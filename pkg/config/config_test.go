package config

import (
	"fmt"
	"testing"

	"github.com/khalilonline/gokart/pkg/config/plugin"
	"github.com/khalilonline/gokart/pkg/testflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type configTestSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	testflags.UnitTest(t)
	suite.Run(t, new(configTestSuite))
}

func (s *configTestSuite) TestNilCfg() {
	err := Load(nil)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "non-nil pointer to a struct")
}

func (s *configTestSuite) TestNonPointer() {
	type cfg struct{ Name string }
	err := Load(cfg{})
	assert.Error(s.T(), err)
}

func (s *configTestSuite) TestNonStruct() {
	n := 42
	err := Load(&n)
	assert.Error(s.T(), err)
}

func (s *configTestSuite) TestDefaultPlugin() {
	type cfg struct {
		Host string `env:"TEST_CFG_HOST"`
		Port int    `env:"TEST_CFG_PORT"`
	}

	t := s.T()
	t.Setenv("TEST_CFG_HOST", "localhost")
	t.Setenv("TEST_CFG_PORT", "9090")

	var c cfg
	err := Load(&c)

	assert.NoError(t, err)
	assert.Equal(t, "localhost", c.Host)
	assert.Equal(t, 9090, c.Port)
}

func (s *configTestSuite) TestMultiplePlugins() {
	type cfg struct {
		Host string `env:"TEST_MULTI_HOST" yaml:"host"`
		Port int    `env:"TEST_MULTI_PORT" yaml:"port"`
	}

	t := s.T()
	t.Setenv("TEST_MULTI_HOST", "from-env")
	t.Setenv("TEST_MULTI_PORT", "1111")

	yamlData := []byte("host: from-yaml\nport: 2222\n")

	var c cfg
	err := Load(&c, plugin.NewEnvPlugin(), plugin.NewYamlPlugin(yamlData))

	assert.NoError(t, err)
	assert.Equal(t, "from-yaml", c.Host, "YAML plugin should override env plugin")
	assert.Equal(t, 2222, c.Port, "YAML plugin should override env plugin")
}

func (s *configTestSuite) TestPluginError() {
	failing := &failPlugin{err: fmt.Errorf("boom")}

	type cfg struct{ Name string }
	var c cfg
	err := Load(&c, failing)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "boom")
}

type failPlugin struct {
	err error
}

func (p *failPlugin) Load(_ any) error {
	return p.err
}
