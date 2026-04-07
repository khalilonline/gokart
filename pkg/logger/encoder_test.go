package logger

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/khalilonline/gokart/pkg/testflags"
)

// --- String escaping ---

func TestAppendEscapedString_PlainASCII(t *testing.T) {
	testflags.UnitTest(t)

	dst := appendEscapedString(nil, "hello world")
	if got := string(dst); got != "hello world" {
		t.Fatalf("got %q, want %q", got, "hello world")
	}
}

func TestAppendEscapedString_Quotes(t *testing.T) {
	testflags.UnitTest(t)

	dst := appendEscapedString(nil, `say "hi"`)
	if got := string(dst); got != `say \"hi\"` {
		t.Fatalf("got %q, want %q", got, `say \"hi\"`)
	}
}

func TestAppendEscapedString_Backslash(t *testing.T) {
	testflags.UnitTest(t)

	dst := appendEscapedString(nil, `a\b`)
	if got := string(dst); got != `a\\b` {
		t.Fatalf("got %q, want %q", got, `a\\b`)
	}
}

func TestAppendEscapedString_ControlChars(t *testing.T) {
	testflags.UnitTest(t)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"newline", "a\nb", `a\nb`},
		{"carriage return", "a\rb", `a\rb`},
		{"tab", "a\tb", `a\tb`},
		{"backspace", "a\bb", `a\bb`},
		{"form feed", "a\fb", `a\fb`},
		{"null byte", "a\x00b", `a\u0000b`},
		{"other control", "a\x01b", `a\u0001b`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := appendEscapedString(nil, tt.in)
			if got := string(dst); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppendEscapedString_ValidUTF8(t *testing.T) {
	testflags.UnitTest(t)

	// Valid UTF-8 should pass through unchanged.
	s := "héllo wörld 日本語"
	dst := appendEscapedString(nil, s)
	if got := string(dst); got != s {
		t.Fatalf("got %q, want %q", got, s)
	}
}

func TestAppendEscapedString_InvalidUTF8(t *testing.T) {
	testflags.UnitTest(t)

	s := "a\xfeb"
	dst := appendEscapedString(nil, s)
	if got := string(dst); got != `a\u00feb` {
		t.Fatalf("got %q, want %q", got, `a\u00feb`)
	}
}

func TestAppendEscapedString_Mixed(t *testing.T) {
	testflags.UnitTest(t)

	s := "a\"b\nc\x00d"
	dst := appendEscapedString(nil, s)
	want := `a\"b\nc\u0000d`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppendEscapedString_Empty(t *testing.T) {
	testflags.UnitTest(t)

	dst := appendEscapedString(nil, "")
	if len(dst) != 0 {
		t.Fatalf("got %q, want empty", string(dst))
	}
}

// --- Typed append functions ---

func TestAppendString(t *testing.T) {
	testflags.UnitTest(t)

	dst := appendString(nil, "name", "alice")
	want := `,"name":"alice"`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppendInt(t *testing.T) {
	testflags.UnitTest(t)

	tests := []struct {
		name string
		val  int64
		want string
	}{
		{"positive", 42, `,"count":42`},
		{"negative", -7, `,"count":-7`},
		{"zero", 0, `,"count":0`},
		{"max", math.MaxInt64, `,"count":9223372036854775807`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := appendInt(nil, "count", tt.val)
			if got := string(dst); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppendFloat(t *testing.T) {
	testflags.UnitTest(t)

	tests := []struct {
		name string
		val  float64
		want string
	}{
		{"normal", 3.14, `,"val":3.14`},
		{"zero", 0.0, `,"val":0`},
		{"negative", -2.5, `,"val":-2.5`},
		{"NaN", math.NaN(), `,"val":null`},
		{"Inf", math.Inf(1), `,"val":null`},
		{"NegInf", math.Inf(-1), `,"val":null`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := appendFloat(nil, "val", tt.val)
			if got := string(dst); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppendBool(t *testing.T) {
	testflags.UnitTest(t)

	dst := appendBool(nil, "ok", true)
	if got := string(dst); got != `,"ok":true` {
		t.Fatalf("got %q, want %q", got, `,"ok":true`)
	}
	dst = appendBool(nil, "ok", false)
	if got := string(dst); got != `,"ok":false` {
		t.Fatalf("got %q, want %q", got, `,"ok":false`)
	}
}

func TestAppendTime(t *testing.T) {
	testflags.UnitTest(t)

	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	dst := appendTime(nil, "ts", ts.UnixNano(), time.RFC3339)
	want := `,"ts":"2024-06-15T10:30:00Z"`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppendDuration(t *testing.T) {
	testflags.UnitTest(t)

	dst := appendInt(nil, "dur", int64(5*time.Second))
	want := `,"dur":5000000000`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppendNull(t *testing.T) {
	testflags.UnitTest(t)

	dst := appendNull(nil, "val")
	want := `,"val":null`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// --- appendField dispatch ---

func TestAppendField_AllTypes(t *testing.T) {
	testflags.UnitTest(t)

	tests := []struct {
		name  string
		field Field
		want  string
	}{
		{"string", Str("k", "v"), `,"k":"v"`},
		{"int", Int("k", 42), `,"k":42`},
		{"int64", Int64("k", 99), `,"k":99`},
		{"float64", Float64("k", 1.5), `,"k":1.5`},
		{"bool_true", Bool("k", true), `,"k":true`},
		{"bool_false", Bool("k", false), `,"k":false`},
		{"duration", Dur("k", 3*time.Second), `,"k":3000000000`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := appendField(nil, tt.field)
			if got := string(dst); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppendField_Error(t *testing.T) {
	testflags.UnitTest(t)

	f := Err(nil)
	dst := appendField(nil, f)
	want := `,"error":"<nil>"`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppendField_Time(t *testing.T) {
	testflags.UnitTest(t)

	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	f := Time("at", ts, time.RFC3339)
	dst := appendField(nil, f)
	want := `,"at":"2024-01-01T00:00:00Z"`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppendField_Bytes(t *testing.T) {
	testflags.UnitTest(t)

	f := Bytes("data", []byte("hello"))
	dst := appendField(nil, f)
	want := `,"data":"hello"`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppendField_Any(t *testing.T) {
	testflags.UnitTest(t)

	// Any() for unknown types now goes through fieldVal path (fmt.Sprint).
	f := Any("x", []int{1, 2, 3})
	dst := appendField(nil, f)
	want := `,"x":"[1 2 3]"`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAppendField_Val(t *testing.T) {
	testflags.UnitTest(t)

	type myStringer struct{ s string }
	s := myStringer{s: "hello"}
	// Val stores raw value; encoder uses fmt.Sprint → calls String() if available.
	f := Val("k", s)
	dst := appendField(nil, f)
	want := `,"k":"{hello}"`
	if got := string(dst); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestVal_PreservesRawValue(t *testing.T) {
	testflags.UnitTest(t)

	type myStruct struct{ X int }
	original := myStruct{X: 42}
	f := Val("k", original)

	got, ok := f.Value().(myStruct)
	if !ok {
		t.Fatalf("Value() type = %T, want myStruct", f.Value())
	}
	if got.X != original.X {
		t.Fatalf("Value().X = %d, want %d", got.X, original.X)
	}
}

func TestAny_FallbackPreservesRawValue(t *testing.T) {
	testflags.UnitTest(t)

	type myStruct struct{ X int }
	original := myStruct{X: 7}
	f := Any("k", original)

	got, ok := f.Value().(myStruct)
	if !ok {
		t.Fatalf("Any() fallback Value() type = %T, want myStruct", f.Value())
	}
	if got.X != original.X {
		t.Fatalf("Any() fallback Value().X = %d, want %d", got.X, original.X)
	}
}

func TestFieldKey_Constants(t *testing.T) {
	testflags.UnitTest(t)

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"ErrorKey", ErrorKey, "error"},
		{"CallerKey", CallerKey, "caller"},
		{"TimestampKey", TimestampKey, "timestamp"},
		{"RequestIDKey", RequestIDKey, "request_id"},
		{"EnvKey", EnvKey, "env"},
		{"ServiceNameKey", ServiceNameKey, "service_name"},
		{"ServiceVersionKey", ServiceVersionKey, "service_version"},
		{"TraceIDKey", TraceIDKey, "trace_id"},
		{"SpanIDKey", SpanIDKey, "span_id"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// --- Buffer pool ---

func TestBufPool_Recycle(t *testing.T) {
	testflags.UnitTest(t)

	bp := getBuf()
	*bp = append(*bp, "hello"...)
	putBuf(bp)

	bp2 := getBuf()
	if len(*bp2) != 0 {
		t.Fatal("recycled buffer should be reset to zero length")
	}
	if cap(*bp2) < bufInitCap {
		t.Fatal("recycled buffer should maintain capacity")
	}
	putBuf(bp2)
}

func TestBufPool_OversizeRejected(t *testing.T) {
	testflags.UnitTest(t)

	bp := getBuf()
	// Grow beyond max pool size.
	*bp = make([]byte, 0, bufMaxPool+1)
	putBuf(bp) // should not be returned to pool

	// Get a new buffer — it should have default capacity, not the oversized one.
	bp2 := getBuf()
	if cap(*bp2) > bufMaxPool {
		t.Fatal("oversized buffer should not be returned to pool")
	}
	putBuf(bp2)
}

// --- JSON validity helpers ---

func TestAppendedFieldsProduceValidJSON(t *testing.T) {
	testflags.UnitTest(t)

	// Build a complete JSON object with all field types.
	ts := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	buf := []byte(`{"message":"test"`)
	buf = appendField(buf, Str("s", "hello \"world\""))
	buf = appendField(buf, Int("i", -42))
	buf = appendField(buf, Float64("f", 3.14))
	buf = appendField(buf, Bool("b", true))
	buf = appendField(buf, Err(nil))
	buf = appendField(buf, Time("t", ts, time.RFC3339))
	buf = appendField(buf, Dur("d", time.Second))
	buf = appendField(buf, Bytes("data", []byte("raw")))
	buf = appendField(buf, Any("any", 123))
	buf = appendField(buf, Val("v", "raw"))
	buf = append(buf, '}')

	if !json.Valid(buf) {
		t.Fatalf("invalid JSON: %s", buf)
	}
}
