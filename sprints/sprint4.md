## Summary

Sprint 4 focuses on **Memory & Artifact Management**, delivering both short-term session storage and long-term vector-based memory, alongside a robust artifact service for handling files and logs. The tasks below break each story into concrete developer tasks with clear deliverables and acceptance criteria.

---

## 1. In-Memory SessionStore

**Story:** Provide a concurrency-safe store for per-session state and event history.

| Task                                                                 | Deliverable                                                                          | Acceptance Criteria                                                                                   |
|----------------------------------------------------------------------|--------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| 1.1 Define `Session` and `State` models                              | Go structs for `Session{ID, CreatedAt, Events}` and `State{Data, Metadata}`          | Fields cover timestamp, arbitrary payload, and metadata maps; JSON-serializable                        |
| 1.2 Implement `SessionStore` interface                               | Interface with `Create`, `Get`, `Update`, `Delete` methods                           | Methods return errors on missing IDs; thread-safe under concurrent calls                              |
| 1.3 In-memory backend                                                | `MemorySessionStore` using `sync.RWMutex` or `sync.Map`                              | Passes race detector (`go test -race`); CRUD operations behave correctly                              |
| 1.4 Unit tests                                                       | Tests covering all CRUD paths and concurrent access                                  | 100% branch coverage for store code; no data races in stress tests                                    |

---

## 2. Vector Memory with Weaviate & PgVector

**Story:** Enable long-term memory and context retrieval via vector similarity search using open-source and local-first tools (Weaviate and PgVector).

| Task ID | Task Description                             | Deliverable                                                                         | Acceptance Criteria                                                                                      |
|---------|----------------------------------------------|-------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| 2.1     | Define `VectorMemory` interface              | Interface with methods: `Store(id, embedding, metadata)`, `Query(embedding, topK)` | Interface covers core use cases; clear method contracts; unit-testable signatures                       |
| 2.2     | Implement Weaviate Driver                    | `weaviate_memory.go` using Weaviate Go client SDK                                  | Correctly stores and queries vectors; handles schema setup and metadata; integration test with Docker   |
| 2.3     | Implement PgVector Driver                    | `pgvector_memory.go` using `github.com/jackc/pgx`                                  | Creates `vector` column; stores embeddings with metadata; top-k query implemented with cosine distance   |
| 2.4     | Add Fallback Memory Abstraction (Priority)   | Memory abstraction layer to route between Weaviate and PgVector                    | Configurable via `config.yaml`; fallback if one driver fails; mock tests validate fallback path         |
| 2.5     | Add Integration Tests with Docker Compose    | Compose setup for Postgres + Weaviate sandbox                                      | Run `make test-memory` spins up infra, loads embeddings, performs queries; test assertions passed       |
| 2.6     | Add Mock Implementations for Testing         | `mock_memory.go` with mock `Store`, `Query`, and error injection                   | Unit tests for agents/toolchains use this to simulate memory injection, latency, and partial failures   |
| 2.7     | Document Usage and Schema in `docs/memory.md`| Markdown file with schema examples, diagrams, and usage tips                       | Shows when to use Weaviate vs PgVector; CLI/Docker commands to run services locally                    |

---

### 💡 Notes:
- **Weaviate Go SDK** is very ergonomic and supports hybrid search (BM25 + vectors), great for structured queries too.
- You'll want to make sure **metadata schema in Weaviate** is flexible enough to support tool calls, session context, or agent identity tagging.
- **PgVector** works better if local/offline is a priority. It’s ideal for deployments where Postgres is already in use.

----
## 3. Artifact Service

**Story:** Manage file artifacts (logs, images, PDFs) with versioning and metadata.

| Task                                                                 | Deliverable                                                                          | Acceptance Criteria                                                                                   |
|----------------------------------------------------------------------|--------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| 3.1 Define `ArtifactService` interface                               | Methods: `Save(sessionID, name, io.Reader) (url, error)`, `List(sessionID)`, `Get`   | Interface GoDoc clearly describes URL scheme and metadata                                             |
| 3.2 Local filesystem backend                                         | Stores files under `/artifacts/{sessionID}/{timestamp}_{name}`                       | Files persisted on disk; permissions secure; path traversal guarded                                   |
| 3.3 Metadata manifest                                                | Generate a JSON manifest per session listing artifacts (name, URL, timestamp)        | Manifest updates atomically; can be queried via service                                              |
| 3.4 Unit tests and cleanup                                           | Tests for save, list, get, and cleanup operations                                    | Temporary test directories cleaned between runs; no leftover files                                     |

---

## 4. Runner Integration & Sample Usage

**Story:** Wire in memory and artifact services so agents can leverage storage.

| Task                                                                 | Deliverable                                                                          | Acceptance Criteria                                                                                   |
|----------------------------------------------------------------------|--------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| 4.1 Dependency injection                                             | Extend `RunnerConfig` to accept `SessionStore`, `VectorMemory`, `ArtifactService`    | Runner fails fast if any service is nil                                                           |
| 4.2 Service access API                                              | Add methods on `Runner` to read/write state and artifacts within agent code         | Sample agent can call `runner.SaveState`, `runner.SaveArtifact`                                      |
| 4.3 End-to-end integration tests                                     | Simulate an agent that stores intermediate state and artifacts                       | Tests assert state persists across calls and artifact URLs resolve to actual files                  |
| 4.4 Example in `examples/`                                           | Go program demonstrating a memory-enabled workflow with an artifact log               | README guides user through running the example and inspecting stored sessions and artifacts        |

---

## 5. Documentation, Benchmarking & Demo

| Task                                                                 | Deliverable                                                                          | Acceptance Criteria                                                                                   |
|----------------------------------------------------------------------|--------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| 5.1 GoDoc updates                                                    | Document all new interfaces and types                                                | Public API fully documented with usage examples                                                     |
| 5.2 Benchmarks                                                       | Benchmarks for `SessionStore` and each `VectorMemory` driver                         | Benchmarks under `benchmarks/` folder; record metrics (ops/sec, p50 latency)                         |
| 5.3 Sprint demo prep                                                 | Demo script showing: session persistence, vector query, artifact retrieval           | PM can run `go run examples/memory_demo.go` and see end-to-end functionality                         |

---

By completing these tasks, Sprint 4 will equip the framework with robust, performant memory and artifact capabilities, ensuring agents can maintain context and traceability across sessions.