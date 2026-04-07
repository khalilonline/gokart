package logger

// Option configures a Logger.
type Option func(*Logger)

// Hook is a function that can inspect and transform fields before they are written.
type Hook func([]Field) []Field

// WithTimestamp adds a timestamp field to every log entry using the given layout.
func WithTimestamp(layout string) Option {
	return func(l *Logger) {
		l.timestampLayout = layout
	}
}

// WithTimestampKey sets the JSON key for the timestamp field. Default is "timestamp".
func WithTimestampKey(key string) Option {
	return func(l *Logger) {
		l.timestampKey = key
	}
}

// WithCaller adds a "caller" field with file:line to every log entry.
func WithCaller() Option {
	return func(l *Logger) {
		l.addCaller = true
	}
}

// WithCallerSkip adjusts the number of stack frames to skip when determining
// the caller. Use this when wrapping the logger.
func WithCallerSkip(n int) Option {
	return func(l *Logger) {
		l.callerSkip = n
	}
}

// WithHook registers a hook that runs before each log entry is written.
func WithHook(h Hook) Option {
	return func(l *Logger) {
		l.hooks = append(l.hooks, h)
	}
}

// WithExitFunc overrides the function called by Fatal. Default is os.Exit.
func WithExitFunc(fn func(int)) Option {
	return func(l *Logger) {
		l.exitFn = fn
	}
}

// WithErrorKey sets the JSON key used by the Err() field constructor.
// Default is "error".
func WithErrorKey(key string) Option {
	return func(l *Logger) {
		l.errorKey = key
	}
}

// Emitter is a callback invoked on every log entry with the structured data.
// Implementations must be safe for concurrent use.
type Emitter func(level Level, msg string, fields []Field)

// WithEmitter registers an emitter that receives every log entry.
func WithEmitter(e Emitter) Option {
	return func(l *Logger) {
		l.emitters = append(l.emitters, e)
	}
}
