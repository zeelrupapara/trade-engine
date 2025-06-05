package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Exchange struct {
	Name   string `yaml:"name"`
	Active bool   `yaml:"active"`
}

type Config struct {
	Exchanges []Exchange `yaml:"exchanges"`
}

func LoadExchangeConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
