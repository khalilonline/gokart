package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/khalilonline/gokart/pkg/testflags"
)

func TestFromCtx_EmptyContext_ReturnsNop(t *testing.T) {
	testflags.UnitTest(t)

	l := FromCtx(context.Background())
	if l == nil {
		t.Fatal("FromCtx returned nil")
	}
	// Nop logger should have no levels enabled.
	if l.Enabled(ALL) {
		t.Fatal("expected Nop logger with no levels enabled")
	}
}

func TestNewCtx_FromCtx_RoundTrip(t *testing.T) {
	testflags.UnitTest(t)

	var buf bytes.Buffer
	original := New(&buf, INFO)
	ctx := NewCtx(context.Background(), original)
	got := FromCtx(ctx)

	if got != original {
		t.Fatal("FromCtx did not return the same logger instance stored by NewCtx")
	}
}

func TestChildLoggerFromContext_IncludesFields(t *testing.T) {
	testflags.UnitTest(t)

	var buf bytes.Buffer
	base := New(&buf, INFO)
	child := base.With(Str("request_id", "abc-123"))

	ctx := NewCtx(context.Background(), child)
	FromCtx(ctx).Info("hello")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if m["request_id"] != "abc-123" {
		t.Fatalf("request_id = %v, want abc-123", m["request_id"])
	}
	if m["message"] != "hello" {
		t.Fatalf("message = %v, want hello", m["message"])
	}
}
