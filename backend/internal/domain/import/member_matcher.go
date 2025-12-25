package importjob

import (
	"strings"
	"unicode"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// MemberMatcher handles matching member names from CSV to existing members
type MemberMatcher struct {
	members       []*member.Member
	nameIndex     map[string]*member.Member
	fuzzyEnabled  bool
}

// NewMemberMatcher creates a new member matcher
func NewMemberMatcher(members []*member.Member, fuzzyEnabled bool) *MemberMatcher {
	matcher := &MemberMatcher{
		members:      members,
		nameIndex:    make(map[string]*member.Member),
		fuzzyEnabled: fuzzyEnabled,
	}

	// Build name index
	for _, m := range members {
		// Index by display name (exact match)
		normalized := normalizeForExactMatch(m.DisplayName())
		matcher.nameIndex[normalized] = m
	}

	return matcher
}

// Match attempts to find a member by name
func (m *MemberMatcher) Match(name string) (*member.Member, error) {
	if name == "" {
		return nil, common.NewValidationError("name is empty", nil)
	}

	// Try exact match first
	normalized := normalizeForExactMatch(name)
	if member, ok := m.nameIndex[normalized]; ok {
		return member, nil
	}

	// Try fuzzy match if enabled
	if m.fuzzyEnabled {
		if member := m.fuzzyMatch(name); member != nil {
			return member, nil
		}
	}

	return nil, nil // Not found (not an error)
}

// fuzzyMatch attempts to find a member using fuzzy matching
func (m *MemberMatcher) fuzzyMatch(name string) *member.Member {
	normalizedInput := normalizeForFuzzyMatch(name)

	for _, member := range m.members {
		normalizedMember := normalizeForFuzzyMatch(member.DisplayName())
		if normalizedInput == normalizedMember {
			return member
		}
	}

	return nil
}

// normalizeForExactMatch normalizes a string for exact matching
func normalizeForExactMatch(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// normalizeForFuzzyMatch normalizes a string for fuzzy matching
// Converts katakana to hiragana, removes whitespace, converts to lowercase
func normalizeForFuzzyMatch(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "　", "") // Full-width space
	s = katakanaToHiragana(s)
	s = fullWidthToHalfWidth(s)
	return s
}

// katakanaToHiragana converts katakana characters to hiragana
func katakanaToHiragana(s string) string {
	var result []rune
	for _, r := range s {
		// Katakana range: U+30A1 to U+30F6
		// Hiragana range: U+3041 to U+3096
		// Offset: 0x60
		if r >= 0x30A1 && r <= 0x30F6 {
			result = append(result, r-0x60)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// fullWidthToHalfWidth converts full-width alphanumeric to half-width
func fullWidthToHalfWidth(s string) string {
	var result []rune
	for _, r := range s {
		// Full-width digits: U+FF10 to U+FF19
		if r >= 0xFF10 && r <= 0xFF19 {
			result = append(result, r-0xFEE0)
		} else if r >= 0xFF21 && r <= 0xFF3A {
			// Full-width uppercase: U+FF21 to U+FF3A
			result = append(result, r-0xFEE0)
		} else if r >= 0xFF41 && r <= 0xFF5A {
			// Full-width lowercase: U+FF41 to U+FF5A
			result = append(result, r-0xFEE0)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// MatchResult represents the result of a member match
type MatchResult struct {
	Member    *member.Member
	Matched   bool
	MatchType string // "exact", "fuzzy", or "none"
}

// MatchAll attempts to match multiple names and returns results for each
func (m *MemberMatcher) MatchAll(names []string) []MatchResult {
	results := make([]MatchResult, len(names))

	for i, name := range names {
		member, _ := m.Match(name)
		if member != nil {
			matchType := "exact"
			normalized := normalizeForExactMatch(name)
			if _, ok := m.nameIndex[normalized]; !ok {
				matchType = "fuzzy"
			}
			results[i] = MatchResult{
				Member:    member,
				Matched:   true,
				MatchType: matchType,
			}
		} else {
			results[i] = MatchResult{
				Matched:   false,
				MatchType: "none",
			}
		}
	}

	return results
}

// isAlphanumeric checks if a rune is alphanumeric
func isAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
