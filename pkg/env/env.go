// Package env provides environment type detection and resolution for application services.
package env

import (
	"fmt"
	"os"
	"strings"
)

// Env represents an application environment.
type Env string

const (
	Development Env = "development"
	Sandbox     Env = "sandbox"
	Staging     Env = "staging"
	Production  Env = "production"
)

const defaultEnvVar = "ENV"

// String returns the string representation of the environment.
func (e Env) String() string {
	return string(e)
}

// Valid reports whether the environment is one of the known constants.
func (e Env) Valid() bool {
	switch e {
	case Development, Sandbox, Staging, Production:
		return true
	default:
		return false
	}
}

// IsDevelopment reports whether the environment is Development.
func (e Env) IsDevelopment() bool {
	return e == Development
}

// IsSandbox reports whether the environment is Sandbox.
func (e Env) IsSandbox() bool {
	return e == Sandbox
}

// IsStaging reports whether the environment is Staging.
func (e Env) IsStaging() bool {
	return e == Staging
}

// IsProduction reports whether the environment is Production.
func (e Env) IsProduction() bool {
	return e == Production
}

// IsNonDevelopment reports whether the environment is valid and not Development.
func (e Env) IsNonDevelopment() bool {
	return e.Valid() && e != Development
}

// IsNonProduction reports whether the environment is valid and not Production.
func (e Env) IsNonProduction() bool {
	return e.Valid() && e != Production
}

// ResolveEnv reads the environment from an environment variable and returns the resolved Env.
// By default it reads "ENV". An optional envVarName overrides the variable name.
// Panics if the variable is unset, empty, or not a recognized environment.
func ResolveEnv(envVarName ...string) Env {
	name := defaultEnvVar
	if len(envVarName) > 0 {
		name = envVarName[0]
	}

	raw := os.Getenv(name)
	if raw == "" {
		panic(fmt.Sprintf("environment variable %s is not set", name))
	}

	e := Env(strings.ToLower(raw))
	if !e.Valid() {
		panic(fmt.Sprintf("unknown environment: %s", raw))
	}

	return e
}
