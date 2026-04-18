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

	// 1. Read the entire input into a byte slice
	rawInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Read Stdin Failed: %v\n", err)
		return
	}

	// 1. Read everything from Stdin
    raw, err := io.ReadAll(os.Stdin)
    if err != nil {
        return
    }

    // 2. PowerShell might send bytes separated by Newlines/Carriage Returns
    // This cleans up the 123 \n 34 \n 99 mess you saw in the debug
    cleaned := bytes.ReplaceAll(raw, []byte("\r"), []byte(""))
    cleaned = bytes.ReplaceAll(cleaned, []byte("\n"), []byte(""))

    // 3. Strip the BOM (the 'ï' thing) if it's there
    cleaned = bytes.TrimPrefix(cleaned, []byte("\xef\xbb\xbf"))

    var input struct {
        CWD string `json:"cwd"`
		Prompt string `json:"prompt"`
    }

    if err := json.Unmarshal(cleaned, &input); err != nil {
        fmt.Fprintf(os.Stderr, "DEBUG: Still failing. Content: %s\n", string(cleaned))
        return
    }

	// 2. Strip the UTF-8 BOM (0xEF, 0xBB, 0xBF) if it exists
	// This is what is causing the 'ï' error
	cleanedInput := bytes.TrimPrefix(rawInput, []byte("\xef\xbb\xbf"))

	// 3. Unmarshal from the cleaned byte slice instead of using the decoder directly
	if err := json.Unmarshal(cleanedInput, &input); err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: JSON Decode Failed: %v\n", err)
		// Print what we actually got so you can see it
		fmt.Fprintf(os.Stderr, "DEBUG: Raw Input was: %s\n", string(cleanedInput))
		return
	}

    if (stat.Mode() & os.ModeCharDevice) != 0 {
        welcomeUi.ShowWelcomeUI()
        return
    }


    // DEBUG 1: Is the JSON coming in correctly?
    if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
        fmt.Fprintf(os.Stderr, "DEBUG: JSON Decode Failed: %v\n", err)
        return
    }

    cfg, err := config.Load()
    // DEBUG 2: Is the config file actually there?
    if err != nil {
        fmt.Fprintf(os.Stderr, "DEBUG: Config Load Failed: %v\n", err)
        return
    }

    // DEBUG 3: Is it just turned off?
    if !cfg.Enabled {
        fmt.Fprintf(os.Stderr, "DEBUG: Cogito is DISABLED in config\n")
        return
    }
    cwd, _ := os.Getwd()
    memoriesRaw := db.GetAllMemories(cwd, 20)

    var memTexts []string
    for _, m := range memoriesRaw {
        memTexts = append(memTexts, fmt.Sprintf("%s: %s", m.FilesTouched, m.CompressedText))
    }

    context := injector.BuildFinalPrompt("Start session", memTexts, cfg)

    output := map[string]interface{}{
        "continue":       true,
        "suppressOutput": true,
        "systemMessage":  context,
    }

    jsonOut, _ := json.Marshal(output)
    fmt.Println(string(jsonOut))
}
