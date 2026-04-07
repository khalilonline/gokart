package logger

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestMaskerDefaultKeys(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Str("password", "hunter2"),
		Str("access_token", "tok_abc123"),
		Str("ssn", "123-45-6789"),
		Str("credit_card", "4111111111111111"),
		Str("username", "alice"),
	}
	out := hook(fields)

	for _, f := range out {
		switch f.key {
		case "username":
			if f.str == DefaultMaskValue {
				t.Errorf("non-sensitive key %q should not be masked", f.key)
			}
		default:
			if f.str != DefaultMaskValue {
				t.Errorf("sensitive key %q should be masked, got %q", f.key, f.str)
			}
		}
	}
}

func TestMaskerCaseInsensitive(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Str("Password", "secret"),
		Str("ACCESS_TOKEN", "tok"),
		Str("Api_Key", "key"),
	}
	out := hook(fields)

	for _, f := range out {
		if f.str != DefaultMaskValue {
			t.Errorf("key %q should be masked case-insensitively, got %q", f.key, f.str)
		}
	}
}

func TestMaskerPreservesKey(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{Str("Password", "secret")}
	out := hook(fields)

	if out[0].key != "Password" {
		t.Errorf("original key should be preserved, got %q", out[0].key)
	}
}

func TestMaskerMasksAllFieldTypes(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Int("token", 42),
		Bool("secret", true),
		Float64("api_key", 3.14),
	}
	out := hook(fields)

	for _, f := range out {
		if f.typ != fieldString || f.str != DefaultMaskValue {
			t.Errorf("field %q (type %d) should be masked to string %q, got type %d str %q",
				f.key, f.typ, DefaultMaskValue, f.typ, f.str)
		}
	}
}

func TestMaskerWithKeys(t *testing.T) {
	hook := NewMasker(WithKeys([]string{"custom_field"})).Hook()
	fields := []Field{
		Str("password", "hunter2"),
		Str("custom_field", "sensitive"),
	}
	out := hook(fields)

	if out[0].str == DefaultMaskValue {
		t.Error("password should NOT be masked when WithKeys replaces defaults")
	}
	if out[1].str != DefaultMaskValue {
		t.Errorf("custom_field should be masked, got %q", out[1].str)
	}
}

func TestMaskerWithExtraKeys(t *testing.T) {
	hook := NewMasker(WithExtraKeys("custom_field")).Hook()
	fields := []Field{
		Str("password", "hunter2"),
		Str("custom_field", "sensitive"),
	}
	out := hook(fields)

	if out[0].str != DefaultMaskValue {
		t.Error("default key should still be masked with WithExtraKeys")
	}
	if out[1].str != DefaultMaskValue {
		t.Errorf("extra key should be masked, got %q", out[1].str)
	}
}

func TestMaskerWithMaskValue(t *testing.T) {
	hook := NewMasker(WithMaskValue("[REDACTED]")).Hook()
	fields := []Field{Str("password", "hunter2")}
	out := hook(fields)

	if out[0].str != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %q", out[0].str)
	}
}

func TestMaskerViaLogOutput(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, ALL, WithHook(NewMasker().Hook()))
	log.Info("test", Str("password", "hunter2"), Str("user", "alice"))

	line := buf.String()
	if strings.Contains(line, "hunter2") {
		t.Errorf("password value should be masked in output: %s", line)
	}
	if !strings.Contains(line, DefaultMaskValue) {
		t.Errorf("output should contain mask value: %s", line)
	}
	if !strings.Contains(line, "alice") {
		t.Errorf("non-sensitive value should be present: %s", line)
	}
}

func TestMaskerViaWith(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, ALL, WithHook(NewMasker().Hook()))
	child := log.With(Str("token", "secret_tok"))
	child.Info("test")

	line := buf.String()
	if strings.Contains(line, "secret_tok") {
		t.Errorf("With() context field should be masked: %s", line)
	}
	if !strings.Contains(line, DefaultMaskValue) {
		t.Errorf("output should contain mask value: %s", line)
	}
}

