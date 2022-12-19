package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigPath = "appman.yaml"
)

type ApplicationConfig struct {
	Name         string   `yaml:"name"`
	Path         string   `yaml:"path"`
	StartCommand []string `yaml:"startCommand"`
	StopCommand  []string `yaml:"stopCommand"`
	Env          []string `yaml:"env"`
	Port         int      `yaml:"port"`
	URL          string   `yaml:"url"`
}

func Config() (*[]ApplicationConfig, error) {
	var err error
	config, err := LoadConfig("")
	if err != nil {
		return nil, err
	}

	return config, nil
}

func LoadConfig(path string) (*[]ApplicationConfig, error) {
	if path == "" {
		path = DefaultConfigPath
	}

	contents, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	conf, err := ParseConfig(contents)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file at %q: %w; ensure that the file is valid; Ansible Vault is known to conflict with it", path, err)
	}

	return conf, nil
}

func ParseConfig(contents []byte) (*[]ApplicationConfig, error) {
	var applicationConfig []ApplicationConfig
	err := yaml.Unmarshal(contents, &applicationConfig)

	return &applicationConfig, err
}
