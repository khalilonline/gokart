package logger

import (
	"errors"
	"io"
	"testing"

	"github.com/khalilonline/gokart/pkg/testflags"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// --- Our logger benchmarks ---

func BenchmarkLogger_DisabledLevel(b *testing.B) {
	testflags.PerformanceTest(b)

	l := New(io.Discard, ERROR) // DEBUG/INFO/WARN disabled
	for b.Loop() {
		l.Info("skip", Str("k", "v"))
	}
}

func BenchmarkLogger_SimpleMessage(b *testing.B) {
	testflags.PerformanceTest(b)

	l := New(io.Discard, ALL)
	for b.Loop() {
		l.Info("server started")
	}
}

func BenchmarkLogger_WithFields(b *testing.B) {
	testflags.PerformanceTest(b)

	l := New(io.Discard, ALL)
	err := errors.New("test error")
	for b.Loop() {
		l.Info("request handled",
			Str("method", "GET"),
			Int("status", 200),
			Float64("latency_ms", 1.234),
			Bool("cached", true),
			Err(err),
		)
	}
}

func BenchmarkLogger_WithContext(b *testing.B) {
	testflags.PerformanceTest(b)

	l := New(io.Discard, ALL)
	child := l.With(
		Str("service", "api"),
		Str("version", "1.0"),
		Str("request_id", "abc-123"),
	)
	for b.Loop() {
		child.Info("request",
			Str("method", "GET"),
			Int("status", 200),
		)
	}
}

func BenchmarkLogger_Parallel(b *testing.B) {
	testflags.PerformanceTest(b)

	l := New(io.Discard, ALL)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel",
				Str("method", "GET"),
				Int("status", 200),
			)
		}
	})
}

// --- Zerolog benchmarks ---

func newZerolog() zerolog.Logger {
	return zerolog.New(io.Discard)
}

func BenchmarkZerolog_DisabledLevel(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZerolog().Level(zerolog.ErrorLevel)
	for b.Loop() {
		l.Info().Str("k", "v").Msg("skip")
	}
}

func BenchmarkZerolog_SimpleMessage(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZerolog()
	for b.Loop() {
		l.Info().Msg("server started")
	}
}

func BenchmarkZerolog_WithFields(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZerolog()
	err := errors.New("test error")
	for b.Loop() {
		l.Info().
			Str("method", "GET").
			Int("status", 200).
			Float64("latency_ms", 1.234).
			Bool("cached", true).
			Err(err).
			Msg("request handled")
	}
}

func BenchmarkZerolog_WithContext(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZerolog().With().
		Str("service", "api").
		Str("version", "1.0").
		Str("request_id", "abc-123").
		Logger()
	for b.Loop() {
		l.Info().
			Str("method", "GET").
			Int("status", 200).
			Msg("request")
	}
}

func BenchmarkZerolog_Parallel(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZerolog()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info().
				Str("method", "GET").
				Int("status", 200).
				Msg("parallel")
		}
	})
}

// --- Zap benchmarks ---

func newZap() *zap.Logger {
	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zapcore.DebugLevel)
	return zap.New(core)
}

func BenchmarkZap_DisabledLevel(b *testing.B) {
	testflags.PerformanceTest(b)

	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zapcore.ErrorLevel)
	l := zap.New(core)
	for b.Loop() {
		l.Info("skip", zap.String("k", "v"))
	}
}

func BenchmarkZap_SimpleMessage(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZap()
	for b.Loop() {
		l.Info("server started")
	}
}

func BenchmarkZap_WithFields(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZap()
	err := errors.New("test error")
	for b.Loop() {
		l.Info("request handled",
			zap.String("method", "GET"),
			zap.Int("status", 200),
			zap.Float64("latency_ms", 1.234),
			zap.Bool("cached", true),
			zap.Error(err),
		)
	}
}

func BenchmarkZap_WithContext(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZap().With(
		zap.String("service", "api"),
		zap.String("version", "1.0"),
		zap.String("request_id", "abc-123"),
	)
	for b.Loop() {
		l.Info("request",
			zap.String("method", "GET"),
			zap.Int("status", 200),
		)
	}
}

func BenchmarkZap_Parallel(b *testing.B) {
	testflags.PerformanceTest(b)

	l := newZap()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel",
				zap.String("method", "GET"),
				zap.Int("status", 200),
			)
		}
	})
}
