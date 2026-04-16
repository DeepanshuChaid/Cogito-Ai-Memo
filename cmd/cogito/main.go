package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			runInstall()
			return
		case "config":
			handleConfig() // Now we handle the config command
			return
		case "compress":
			fmt.Println("Janitor is coming soon... (Step 3)")
			return
		}
	}

	// REMOVED: fmt.Println("ITS WORKING") <- This would break the hook!

	handleHook()
}

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: cogito config <enabled|intensity> <value>")
		return
	}

	cfg, _ := config.Load() // Loads current or creates defaults
	key := os.Args[2]
	val := os.Args[3]

	switch key {
	case "enabled":
		cfg.Enabled = (val == "on" || val == "true")
	case "intensity":
		cfg.Intensity = val
	default:
		fmt.Println("Unknown setting. Use 'enabled' or 'intensity'")
		return
	}

	config.Save(cfg)
	fmt.Printf("✅ Updated %s to %s\n", key, val)
}

func handleHook() {
	var input struct {
		CWD string `json:"cwd"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return
	}

	// 1. LOAD CONFIG
	// config.Load() automatically handles: "If file doesn't exist, return defaults"
	cfg, err := config.Load()
	if err != nil {
		return
	}

	// If user disabled Cogito in config, exit silently so Codex runs normally
	if !cfg.Enabled {
		return
	}

	// 2. Get all compressed memories from SQLite
	memoriesRaw, _ := db.GetAllMemories()
	var memTexts []string
	for _, m := range memoriesRaw {
		memTexts = append(memTexts, fmt.Sprintf("%s: %s", m.FilePath, m.CompressedText))
	}

	// 3. INJECT (Passing the config object as the 3rd argument)
	context := injector.BuildFinalPrompt("Start session", memTexts, cfg)

	output := map[string]interface{}{
		"continue":      true,
		"suppressOutput": true,
		"systemMessage":  context,
	}

	jsonOut, _ := json.Marshal(output)
	fmt.Println(string(jsonOut))
}

func runInstall() {
	fmt.Println("Installing Cogito...")

	execPath, _ := os.Executable()
	cwd, _ := os.Getwd()

	// 1. Create .cogito folder for configs (Use MkdirAll to prevent error if it exists)
	cogitoDir := filepath.Join(os.Getenv("USERPROFILE"), ".cogito") // Use User Profile for Windows
	os.MkdirAll(cogitoDir, 0755)

	// 2. Initialize default config file immediately upon installation
	defaultCfg := &config.Config{Enabled: true, Intensity: "full"}
	config.Save(defaultCfg)

	// 3. Install Codex Hooks
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
