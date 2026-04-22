package db

import (
	"strings"
	"sync"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

type ObservationCache struct {
	mu    sync.RWMutex
	store map[string][]schemaModels.Observation // project → observations
}

var ObsCache = &ObservationCache{
	store: make(map[string][]schemaModels.Observation),
}

// CACHE LAYER FOR OBSERVATIONS (In-Memory, per session, for quick access during context assembly)
// WORKERS write to DB and update cache, Context Hook reads from cache first, then DB if needed
func (c *ObservationCache) Set(project string, obs []schemaModels.Observation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[project] = obs
}

func (c *ObservationCache) Get(project string) ([]schemaModels.Observation, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[project]
	return val, ok
}

func (c *ObservationCache) Append(project string, o schemaModels.Observation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[project] = append([]schemaModels.Observation{o}, c.store[project]...)
}


// CreateObservation writes distilled memory (Thick Worker - After LLM processing)
// Called by: Worker after distillation
func CreateObservation(SessionID, project, obsType, title, compressedText, facts, filesTouched string, discoveryTokens int) error {
	now := time.Now()

	res, err := DB.Exec(`
		INSERT INTO observations (
			session_id, project, obs_type, title,
			compressed_text, facts, files_touched,
			discovery_tokens, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, SessionID, project, obsType, title, compressedText, facts, filesTouched, discoveryTokens, now.Format("2006-01-02 15:04:05"))

	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()

	// 🔥 update cache instantly
	ObsCache.Append(project, schemaModels.Observation{
		ID:              int(id),
		SessionID:       SessionID,
		Project:         project,
		ObsType:         obsType,
		Title:           title,
		CompressedText:  compressedText,
		Facts:           facts,
		FilesTouched:    filesTouched,
		DiscoveryTokens: discoveryTokens,
		CreatedAt:       now,
	})

	return nil
}


// GetRecentObservations fetches observations for a project (Progressive Disclosure - Step 1)
// Called by: context-hook (SessionStart)
func GetRecentObservations(project string, limit int) ([]schemaModels.Observation, error) {
	// 🔥 1. Try cache first
	if cached, ok := ObsCache.Get(project); ok && len(cached) >= limit {
		return cached[:limit], nil
	}

	// 2. fallback to DB
	rows, err := DB.Query(`
		SELECT id, session_id, project, obs_type, title,
		       compressed_text, facts, files_touched,
		       discovery_tokens, created_at
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

		err := rows.Scan(&o.ID, &o.SessionID, &o.Project, &o.ObsType,
			&o.Title, &o.CompressedText, &o.Facts, &o.FilesTouched,
			&o.DiscoveryTokens, &createdAt)

		if err != nil {
			continue
		}

		o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		obs = append(obs, o)
	}

	// 🔥 populate cache
	ObsCache.Set(project, obs)

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

func FastSearch(project, query string, limit int) ([]schemaModels.ObservationSearchResult, error) {
	// 🔥 try FTS first
	results, err := SearchObservationsFTS(project, query, limit)
	if err == nil && len(results) > 0 {
		return results, nil
	}

	// fallback → cache scan (cheap)
	if cached, ok := ObsCache.Get(project); ok {
		var out []schemaModels.ObservationSearchResult
		for _, o := range cached {
			if strings.Contains(strings.ToLower(o.Title), strings.ToLower(query)) {
				out = append(out, schemaModels.ObservationSearchResult{
					ID:    o.ID,
					Title: o.Title,
					Rank:  1,
				})
			}
		}
		return out, nil
	}

	return nil, nil
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

