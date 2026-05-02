package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Paths  []string `json:"paths"`
	DryRun bool     `json:"dry_run"`
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".goclean.json"
	}
	return filepath.Join(home, ".goclean.json")
}

func Load() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return Config{}
	}
	var cfg Config
	json.Unmarshal(data, &cfg)
	return cfg
}

func Save(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

func (c *Config) AddPath(p string) {
	for _, existing := range c.Paths {
		if existing == p {
			return
		}
	}
	c.Paths = append(c.Paths, p)
}

func (c *Config) RemovePath(p string) {
	var filtered []string
	for _, existing := range c.Paths {
		if existing != p {
			filtered = append(filtered, existing)
		}
	}
	c.Paths = filtered
}
