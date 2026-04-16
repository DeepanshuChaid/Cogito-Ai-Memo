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
			handleConfig()
			return
		case "compress":
			fmt.Println("Janitor is coming soon... (Step 3)")
			return
		}
	}

	handleHook()
}

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: cogito config <enabled|intensity> <value>")
		return
	}

	cfg, _ := config.MustLoad()
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

	cfg, err := config.MustLoad()
	if err != nil {
		return
	}

	if !cfg.Enabled {
		return
	}

	memoriesRaw, _ := db.GetAllMemories()
	var memTexts []string
	for _, m := range memoriesRaw {
		memTexts = append(memTexts, fmt.Sprintf("%s: %s", m.FilePath, m.CompressedText))
	}

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

	// --- CROSS-PLATFORM FIX START ---
	// Use os.UserHomeDir() instead of os.Getenv("USERPROFILE")
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Failed to find home directory: %v\n", err)
		os.Exit(1)
	}
	cogitoDir := filepath.Join(home, ".cogito")
	// --- CROSS-PLATFORM FIX END ---

	os.MkdirAll(cogitoDir, 0755)

	defaultCfg := &config.Config{Enabled: true, Intensity: "full"}
	config.Save(defaultCfg)

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
