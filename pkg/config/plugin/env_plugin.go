package plugin

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

// EnvPlugin reads environment variables into struct fields tagged with `env:"VAR_NAME"`.
// Supports string, int, int64, bool, float64, and time.Duration fields.
// An optional `envDefault:"value"` tag provides a fallback when the variable is unset or empty.
type EnvPlugin struct{}

// NewEnvPlugin returns an EnvPlugin.
func NewEnvPlugin() *EnvPlugin {
	return &EnvPlugin{}
}

var _ Plugin = (*EnvPlugin)(nil)

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

		if err := setField(fieldVal, raw); err != nil {
			return fmt.Errorf("field %s (env %q): %w", field.Name, envKey, err)
		}
	}

	return nil
}

var textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()

func setField(field reflect.Value, raw string) error {
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
