package commands

import (
	"fmt"
	"os"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
)


func HandleConfig() {
	if len(os.Args) < 3 {
		printHelp()
		return
	}

	key := os.Args[2]

	switch key {
	case "get":
		handleGet()
	case "set":
		handleSet()
	case "list":
		handleList()
	default:
		fmt.Printf("❌ Unknown command: %s\n", key)
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage: cogito config <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  get <key>          - Get a config value")
	fmt.Println("  set <key> <value>  - Set a config value")
	fmt.Println("  list               - List all config values")
	fmt.Println("")
	fmt.Println("Keys: enabled, intensity")
	fmt.Println("Intensity values: lite, normal, ultra")
}

func handleGet() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	if len(os.Args) < 4 {
		fmt.Println("Usage: cogito config get <key>")
		return
	}

	key := os.Args[3]
	switch key {
	case "enabled":
		fmt.Printf("enabled = %v\n", cfg.Enabled)
	case "intensity":
		fmt.Printf("intensity = %s\n", cfg.Intensity)
	default:
		fmt.Printf("❌ Unknown key: %s\n", key)
	}
}

func handleSet() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: cogito config set <key> <value>")
		return
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	key := os.Args[3]
	value := os.Args[4]

	switch key {
	case "enabled":
		cfg.Enabled = (value == "on" || value == "true" || value == "1")
		fmt.Printf("✅ Set enabled = %v\n", cfg.Enabled)
	case "intensity":
		intensity := config.Intensity(value)

		if valid := config.IsValid(intensity); !valid {
			fmt.Printf("❌ Invalid intensity. Valid values: lite, normal, ultra\n")
			return
		}

		cfg.Intensity = intensity
		fmt.Printf("✅ Set intensity = %s\n", intensity)
	default:
		fmt.Printf("❌ Unknown key: %s\n", key)
		return
	}

	if err := config.Save(cfg); err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
		return
	}
}

func handleList() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	fmt.Println("Current configuration:")
	fmt.Printf("  enabled   = %v\n", cfg.Enabled)
	fmt.Printf("  intensity = %s\n", cfg.Intensity)
}
