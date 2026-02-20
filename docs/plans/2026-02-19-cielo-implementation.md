# Cielo Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Trello-style Kanban board with a Go backend (Fiber v3 + MCP server) and React frontend for AI agent orchestration.

**Architecture:** Single Go binary serving both a REST API (Fiber v3) and a JSON-RPC MCP server (Streamable HTTP transport), sharing a common service layer backed by SQLite. React + Vite frontend with real-time SSE updates.

**Tech Stack:** Go 1.26, Fiber v3, modernc.org/sqlite, UUIDv7, React 19, TypeScript, Vite, Tailwind CSS v4, TanStack Query, dnd-kit, React Router

**Reference Docs:**
- Design: `docs/plans/2026-02-19-cielo-design.md`
- Requirements: `docs/requirements.md`

---

## Task 1: Go Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/cielo/main.go`
- Create: `internal/config/config.go`
- Create: `Makefile`

**Step 1: Initialize Go module**

Run:
```bash
cd /Users/aellingwood/dev/personal/cielo
go mod init github.com/aellingwood/cielo
```

**Step 2: Create directory structure**

Run:
```bash
mkdir -p cmd/cielo internal/{config,model,store,service,event,api,mcp} migrations web
```

**Step 3: Write config**

Create `internal/config/config.go`:
```go
package config

import (
	"os"
)

type Config struct {
	HTTPAddr string
	DBPath   string
}

func Load() *Config {
	return &Config{
		HTTPAddr: envOr("CIELO_HTTP_ADDR", ":8080"),
		DBPath:   envOr("CIELO_DB_PATH", "cielo.db"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

**Step 4: Write minimal main.go**

Create `cmd/cielo/main.go`:
```go
package main

import (
	"fmt"
	"log"

	"github.com/aellingwood/cielo/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Cielo starting on %s (db: %s)\n", cfg.HTTPAddr, cfg.DBPath)
	log.Println("Server not yet implemented")
}
```

**Step 5: Write Makefile**

Create `Makefile`:
```makefile
.PHONY: build run test clean

build:
	go build -o bin/cielo ./cmd/cielo

run: build
	./bin/cielo

test:
	go test ./... -v

clean:
	rm -rf bin/
```

**Step 6: Install initial dependencies**

Run:
```bash
go mod tidy
```

**Step 7: Verify build**

Run: `make build && make run`
Expected: Prints "Cielo starting on :8080 (db: cielo.db)"

**Step 8: Commit**

```bash
git add -A
git commit -m "feat: scaffold Go project with config, main entry point, and Makefile"
```

---

## Task 2: SQLite Schema & Migration

**Files:**
- Create: `migrations/001_initial.sql`
- Create: `internal/store/migrate.go`

**Step 1: Write migration SQL**

Create `migrations/001_initial.sql`:
```sql
-- Cielo initial schema
-- All IDs are UUIDv7 stored as TEXT (36-char string)

PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

