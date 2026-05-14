package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string   `yaml:"version"`
	Source  string   `yaml:"source"`
	Ref     string   `yaml:"ref"`
	Profile string   `yaml:"profile"`
	Flavors []string `yaml:"flavors"`
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
	defer f.Close()
	return yaml.NewEncoder(f).Encode(cfg)
}
