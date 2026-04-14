package types

type HookEvent string

// HOOKEVENT REPRESENTS THE TYPE OF HOOK BEING TRGIGERED
// FOR THOSE WHO DONT KNOW HOOKS ARE BASICALLY LIKE EVENTS HANDLERS BUT FOR CLI BASED AI'S
const (
	EventSessionStart HookEvent = "session_start"
	EventUserPrompt   HookEvent = "user_prompt"
	EventPreToolUse   HookEvent = "pre_tool_use"
	EventPostToolUse  HookEvent = "post_tool_use"
	EventSessionEnd   HookEvent = "session_end"
)

// THIS REPRESENTS THE AI CLI PLATFORMS
type Platform string

const (
	PlatformClaudeCode Platform = "claude_code"
	PlatformCursor     Platform = "cursor"
	PlatformGemini     Platform = "gemini"
	PlatformWindsurf   Platform = "windsurf"
	PlatformDefault    Platform = "default"
)

// HookInput is the normalized input from the Platforms basically a normalized struct for every type of AI
type HookInput struct {
	Platform     Platform               `json:"platform"`
	Event        HookEvent              `json:"event"`
	SessionID    string                 `json:"session_id"`
	CWD          string                 `json:"cwd"`
	Prompt       string                 `json:"prompt,omitempty"` // OMIT EMPTY MEANS IF ITS VALUE IS GIVEN THE KEY STAYS IF NOT GETS CUCKED
	ToolName     string                 `json:"tool_name,omitempty"`
	ToolInput    map[string]interface{} `json:"tool_input,omitempty"`
	ToolResponse map[string]interface{} `json:"tool_response,omitempty"`

	TranscriptPath string `json:"transcript_path,omitempty"`
	FilePath       string `json:"file_path,omitempty"`
}

// HOOK_OUTPUT IS WHAT WE SEND BACK TO THE PLATFORM (CLI BASED AI'S)
type HookOutput struct {
	Continue        bool                   `json:"continue"`
	SupressOutput   bool                   `json:"supress_output"`
	AdditionContext string                 `json:"additional_context,omitempty"`
	SystemMessage   string                 `json:"system_message,omitempty"`
	UpdatedInput    map[string]interface{} `json:"updated_input,omitempty"`
}

// CONFIG HOLDS THE COGITO CONFIGURATIONS
type Config struct {
	Enabled        bool     `yaml:"enabled"`
	Platform       Platform `yaml:"platform"`
	ContextFile    string   `yaml:"context_file"`
	CompressOutput bool     `yaml:"compress_output"`
	Debug          bool     `yaml:"debug"`
}

// type Config struct {
// 	Enabled        bool     `yaml:"enabled"`
// 	Platform       Platform `yaml:"platform"`
// 	ContextFile    string   `yaml:"context_file"`
// 	CompressOutput bool     `yaml:"compress_output"`
// 	Debug          bool     `yaml:"debug"`
// }
