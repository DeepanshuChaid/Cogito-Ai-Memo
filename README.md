# Cogito 🧠

Cogito is a project-memory MCP server for coding agents.
It stores durable engineering context per project and reuses it in future sessions.

## ✨ What It Does

- Creates project-scoped memory sessions.
- Stores durable observations with `create_observation`.
- Stores session summaries with `create_summary`.
- Retrieves memory with:
  - `get_project_memory` (past sessions, same project)
  - `get_recent_context` (latest observations + summaries)
- Builds a codebase substrate map with `get_codebase_map`.
- Supports stricter/terser agent workflows (for example caveman-style policies).

## 📦 Install

```bash
git clone https://github.com/DeepanshuChaid/Cogito-Ai.git
cd Cogito-Ai
go install ./cmd/cogito
```

## 🚀 Quick Start

```bash
cogito install
```

`cogito install` writes:

- Root `AGENTS.md` policy block
- Skills inside the project
- MCP server registration in `~/.codex/config.toml`

## 🛠️ MCP Tools

- `create_observation`
  - Input:
    - `memory` (required)
    - `facts` (optional JSON-array string)
- `create_summary`
  - Input:
    - `request`, `learned`, `nextSteps`
  - Guard:
    - Fails if current session has zero observations
- `get_project_memory`
  - Input:
    - `limit` (optional, default `8`)
  - Returns:
    - Past-session observations + summaries for current project
- `get_recent_context`
  - Input:
    - `limit` (optional, default `10`)
  - Returns:
    - Latest observations + summaries for current project
- `get_codebase_map`
  - Returns:
    - `.cogito/substrate.txt` map

## 🗃️ Memory Model

- DB path: `~/.cogito/cogito.db`
- Core tables:
  - `sdk_sessions`
  - `observations`
  - `session_summaries`
  - `observations_fts` (FTS5)

## 💻 CLI Commands

- `cogito install`
- `cogito uninstall`
- `cogito serve-mcp`
- `cogito build-map`
- `cogito --help`
- `cogito -v`

## 📝 Notes

- Session auto-summary is skipped when no observations were created.
- `get_project_memory` excludes the current active session by design.
