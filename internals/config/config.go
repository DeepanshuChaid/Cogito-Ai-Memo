package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Enabled   bool   `json:"enabled"`
	Intensity string `json:"intensity"`
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cogito", "config.json")
}

func MustLoad() (*Config, error) {
	path := GetConfigPath()

	// TRY TO READ THE FILE IF WE ARE NOT ABLE TO WE RETURN THE DEFAULT CONFIG
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{Enabled: true, Intensity: "full"}, err
	}

	// IF WE ARE NOT ABLE TO NOT DECODE THE CONFIGS WE RETURN DEFAULT
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return &Config{Enabled: true, Intensity: "full"}, err
	}

	return &cfg, nil
}


func Save(cfg *Config) error {
	path := GetConfigPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(cfg, "", "")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
