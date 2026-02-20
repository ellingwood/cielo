# Cielo Build Transcript

A running log of the design and build process for Cielo — a Trello-style Kanban board for AI agent orchestration.

---

## Phase 1: Brainstorming & Design

**Date:** 2026-02-19

### The Idea

Build a Trello clone purpose-built for orchestrating projects that AI agents work on. The core insight: agents need a structured way to pick up tasks, report progress, and coordinate with each other and humans. A Kanban board with an MCP interface gives any LLM agent (Claude, GPT, Gemini, etc.) a standard way to interact with a project board.

### Key Design Decisions

**Q: What kinds of agents will use this?**
A: Mixed LLM agents — the board should be agent-agnostic. Any agent that speaks MCP can use it.

**Q: How should agents interact with the board?**
A: Both modes — agents can self-assign cards autonomously, or humans can assign cards to specific agents. Flexible orchestration.

**Q: What database?**
A: SQLite. Simple, file-based, zero config. Perfect for the initial version.

**Q: Which orchestration features matter most?**
A: All three proposed features:
- **Agent activity log** — track what each agent did, when
- **Card dependencies** — cards can block other cards
- **Labels/tags for agent type** — tag cards with coding, research, review, etc.

**Q: Single binary or separate processes?**
A: Single binary. One Go process serves both the Fiber v3 HTTP API and the JSON-RPC MCP server, sharing a common service layer and database. Simplest deployment, no state coordination issues.

**Q: Real-time updates?**
A: Server-Sent Events (SSE). Unidirectional push from server to client — simple to implement, sufficient for board updates. Events scoped per board with reconnection support via Last-Event-ID.

### Architecture: Monolith with Shared Service Layer

We evaluated three approaches:

1. **Monolith with shared data layer** — single binary, shared service layer (chosen)
2. **Layered with message queue** — adds embedded NATS, unnecessary complexity for SQLite
3. **Separate processes with shared DB** — coordination problems with SQLite write locking

The monolith won because a shared service layer means a card created via MCP and a card created via the UI go through the same validation, persistence, and event broadcasting. No duplication. Clean boundary if we ever need to split.

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.26 |
| HTTP Framework | Go Fiber v3 |
| Database | SQLite (modernc.org/sqlite, pure Go) |
| IDs | UUIDv7 (time-ordered, index-efficient) |
| MCP Transport | Streamable HTTP (JSON-RPC 2.0, POST /mcp) |
| Frontend | React 19 + TypeScript + Vite |
| Styling | Tailwind CSS v4 |
| Drag & Drop | dnd-kit |
| Server State | TanStack Query |
| Routing | React Router |

### Data Model

Six core entities: Board, List, Card, CardDependency, Label (with CardLabel join table), and ActivityLog. All IDs are UUIDv7. Position-based ordering for lists and cards. Assignee is a plain string (agents identify themselves by name — no user table needed). Activity log detail stored as JSON for flexibility.

### MCP Tools

7 read tools and 13 write tools covering full board/list/card CRUD, dependencies, labels, comments, and search. All write tools emit events to the SSE bus and log to the activity log with actor attribution.

### HTTP API

Full REST API under `/api/v1` mirroring the MCP tool set. SSE endpoint at `/api/v1/boards/:boardId/events` for real-time board updates. Same service layer as MCP — consistent behavior regardless of entry point.

### Frontend

Kanban board with drag-and-drop (dnd-kit), card detail modals, markdown descriptions, dependency visualization, and a live activity feed. SSEProvider context manages the EventSource connection and invalidates TanStack Query cache on incoming events for automatic UI updates.

### Documents Produced

- [`docs/requirements.md`](./requirements.md) — 60+ numbered requirements across 6 categories
- [`docs/plans/2026-02-19-cielo-design.md`](./plans/2026-02-19-cielo-design.md) — Full technical design with data model, API surface, component tree, and architecture diagrams

---

## Phase 2: Implementation

**Date:** 2026-02-19

### Approach

Full-stack implementation in a single session. Go backend first, then React frontend, then Docker Compose to tie it all together. Target: a running application at `http://localhost:8080`.

### Go Backend

Built the complete backend in Go with Fiber v3:

