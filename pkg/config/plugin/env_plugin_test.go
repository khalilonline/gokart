package plugin

import (
	"testing"
	"time"

	"github.com/khalilonline/gokart/pkg/logger"
	"github.com/khalilonline/gokart/pkg/testflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type envPluginTestSuite struct {
	suite.Suite
}

func TestEnvPluginSuite(t *testing.T) {
	testflags.UnitTest(t)
	suite.Run(t, new(envPluginTestSuite))
}

type envTestConfig struct {
	Host    string        `env:"TEST_HOST"`
	Port    int           `env:"TEST_PORT" envDefault:"8080"`
	Debug   bool          `env:"TEST_DEBUG"`
	Rate    float64       `env:"TEST_RATE"`
	Timeout time.Duration `env:"TEST_TIMEOUT"`
	Unset   string
}

func (s *envPluginTestSuite) TestAllFields() {
	t := s.T()
	t.Setenv("TEST_HOST", "localhost")
	t.Setenv("TEST_PORT", "3000")
	t.Setenv("TEST_DEBUG", "true")
	t.Setenv("TEST_RATE", "1.5")
	t.Setenv("TEST_TIMEOUT", "5s")

	var cfg envTestConfig
	err := NewEnvPlugin().Load(&cfg)

	assert.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 3000, cfg.Port)
	assert.True(t, cfg.Debug)
	assert.Equal(t, 1.5, cfg.Rate)
	assert.Equal(t, 5*time.Second, cfg.Timeout)
	assert.Empty(t, cfg.Unset)
}

func (s *envPluginTestSuite) TestDefaults() {
	t := s.T()
	t.Setenv("TEST_HOST", "localhost")

	var cfg envTestConfig
	err := NewEnvPlugin().Load(&cfg)

	assert.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port, "should use envDefault when env var is unset")
}

func (s *envPluginTestSuite) TestEmptyEnv() {
	t := s.T()

	var cfg envTestConfig
	err := NewEnvPlugin().Load(&cfg)

	assert.NoError(t, err)
	assert.Empty(t, cfg.Host)
	assert.Equal(t, 8080, cfg.Port, "should use envDefault")
	assert.False(t, cfg.Debug)
	assert.Equal(t, 0.0, cfg.Rate)
}

func (s *envPluginTestSuite) TestInvalidInt() {
	t := s.T()
	t.Setenv("TEST_PORT", "abc")

	var cfg envTestConfig
	err := NewEnvPlugin().Load(&cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_PORT")
}

func (s *envPluginTestSuite) TestTextUnmarshaler() {
	type cfg struct {
		Level logger.Level `env:"TEST_LOG_LEVEL"`
	}

	t := s.T()
	t.Setenv("TEST_LOG_LEVEL", "info")

	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.NoError(t, err)
	assert.Equal(t, logger.INFO, c.Level)
}

func (s *envPluginTestSuite) TestTextUnmarshalerInvalidValue() {
	type cfg struct {
		Level logger.Level `env:"TEST_LOG_LEVEL"`
	}

	t := s.T()
	t.Setenv("TEST_LOG_LEVEL", "invalid")

	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_LOG_LEVEL")
}

func (s *envPluginTestSuite) TestStringSlice() {
	type cfg struct {
		Hosts []string `env:"TEST_HOSTS"`
	}

	t := s.T()
	t.Setenv("TEST_HOSTS", "a.example.com,b.example.com,c.example.com")

	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.NoError(t, err)
	assert.Equal(t, []string{"a.example.com", "b.example.com", "c.example.com"}, c.Hosts)
}

func (s *envPluginTestSuite) TestStringSliceTrimsWhitespaceAndSkipsEmpty() {
	type cfg struct {
		Hosts []string `env:"TEST_HOSTS"`
	}

	t := s.T()
	// Trailing separator + interior whitespace + empty middle entry.
	t.Setenv("TEST_HOSTS", "a.example.com,  b.example.com  ,,c.example.com,")

	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.NoError(t, err)
	assert.Equal(t, []string{"a.example.com", "b.example.com", "c.example.com"}, c.Hosts)
}

func (s *envPluginTestSuite) TestStringSliceCustomSeparator() {
	type cfg struct {
		Tokens []string `env:"TEST_TOKENS" envSeparator:"|"`
	}

	t := s.T()
	t.Setenv("TEST_TOKENS", "alpha|beta|gamma")

	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, c.Tokens)
}

func (s *envPluginTestSuite) TestIntSlice() {
	type cfg struct {
		Ports []int `env:"TEST_PORTS"`
	}

	t := s.T()
	t.Setenv("TEST_PORTS", "8080,8081,8082")

	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.NoError(t, err)
	assert.Equal(t, []int{8080, 8081, 8082}, c.Ports)
}

func (s *envPluginTestSuite) TestIntSliceInvalidElement() {
	type cfg struct {
		Ports []int `env:"TEST_PORTS"`
	}

	t := s.T()
	t.Setenv("TEST_PORTS", "8080,not-a-number,8082")

	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_PORTS")
	assert.Contains(t, err.Error(), "not-a-number")
}

func (s *envPluginTestSuite) TestSliceFromDefault() {
	type cfg struct {
		Hosts []string `env:"TEST_HOSTS_UNSET" envDefault:"x.example.com,y.example.com"`
	}

	t := s.T()
	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.NoError(t, err)
	assert.Equal(t, []string{"x.example.com", "y.example.com"}, c.Hosts)
}

func (s *envPluginTestSuite) TestEmptySliceEnvLeavesFieldZero() {
	type cfg struct {
		Hosts []string `env:"TEST_HOSTS_EMPTY"`
	}

	// Empty/unset env value → field stays nil. Same behaviour as the
	// scalar empty-env case.
	t := s.T()
	var c cfg
	err := NewEnvPlugin().Load(&c)

	assert.NoError(t, err)
	assert.Nil(t, c.Hosts)
}

func (s *envPluginTestSuite) TestNestedStruct() {
	type inner struct {
		Value string `env:"TEST_INNER_VAL"`
	}
	type outer struct {
		Name  string `env:"TEST_OUTER_NAME"`
		Inner inner
	}

	t := s.T()
	t.Setenv("TEST_OUTER_NAME", "outer")
	t.Setenv("TEST_INNER_VAL", "inner")

	var cfg outer
	err := NewEnvPlugin().Load(&cfg)

	assert.NoError(t, err)
	assert.Equal(t, "outer", cfg.Name)
	assert.Equal(t, "inner", cfg.Inner.Value)
}
