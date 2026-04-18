package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
)

func HandleHooks(hookType string) {
	switch hookType {
	case "session-start":
		handleSessionStart()
	case "tool-use":
		handleToolUse()
	case "session-end":
		handleSessionEnd()
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown hook type: %s\n", hookType)
		os.Exit(1)
	}
}

// =============================================================================
// HOOK 1: SESSION START
// Creates session + injects memory context
// =============================================================================
func handleSessionStart() {
	// READ STDIN ONCE
	rawInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ SessionStart: Read Failed: %v\n", err)
		return
	}

	// CLEAN
	cleaned := bytes.TrimPrefix(rawInput, []byte("\xef\xbb\xbf"))
	cleaned = bytes.ReplaceAll(cleaned, []byte("\r"), []byte(""))
	cleaned = bytes.ReplaceAll(cleaned, []byte("\n"), []byte(""))

	// PARSE
	var input struct {
		CWD       string `json:"cwd"`
		SessionID string `json:"session_id"`
		Prompt    string `json:"prompt"`
	}
	if err := json.Unmarshal(cleaned, &input); err != nil {
		fmt.Fprintf(os.Stderr, "❌ SessionStart: JSON Failed: %v\n", err)
		return
	}

	// CREATE SESSION
	_, err = db.CreateSession(input.SessionID, input.CWD, input.Prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  SessionStart: CreateSession Failed: %v\n", err)
	}

	// FETCH MEMORIES (Progressive Disclosure)
	summary, _ := db.GetLatestSessionSummary(input.CWD)
	teasers, _ := db.GetObservationTeasers(input.CWD, 5)

	var contextBuilder strings.Builder
	if summary != nil {
		contextBuilder.WriteString("## Previous Session Summary\n")
		contextBuilder.WriteString(fmt.Sprintf("- Goal: %s\n", summary.Request))
		contextBuilder.WriteString(fmt.Sprintf("- Learned: %s\n", summary.Learned))
		contextBuilder.WriteString(fmt.Sprintf("- Next Steps: %s\n", summary.NextSteps))
		contextBuilder.WriteString("\n")
	}
	if len(teasers) > 0 {
		contextBuilder.WriteString("## Recent Discoveries\n")
		for _, t := range teasers {
			contextBuilder.WriteString(fmt.Sprintf("- [%d] [%s] %s\n", t.ID, t.ObsType, t.Title))
		}
	}


	var memTexts []string
	if contextBuilder.Len() > 0 {
		memTexts = append(memTexts, contextBuilder.String())
	}

	cfg, _ := config.Load()

	context := injector.BuildFinalPrompt("Start session", memTexts, cfg)

	// OUTPUT
	output := map[string]interface{}{
		"continue":       true,
		"suppressOutput": true,
		"systemMessage":  context,
	}
	jsonOut, _ := json.Marshal(output)
	fmt.Println(string(jsonOut))
}

// =============================================================================
// HOOK 2: POST TOOL USE
// Captures AI actions → Creates Observation
// =============================================================================
func handleToolUse() {
	// READ STDIN ONCE
	rawInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ ToolUse: Read Failed: %v\n", err)
		return
	}

	// CLEAN
	cleaned := bytes.TrimPrefix(rawInput, []byte("\xef\xbb\xbf"))
	cleaned = bytes.ReplaceAll(cleaned, []byte("\r"), []byte(""))
	cleaned = bytes.ReplaceAll(cleaned, []byte("\n"), []byte(""))

	// PARSE
	var input struct {
		CWD        string `json:"cwd"`
		SessionID  string `json:"session_id"`
		ToolName   string `json:"tool_name"`
		ToolInput  string `json:"tool_input"`
		ToolOutput string `json:"tool_output"`
	}
	if err := json.Unmarshal(cleaned, &input); err != nil {
		fmt.Fprintf(os.Stderr, "❌ ToolUse: JSON Failed: %v\nRaw: %s\n", err, string(cleaned))
		return
	}

	// GET SESSION
	session, err := db.GetSessionByContentID(input.SessionID)
	if err != nil || session == nil {
		fmt.Fprintf(os.Stderr, "⚠️  ToolUse: Session Not Found: %s\n", input.SessionID)
		output := map[string]interface{}{"continue": true, "suppressOutput": true}
		jsonOut, _ := json.Marshal(output)
		fmt.Println(string(jsonOut))
		return
	}

	// DECIDE IF WORTH SAVING
	if !shouldSaveObservation(input.ToolName, input.ToolOutput) {
		output := map[string]interface{}{"continue": true, "suppressOutput": true}
		jsonOut, _ := json.Marshal(output)
		fmt.Println(string(jsonOut))
		return
	}

	// CATEGORIZE
	obsType := categorizeObservation(input.ToolName, input.ToolOutput)

	// EXTRACT FILES (Simple parsing - improve in Stage 3)
	filesTouched := extractFiles(input.ToolInput, input.ToolOutput)

	// CREATE OBSERVATION (Direct Write - No Queue)
	err = db.CreateObservation(
		*session.MemorySessionID,
		input.CWD,
		obsType,
		fmt.Sprintf("%s: %s", input.ToolName, getActionSummary(input.ToolName, input.ToolOutput)),
		input.ToolOutput, // In Stage 3, LLM compresses this
		"[]",             // facts (JSON array)
		filesTouched,     // files_touched (JSON array)
		0,                // discovery_tokens
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  ToolUse: CreateObservation Failed: %v\n", err)
	}

	// OUTPUT
	output := map[string]interface{}{"continue": true, "suppressOutput": true}
	jsonOut, _ := json.Marshal(output)
	fmt.Println(string(jsonOut))
}

