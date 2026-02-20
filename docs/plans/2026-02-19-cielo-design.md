# Cielo — Technical Design

A Trello-style Kanban board for AI agent orchestration. Single Go binary serving both a REST API (Fiber v3) and an MCP server (JSON-RPC 2.0), with a React + Vite frontend.

## Architecture Overview

```
┌─────────────────────────────────────┐
│         Single Go Binary            │
│                                     │
│  ┌───────────┐   ┌───────────────┐  │
│  │ Fiber v3  │   │ JSON-RPC MCP  │  │
│  │ HTTP API  │   │ Server        │  │
│  │ :8080     │   │ (on :8080)    │  │
│  └─────┬─────┘   └──────┬────────┘  │
│        │                │           │
│        └───────┬────────┘           │
│          ┌─────▼─────┐             │
│          │  Service   │             │
│          │   Layer    │             │
│          └─────┬─────┘             │
│          ┌─────▼─────┐             │
│          │  SQLite    │             │
│          └───────────┘             │
└─────────────────────────────────────┘
```

Both the HTTP API and MCP server share a single service layer and database. All mutations flow through the service layer, which handles validation, persistence, and event emission — ensuring consistent behavior regardless of entry point.

## Data Model

All entity IDs use UUIDv7 for time-ordered, B-tree-friendly indexing.

### Board
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | Primary key |
| name | TEXT | Required |
| description | TEXT | Optional |
| created_at | DATETIME | Auto-set |
| updated_at | DATETIME | Auto-updated |

### List
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | Primary key |
| board_id | UUIDv7 | FK → Board, ON DELETE CASCADE |
| name | TEXT | Required |
| position | INTEGER | Ordering within board |
| created_at | DATETIME | Auto-set |
| updated_at | DATETIME | Auto-updated |

### Card
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | Primary key |
| list_id | UUIDv7 | FK → List, ON DELETE CASCADE |
| title | TEXT | Required |
| description | TEXT | Markdown, optional |
| position | INTEGER | Ordering within list |
| assignee | TEXT | Agent name or user, optional |
| status | TEXT | Enum: unassigned, assigned, in_progress, blocked, done |
| priority | TEXT | Enum: low, medium, high, critical |
| due_date | DATETIME | Optional |
| created_at | DATETIME | Auto-set |
| updated_at | DATETIME | Auto-updated |

### CardDependency
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | Primary key |
| card_id | UUIDv7 | FK → Card (the blocked card) |
| depends_on_card_id | UUIDv7 | FK → Card (the blocking card) |
| created_at | DATETIME | Auto-set |

Unique constraint on (card_id, depends_on_card_id).

### Label
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | Primary key |
| board_id | UUIDv7 | FK → Board, ON DELETE CASCADE |
| name | TEXT | e.g., "coding", "research", "review" |
| color | TEXT | Hex color string |

### CardLabel (join table)
| Column | Type | Notes |
|--------|------|-------|
| card_id | UUIDv7 | FK → Card, ON DELETE CASCADE |
| label_id | UUIDv7 | FK → Label, ON DELETE CASCADE |

Primary key on (card_id, label_id).

### ActivityLog
| Column | Type | Notes |
|--------|------|-------|
| id | UUIDv7 | Primary key |
| card_id | UUIDv7 | FK → Card, ON DELETE CASCADE |
| actor | TEXT | Agent name or "user" |
| action | TEXT | Enum: created, moved, assigned, unassigned, status_changed, comment, dependency_added, dependency_removed, label_added, label_removed |
| detail | TEXT | JSON blob — before/after state, comment text, etc. |
| created_at | DATETIME | Auto-set |

## Backend Architecture

### Project Structure

```
cielo/
├── cmd/
│   └── cielo/
│       └── main.go              # Entry point — starts Fiber + MCP server
├── internal/
│   ├── config/
│   │   └── config.go            # App config (ports, DB path, etc.)
│   ├── model/
│   │   └── models.go            # Struct definitions matching data model
│   ├── store/
│   │   ├── store.go             # Store interface (repository pattern)
│   │   └── sqlite.go            # SQLite implementation
│   ├── service/
│   │   └── service.go           # Business logic layer (shared by HTTP + MCP)
│   ├── event/
│   │   └── bus.go               # In-process event bus for SSE broadcasting
│   ├── api/
│   │   ├── router.go            # Fiber v3 route definitions
│   │   ├── board.go             # Board handlers
│   │   ├── list.go              # List handlers
│   │   ├── card.go              # Card handlers
│   │   └── middleware.go        # CORS, logging, error handling
│   └── mcp/
│       ├── server.go            # JSON-RPC dispatcher + MCP protocol handling
│       ├── tools.go             # Tool definitions (read + write tools)
│       └── transport.go         # Streamable HTTP transport (POST /mcp)
├── web/                         # React + Vite frontend
├── migrations/
│   └── 001_initial.sql          # SQLite schema
├── go.mod
├── go.sum
└── Makefile
```

