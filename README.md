# Cielo

A Kanban board designed for AI agent orchestration via the Model Context Protocol (MCP).

## Overview

Cielo gives AI agents a structured way to coordinate work. Instead of passing tasks through unstructured text, agents interact with a Kanban board through 20 MCP tools — creating cards, moving them between lists, tracking dependencies, and logging activity.

The problem: multi-agent workflows need shared state. Agents need to claim tasks, signal blockers, and see what others are doing. Chat threads and flat task lists don't provide the spatial organization or dependency tracking that complex workflows require.

Cielo solves this with a full Kanban board backed by a REST API and MCP server. Agents read and write cards, manage dependencies, and subscribe to real-time updates — all through a single binary that serves the API, the MCP endpoint, and a React frontend.

## Features

### Kanban Board

- Multiple boards with ordered lists and drag-and-drop cards
- Card priority levels (low, medium, high, critical) and statuses (unassigned, assigned, in_progress, blocked, done)
- Labels with custom colors, due dates, and rich descriptions

### Agent Orchestration

- 20 MCP tools for full board interaction
- Card assignment and status tracking per agent
- Card-to-card dependency graphs (blocker/dependent relationships)
- Activity log with actor attribution for audit trails

### Real-time Updates

- Server-Sent Events (SSE) per board
- Automatic UI refresh on card, list, label, and activity changes

### Search & Filtering

- Search cards by title, assignee, status, or label
- Board-scoped activity logs with configurable limits

## Architecture

```text
┌──────────────────────────────────────────────────────┐
│                React Frontend (Vite)                 │
│         React Query · dnd-kit · Tailwind CSS         │
└──────────────────────┬───────────────────────────────┘
                       │ HTTP
┌──────────────────────▼───────────────────────────────┐
│                   Fiber HTTP Server                  │
│                                                      │
│  /api/v1/*  REST endpoints                           │
│  /mcp       MCP JSON-RPC 2.0                         │
│  /events    Server-Sent Events                       │
└──────────────────────┬───────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────┐
│                   Service Layer                      │
│        Business logic · Activity logging             │
│              Event bus (pub/sub)                      │
└──────────────────────┬───────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────┐
│                  SQLite (WAL mode)                    │
│     boards · lists · cards · labels · activity_log   │
│           card_dependencies · card_labels            │
└──────────────────────────────────────────────────────┘
```

Everything compiles into a single binary. The frontend is embedded at build time and served as static files.

## Getting Started

### Docker Compose (quickest)

```bash
docker compose up -d
```

The app is available at `http://localhost:8080`. Data persists in a named volume.

### Local Development

Prerequisites: Go 1.26+, Node 20+

```bash
# Build and run
make build
make run

# Run tests
make test
```

For frontend development with hot reload:

```bash
cd web
npm install
npm run dev
```

The Vite dev server proxies API requests to the Go backend on `:8080`.

### Building from Source

```bash
# Build frontend
cd web && npm install && npm run build && cd ..

# Build backend (embeds web/dist/)
go build -o bin/cielo ./cmd/cielo

# Run
./bin/cielo
```

## Configuration

| Variable | Description | Default |
| --- | --- | --- |
| `CIELO_HTTP_ADDR` | HTTP server listen address | `:8080` |
| `CIELO_DB_PATH` | SQLite database file path | `cielo.db` |

## API Reference

All endpoints are prefixed with `/api/v1`.

### Boards

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/boards` | List all boards |
| `POST` | `/boards` | Create a board |
| `GET` | `/boards/:id` | Get board with lists and cards |
| `PUT` | `/boards/:id` | Update board |
| `DELETE` | `/boards/:id` | Delete board |

### Lists

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/boards/:boardId/lists` | Create a list |
| `PUT` | `/lists/:id` | Update list (name, position) |
| `DELETE` | `/lists/:id` | Delete list |

