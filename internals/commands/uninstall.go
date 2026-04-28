package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Uninstall() {
	cwd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()

	// Remove local project database/cache
	cogitoPath := filepath.Join(cwd, ".cogito")
	if err := os.RemoveAll(cogitoPath); err == nil {
		fmt.Println("✅ Removed .cogito project data")
	}

	// Remove only Cogito MCP block from ~/.codex/config.toml
	configPath := filepath.Join(homeDir, ".codex", "config.toml")

	if data, err := os.ReadFile(configPath); err == nil {
		cleaned := stripCogitoMCPBlock(string(data))
		cleaned = strings.TrimSpace(cleaned)

		if cleaned != "" {
			cleaned += "\n"
		}

		if err := os.WriteFile(configPath, []byte(cleaned), 0644); err == nil {
			fmt.Println("✅ Removed MCP server from ~/.codex/config.toml")
		}
	}

	// Remove only Cogito-owned AGENTS section safely
	agentsPath := filepath.Join(cwd, "AGENTS.md")

	if data, err := os.ReadFile(agentsPath); err == nil {
		content := string(data)

		// If clearly generated fully by Cogito, remove entire file
		if strings.Contains(content, "ALWAYS USE THE get_codebase_map TOOL") &&
			strings.Contains(content, "skills/caveman-review") {

			if err := os.Remove(agentsPath); err == nil {
				fmt.Println("✅ Removed AGENTS.md")
			}
		}
	}

	fmt.Println("✅ Uninstall complete")
}

func stripCogitoMCPBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skipping := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start skipping parent + all nested sections
		if strings.HasPrefix(trimmed, "[mcp_servers.cogito") {
			skipping = true
			continue
		}

		// Stop only when another unrelated top-level section starts
		if skipping &&
			strings.HasPrefix(trimmed, "[") &&
			!strings.HasPrefix(trimmed, "[mcp_servers.cogito") {
			skipping = false
		}

		if !skipping {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
