package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Test utility for inspecting core Cogito tables:
//
// - sdk_sessions
// - observations
// - session_summaries
//
// Also supports:
//
// go run test.go reset
//
// which fully wipes the DB including FTS tables + triggers.

func main() {
	// FAKE COMMENT: this line is intentionally added for testing.
	// Resolve DB path
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("❌ Could not resolve home directory: %v", err)
	}

	dbPath := filepath.Join(home, ".cogito", "cogito.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("❌ Could not open database: %v", err)
	}
	defer db.Close()

	// RESET MODE
	if len(os.Args) > 1 && os.Args[1] == "reset" {
		resetDatabase(db)
		return
	}

	// INSPECTION MODE
	fmt.Printf("🔍 Inspecting Database: %s\n", dbPath)
	fmt.Println(strings.Repeat("=", 100))

	printSessions(db)
	printObservations(db)
	printSummaries(db)
}

func resetDatabase(db *sql.DB) {
	fmt.Println("⚠️ RESET MODE: Wiping Cogito database...")

	// Drop trigger first
	triggers := []string{
		"observations_ai",
		"observations_au",
		"observations_ad",
	}

	for _, trigger := range triggers {
		_, err := db.Exec(fmt.Sprintf(
			"DROP TRIGGER IF EXISTS %s;",
			trigger,
		))
		if err != nil {
			fmt.Printf("❌ Error dropping trigger %s: %v\n", trigger, err)
		} else {
			fmt.Printf("✅ Dropped trigger %s\n", trigger)
		}
	}

	// Drop FTS helper tables + core tables
	tables := []string{
		"observations_fts",
		"observations_fts_config",
		"observations_fts_data",
		"observations_fts_docsize",
		"observations_fts_idx",

		"observations",
		"session_summaries",
		"sdk_sessions",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf(
			"DROP TABLE IF EXISTS %s;",
			table,
		))
		if err != nil {
			fmt.Printf("❌ Error dropping %s: %v\n", table, err)
		} else {
			fmt.Printf("✅ Dropped %s\n", table)
		}
	}

	fmt.Println()
	fmt.Println("✨ Database wiped successfully.")
	fmt.Println("Your MCP server will recreate schema on next start.")
}

//
// SDK SESSIONS
//

func printSessions(db *sql.DB) {
	fmt.Println("\n📂 [SDK_SESSIONS]")

	rows, err := db.Query(`
		SELECT
			id,
			session_id,
			project,
			started_at,
			COALESCE(completed_at, '')
		FROM sdk_sessions
		ORDER BY started_at DESC
	`)
	if err != nil {
		fmt.Printf("⚠️ Error fetching sessions: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf(
		"%-4s | %-24s | %-40s | %-20s | %-20s\n",
		"ID",
		"SessionID",
		"Project",
		"StartedAt",
		"CompletedAt",
	)

	fmt.Println(strings.Repeat("-", 130))

	for rows.Next() {
		var (
			id          int
			sessionID   string
			project     string
			startedAt   string
			completedAt string
		)

		err := rows.Scan(
			&id,
			&sessionID,
			&project,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			fmt.Printf("⚠️ Scan error: %v\n", err)
			continue
		}

		project = trim(project, 38)

		fmt.Printf(
			"%-4d | %-24s | %-40s | %-20s | %-20s\n",
			id,
			sessionID,
			project,
			startedAt,
			completedAt,
		)
	}
}

//
// OBSERVATIONS
//

func printObservations(db *sql.DB) {
	fmt.Println("\n🧠 [OBSERVATIONS]")

	rows, err := db.Query(`
		SELECT
			id,
			session_id,
			project,
			memory,
			facts,
			created_at
		FROM observations
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("⚠️ Error fetching observations: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf(
		"%-4s | %-24s | %-35s | %-45s | %-35s | %-20s\n",
		"ID",
		"SessionID",
		"Project",
		"Memory",
		"Facts",
		"CreatedAt",
	)

	fmt.Println(strings.Repeat("-", 180))

	for rows.Next() {
		var (
			id        int
			sessionID string
			project   string
			memory    string
			facts     string
			createdAt string
		)

		err := rows.Scan(
			&id,
			&sessionID,
			&project,
			&memory,
			&facts,
			&createdAt,
		)
		if err != nil {
			fmt.Printf("⚠️ Scan error: %v\n", err)
			continue
		}

		project = trim(project, 33)
		memory = trim(memory, 43)
		facts = trim(facts, 33)

		fmt.Printf(
			"%-4d | %-24s | %-35s | %-45s | %-35s | %-20s\n",
			id,
			sessionID,
			project,
			memory,
			facts,
			createdAt,
		)
	}
}

//
// SESSION SUMMARIES
//

func printSummaries(db *sql.DB) {
	fmt.Println("\n📝 [SESSION_SUMMARIES]")

	rows, err := db.Query(`
		SELECT
			id,
			session_id,
			project,
			request,
			learned,
			next_steps,
			created_at
		FROM session_summaries
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("⚠️ Error fetching summaries: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf(
		"%-4s | %-24s | %-30s | %-30s | %-30s | %-30s | %-20s\n",
		"ID",
		"SessionID",
		"Project",
		"Request",
		"Learned",
		"NextSteps",
		"CreatedAt",
	)

	fmt.Println(strings.Repeat("-", 210))

	for rows.Next() {
		var (
			id        int
			sessionID string
			project   string
			request   string
			learned   string
			nextSteps string
			createdAt string
		)

		err := rows.Scan(
			&id,
			&sessionID,
			&project,
			&request,
			&learned,
			&nextSteps,
			&createdAt,
		)
		if err != nil {
			fmt.Printf("⚠️ Scan error: %v\n", err)
			continue
		}

		project = trim(project, 28)
		request = trim(request, 28)
		learned = trim(learned, 28)
		nextSteps = trim(nextSteps, 28)

		fmt.Printf(
			"%-4d | %-24s | %-30s | %-30s | %-30s | %-30s | %-20s\n",
			id,
			sessionID,
			project,
			request,
			learned,
			nextSteps,
			createdAt,
		)
	}
}

// ================= //
// Helpers
// ================= //

func trim(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[:max-3] + "..."
}
