package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/khalilonline/gokart/pkg/testflags"
)

// helper to capture logger output.
func capture(levels Level, opts ...Option) (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := New(&buf, levels, opts...)
	return l, &buf
}

// --- Level gating ---

func TestDisabledLevel_NoOutput(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(ERROR) // only ERROR enabled
	l.Info("should not appear")
	if buf.Len() != 0 {
		t.Fatalf("disabled level produced output: %s", buf.String())
	}
}

func TestEnabledLevel_WritesJSON(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	l.Info("hello")
	if buf.Len() == 0 {
		t.Fatal("enabled level produced no output")
	}
	if !json.Valid(buf.Bytes()) {
		t.Fatalf("invalid JSON: %s", buf.String())
	}
}

// --- Output correctness ---

func TestSimpleMessage(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	l.Info("server started")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if m["level"] != "info" {
		t.Fatalf("level = %v, want info", m["level"])
	}
	if m["message"] != "server started" {
		t.Fatalf("message = %v, want server started", m["message"])
	}
}

func TestAllFieldTypes(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	l.Info("test",
		Str("s", "val"),
		Int("i", 42),
		Int64("i64", 99),
		Float64("f", 3.14),
		Bool("b", true),
		Err(errors.New("boom")),
		Time("ts", ts, time.RFC3339),
		Dur("dur", 5*time.Second),
		Bytes("data", []byte("raw")),
		Any("any", []int{1, 2}),
	)

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if m["s"] != "val" {
		t.Fatalf("s = %v", m["s"])
	}
	if m["i"] != float64(42) {
		t.Fatalf("i = %v", m["i"])
	}
	if m["i64"] != float64(99) {
		t.Fatalf("i64 = %v", m["i64"])
	}
	if m["f"] != 3.14 {
		t.Fatalf("f = %v", m["f"])
	}
	if m["b"] != true {
		t.Fatalf("b = %v", m["b"])
	}
	if m["error"] != "boom" {
		t.Fatalf("error = %v", m["error"])
	}
	if m["ts"] != "2024-06-15T10:30:00Z" {
		t.Fatalf("ts = %v", m["ts"])
	}
	if m["dur"] != float64(5*time.Second) {
		t.Fatalf("dur = %v", m["dur"])
	}
	if m["data"] != "raw" {
		t.Fatalf("data = %v", m["data"])
	}
	if m["any"] != "[1 2]" {
		t.Fatalf("any = %v", m["any"])
	}
}

func TestNilError(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	l.Info("test", Err(nil))

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["error"] != "<nil>" {
		t.Fatalf("error = %v, want <nil>", m["error"])
	}
}

func TestEmptyMessage(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	l.Info("")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if m["message"] != "" {
		t.Fatalf("message = %v, want empty", m["message"])
	}
}

func TestMessageWithSpecialChars(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	l.Info("say \"hello\"\nnewline")

	if !json.Valid(buf.Bytes()) {
		t.Fatalf("invalid JSON: %s", buf.String())
	}
}

// --- Options ---

func TestWithTimestamp(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO, WithTimestamp(time.RFC3339))
	l.Info("test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	ts, ok := m["timestamp"].(string)
	if !ok || ts == "" {
		t.Fatal("timestamp missing or empty")
	}
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		t.Fatalf("cannot parse timestamp %q: %v", ts, err)
	}
}

func TestWithTimestampKey(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO, WithTimestamp(time.RFC3339), WithTimestampKey("time"))
	l.Info("test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := m["time"]; !ok {
		t.Fatal("custom timestamp key 'time' not found")
	}
	if _, ok := m["timestamp"]; ok {
		t.Fatal("default timestamp key should not be present")
	}
}

func TestWithCaller(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO, WithCaller())
	l.Info("test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	caller, ok := m["caller"].(string)
	if !ok || caller == "" {
		t.Fatal("caller missing or empty")
	}
	if !strings.Contains(caller, "logger_test.go:") {
		t.Fatalf("caller = %q, should contain logger_test.go:", caller)
	}
}

func TestWithHook(t *testing.T) {
	testflags.UnitTest(t)

	hookCalled := false
	hook := func(fields []Field) []Field {
		hookCalled = true
		return fields
	}

	l, buf := capture(INFO, WithHook(hook))
	l.Info("test")

	if !hookCalled {
		t.Fatal("hook was not called")
	}
	if !json.Valid(buf.Bytes()) {
		t.Fatalf("invalid JSON: %s", buf.String())
	}
}

func TestWithExitFunc(t *testing.T) {
	testflags.UnitTest(t)

	exitCode := -1
	l, _ := capture(FATAL, WithExitFunc(func(code int) {
		exitCode = code
	}))
	l.Fatal("bye")

	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
}