CREATE TABLE IF NOT EXISTS boards (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE IF NOT EXISTS lists (
    id         TEXT PRIMARY KEY,
    board_id   TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    position   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_lists_board_id ON lists(board_id);

CREATE TABLE IF NOT EXISTS cards (
    id          TEXT PRIMARY KEY,
    list_id     TEXT NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    position    INTEGER NOT NULL DEFAULT 0,
    assignee    TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'unassigned' CHECK(status IN ('unassigned','assigned','in_progress','blocked','done')),
    priority    TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low','medium','high','critical')),
    due_date    TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_cards_list_id ON cards(list_id);

CREATE TABLE IF NOT EXISTS card_dependencies (
    id                 TEXT PRIMARY KEY,
    card_id            TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    depends_on_card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE(card_id, depends_on_card_id)
);

CREATE INDEX idx_card_deps_card_id ON card_dependencies(card_id);
CREATE INDEX idx_card_deps_depends_on ON card_dependencies(depends_on_card_id);

CREATE TABLE IF NOT EXISTS labels (
    id       TEXT PRIMARY KEY,
    board_id TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name     TEXT NOT NULL,
    color    TEXT NOT NULL DEFAULT '#6b7280'
);

CREATE INDEX idx_labels_board_id ON labels(board_id);

CREATE TABLE IF NOT EXISTS card_labels (
    card_id  TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    label_id TEXT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (card_id, label_id)
);

CREATE TABLE IF NOT EXISTS activity_log (
    id         TEXT PRIMARY KEY,
    card_id    TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    actor      TEXT NOT NULL,
    action     TEXT NOT NULL CHECK(action IN ('created','moved','assigned','unassigned','status_changed','comment','dependency_added','dependency_removed','label_added','label_removed')),
    detail     TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_activity_card_id ON activity_log(card_id);
```

**Step 2: Write migration runner**

Create `internal/store/migrate.go`:
```go
package store

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed ../../migrations/*.sql
// NOTE: This embed path won't work from internal/store.
// Instead, embed from the migrations package or pass SQL as string.
// See implementation note below.

// RunMigrations executes all migration files against the database.
// For now, we embed the SQL directly since Go embed requires the
// files to be in or below the embedding package's directory.
func RunMigrations(db *sql.DB) error {
	// Read and execute migrations/001_initial.sql
	// Implementation: either embed from cmd/cielo and pass down,
	// or use a migrations package at the project root.
	return nil
}
```

**Implementation note:** Go's `//go:embed` requires files to be within the package directory tree. Two options:
1. Create a top-level `migrations` package with an `embed.go` that exports the embedded filesystem
2. Embed from `cmd/cielo/main.go` and pass the `embed.FS` to the store

Preferred approach — create `migrations/migrations.go`:
```go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

Then `internal/store/migrate.go` imports `migrations.FS` and iterates over SQL files.

**Step 3: Write migration test**

Create `internal/store/migrate_test.go`:
```go
package store_test

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/aellingwood/cielo/internal/store"
)

func TestRunMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := store.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify tables exist
	tables := []string{"boards", "lists", "cards", "card_dependencies", "labels", "card_labels", "activity_log"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}
```

**Step 4: Run test to verify it fails**

Run: `go test ./internal/store/ -v -run TestRunMigrations`
Expected: FAIL (RunMigrations is a no-op)

**Step 5: Implement RunMigrations**

Complete the implementation so it reads from `migrations.FS`, sorts files, and executes them.

**Step 6: Install SQLite dependency**

Run:
```bash
go get modernc.org/sqlite
go mod tidy
```

**Step 7: Run test to verify it passes**

Run: `go test ./internal/store/ -v -run TestRunMigrations`
Expected: PASS

**Step 8: Commit**

```bash
git add -A
git commit -m "feat: add SQLite schema migration with all 7 tables"
```

---

## Task 3: Models & UUIDv7

**Files:**
- Create: `internal/model/models.go`
- Create: `internal/model/models_test.go`

**Step 1: Install UUIDv7 library**

Run:
```bash
go get github.com/google/uuid
```

The `github.com/google/uuid` package supports UUIDv7 via `uuid.NewV7()`.

**Step 2: Write model structs**

Create `internal/model/models.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// NewID generates a UUIDv7 string.
func NewID() string {
	return uuid.Must(uuid.NewV7()).String()
}

type Board struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type List struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Cards     []Card    `json:"cards,omitempty"`
}

type Card struct {
	ID           string         `json:"id"`
	ListID       string         `json:"list_id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Position     int            `json:"position"`
	Assignee     string         `json:"assignee"`
	Status       string         `json:"status"`
	Priority     string         `json:"priority"`
	DueDate      *time.Time     `json:"due_date,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Labels       []Label        `json:"labels,omitempty"`
	Dependencies []Card         `json:"dependencies,omitempty"`
	Dependents   []Card         `json:"dependents,omitempty"`
	Activity     []ActivityLog  `json:"activity,omitempty"`
}

type CardDependency struct {
	ID              string    `json:"id"`
	CardID          string    `json:"card_id"`
	DependsOnCardID string    `json:"depends_on_card_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type Label struct {
	ID      string `json:"id"`
	BoardID string `json:"board_id"`
	Name    string `json:"name"`
	Color   string `json:"color"`
}

type ActivityLog struct {
	ID        string    `json:"id"`
	CardID    string    `json:"card_id"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

// Valid status values
const (
	StatusUnassigned = "unassigned"
	StatusAssigned   = "assigned"
	StatusInProgress = "in_progress"
	StatusBlocked    = "blocked"
	StatusDone       = "done"
)

// Valid priority values
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Valid action values
const (
	ActionCreated           = "created"
	ActionMoved             = "moved"
	ActionAssigned          = "assigned"
	ActionUnassigned        = "unassigned"
	ActionStatusChanged     = "status_changed"
	ActionComment           = "comment"
	ActionDependencyAdded   = "dependency_added"
	ActionDependencyRemoved = "dependency_removed"
	ActionLabelAdded        = "label_added"
	ActionLabelRemoved      = "label_removed"
)
```

**Step 3: Write model test**

Create `internal/model/models_test.go`:
```go
package model_test

import (
	"testing"

	"github.com/aellingwood/cielo/internal/model"
)

func TestNewID_IsValidUUIDv7(t *testing.T) {
	id := model.NewID()
	if len(id) != 36 {
		t.Errorf("expected UUID length 36, got %d: %s", len(id), id)
	}
	// UUIDv7 has version nibble '7' at position 14
	if id[14] != '7' {
		t.Errorf("expected version 7 at position 14, got %c in %s", id[14], id)
	}
}

func TestNewID_IsUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := model.NewID()
		if seen[id] {
			t.Fatalf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/model/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add domain models and UUIDv7 ID generation"
```

---

## Task 4: Store Interface

**Files:**
- Create: `internal/store/store.go`

**Step 1: Define the Store interface**

Create `internal/store/store.go`:
```go
package store

import (
	"context"

	"github.com/aellingwood/cielo/internal/model"
)

// Store defines the data access interface for Cielo.
// All methods accept a context for cancellation.
type Store interface {
	// Boards
	CreateBoard(ctx context.Context, board *model.Board) error
	GetBoard(ctx context.Context, id string) (*model.Board, error)
	ListBoards(ctx context.Context) ([]model.Board, error)
	UpdateBoard(ctx context.Context, board *model.Board) error
	DeleteBoard(ctx context.Context, id string) error

	// Lists
	CreateList(ctx context.Context, list *model.List) error
	GetList(ctx context.Context, id string) (*model.List, error)
	ListListsByBoard(ctx context.Context, boardID string) ([]model.List, error)
	UpdateList(ctx context.Context, list *model.List) error
	DeleteList(ctx context.Context, id string) error

	// Cards
	CreateCard(ctx context.Context, card *model.Card) error
	GetCard(ctx context.Context, id string) (*model.Card, error)
	ListCardsByList(ctx context.Context, listID string) ([]model.Card, error)
	UpdateCard(ctx context.Context, card *model.Card) error
	MoveCard(ctx context.Context, cardID, targetListID string, position int) error
	DeleteCard(ctx context.Context, id string) error
	SearchCards(ctx context.Context, boardID string, query, assignee, status, label string) ([]model.Card, error)

	// Dependencies
	AddDependency(ctx context.Context, dep *model.CardDependency) error
	RemoveDependency(ctx context.Context, cardID, dependsOnCardID string) error
	GetDependencies(ctx context.Context, cardID string) ([]model.Card, error)   // cards that block this card
	GetDependents(ctx context.Context, cardID string) ([]model.Card, error)     // cards blocked by this card

	// Labels
	CreateLabel(ctx context.Context, label *model.Label) error
	GetLabel(ctx context.Context, id string) (*model.Label, error)
	ListLabelsByBoard(ctx context.Context, boardID string) ([]model.Label, error)
	UpdateLabel(ctx context.Context, label *model.Label) error
	DeleteLabel(ctx context.Context, id string) error
	AddLabelToCard(ctx context.Context, cardID, labelID string) error
	RemoveLabelFromCard(ctx context.Context, cardID, labelID string) error
	GetLabelsForCard(ctx context.Context, cardID string) ([]model.Label, error)

	// Activity
	CreateActivity(ctx context.Context, entry *model.ActivityLog) error
	ListActivityByCard(ctx context.Context, cardID string, limit int) ([]model.ActivityLog, error)
	ListActivityByBoard(ctx context.Context, boardID string, limit int) ([]model.ActivityLog, error)
}
```

**Step 2: Commit**

```bash
git add internal/store/store.go
git commit -m "feat: define Store interface (repository pattern)"
```

---

## Task 5: SQLite Store Implementation — Boards

**Files:**
- Create: `internal/store/sqlite.go`
- Create: `internal/store/sqlite_test.go`

**Step 1: Write board store tests**

Create `internal/store/sqlite_test.go` with a test helper that opens an in-memory DB, runs migrations, and returns a `*SQLiteStore`. Write tests for:
- `TestCreateBoard` — create and verify it can be retrieved
- `TestListBoards` — create two boards, list returns both
- `TestUpdateBoard` — create, update name, verify change
- `TestDeleteBoard` — create, delete, verify gone

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/store/ -v -run TestCreate`
Expected: FAIL (SQLiteStore not implemented)

**Step 3: Implement SQLiteStore with board methods**

Create `internal/store/sqlite.go` with:
- `SQLiteStore` struct wrapping `*sql.DB`
- `NewSQLiteStore(db *sql.DB) *SQLiteStore`
- Board CRUD methods using prepared statements
- Time parsing helpers for SQLite datetime strings

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/store/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: implement SQLite store for boards"
```

---

## Task 6: SQLite Store Implementation — Lists

**Files:**
- Modify: `internal/store/sqlite.go`
- Modify: `internal/store/sqlite_test.go`

**Step 1: Write list store tests**

Add tests to `sqlite_test.go`:
- `TestCreateList` — create list in a board, verify retrieval
- `TestListListsByBoard` — create multiple lists, verify order by position
- `TestUpdateList` — rename list, verify
- `TestDeleteList` — delete list, verify gone
- `TestDeleteList_CascadesCards` — create list with cards, delete list, verify cards gone

**Step 2: Run to verify fail, implement, verify pass**

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: implement SQLite store for lists"
```

---

## Task 7: SQLite Store Implementation — Cards

**Files:**
- Modify: `internal/store/sqlite.go`
- Modify: `internal/store/sqlite_test.go`

**Step 1: Write card store tests**

- `TestCreateCard` — create card in list, verify all fields
- `TestListCardsByList` — multiple cards, ordered by position
- `TestUpdateCard` — update fields, verify
- `TestMoveCard` — move card between lists, verify list_id and position changed
- `TestDeleteCard`
- `TestSearchCards` — create cards with different assignees/statuses/labels, verify search filters work

**Step 2: Run to verify fail, implement, verify pass**

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: implement SQLite store for cards with search"
```

---

## Task 8: SQLite Store Implementation — Dependencies, Labels, Activity

**Files:**
- Modify: `internal/store/sqlite.go`
- Modify: `internal/store/sqlite_test.go`

**Step 1: Write dependency tests**

- `TestAddDependency` — add dep, verify GetDependencies returns it
- `TestRemoveDependency`
- `TestGetDependents` — verify reverse lookup
- `TestAddDependency_NoDuplicates` — verify unique constraint

**Step 2: Write label tests**

- `TestCreateLabel`
- `TestAddLabelToCard` / `TestRemoveLabelFromCard`
- `TestGetLabelsForCard`

**Step 3: Write activity tests**

- `TestCreateActivity`
- `TestListActivityByCard` — verify ordering (newest first) and limit
- `TestListActivityByBoard` — verify it finds activity across all cards in the board

**Step 4: Implement all three, run tests**

Run: `go test ./internal/store/ -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: implement SQLite store for dependencies, labels, and activity log"
```

---

## Task 9: Event Bus

**Files:**
- Create: `internal/event/bus.go`
- Create: `internal/event/bus_test.go`

**Step 1: Write event bus tests**

```go
// Test that a subscriber receives published events
func TestBus_SubscribeAndPublish(t *testing.T)

// Test that events are scoped to board ID
func TestBus_BoardScoping(t *testing.T)

// Test that unsubscribe stops delivery
func TestBus_Unsubscribe(t *testing.T)

// Test multiple subscribers on same board
func TestBus_MultipleSubscribers(t *testing.T)
```

**Step 2: Run to verify fail**

**Step 3: Implement event bus**

Create `internal/event/bus.go`:
- `Event` struct: `Type string`, `BoardID string`, `Payload any`, `SeqID uint64`
- `Bus` struct with `sync.RWMutex`, map of `boardID → []*Subscriber`
- `Subscriber` struct wrapping a `chan Event` with buffer size 64
- `NewBus() *Bus`
- `Subscribe(boardID string) *Subscriber`
- `Unsubscribe(sub *Subscriber)`
- `Publish(event Event)` — fan-out to all subscribers for that board, non-blocking send (drop if channel full)
- Atomic counter for `SeqID`

**Step 4: Run tests, verify pass**

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: implement in-process event bus with board-scoped pub/sub"
```

---

## Task 10: Service Layer

**Files:**
- Create: `internal/service/service.go`
- Create: `internal/service/service_test.go`

**Step 1: Write service tests**

The service layer wraps the store and adds:
- Validation (e.g., card title not empty, valid status/priority enums)
- Activity log entries on mutations
- Event bus publishing on mutations

Test key behaviors:
- `TestCreateBoard_Valid` — happy path
- `TestCreateBoard_EmptyName` — returns error
- `TestCreateCard_LogsActivity` — verify activity log entry created
- `TestMoveCard_EmitsEvent` — verify event bus receives card.moved
- `TestAssignCard_UpdatesStatus` — assigning sets status to "assigned" if "unassigned"
- `TestAddDependency_SelfReference` — error when card depends on itself

**Step 2: Run to verify fail**

**Step 3: Implement service**

Create `internal/service/service.go`:
```go
type Service struct {
    store store.Store
    bus   *event.Bus
}

func New(store store.Store, bus *event.Bus) *Service
```

Methods mirror the store but add validation, activity logging, and event emission. Each write method:
1. Validates input
2. Calls store
3. Creates activity log entry via store
4. Publishes event to bus
5. Returns result

**Step 4: Run tests, verify pass**

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: implement service layer with validation, activity logging, and events"
```

---

## Task 11: Fiber v3 HTTP Server Skeleton

**Files:**
- Modify: `cmd/cielo/main.go`
- Create: `internal/api/router.go`
- Create: `internal/api/middleware.go`

**Step 1: Install Fiber v3**

Run:
```bash
go get github.com/gofiber/fiber/v3
```

**Step 2: Write router**

Create `internal/api/router.go`:
```go
package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/aellingwood/cielo/internal/service"
	"github.com/aellingwood/cielo/internal/event"
)

func SetupRouter(app *fiber.App, svc *service.Service, bus *event.Bus) {
	api := app.Group("/api/v1")

	// Board routes
	api.Get("/boards", listBoards(svc))
	api.Post("/boards", createBoard(svc))
	api.Get("/boards/:id", getBoard(svc))
	api.Put("/boards/:id", updateBoard(svc))
	api.Delete("/boards/:id", deleteBoard(svc))

	// List routes
	api.Post("/boards/:boardId/lists", createList(svc))
	api.Put("/lists/:id", updateList(svc))
	api.Delete("/lists/:id", deleteList(svc))

	// Card routes
	api.Post("/lists/:listId/cards", createCard(svc))
	api.Get("/cards/:id", getCard(svc))
	api.Put("/cards/:id", updateCard(svc))
	api.Delete("/cards/:id", deleteCard(svc))
	api.Put("/cards/:id/move", moveCard(svc))
	api.Put("/cards/:id/assign", assignCard(svc))

	// Dependency routes
	api.Post("/cards/:id/dependencies", addDependency(svc))
	api.Delete("/cards/:id/dependencies/:depId", removeDependency(svc))

	// Label routes
	api.Get("/boards/:boardId/labels", listLabels(svc))
	api.Post("/boards/:boardId/labels", createLabel(svc))
	api.Put("/labels/:id", updateLabel(svc))
	api.Delete("/labels/:id", deleteLabel(svc))
	api.Post("/cards/:id/labels", addLabelToCard(svc))
	api.Delete("/cards/:id/labels/:labelId", removeLabelFromCard(svc))

	// Activity & search
	api.Get("/cards/:id/activity", getCardActivity(svc))
	api.Get("/boards/:boardId/activity", getBoardActivity(svc))
	api.Get("/boards/:boardId/search", searchCards(svc))

	// SSE
	api.Get("/boards/:boardId/events", boardSSE(bus))
}
```

**Step 3: Write middleware**

Create `internal/api/middleware.go` with:
- CORS middleware (allow all origins for dev)
- Request logging middleware
- Error handling middleware that returns JSON `{"error": "message"}`

**Step 4: Wire up main.go**

Update `cmd/cielo/main.go` to:
1. Load config
2. Open SQLite DB, run migrations
3. Create store, event bus, service
4. Create Fiber app with middleware
5. Call `SetupRouter`
6. Start listening

**Step 5: Verify it compiles and starts**

Run: `make run`
Expected: Server starts on :8080, no errors. `curl localhost:8080/api/v1/boards` returns `[]` (empty array) or a stub.

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: wire up Fiber v3 HTTP server with route skeleton and middleware"
```

---

## Task 12: Board API Handlers

**Files:**
- Create: `internal/api/board.go`
- Create: `internal/api/board_test.go`

**Step 1: Write handler tests**

Use `net/http/httptest` and Fiber's `app.Test()` to test:
- `TestCreateBoard_API` — POST JSON, verify 201 + returned board
- `TestListBoards_API` — create two, GET, verify array
- `TestGetBoard_API` — create, GET by ID, verify
- `TestUpdateBoard_API` — create, PUT, verify updated
- `TestDeleteBoard_API` — create, DELETE, verify 204
- `TestGetBoard_NotFound` — GET nonexistent ID, verify 404

**Step 2: Run to verify fail**

**Step 3: Implement board handlers**

Each handler:
- Parses path params / request body
- Calls service method
- Returns JSON response with appropriate status code

**Step 4: Run tests, verify pass**

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: implement board API handlers with tests"
```

---

## Task 13: List API Handlers

**Files:**
- Create: `internal/api/list.go`
- Create: `internal/api/list_test.go`

Same pattern as Task 12 but for list endpoints.

Tests:
- `TestCreateList_API`
- `TestUpdateList_API`
- `TestDeleteList_API`

**Commit:**
```bash
git add -A
git commit -m "feat: implement list API handlers with tests"
```

---

## Task 14: Card API Handlers

**Files:**
- Create: `internal/api/card.go`
- Create: `internal/api/card_test.go`

Tests:
- `TestCreateCard_API`
- `TestGetCard_API` — verify labels, deps, activity are included
- `TestUpdateCard_API`
- `TestDeleteCard_API`
- `TestMoveCard_API`
- `TestAssignCard_API`
- `TestSearchCards_API`

**Commit:**
```bash
git add -A
git commit -m "feat: implement card API handlers with move, assign, and search"
```

---

## Task 15: Dependency, Label, and Activity API Handlers

**Files:**
- Modify: `internal/api/card.go` (or create `internal/api/dependency.go`, `internal/api/label.go`, `internal/api/activity.go`)
- Tests for each

Tests:
- `TestAddDependency_API` / `TestRemoveDependency_API`
- `TestCreateLabel_API` / `TestAddLabelToCard_API` / `TestRemoveLabelFromCard_API`
- `TestGetCardActivity_API` / `TestGetBoardActivity_API`

**Commit:**
```bash
git add -A
git commit -m "feat: implement dependency, label, and activity API handlers"
```

---

## Task 16: SSE Endpoint

**Files:**
- Create: `internal/api/sse.go`
- Create: `internal/api/sse_test.go`

**Step 1: Write SSE test**

Test that:
- Connecting to `/api/v1/boards/:boardId/events` returns `text/event-stream` content type
- Publishing an event via the bus delivers it to the connected SSE client
- Events include `id:` field with sequence number

**Step 2: Implement SSE handler**

Use Fiber v3's `SendStreamWriter` to stream events:
```go
func boardSSE(bus *event.Bus) fiber.Handler {
    return func(c fiber.Ctx) error {
        boardID := c.Params("boardId")
        c.Set("Content-Type", "text/event-stream")
        c.Set("Cache-Control", "no-cache")
        c.Set("Connection", "keep-alive")

        sub := bus.Subscribe(boardID)
        defer bus.Unsubscribe(sub)

        return c.SendStreamWriter(func(w *bufio.Writer) {
            for evt := range sub.Ch {
                data, _ := json.Marshal(evt.Payload)
                fmt.Fprintf(w, "id: %d\n", evt.SeqID)
                fmt.Fprintf(w, "event: %s\n", evt.Type)
                fmt.Fprintf(w, "data: %s\n\n", data)
                w.Flush()
            }
        })
    }
}
```

**Step 3: Run tests, verify pass**

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: implement SSE endpoint for real-time board events"
```

---

## Task 17: MCP Server — JSON-RPC Dispatcher

**Files:**
- Create: `internal/mcp/server.go`
- Create: `internal/mcp/server_test.go`

**Step 1: Write dispatcher tests**

- `TestInitialize` — send `initialize` request, verify capabilities response
- `TestToolsList` — send `tools/list`, verify all tools returned with schemas
- `TestToolsCall_UnknownTool` — verify error response
- `TestInvalidJSON` — verify parse error

**Step 2: Implement JSON-RPC dispatcher**

Create `internal/mcp/server.go`:
```go
type Server struct {
    svc *service.Service
}

type JSONRPCRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      any             `json:"id"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
    JSONRPC string `json:"jsonrpc"`
    ID      any    `json:"id"`
    Result  any    `json:"result,omitempty"`
    Error   *RPCError `json:"error,omitempty"`
}

func (s *Server) HandleRequest(ctx context.Context, req JSONRPCRequest) JSONRPCResponse
```

Dispatch on `req.Method`:
- `"initialize"` → return server info + capabilities (tools)
- `"tools/list"` → return tool definitions
- `"tools/call"` → route to tool handler by `params.name`

**Step 3: Run tests, verify pass**

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: implement MCP JSON-RPC dispatcher with initialize and tools/list"
```

---

## Task 18: MCP Tool Definitions & Handlers

**Files:**
- Create: `internal/mcp/tools.go`
- Create: `internal/mcp/tools_test.go`

**Step 1: Write tool handler tests**

Test each tool via the JSON-RPC dispatcher:
- `TestTool_ListBoards` — create boards via service, call tool, verify response
- `TestTool_CreateBoard` — call tool, verify board created
- `TestTool_CreateCard` — call tool, verify card created with activity log
- `TestTool_MoveCard` — call tool, verify card moved
- `TestTool_AssignCard` — call tool, verify assignee + status
- `TestTool_AddComment` — call tool, verify activity entry
- `TestTool_SearchCards` — call tool, verify filtered results

**Step 2: Implement tool definitions**

Each tool has:
- `name` and `description`
- `inputSchema` (JSON Schema object)
- Handler function that parses params, calls service, returns result

Define all 20 tools (7 read + 13 write) per the design doc.

**Step 3: Run tests, verify pass**

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: implement all 20 MCP tool handlers"
```

---

## Task 19: MCP Streamable HTTP Transport

**Files:**
- Create: `internal/mcp/transport.go`
- Modify: `internal/api/router.go` (add `/mcp` route)

**Step 1: Write transport test**

Test that `POST /mcp` with a JSON-RPC request body returns the correct JSON-RPC response.

**Step 2: Implement transport**

Create a Fiber handler that:
1. Reads the JSON-RPC request from the body
2. Passes it to `Server.HandleRequest`
3. Returns the JSON-RPC response

Wire it into the Fiber router:
```go
app.Post("/mcp", mcpHandler(mcpServer))
```

**Step 3: Run tests, verify pass**

**Step 4: Manual smoke test**

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

Expected: JSON response with all tool definitions.

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: implement MCP streamable HTTP transport on POST /mcp"
```

---

## Task 20: Frontend Scaffolding

**Files:**
- Create: `web/` (Vite project)

**Step 1: Scaffold React + Vite + TypeScript**

Run:
```bash
cd /Users/aellingwood/dev/personal/cielo
npm create vite@latest web -- --template react-ts
cd web
npm install
```

**Step 2: Install dependencies**

```bash
cd /Users/aellingwood/dev/personal/cielo/web
npm install @tanstack/react-query react-router-dom @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
npm install -D tailwindcss @tailwindcss/vite
```

**Step 3: Configure Tailwind CSS v4**

Add Tailwind to `vite.config.ts`:
```typescript
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
})
```

Replace `src/index.css` contents with:
```css
@import "tailwindcss";
```

**Step 4: Configure Vite proxy**

Update `vite.config.ts` to proxy `/api` and `/mcp` to `localhost:8080`:
```typescript
server: {
  proxy: {
    '/api': 'http://localhost:8080',
    '/mcp': 'http://localhost:8080',
  }
}
```

**Step 5: Set up React Router and TanStack Query**

Update `src/main.tsx` with `BrowserRouter` and `QueryClientProvider`.

**Step 6: Clean up default Vite boilerplate**

Remove default `App.tsx` content, `App.css`, `assets/`, etc.

**Step 7: Verify dev server starts**

Run: `cd web && npm run dev`
Expected: Vite dev server starts, blank page loads at localhost:5173

**Step 8: Commit**

```bash
git add -A
git commit -m "feat: scaffold React + Vite + Tailwind CSS + TanStack Query frontend"
```

---

## Task 21: Frontend API Client & Types

**Files:**
- Create: `web/src/api/client.ts`
- Create: `web/src/api/types.ts`

**Step 1: Define TypeScript types matching Go models**

Create `web/src/api/types.ts` with interfaces for Board, List, Card, Label, ActivityLog, CardDependency.

**Step 2: Create API client**

Create `web/src/api/client.ts` with fetch wrappers for every endpoint:
```typescript
const BASE = '/api/v1'

export const api = {
  boards: {
    list: () => fetch(`${BASE}/boards`).then(r => r.json()),
    create: (data: {name: string; description?: string}) => ...,
    get: (id: string) => ...,
    update: (id: string, data: Partial<Board>) => ...,
    delete: (id: string) => ...,
  },
  lists: { ... },
  cards: { ... },
  labels: { ... },
  activity: { ... },
}
```

**Step 3: Create TanStack Query hooks**

Create `web/src/api/hooks.ts` with custom hooks:
```typescript
export function useBoards() { return useQuery({queryKey: ['boards'], queryFn: api.boards.list}) }
export function useBoard(id: string) { ... }
export function useCreateBoard() { return useMutation({...}) }
// etc.
```

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: add TypeScript API client, types, and TanStack Query hooks"
```

---

## Task 22: Frontend — Layout & Board List Page

**Files:**
- Create: `web/src/components/Layout.tsx`
- Create: `web/src/pages/BoardList.tsx`
- Modify: `web/src/App.tsx`

**Step 1: Create Layout component**

Navigation bar with app name "Cielo" and link to home.

**Step 2: Create BoardList page**

Grid of board cards. Each card shows name, description, card count. "Create Board" button with inline form.

**Step 3: Wire up React Router in App.tsx**

```typescript
<Routes>
  <Route path="/" element={<BoardList />} />
  <Route path="/boards/:boardId" element={<BoardView />} />
</Routes>
```

**Step 4: Verify**

Run frontend + backend, verify board list page loads and can create a board.

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add layout component and board list page"
```

---

## Task 23: Frontend — Board View (Kanban)

**Files:**
- Create: `web/src/pages/BoardView.tsx`
- Create: `web/src/components/ListColumn.tsx`
- Create: `web/src/components/ListHeader.tsx`
- Create: `web/src/components/CardTile.tsx`
- Create: `web/src/components/AddListButton.tsx`

**Step 1: Create BoardView page**

Fetches board data (lists + cards) via TanStack Query. Renders horizontal scrollable container of `ListColumn` components.

**Step 2: Create ListColumn**

Droppable zone (dnd-kit `useDroppable`). Renders `ListHeader` + vertical list of `CardTile` components.

**Step 3: Create CardTile**

Draggable card (dnd-kit `useDraggable`). Shows title, label chips, assignee badge, priority indicator.

**Step 4: Create AddListButton**

Button at end of columns. Click reveals inline text input to create a new list.

**Step 5: Integrate dnd-kit**

Wrap board in `DndContext`. Handle `onDragEnd` to call `moveCard` API. Implement optimistic update — move card in local state immediately, revert on API failure.

**Step 6: Verify**

Create a board, add lists, add cards, drag cards between lists.

**Step 7: Commit**

```bash
git add -A
git commit -m "feat: implement Kanban board view with drag-and-drop"
```

---

## Task 24: Frontend — Card Detail Modal

**Files:**
- Create: `web/src/components/CardDetail.tsx`
- Create: `web/src/components/ActivityFeed.tsx`
- Create: `web/src/components/DependencyList.tsx`
- Create: `web/src/components/LabelManager.tsx`

**Step 1: Create CardDetail modal**

Opened when clicking a card tile. URL changes to `/boards/:boardId/cards/:cardId` (modal overlay, board still visible behind). Contains:
- Title (editable inline)
- Description (markdown editor — use a simple textarea with preview for v1)
- Assignee, status, priority, due date selects
- LabelManager
- DependencyList
- ActivityFeed

**Step 2: Create ActivityFeed**

Chronological list of activity entries. Shows actor, action, timestamp, detail.

**Step 3: Create DependencyList**

Shows "Blocked by" and "Blocking" sections. Add dependency via card search/select.

**Step 4: Create LabelManager**

Shows current labels as chips. Dropdown to add/remove labels.

**Step 5: Verify**

Click a card, edit fields, add labels, add dependencies, view activity.

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: implement card detail modal with activity, deps, and labels"
```

---

## Task 25: Frontend — SSE Integration

**Files:**
- Create: `web/src/components/SSEProvider.tsx`
- Modify: `web/src/pages/BoardView.tsx`

**Step 1: Create SSEProvider**

React context provider that:
- Connects to `/api/v1/boards/:boardId/events` via `EventSource`
- Parses incoming events and invalidates relevant TanStack Query cache keys
- Handles reconnection with `Last-Event-ID`
- Cleans up on unmount

**Step 2: Wrap BoardView in SSEProvider**

**Step 3: Verify**

Open two browser tabs on the same board. Create a card in one tab, verify it appears in the other tab within a second.

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: add SSE provider for real-time board updates"
```

---

## Task 26: End-to-End Smoke Test

**Step 1: Start the server**

```bash
make run
```

**Step 2: Test MCP flow**

```bash
# Initialize
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'

# Create a board via MCP
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"create_board","arguments":{"name":"Agent Project","description":"Test board"}}}'

# List boards via MCP
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/list"}'
```

**Step 3: Test HTTP API flow**

```bash
# List boards
curl -s http://localhost:8080/api/v1/boards | jq .

# Verify board created via MCP appears in API
```

**Step 4: Test frontend**

Open `http://localhost:5173`, verify board created via MCP is visible.

**Step 5: Test SSE**

Open board in browser. Create a card via curl. Verify card appears in browser without refresh.

**Step 6: Commit any fixes**

```bash
git add -A
git commit -m "fix: address issues found during end-to-end smoke testing"
```

---

## Task 27: Update Transcript

**Files:**
- Modify: `docs/transcript.md`

**Step 1: Add Phase 2 section to transcript**

Document the implementation journey: what went smoothly, what required iteration, key decisions made during build.

**Step 2: Commit**

```bash
git add docs/transcript.md
git commit -m "docs: update build transcript with implementation notes"
```

---

## Summary

| Task | Description | Dependencies |
|------|-------------|-------------|
| 1 | Go project scaffolding | — |
| 2 | SQLite schema & migration | 1 |
| 3 | Models & UUIDv7 | 1 |
| 4 | Store interface | 3 |
| 5 | SQLite store — boards | 2, 4 |
| 6 | SQLite store — lists | 5 |
| 7 | SQLite store — cards | 6 |
| 8 | SQLite store — deps, labels, activity | 7 |
| 9 | Event bus | 3 |
| 10 | Service layer | 8, 9 |
| 11 | Fiber HTTP server skeleton | 10 |
| 12 | Board API handlers | 11 |
| 13 | List API handlers | 11 |
| 14 | Card API handlers | 11 |
| 15 | Dep, label, activity handlers | 11 |
| 16 | SSE endpoint | 11 |
| 17 | MCP JSON-RPC dispatcher | 10 |
| 18 | MCP tool handlers | 17 |
| 19 | MCP HTTP transport | 18, 11 |
| 20 | Frontend scaffolding | — |
| 21 | Frontend API client & types | 20 |
| 22 | Frontend board list page | 21 |
| 23 | Frontend Kanban board view | 22 |
| 24 | Frontend card detail modal | 23 |
| 25 | Frontend SSE integration | 24, 16 |
| 26 | End-to-end smoke test | all |
| 27 | Update transcript | 26 |
