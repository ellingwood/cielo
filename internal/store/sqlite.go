package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aellingwood/cielo/internal/model"
)

const timeLayout = "2006-01-02T15:04:05.000Z"

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(timeLayout, s)
	return t
}

func now() string {
	return time.Now().UTC().Format(timeLayout)
}

// --- Boards ---

func (s *SQLiteStore) CreateBoard(ctx context.Context, board *model.Board) error {
	ts := now()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO boards (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		board.ID, board.Name, board.Description, ts, ts)
	if err != nil {
		return err
	}
	board.CreatedAt = parseTime(ts)
	board.UpdatedAt = parseTime(ts)
	return nil
}

func (s *SQLiteStore) GetBoard(ctx context.Context, id string) (*model.Board, error) {
	var b model.Board
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at, updated_at FROM boards WHERE id = ?", id).
		Scan(&b.ID, &b.Name, &b.Description, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("board not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	b.CreatedAt = parseTime(createdAt)
	b.UpdatedAt = parseTime(updatedAt)
	return &b, nil
}

func (s *SQLiteStore) ListBoards(ctx context.Context) ([]model.Board, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, description, created_at, updated_at FROM boards ORDER BY created_at ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var boards []model.Board
	for rows.Next() {
		var b model.Board
		var createdAt, updatedAt string
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		b.CreatedAt = parseTime(createdAt)
		b.UpdatedAt = parseTime(updatedAt)
		boards = append(boards, b)
	}
	return boards, rows.Err()
}

func (s *SQLiteStore) UpdateBoard(ctx context.Context, board *model.Board) error {
	ts := now()
	res, err := s.db.ExecContext(ctx,
		"UPDATE boards SET name = ?, description = ?, updated_at = ? WHERE id = ?",
		board.Name, board.Description, ts, board.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("board not found: %s", board.ID)
	}
	board.UpdatedAt = parseTime(ts)
	return nil
}

func (s *SQLiteStore) DeleteBoard(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM boards WHERE id = ?", id)
	return err
}

// --- Lists ---

func (s *SQLiteStore) CreateList(ctx context.Context, list *model.List) error {
	ts := now()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO lists (id, board_id, name, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		list.ID, list.BoardID, list.Name, list.Position, ts, ts)
	if err != nil {
		return err
	}
	list.CreatedAt = parseTime(ts)
	list.UpdatedAt = parseTime(ts)
	return nil
}

func (s *SQLiteStore) GetList(ctx context.Context, id string) (*model.List, error) {
	var l model.List
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, board_id, name, position, created_at, updated_at FROM lists WHERE id = ?", id).
		Scan(&l.ID, &l.BoardID, &l.Name, &l.Position, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("list not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	l.CreatedAt = parseTime(createdAt)
	l.UpdatedAt = parseTime(updatedAt)
	return &l, nil
}

func (s *SQLiteStore) ListListsByBoard(ctx context.Context, boardID string) ([]model.List, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, board_id, name, position, created_at, updated_at FROM lists WHERE board_id = ? ORDER BY position ASC",
		boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lists []model.List
	for rows.Next() {
		var l model.List
		var createdAt, updatedAt string
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Position, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		l.CreatedAt = parseTime(createdAt)
		l.UpdatedAt = parseTime(updatedAt)
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func (s *SQLiteStore) UpdateList(ctx context.Context, list *model.List) error {
	ts := now()
	res, err := s.db.ExecContext(ctx,
		"UPDATE lists SET name = ?, position = ?, updated_at = ? WHERE id = ?",
		list.Name, list.Position, ts, list.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("list not found: %s", list.ID)
	}
	list.UpdatedAt = parseTime(ts)
	return nil
}

func (s *SQLiteStore) DeleteList(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM lists WHERE id = ?", id)
	return err
}

// --- Cards ---

func (s *SQLiteStore) CreateCard(ctx context.Context, card *model.Card) error {
	ts := now()
	var dueDate *string
	if card.DueDate != nil {
		d := card.DueDate.Format(timeLayout)
		dueDate = &d
	}
	if card.Status == "" {
		card.Status = model.StatusUnassigned
	}
	if card.Priority == "" {
		card.Priority = model.PriorityMedium
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO cards (id, list_id, title, description, position, assignee, status, priority, due_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		card.ID, card.ListID, card.Title, card.Description, card.Position,
		card.Assignee, card.Status, card.Priority, dueDate, ts, ts)
	if err != nil {
		return err
	}
	card.CreatedAt = parseTime(ts)
	card.UpdatedAt = parseTime(ts)
	return nil
}

func (s *SQLiteStore) scanCard(row interface{ Scan(...any) error }) (*model.Card, error) {
	var c model.Card
	var createdAt, updatedAt string
	var dueDate sql.NullString
	err := row.Scan(&c.ID, &c.ListID, &c.Title, &c.Description, &c.Position,
		&c.Assignee, &c.Status, &c.Priority, &dueDate, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	c.CreatedAt = parseTime(createdAt)
	c.UpdatedAt = parseTime(updatedAt)
	if dueDate.Valid {
		t := parseTime(dueDate.String)
		c.DueDate = &t
	}
	return &c, nil
}

func (s *SQLiteStore) GetCard(ctx context.Context, id string) (*model.Card, error) {
	c, err := s.scanCard(s.db.QueryRowContext(ctx,
		`SELECT id, list_id, title, description, position, assignee, status, priority, due_date, created_at, updated_at
		 FROM cards WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("card not found: %s", id)
	}
	return c, err
}

func (s *SQLiteStore) ListCardsByList(ctx context.Context, listID string) ([]model.Card, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, list_id, title, description, position, assignee, status, priority, due_date, created_at, updated_at
		 FROM cards WHERE list_id = ? ORDER BY position ASC`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []model.Card
	for rows.Next() {
		c, err := s.scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, *c)
	}
	return cards, rows.Err()
}

func (s *SQLiteStore) UpdateCard(ctx context.Context, card *model.Card) error {
	ts := now()
	var dueDate *string
	if card.DueDate != nil {
		d := card.DueDate.Format(timeLayout)
		dueDate = &d
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE cards SET list_id=?, title=?, description=?, position=?, assignee=?, status=?, priority=?, due_date=?, updated_at=?
		 WHERE id=?`,
		card.ListID, card.Title, card.Description, card.Position,
		card.Assignee, card.Status, card.Priority, dueDate, ts, card.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("card not found: %s", card.ID)
	}
	card.UpdatedAt = parseTime(ts)
	return nil
}

func (s *SQLiteStore) MoveCard(ctx context.Context, cardID, targetListID string, position int) error {
	ts := now()
	res, err := s.db.ExecContext(ctx,
		"UPDATE cards SET list_id = ?, position = ?, updated_at = ? WHERE id = ?",
		targetListID, position, ts, cardID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("card not found: %s", cardID)
	}
	return nil
}

func (s *SQLiteStore) DeleteCard(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM cards WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) SearchCards(ctx context.Context, boardID, query, assignee, status, label string) ([]model.Card, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, "l.board_id = ?")
	args = append(args, boardID)

	if query != "" {
		conditions = append(conditions, "(c.title LIKE ? OR c.description LIKE ?)")
		q := "%" + query + "%"
		args = append(args, q, q)
	}
	if assignee != "" {
		conditions = append(conditions, "c.assignee = ?")
		args = append(args, assignee)
	}
	if status != "" {
		conditions = append(conditions, "c.status = ?")
		args = append(args, status)
	}

	baseQuery := `SELECT DISTINCT c.id, c.list_id, c.title, c.description, c.position, c.assignee, c.status, c.priority, c.due_date, c.created_at, c.updated_at
		FROM cards c JOIN lists l ON c.list_id = l.id`

	if label != "" {
		baseQuery += " JOIN card_labels cl ON c.id = cl.card_id JOIN labels lb ON cl.label_id = lb.id"
		conditions = append(conditions, "lb.name = ?")
		args = append(args, label)
	}

	baseQuery += " WHERE " + strings.Join(conditions, " AND ") + " ORDER BY c.position ASC"

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []model.Card
	for rows.Next() {
		c, err := s.scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, *c)
	}
	return cards, rows.Err()
}

// --- Dependencies ---

func (s *SQLiteStore) AddDependency(ctx context.Context, dep *model.CardDependency) error {
	ts := now()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO card_dependencies (id, card_id, depends_on_card_id, created_at) VALUES (?, ?, ?, ?)",
		dep.ID, dep.CardID, dep.DependsOnCardID, ts)
	if err != nil {
		return err
	}
	dep.CreatedAt = parseTime(ts)
	return nil
}

func (s *SQLiteStore) RemoveDependency(ctx context.Context, cardID, dependsOnCardID string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM card_dependencies WHERE card_id = ? AND depends_on_card_id = ?",
		cardID, dependsOnCardID)
	return err
}

func (s *SQLiteStore) GetDependencies(ctx context.Context, cardID string) ([]model.Card, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.list_id, c.title, c.description, c.position, c.assignee, c.status, c.priority, c.due_date, c.created_at, c.updated_at
		 FROM cards c JOIN card_dependencies d ON c.id = d.depends_on_card_id WHERE d.card_id = ?`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []model.Card
	for rows.Next() {
		c, err := s.scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, *c)
	}
	return cards, rows.Err()
}

func (s *SQLiteStore) GetDependents(ctx context.Context, cardID string) ([]model.Card, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.list_id, c.title, c.description, c.position, c.assignee, c.status, c.priority, c.due_date, c.created_at, c.updated_at
		 FROM cards c JOIN card_dependencies d ON c.id = d.card_id WHERE d.depends_on_card_id = ?`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []model.Card
	for rows.Next() {
		c, err := s.scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, *c)
	}
	return cards, rows.Err()
}

// --- Labels ---

func (s *SQLiteStore) CreateLabel(ctx context.Context, label *model.Label) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO labels (id, board_id, name, color) VALUES (?, ?, ?, ?)",
		label.ID, label.BoardID, label.Name, label.Color)
	return err
}

func (s *SQLiteStore) GetLabel(ctx context.Context, id string) (*model.Label, error) {
	var l model.Label
	err := s.db.QueryRowContext(ctx,
		"SELECT id, board_id, name, color FROM labels WHERE id = ?", id).
		Scan(&l.ID, &l.BoardID, &l.Name, &l.Color)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("label not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *SQLiteStore) ListLabelsByBoard(ctx context.Context, boardID string) ([]model.Label, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, board_id, name, color FROM labels WHERE board_id = ?", boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var labels []model.Label
	for rows.Next() {
		var l model.Label
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

func (s *SQLiteStore) UpdateLabel(ctx context.Context, label *model.Label) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE labels SET name = ?, color = ? WHERE id = ?",
		label.Name, label.Color, label.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("label not found: %s", label.ID)
	}
	return nil
}

func (s *SQLiteStore) DeleteLabel(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM labels WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) AddLabelToCard(ctx context.Context, cardID, labelID string) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO card_labels (card_id, label_id) VALUES (?, ?)",
		cardID, labelID)
	return err
}

func (s *SQLiteStore) RemoveLabelFromCard(ctx context.Context, cardID, labelID string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM card_labels WHERE card_id = ? AND label_id = ?",
		cardID, labelID)
	return err
}

func (s *SQLiteStore) GetLabelsForCard(ctx context.Context, cardID string) ([]model.Label, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.id, l.board_id, l.name, l.color FROM labels l
		 JOIN card_labels cl ON l.id = cl.label_id WHERE cl.card_id = ?`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var labels []model.Label
	for rows.Next() {
		var l model.Label
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// --- Activity ---

func (s *SQLiteStore) CreateActivity(ctx context.Context, entry *model.ActivityLog) error {
	ts := now()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO activity_log (id, card_id, actor, action, detail, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		entry.ID, entry.CardID, entry.Actor, entry.Action, entry.Detail, ts)
	if err != nil {
		return err
	}
	entry.CreatedAt = parseTime(ts)
	return nil
}

func (s *SQLiteStore) ListActivityByCard(ctx context.Context, cardID string, limit int) ([]model.ActivityLog, error) {
	q := "SELECT id, card_id, actor, action, detail, created_at FROM activity_log WHERE card_id = ? ORDER BY created_at DESC"
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.QueryContext(ctx, q, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActivities(rows)
}

func (s *SQLiteStore) ListActivityByBoard(ctx context.Context, boardID string, limit int) ([]model.ActivityLog, error) {
	q := `SELECT a.id, a.card_id, a.actor, a.action, a.detail, a.created_at
		  FROM activity_log a
		  JOIN cards c ON a.card_id = c.id
		  JOIN lists l ON c.list_id = l.id
		  WHERE l.board_id = ?
		  ORDER BY a.created_at DESC`
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.QueryContext(ctx, q, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActivities(rows)
}

func scanActivities(rows *sql.Rows) ([]model.ActivityLog, error) {
	var entries []model.ActivityLog
	for rows.Next() {
		var e model.ActivityLog
		var createdAt string
		if err := rows.Scan(&e.ID, &e.CardID, &e.Actor, &e.Action, &e.Detail, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt = parseTime(createdAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
