package importjob

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

func createTestMember(t *testing.T, displayName string) *member.Member {
	t.Helper()
	tenantID := common.NewTenantIDWithTime(time.Now())
	m, err := member.NewMember(time.Now(), tenantID, displayName, "", "")
	if err != nil {
		t.Fatalf("failed to create test member: %v", err)
	}
	return m
}

func TestMemberMatcher_Match_ExactMatch(t *testing.T) {
	members := []*member.Member{
		createTestMember(t, "ラット"),
		createTestMember(t, "もやし"),
		createTestMember(t, "おおちゃん"),
	}

	matcher := NewMemberMatcher(members, false)

	tests := []struct {
		name      string
		input     string
		wantMatch bool
		wantName  string
	}{
		{
			name:      "完全一致",
			input:     "ラット",
			wantMatch: true,
			wantName:  "ラット",
		},
		{
			name:      "大文字小文字無視（英字）",
			input:     "らっと", // This won't match because we're matching display name
			wantMatch: false,
		},
		{
			name:      "存在しないメンバー",
			input:     "存在しない",
			wantMatch: false,
		},
		{
			name:      "空文字",
			input:     "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matcher.Match(tt.input)

			if tt.input == "" {
				if err == nil {
					t.Error("expected error for empty input")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.wantMatch {
				if result == nil {
					t.Error("expected match, got nil")
					return
				}
				if result.DisplayName() != tt.wantName {
					t.Errorf("got %q, want %q", result.DisplayName(), tt.wantName)
				}
			} else {
				if result != nil {
					t.Errorf("expected no match, got %q", result.DisplayName())
				}
			}
		})
	}
}

func TestMemberMatcher_Match_FuzzyMatch(t *testing.T) {
	members := []*member.Member{
		createTestMember(t, "ラット"),
		createTestMember(t, "モヤシ"), // カタカナ
		createTestMember(t, "おおちゃん"),
	}

	matcher := NewMemberMatcher(members, true) // fuzzy enabled

	tests := []struct {
		name      string
		input     string
		wantMatch bool
		wantName  string
	}{
		{
			name:      "カタカナ→ひらがな変換",
			input:     "もやし", // ひらがなで入力
			wantMatch: true,
			wantName:  "モヤシ",
		},
		{
			name:      "ひらがな→カタカナは不一致（正規化後の比較）",
			input:     "らっと", // ひらがなで入力、元はカタカナ
			wantMatch: true,
			wantName:  "ラット",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matcher.Match(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.wantMatch {
				if result == nil {
					t.Error("expected match, got nil")
					return
				}
				if result.DisplayName() != tt.wantName {
					t.Errorf("got %q, want %q", result.DisplayName(), tt.wantName)
				}
			} else {
				if result != nil {
					t.Errorf("expected no match, got %q", result.DisplayName())
				}
			}
		})
	}
}

func TestKatakanaToHiragana(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"アイウエオ", "あいうえお"},
		{"カタカナ", "かたかな"},
		{"ラット", "らっと"},
		{"モヤシ", "もやし"},
		{"ひらがな", "ひらがな"},     // Already hiragana
		{"ABC123", "ABC123"}, // Non-Japanese
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := katakanaToHiragana(tt.input)
			if got != tt.want {
				t.Errorf("katakanaToHiragana(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFullWidthToHalfWidth(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"０１２３４５", "012345"},
		{"ＡＢＣＤ", "ABCD"},
		{"ａｂｃｄ", "abcd"},
		{"abc123", "abc123"}, // Already half-width
		{"あいうえお", "あいうえお"},   // Non-alphanumeric
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := fullWidthToHalfWidth(tt.input)
			if got != tt.want {
				t.Errorf("fullWidthToHalfWidth(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMemberMatcher_MatchAll(t *testing.T) {
	members := []*member.Member{
		createTestMember(t, "ラット"),
		createTestMember(t, "もやし"),
	}

	matcher := NewMemberMatcher(members, false)

	names := []string{"ラット", "存在しない", "もやし"}
	results := matcher.MatchAll(names)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// First: match
	if !results[0].Matched || results[0].Member.DisplayName() != "ラット" {
		t.Errorf("results[0] should match 'ラット'")
	}

	// Second: no match
	if results[1].Matched {
		t.Errorf("results[1] should not match")
	}

	// Third: match
	if !results[2].Matched || results[2].Member.DisplayName() != "もやし" {
		t.Errorf("results[2] should match 'もやし'")
	}
}

func TestNormalizeForExactMatch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ABC", "abc"},
		{"  test  ", "test"},
		{"Mixed CASE", "mixed case"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeForExactMatch(tt.input)
			if got != tt.want {
				t.Errorf("normalizeForExactMatch(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeForFuzzyMatch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ラット", "らっと"},
		{"モ ヤ シ", "もやし"},
		{"Ａ Ｂ　Ｃ", "abc"},
		{"  test  ", "test"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeForFuzzyMatch(tt.input)
			if got != tt.want {
				t.Errorf("normalizeForFuzzyMatch(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
