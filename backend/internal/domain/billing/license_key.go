package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// LicenseKeyID represents the unique identifier for a license key
type LicenseKeyID string

// NewLicenseKeyID generates a new LicenseKeyID
func NewLicenseKeyID() LicenseKeyID {
	return LicenseKeyID(common.NewULID())
}

// ParseLicenseKeyID parses a string into LicenseKeyID
func ParseLicenseKeyID(id string) (LicenseKeyID, error) {
	if err := common.ValidateULID(id); err != nil {
		return "", err
	}
	return LicenseKeyID(id), nil
}

// String returns the string representation
func (id LicenseKeyID) String() string {
	return string(id)
}

// LicenseKeyStatus represents the status of a license key
type LicenseKeyStatus string

const (
	LicenseKeyStatusUnused  LicenseKeyStatus = "unused"
	LicenseKeyStatusUsed    LicenseKeyStatus = "used"
	LicenseKeyStatusRevoked LicenseKeyStatus = "revoked"
)

// String returns the string representation
func (s LicenseKeyStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s LicenseKeyStatus) IsValid() bool {
	return s == LicenseKeyStatusUnused || s == LicenseKeyStatusUsed || s == LicenseKeyStatusRevoked
}

// HashLicenseKey computes SHA-256 hash of a license key
func HashLicenseKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// LicenseKey represents a BOOTH license key
type LicenseKey struct {
	keyID        LicenseKeyID
	keyHash      string
	status       LicenseKeyStatus
	batchID      *string
	expiresAt    *time.Time
	memo         string
	usedAt       *time.Time
	usedTenantID *common.TenantID
	revokedAt    *time.Time
	createdAt    time.Time
}

// NewLicenseKey creates a new LicenseKey entity
func NewLicenseKey(
	now time.Time,
	keyHash string,
	expiresAt *time.Time,
	memo string,
) (*LicenseKey, error) {
	key := &LicenseKey{
		keyID:        NewLicenseKeyID(),
		keyHash:      keyHash,
		status:       LicenseKeyStatusUnused,
		batchID:      nil,
		expiresAt:    expiresAt,
		memo:         memo,
		usedAt:       nil,
		usedTenantID: nil,
		revokedAt:    nil,
		createdAt:    now,
	}

	if err := key.validate(); err != nil {
		return nil, err
	}

	return key, nil
}

// ReconstructLicenseKey reconstructs a LicenseKey entity from persistence
func ReconstructLicenseKey(
	keyID LicenseKeyID,
	keyHash string,
	status LicenseKeyStatus,
	batchID *string,
	expiresAt *time.Time,
	memo string,
	usedAt *time.Time,
	usedTenantID *common.TenantID,
	revokedAt *time.Time,
	createdAt time.Time,
) (*LicenseKey, error) {
	key := &LicenseKey{
		keyID:        keyID,
		keyHash:      keyHash,
		status:       status,
		batchID:      batchID,
		expiresAt:    expiresAt,
		memo:         memo,
		usedAt:       usedAt,
		usedTenantID: usedTenantID,
		revokedAt:    revokedAt,
		createdAt:    createdAt,
	}

	if err := key.validate(); err != nil {
		return nil, err
	}

	return key, nil
}

func (k *LicenseKey) validate() error {
	if k.keyHash == "" {
		return common.NewValidationError("key_hash is required", nil)
	}
	if len(k.keyHash) != 64 {
		return common.NewValidationError("key_hash must be 64 characters (SHA-256 hex)", nil)
	}
	if !k.status.IsValid() {
		return common.NewValidationError("invalid license key status", nil)
	}
	return nil
}

// Getters

func (k *LicenseKey) KeyID() LicenseKeyID {
	return k.keyID
}

func (k *LicenseKey) KeyHash() string {
	return k.keyHash
}

func (k *LicenseKey) Status() LicenseKeyStatus {
	return k.status
}

func (k *LicenseKey) BatchID() *string {
	return k.batchID
}

func (k *LicenseKey) ExpiresAt() *time.Time {
	return k.expiresAt
}

func (k *LicenseKey) Memo() string {
	return k.memo
}

func (k *LicenseKey) UsedAt() *time.Time {
	return k.usedAt
}

// ClaimedAt returns when the key was used (alias for UsedAt)
func (k *LicenseKey) ClaimedAt() *time.Time {
	return k.usedAt
}

// ClaimedBy returns the tenant that used this key (alias for UsedTenantID)
func (k *LicenseKey) ClaimedBy() *common.TenantID {
	return k.usedTenantID
}

func (k *LicenseKey) UsedTenantID() *common.TenantID {
	return k.usedTenantID
}

func (k *LicenseKey) RevokedAt() *time.Time {
	return k.revokedAt
}

func (k *LicenseKey) CreatedAt() time.Time {
	return k.createdAt
}

// IsUnused checks if the key is unused
func (k *LicenseKey) IsUnused() bool {
	return k.status == LicenseKeyStatusUnused
}

// IsUsed checks if the key has been used
func (k *LicenseKey) IsUsed() bool {
	return k.status == LicenseKeyStatusUsed
}

// IsRevoked checks if the key has been revoked
func (k *LicenseKey) IsRevoked() bool {
	return k.status == LicenseKeyStatusRevoked
}

// MarkAsUsed marks the key as used
func (k *LicenseKey) MarkAsUsed(now time.Time, tenantID common.TenantID) error {
	if k.status != LicenseKeyStatusUnused {
		return common.NewValidationError("license key is not available", nil)
	}
	k.status = LicenseKeyStatusUsed
	k.usedAt = &now
	k.usedTenantID = &tenantID
	return nil
}

// Revoke revokes the key
func (k *LicenseKey) Revoke(now time.Time) error {
	if k.status == LicenseKeyStatusRevoked {
		return common.NewValidationError("license key is already revoked", nil)
	}
	k.status = LicenseKeyStatusRevoked
	k.revokedAt = &now
	return nil
}
