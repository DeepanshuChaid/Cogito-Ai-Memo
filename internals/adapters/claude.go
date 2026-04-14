package adapters

import (
	"encoding/json"

	"github.com/DeepanshuChaid/Cogito-Ai.git/pkg/types"
)

type ClaudeAdapter struct{}

func (a *ClaudeAdapter) Name() types.Platform {
	return types.PlatformClaudeCode
}

// CLAUDE_CODE RAW INPUT STRUCTURE
type claudeInput struct {
	SessionID string `json:"session_id"`
	CWD string `json:"cwd"`
	Prompt string `json:"prompt"`
	ToolName string `json:"tool_name"`
	ToolInput interface{} `json:"tool_input"`
	ToolResponse interface{} `json:"tool_response"`
	TranscriptPath string `json:"transcript_path"`
}

func (a *ClaudeAdapter) ParseInput(raw []byte) (*types.HookInput, error) {
	var input claudeInput

	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, err
	}

	return &types.HookInput{
		Platform: types.PlatformClaudeCode,
		SessionID: input.SessionID,
		CWD: input.CWD,
		Prompt: input.Prompt,
		ToolName: input.ToolName,
		ToolInput: mapify(input.ToolInput),
		ToolResponse: mapify(input.ToolResponse),
		TranscriptPath: input.TranscriptPath,
	}, nil
}

func (a *ClaudeAdapter) FormatOutput(output types.HookOutput) ([]byte, error) {
	result := map[string]interface{}{
		"continue": output.Continue,
		"supressOutput": output.SupressOutput,
	}

	if output.AdditionContext != "" {
		result["hookSpecificOutput"] = map[string]interface{}{
			"hookEventName": "sessionStart",
			"additionalContext": output.AdditionContext,
		}
	}

	if output.SystemMessage != "" {
		result["systemMessage"] = output.SystemMessage
	}

	return json.Marshal(result)
}

// GET HOOK EVENT
func (a *ClaudeAdapter) GetHookEvent (raw []byte) (types.HookEvent, error) {
	// CLAUDE CODE PASSES EVENT TYPE VIA COMMAND LINE ARG NOT STDIN
	// THIS IS HANDLED BY THE MAIN CLI
	return types.EventSessionStart, nil
}


// HELPER TO CHECK THE TYPE OF THE INTERFACE
func mapify (v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}

	if m, ok := v.(map[string]interface{}); ok {
		return m
	}

	return nil
}
