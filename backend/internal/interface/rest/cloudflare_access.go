package rest

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// decodeJSONBody decodes a JSON body into the given struct
func decodeJSONBody(body io.Reader, v interface{}) error {
	return json.NewDecoder(body).Decode(v)
}

// CloudflareAccessConfig holds the configuration for Cloudflare Access validation
type CloudflareAccessConfig struct {
	// TeamDomain is your Cloudflare Access team domain (e.g., "yourteam.cloudflareaccess.com")
	TeamDomain string
	// PolicyAUD is the Application Audience (AUD) tag from Cloudflare Access
	PolicyAUD string
	// AllowedEmails is a list of email addresses allowed to access (optional additional check)
	AllowedEmails []string
	// Enabled controls whether Cloudflare Access validation is enabled
	Enabled bool
}

// ContextKeyCFEmail is the context key for Cloudflare Access email
const ContextKeyCFEmail ContextKey = "cf_access_email"

// CloudflareAccessMiddleware validates Cloudflare Access JWT tokens
// When enabled, it verifies the CF-Access-JWT-Assertion header
// When disabled (local development), it allows all requests
func CloudflareAccessMiddleware(config CloudflareAccessConfig) func(http.Handler) http.Handler {
	// Cache for Cloudflare's public keys
	var (
		cachedKeys   map[string]interface{}
		keysMutex    sync.RWMutex
		keysExpiry   time.Time
		keysCacheTTL = 1 * time.Hour
	)

	// Fetch Cloudflare's public keys
	fetchPublicKeys := func() (map[string]interface{}, error) {
		keysMutex.RLock()
		if cachedKeys != nil && time.Now().Before(keysExpiry) {
			defer keysMutex.RUnlock()
			return cachedKeys, nil
		}
		keysMutex.RUnlock()

		// Fetch keys from Cloudflare
		certsURL := fmt.Sprintf("https://%s/cdn-cgi/access/certs", config.TeamDomain)
		resp, err := http.Get(certsURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Cloudflare certs: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch Cloudflare certs: status %d", resp.StatusCode)
		}

		// Parse the response (Cloudflare returns PEM-encoded public keys)
		// For simplicity, we'll store the raw response and parse as needed
		keysMutex.Lock()
		defer keysMutex.Unlock()

		// In production, parse the JWKS or PEM keys here
		// For now, we'll use a simplified approach
		cachedKeys = make(map[string]interface{})
		keysExpiry = time.Now().Add(keysCacheTTL)

		return cachedKeys, nil
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip validation if disabled (local development)
			if !config.Enabled {
				// In development, optionally check for a bypass header
				devEmail := r.Header.Get("X-Admin-Email")
				if devEmail != "" {
					ctx := context.WithValue(r.Context(), ContextKeyCFEmail, devEmail)
					r = r.WithContext(ctx)
				}
				next.ServeHTTP(w, r)
				return
			}

			// Get the Cloudflare Access JWT from header
			cfJWT := r.Header.Get("CF-Access-JWT-Assertion")
			if cfJWT == "" {
				// Also check cookie (Cloudflare sets both)
				cookie, err := r.Cookie("CF_Authorization")
				if err == nil {
					cfJWT = cookie.Value
				}
			}

			if cfJWT == "" {
				RespondError(w, http.StatusUnauthorized, "ERR_CF_ACCESS_REQUIRED",
					"Cloudflare Access authentication required", nil)
				return
			}

			// Parse and validate the JWT
			token, err := jwt.Parse(cfJWT, func(token *jwt.Token) (interface{}, error) {
				// Verify the signing method
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				// Get kid from token header
				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, fmt.Errorf("missing kid in token header")
				}

				// Fetch public keys from Cloudflare
				_, err := fetchPublicKeys()
				if err != nil {
					return nil, err
				}

				// In production, look up the key by kid
				// For now, we'll fetch the cert directly
				certsURL := fmt.Sprintf("https://%s/cdn-cgi/access/certs", config.TeamDomain)
				return fetchCertByKid(certsURL, kid)
			})

			if err != nil {
				RespondError(w, http.StatusUnauthorized, "ERR_CF_ACCESS_INVALID",
					"Invalid Cloudflare Access token", map[string]interface{}{"detail": err.Error()})
				return
			}

			if !token.Valid {
				RespondError(w, http.StatusUnauthorized, "ERR_CF_ACCESS_INVALID",
					"Invalid Cloudflare Access token", nil)
				return
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				RespondError(w, http.StatusUnauthorized, "ERR_CF_ACCESS_INVALID",
					"Invalid token claims", nil)
				return
			}

			// Verify audience
			aud, ok := claims["aud"].([]interface{})
			if !ok {
				// Try single audience
				singleAud, ok := claims["aud"].(string)
				if !ok || singleAud != config.PolicyAUD {
					RespondError(w, http.StatusForbidden, "ERR_CF_ACCESS_AUD_MISMATCH",
						"Token audience mismatch", nil)
					return
				}
			} else {
				audMatch := false
				for _, a := range aud {
					if a == config.PolicyAUD {
						audMatch = true
						break
					}
				}
				if !audMatch {
					RespondError(w, http.StatusForbidden, "ERR_CF_ACCESS_AUD_MISMATCH",
						"Token audience mismatch", nil)
					return
				}
			}

			// Extract email
			email, _ := claims["email"].(string)

			// Check allowed emails if configured
			if len(config.AllowedEmails) > 0 {
				allowed := false
				for _, allowedEmail := range config.AllowedEmails {
					if strings.EqualFold(email, allowedEmail) {
						allowed = true
						break
					}
				}
				if !allowed {
					RespondError(w, http.StatusForbidden, "ERR_CF_ACCESS_NOT_ALLOWED",
						"Email not in allowed list", nil)
					return
				}
			}

			// Store email in context
			ctx := context.WithValue(r.Context(), ContextKeyCFEmail, email)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// fetchCertByKid fetches the public key from Cloudflare by kid
func fetchCertByKid(certsURL, kid string) (interface{}, error) {
	resp, err := http.Get(certsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response - Cloudflare returns a JSON with public_certs array
	// Each cert has a kid and cert (PEM-encoded)
	type CertResponse struct {
		Keys       []map[string]interface{} `json:"keys"`
		PublicCert struct {
			Kid  string `json:"kid"`
			Cert string `json:"cert"`
		} `json:"public_cert"`
		PublicCerts []struct {
			Kid  string `json:"kid"`
			Cert string `json:"cert"`
		} `json:"public_certs"`
	}

	var certResp CertResponse
	if err := decodeJSONBody(resp.Body, &certResp); err != nil {
		return nil, fmt.Errorf("failed to parse certs response: %w", err)
	}

	// Find the cert with matching kid
	for _, cert := range certResp.PublicCerts {
		if cert.Kid == kid {
			// Parse PEM
			block, _ := pem.Decode([]byte(cert.Cert))
			if block == nil {
				return nil, fmt.Errorf("failed to parse PEM block")
			}

			// Parse certificate
			parsedCert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}

			return parsedCert.PublicKey, nil
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// GetCloudflareEmail extracts the Cloudflare Access email from context
func GetCloudflareEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(ContextKeyCFEmail).(string)
	return email, ok
}

// LoadCloudflareAccessConfig loads configuration from environment variables
func LoadCloudflareAccessConfig() CloudflareAccessConfig {
	teamDomain := os.Getenv("CF_ACCESS_TEAM_DOMAIN")
	policyAUD := os.Getenv("CF_ACCESS_POLICY_AUD")
	allowedEmailsStr := os.Getenv("CF_ACCESS_ALLOWED_EMAILS")

	var allowedEmails []string
	if allowedEmailsStr != "" {
		allowedEmails = strings.Split(allowedEmailsStr, ",")
		for i := range allowedEmails {
			allowedEmails[i] = strings.TrimSpace(allowedEmails[i])
		}
	}

	// Enable only if team domain and policy AUD are set
	enabled := teamDomain != "" && policyAUD != ""

	return CloudflareAccessConfig{
		TeamDomain:    teamDomain,
		PolicyAUD:     policyAUD,
		AllowedEmails: allowedEmails,
		Enabled:       enabled,
	}
}
