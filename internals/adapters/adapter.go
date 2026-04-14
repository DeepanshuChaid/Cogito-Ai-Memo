package adapters

import "github.com/DeepanshuChaid/Cogito-Ai.git/pkg/types"

// ADAPTERS DEFINES THE INTERFACE BETWEEN INTERFACE FOR PLATFORM SPECIFIC ADAPTERS
type Adapter interface {
	// NAME RETURNS THE PLATFORM NAME
	Name() types.Platform

	// PARSE INPUT CONVERTS THE RAW PLATFORM INPUT TO NORMALIZED HOOKINPUT
	ParseInput(raw []byte) (*types.HookInput, error)

	// FORMAT_OUTPUT CONVERTS HOOKOUTPUT TO PLATFORM SEPCIFIC JSON
	FormatOuput(output *types.HookInput) ([]byte, error)

	// GET_HOOK_EVENT EXTRACTS THE EVENT TYPE FROM THE RAW INPUT
	GetHookEvent(raw []byte) (types.HookEvent, error)
}

func GetAdapter(platform types.Platform) Adapter {
	switch platform {
	case types.PlatformClaudeCode:
		// return &ClaudeAdapter{}

	case types.PlatformCursor:
		// return &CursorAdapter{}

	case types.PlatformGemini:
		// return &GeminiAdapter{}

	default:
		// return &DefaultAdapter
	}
	return nil
}