// =============================================================================
// HOOK 3: SESSION END
// Creates summary + marks session complete
// =============================================================================
func handleSessionEnd() {
	// READ STDIN ONCE
	rawInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ SessionEnd: Read Failed: %v\n", err)
		return
	}

	// CLEAN
	cleaned := bytes.TrimPrefix(rawInput, []byte("\xef\xbb\xbf"))
	cleaned = bytes.ReplaceAll(cleaned, []byte("\r"), []byte(""))
	cleaned = bytes.ReplaceAll(cleaned, []byte("\n"), []byte(""))

	// PARSE
	var input struct {
		CWD       string `json:"cwd"`
		SessionID string `json:"session_id"`
		Summary   string `json:"summary"`
	}
	if err := json.Unmarshal(cleaned, &input); err != nil {
		fmt.Fprintf(os.Stderr, "❌ SessionEnd: JSON Failed: %v\n", err)
		return
	}

	// GET SESSION
	session, err := db.GetSessionByContentID(input.SessionID)
	if err != nil || session == nil {
		fmt.Fprintf(os.Stderr, "⚠️  SessionEnd: Session Not Found: %s\n", input.SessionID)
		output := map[string]interface{}{"continue": true, "suppressOutput": true}
		jsonOut, _ := json.Marshal(output)
		fmt.Println(string(jsonOut))
		return
	}

	// CREATE SUMMARY
	err = db.CreateSessionSummary(
		*session.MemorySessionID,
		input.CWD,
		session.UserPrompt, // request
		input.Summary,      // learned
		"Continue work",    // next_steps
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  SessionEnd: CreateSummary Failed: %v\n", err)
	}

	// COMPLETE SESSION
	err = db.CompleteSession(input.SessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  SessionEnd: CompleteSession Failed: %v\n", err)
	}

	// OUTPUT
	output := map[string]interface{}{"continue": true, "suppressOutput": true}
	jsonOut, _ := json.Marshal(output)
	fmt.Println(string(jsonOut))
}

// =============================================================================
// HELPERS
// =============================================================================

func shouldSaveObservation(toolName, output string) bool {
	// Save if: file modified, command succeeded, error fixed
	saveTools := []string{"write_file", "edit_file", "run_command", "search"}
	for _, t := range saveTools {
		if strings.Contains(toolName, t) {
			return true
		}
	}
	if strings.Contains(strings.ToLower(output), "error") ||
	   strings.Contains(strings.ToLower(output), "fixed") ||
	   strings.Contains(strings.ToLower(output), "created") {
		return true
	}
	return false
}

func categorizeObservation(toolName, output string) string {
	if strings.Contains(strings.ToLower(output), "error") ||
	   strings.Contains(strings.ToLower(output), "bug") ||
	   strings.Contains(strings.ToLower(output), "fix") {
		return "bugfix"
	}
	if strings.Contains(toolName, "write") || strings.Contains(toolName, "edit") {
		return "change"
	}
	if strings.Contains(strings.ToLower(output), "found") ||
	   strings.Contains(strings.ToLower(output), "discovered") {
		return "discovery"
	}
	return "observation"
}

func extractFiles(toolInput, toolOutput string) string {
	// Simple extraction - returns JSON array of files
	// Improve with regex in Stage 3
	files := []string{}

	// Look for file paths in tool input/output
	if strings.Contains(toolInput, ".go") || strings.Contains(toolInput, ".py") {
		// Add basic parsing logic here
	}

	result, _ := json.Marshal(files)
	return string(result)
}

func getActionSummary(toolName, output string) string {
	if len(output) > 100 {
		return output[:97] + "..."
	}
	return output
}
