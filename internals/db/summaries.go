package db

import (
	"database/sql"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// CreateSessionSummary writes the global context (SessionEnd)
// Called by: summary-hook
func CreateSessionSummary(SessionID, project, request, learned, nextSteps string) error {
	now := time.Now()
	_, err := DB.Exec(`
		INSERT INTO session_summaries (session_id, project, request, learned, next_steps, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			project = excluded.project,
			request = excluded.request,
			learned = excluded.learned,
			next_steps = excluded.next_steps,
			created_at = excluded.created_at
	`, SessionID, project, request, learned, nextSteps, now.Format("2006-01-02 15:04:05"))
	return err
}

// GetLatestSessionSummary fetches the last session's summary (Wide Lens)
// Called by: context-hook (SessionStart)
func GetLatestSessionSummary(project string) (*schemaModels.SessionSummary, error) {
	summary := &schemaModels.SessionSummary{}
	var createdAt string

	err := DB.QueryRow(`
		SELECT id, session_id, project, request, learned, next_steps, created_at
		FROM session_summaries
		WHERE project = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, project).Scan(
		&summary.ID, &summary.SessionID, &summary.Project,
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

// GetRecentSessionSummaries fetches latest summaries for prompt/context use
func GetRecentSessionSummaries(project string, limit int) ([]schemaModels.SessionSummary, error) {
	rows, err := DB.Query(`
		SELECT id, session_id, project, request, learned, next_steps, created_at
		FROM session_summaries
		WHERE project = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, project, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []schemaModels.SessionSummary
	for rows.Next() {
		var s schemaModels.SessionSummary
		var createdAt string

		if err := rows.Scan(
			&s.ID,
			&s.SessionID,
			&s.Project,
			&s.Request,
			&s.Learned,
			&s.NextSteps,
			&createdAt,
		); err != nil {
			continue
		}

		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

// GetRecentSessionSummariesExcludingSession fetches latest summaries excluding one session
func GetRecentSessionSummariesExcludingSession(project, excludeSessionID string, limit int) ([]schemaModels.SessionSummary, error) {
	rows, err := DB.Query(`
		SELECT id, session_id, project, request, learned, next_steps, created_at
		FROM session_summaries
		WHERE project = ?
			AND session_id != ?
		ORDER BY created_at DESC
		LIMIT ?
	`, project, excludeSessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []schemaModels.SessionSummary
	for rows.Next() {
		var s schemaModels.SessionSummary
		var createdAt string

		if err := rows.Scan(
			&s.ID,
			&s.SessionID,
			&s.Project,
			&s.Request,
			&s.Learned,
			&s.NextSteps,
			&createdAt,
		); err != nil {
			continue
		}

		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return summaries, nil
}
