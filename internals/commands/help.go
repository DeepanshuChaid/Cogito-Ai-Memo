package commands

import "fmt"

// commands/help.go
func Help() {
    fmt.Println(`Commands:
  cogito install          Install hooks
  cogito uninstall        Remove hooks
  cogito config get <key> Get config value
  cogito config set <key> <value>  Set config value
  cogito config list      Show all config
  cogito config reset     Reset to defaults
  cogito --help           Show this help
  cogito --version        Show version`)
}
