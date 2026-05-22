package compose

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProfileConfig struct {
	Extends []string `yaml:"extends"`
}

func LoadProfileConfig(sourceDir, profile string) ProfileConfig {
	path := filepath.Join(sourceDir, "profiles", profile, "profile.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return ProfileConfig{}
	}
	var cfg ProfileConfig
	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}
