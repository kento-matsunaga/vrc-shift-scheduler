package services

// PasswordHasher is an interface for password hashing and verification.
// This interface allows the Application layer to be independent of
// the specific hashing implementation (e.g., bcrypt).
type PasswordHasher interface {
	// Hash generates a hash from the given password
	Hash(password string) (string, error)
	// Compare compares a hash with a password and returns nil if they match
	Compare(hash, password string) error
}
