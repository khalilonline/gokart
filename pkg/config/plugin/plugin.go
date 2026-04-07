// Package plugin defines the interface for configuration loading plugins.
package plugin

// Plugin loads configuration into a target struct.
type Plugin interface {
	// Load populates the struct pointed to by cfg.
	// cfg is guaranteed to be a non-nil pointer to a struct.
	Load(cfg any) error
}
