package logger

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	bufInitCap = 512
	bufMaxPool = 64 * 1024 // buffers larger than this are not returned to the pool
)

// bufPool is a pool of *[]byte to avoid interface boxing allocation.
var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, bufInitCap)
		return &b
	},
}

func getBuf() *[]byte {
	bp := bufPool.Get().(*[]byte)
	*bp = (*bp)[:0]
	return bp
}

func putBuf(bp *[]byte) {
	if cap(*bp) > bufMaxPool {
		return // let GC collect oversized buffers
	}
	bufPool.Put(bp)
}

// noEscapeTable marks bytes that do NOT need JSON escaping.
// true = safe to append directly.
var noEscapeTable [256]bool

func init() {
	for i := range 256 {
		// Only safe printable ASCII: 0x20–0x7E excluding " and \.
		// Bytes >= 0x80 go through slow path for UTF-8 validation.
		noEscapeTable[i] = i >= 0x20 && i <= 0x7E && i != '"' && i != '\\'
	}
}

// appendKey writes ,"key": to dst.
func appendKey(dst []byte, key string) []byte {
	dst = append(dst, ',', '"')
	dst = appendEscapedString(dst, key)
	dst = append(dst, '"', ':')
	return dst
}

// appendString writes ,"key":"escaped_val" to dst.
func appendString(dst []byte, key, val string) []byte {
	dst = appendKey(dst, key)
	dst = append(dst, '"')
	dst = appendEscapedString(dst, val)
	dst = append(dst, '"')
	return dst
}

// appendInt writes ,"key":val to dst.
func appendInt(dst []byte, key string, val int64) []byte {
	dst = appendKey(dst, key)
	dst = strconv.AppendInt(dst, val, 10)
	return dst
}

// appendFloat writes ,"key":val to dst. NaN/Inf are written as null.
func appendFloat(dst []byte, key string, val float64) []byte {
	dst = appendKey(dst, key)
	if math.IsNaN(val) || math.IsInf(val, 0) {
		dst = append(dst, "null"...)
		return dst
	}
	dst = strconv.AppendFloat(dst, val, 'f', -1, 64)
	return dst
}

// appendBool writes ,"key":true or ,"key":false to dst.
func appendBool(dst []byte, key string, val bool) []byte {
	dst = appendKey(dst, key)
	dst = strconv.AppendBool(dst, val)
	return dst
}

// appendTime writes ,"key":"formatted_time" to dst using time.AppendFormat.
func appendTime(dst []byte, key string, unixNano int64, layout string) []byte {
	dst = appendKey(dst, key)
	dst = append(dst, '"')
	t := time.Unix(0, unixNano).UTC()
	dst = t.AppendFormat(dst, layout)
	dst = append(dst, '"')
	return dst
}

// appendNull writes ,"key":null to dst.
func appendNull(dst []byte, key string) []byte {
	dst = appendKey(dst, key)
	dst = append(dst, "null"...)
	return dst
}

// appendField dispatches to the appropriate append function based on field type.
func appendField(dst []byte, f Field) []byte {
	switch f.typ {
	case fieldString, fieldError, fieldBytes:
		return appendString(dst, f.key, f.str)
	case fieldVal:
		return appendString(dst, f.key, fmt.Sprint(f.val))
	case fieldInt, fieldDuration:
		return appendInt(dst, f.key, f.num)
	case fieldFloat:
		return appendFloat(dst, f.key, math.Float64frombits(uint64(f.num)))
	case fieldBool:
		return appendBool(dst, f.key, f.num != 0)
	case fieldTime:
		return appendTime(dst, f.key, f.num, f.str)
	default:
		return appendNull(dst, f.key)
	}
}

// appendEscapedString appends s to dst with JSON escaping.
// Fast path: if no bytes need escaping, copies directly.
func appendEscapedString(dst []byte, s string) []byte {
	// Fast path: scan for any byte that needs escaping.
	for i := range len(s) {
		if !noEscapeTable[s[i]] {
			return appendEscapedStringComplex(dst, s, i)
		}
	}
	return append(dst, s...)
}

// appendEscapedStringComplex handles the slow path for strings with escape characters.
// start is the index of the first byte that needs escaping.
func appendEscapedStringComplex(dst []byte, s string, start int) []byte {
	// Append the clean prefix.
	dst = append(dst, s[:start]...)

	for i := start; i < len(s); {
		b := s[i]
		if noEscapeTable[b] {
			// Find the next byte needing escape.
			j := i + 1
			for j < len(s) && noEscapeTable[s[j]] {
				j++
			}
			dst = append(dst, s[i:j]...)
			i = j
			continue
		}

		switch b {
		case '"':
			dst = append(dst, '\\', '"')
			i++
		case '\\':
			dst = append(dst, '\\', '\\')
			i++
		case '\n':
			dst = append(dst, '\\', 'n')
			i++
		case '\r':
			dst = append(dst, '\\', 'r')
			i++
		case '\t':
			dst = append(dst, '\\', 't')
			i++
		case '\b':
			dst = append(dst, '\\', 'b')
			i++
		case '\f':
			dst = append(dst, '\\', 'f')
			i++
		default:
			if b < 0x20 {
				// Control characters as \u00XX.
				dst = append(dst, '\\', 'u', '0', '0')
				dst = append(dst, hexDigits[b>>4], hexDigits[b&0x0f])
				i++
			} else if b >= utf8.RuneSelf {
				// Multi-byte UTF-8: validate and pass through or escape.
				r, size := utf8.DecodeRuneInString(s[i:])
				if r == utf8.RuneError && size == 1 {
					// Invalid UTF-8 byte — escape as \u00XX.
					dst = append(dst, '\\', 'u', '0', '0')
					dst = append(dst, hexDigits[b>>4], hexDigits[b&0x0f])
					i++
				} else {
					dst = append(dst, s[i:i+size]...)
					i += size
				}
			} else {
				dst = append(dst, b)
				i++
			}
		}
	}

	return dst
}

var hexDigits = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}
