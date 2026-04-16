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
	Enabled   bool   `json:"enabled"`
	Intensity Intensity `json:"intensity"`
}

func GetConfigPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".cogito", "config.json")
}


func Load() (*Config, error) {
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
