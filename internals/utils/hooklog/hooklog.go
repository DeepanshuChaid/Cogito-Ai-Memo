// internals/utils/hooklog/hooklog.go
package hooklog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Log writes a line to $HOME/.cogito/hook.log.
// It creates the parent directory if it doesn't exist.
// Returns any error – callers can ignore it (so the hook never crashes) or log it to stderr for debugging.
func Log(event, payload string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot resolve home directory: %w", err)
	}
	dir := filepath.Join(home, ".cogito")
	logPath := filepath.Join(dir, "hook.log")

	// Ensure the directory exists.
	if mkErr := os.MkdirAll(dir, 0o750); mkErr != nil {
		return fmt.Errorf("cannot create %s: %w", dir, mkErr)
	}

	// Open (or create) the log file.
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", logPath, err)
	}
	defer f.Close()

	ts := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("%s | %s | %s\n", ts, event, payload)

	if _, werr := f.WriteString(line); werr != nil {
		return fmt.Errorf("cannot write to %s: %w", logPath, werr)
	}
	return nil
}
