package store

import (
	"context"

	"github.com/aellingwood/cielo/internal/model"
)

type Store interface {
	CreateBoard(ctx context.Context, board *model.Board) error
	GetBoard(ctx context.Context, id string) (*model.Board, error)
	ListBoards(ctx context.Context) ([]model.Board, error)
	UpdateBoard(ctx context.Context, board *model.Board) error
	DeleteBoard(ctx context.Context, id string) error

	CreateList(ctx context.Context, list *model.List) error
	GetList(ctx context.Context, id string) (*model.List, error)
	ListListsByBoard(ctx context.Context, boardID string) ([]model.List, error)
	UpdateList(ctx context.Context, list *model.List) error
	DeleteList(ctx context.Context, id string) error

	CreateCard(ctx context.Context, card *model.Card) error
	GetCard(ctx context.Context, id string) (*model.Card, error)
	ListCardsByList(ctx context.Context, listID string) ([]model.Card, error)
	UpdateCard(ctx context.Context, card *model.Card) error
	MoveCard(ctx context.Context, cardID, targetListID string, position int) error
	DeleteCard(ctx context.Context, id string) error
	SearchCards(ctx context.Context, boardID, query, assignee, status, label string) ([]model.Card, error)

	AddDependency(ctx context.Context, dep *model.CardDependency) error
	RemoveDependency(ctx context.Context, cardID, dependsOnCardID string) error
	GetDependencies(ctx context.Context, cardID string) ([]model.Card, error)
	GetDependents(ctx context.Context, cardID string) ([]model.Card, error)

	CreateLabel(ctx context.Context, label *model.Label) error
	GetLabel(ctx context.Context, id string) (*model.Label, error)
	ListLabelsByBoard(ctx context.Context, boardID string) ([]model.Label, error)
	UpdateLabel(ctx context.Context, label *model.Label) error
	DeleteLabel(ctx context.Context, id string) error
	AddLabelToCard(ctx context.Context, cardID, labelID string) error
	RemoveLabelFromCard(ctx context.Context, cardID, labelID string) error
	GetLabelsForCard(ctx context.Context, cardID string) ([]model.Label, error)

	CreateActivity(ctx context.Context, entry *model.ActivityLog) error
	ListActivityByCard(ctx context.Context, cardID string, limit int) ([]model.ActivityLog, error)
	ListActivityByBoard(ctx context.Context, boardID string, limit int) ([]model.ActivityLog, error)
}
