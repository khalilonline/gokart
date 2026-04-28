package plugin

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// EnvPlugin reads environment variables into struct fields tagged with `env:"VAR_NAME"`.
// Supports string, int, int64, bool, float64, time.Duration, and slices of any of
// those (parsed as comma-separated by default; override the separator with
// `envSeparator:"|"` on the field).
//
// An optional `envDefault:"value"` tag provides a fallback when the variable is unset
// or empty. A field whose type implements `encoding.TextUnmarshaler` (e.g. logger.Level)
// is parsed via that interface — this also applies to slice element types.
type EnvPlugin struct{}

// NewEnvPlugin returns an EnvPlugin.
func NewEnvPlugin() *EnvPlugin {
	return &EnvPlugin{}
}

var _ Plugin = (*EnvPlugin)(nil)

// defaultEnvSeparator splits slice-valued env vars when no `envSeparator` tag is set.
const defaultEnvSeparator = ","

// Load implements Plugin.
func (p *EnvPlugin) Load(cfg any) error {
	return loadEnvFields(reflect.ValueOf(cfg).Elem())
}

func loadEnvFields(v reflect.Value) error {
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// Recurse into embedded or nested structs.
		if field.Type.Kind() == reflect.Struct {
			if err := loadEnvFields(fieldVal); err != nil {
				return err
			}
			continue
		}

		envKey := field.Tag.Get("env")
		if envKey == "" {
			continue
		}

		raw := os.Getenv(envKey)
		if raw == "" {
			raw = field.Tag.Get("envDefault")
		}
		if raw == "" {
			continue
		}

		separator := field.Tag.Get("envSeparator")
		if separator == "" {
			separator = defaultEnvSeparator
		}

		if err := setField(fieldVal, raw, separator); err != nil {
			return fmt.Errorf("field %s (env %q): %w", field.Name, envKey, err)
		}
	}

	return nil
}

var textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()

// setField populates field from raw. For slice fields, raw is split by separator
// and each element is parsed individually via setScalarField. Empty elements
// (e.g. a trailing separator) are skipped silently rather than treated as zero
// values, since "a,b," in env-string form almost always represents a list of
// two and not a list of three.
func setField(field reflect.Value, raw, separator string) error {
	if field.Kind() == reflect.Slice {
		return setSliceField(field, raw, separator)
	}

	return setScalarField(field, raw)
}

func setSliceField(field reflect.Value, raw, separator string) error {
	parts := strings.Split(raw, separator)
	out := reflect.MakeSlice(field.Type(), 0, len(parts))
	elemType := field.Type().Elem()

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		elem := reflect.New(elemType).Elem()
		if err := setScalarField(elem, part); err != nil {
			return fmt.Errorf("element %q: %w", part, err)
		}

		out = reflect.Append(out, elem)
	}

	field.Set(out)
	return nil
}

func setScalarField(field reflect.Value, raw string) error {
	// If the field (or a pointer to it) implements encoding.TextUnmarshaler,
	// delegate parsing to it. This handles custom types like logger.Level.
	if field.CanAddr() && field.Addr().Type().Implements(textUnmarshalerType) {
		return field.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(raw))
	}

	// Handle time.Duration specially.
	if field.Type() == reflect.TypeFor[time.Duration]() {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		field.SetInt(int64(d))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("unsupported type %s", field.Type())
	}

	return nil
}
