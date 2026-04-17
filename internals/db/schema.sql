CREATE TABLE IF NOT EXISTS sdk_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content_session_id TEXT UNIQUE NOT NULL,  -- IDE?CODEX SESSION ID
    memory_session_id TEXT, -- COGITO UNIQUE ID
    project TEXT NOT NULL, -- ABSOLUTE CWD PATH
    status TEXT DEFAULT 'active', -- active, completed, failed
    user_prompt TEXT, -- initial prompt
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- 2. PENDING_QUEUE (The "Waiting Room" for Distillation)
-- This is crucial. Hooks write here instantly. Worker processes this async.
CREATE TABLE IF NOT EXISTS pending_observations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    memory_session_id TEXT NOT NULL,
    raw_input TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN DEFAULT 0
);


-- 3. OBSERVATIONS (The "Gold" - Distilled Memory)
-- Worker writes here after LLM distillation
CREATE TABLE IF NOT EXISTS observations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    memory_session_id TEXT NOT NULL,
    project TEXT NOT NULL,
    obs_type TEXT,                           -- bugfix, decision, discovery
    title TEXT,                              -- Short summary
    compressed_text TEXT,                          -- The compressed memory
    facts TEXT,                              -- JSON array of pure facts
    files_touched TEXT,                      -- JSON array of paths
    discovery_tokens INTEGER DEFAULT 0,      -- ROI Tracking (Tier 3 lite)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 4. SESSION_SUMMARIES (The "Global Context")
-- Generated at SessionEnd
CREATE TABLE IF NOT EXISTS session_summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    memory_session_id TEXT UNIQUE NOT NULL,
    project TEXT NOT NULL,
    request TEXT,                            -- What user asked
    learned TEXT,                            -- What AI learned
    next_steps TEXT,                         -- What's next
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


-- 5. FTS5 VIRTUAL TABLE (The Search Engine - Tier 2)
-- Mirrors observations for fast keyword search without loading full rows
CREATE VIRTUAL TABLE IF NOT EXISTS observations_fts USING fts5(
    title, compressed_text, facts, files_touched,
    content='observations',
    content_rowid='id'
);

-- Triggers to keep FTS in sync (Auto-update when observations change)
CREATE TRIGGER IF NOT EXISTS observations_ai AFTER INSERT ON observations BEGIN
  INSERT INTO observations_fts(rowid, title, compressed_text, facts, files_touched)
  VALUES (new.id, new.title, new.compressed_text, new.facts, new.files_touched);
END;