func TestMaskerEmptyFields(t *testing.T) {
	hook := NewMasker().Hook()
	out := hook(nil)
	if out != nil {
		t.Errorf("nil input should return nil, got %v", out)
	}

	out = hook([]Field{})
	if len(out) != 0 {
		t.Errorf("empty input should return empty, got %v", out)
	}
}

func TestFieldKey(t *testing.T) {
	f := Str("my_key", "val")
	if f.Key() != "my_key" {
		t.Errorf("Key() = %q, want %q", f.Key(), "my_key")
	}
}

// --- Pattern-based masking tests ---

func TestPatternBearerToken_LongToken(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Str("headers", "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.longtoken"),
	}
	out := hook(fields)

	val := out[0].str
	if strings.Contains(val, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("bearer token should be redacted: %s", val)
	}
	// Should preserve first 3 and last 3 chars of token
	if !strings.Contains(val, "eyJ") {
		t.Errorf("should reveal first 3 chars of long token: %s", val)
	}
	if !strings.Contains(val, "ken") {
		t.Errorf("should reveal last 3 chars of long token: %s", val)
	}
	if !strings.Contains(val, "Authorization: Bearer ") {
		t.Errorf("should preserve prefix: %s", val)
	}
}

func TestPatternBearerToken_ShortToken(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Str("headers", "Authorization: Bearer short"),
	}
	out := hook(fields)

	val := out[0].str
	if strings.Contains(val, "short") {
		t.Errorf("short token should be fully masked: %s", val)
	}
	if !strings.Contains(val, "xxxxxx") {
		t.Errorf("should contain mask: %s", val)
	}
}

func TestPatternBearerToken_CaseInsensitive(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Str("raw", "authorization: bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.longtoken"),
	}
	out := hook(fields)

	if strings.Contains(out[0].str, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("case-insensitive bearer should be redacted: %s", out[0].str)
	}
}

func TestPatternAPIKey_LongKey(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Str("headers", "X-API-Key: sk_live_abcdefghijklmnop"),
	}
	out := hook(fields)

	val := out[0].str
	if strings.Contains(val, "abcdefghijklmnop") {
		t.Errorf("API key should be redacted: %s", val)
	}
	if !strings.Contains(val, "sk_") {
		t.Errorf("should reveal first 3 chars: %s", val)
	}
	if !strings.Contains(val, "nop") {
		t.Errorf("should reveal last 3 chars: %s", val)
	}
}

func TestPatternAPIKey_ShortKey(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Str("headers", "X-API-Key: short"),
	}
	out := hook(fields)

	val := out[0].str
	if strings.Contains(val, "short") {
		t.Errorf("short API key should be fully masked: %s", val)
	}
}

func TestPatternSkippedForKeyMaskedFields(t *testing.T) {
	hook := NewMasker().Hook()
	// "authorization" is a sensitive key — should get full key-based masking,
	// not pattern-based partial redaction.
	fields := []Field{
		Str("authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.longtoken"),
	}
	out := hook(fields)

	if out[0].str != DefaultMaskValue {
		t.Errorf("key-matched field should be fully masked, got %q", out[0].str)
	}
}

func TestPatternOnErrorField(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		{key: "error", typ: fieldError, str: "failed: Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.longtoken"},
	}
	out := hook(fields)

	if strings.Contains(out[0].str, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("pattern should apply to error field content: %s", out[0].str)
	}
}

func TestPatternOnAnyField(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Val("debug", "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.longtoken"),
	}
	out := hook(fields)

	if strings.Contains(out[0].str, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("pattern should apply to val-typed field: %s", out[0].str)
	}
}

