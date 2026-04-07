package plugin

import "gopkg.in/yaml.v3"

// YamlPlugin unmarshals YAML bytes into a struct using `yaml` struct tags.
type YamlPlugin struct {
	data []byte
}

// NewYamlPlugin returns a YamlPlugin that will unmarshal the given YAML bytes.
func NewYamlPlugin(data []byte) *YamlPlugin {
	return &YamlPlugin{data: data}
}

var _ Plugin = (*YamlPlugin)(nil)

// Load implements Plugin.
func (p *YamlPlugin) Load(cfg any) error {
	if len(p.data) == 0 {
		return nil
	}
	return yaml.Unmarshal(p.data, cfg)
}
