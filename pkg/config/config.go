// Package config provides a plugin-based configuration loader for populating structs
// from environment variables, YAML, or custom sources.
package config

import (
	"fmt"
	"reflect"

	"github.com/khalilonline/gokart/pkg/config/plugin"
)

// Load populates cfg by applying each plugin in order.
// cfg must be a non-nil pointer to a struct.
// If no plugins are provided, it defaults to loading from environment variables.
// Later plugins override earlier ones for overlapping fields.
func Load(cfg any, plugins ...plugin.Plugin) error {
	rv := reflect.ValueOf(cfg)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config: cfg must be a non-nil pointer to a struct, got %T", cfg)
	}

	if len(plugins) == 0 {
		plugins = []plugin.Plugin{plugin.NewEnvPlugin()}
	}

	for _, p := range plugins {
		if err := p.Load(cfg); err != nil {
			return fmt.Errorf("config: plugin failed: %w", err)
		}
	}

	return nil
}
