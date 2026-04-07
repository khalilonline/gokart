package logger

import (
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Logger is a high-performance structured JSON logger.
// All fields are read-only after construction except for the mutex-guarded writer.
type Logger struct {
	w               io.Writer
	levels          Level
	mu              sync.Mutex
	context         []byte // pre-serialized JSON fragment from With()
	timestampLayout string
	timestampKey    string
	addCaller       bool
	callerSkip      int
	hooks           []Hook
	exitFn          func(int)
	errorKey        string
	emitters        []Emitter
	contextFields   []Field // original fields from With(), for emitters
}

// callerSkipBase is the number of frames to skip to reach the caller of
// Debug/Info/Warn/Error/Fatal → log() → runtime.Caller.
const callerSkipBase = 2

// New creates a Logger that writes JSON to w for the given levels.
func New(w io.Writer, levels Level, opts ...Option) *Logger {
	l := &Logger{
		w:            w,
		levels:       levels,
		timestampKey: TimestampKey,
		exitFn:       os.Exit,
		errorKey:     ErrorKey,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Nop returns a Logger that never writes anything. Useful as a default/placeholder.
func Nop() *Logger {
	return &Logger{
		w:      io.Discard,
		levels: 0,
		exitFn: os.Exit,
	}
}

// Enabled reports whether the given level is enabled.
func (l *Logger) Enabled(level Level) bool {
	return l.levels&level != 0
}

// Debug logs at DEBUG level.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(DEBUG, msg, fields)
}

// Info logs at INFO level.
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(INFO, msg, fields)
}

// Warn logs at WARN level.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(WARN, msg, fields)
}

// Error logs at ERROR level.
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(ERROR, msg, fields)
}

// Fatal logs at FATAL level and then calls exitFn(1).
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(FATAL, msg, fields)
	l.exitFn(1)
}

// With returns a child Logger that includes the given fields in every entry.
// The child has its own mutex and pre-serialized context, so it does not
// contend with the parent. Hooks (e.g. masking) are applied before
// serialization so context fields receive the same transforms as per-entry fields.
func (l *Logger) With(fields ...Field) *Logger {
	// Apply hooks so transforms (e.g. masking) cover context fields too.
	for _, h := range l.hooks {
		fields = h(fields)
	}

	// Pre-serialize the fields into a JSON fragment.
	ctx := make([]byte, 0, len(l.context)+len(fields)*32)
	ctx = append(ctx, l.context...)
	for _, f := range fields {
		ctx = l.encodeField(ctx, f)
	}

	// Accumulate context fields for emitters.
	var cf []Field
	if len(l.emitters) > 0 {
		cf = make([]Field, len(l.contextFields)+len(fields))
		copy(cf, l.contextFields)
		copy(cf[len(l.contextFields):], fields)
	}

	return &Logger{
		w:               l.w,
		levels:          l.levels,
		context:         ctx,
		timestampLayout: l.timestampLayout,
		timestampKey:    l.timestampKey,
		addCaller:       l.addCaller,
		callerSkip:      l.callerSkip,
		hooks:           l.hooks,
		exitFn:          l.exitFn,
		errorKey:        l.errorKey,
		emitters:        l.emitters,
		contextFields:   cf,
	}
}

// log is the hot-path method that builds and writes a JSON log entry.
//
//go:noinline
func (l *Logger) log(level Level, msg string, fields []Field) {
	// 1. Bitmask check.
	if l.levels&level == 0 {
		return
	}

	// 2. Acquire buffer from pool.
	bp := getBuf()
	buf := *bp

	// 3. Write pre-computed level prefix: {"level":"xxx","message":"
	idx := levelIndex(level)
	buf = append(buf, levelPrefixes[idx]...)

	// 4. Write escaped message and closing quote.
	buf = appendEscapedString(buf, msg)
	buf = append(buf, '"')

	// 5. Write pre-serialized context fields (from With()).
	buf = append(buf, l.context...)

	// 6. Timestamp.
	if l.timestampLayout != "" {
		buf = appendTime(buf, l.timestampKey, time.Now().UnixNano(), l.timestampLayout)
	}

	// 7. Caller.
	if l.addCaller {
		_, file, line, ok := runtime.Caller(callerSkipBase + l.callerSkip)
		if ok {
			buf = appendKey(buf, CallerKey)
			buf = append(buf, '"')
			buf = appendEscapedString(buf, file)
			buf = append(buf, ':')
			buf = strconv.AppendInt(buf, int64(line), 10)
			buf = append(buf, '"')
		}
	}

	// 8. Run hooks (before encoding so transforms take effect).
	if len(l.hooks) > 0 {
		for _, h := range l.hooks {
			fields = h(fields)
		}
	}

	// 8b. Emit to registered emitters (e.g. OTel log bridge).
	if len(l.emitters) > 0 {
		emitFields := fields
		if len(l.contextFields) > 0 {
			emitFields = make([]Field, 0, len(l.contextFields)+len(fields))
			emitFields = append(emitFields, l.contextFields...)
			emitFields = append(emitFields, fields...)
		}

		for _, e := range l.emitters {
			e(level, msg, emitFields)
		}
	}

	// 9. Encode entry fields.
	for i := range fields {
		buf = l.encodeField(buf, fields[i])
	}

	// 10. Close JSON object + newline.
	buf = append(buf, '}', '\n')

	// 11. Write under lock.
	l.mu.Lock()
	_, _ = l.w.Write(buf)
	l.mu.Unlock()

	// 12. Return buffer to pool.
	*bp = buf
	putBuf(bp)
}

// encodeField appends a field to dst, remapping the error key if needed.
func (l *Logger) encodeField(dst []byte, f Field) []byte {
	if f.typ == fieldError {
		f.key = l.errorKey
	}
	return appendField(dst, f)
}
