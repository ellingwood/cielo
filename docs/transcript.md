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

*Next up: Implementation planning and build.*