### Layering

```
HTTP Handlers (Fiber)     MCP Tool Handlers (JSON-RPC)
        │                          │
        └──────────┬───────────────┘
                   │
            Service Layer          ← validation, business rules, event emission
                   │
              Store Interface      ← repository pattern
                   │
              SQLite (modernc.org/sqlite)
```

- **Store interface**: All DB access goes through a `Store` interface — testable and swappable.
- **Service layer**: Single source of truth for business logic. Both HTTP and MCP handlers call the same service methods.
- **Event bus**: In-process pub/sub (channel-based). Service layer publishes events on mutation; SSE endpoint subscribes and pushes to connected clients.
- **SQLite driver**: `modernc.org/sqlite` — pure Go, no CGo dependency.

## MCP Server

### Protocol

Implements the Model Context Protocol (2025-11-25 spec) using Streamable HTTP transport:
- Single `POST /mcp` endpoint accepting JSON-RPC 2.0 messages
- Supports `initialize`, `tools/list`, and `tools/call` methods
- Tool definitions include JSON Schema for `inputSchema`

### Read Tools

| Tool | Description | Key Params |
|------|-------------|------------|
| `list_boards` | List all boards | _(none)_ |
| `get_board` | Get board with lists and card counts | `board_id` |
| `list_lists` | Get all lists for a board with cards | `board_id` |
| `get_card` | Get full card detail (labels, deps, activity) | `card_id` |
| `search_cards` | Search cards by title, assignee, status, label | `board_id`, `query?`, `assignee?`, `status?`, `label?` |
| `get_card_dependencies` | Get blockers and dependents for a card | `card_id` |
| `get_activity_log` | Get activity history for a card or board | `card_id?`, `board_id?`, `limit?` |

### Write Tools

| Tool | Description | Key Params |
|------|-------------|------------|
| `create_board` | Create a new board | `name`, `description?` |
| `create_list` | Add a list to a board | `board_id`, `name`, `position?` |
| `create_card` | Create a card in a list | `list_id`, `title`, `description?`, `assignee?`, `priority?`, `labels?` |
| `move_card` | Move card to a different list/position | `card_id`, `list_id`, `position?` |
| `update_card` | Update card fields | `card_id`, `title?`, `description?`, `assignee?`, `status?`, `priority?`, `due_date?` |
| `assign_card` | Assign/unassign a card | `card_id`, `assignee` |
| `add_comment` | Add comment to card activity log | `card_id`, `actor`, `text` |
| `add_dependency` | Create dependency between cards | `card_id`, `depends_on_card_id` |
| `remove_dependency` | Remove a dependency | `card_id`, `depends_on_card_id` |
| `add_label_to_card` | Tag a card with a label | `card_id`, `label_id` |
| `remove_label_from_card` | Remove label from card | `card_id`, `label_id` |
| `delete_card` | Delete a card | `card_id` |
| `delete_list` | Delete a list and its cards | `list_id` |

All write tools:
- Accept an `actor` parameter for activity log attribution
- Return the updated entity
- Emit events to the SSE bus

## HTTP API

All routes prefixed with `/api/v1`.

### Boards
- `GET /api/v1/boards` — list all boards
- `POST /api/v1/boards` — create board
- `GET /api/v1/boards/:id` — get board detail (with lists + cards)
- `PUT /api/v1/boards/:id` — update board
- `DELETE /api/v1/boards/:id` — delete board

### Lists
- `POST /api/v1/boards/:boardId/lists` — create list
- `PUT /api/v1/lists/:id` — update list (name, position)
- `DELETE /api/v1/lists/:id` — delete list

