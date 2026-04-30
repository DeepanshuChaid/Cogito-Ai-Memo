package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

var nonWordRegex = regexp.MustCompile(`[^a-z0-9\s]+`)
var duplicateStopwords = map[string]struct{}{
	"the": {}, "a": {}, "an": {}, "and": {}, "or": {}, "to": {}, "of": {}, "in": {}, "on": {}, "for": {},
	"with": {}, "by": {}, "from": {}, "at": {}, "is": {}, "are": {}, "was": {}, "were": {}, "be": {}, "as": {},
	"this": {}, "that": {}, "it": {}, "into": {}, "we": {}, "you": {}, "i": {}, "our": {}, "their": {},
}

var obsDebugLogMu sync.Mutex
const observationDebugLogPath = "C:\\Users\\HP\\Downloads\\CODING\\Cogito\\debug-2ed107.log"

func writeObservationDebugLog(runID, hypothesisID, location, message string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"sessionId":    "2ed107",
		"runId":        runID,
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}

	obsDebugLogMu.Lock()
	defer obsDebugLogMu.Unlock()

	path := observationDebugLogPath
	if !filepath.IsAbs(path) {
		path = "C:\\Users\\HP\\Downloads\\CODING\\Cogito\\debug-2ed107.log"
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(append(raw, '\n'))
}

func normalizeMeaningText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonWordRegex.ReplaceAllString(s, " ")
	fields := strings.Fields(s)
	filtered := make([]string, 0, len(fields))
	for _, token := range fields {
		if _, skip := duplicateStopwords[token]; skip {
			continue
		}
		filtered = append(filtered, token)
	}
	return strings.Join(filtered, " ")
}

func jaccardSimilarity(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}
	setA := map[string]struct{}{}
	setB := map[string]struct{}{}
	for _, token := range strings.Fields(a) {
		setA[token] = struct{}{}
	}
	for _, token := range strings.Fields(b) {
		setB[token] = struct{}{}
	}
	if len(setA) == 0 || len(setB) == 0 {
		return 0
	}

	intersection := 0
	union := map[string]struct{}{}
	for k := range setA {
		union[k] = struct{}{}
		if _, ok := setB[k]; ok {
			intersection++
		}
	}
	for k := range setB {
		union[k] = struct{}{}
	}

	return float64(intersection) / float64(len(union))
}

func IsDuplicateObservation(project, memory, facts string, recentLimit int) (bool, string, float64, error) {
	if recentLimit <= 0 {
		recentLimit = 30
	}

	current := normalizeMeaningText(memory + " " + facts)
	if current == "" {
		return false, "", 0, nil
	}

	// #region agent log
	writeObservationDebugLog(
		"run1",
		"H4",
		"internals/db/observations.go:IsDuplicateObservation",
		"starting duplicate scan",
		map[string]interface{}{
			"project":      project,
			"recentLimit":  recentLimit,
			"currentTokens": len(strings.Fields(current)),
		},
	)
	// #endregion

	recent, err := GetRecentObservations(project, recentLimit)
	if err != nil {
		return false, "", 0, err
	}

	bestScore := 0.0
	bestCandidate := ""
	for _, obs := range recent {
		candidate := normalizeMeaningText(obs.Memory + " " + obs.Facts)
		if candidate == "" {
			continue
		}
		if current == candidate {
			// #region agent log
			writeObservationDebugLog(
				"run1",
				"H4",
				"internals/db/observations.go:IsDuplicateObservation",
				"exact normalized duplicate found",
				map[string]interface{}{
					"score": 1,
				},
			)
			// #endregion
			return true, obs.Memory, 1, nil
		}

		score := jaccardSimilarity(current, candidate)
		if score > bestScore {
			bestScore = score
			bestCandidate = obs.Memory
		}
		if score >= 0.72 {
			// #region agent log
			writeObservationDebugLog(
				"run1",
				"H5",
				"internals/db/observations.go:IsDuplicateObservation",
				"near duplicate found",
				map[string]interface{}{
					"score": score,
				},
			)
			// #endregion
			return true, obs.Memory, score, nil
		}
	}

	// #region agent log
	writeObservationDebugLog(
		"run1",
		"H5",
		"internals/db/observations.go:IsDuplicateObservation",
		"no duplicate found",
		map[string]interface{}{
			"bestScore": bestScore,
		},
	)
	// #endregion

	return false, bestCandidate, bestScore, nil
}