- **SQLite schema** — 7 tables (boards, lists, cards, card_dependencies, labels, card_labels, activity_log) with foreign keys, cascade deletes, and CHECK constraints. WAL journal mode for concurrent reads.
- **Store layer** — Repository pattern with `Store` interface and full SQLite implementation (~450 lines). Dynamic query building for card search with optional JOIN filtering by label.
- **Event bus** — In-process pub/sub with board-scoped channels. Non-blocking publish, atomic sequence IDs. 64-slot buffered channels per subscriber.
- **Service layer** — Shared business logic between HTTP API and MCP. Validates inputs, logs activity, publishes events. Auto-sets card status on assign/unassign.
- **HTTP API** — Full REST under `/api/v1` with 22 routes: CRUD for boards, lists, cards, labels. Plus move, assign, dependencies, search, activity, and SSE.
- **MCP server** — JSON-RPC 2.0 dispatcher handling `initialize`, `tools/list`, `tools/call`. 20 tools (7 read + 13 write) with full JSON Schema input definitions. MCP protocol version 2025-11-25.
- **SSE endpoint** — Board-scoped Server-Sent Events with `Last-Event-ID` reconnection support.
- **12 unit tests** — store layer (migrations, board/list/card CRUD, move, dependencies, labels, activity, search) and event bus (subscribe/publish, board scoping, multiple subscribers). All passing.

### React Frontend

Built with React 19 + TypeScript + Vite + Tailwind CSS v4:

- **API client** — Typed fetch wrappers for all endpoints with error handling.
- **TanStack Query hooks** — 17 hooks covering all CRUD operations with cache invalidation.
- **Board list page** — Grid of boards with create/delete.
- **Kanban board view** — Full drag-and-drop with dnd-kit (PointerSensor, closestCorners collision). DragOverlay for visual feedback during drags. Cards show priority color dots, assignee badges, label chips.
- **Card detail modal** — Inline title editing, status/priority dropdowns, assignee input, description textarea, label management, dependency management (add/remove), full activity feed. All changes save on blur.
- **SSE integration** — `useSSE` hook creates EventSource per board, invalidates query cache on events for automatic UI refresh.

### Docker Setup

Multi-stage Dockerfile:
1. `node:22-alpine` — builds frontend (`npm ci` + `npm run build`)
2. `golang:1.25-alpine` — builds backend (`CGO_ENABLED=0 go build`)
3. `alpine:3.21` — runtime with just the binary + static files

Docker Compose exposes port 8080 with a named volume for SQLite persistence at `/data/cielo.db`.

### Issues Encountered & Fixed

1. **Go version mismatch** — `go.mod` requires 1.25, initial Dockerfile used `golang:1.24`. Fixed by updating to `golang:1.25-alpine`.
2. **Private npm registry in lockfile** — `package-lock.json` pointed to a private JFrog Artifactory. Regenerated with public registry.
3. **Binary committed to git** — 18MB `cielo` binary accidentally tracked. Removed with `git rm --cached`, added `.gitignore`.
4. **Overly broad .gitignore** — Pattern `cielo` matched `cmd/cielo/` directory. Fixed to `/cielo` (root only).
5. **Type assertion error** — `[]any{}.([]any)` invalid in Go. Changed to `return c.JSON([]any{})`.
6. **TypeScript verbatimModuleSyntax** — DragEndEvent/DragStartEvent imported as values. Fixed with `type` keyword imports.

### Verification

End-to-end testing confirmed:
- Frontend serves at `http://localhost:8080` with SPA routing
- Board CRUD, list CRUD, card CRUD all functional
- Card move, assign (with auto-status), dependencies, labels all working
- Search by title/description returns correct results
- Activity log tracks all mutations with actor/action/detail/timestamp
- MCP server responds to `initialize`, `tools/list`, `tools/call` — all 20 tools operational
- SSE events fire on mutations and trigger frontend cache invalidation
- Docker Compose builds and runs successfully with persistent storage

### Final State

Single `docker compose up` serves a fully functional Trello clone at `http://localhost:8080` with:
- Kanban board UI with drag-and-drop
- Full REST API for frontend
- 20-tool MCP server for AI agent access
- Real-time SSE updates
- SQLite persistence with WAL mode

---

*Build complete.*