### Cards
- `POST /api/v1/lists/:listId/cards` — create card
- `GET /api/v1/cards/:id` — get card detail
- `PUT /api/v1/cards/:id` — update card
- `DELETE /api/v1/cards/:id` — delete card
- `PUT /api/v1/cards/:id/move` — move card (list + position)
- `PUT /api/v1/cards/:id/assign` — assign/unassign

### Dependencies
- `POST /api/v1/cards/:id/dependencies` — add dependency
- `DELETE /api/v1/cards/:id/dependencies/:depId` — remove dependency

### Labels
- `GET /api/v1/boards/:boardId/labels` — list labels
- `POST /api/v1/boards/:boardId/labels` — create label
- `PUT /api/v1/labels/:id` — update label
- `DELETE /api/v1/labels/:id` — delete label
- `POST /api/v1/cards/:id/labels` — add label to card
- `DELETE /api/v1/cards/:id/labels/:labelId` — remove label from card

### Activity & Search
- `GET /api/v1/cards/:id/activity` — card activity log
- `GET /api/v1/boards/:boardId/activity` — board activity log
- `GET /api/v1/boards/:boardId/search?q=&assignee=&status=&label=` — search cards

### SSE
- `GET /api/v1/boards/:boardId/events` — SSE stream for real-time board updates

## Real-time Events (SSE)

### Event Bus

```
Service Layer (mutation occurs)
        │
        ▼
   Event Bus (in-process, channel-based)
        │
        ├──▶ SSE Subscriber (board-123)  ──▶  Browser client
        ├──▶ SSE Subscriber (board-123)  ──▶  Browser client
        └──▶ SSE Subscriber (board-456)  ──▶  Browser client
```

### Event Types

| Event | Payload | Triggered by |
|-------|---------|--------------|
| `card.created` | Full card object | New card via API or MCP |
| `card.updated` | Card ID + changed fields | Any card field update |
| `card.moved` | Card ID, from list/pos, to list/pos | Drag-drop or MCP move |
| `card.deleted` | Card ID | Deletion |
| `list.created` | Full list object | New list |
| `list.updated` | List ID + changed fields | Rename, reorder |
| `list.deleted` | List ID | Deletion |
| `label.created` | Full label object | New label |
| `label.updated` | Label ID + changed fields | Update |
| `label.deleted` | Label ID | Deletion |
| `activity.new` | Activity log entry | Any logged action |

### Design Details

- **Scoped by board**: Clients subscribe to `/api/v1/boards/:boardId/events` and only receive events for that board.
- **Fan-out per board**: Event bus maintains a map of `boardId → []subscriber`. Each subscriber is a channel drained by the SSE handler.
- **Reconnect support**: Events include a monotonically increasing sequence ID. Clients send `Last-Event-ID` on reconnect to catch up from an in-memory ring buffer per board.

## Frontend

### Stack
- React 19 + TypeScript
- Vite (build + dev server)
- TanStack Query (server state management)
- dnd-kit (drag-and-drop)
- Tailwind CSS v4 (styling)
- React Router (navigation)

### Routes

| Path | View |
|------|------|
| `/` | Board list (home) |
| `/boards/:boardId` | Board view (main Kanban) |
| `/boards/:boardId/cards/:cardId` | Card detail (modal overlay) |

### Component Tree

```
App
├── BoardList           — Grid of boards on home page
├── BoardView           — The main Kanban board
│   ├── ListColumn      — A single column (droppable zone)
│   │   ├── ListHeader  — Column name, card count, add card button
│   │   └── CardTile    — Compact card preview (draggable)
│   │       ├── Labels  — Color chips
│   │       ├── Assignee
│   │       └── Priority indicator
│   └── AddListButton   — Add new column
├── CardDetail          — Modal with full card info
│   ├── Description     — Markdown editor
│   ├── AssigneeSelect
│   ├── StatusSelect
│   ├── PrioritySelect
│   ├── LabelManager
│   ├── DependencyList  — Shows blockers/dependents
│   └── ActivityFeed    — Chronological activity log
└── Shared
    ├── SSEProvider     — Context provider managing EventSource connection
    └── Layout          — Nav bar, board switcher
```

### SSE Integration
- `SSEProvider` wraps the board view, connects to `/api/v1/boards/:boardId/events`
- On receiving events, invalidates relevant TanStack Query cache keys for automatic refetch
- Handles reconnection with `Last-Event-ID`

### Optimistic Updates
- Drag-and-drop moves update UI immediately, then sync with server
- On failure, UI reverts to server state
