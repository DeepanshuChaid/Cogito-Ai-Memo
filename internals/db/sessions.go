package db

import (
	"database/sql"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// CreateSession creates a new session or returns existing one (idempotent)
// Called by: context-hook (SessionStart)
func CreateSession(contentSessionID, project, userPrompt string) (*schemaModels.Session, error) {
	now := time.Now()

	existing := &schemaModels.Session{}
	var memorySessionID *string  // ✅ Pointer for NULL handling
	var completedAt *time.Time

	err := DB.QueryRow(`
		SELECT id, content_session_id, memory_session_id, project, status, user_prompt, started_at, completed_at
		FROM sdk_sessions
		WHERE content_session_id = ?
	`, contentSessionID).Scan(
		&existing.ID,
		&existing.ContentSessionID,
		&memorySessionID,  // ✅ Scan into pointer
		&existing.Project,
		&existing.Status,
		&existing.UserPrompt,
		&existing.StartedAt,
		&completedAt,
	)

	if err == nil {
		existing.MemorySessionID = memorySessionID  // ✅ Assign pointer
		existing.CompletedAt = completedAt
		return existing, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new session
	result, err := DB.Exec(`
		INSERT INTO sdk_sessions (content_session_id, memory_session_id, project, status, user_prompt, started_at)
		VALUES (?, NULL, ?, 'active', ?, ?)
	`, contentSessionID, project, userPrompt, now.Format("2006-01-02 15:04:05"))

	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()

	return &schemaModels.Session{
		ID:               int(id),
		ContentSessionID: contentSessionID,
		MemorySessionID:  nil,  // ✅ Start as NULL
		Project:          project,
		Status:           "active",
		UserPrompt:       userPrompt,
		StartedAt:        now,
		CompletedAt: nil,
	}, nil
}

func GetSessionByContentID(contentSessionID string) (*schemaModels.Session, error) {
	session := &schemaModels.Session{}
	var memorySessionID *string  // ✅ Pointer for NULL
	var completedAt *time.Time    // ✅ FIX: Pointer for NULL

	err := DB.QueryRow(`
		SELECT id, content_session_id, memory_session_id, project, status, user_prompt, started_at, completed_at
		FROM sdk_sessions
		WHERE content_session_id = ?
	`, contentSessionID).Scan(
		&session.ID,
		&session.ContentSessionID,
		&memorySessionID,  // ✅ Scan into pointer
		&session.Project,
		&session.Status,
		&session.UserPrompt,
		&session.StartedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	session.MemorySessionID = memorySessionID  // ✅ Assign pointer
	session.CompletedAt = completedAt  // ✅ Assign pointer
	return session, nil
}


// UpdateMemorySessionID links a Cogito memory_session_id to the session
// Called by: Worker after first SDK response
func UpdateMemorySessionID(sessionID int, memorySessionID string) error {
	_, err := DB.Exec(`
		UPDATE sdk_sessions
		SET memory_session_id = ?
		WHERE id = ?
	`, memorySessionID, sessionID)

	return err
}

// CompleteSession marks a session as completed
// Called by: summary-hook (SessionEnd)
func CompleteSession(contentSessionID string) error {
	now := time.Now()
	_, err := DB.Exec(`
		UPDATE sdk_sessions
		SET status = 'completed', completed_at = ?
		WHERE content_session_id = ?
	`, now.Format("2006-01-02 15:04:05"), contentSessionID)
	return err
}

