package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			runInstall()
			return
		case "compress":
			fmt.Println("Janitor is coming soon... (Step 3)")
			return
		}
	}

	// This part handles the Codex Hook (JSON input/output)
	handleHook()
}

func handleHook() {
	// Codex sends JSON to Stdin
	var input struct {
		CWD string `json:"cwd"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return
	}

	// 1. Get all compressed memories from SQLite
	memoriesRaw, _ := db.GetAllMemories()
	var memTexts []string
	for _, m := range memoriesRaw {
		memTexts = append(memTexts, fmt.Sprintf("%s: %s", m.FilePath, m.CompressedText))
	}

	// 2. We don't have a user query yet during SessionStart,
	// so we just inject the rules and the project knowledge as a system message.
	context := injector.BuildFinalPrompt("Start session", memTexts)

	output := map[string]interface{}{
		"continue": true,
		"systemMessage": context,
	}

	jsonOut, _ := json.Marshal(output)
	fmt.Println(string(jsonOut))
}

func runInstall() {
	fmt.Println("Installing Cogito...")

	execPath, _ := os.Executable()
	cwd, _ := os.Getwd()
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
	fmt.Printf("Installed hook at %s\n", execPath)
}
