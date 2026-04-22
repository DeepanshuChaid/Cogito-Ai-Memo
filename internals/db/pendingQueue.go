package db

import (
	"log"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// 🔥 Resolve project from session
func GetProjectBySession(sessionID string) (string, error) {
	var project string

	err := DB.QueryRow(`
		SELECT project FROM sdk_sessions WHERE session_id = ?
	`, sessionID).Scan(&project)

	return project, err
}

func StartPendingWorker() {
	go func() {
		for {
			items, err := GetUnprocessedObservations(10)
			if err != nil {
				log.Println("worker fetch error:", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if len(items) == 0 {
				time.Sleep(2 * time.Second)
				continue
			}

			for _, p := range items {

				project, err := GetProjectBySession(p.SessionID)
				if err != nil {
					log.Println("project lookup error:", err)
					continue
				}

				err = CreateObservation(
					p.SessionID,
					project,
					"raw",
					p.RawInput,
					p.RawInput,
					"",
					"",
					0,
				)

				if err != nil {
					log.Println("worker insert error:", err)
					continue
				}

				err = MarkObservationProcessed(p.ID)
				if err != nil {
					log.Println("worker mark error:", err)
				}
			}
		}
	}()
}


// INSERT
func CreatePendingObservation(sessionID, rawInput string) error {
	now := time.Now()

	_, err := DB.Exec(`
		INSERT INTO pending_observations (session_id, raw_input, created_at, processed)
		VALUES (?, ?, ?, 0)
	`, sessionID, rawInput, now.Format("2006-01-02 15:04:05"))

	return err
}

// FETCH
func GetUnprocessedObservations(limit int) ([]schemaModels.PendingObservation, error) {
	rows, err := DB.Query(`
		SELECT id, session_id, raw_input, created_at, processed
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

		err := rows.Scan(
			&p.ID,
			&p.SessionID,
			&p.RawInput,
			&createdAt,
			&p.Processed,
		)
		if err != nil {
			continue
		}

		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		pending = append(pending, p)
	}

	return pending, nil
}

// UPDATE
func MarkObservationProcessed(id int) error {
	_, err := DB.Exec(`
		UPDATE pending_observations
		SET processed = 1
		WHERE id = ?
	`, id)

	return err
}
