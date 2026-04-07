package env

import (
	"testing"

	"github.com/khalilonline/gokart/pkg/testflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
}

func TestEnvSuite(t *testing.T) {
	testflags.UnitTest(t)
	suite.Run(t, new(envTestSuite))
}

func (s *envTestSuite) TestString() {
	tests := []struct {
		input    Env
		expected string
	}{
		{Development, "development"},
		{Sandbox, "sandbox"},
		{Staging, "staging"},
		{Production, "production"},
	}

	for _, tt := range tests {
		assert.Equal(s.T(), tt.expected, tt.input.String())
	}
}

func (s *envTestSuite) TestValid() {
	tests := []struct {
		input    Env
		expected bool
	}{
		{Development, true},
		{Sandbox, true},
		{Staging, true},
		{Production, true},
		{"", false},
		{"unknown", false},
		{"PRODUCTION", false},
	}

	for _, tt := range tests {
		assert.Equal(s.T(), tt.expected, tt.input.Valid(), "Env(%q).Valid()", tt.input)
	}
}

func (s *envTestSuite) TestIsDevelopment() {
	assert.True(s.T(), Development.IsDevelopment())
	assert.False(s.T(), Sandbox.IsDevelopment())
	assert.False(s.T(), Staging.IsDevelopment())
	assert.False(s.T(), Production.IsDevelopment())
	assert.False(s.T(), Env("unknown").IsDevelopment())
}

func (s *envTestSuite) TestIsSandbox() {
	assert.True(s.T(), Sandbox.IsSandbox())
	assert.False(s.T(), Development.IsSandbox())
	assert.False(s.T(), Staging.IsSandbox())
	assert.False(s.T(), Production.IsSandbox())
	assert.False(s.T(), Env("unknown").IsSandbox())
}

func (s *envTestSuite) TestIsStaging() {
	assert.True(s.T(), Staging.IsStaging())
	assert.False(s.T(), Development.IsStaging())
	assert.False(s.T(), Sandbox.IsStaging())
	assert.False(s.T(), Production.IsStaging())
	assert.False(s.T(), Env("unknown").IsStaging())
}

func (s *envTestSuite) TestIsProduction() {
	assert.True(s.T(), Production.IsProduction())
	assert.False(s.T(), Development.IsProduction())
	assert.False(s.T(), Sandbox.IsProduction())
	assert.False(s.T(), Staging.IsProduction())
	assert.False(s.T(), Env("unknown").IsProduction())
}

func (s *envTestSuite) TestIsNonDevelopment() {
	assert.True(s.T(), Sandbox.IsNonDevelopment())
	assert.True(s.T(), Staging.IsNonDevelopment())
	assert.True(s.T(), Production.IsNonDevelopment())
	assert.False(s.T(), Development.IsNonDevelopment())
	assert.False(s.T(), Env("unknown").IsNonDevelopment(), "invalid env should return false")
}

func (s *envTestSuite) TestIsNonProduction() {
	assert.True(s.T(), Development.IsNonProduction())
	assert.True(s.T(), Sandbox.IsNonProduction())
	assert.True(s.T(), Staging.IsNonProduction())
	assert.False(s.T(), Production.IsNonProduction())
	assert.False(s.T(), Env("unknown").IsNonProduction(), "invalid env should return false")
}

func (s *envTestSuite) TestResolveEnv() {
	s.T().Run("resolves valid env values", func(t *testing.T) {
		tests := []struct {
			value    string
			expected Env
		}{
			{"development", Development},
			{"Development", Development},
			{"DEVELOPMENT", Development},
			{"sandbox", Sandbox},
			{"SANDBOX", Sandbox},
			{"staging", Staging},
			{"STAGING", Staging},
			{"production", Production},
			{"PRODUCTION", Production},
		}

		for _, tt := range tests {
			t.Setenv("ENV", tt.value)
			assert.Equal(t, tt.expected, ResolveEnv(), "ResolveEnv() with ENV=%q", tt.value)
		}
	})

	s.T().Run("panics on empty env var", func(t *testing.T) {
		t.Setenv("ENV", "")
		assert.Panics(t, func() { ResolveEnv() })
	})

	s.T().Run("panics on unset env var", func(t *testing.T) {
		assert.Panics(t, func() { ResolveEnv("UNSET_VAR_FOR_TEST") })
	})

	s.T().Run("panics on invalid value", func(t *testing.T) {
		t.Setenv("ENV", "invalid")
		assert.Panics(t, func() { ResolveEnv() })
	})

	s.T().Run("reads custom env var name", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		assert.Equal(t, Production, ResolveEnv("APP_ENV"))
	})
}