// CreateObservation writes durable memory directly to DB
func CreateObservation(sessionID, project, memory, facts string) error {
	now := time.Now()

	_, err := DB.Exec(`
		INSERT INTO observations (
			session_id,
			project,
			memory,
			facts,
			created_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		sessionID,
		project,
		memory,
		facts,
		now.Format("2006-01-02 15:04:05"),
	)

	return err
}

// GetRecentObservations fetches latest observations for prompt context
func GetRecentObservations(project string, limit int) ([]schemaModels.Observation, error) {
	rows, err := DB.Query(`
		SELECT
			memory,
			facts,
			created_at
		FROM observations
		WHERE project = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, project, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var observations []schemaModels.Observation

	for rows.Next() {
		var o schemaModels.Observation
		var createdAt string

		err := rows.Scan(
			&o.Memory,
			&o.Facts,
			&createdAt,
		)

		if err != nil {
			continue
		}

		o.CreatedAt, _ = time.Parse(
			"2006-01-02 15:04:05",
			createdAt,
		)

		observations = append(observations, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return observations, nil
}

// GetRecentObservationsExcludingSession fetches latest observations excluding one session
func GetRecentObservationsExcludingSession(project, excludeSessionID string, limit int) ([]schemaModels.Observation, error) {
	rows, err := DB.Query(`
		SELECT
			memory,
			facts,
			created_at
		FROM observations
		WHERE project = ?
			AND session_id != ?
		ORDER BY created_at DESC
		LIMIT ?
	`, project, excludeSessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var observations []schemaModels.Observation
	for rows.Next() {
		var o schemaModels.Observation
		var createdAt string

		if err := rows.Scan(&o.Memory, &o.Facts, &createdAt); err != nil {
			continue
		}
		o.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		observations = append(observations, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return observations, nil
}

// GetSessionObservations fetches observations for a specific session
func GetSessionObservations(sessionID string) ([]schemaModels.Observation, error) {
	rows, err := DB.Query(`
		SELECT
			memory,
			facts,
			created_at
		FROM observations
		WHERE session_id = ?
		ORDER BY created_at ASC
	`, sessionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var observations []schemaModels.Observation

	for rows.Next() {
		var o schemaModels.Observation
		var createdAt string

		err := rows.Scan(
			&o.Memory,
			&o.Facts,
			&createdAt,
		)

		if err != nil {
			continue
		}

		o.CreatedAt, _ = time.Parse(
			"2006-01-02 15:04:05",
			createdAt,
		)

		observations = append(observations, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return observations, nil
}

// SearchObservationsFTS performs fast keyword lookup using FTS5
func SearchObservationsFTS(project, query string, limit int) ([]schemaModels.ObservationSearchResult, error) {
	rows, err := DB.Query(`
		SELECT
			o.id,
			o.memory,
			bm25(observations_fts) as rank
		FROM observations o
		JOIN observations_fts
			ON o.id = observations_fts.rowid
		WHERE
			o.project = ?
			AND observations_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, project, query, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []schemaModels.ObservationSearchResult

	for rows.Next() {
		var r schemaModels.ObservationSearchResult

		err := rows.Scan(
			&r.ID,
			&r.Memory,
			&r.Rank,
		)
		if err != nil {
			continue
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}


// GetObservationByID fetches one full observation
func GetObservationByID(id int) (*schemaModels.Observation, error) {
	row := DB.QueryRow(`
		SELECT
			id,
			session_id,
			project,
			memory,
			facts,
			created_at
		FROM observations
		WHERE id = ?
	`, id)

	var o schemaModels.Observation
	var createdAt string

	err := row.Scan(
		&o.ID,
		&o.SessionID,
		&o.Project,
		&o.Memory,
		&o.Facts,
		&createdAt,
	)

	if err != nil {
		return nil, err
	}

	o.CreatedAt, _ = time.Parse(
		"2006-01-02 15:04:05",
		createdAt,
	)

	return &o, nil
}
