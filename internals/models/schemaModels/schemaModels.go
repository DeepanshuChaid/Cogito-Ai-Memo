package schemaModels

import "time"

// =============================================================================
// SESSIONS (sdk_sessions table)
// Tracks every session with both IDE and Cogito IDs
// =============================================================================
type Session struct {
	ID                int       `json:"id"`
	ContentSessionID  string    `json:"content_session_id"` // The IDE/Codex Session ID
	MemorySessionID   *string    `json:"memory_session_id"`  // Cogito's Unique ID (can be null initially)
	Project           string    `json:"project"`            // Absolute CWD Path
	Status            string    `json:"status"`             // active, completed, failed
	UserPrompt        string    `json:"user_prompt"`        // Initial prompt that started session
	StartedAt         time.Time `json:"started_at"`
	CompletedAt       *time.Time `json:"completed_at"`
}



// =============================================================================
// PENDING QUEUE (pending_observations table)
// The "Waiting Room" - Hooks write here, Worker processes async
// =============================================================================
type PendingObservation struct {
	ID          int       `json:"id"`
	MemorySessionID string `json:"memory_session_id"`
	RawInput    string    `json:"raw_input"`      // The raw tool output/log (uncleaned)
	CreatedAt   time.Time `json:"created_at"`
	Processed   bool      `json:"processed"`      // Flag for Worker to pick up
}

// =============================================================================
// OBSERVATIONS (observations table + observations_fts virtual table)
// The "Gold" - Distilled Memory (what gets sent to AI)
// =============================================================================
type Observation struct {
	ID              int       `json:"id"`
	MemorySessionID string    `json:"memory_session_id"`
	Project         string    `json:"project"`
	ObsType         string    `json:"obs_type"`           // bugfix, decision, discovery, feature, refactor
	Title           string    `json:"title"`              // Short 1-sentence summary
	CompressedText       string    `json:"compressed_text"`          // The compressed memory (THIS IS WHAT GETS SENT TO AI)
	Facts           string    `json:"facts"`              // JSON array of pure facts
	FilesTouched    string    `json:"files_touched"`      // JSON array of file paths
	DiscoveryTokens int       `json:"discovery_tokens"`   // ROI Tracking (Tier 3 lite)
	CreatedAt       time.Time `json:"created_at"`
}


type ObservationTeaser struct {
	ID int `json:"id"`
	ObsType         string    `json:"obs_type"`           // bugfix, decision, discovery, feature, refactor
	Title           string    `json:"title"`              // Short 1-sentence summary
	FilesTouched    string    `json:"files_touched"`      // JSON array of file paths
	CreatedAt       time.Time `json:"created_at"`
}

// =============================================================================
// OBSERVATION SEARCH RESULT (observations_fts virtual table)
// For Progressive Disclosure - Step 1 (Cheap Token Cost)
// =============================================================================
type ObservationSearchResult struct {
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Rank  float64 `json:"rank"`
}

// =============================================================================
// SESSION SUMMARIES (session_summaries table)
// The "Global Context" - Generated at SessionEnd
// =============================================================================
type SessionSummary struct {
	ID            int       `json:"id"`
	MemorySessionID string `json:"memory_session_id"`
	Project       string    `json:"project"`
	Request       string    `json:"request"`        // What user asked for
	Investigated  string    `json:"investigated"`   // What was explored
	Learned       string    `json:"learned"`        // What AI learned
	Completed     string    `json:"completed"`      // What was finished
	NextSteps     string    `json:"next_steps"`     // What's next
	CreatedAt     time.Time `json:"created_at"`
}

