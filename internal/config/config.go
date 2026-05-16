package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version       string            `yaml:"version"`
	Source        string            `yaml:"source"`
	Ref           string            `yaml:"ref"`
	Profile       string            `yaml:"profile"`
	Flavors       []string          `yaml:"flavors"`
	DeployedFiles map[string]string `yaml:"deployed_files,omitempty"`
}

func Read(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Write(path string, cfg Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		f.Close()
		return err
	}
	if err := enc.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
