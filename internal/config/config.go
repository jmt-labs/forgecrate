package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var validPermissionModes = []string{"bypass", "plan", "ask", "auto"}

type Config struct {
	Version        string            `yaml:"version"`
	Source         string            `yaml:"source"`
	Ref            string            `yaml:"ref"`
	Profile        string            `yaml:"profile"`
	Flavors        []string          `yaml:"flavors"`
	PermissionMode string            `yaml:"permission_mode,omitempty"`
	DeployedFiles  map[string]string `yaml:"deployed_files,omitempty"`
}

func ValidatePermissionMode(mode string) error {
	for _, m := range validPermissionModes {
		if mode == m {
			return nil
		}
	}
	return fmt.Errorf("ungültiger Modus %q — erlaubt: bypass, plan, ask, auto", mode)
}

func (c Config) Validate() error {
	if c.PermissionMode != "" {
		return ValidatePermissionMode(c.PermissionMode)
	}
	return nil
}

func (c Config) HasFlavor(name string) bool {
	for _, f := range c.Flavors {
		if f == name {
			return true
		}
	}
	return false
}

func Read(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer func() { _ = f.Close() }()

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
		_ = f.Close()
		return err
	}
	if err := enc.Close(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
