package logger

import (
	"fmt"
	"regexp"
	"strings"
)

// DefaultMaskValue is the replacement string used by NewMasker when no
// WithMaskValue option is supplied.
const DefaultMaskValue = "***"

// defaultSensitiveKeys is the set of field keys masked by NewMasker when no
// WithKeys option is supplied. Keys are matched case-insensitively.
var defaultSensitiveKeys = []string{
	// Credentials
	"password", "passwd", "secret", "secret_key",

	// Tokens & sessions
	"access_token", "refresh_token", "id_token", "token",
	"session_id", "session_token", "bearer",

	// API keys & auth
	"api_key", "apikey", "api_secret", "client_secret",
	"auth", "authorization", "private_key", "signing_key",

	// PII
	"ssn", "social_security_number",
	"date_of_birth", "dob", "email",

	// Financial
	"credit_card", "card_number", "cvv", "cvc",
	"account_number", "routing_number",
}

// MaskPattern pairs a compiled regex with a replacement function. Every match
// of Regex within a string field value is passed to Replace; the return value
// is substituted in place.
type MaskPattern struct {
	re      *regexp.Regexp
	replace func(match string) string
	minLen  int // if > 0, skip regex for strings shorter than this
}

// NewMaskPattern creates a MaskPattern.
func NewMaskPattern(re *regexp.Regexp, replace func(match string) string) MaskPattern {
	return MaskPattern{re: re, replace: replace}
}

// PartialRedact returns a replacement function that preserves the first and
// last reveal characters of strings longer than 2*reveal, replacing the
// middle with mask. Shorter strings are fully replaced.
func PartialRedact(reveal int, mask string) func(string) string {
	return func(s string) string {
		if len(s) <= reveal*2 {
			return mask
		}
		return s[:reveal] + mask + s[len(s)-reveal:]
	}
}

var defaultRedact = PartialRedact(3, "xxxxxx")

// defaultPatterns is the set of regex patterns applied to string field values
// when no WithPatterns option is supplied.
var defaultPatterns = []MaskPattern{
	bearerTokenPattern(),
	apiKeyPattern(),
}

func bearerTokenPattern() MaskPattern {
	re := regexp.MustCompile(`(?i)Authorization:\s*Bearer\s+\S+`)
	return MaskPattern{
		re:     re,
		minLen: 21, // len("Authorization:Bearer X")
		replace: func(match string) string {
			lower := strings.ToLower(match)
			idx := strings.Index(lower, "bearer ")
			if idx < 0 {
				return match
			}
			tokenStart := idx + 7
			prefix := match[:tokenStart]
			token := match[tokenStart:]
			return prefix + defaultRedact(token)
		},
	}
}

func apiKeyPattern() MaskPattern {
	re := regexp.MustCompile(`(?im)^X-API-Key:\s+\S+`)
	return MaskPattern{
		re:     re,
		minLen: 12, // len("X-API-Key: X")
		replace: func(match string) string {
			idx := strings.Index(match, ": ")
			if idx < 0 {
				return match
			}
			keyStart := idx + 2
			prefix := match[:keyStart]
			value := match[keyStart:]
			return prefix + defaultRedact(value)
		},
	}
}

// Masker is a Hook that replaces the values of fields whose keys match a
// configured set with a mask string, and applies regex patterns to the
// string content of remaining fields. Keys are matched case-insensitively.
type Masker struct {
	keys     map[string]struct{}
	maskVal  string
	patterns []MaskPattern
}

// MaskerOption configures a Masker.
type MaskerOption func(*Masker)

// NewMasker creates a Masker that masks defaultSensitiveKeys with "***"
// and applies defaultPatterns to string field values. Use options to customize.
func NewMasker(opts ...MaskerOption) *Masker {
	m := &Masker{
		keys:     make(map[string]struct{}, len(defaultSensitiveKeys)),
		maskVal:  DefaultMaskValue,
		patterns: defaultPatterns,
	}
	for _, k := range defaultSensitiveKeys {
		m.keys[strings.ToLower(k)] = struct{}{}
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WithKeys replaces the entire set of sensitive keys.
func WithKeys(keys []string) MaskerOption {
	return func(m *Masker) {
		m.keys = make(map[string]struct{}, len(keys))
		for _, k := range keys {
			m.keys[strings.ToLower(k)] = struct{}{}
		}
	}
}

// WithExtraKeys adds keys on top of the defaults.
func WithExtraKeys(keys ...string) MaskerOption {
	return func(m *Masker) {
		for _, k := range keys {
			m.keys[strings.ToLower(k)] = struct{}{}
		}
	}
}

// WithMaskValue sets the replacement string for key-based masking. Default is "***".
func WithMaskValue(val string) MaskerOption {
	return func(m *Masker) {
		m.maskVal = val
	}
}

// WithPatterns replaces the entire set of regex patterns.
func WithPatterns(patterns []MaskPattern) MaskerOption {
	return func(m *Masker) {
		m.patterns = patterns
	}
}

// WithExtraPatterns adds patterns on top of the defaults.
func WithExtraPatterns(patterns ...MaskPattern) MaskerOption {
	return func(m *Masker) {
		m.patterns = append(m.patterns, patterns...)
	}
}

// Hook returns a Hook function suitable for WithHook.
//
// For each field the hook:
//  1. If the field key matches a sensitive key, replaces the entire value.
//  2. Otherwise, if the field carries string content, applies regex patterns
//     to redact embedded secrets (e.g. bearer tokens in headers).
func (m *Masker) Hook() Hook {
	return func(fields []Field) []Field {
		for i, f := range fields {
			if _, ok := m.keys[strings.ToLower(f.key)]; ok {
				fields[i] = Str(f.key, m.maskVal)
				continue
			}

			if len(m.patterns) > 0 {
				if str, ok := fieldStr(f); ok {
					masked := str
					for _, p := range m.patterns {
						if len(masked) > p.minLen {
							masked = p.re.ReplaceAllStringFunc(masked, p.replace)
						}
					}

					if masked != str {
						fields[i] = Str(f.key, masked)
					}
				}
			}
		}

		return fields
	}
}

// fieldStr returns the string content of f for pattern matching.
// For fieldVal, the string is computed lazily via fmt.Sprint.
// Returns ("", false) for numeric/bool/time fields.
func fieldStr(f Field) (string, bool) {
	switch f.typ {
	case fieldString, fieldError, fieldBytes:
		return f.str, true
	case fieldVal:
		return fmt.Sprint(f.val), true
	default:
		return "", false
	}
}
