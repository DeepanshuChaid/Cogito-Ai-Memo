package db

import (
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// CREATE PENDING OBSERVATIONS CREATE RAW DATA TO THE QUEUE - FAST
// CALLED BY THE POSTTOOLUSE HOOK
func CreatePendingObservation(memorySessionID, rawInput string) error {
	now := time.Now()
	_, err := DB.Exec(`
		INSERT INTO pending_observations (memory_session_id, raw_input, created_at, processed)
		VALUES (?, ?, ?, 0)
	`, memorySessionID, rawInput, now.Format("2006-01-02 15:04:05"))
	return err
}

// GetUnprocessedObservations fetches pending items for the Worker (Thick Worker)
// Called by: Worker background goroutine
func GetUnprocessedObservations(limit int) ([]schemaModels.PendingObservation, error) {
	rows, err := DB.Query(`
		SELECT id, memory_session_id, raw_input, created_at, processed
		FROM pending_observations
		WHERE processed = 0
		ORDER BY created_at ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pending []schemaModels.PendingObservation
	for rows.Next() {
		var p schemaModels.PendingObservation
		var createdAt string
		err := rows.Scan(&p.ID, &p.MemorySessionID, &p.RawInput, &createdAt, &p.Processed)
		if err != nil {
			continue
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		pending = append(pending, p)
	}

	return pending, nil
}

// MarkObservationProcessed flags a pending item as done
// Called by: Worker after LLM distillation
func MarkObservationProcessed(id int) error {
	_, err := DB.Exec(`
		UPDATE pending_observations
		SET processed = 1
		WHERE id = ?
	`, id)
	return err
}
