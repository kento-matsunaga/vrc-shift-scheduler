package security

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	AdminID  string `json:"admin_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenIssuer is an interface for issuing JWT tokens
type TokenIssuer interface {
	Issue(adminID, tenantID, role string) (token string, expiresAt time.Time, err error)
}

// TokenVerifier is an interface for verifying JWT tokens
type TokenVerifier interface {
	Verify(token string) (*JWTClaims, error)
}

// JWTManager implements both TokenIssuer and TokenVerifier
type JWTManager struct {
	secretKey      []byte
	expirationTime time.Duration
}

// NewJWTManager creates a new JWTManager
// JWT_SECRET 環境変数が必須。なければpanicする。
func NewJWTManager() *JWTManager {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	// デフォルトの有効期限は24時間
	expirationTime := 24 * time.Hour
	return &JWTManager{
		secretKey:      []byte(secret),
		expirationTime: expirationTime,
	}
}

// NewJWTManagerWithExpiration creates a new JWTManager with custom expiration time
func NewJWTManagerWithExpiration(expirationTime time.Duration) *JWTManager {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	return &JWTManager{
		secretKey:      []byte(secret),
		expirationTime: expirationTime,
	}
}

// Issue issues a new JWT token
func (m *JWTManager) Issue(adminID, tenantID, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.expirationTime)

	claims := JWTClaims{
		AdminID:  adminID,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// Verify verifies a JWT token and returns the claims
func (m *JWTManager) Verify(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
