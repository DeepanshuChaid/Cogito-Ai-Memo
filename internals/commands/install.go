package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/configTui"
)

func Install() {
	configTui.RunConfigTUI()

	execPath, _ := os.Executable()
	execPath, _ = filepath.EvalSymlinks(execPath)

	cwd, _ := os.Getwd()

	// Create .cogito
	cogitoDir := filepath.Join(cwd, ".cogito")
	os.MkdirAll(cogitoDir, 0755)

	// Create .codex
	hooksDir := filepath.Join(cwd, ".codex")
	os.MkdirAll(hooksDir, 0755)

	// FOR EDGE CASE WE GONNA WRITE IT IN THE HOME DIR AS WELL
	homeDir, _ := os.UserHomeDir()
	homeCodexDir := filepath.Join(homeDir, ".codex")

	// ✅ All 3 Hooks with args
	hooksConfig := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]interface{}{
				{
					"type":    "command",
					"command": execPath,
					"args":    []string{"hook", "session-start"},
				},
			},
			"PostToolUse": []map[string]interface{}{
				{
					"type":    "command",
					"command": execPath,
					"args":    []string{"hook", "tool-use"},
				},
			},
			"SessionEnd": []map[string]interface{}{
				{
					"type":    "command",
					"command": execPath,
					"args":    []string{"hook", "session-end"},
				},
			},
		},
	}

	content, _ := json.MarshalIndent(hooksConfig, "", "  ")
	os.WriteFile(filepath.Join(hooksDir, "hooks.json"), content, 0644)
	os.WriteFile(filepath.Join(homeCodexDir, "hooks.json"), content, 0644)

	fmt.Println("✅ Cogito hooks installed successfully!")
	fmt.Println("📍 Config:", filepath.Join(hooksDir, "hooks.json"))
	fmt.Println("\nRegistered hooks:")
	fmt.Println("  - SessionStart  → cogito hook session-start")
	fmt.Println("  - PostToolUse   → cogito hook tool-use")
	fmt.Println("  - SessionEnd    → cogito hook session-end")

	// Create/append to .instructions.md with caveman-review prompt
	instructionsPath := filepath.Join(cwd, ".instructions.md")
	cavemanPrompt := `# Must Follow

Ultra-compressed code review comments. Cuts noise from PR feedback while preserving
the actionable signal. Each comment is one line: location, problem, fix. Use when user
says "review this PR", "code review", "review the diff", "/review", or invokes
/caveman-review. Auto-triggers when reviewing pull requests.

---

Write code review comments terse and actionable. One line per finding. Location, problem, fix. No throat-clearing.

## Rules

**Format:** L<line>: <problem>. <fix>. — or <file>:L<line>: ... when reviewing multi-file diffs.

**Severity prefix (optional, when mixed):**
- 🔴 bug: — broken behavior, will cause incident
- 🟡 risk: — works but fragile (race, missing null check, swallowed error)
- 🔵 nit: — style, naming, micro-optim. Author can ignore
- ❓ q: — genuine question, not a suggestion

**Drop:**
- "I noticed that...", "It seems like...", "You might want to consider..."
- "This is just a suggestion but..." — use nit: instead
- "Great work!", "Looks good overall but..." — say it once at the top, not per comment
- Restating what the line does — the reviewer can read the diff
- Hedging ("perhaps", "maybe", "I think") — if unsure use q:

**Keep:**
- Exact line numbers
- Exact symbol/function/variable names in backticks
- Concrete fix, not "consider refactoring this"
- The *why* if the fix isn't obvious from the problem statement

## Examples

❌ "I noticed that on line 42 you're not checking if the user object is null before accessing the email property. This could potentially cause a crash if the user is not found in the database. You might want to add a null check here."

✅ L42: 🔴 bug: user can be null after .find(). Add guard before .email.

❌ "It looks like this function is doing a lot of things and might benefit from being broken up into smaller functions for readability."

✅ L88-140: 🔵 nit: 50-line fn does 4 things. Extract validate/normalize/persist.

❌ "Have you considered what happens if the API returns a 429? I think we should probably handle that case?"

✅ L23: 🟡 risk: no retry on 429. Wrap in withBackoff(3).

## Auto-Clarity

Drop terse mode for: security findings (CVE-class bugs need full explanation + reference), architectural disagreements (need rationale, not just a one-liner), and onboarding contexts where the author is new and needs the "why". In those cases write a normal paragraph, then resume terse for the rest.

## Boundaries

Reviews only — does not write the code fix, does not approve/request-changes, does not run linters. Output the comment(s) ready to paste into the PR. "stop caveman-review" or "normal mode": revert to verbose review style.
`

	// Check if file exists
	fileInfo, err := os.Stat(instructionsPath)
	if err == nil && fileInfo.Size() > 0 {
		// File exists, append if caveman prompt not already there
		existingContent, _ := os.ReadFile(instructionsPath)
		if !strings.Contains(string(existingContent), "Cogito Review Mode") {
			f, _ := os.OpenFile(instructionsPath, os.O_APPEND|os.O_WRONLY, 0644)
			f.WriteString("\n\n" + cavemanPrompt)
			f.Close()
			fmt.Println("✅ Caveman-review prompt appended to .instructions.md")
		}
	} else {
		// File doesn't exist, create it
		os.WriteFile(instructionsPath, []byte(cavemanPrompt), 0644)
		fmt.Println("✅ Created .instructions.md with caveman-review prompt")
	}

	if err := upsertCodexMCPServer(homeDir); err != nil {
		fmt.Println("❌ Failed to register MCP server in Codex config:", err)
		return
	}

	fmt.Println("✅ MCP Server registered in ~/.codex/config.toml")
	fmt.Println("📍 Codex will call: cogito serve-mcp")
}

func upsertCodexMCPServer(homeDir string) error {
	codexDir := filepath.Join(homeDir, ".codex")
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(codexDir, "config.toml")
	existing, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(existing)
	content = stripCogitoMCPBlock(content)
	if strings.TrimSpace(content) != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content += "[mcp_servers.cogito]\n"
	content += "command = \"cogito\"\n"
	content += "args = [\"serve-mcp\"]\n"

	return os.WriteFile(configPath, []byte(content), 0644)
}

func stripCogitoMCPBlock(content string) string {
	blockPattern := `(?ms)\n?\[mcp_servers\.cogito\]\n(?:[^\[]*\n?)*`
	re := regexp.MustCompile(blockPattern)
	updated := re.ReplaceAllString(content, "")
	return strings.TrimRight(updated, "\n") + "\n"
}
