package plugin

import (
	"testing"

	"github.com/khalilonline/gokart/pkg/testflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type yamlPluginTestSuite struct {
	suite.Suite
}

func TestYamlPluginSuite(t *testing.T) {
	testflags.UnitTest(t)
	suite.Run(t, new(yamlPluginTestSuite))
}

func (s *yamlPluginTestSuite) TestValidYaml() {
	type cfg struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	data := []byte("host: localhost\nport: 3000\n")

	var c cfg
	err := NewYamlPlugin(data).Load(&c)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "localhost", c.Host)
	assert.Equal(s.T(), 3000, c.Port)
}

func (s *yamlPluginTestSuite) TestInvalidYaml() {
	type cfg struct {
		Host string `yaml:"host"`
	}

	data := []byte(":\n  :\n    - invalid: [")

	var c cfg
	err := NewYamlPlugin(data).Load(&c)

	assert.Error(s.T(), err)
}

func (s *yamlPluginTestSuite) TestEmptyBytes() {
	type cfg struct {
		Host string `yaml:"host"`
	}

	var c cfg
	err := NewYamlPlugin(nil).Load(&c)

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), c.Host)
}

func (s *yamlPluginTestSuite) TestNestedStruct() {
	type db struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	type cfg struct {
		DB db `yaml:"db"`
	}

	data := []byte("db:\n  host: localhost\n  port: 5432\n")

	var c cfg
	err := NewYamlPlugin(data).Load(&c)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "localhost", c.DB.Host)
	assert.Equal(s.T(), 5432, c.DB.Port)
}
