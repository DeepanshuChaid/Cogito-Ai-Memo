package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

// commands/uninstall.go
func Uninstall() {
    cwd, _ := os.Getwd()
    os.RemoveAll(filepath.Join(cwd, ".cogito"))
    os.RemoveAll(filepath.Join(cwd, ".codex"))
    fmt.Println("✅ Uninstalled")
}
