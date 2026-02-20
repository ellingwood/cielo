package store_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/aellingwood/cielo/internal/model"
	"github.com/aellingwood/cielo/internal/store"
)

func setupTestDB(t *testing.T) (*store.SQLiteStore, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("PRAGMA foreign_keys = ON")
	if err := store.RunMigrations(db); err != nil {
		t.Fatal(err)
	}
	return store.NewSQLiteStore(db), db
}

func TestRunMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := store.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	tables := []string{"boards", "lists", "cards", "card_dependencies", "labels", "card_labels", "activity_log"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestBoardCRUD(t *testing.T) {
	s, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	b := &model.Board{ID: model.NewID(), Name: "Test Board", Description: "desc"}
	if err := s.CreateBoard(ctx, b); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetBoard(ctx, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Test Board" {
		t.Errorf("expected name 'Test Board', got %q", got.Name)
	}

	boards, err := s.ListBoards(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(boards) != 1 {
		t.Errorf("expected 1 board, got %d", len(boards))
	}

	b.Name = "Updated"
	if err := s.UpdateBoard(ctx, b); err != nil {
		t.Fatal(err)
	}
	got, _ = s.GetBoard(ctx, b.ID)
	if got.Name != "Updated" {
		t.Errorf("expected updated name")
	}

	if err := s.DeleteBoard(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	_, err = s.GetBoard(ctx, b.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestListCRUD(t *testing.T) {
	s, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	b := &model.Board{ID: model.NewID(), Name: "Board"}
	s.CreateBoard(ctx, b)

	l := &model.List{ID: model.NewID(), BoardID: b.ID, Name: "Todo", Position: 0}
	if err := s.CreateList(ctx, l); err != nil {
		t.Fatal(err)
	}

	lists, err := s.ListListsByBoard(ctx, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 1 || lists[0].Name != "Todo" {
		t.Errorf("unexpected lists: %+v", lists)
	}
}

func TestCardCRUD(t *testing.T) {
	s, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	b := &model.Board{ID: model.NewID(), Name: "Board"}
	s.CreateBoard(ctx, b)
	l := &model.List{ID: model.NewID(), BoardID: b.ID, Name: "Todo", Position: 0}
	s.CreateList(ctx, l)

	c := &model.Card{ID: model.NewID(), ListID: l.ID, Title: "Task 1", Position: 0, Status: "unassigned", Priority: "medium"}
	if err := s.CreateCard(ctx, c); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetCard(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Task 1" {
		t.Errorf("expected 'Task 1', got %q", got.Title)
	}

	cards, err := s.ListCardsByList(ctx, l.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestMoveCard(t *testing.T) {
	s, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	b := &model.Board{ID: model.NewID(), Name: "Board"}
	s.CreateBoard(ctx, b)
	l1 := &model.List{ID: model.NewID(), BoardID: b.ID, Name: "Todo", Position: 0}
	l2 := &model.List{ID: model.NewID(), BoardID: b.ID, Name: "Done", Position: 1}
	s.CreateList(ctx, l1)
	s.CreateList(ctx, l2)

	c := &model.Card{ID: model.NewID(), ListID: l1.ID, Title: "Task", Position: 0, Status: "unassigned", Priority: "medium"}
	s.CreateCard(ctx, c)

	if err := s.MoveCard(ctx, c.ID, l2.ID, 0); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetCard(ctx, c.ID)
	if got.ListID != l2.ID {
		t.Errorf("expected card in list %s, got %s", l2.ID, got.ListID)
	}
}

func TestDependencies(t *testing.T) {
	s, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	b := &model.Board{ID: model.NewID(), Name: "Board"}
	s.CreateBoard(ctx, b)
	l := &model.List{ID: model.NewID(), BoardID: b.ID, Name: "Todo", Position: 0}
	s.CreateList(ctx, l)

	c1 := &model.Card{ID: model.NewID(), ListID: l.ID, Title: "Blocker", Position: 0, Status: "unassigned", Priority: "medium"}
	c2 := &model.Card{ID: model.NewID(), ListID: l.ID, Title: "Blocked", Position: 1, Status: "unassigned", Priority: "medium"}
	s.CreateCard(ctx, c1)
	s.CreateCard(ctx, c2)

	dep := &model.CardDependency{ID: model.NewID(), CardID: c2.ID, DependsOnCardID: c1.ID}
	if err := s.AddDependency(ctx, dep); err != nil {
		t.Fatal(err)
	}

	deps, err := s.GetDependencies(ctx, c2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(deps) != 1 || deps[0].ID != c1.ID {
		t.Errorf("expected c1 as dependency")
	}

	dependents, err := s.GetDependents(ctx, c1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(dependents) != 1 || dependents[0].ID != c2.ID {
		t.Errorf("expected c2 as dependent")
	}

	if err := s.RemoveDependency(ctx, c2.ID, c1.ID); err != nil {
		t.Fatal(err)
	}
	deps, _ = s.GetDependencies(ctx, c2.ID)
	if len(deps) != 0 {
		t.Errorf("expected no dependencies after removal")
	}
}

func TestLabels(t *testing.T) {
	s, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	b := &model.Board{ID: model.NewID(), Name: "Board"}
	s.CreateBoard(ctx, b)
	l := &model.List{ID: model.NewID(), BoardID: b.ID, Name: "Todo", Position: 0}
	s.CreateList(ctx, l)
	c := &model.Card{ID: model.NewID(), ListID: l.ID, Title: "Task", Position: 0, Status: "unassigned", Priority: "medium"}
	s.CreateCard(ctx, c)

	label := &model.Label{ID: model.NewID(), BoardID: b.ID, Name: "coding", Color: "#ff0000"}
	if err := s.CreateLabel(ctx, label); err != nil {
		t.Fatal(err)
	}

	if err := s.AddLabelToCard(ctx, c.ID, label.ID); err != nil {
		t.Fatal(err)
	}

	labels, err := s.GetLabelsForCard(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 1 || labels[0].Name != "coding" {
		t.Errorf("expected coding label")
	}

	if err := s.RemoveLabelFromCard(ctx, c.ID, label.ID); err != nil {
		t.Fatal(err)
	}
	labels, _ = s.GetLabelsForCard(ctx, c.ID)
	if len(labels) != 0 {
		t.Errorf("expected no labels after removal")
	}
}

func TestActivity(t *testing.T) {
	s, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	b := &model.Board{ID: model.NewID(), Name: "Board"}
	s.CreateBoard(ctx, b)
	l := &model.List{ID: model.NewID(), BoardID: b.ID, Name: "Todo", Position: 0}
	s.CreateList(ctx, l)
	c := &model.Card{ID: model.NewID(), ListID: l.ID, Title: "Task", Position: 0, Status: "unassigned", Priority: "medium"}
	s.CreateCard(ctx, c)

	entry := &model.ActivityLog{ID: model.NewID(), CardID: c.ID, Actor: "agent-1", Action: "created", Detail: `{"title":"Task"}`}
	if err := s.CreateActivity(ctx, entry); err != nil {
		t.Fatal(err)
	}

	entries, err := s.ListActivityByCard(ctx, c.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Actor != "agent-1" {
		t.Errorf("unexpected activity: %+v", entries)
	}

	entries, err = s.ListActivityByBoard(ctx, b.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 board activity, got %d", len(entries))
	}
}

func TestSearchCards(t *testing.T) {
	s, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	b := &model.Board{ID: model.NewID(), Name: "Board"}
	s.CreateBoard(ctx, b)
	l := &model.List{ID: model.NewID(), BoardID: b.ID, Name: "Todo", Position: 0}
	s.CreateList(ctx, l)

	c1 := &model.Card{ID: model.NewID(), ListID: l.ID, Title: "Fix bug", Assignee: "alice", Status: "in_progress", Priority: "high", Position: 0}
	c2 := &model.Card{ID: model.NewID(), ListID: l.ID, Title: "Add feature", Assignee: "bob", Status: "unassigned", Priority: "medium", Position: 1}
	s.CreateCard(ctx, c1)
	s.CreateCard(ctx, c2)

	cards, err := s.SearchCards(ctx, b.ID, "bug", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 || cards[0].Title != "Fix bug" {
		t.Errorf("search by text failed")
	}

	cards, _ = s.SearchCards(ctx, b.ID, "", "alice", "", "")
	if len(cards) != 1 {
		t.Errorf("search by assignee failed")
	}

	cards, _ = s.SearchCards(ctx, b.ID, "", "", "in_progress", "")
	if len(cards) != 1 {
		t.Errorf("search by status failed")
	}
}