### Cards

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/lists/:listId/cards` | Create a card |
| `GET` | `/cards/:id` | Get card with full details |
| `PUT` | `/cards/:id` | Update card fields |
| `DELETE` | `/cards/:id` | Delete card |
| `PUT` | `/cards/:id/move` | Move card to a different list/position |
| `PUT` | `/cards/:id/assign` | Assign or unassign a card |

### Labels

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/boards/:boardId/labels` | List board labels |
| `POST` | `/boards/:boardId/labels` | Create a label |
| `PUT` | `/labels/:id` | Update label |
| `DELETE` | `/labels/:id` | Delete label |
| `POST` | `/cards/:id/labels` | Add label to card |
| `DELETE` | `/cards/:id/labels/:labelId` | Remove label from card |

### Dependencies

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/cards/:id/dependencies` | Add a dependency |
| `DELETE` | `/cards/:id/dependencies/:depId` | Remove a dependency |

### Activity

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/cards/:id/activity` | Card activity log |
| `GET` | `/boards/:boardId/activity` | Board activity log |

### Search

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/boards/:boardId/search` | Search cards (`q`, `assignee`, `status`, `label`) |

### Real-time Events

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/boards/:boardId/events` | SSE stream for board changes |

## MCP Tools

Connect to the MCP endpoint at `/mcp` (JSON-RPC 2.0, protocol version `2025-11-25`).

### Read Tools

| Tool | Description |
| --- | --- |
| `list_boards` | List all boards |
| `get_board` | Get board with lists and card counts |
| `list_lists` | Get all lists for a board with their cards |
| `get_card` | Get full card detail including labels, dependencies, activity |
| `search_cards` | Search cards by title, assignee, status, or label |
| `get_card_dependencies` | Get blockers and dependents for a card |
| `get_activity_log` | Get activity history for a card or board |

### Write Tools

| Tool | Description |
| --- | --- |
| `create_board` | Create a new board |
| `create_list` | Add a list to a board |
| `create_card` | Create a card in a list |
| `update_card` | Update card fields (title, description, status, priority, assignee) |
| `move_card` | Move a card to a different list and/or position |
| `assign_card` | Assign or unassign a card |
| `add_comment` | Add a comment to a card's activity log |
| `add_dependency` | Create a dependency between two cards |
| `remove_dependency` | Remove a dependency between two cards |
| `add_label_to_card` | Tag a card with a label |
| `remove_label_from_card` | Remove a label from a card |
| `delete_card` | Delete a card |
| `delete_list` | Delete a list and its cards |

## Project Structure

```tree
cmd/cielo/           Entry point — wires up config, database, services, and HTTP server
internal/
  api/               HTTP handlers and routing (Fiber)
    router.go        Route definitions
    board.go         Board endpoints
    card.go          Card endpoints
    list.go          List endpoints
    sse.go           Server-Sent Events handler
    middleware.go    CORS and request logging
  config/            Environment-based configuration
  event/             Pub/sub event bus for real-time updates
  mcp/               MCP server and tool definitions (JSON-RPC 2.0)
  model/             Data models (Board, List, Card, Label, Activity)
  service/           Business logic and activity logging
  store/             SQLite persistence and store interface
migrations/          SQL migration files (embedded at build time)
web/                 React frontend (Vite + TypeScript + Tailwind)
  src/api/           API client and React Query hooks
  src/components/    UI components (CardTile, ListColumn, CardDetail)
  src/pages/         Page components (BoardList, BoardView)
docs/                Design documents and transcripts
```

## Tech Stack

| Layer | Technology |
| --- | --- |
| Language | Go 1.26 |
| HTTP | Fiber v3 |
| Database | SQLite (WAL mode) via modernc.org/sqlite |
| Frontend | React 19, TypeScript 5.9, Vite 7 |
| Styling | Tailwind CSS 4 |
| State | TanStack React Query 5 |
| Drag & Drop | dnd-kit |
| Routing | React Router 7 |
| Container | Docker (multi-stage build) |

## License

[Apache License 2.0](LICENSE)
