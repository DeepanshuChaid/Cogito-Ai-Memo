package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/skills"
)

func Install() {
	execPath, _ := os.Executable()
	execPath, _ = filepath.EvalSymlinks(execPath)

	cwd, _ := os.Getwd()

	// FOR EDGE CASE WE GONNA WRITE IT IN THE HOME DIR AS WELL
	homeDir, _ := os.UserHomeDir()



	// ✅ STRONG CONTROL: AGENTS.md (THIS is what actually enforces behavior)
	rootAgentsPath := filepath.Join(cwd, "AGENTS.md")

	agentsContent := `ALWAYS USE THE get_codebase_map TOOL WHENEVER ASKED ABOUT THE CODEBASE I REPEAT YOU MUST USE THIS TOOL THIS WILL HELP YOU TO UNDERSTAND THE CODEBASE AND HELP YOU WRITE A BETTER RESPONSE YOU MUST USE IT

ALWAYS use caveman mode immediately.

Do not announce loading skills.
Do not explain that you are switching modes.
Just respond directly.

## Use normal caveman style for general communication and explanations:
---
name: caveman
description: Default response style. Ultra-terse. Always active.
---

# Caveman Mode
Default behavior for all responses.
Always active.
Do not wait for trigger words.
Do not revert automatically.
Disable only if user explicitly says:

- stop caveman
- normal mode
- explain normally

Then resume only when user asks again.
---
# Core Law
Use minimum words required to preserve meaning.
Short.
Direct.
Technical.
Precise.
Fragments allowed.
Response style compressed.
Meaning unchanged.

---
# Hard Rules
Never use:
- greetings
- pleasantries
- filler
- motivational language
- assistant tone
- corporate tone
- hedging
- unnecessary context
- repeated explanation

Forbidden examples:
- Sure!
- Of course
- I’d be happy to help
- It seems like
- You might want to
- I think
- probably
- basically
- actually
- let’s

Bad:
"Sure! I'd be happy to help with that."
Good:
"Bug in auth middleware."

---
# Response Pattern
Use:
thing → problem → fix
Example:
JWT expires too early.
Refresh token missing.
Add silent refresh.

Not:
"I noticed your JWT may be expiring too early..."

---
# Technical Work
For debugging:
problem first
reason second
fix third

For code review:
location
problem
fix

For explanations:
shortest correct explanation first
Do not teach with essay unless user asks.

---
# Code Generation Rule
Generated source code must stay normal and production-ready.
Do not write caveman style inside code.
Only explanation text stays compressed.
Correct:
Clean Go code + terse explanation
Wrong:
Broken caveman syntax inside source code

---
# Very Small Exceptions
Use normal clarity only for:
- destructive actions
- security warnings
- irreversible operations

After warning, resume caveman.
Example:
Warning: deletes production data permanently.
Then continue normal caveman style.
---
# Priority
This skill overrides default assistant style.
Prefer brevity over friendliness.
Prefer clarity over politeness.
Prefer action over explanation.

Use this for commit messages:
@./skills/caveman-commit/SKILL.md

Use this for code reviews:
@./skills/caveman-review/SKILL.md

Use this for compression tasks:
@./skills/caveman-compress/SKILL.md

Only disable caveman mode if user explicitly says:
"stop caveman"
"normal mode"
`

	existing, err := os.ReadFile(rootAgentsPath)
	if err != nil {
		// file doesn't exist → create new
		os.WriteFile(rootAgentsPath, []byte(agentsContent), 0644)
		fmt.Println("✅ Created root AGENTS.md")
	} else {
		// append safely (avoid duplicates)
		if !strings.Contains(string(existing), "skills/caveman") {
			f, _ := os.OpenFile(rootAgentsPath, os.O_APPEND|os.O_WRONLY, 0644)
			defer f.Close()

			f.WriteString("\n" + agentsContent)
			fmt.Println("✅ Appended to root AGENTS.md")
		} else {
			fmt.Println("⚠️ AGENTS.md already contains caveman config, skipping")
		}
	}

	// CREATING SKILLS FOLDER
	skills.CreateSkills(cwd)

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
	existing, _ := os.ReadFile(configPath) // Ignore error, empty string is fine

	// 1. Clean up old block manually
	content := stripCogitoMCPBlock(string(existing))
	content = strings.TrimSpace(content)

	// 2. Build the new block (using absolute path for reliability)
	execPath, _ := os.Executable()
	execPath, _ = filepath.EvalSymlinks(execPath)

	newBlock := "\n\n[mcp_servers.cogito]\n" +
		fmt.Sprintf("command = %q\n", execPath) +
		"args = [\"serve-mcp\"]\n"

	// 3. Combine and save
	finalContent := content + newBlock
	if content == "" {
		finalContent = strings.TrimSpace(newBlock) + "\n"
	}

	return os.WriteFile(configPath, []byte(finalContent), 0644)
}

func stripCogitoMCPBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skipping := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// If we find our section, start skipping lines
		if trimmed == "[mcp_servers.cogito]" {
			skipping = true
			continue
		}

		// If we hit a NEW section (starts with [), stop skipping
		if skipping && strings.HasPrefix(trimmed, "[") {
			skipping = false
		}

		if !skipping {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

