package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/configTui"
)

func Install() {
	configTui.RunConfigTUI()
	fmt.Println("Installing Cogito...")

	execPath, _ := os.Executable()
	cwd, _ := os.Getwd()

	// Create .cogito in current directory, not home
	cogitoDir := filepath.Join(cwd, ".cogito")
	os.MkdirAll(cogitoDir, 0755)


	// Create .codex hooks directory
	hooksDir := filepath.Join(cwd, ".codex")
	os.MkdirAll(hooksDir, 0755)

	relPath, _ := filepath.Rel(cwd, execPath)

	hooksConfig := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]string{
				{"type": "command", "command": relPath},  // Use relative
			},
		},
	}

	content, _ := json.MarshalIndent(hooksConfig, "", "  ")
	os.WriteFile(filepath.Join(hooksDir, "hooks.json"), content, 0644)
	fmt.Printf("✅ Installed hook at %s\n", execPath)
}

