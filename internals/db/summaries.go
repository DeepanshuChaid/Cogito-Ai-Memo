package db

import (
	"database/sql"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// CreateSessionSummary writes the global context (SessionEnd)
// Called by: summary-hook
func CreateSessionSummary(memorySessionID, project, request, learned, nextSteps string) error {
	now := time.Now()
	_, err := DB.Exec(`
		INSERT INTO session_summaries (memory_session_id, project, request, learned, next_steps, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, memorySessionID, project, request, learned, nextSteps, now.Format("2006-01-02 15:04:05"))
	return err
}

// GetLatestSessionSummary fetches the last session's summary (Wide Lens)
// Called by: context-hook (SessionStart)
func GetLatestSessionSummary(project string) (*schemaModels.SessionSummary, error) {
	summary := &schemaModels.SessionSummary{}
	var createdAt string

	err := DB.QueryRow(`
		SELECT id, memory_session_id, project, request, learned, next_steps, created_at
		FROM session_summaries
		WHERE project = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, project).Scan(
		&summary.ID, &summary.MemorySessionID, &summary.Project,
		&summary.Request, &summary.Learned, &summary.NextSteps, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	summary.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return summary, nil
}