func TestWithErrorKey(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO, WithErrorKey("err"))
	l.Info("test", Err(errors.New("fail")))

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := m["err"]; !ok {
		t.Fatal("custom error key 'err' not found")
	}
	if _, ok := m["error"]; ok {
		t.Fatal("default error key should not be present")
	}
}

// --- Context (With) ---

func TestWith_ChildIncludesParentFields(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	child := l.With(Str("req_id", "abc"))
	child.Info("hello")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["req_id"] != "abc" {
		t.Fatalf("req_id = %v, want abc", m["req_id"])
	}
}

func TestWith_GrandchildChains(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	child := l.With(Str("a", "1"))
	grandchild := child.With(Str("b", "2"))
	grandchild.Info("test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["a"] != "1" {
		t.Fatalf("a = %v, want 1", m["a"])
	}
	if m["b"] != "2" {
		t.Fatalf("b = %v, want 2", m["b"])
	}
}

func TestWith_ChildDoesNotAffectParent(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO)
	_ = l.With(Str("child_only", "yes"))
	l.Info("parent")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := m["child_only"]; ok {
		t.Fatal("parent should not have child's context fields")
	}
}

func TestWith_ErrorKeyPropagated(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(INFO, WithErrorKey("err"))
	child := l.With(Err(errors.New("ctx_err")))
	child.Info("test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := m["err"]; !ok {
		t.Fatal("With() should use custom error key")
	}
}

// --- Concurrency ---

func TestConcurrency(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(ALL)

	var wg sync.WaitGroup
	const goroutines = 100
	const iterations = 50

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				l.Info("concurrent", Str("k", "v"), Int("n", 1))
			}
		}()
	}
	wg.Wait()

	// Each line should be valid JSON.
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != goroutines*iterations {
		t.Fatalf("got %d lines, want %d", len(lines), goroutines*iterations)
	}
	for i, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatalf("line %d is invalid JSON: %s", i, line)
		}
	}
}

// --- JSON validity ---

func TestAllLevels_ValidJSON(t *testing.T) {
	testflags.UnitTest(t)

	levels := []struct {
		name string
		lvl  Level
		fn   func(*Logger, string, ...Field)
	}{
		{"debug", DEBUG, (*Logger).Debug},
		{"info", INFO, (*Logger).Info},
		{"warn", WARN, (*Logger).Warn},
		{"error", ERROR, (*Logger).Error},
	}
	for _, tt := range levels {
		t.Run(tt.name, func(t *testing.T) {
			l, buf := capture(tt.lvl)
			tt.fn(l, "test msg", Str("key", "val"))
			if !json.Valid(buf.Bytes()) {
				t.Fatalf("invalid JSON for %s: %s", tt.name, buf.String())
			}
		})
	}
}

func TestFatal_ValidJSON(t *testing.T) {
	testflags.UnitTest(t)

	l, buf := capture(FATAL, WithExitFunc(func(int) {}))
	l.Fatal("fatal msg")
	if !json.Valid(buf.Bytes()) {
		t.Fatalf("invalid JSON: %s", buf.String())
	}
}

// --- Enabled ---

func TestEnabled(t *testing.T) {
	testflags.UnitTest(t)

	l, _ := capture(INFO | ERROR)
	if !l.Enabled(INFO) {
		t.Fatal("INFO should be enabled")
	}
	if !l.Enabled(ERROR) {
		t.Fatal("ERROR should be enabled")
	}
	if l.Enabled(DEBUG) {
		t.Fatal("DEBUG should not be enabled")
	}
	if l.Enabled(WARN) {
		t.Fatal("WARN should not be enabled")
	}
}

// --- Nop ---

func TestNop(t *testing.T) {
	testflags.UnitTest(t)

	l := Nop()
	// Should not panic and produce no output.
	l.Info("ignored")
	l.Debug("ignored")
	l.Error("ignored")
	if l.Enabled(ALL) {
		t.Fatal("Nop logger should have no levels enabled")
	}
}

// --- Level output correctness ---

func TestLevelStrings(t *testing.T) {
	testflags.UnitTest(t)

	tests := []struct {
		level Level
		want  string
	}{
		{DEBUG, "debug"},
		{INFO, "info"},
		{WARN, "warn"},
		{ERROR, "error"},
		{FATAL, "fatal"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			l, buf := capture(tt.level, WithExitFunc(func(int) {}))
			switch tt.level {
			case DEBUG:
				l.Debug("msg")
			case INFO:
				l.Info("msg")
			case WARN:
				l.Warn("msg")
			case ERROR:
				l.Error("msg")
			case FATAL:
				l.Fatal("msg")
			}

			var m map[string]any
			if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if m["level"] != tt.want {
				t.Fatalf("level = %v, want %s", m["level"], tt.want)
			}
		})
	}
}
