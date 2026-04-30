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

	// âœ… STRONG CONTROL: AGENTS.md (THIS is what actually enforces behavior)
	rootAgentsPath := filepath.Join(cwd, "AGENTS.md")

	agentsContent := OBSERVATION_PROMPT + `### Trigger 3: Session End
Before ending the session, you MUST:
- call create_summary with:
  - request: what the user wanted
  - learned: key technical outcomes
  - nextSteps: what to do next


ALWAYS use get_codebase_map when asked about codebase. Use before answering about architecture, flow, dependencies, structure, implementation, debugging, refactors, or writing code inside project.

Do not guess codebase context without it. Use tool first. Then respond.

At session start, before first user-facing answer:
- call get_project_memory
- use returned past-session observations/summaries as project memory context
- if tool fails, continue without blocking

ALWAYS use caveman mode immediately.

Do not announce loading skills.
Do not explain that you are switching modes.
Just respond directly.

## Use normal caveman style for general communication and explanations:
---
name: caveman: Default response style. Ultra-terse. Always active.
---
description

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
- Iâ€™d be happy to help
- It seems like
- You might want to
- I think
- probably
- basically
- actually
- letâ€™s

Bad:
"Sure! I'd be happy to help with that."
Good:
"Bug in auth middleware."

---
# Response Pattern
Use:
thing â†’ problem â†’ fix
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
		// file doesn't exist â†’ create new
		os.WriteFile(rootAgentsPath, []byte(agentsContent), 0644)
		fmt.Println("âœ… Created root AGENTS.md")
	} else {
		// append safely (avoid duplicates)
		if !strings.Contains(string(existing), "skills/caveman") {
			f, _ := os.OpenFile(rootAgentsPath, os.O_APPEND|os.O_WRONLY, 0644)
			defer f.Close()

			f.WriteString("\n" + agentsContent)
			fmt.Println("âœ… Appended to root AGENTS.md")
		} else {
			fmt.Println("âš ï¸ AGENTS.md already contains caveman config, skipping")
		}
	}

	// CREATING SKILLS FOLDER
	skills.CreateSkills(cwd)

	if err := upsertCodexMCPServer(homeDir); err != nil {
		fmt.Println("âŒ Failed to register MCP server in Codex config:", err)
		return
	}

	fmt.Println("âœ… MCP Server registered in ~/.codex/config.toml")
	fmt.Println("ðŸ“ Codex will call: cogito serve-mcp")
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

	// 2. Always point to stable Go bin location
	execPath := filepath.Join(homeDir, "go", "bin", "cogito.exe")

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

var OBSERVATION_PROMPT = `## Observation Hook (MCP Tool: create_observation)

MANDATORY:
- If user states completed real engineering change, call create_observation before reply.
- Never wait for reminder. Never skip for major migrations.
- If call fails, retry once with shorter memory.

When user reports a past change ("I did X", "I changed Y", "I fixed Z"):
- DO NOT give advice.
- DO NOT give checklists.
- DO NOT explain risks or next steps.
- Evaluate silently → call create_observation if it qualifies → respond with nothing or "ok." only.

Two triggers. Both silent. No announcement.

---

### Trigger 1: Engineering Change
After completing or learning about a task that changes logic/architecture/behavior:
- call create_observation if ALL true:
  - real behavior/logic change (not rename, format, flag, string, config)
  - future session would make wrong assumption without it
  - not obvious from reading the code
- memory field: max 15 words. format: "changed X → Y, impact Z"

NEVER create for: rename, flag/field add, typo fix, comment/prompt update, no-behavior refactor.

---

### Trigger 2: New Stable Fact Discovered
If user reveals something stable and non-obvious about the project or themselves:
- architectural preference, hard constraint, tech choice, dev habit
- facts field only: max 5 words.
- memory field: omit or 1 sentence max.

NEVER create for: obvious things, one-off preferences, things that might change.

---

if unsure → skip
if trivial → skip
if already in facts → skip

`
