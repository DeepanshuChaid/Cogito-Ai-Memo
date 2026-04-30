package mcpServer

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
)

// 🔥 KEEP IT SHORT (important)
const PROMPT = `
Code review mode.

STRICT RULES:
- Max 12 words per line
- Max 5 lines total
- If exceeded → output INVALID

Format: L<line>: <problem>. <fix>.
No fluff.
`

var lineFormat = regexp.MustCompile(`^L[0-9]+:\s+.+\.\s+.+\.$`)

func errorResponse(code int, msg string) map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": msg,
		},
	}
}

func validateOutput(text string) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return fmt.Errorf("empty output")
	}

	lines := strings.Split(trimmed, "\n")

	if len(lines) > 5 {
		return fmt.Errorf("too many lines")
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			return fmt.Errorf("empty line")
		}
		if len(strings.Fields(line)) > 12 {
			return fmt.Errorf("line too long")
		}
		if !lineFormat.MatchString(line) {
			return fmt.Errorf("bad format")
		}
	}

	return nil
}

func trimInput(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) > 200 {
		lines = lines[:200]
	}
	return strings.Join(lines, "\n")
}

func runCaveman(prompt string) (string, error) {
	var output string

	for i := 0; i < 2; i++ {
		out, err := callModel(prompt)
		if err != nil {
			return "", err
		}

		if err := validateOutput(out); err == nil {
			return out, nil
		} else {
			prompt = fmt.Sprintf(`

%s

Rewrite only. Previous output INVALID.

ERROR: %v

OUTPUT:
%s

STRICT RULES:
- Max 12 words per line
- Max 5 lines total
- Exact format: L<line>: <problem>. <fix>.
- Output INVALID if rules broken

Return corrected lines only.
`, PROMPT, err, out)

			output = out
		}
	}

	return output, fmt.Errorf("failed to enforce caveman constraints")
}

func callModel(prompt string) (string, error) {
	cmd := exec.Command("codex", "exec", "-")

	cmd.Stdin = strings.NewReader(prompt)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("model error: %v: %s", err, strings.TrimSpace(stderr.String()))
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		errText := strings.TrimSpace(stderr.String())
		if errText != "" {
			return "", fmt.Errorf("empty model output: %s", errText)
		}
		return "", fmt.Errorf("empty model output")
	}

	return result, nil
}

func newSessionID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("session-%d-%s", time.Now().UnixNano(), hex.EncodeToString(b[:]))
}

func GenerateAutoSummary(sessionID string, project string) error {
	// 1. Fetch observations from this session
	obs, err := db.GetSessionObservations(sessionID)
	if err != nil {
		return err
	}

	// Always persist a session summary, even when no observations were captured.
	if len(obs) == 0 {
		return nil
	}

	var obsText bytes.Buffer
	for i, o := range obs {
		obsText.WriteString(fmt.Sprintf("%d. %s (Facts: %s)\n", i+1, o.Memory, o.Facts))
	}

	prompt := fmt.Sprintf(`
You are a senior engineer summarizing a coding session.
Write a dense, technical handover summary for the next session.

OBSERVATIONS FROM THIS SESSION:
%s

FORMAT RULES:
- Be extremely concise. Caveman style allowed if clear.
- Output ONLY the summary text, no pleasantries, no markdown blocks.

SUMMARY:
`, obsText.String())

	// 2. Call the model (runs codex exec -)
	summary, err := callModel(prompt)
	if err != nil {
		// During Ctrl+C shutdown, model subprocess may fail.
		// Persist a deterministic fallback summary so next session has context.
		var fallback bytes.Buffer
		fallback.WriteString("Model summary unavailable during shutdown. Key observations:\n")
		max := len(obs)
		if max > 5 {
			max = 5
		}
		for i := 0; i < max; i++ {
			fallback.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(obs[i].Memory)))
		}
		return db.CreateSessionSummary(
			sessionID,
			project,
			"Auto-Generated via Shutdown Hook",
			strings.TrimSpace(fallback.String()),
			"Review observations and continue from listed changes.",
		)
	}

	// 3. Save to DB
	return db.CreateSessionSummary(sessionID, project, "Auto-Generated via Shutdown Hook", summary, "Check observations for details.")
}
