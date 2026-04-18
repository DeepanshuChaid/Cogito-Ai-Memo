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

	execPath, _ := os.Executable()
	execPath, _ = filepath.EvalSymlinks(execPath)

	cwd, _ := os.Getwd()

	// Create .cogito
	cogitoDir := filepath.Join(cwd, ".cogito")
	os.MkdirAll(cogitoDir, 0755)

	// Create .codex
	hooksDir := filepath.Join(cwd, ".codex")
	os.MkdirAll(hooksDir, 0755)

	// ✅ All 3 Hooks with args
	hooksConfig := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]interface{}{
				{
					"type":    "command",
					"command": execPath,
					"args":    []string{"hook", "session-start"},
				},
			},
			"PostToolUse": []map[string]interface{}{
				{
					"type":    "command",
					"command": execPath,
					"args":    []string{"hook", "tool-use"},
				},
			},
			"SessionEnd": []map[string]interface{}{
				{
					"type":    "command",
					"command": execPath,
					"args":    []string{"hook", "session-end"},
				},
			},
		},
	}

	content, _ := json.MarshalIndent(hooksConfig, "", "  ")
	os.WriteFile(filepath.Join(hooksDir, "hooks.json"), content, 0644)

	fmt.Println("✅ Cogito hooks installed successfully!")
	fmt.Println("📍 Config:", filepath.Join(hooksDir, "hooks.json"))
	fmt.Println("\nRegistered hooks:")
	fmt.Println("  - SessionStart  → cogito hook session-start")
	fmt.Println("  - PostToolUse   → cogito hook tool-use")
	fmt.Println("  - SessionEnd    → cogito hook session-end")
}
