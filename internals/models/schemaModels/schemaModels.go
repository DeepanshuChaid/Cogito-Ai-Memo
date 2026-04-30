package schemaModels

import "time"

// =============================================================================
// PROJECTS
// =============================================================================
// type Project struct {
// 	ID           int       `json:"id"`
// 	Name         string    `json:"name"`
// 	RootPath     string    `json:"root_path"`
// 	TechStack    string    `json:"tech_stack"`
// 	LastAccessed time.Time `json:"last_accessed"`
// }

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
// type PendingObservation struct {
// 	ID        int       `json:"id"`
// 	SessionID string    `json:"session_id"`
// 	RawInput  string    `json:"raw_input"`
// 	CreatedAt time.Time `json:"created_at"`
// 	Processed bool      `json:"processed"`
// }

// =============================================================================
// OBSERVATIONS
// =============================================================================
type Observation struct {
	ID        int       `json:"id,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	Project   string    `json:"project,omitempty"`
	Memory    string    `json:"memory,omitempty"`
	Facts     string    `json:"facts,omitempty"` // JSON array string
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type ObservationSearchResult struct {
	ID     int     `json:"id"`
	Memory string  `json:"memory"`
	Rank   float64 `json:"rank"`
}

// type ObservationTeaser struct {
// 	ID           int       `json:"id"`
// 	ObsType      string    `json:"obs_type"`
// 	Title        string    `json:"title"`
// 	FilesTouched string    `json:"files_touched"`
// 	CreatedAt    time.Time `json:"created_at"`
// }


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
