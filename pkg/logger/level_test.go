package logger

import (
	"testing"

	"github.com/khalilonline/gokart/pkg/testflags"
)

func TestParseLevel(t *testing.T) {
	testflags.UnitTest(t)

	tests := []struct {
		input   string
		want    Level
		wantErr bool
	}{
		{"debug", DEBUG, false},
		{"info", INFO, false},
		{"warn", WARN, false},
		{"error", ERROR, false},
		{"fatal", FATAL, false},
		{"DEBUG", DEBUG, false},
		{"Info", INFO, false},
		{"WARN", WARN, false},
		{"Error", ERROR, false},
		{"Fatal", FATAL, false},
		{"unknown", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseLevel(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseLevel(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseLevel(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestLevelsAbove(t *testing.T) {
	testflags.UnitTest(t)

	tests := []struct {
		name string
		min  Level
		want Level
	}{
		{"from DEBUG", DEBUG, ALL},
		{"from INFO", INFO, INFO | WARN | ERROR | FATAL},
		{"from WARN", WARN, WARN | ERROR | FATAL},
		{"from ERROR", ERROR, ERROR | FATAL},
		{"from FATAL", FATAL, FATAL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LevelsAbove(tt.min)
			if got != tt.want {
				t.Fatalf("LevelsAbove(%d) = 0b%08b, want 0b%08b", tt.min, got, tt.want)
			}
		})
	}
}

func TestUnmarshalText(t *testing.T) {
	testflags.UnitTest(t)

	var l Level
	if err := l.UnmarshalText([]byte("warn")); err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}
	if l != WARN {
		t.Fatalf("got %d, want WARN (%d)", l, WARN)
	}

	if err := l.UnmarshalText([]byte("invalid")); err == nil {
		t.Fatal("expected error for invalid level")
	}
}
