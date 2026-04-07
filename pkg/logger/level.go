package logger

import (
	"fmt"
	"math/bits"
	"strings"
)

// Level is a bitmask type enabling fast bitwise checks for log level gating.
// Multiple levels can be combined with OR: INFO | ERROR.
type Level uint8

const (
	DEBUG Level = 1 << iota // 0x01
	INFO                    // 0x02
	WARN                    // 0x04
	ERROR                   // 0x08
	FATAL                   // 0x10
	ALL   Level = DEBUG | INFO | WARN | ERROR | FATAL
)

const numLevels = 5

// levelStrings maps level index to its string representation.
var levelStrings = [numLevels]string{
	"debug",
	"info",
	"warn",
	"error",
	"fatal",
}

// levelPrefixes contains pre-computed JSON prefixes for each level.
// e.g. {"level":"info","message":"
var levelPrefixes [numLevels][]byte

func init() {
	for i := range numLevels {
		levelPrefixes[i] = []byte(`{"level":"` + levelStrings[i] + `","message":"`)
	}
}

// levelIndex converts a single-bit Level to an array index [0..4]
// using a single CPU instruction (TZCNT/BSF).
func levelIndex(l Level) int {
	return bits.TrailingZeros8(uint8(l))
}

// ParseLevel converts a case-insensitive string to a Level.
// Accepted values: "debug", "info", "warn", "error", "fatal".
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return DEBUG, nil
	case "info":
		return INFO, nil
	case "warn":
		return WARN, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	default:
		return 0, fmt.Errorf("unknown log level: %q", s)
	}
}

// LevelsAbove returns a bitmask containing min and all higher-severity levels.
func LevelsAbove(min Level) Level {
	// For a single-bit level like WARN (0x04), we want WARN | ERROR | FATAL.
	// All bits from min upward within the valid range: mask = ^(min - 1) & ALL.
	return ^(min - 1) & ALL
}

// UnmarshalText implements encoding.TextUnmarshaler, enabling automatic
// parsing from YAML, JSON, and environment variable config loaders.
func (l *Level) UnmarshalText(b []byte) error {
	parsed, err := ParseLevel(string(b))
	if err != nil {
		return err
	}
	*l = parsed
	return nil
}
