# Cogito

### Project Overview

Cogito is a CLI tool designed to create a persistent "memo" for AI agents to prevent token wastage. Instead of re-analyzing a codebase every time a session starts, Cogito saves the project state and incrementally updates it based on file changes.

---

### The Four Pillars

#### 1. Go (Golang) Basics

Built with Go for maximum performance.

- **Syntax & Concurrency:** Utilizing Goroutines for high-speed file scanning.
- **CLI:** Implementing a robust command-line interface for developer workflows.

#### 2. SQL Fundamentals

Powered by **SQLite** to manage local project history.

- **Persistence:** Storing metadata and file hashes to track changes over time.
- **Efficiency:** Quick queries to retrieve the last "known good" state of the codebase.

#### 3. The "RAG" Concept

Implementing **Retrieval-Augmented Generation**.

- **Source Management:** Providing the AI with a specific "library" of your code.
- **Context Accuracy:** Ensuring the AI looks at the memo before generating answers to reduce hallucinations.

#### 4. Vector DBs

Using semantic search to handle complex queries.

- **Meaning-Based Search:** Searching for context and intent rather than just exact keywords.
- **Optimization:** Pulling only the most relevant snippets into the active context window.

---

### Problem Statement

**Why:** AI agents currently re-scan full codebases every session, wasting tokens and money.
**How:** By using a local graph-based summary and incremental hashing.
**Where:** A local CLI service that integrates with existing AI coding tools.
**Solution:** Cogito creates a "save state" for your project context, making AI interactions faster and cheaper.
