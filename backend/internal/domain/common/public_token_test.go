package common

import (
	"testing"
)

func TestNewPublicToken(t *testing.T) {
	token := NewPublicToken()

	// Should be a valid UUID
	if err := token.Validate(); err != nil {
		t.Errorf("NewPublicToken() generated invalid token: %v", err)
	}

	// Should be 36 characters (UUID format with hyphens)
	if len(token.String()) != 36 {
		t.Errorf("NewPublicToken() length = %d, want 36", len(token.String()))
	}
}

func TestValidatePublicToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid UUID v4",
			token:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "valid UUID v4 lowercase",
			token:   "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - too short",
			token:   "abc123",
			wantErr: true,
		},
		{
			name:    "invalid format - nanoid style",
			token:   "V1StGXR8_Z5jdHi6B-myT",
			wantErr: true,
		},
		{
			name:    "valid UUID without hyphens (also accepted)",
			token:   "550e8400e29b41d4a716446655440000",
			wantErr: false,
		},
		{
			name:    "invalid format - wrong characters",
			token:   "550e8400-e29b-41d4-a716-44665544000g",
			wantErr: true,
		},
		{
			name:    "invalid format - ULID",
			token:   "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePublicToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePublicToken(%q) error = %v, wantErr %v", tt.token, err, tt.wantErr)
			}
		})
	}
}

func TestParsePublicToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		validUUID := "550e8400-e29b-41d4-a716-446655440000"
		token, err := ParsePublicToken(validUUID)
		if err != nil {
			t.Errorf("ParsePublicToken(%q) unexpected error: %v", validUUID, err)
		}
		if token.String() != validUUID {
			t.Errorf("ParsePublicToken(%q) = %q, want %q", validUUID, token.String(), validUUID)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		invalidToken := "not-a-uuid"
		_, err := ParsePublicToken(invalidToken)
		if err == nil {
			t.Errorf("ParsePublicToken(%q) expected error, got nil", invalidToken)
		}
	})
}

func TestPublicTokenUniqueness(t *testing.T) {
	// Generate multiple tokens and ensure they are unique
	tokens := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		token := NewPublicToken()
		if tokens[token.String()] {
			t.Errorf("NewPublicToken() generated duplicate token: %s", token)
		}
		tokens[token.String()] = true
	}
}

