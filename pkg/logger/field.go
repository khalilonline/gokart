package logger

import (
	"math"
	"time"
	"unsafe"
)

type fieldType uint8

const (
	fieldString   fieldType = iota
	fieldInt                // int, int64
	fieldFloat              // float64
	fieldBool               // bool (stored as 0/1 in num)
	fieldError              // error message string
	fieldTime               // time as unix nanos + layout
	fieldDuration           // duration as nanoseconds
	fieldBytes              // byte slice stored as string
	fieldVal                // any value, raw — preserved for type-asserting consumers
)

// Well-known field key constants used throughout the application.
// Centralising these here avoids duplicated string literals and ensures
// all producers and consumers refer to the same key names.
const (
	ErrorKey          = "error"
	CallerKey         = "caller"
	TimestampKey      = "timestamp"
	RequestIDKey      = "request_id"
	Requester         = "requester"
	EnvKey            = "env"
	ServiceNameKey    = "service_name"
	ServiceVersionKey = "service_version"
	SessionIDKey      = "session_id"
	TraceIDKey        = "trace_id"
	SpanIDKey         = "span_id"
)

// Field is a typed key-value pair for structured logging.
// It is designed as a value type (64 bytes) that stays on the stack for
// all types except fieldVal, which stores a pointer via the val interface.
// The num field is a union: it stores int64, bool (0/1), duration (ns),
// time unix nanos, or float64 bits depending on typ.
type Field struct {
	key string    // 16 bytes
	typ fieldType // 1 byte + 7 padding
	str string    // 16 bytes — string values, error messages, time layouts
	num int64     // 8 bytes — int, bool, duration, time nanos, or float64 bits
	val any       // 16 bytes — raw value for fieldVal; nil for all other types
}

// Key returns the field's key name.
func (f Field) Key() string {
	return f.key
}

// Value returns the field's value as the closest Go type.
func (f Field) Value() any {
	switch f.typ {
	case fieldString, fieldError:
		return f.str
	case fieldVal:
		return f.val
	case fieldInt:
		return f.num
	case fieldFloat:
		return math.Float64frombits(uint64(f.num))
	case fieldBool:
		return f.num != 0
	case fieldTime:
		return time.Unix(0, f.num)
	case fieldDuration:
		return time.Duration(f.num)
	case fieldBytes:
		return []byte(f.str)
	default:
		return nil
	}
}

// Str creates a string field.
func Str(key, val string) Field {
	return Field{key: key, typ: fieldString, str: val}
}

// Int creates an int field.
func Int(key string, val int) Field {
	return Field{key: key, typ: fieldInt, num: int64(val)}
}

// Int64 creates an int64 field.
func Int64(key string, val int64) Field {
	return Field{key: key, typ: fieldInt, num: val}
}

// Float64 creates a float64 field. The value is stored as int64 bits.
func Float64(key string, val float64) Field {
	return Field{key: key, typ: fieldFloat, num: int64(math.Float64bits(val))}
}

// Bool creates a bool field.
func Bool(key string, val bool) Field {
	var n int64
	if val {
		n = 1
	}
	return Field{key: key, typ: fieldBool, num: n}
}

// Err creates an error field with the key "error".
// If err is nil, the value is "<nil>".
func Err(err error) Field {
	msg := "<nil>"
	if err != nil {
		msg = err.Error()
	}

	return Field{key: ErrorKey, typ: fieldError, str: msg}
}

// Time creates a time field. The time is stored as UnixNano in num and formatted
// lazily during encoding using the layout stored in str.
func Time(key string, val time.Time, layout string) Field {
	return Field{key: key, typ: fieldTime, num: val.UnixNano(), str: layout}
}

// Dur creates a duration field stored as nanoseconds.
func Dur(key string, val time.Duration) Field {
	return Field{key: key, typ: fieldDuration, num: int64(val)}
}

// Bytes creates a field from a byte slice. Uses unsafe zero-copy conversion
// to string to avoid allocation. The caller must not modify val after this call.
func Bytes(key string, val []byte) Field {
	return Field{key: key, typ: fieldBytes, str: unsafe.String(unsafe.SliceData(val), len(val))}
}

// Val creates a field that preserves the raw value without converting to a
// string. Consumers can recover the original value via f.Value().(T).
// For JSON/stdout encoding, fmt.Sprint is used, which calls .String() on
// types that implement fmt.Stringer (e.g. trace.TraceID, trace.SpanID).
func Val(key string, val any) Field {
	return Field{key: key, typ: fieldVal, val: val}
}

// Any creates a field from an arbitrary value. It uses a type switch to
// select the most efficient typed encoding, falling back to fmt.Sprint only
// for types that have no dedicated Field constructor.
func Any(key string, val any) Field {
	switch v := val.(type) {
	case string:
		return Str(key, v)
	case int:
		return Int(key, v)
	case int64:
		return Int64(key, v)
	case float64:
		return Float64(key, v)
	case float32:
		return Float64(key, float64(v))
	case bool:
		return Bool(key, v)
	case time.Duration:
		return Dur(key, v)
	case error:
		if v == nil {
			return Str(key, "<nil>")
		}
		return Str(key, v.Error())
	case []byte:
		return Bytes(key, v)
	default:
		return Val(key, val)
	}
}
