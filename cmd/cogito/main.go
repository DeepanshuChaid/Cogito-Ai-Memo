package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	// "path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/welcomeUi"
)

func main() {
	if err := db.InitDB(); err != nil {
        fmt.Fprintf(os.Stderr, "Critical Error: Could not initialize DB: %v\n", err)
        os.Exit(1)
    }

	fmt.Println("DB INTIALIZED SUCCESSFULLY!")

	if len(os.Args) > 1 {
		switch os.Args[1] {

		case "install":
			commands.Install()
			return

		case "config":
			commands.HandleConfig()
			return

		case "--help":
			commands.Help()
			return

		case "uninstall":
			commands.Uninstall()
			return

		case "--version", "-v":
			commands.Version()
			return

		case "compress":
			fmt.Println("Janitor is coming soon... (Step 3)")
			return

		default:
			commands.Unknown(os.Args[1])
			return
		}
	}

	handleHook()
}

func handleHook() {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		welcomeUi.ShowWelcomeUI()
		return
	}

	// 1. Read the input ONLY ONCE
	rawInput, err := io.ReadAll(os.Stdin)
	if err != nil || len(rawInput) == 0 {
		fmt.Fprintf(os.Stderr, "DEBUG: Read Stdin Failed or Empty: %v\n", err)
		return
	}

	// 2. Clean the Windows garbage (Newlines, Returns, and UTF-8 BOM)
	cleaned := bytes.ReplaceAll(rawInput, []byte("\r"), []byte(""))
	cleaned = bytes.ReplaceAll(cleaned, []byte("\n"), []byte(""))
	cleaned = bytes.TrimPrefix(cleaned, []byte("\xef\xbb\xbf"))

	// 3. Parse the JSON
	var input struct {
		CWD    string `json:"cwd"`
		Prompt string `json:"prompt"`
	}

	if err := json.Unmarshal(cleaned, &input); err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: JSON Decode Failed. Content: '%s' Error: %v\n", string(cleaned), err)
		return
	}

	// 4. Load Config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Config Load Failed: %v\n", err)
		return
	}

	if !cfg.Enabled {
		fmt.Fprintf(os.Stderr, "DEBUG: Cogito is DISABLED in config\n")
		return
	}

	// 5. Fetch Memories and Build Context
	cwd, _ := os.Getwd()
	memoriesRaw := db.GetAllMemories(cwd, 20)

	var memTexts []string
	for _, m := range memoriesRaw {
		memTexts = append(memTexts, fmt.Sprintf("%s: %s", m.FilesTouched, m.CompressedText))
	}

	context := injector.BuildFinalPrompt("Start session", memTexts, cfg)

	// 6. Output to the AI
	output := map[string]interface{}{
		"continue":       true,
		"suppressOutput": true,
		"systemMessage":  context,
	}

	jsonOut, _ := json.Marshal(output)
	fmt.Println(string(jsonOut))
}
