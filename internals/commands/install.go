package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
)

func Install() {
	fmt.Println("Installing Cogito...")

	execPath, _ := os.Executable()
	cwd, _ := os.Getwd()

	// Create .cogito in current directory, not home
	cogitoDir := filepath.Join(cwd, ".cogito")
	os.MkdirAll(cogitoDir, 0755)

	// Save config to ./cogito/config.json
	defaultCfg := &config.Config{Enabled: true, Intensity: config.IntensityNormal}
	if err := config.Save(defaultCfg); err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
	} else {
		fmt.Printf("✅ Config saved at %s\n", config.GetConfigPath())
	}

	// Create .codex hooks directory
	hooksDir := filepath.Join(cwd, ".codex")
	os.MkdirAll(hooksDir, 0755)

	hooksConfig := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]string{
				{"type": "command", "command": execPath},
			},
		},
	}

	content, _ := json.MarshalIndent(hooksConfig, "", "  ")
	os.WriteFile(filepath.Join(hooksDir, "hooks.json"), content, 0644)
	fmt.Printf("✅ Installed hook at %s\n", execPath)
}
