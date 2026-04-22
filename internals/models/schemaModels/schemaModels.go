package schemaModels

import "time"

// =============================================================================
// PROJECTS
// =============================================================================
type Project struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	RootPath     string    `json:"root_path"`
	TechStack    string    `json:"tech_stack"`
	LastAccessed time.Time `json:"last_accessed"`
}

// =============================================================================
// SESSIONS
// =============================================================================
type Session struct {
	ID          int        `json:"id"`
	SessionID   string     `json:"session_id"`
	Project     string     `json:"project"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// =============================================================================
// PENDING OBSERVATIONS
// =============================================================================
type PendingObservation struct {
	ID        int       `json:"id"`
	SessionID string    `json:"session_id"`
	RawInput  string    `json:"raw_input"`
	CreatedAt time.Time `json:"created_at"`
	Processed bool      `json:"processed"`
}

// =============================================================================
// OBSERVATIONS
// =============================================================================
type Observation struct {
	ID              int       `json:"id"`
	SessionID       string    `json:"session_id"`
	Project         string    `json:"project"`
	ObsType         string    `json:"obs_type"`
	Title           string    `json:"title"`
	CompressedText  string    `json:"compressed_text"`
	Facts           string    `json:"facts"`
	FilesTouched    string    `json:"files_touched"`
	DiscoveryTokens int       `json:"discovery_tokens"`
	CreatedAt       time.Time `json:"created_at"`
}

type ObservationTeaser struct {
	ID           int       `json:"id"`
	ObsType      string    `json:"obs_type"`
	Title        string    `json:"title"`
	FilesTouched string    `json:"files_touched"`
	CreatedAt    time.Time `json:"created_at"`
}

type ObservationSearchResult struct {
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Rank  float64 `json:"rank"`
}

// =============================================================================
// SESSION SUMMARIES
// =============================================================================
type SessionSummary struct {
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	Project    string    `json:"project"`
	Request    string    `json:"request"`
	Learned    string    `json:"learned"`
	NextSteps  string    `json:"next_steps"`
	CreatedAt  time.Time `json:"created_at"`
}
