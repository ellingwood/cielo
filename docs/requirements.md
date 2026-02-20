# Cielo — Requirements

A Trello-style Kanban board for orchestrating projects worked on by AI agents and humans.

## 1. Core Kanban Features

### 1.1 Boards
- R-101: Users can create, rename, and delete boards
- R-102: Each board has a name and optional description
- R-103: Home page displays all boards in a grid

### 1.2 Lists (Columns)
- R-201: Users can create, rename, reorder, and delete lists within a board
- R-202: Lists are ordered by a position value
- R-203: Deleting a list deletes all cards within it

### 1.3 Cards
- R-301: Users can create, edit, and delete cards within a list
- R-302: Cards have: title, description (markdown), assignee, status, priority, and optional due date
- R-303: Cards are ordered by position within their list
- R-304: Cards can be moved between lists via drag-and-drop (frontend) or API/MCP calls
- R-305: Card status enum: unassigned, assigned, in_progress, blocked, done
- R-306: Card priority enum: low, medium, high, critical

### 1.4 Labels
- R-401: Boards have a set of labels, each with a name and color
- R-402: Cards can have zero or more labels
- R-403: Labels are used to tag agent type (e.g., coding, research, review)

## 2. AI Agent Orchestration

### 2.1 Agent Assignment
- R-501: Cards can be assigned to an agent (identified by name string)
- R-502: Agents can self-assign cards (pick up work autonomously)
- R-503: Humans can assign cards to specific agents
- R-504: Agents can unassign themselves from cards

### 2.2 Card Dependencies
- R-601: Cards can depend on other cards (A is blocked by B)
- R-602: Dependencies are directional — a card can have multiple blockers and multiple dependents
- R-603: Dependencies are queryable (get all blockers for a card, get all cards blocked by a card)

### 2.3 Activity Log
- R-701: Every mutation to a card is logged with: actor (agent name or "user"), action type, detail (JSON), and timestamp
- R-702: Action types: created, moved, assigned, unassigned, status_changed, comment, dependency_added, dependency_removed, label_added, label_removed
- R-703: Agents can add comments to cards via the activity log
- R-704: Activity log is queryable per card and per board

### 2.4 Search
- R-801: Cards are searchable by title text, assignee, status, and label within a board

## 3. MCP Server (JSON-RPC)

### 3.1 Protocol
- R-901: MCP server implements the Model Context Protocol (2025-11-25 spec)
- R-902: Transport: Streamable HTTP — single POST /mcp endpoint accepting JSON-RPC 2.0 messages
- R-903: Server exposes tools via `tools/list` and executes them via `tools/call`

### 3.2 Read Tools
- R-1001: `list_boards` — list all boards
- R-1002: `get_board` — get board with lists and card counts
- R-1003: `list_lists` — get all lists for a board with their cards
- R-1004: `get_card` — get full card detail including labels, dependencies, and activity
- R-1005: `search_cards` — search cards by title, assignee, status, label
- R-1006: `get_card_dependencies` — get blockers and dependents for a card
- R-1007: `get_activity_log` — get activity history for a card or board

### 3.3 Write Tools
- R-1101: `create_board` — create a new board
- R-1102: `create_list` — add a list to a board
- R-1103: `create_card` — create a card in a list
- R-1104: `move_card` — move card to a different list and/or position
- R-1105: `update_card` — update card fields
- R-1106: `assign_card` — assign or unassign a card
- R-1107: `add_comment` — add a comment to a card
- R-1108: `add_dependency` — create a dependency between cards
- R-1109: `remove_dependency` — remove a dependency
- R-1110: `add_label_to_card` — tag a card with a label
- R-1111: `remove_label_from_card` — remove a label from a card
- R-1112: `delete_card` — delete a card
- R-1113: `delete_list` — delete a list and its cards

### 3.4 Agent Interaction
- R-1201: All write tools accept an `actor` parameter for activity log attribution
- R-1202: Write tools return the updated entity after mutation
- R-1203: Write tools emit events to the SSE bus for real-time UI updates

## 4. HTTP API

### 4.1 REST Endpoints
- R-1301: Full CRUD REST API under `/api/v1` for boards, lists, cards, labels, dependencies, and activity
- R-1302: API routes follow RESTful conventions with proper HTTP methods and status codes
- R-1303: API goes through the same service layer as MCP, ensuring consistent behavior

### 4.2 Real-time (SSE)
- R-1401: `GET /api/v1/boards/:boardId/events` provides an SSE stream of board changes
- R-1402: Events are scoped per board — clients only receive events for the board they subscribe to
- R-1403: Event types: card.created, card.updated, card.moved, card.deleted, list.created, list.updated, list.deleted, label.created, label.updated, label.deleted, activity.new
- R-1404: Events include a sequence ID for reconnection support via Last-Event-ID

## 5. Frontend

### 5.1 Core UI
- R-1501: Board list page showing all boards
- R-1502: Board view with draggable columns and cards (Kanban layout)
- R-1503: Card detail modal with full editing capabilities
- R-1504: Drag-and-drop card movement between lists with optimistic updates

### 5.2 Card Detail
- R-1601: Markdown description editor
- R-1602: Assignee, status, priority, and due date fields
- R-1603: Label management (add/remove labels)
- R-1604: Dependency list showing blockers and dependents
- R-1605: Chronological activity feed

### 5.3 Real-time
- R-1701: Frontend subscribes to SSE stream for the active board
- R-1702: UI updates automatically when agents modify the board via MCP
- R-1703: Handles SSE reconnection with Last-Event-ID

## 6. Technical Constraints

- R-1801: Backend: Go 1.26
- R-1802: HTTP framework: Go Fiber v3
- R-1803: Database: SQLite (via modernc.org/sqlite, pure Go)
- R-1804: All entity IDs: UUIDv7 for time-ordered, index-efficient identifiers
- R-1805: Single binary deployment — HTTP API and MCP server in one process
- R-1806: Frontend: React 19 + TypeScript + Vite
- R-1807: Frontend styling: Tailwind CSS v4
- R-1808: Frontend drag-and-drop: dnd-kit
- R-1809: Frontend server state: TanStack Query
