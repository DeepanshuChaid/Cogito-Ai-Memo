package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Intensity string

const (
	IntensityLite   Intensity = "lite"
	IntensityNormal Intensity = "normal"
	IntensityUltra  Intensity = "ultra"
)

type Config struct {
	Enabled   bool      `json:"enabled"`
	Intensity Intensity `json:"intensity"`
}

func defaultConfig() *Config {
	return &Config{Enabled: true, Intensity: IntensityNormal}
}

func GetConfigPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".cogito", "config.json")
}

func GetConfigPathForDir(dir string) string {
	if dir == "" {
		return GetConfigPath()
	}
	return filepath.Join(dir, ".cogito", "config.json")
}

func Load() (*Config, error) {
	return LoadForDir("")
}

func LoadForDir(dir string) (*Config, error) {
	path := GetConfigPathForDir(dir)

	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig(), err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return defaultConfig(), err
	}

	if !IsValid(cfg.Intensity) {
		cfg.Intensity = IntensityNormal
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	// Save to current directory instead of home
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, ".cogito", "config.json")

	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func IsValid(Intensity Intensity) bool {
	if Intensity != IntensityLite && Intensity != IntensityNormal && Intensity != IntensityUltra {
		return false
	}
	return true
}
