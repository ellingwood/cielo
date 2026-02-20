package model

import (
	"time"

	"github.com/google/uuid"
)

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
	ID           string        `json:"id"`
	ListID       string        `json:"list_id"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Position     int           `json:"position"`
	Assignee     string        `json:"assignee"`
	Status       string        `json:"status"`
	Priority     string        `json:"priority"`
	DueDate      *time.Time    `json:"due_date,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Labels       []Label       `json:"labels,omitempty"`
	Dependencies []Card        `json:"dependencies,omitempty"`
	Dependents   []Card        `json:"dependents,omitempty"`
	Activity     []ActivityLog `json:"activity,omitempty"`
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

const (
	StatusUnassigned = "unassigned"
	StatusAssigned   = "assigned"
	StatusInProgress = "in_progress"
	StatusBlocked    = "blocked"
	StatusDone       = "done"
)

const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

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

func ValidStatus(s string) bool {
	switch s {
	case StatusUnassigned, StatusAssigned, StatusInProgress, StatusBlocked, StatusDone:
		return true
	}
	return false
}

func ValidPriority(p string) bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	}
	return false
}
