package main

import (
	"encoding/json"
	"fmt"
	"os"

	// "path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/welcomeUi"
)

func main() {
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
			commands.Unkown()
			return
		}
	}

	handleHook()
}

func handleHook() {
	stat, _ := os.Stdin.Stat()

	// If stdin is from terminal → NOT a hook
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		welcomeUi.ShowWelcomeUI()
		return
	}

	var input struct {
		CWD string `json:"cwd"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return
	}

	cfg, err := config.Load()
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

