package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

var DB *sql.DB

func resolveDBPath() (string, error) {
    // 1. Get the User Home Directory
    home, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("could not find user home directory: %v", err)
    }

    // 2. Define the global path (~/.cogito)
    dirPath := filepath.Join(home, ".cogito")

    // 3. Create the folder if it doesn't exist
    if err := os.MkdirAll(dirPath, 0755); err != nil {
        return "", fmt.Errorf("failed to create global config directory at %s: %v", dirPath, err)
    }

    // 4. Set the DB file path
    dbPath := filepath.Join(dirPath, "cogito.db")

    // 5. Verify writability by touching the file
    file, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        return "", fmt.Errorf("global database file %s is not writable: %v", dbPath, err)
    }
    file.Close()

    return dbPath, nil
}

func InitDB() error {
	dbPath, err := resolveDBPath()
	if err != nil {
		return err
	}

	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	if err := DB.Ping(); err != nil {
		return err
	}

	if _, err := DB.Exec(schemaSQL); err != nil {
		return err
	}
	if _, err := DB.Exec(`
		DROP TRIGGER IF EXISTS observations_ai;
		DROP TABLE IF EXISTS observations_fts;
		CREATE VIRTUAL TABLE observations_fts USING fts5(
			title, compressed_text, facts, files_touched,
			content='observations',
			content_rowid='id'
		);
		CREATE TRIGGER observations_ai AFTER INSERT ON observations BEGIN
			INSERT INTO observations_fts(rowid, title, compressed_text, facts, files_touched)
			VALUES (new.id, new.title, new.compressed_text, new.facts, new.files_touched);
		END;
		INSERT INTO observations_fts(rowid, title, compressed_text, facts, files_touched)
		SELECT id, title, compressed_text, facts, files_touched FROM observations;
	`); err != nil {
		return err
	}

	return nil
}

func GetAllMemories(cwd string, limit int) []schemaModels.Observation {
	if DB == nil {
		return nil
	}

	projectPath := filepath.Clean(strings.TrimSpace(cwd))
	projectPathAlt := filepath.ToSlash(projectPath)

	query := `
		SELECT compressed_text, files_touched
		FROM observations
		WHERE project = ? OR project = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := DB.Query(query, projectPath, projectPathAlt, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var observations []schemaModels.Observation

	for rows.Next() {
		var observation schemaModels.Observation

		err := rows.Scan(&observation.CompressedText, &observation.FilesTouched)
		if err != nil {
			continue
		}

		observations = append(observations, observation)
	}

	return observations
}