func TestPatternOnValField(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Val("header", "Authorization: Bearer longtoken1234567890ABCDEF"),
	}
	out := hook(fields)

	val := out[0].str
	if strings.Contains(val, "longtoken1234567890ABCDEF") {
		t.Errorf("bearer token in val field should be redacted: %s", val)
	}
	if !strings.Contains(val, "lon") {
		t.Errorf("should reveal first 3 chars of long token: %s", val)
	}
}

func TestPatternDoesNotApplyToNumericFields(t *testing.T) {
	hook := NewMasker().Hook()
	fields := []Field{
		Int("count", 42),
		Float64("rate", 3.14),
		Bool("active", true),
	}
	out := hook(fields)

	// Numeric fields should pass through unchanged.
	if out[0].typ != fieldInt || out[0].num != 42 {
		t.Errorf("int field should be unchanged")
	}
	if out[1].typ != fieldFloat {
		t.Errorf("float field should be unchanged")
	}
	if out[2].typ != fieldBool {
		t.Errorf("bool field should be unchanged")
	}
}

func TestWithPatterns_ReplacesDefaults(t *testing.T) {
	custom := NewMaskPattern(
		regexp.MustCompile(`secret_\w+`),
		func(match string) string { return "REDACTED" },
	)
	hook := NewMasker(WithPatterns([]MaskPattern{custom})).Hook()

	fields := []Field{
		Str("msg", "got secret_abc123 from api"),
		Str("headers", "Authorization: Bearer longtoken1234567890"),
	}
	out := hook(fields)

	if !strings.Contains(out[0].str, "REDACTED") {
		t.Errorf("custom pattern should match: %s", out[0].str)
	}
	// Default bearer pattern should NOT apply since we replaced all patterns.
	if !strings.Contains(out[1].str, "longtoken1234567890") {
		t.Errorf("default pattern should not apply after WithPatterns: %s", out[1].str)
	}
}

func TestWithExtraPatterns(t *testing.T) {
	custom := NewMaskPattern(
		regexp.MustCompile(`secret_\w+`),
		func(match string) string { return "REDACTED" },
	)
	hook := NewMasker(WithExtraPatterns(custom)).Hook()

	fields := []Field{
		Str("msg", "got secret_abc123 from api"),
		Str("headers", "Authorization: Bearer longtoken1234567890"),
	}
	out := hook(fields)

	if !strings.Contains(out[0].str, "REDACTED") {
		t.Errorf("extra pattern should match: %s", out[0].str)
	}
	// Default bearer pattern should still apply.
	if strings.Contains(out[1].str, "longtoken1234567890") {
		t.Errorf("default pattern should still apply with WithExtraPatterns: %s", out[1].str)
	}
}

func TestPartialRedact(t *testing.T) {
	redact := PartialRedact(3, "***")

	tests := []struct {
		input string
		want  string
	}{
		{"abcdefghijklmnop", "abc***nop"},
		{"abcdef", "***"},        // len == 2*reveal
		{"ab", "***"},            // len < 2*reveal
		{"", "***"},              // empty
		{"abcdefg", "abc***efg"}, // len == 2*reveal + 1
	}
	for _, tt := range tests {
		got := redact(tt.input)
		if got != tt.want {
			t.Errorf("PartialRedact(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPatternViaLogOutput(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, ALL, WithHook(NewMasker().Hook()))
	log.Info("request",
		Str("raw_headers", "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.longtoken"),
		Str("user", "alice"),
	)

	line := buf.String()
	if strings.Contains(line, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("bearer token should be redacted in log output: %s", line)
	}
	if !strings.Contains(line, "alice") {
		t.Errorf("non-sensitive value should be present: %s", line)
	}
}

func TestPatternViaWith(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, ALL, WithHook(NewMasker().Hook()))
	child := log.With(Str("raw_auth", "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.longtoken"))
	child.Info("test")

	line := buf.String()
	if strings.Contains(line, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("pattern masking should apply to With() context fields: %s", line)
	}
}
