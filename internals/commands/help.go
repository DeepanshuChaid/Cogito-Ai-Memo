package commands

import "fmt"

func Help() {
	fmt.Println(`Usage:
  cogito <command>

Core Commands:
  install        Install and configure Cogito
  uninstall      Remove Cogito and cleanup
  build-map      Generate full codebase substrate map

Info:
  --help         Show help information
  --version, -v  Show current version

Internal / MCP:
  serve-mcp      Start MCP stdio server for Codex/Claude

Notes:
  • serve-mcp is internal and should not be run manually
  • install registers the MCP server in ~/.codex/config.toml
  • uninstall removes only Cogito-owned MCP config + local project data
`)
}
