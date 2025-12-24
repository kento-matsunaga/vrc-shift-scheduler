package billing

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// =====================================================
// NewLicenseKeyID Tests
// =====================================================

func TestNewLicenseKeyID_Success(t *testing.T) {
	id := NewLicenseKeyID()

	if id == "" {
		t.Error("NewLicenseKeyID() should not return empty string")
	}
}

func TestNewLicenseKeyIDWithTime_Success(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	id := NewLicenseKeyIDWithTime(fixedTime)

	if id == "" {
		t.Error("NewLicenseKeyIDWithTime() should not return empty string")
	}
}

func TestParseLicenseKeyID(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		validID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
		id, err := ParseLicenseKeyID(validID)
		if err != nil {
			t.Errorf("ParseLicenseKeyID(%q) unexpected error: %v", validID, err)
		}
		if id.String() != validID {
			t.Errorf("ParseLicenseKeyID(%q) = %q, want %q", validID, id.String(), validID)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		invalidID := "invalid"
		_, err := ParseLicenseKeyID(invalidID)
		if err == nil {
			t.Errorf("ParseLicenseKeyID(%q) expected error, got nil", invalidID)
		}
	})
}

// =====================================================
// LicenseKeyStatus Tests
// =====================================================

func TestLicenseKeyStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status LicenseKeyStatus
		want   bool
	}{
		{"unused is valid", LicenseKeyStatusUnused, true},
		{"used is valid", LicenseKeyStatusUsed, true},
		{"revoked is valid", LicenseKeyStatusRevoked, true},
		{"invalid status", LicenseKeyStatus("invalid"), false},
		{"empty status", LicenseKeyStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicenseKeyStatus_String(t *testing.T) {
	tests := []struct {
		status LicenseKeyStatus
		want   string
	}{
		{LicenseKeyStatusUnused, "unused"},
		{LicenseKeyStatusUsed, "used"},
		{LicenseKeyStatusRevoked, "revoked"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

// =====================================================
// HashLicenseKey Tests
// =====================================================

func TestHashLicenseKey_Success(t *testing.T) {
	key := "TEST-LICENSE-KEY-12345"
	hash := HashLicenseKey(key)

	// SHA-256 produces 64-character hex string
	if len(hash) != 64 {
		t.Errorf("HashLicenseKey() length = %d, want 64", len(hash))
	}

	// Same input should produce same hash
	hash2 := HashLicenseKey(key)
	if hash != hash2 {
		t.Error("HashLicenseKey() should produce consistent results")
	}
}

func TestHashLicenseKey_DifferentInputsDifferentHashes(t *testing.T) {
	hash1 := HashLicenseKey("KEY-1")
	hash2 := HashLicenseKey("KEY-2")

	if hash1 == hash2 {
		t.Error("HashLicenseKey() should produce different hashes for different inputs")
	}
}

// =====================================================
// NewLicenseKey Tests - Success Cases
// =====================================================

func TestNewLicenseKey_Success(t *testing.T) {
	now := time.Now()
	keyHash := HashLicenseKey("TEST-LICENSE-KEY")
	memo := "Test key"

	key, err := NewLicenseKey(now, keyHash, nil, memo)

	if err != nil {
		t.Fatalf("NewLicenseKey() should succeed, but got error: %v", err)
	}

	if key == nil {
		t.Fatal("NewLicenseKey() returned nil")
	}

	// Basic field validation
	if key.KeyHash() != keyHash {
		t.Errorf("KeyHash: expected %s, got %s", keyHash, key.KeyHash())
	}
	if key.Memo() != memo {
		t.Errorf("Memo: expected %s, got %s", memo, key.Memo())
	}

	// Default values
	if key.Status() != LicenseKeyStatusUnused {
		t.Errorf("Status: expected %s, got %s", LicenseKeyStatusUnused, key.Status())
	}
	if !key.IsUnused() {
		t.Error("IsUnused should be true by default")
	}
	if key.IsUsed() {
		t.Error("IsUsed should be false by default")
	}
	if key.IsRevoked() {
		t.Error("IsRevoked should be false by default")
	}

	// ID should be generated
	if key.KeyID() == "" {
		t.Error("KeyID should not be empty")
	}

	// Optional fields should be nil
	if key.BatchID() != nil {
		t.Error("BatchID should be nil by default")
	}
	if key.UsedAt() != nil {
		t.Error("UsedAt should be nil by default")
	}
	if key.UsedTenantID() != nil {
		t.Error("UsedTenantID should be nil by default")
	}
	if key.RevokedAt() != nil {
		t.Error("RevokedAt should be nil by default")
	}
}

func TestNewLicenseKey_SuccessWithExpiresAt(t *testing.T) {
	now := time.Now()
	keyHash := HashLicenseKey("TEST-LICENSE-KEY")
	expiresAt := now.Add(30 * 24 * time.Hour)

	key, err := NewLicenseKey(now, keyHash, &expiresAt, "")

	if err != nil {
		t.Fatalf("NewLicenseKey() should succeed, but got error: %v", err)
	}

	if key.ExpiresAt() == nil {
		t.Fatal("ExpiresAt should not be nil")
	}
	if !key.ExpiresAt().Equal(expiresAt) {
		t.Errorf("ExpiresAt: expected %v, got %v", expiresAt, key.ExpiresAt())
	}
}

// =====================================================
// NewLicenseKey Tests - Error Cases
// =====================================================

func TestNewLicenseKey_ErrorWhenKeyHashEmpty(t *testing.T) {
	now := time.Now()
	keyHash := "" // Empty

	key, err := NewLicenseKey(now, keyHash, nil, "")

	if err == nil {
		t.Fatal("NewLicenseKey() should return error when key_hash is empty")
	}
	if key != nil {
		t.Error("NewLicenseKey() should return nil when validation fails")
	}
}

func TestNewLicenseKey_ErrorWhenKeyHashWrongLength(t *testing.T) {
	now := time.Now()
	keyHash := "tooshort" // Not 64 characters

	key, err := NewLicenseKey(now, keyHash, nil, "")

	if err == nil {
		t.Fatal("NewLicenseKey() should return error when key_hash is not 64 characters")
	}
	if key != nil {
		t.Error("NewLicenseKey() should return nil when validation fails")
	}
}

// =====================================================
// MarkAsUsed Tests
// =====================================================

func TestLicenseKey_MarkAsUsed_Success(t *testing.T) {
	now := time.Now()
	keyHash := HashLicenseKey("TEST-KEY")
	key, _ := NewLicenseKey(now, keyHash, nil, "")

	tenantID := common.NewTenantID()
	useTime := now.Add(1 * time.Hour)
	err := key.MarkAsUsed(useTime, tenantID)

	if err != nil {
		t.Fatalf("MarkAsUsed() should succeed, but got error: %v", err)
	}

	if key.Status() != LicenseKeyStatusUsed {
		t.Errorf("Status: expected %s, got %s", LicenseKeyStatusUsed, key.Status())
	}
	if !key.IsUsed() {
		t.Error("IsUsed should be true after MarkAsUsed()")
	}
	if key.IsUnused() {
		t.Error("IsUnused should be false after MarkAsUsed()")
	}
	if key.UsedAt() == nil {
		t.Fatal("UsedAt should not be nil after MarkAsUsed()")
	}
	if !key.UsedAt().Equal(useTime) {
		t.Errorf("UsedAt: expected %v, got %v", useTime, key.UsedAt())
	}
	if key.UsedTenantID() == nil {
		t.Fatal("UsedTenantID should not be nil after MarkAsUsed()")
	}
	if *key.UsedTenantID() != tenantID {
		t.Errorf("UsedTenantID: expected %s, got %s", tenantID, *key.UsedTenantID())
	}

	// Test alias methods
	if key.ClaimedAt() == nil {
		t.Fatal("ClaimedAt should not be nil after MarkAsUsed()")
	}
	if key.ClaimedBy() == nil {
		t.Fatal("ClaimedBy should not be nil after MarkAsUsed()")
	}
}

func TestLicenseKey_MarkAsUsed_ErrorWhenAlreadyUsed(t *testing.T) {
	now := time.Now()
	keyHash := HashLicenseKey("TEST-KEY")
	key, _ := NewLicenseKey(now, keyHash, nil, "")

	tenantID := common.NewTenantID()
	_ = key.MarkAsUsed(now, tenantID)

	// Try to use again
	anotherTenantID := common.NewTenantID()
	err := key.MarkAsUsed(now, anotherTenantID)

	if err == nil {
		t.Fatal("MarkAsUsed() should return error when key is already used")
	}
}

func TestLicenseKey_MarkAsUsed_ErrorWhenRevoked(t *testing.T) {
	now := time.Now()
	keyHash := HashLicenseKey("TEST-KEY")
	key, _ := NewLicenseKey(now, keyHash, nil, "")

	// Revoke first
	_ = key.Revoke(now)

	// Try to use
	tenantID := common.NewTenantID()
	err := key.MarkAsUsed(now, tenantID)

	if err == nil {
		t.Fatal("MarkAsUsed() should return error when key is revoked")
	}
}

// =====================================================
// Revoke Tests
// =====================================================

func TestLicenseKey_Revoke_Success(t *testing.T) {
	now := time.Now()
	keyHash := HashLicenseKey("TEST-KEY")
	key, _ := NewLicenseKey(now, keyHash, nil, "")

	revokeTime := now.Add(1 * time.Hour)
	err := key.Revoke(revokeTime)

	if err != nil {
		t.Fatalf("Revoke() should succeed, but got error: %v", err)
	}

	if key.Status() != LicenseKeyStatusRevoked {
		t.Errorf("Status: expected %s, got %s", LicenseKeyStatusRevoked, key.Status())
	}
	if !key.IsRevoked() {
		t.Error("IsRevoked should be true after Revoke()")
	}
	if key.RevokedAt() == nil {
		t.Fatal("RevokedAt should not be nil after Revoke()")
	}
	if !key.RevokedAt().Equal(revokeTime) {
		t.Errorf("RevokedAt: expected %v, got %v", revokeTime, key.RevokedAt())
	}
}

func TestLicenseKey_Revoke_ErrorWhenAlreadyRevoked(t *testing.T) {
	now := time.Now()
	keyHash := HashLicenseKey("TEST-KEY")
	key, _ := NewLicenseKey(now, keyHash, nil, "")

	// Revoke first
	_ = key.Revoke(now)

	// Try to revoke again
	err := key.Revoke(now)

	if err == nil {
		t.Fatal("Revoke() should return error when key is already revoked")
	}
}

func TestLicenseKey_Revoke_SuccessWhenUsed(t *testing.T) {
	now := time.Now()
	keyHash := HashLicenseKey("TEST-KEY")
	key, _ := NewLicenseKey(now, keyHash, nil, "")

	// Use first
	tenantID := common.NewTenantID()
	_ = key.MarkAsUsed(now, tenantID)

	// Can still revoke even if used
	err := key.Revoke(now)

	if err != nil {
		t.Fatalf("Revoke() should succeed even for used keys, but got error: %v", err)
	}
	if !key.IsRevoked() {
		t.Error("IsRevoked should be true after Revoke()")
	}
}

// =====================================================
// ReconstructLicenseKey Tests
// =====================================================

func TestReconstructLicenseKey_Success(t *testing.T) {
	keyID := NewLicenseKeyID()
	keyHash := HashLicenseKey("TEST-KEY")
	now := time.Now()

	key, err := ReconstructLicenseKey(
		keyID,
		keyHash,
		LicenseKeyStatusUnused,
		nil,
		nil,
		"",
		nil,
		nil,
		nil,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructLicenseKey() should succeed, but got error: %v", err)
	}
	if key == nil {
		t.Fatal("ReconstructLicenseKey() returned nil")
	}
	if key.KeyID() != keyID {
		t.Errorf("KeyID: expected %s, got %s", keyID, key.KeyID())
	}
}

func TestReconstructLicenseKey_WithUsedState(t *testing.T) {
	keyID := NewLicenseKeyID()
	keyHash := HashLicenseKey("TEST-KEY")
	now := time.Now()
	usedAt := now.Add(-1 * time.Hour)
	tenantID := common.NewTenantID()

	key, err := ReconstructLicenseKey(
		keyID,
		keyHash,
		LicenseKeyStatusUsed,
		nil,
		nil,
		"Used key",
		&usedAt,
		&tenantID,
		nil,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructLicenseKey() should succeed, but got error: %v", err)
	}
	if !key.IsUsed() {
		t.Error("LicenseKey should be marked as used")
	}
	if key.UsedAt() == nil {
		t.Fatal("UsedAt should not be nil")
	}
	if key.UsedTenantID() == nil {
		t.Fatal("UsedTenantID should not be nil")
	}
}

func TestReconstructLicenseKey_WithRevokedState(t *testing.T) {
	keyID := NewLicenseKeyID()
	keyHash := HashLicenseKey("TEST-KEY")
	now := time.Now()
	revokedAt := now.Add(-1 * time.Hour)

	key, err := ReconstructLicenseKey(
		keyID,
		keyHash,
		LicenseKeyStatusRevoked,
		nil,
		nil,
		"",
		nil,
		nil,
		&revokedAt,
		now,
	)

	if err != nil {
		t.Fatalf("ReconstructLicenseKey() should succeed, but got error: %v", err)
	}
	if !key.IsRevoked() {
		t.Error("LicenseKey should be marked as revoked")
	}
	if key.RevokedAt() == nil {
		t.Fatal("RevokedAt should not be nil")
	}
}

func TestReconstructLicenseKey_ErrorWhenValidationFails(t *testing.T) {
	keyID := NewLicenseKeyID()
	keyHash := "invalid" // Not 64 characters
	now := time.Now()

	key, err := ReconstructLicenseKey(
		keyID,
		keyHash,
		LicenseKeyStatusUnused,
		nil,
		nil,
		"",
		nil,
		nil,
		nil,
		now,
	)

	if err == nil {
		t.Fatal("ReconstructLicenseKey() should return error when validation fails")
	}
	if key != nil {
		t.Error("ReconstructLicenseKey() should return nil when validation fails")
	}
}

func TestReconstructLicenseKey_ErrorWhenStatusInvalid(t *testing.T) {
	keyID := NewLicenseKeyID()
	keyHash := HashLicenseKey("TEST-KEY")
	now := time.Now()

	key, err := ReconstructLicenseKey(
		keyID,
		keyHash,
		LicenseKeyStatus("invalid"), // Invalid status
		nil,
		nil,
		"",
		nil,
		nil,
		nil,
		now,
	)

	if err == nil {
		t.Fatal("ReconstructLicenseKey() should return error when status is invalid")
	}
	if key != nil {
		t.Error("ReconstructLicenseKey() should return nil when validation fails")
	}
}
