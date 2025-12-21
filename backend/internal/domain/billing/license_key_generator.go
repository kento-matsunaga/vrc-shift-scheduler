package billing

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// GenerateLicenseKey generates a new license key in format XXXX-XXXX-XXXX-XXXX
func GenerateLicenseKey() (string, error) {
	// Generate 8 random bytes (64 bits)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Convert to hex string (16 characters)
	hexStr := strings.ToUpper(hex.EncodeToString(randomBytes))

	// Format as XXXX-XXXX-XXXX-XXXX
	return formatLicenseKey(hexStr), nil
}

// formatLicenseKey formats a 16-character string as XXXX-XXXX-XXXX-XXXX
func formatLicenseKey(s string) string {
	if len(s) != 16 {
		return s
	}
	return s[0:4] + "-" + s[4:8] + "-" + s[8:12] + "-" + s[12:16]
}

// NormalizeLicenseKey removes hyphens from a license key for hashing
func NormalizeLicenseKey(key string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(key)), "-", "")
}

// ValidateLicenseKeyFormat validates the format of a license key
func ValidateLicenseKeyFormat(key string) bool {
	normalized := NormalizeLicenseKey(key)
	if len(normalized) != 16 {
		return false
	}
	// Check if all characters are valid hex characters
	for _, c := range normalized {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
