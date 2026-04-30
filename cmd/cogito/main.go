package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/welcomeUi"
	"github.com/DeepanshuChaid/Cogito-Ai.git/mcpServer"
)

func main() {
	if err := db.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Critical Error: (I MAY ACT NON CHALANT BUT I AM KINDA JUST A BITCH) Could not initialize DB: %v\n", err)
		os.Exit(1)
	}

	// DO NOT PRINT ANYTHING TO STDOUT HERE.
	// Codex expects pure JSON output only.

	if len(os.Args) > 1 {
		switch os.Args[1] {

		case "install":
			commands.Install()
			return

		case "build-map":
			commands.BuildMap(true)
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

		case "serve-mcp":
			mcpServer.ServeMcp()
			return

		case "createSum":
			cwd, _ := os.Getwd()
			observatons, err := db.GetRecentObservations(cwd, 10)
			if err != nil {
				fmt.Println("ERROR: ", err)
			}
			json.NewEncoder(os.Stdout).Encode(observatons)
			return

		// Replace case "createSum" (around line 50)
		case "summarize":
			if len(os.Args) < 5 {
				fmt.Println("Usage: cogito summarize <request> <learned> <next_steps>")
				return
			}
			cwd, _ := os.Getwd()
			// Using a dummy session ID for manual CLI testing
			err := db.CreateSessionSummary("CLI_TEST", cwd, os.Args[2], os.Args[3], os.Args[4])
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("✅ Summary created.")
			}
			return


		default:
			commands.Unknown(os.Args[1])
			return
		}
	}

	welcomeUi.ShowWelcomeUI()
}

