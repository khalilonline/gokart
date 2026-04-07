package logger

import "context"

type ctxKey struct{}

// NewCtx returns a new context with the given Logger stored in it.
func NewCtx(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromCtx retrieves the Logger stored in ctx. If none is found, Nop() is returned.
func FromCtx(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return l
	}
	return Nop()
}
