package security

import "golang.org/x/crypto/bcrypt"

// PasswordHasher is an interface for hashing and comparing passwords
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// BcryptHasher is a bcrypt implementation of PasswordHasher
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new BcryptHasher with the default cost
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{cost: bcrypt.DefaultCost}
}

// NewBcryptHasherWithCost creates a new BcryptHasher with a custom cost
func NewBcryptHasherWithCost(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

// Hash hashes a password using bcrypt
func (h *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare compares a hash with a password
// Returns nil if the password matches the hash, otherwise returns an error
func (h *BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
