package services

import "time"

// TokenIssuer is an interface for issuing JWT tokens.
// This interface allows the Application layer to be independent of
// the specific JWT implementation.
type TokenIssuer interface {
	// Issue generates a new JWT token with the given admin info
	Issue(adminID, tenantID, role string) (token string, expiresAt time.Time, err error)
}
