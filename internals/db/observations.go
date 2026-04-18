package db

import (
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// CreateObservation writes distilled memory (Thick Worker - After LLM processing)
// Called by: Worker after distillation
func CreateObservation(memorySessionID, project, obsType, title, compressedText, facts, filesTouched string, discoveryTokens int) error {
	now := time.Now()
	_, err := DB.Exec(`
		INSERT INTO observations (memory_session_id, project, obs_type, title, compressed_text, facts, files_touched, discovery_tokens, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, memorySessionID, project, obsType, title, compressedText, facts, filesTouched, discoveryTokens, now.Format("2006-01-02 15:04:05"))
	return err
}


// GetRecentObservations fetches observations for a project (Progressive Disclosure - Step 1)
// Called by: context-hook (SessionStart)
func GetRecentObservations(project string, limit int) ([]schemaModels.Observation, error) {
	rows, err := DB.Query(`
		SELECT id, memory_session_id, project, obs_type, title, compressed_text, facts, files_touched, discovery_tokens, created_at
		FROM observations
		WHERE project = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, project, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var obs []schemaModels.Observation
	for rows.Next() {
		var o schemaModels.Observation
		var createdAt string
		err := rows.Scan(&o.ID, &o.MemorySessionID, &o.Project, &o.ObsType, &o.Title, &o.CompressedText, &o.Facts, &o.FilesTouched, &o.DiscoveryTokens, &createdAt)
		if err != nil {
			continue
		}
		o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		obs = append(obs, o)
	}

	return obs, nil
}

// GetObservationTeasers fetches only titles for token-efficient context (Progressive Disclosure - Step 2)
// Called by: context-hook (SessionStart)
func GetObservationTeasers(project string, limit int) ([]schemaModels.ObservationTeaser, error) {
	rows, err := DB.Query(`
		SELECT id, obs_type, title, files_touched, created_at
		FROM observations
		WHERE project = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, project, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teasers []schemaModels.ObservationTeaser
	for rows.Next() {
		var t schemaModels.ObservationTeaser
		var createdAt time.Time
		err := rows.Scan(&t.ID, &t.ObsType, &t.Title, &t.FilesTouched, &createdAt)
		if err != nil {
			continue
		}
		t.CreatedAt = createdAt
		teasers = append(teasers, t)
	}

	return teasers, nil
}

// SearchObservationsFTS uses FTS5 for keyword search (Progressive Disclosure - Step 3)
// Called by: Worker or MCP search tool
func SearchObservationsFTS(project, query string, limit int) ([]schemaModels.ObservationSearchResult, error) {
	rows, err := DB.Query(`
		SELECT o.id, o.title, f.rank
		FROM observations o
		JOIN observations_fts f ON o.id = f.rowid
		WHERE o.project = ? AND observations_fts MATCH ?
		ORDER BY f.rank ASC
		LIMIT ?
	`, project, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []schemaModels.ObservationSearchResult
	for rows.Next() {
		var r schemaModels.ObservationSearchResult
		err := rows.Scan(&r.ID, &r.Title, &r.Rank)
		if err != nil {
			continue
		}
		results = append(results, r)
	}

	return results, nil
}
