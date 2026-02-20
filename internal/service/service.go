package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aellingwood/cielo/internal/event"
	"github.com/aellingwood/cielo/internal/model"
	"github.com/aellingwood/cielo/internal/store"
)

type Service struct {
	store store.Store
	bus   *event.Bus
}

func New(s store.Store, bus *event.Bus) *Service {
	return &Service{store: s, bus: bus}
}

func (s *Service) publish(typ, boardID string, payload any) {
	s.bus.Publish(event.Event{Type: typ, BoardID: boardID, Payload: payload})
}

func (s *Service) logActivity(ctx context.Context, cardID, actor, action string, detail any) {
	d, _ := json.Marshal(detail)
	s.store.CreateActivity(ctx, &model.ActivityLog{
		ID:     model.NewID(),
		CardID: cardID,
		Actor:  actor,
		Action: action,
		Detail: string(d),
	})
}

// --- Boards ---

func (s *Service) CreateBoard(ctx context.Context, name, description, actor string) (*model.Board, error) {
	if name == "" {
		return nil, fmt.Errorf("board name is required")
	}
	b := &model.Board{ID: model.NewID(), Name: name, Description: description}
	if err := s.store.CreateBoard(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) GetBoard(ctx context.Context, id string) (*model.Board, error) {
	return s.store.GetBoard(ctx, id)
}

func (s *Service) ListBoards(ctx context.Context) ([]model.Board, error) {
	return s.store.ListBoards(ctx)
}

func (s *Service) UpdateBoard(ctx context.Context, id, name, description string) (*model.Board, error) {
	b, err := s.store.GetBoard(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		b.Name = name
	}
	b.Description = description
	if err := s.store.UpdateBoard(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) DeleteBoard(ctx context.Context, id string) error {
	return s.store.DeleteBoard(ctx, id)
}

// --- Lists ---

func (s *Service) CreateList(ctx context.Context, boardID, name string, position int, actor string) (*model.List, error) {
	if name == "" {
		return nil, fmt.Errorf("list name is required")
	}
	l := &model.List{ID: model.NewID(), BoardID: boardID, Name: name, Position: position}
	if err := s.store.CreateList(ctx, l); err != nil {
		return nil, err
	}
	s.publish("list.created", boardID, l)
	return l, nil
}

func (s *Service) GetList(ctx context.Context, id string) (*model.List, error) {
	return s.store.GetList(ctx, id)
}

func (s *Service) ListListsByBoard(ctx context.Context, boardID string) ([]model.List, error) {
	lists, err := s.store.ListListsByBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	for i := range lists {
		cards, err := s.store.ListCardsByList(ctx, lists[i].ID)
		if err != nil {
			return nil, err
		}
		if cards == nil {
			cards = []model.Card{}
		}
		for j := range cards {
			labels, err := s.store.GetLabelsForCard(ctx, cards[j].ID)
			if err != nil {
				return nil, err
			}
			if labels == nil {
				labels = []model.Label{}
			}
			cards[j].Labels = labels
		}
		lists[i].Cards = cards
	}
	return lists, nil
}

func (s *Service) UpdateList(ctx context.Context, id, name string, position int) (*model.List, error) {
	l, err := s.store.GetList(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		l.Name = name
	}
	l.Position = position
	if err := s.store.UpdateList(ctx, l); err != nil {
		return nil, err
	}
	s.publish("list.updated", l.BoardID, l)
	return l, nil
}

func (s *Service) DeleteList(ctx context.Context, id string) error {
	l, err := s.store.GetList(ctx, id)
	if err != nil {
		return err
	}
	if err := s.store.DeleteList(ctx, id); err != nil {
		return err
	}
	s.publish("list.deleted", l.BoardID, map[string]string{"id": id})
	return nil
}

// --- Cards ---

func (s *Service) CreateCard(ctx context.Context, listID, title, description, assignee, priority, actor string, position int) (*model.Card, error) {
	if title == "" {
		return nil, fmt.Errorf("card title is required")
	}
	if priority == "" {
		priority = model.PriorityMedium
	}
	if !model.ValidPriority(priority) {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}
	status := model.StatusUnassigned
	if assignee != "" {
		status = model.StatusAssigned
	}
	c := &model.Card{
		ID:          model.NewID(),
		ListID:      listID,
		Title:       title,
		Description: description,
		Position:    position,
		Assignee:    assignee,
		Status:      status,
		Priority:    priority,
	}
	if err := s.store.CreateCard(ctx, c); err != nil {
		return nil, err
	}
	c.Labels = []model.Label{}
	l, _ := s.store.GetList(ctx, listID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	s.logActivity(ctx, c.ID, actor, model.ActionCreated, map[string]string{"title": title})
	s.publish("card.created", boardID, c)
	return c, nil
}

func (s *Service) GetCard(ctx context.Context, id string) (*model.Card, error) {
	c, err := s.store.GetCard(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Labels, _ = s.store.GetLabelsForCard(ctx, id)
	if c.Labels == nil {
		c.Labels = []model.Label{}
	}
	c.Dependencies, _ = s.store.GetDependencies(ctx, id)
	if c.Dependencies == nil {
		c.Dependencies = []model.Card{}
	}
	c.Dependents, _ = s.store.GetDependents(ctx, id)
	if c.Dependents == nil {
		c.Dependents = []model.Card{}
	}
	c.Activity, _ = s.store.ListActivityByCard(ctx, id, 50)
	if c.Activity == nil {
		c.Activity = []model.ActivityLog{}
	}
	return c, nil
}

func (s *Service) UpdateCard(ctx context.Context, id string, updates map[string]any, actor string) (*model.Card, error) {
	c, err := s.store.GetCard(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := updates["title"].(string); ok && v != "" {
		c.Title = v
	}
	if v, ok := updates["description"].(string); ok {
		c.Description = v
	}
	if v, ok := updates["assignee"].(string); ok {
		c.Assignee = v
	}
	if v, ok := updates["status"].(string); ok {
		if !model.ValidStatus(v) {
			return nil, fmt.Errorf("invalid status: %s", v)
		}
		c.Status = v
	}
	if v, ok := updates["priority"].(string); ok {
		if !model.ValidPriority(v) {
			return nil, fmt.Errorf("invalid priority: %s", v)
		}
		c.Priority = v
	}
	if err := s.store.UpdateCard(ctx, c); err != nil {
		return nil, err
	}
	l, _ := s.store.GetList(ctx, c.ListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	s.logActivity(ctx, c.ID, actor, model.ActionStatusChanged, updates)
	s.publish("card.updated", boardID, c)
	return s.GetCard(ctx, id)
}

func (s *Service) MoveCard(ctx context.Context, cardID, targetListID string, position int, actor string) (*model.Card, error) {
	c, err := s.store.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}
	fromListID := c.ListID
	if err := s.store.MoveCard(ctx, cardID, targetListID, position); err != nil {
		return nil, err
	}
	l, _ := s.store.GetList(ctx, targetListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	s.logActivity(ctx, cardID, actor, model.ActionMoved, map[string]string{
		"from_list": fromListID, "to_list": targetListID,
	})
	s.publish("card.moved", boardID, map[string]any{
		"card_id": cardID, "from_list": fromListID, "to_list": targetListID, "position": position,
	})
	return s.GetCard(ctx, cardID)
}

func (s *Service) AssignCard(ctx context.Context, cardID, assignee, actor string) (*model.Card, error) {
	c, err := s.store.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}
	oldAssignee := c.Assignee
	c.Assignee = assignee
	if assignee != "" && c.Status == model.StatusUnassigned {
		c.Status = model.StatusAssigned
	}
	if assignee == "" && c.Status == model.StatusAssigned {
		c.Status = model.StatusUnassigned
	}
	if err := s.store.UpdateCard(ctx, c); err != nil {
		return nil, err
	}
	l, _ := s.store.GetList(ctx, c.ListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	action := model.ActionAssigned
	if assignee == "" {
		action = model.ActionUnassigned
	}
	s.logActivity(ctx, cardID, actor, action, map[string]string{
		"from": oldAssignee, "to": assignee,
	})
	s.publish("card.updated", boardID, c)
	return s.GetCard(ctx, cardID)
}

func (s *Service) DeleteCard(ctx context.Context, id, actor string) error {
	c, err := s.store.GetCard(ctx, id)
	if err != nil {
		return err
	}
	l, _ := s.store.GetList(ctx, c.ListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	if err := s.store.DeleteCard(ctx, id); err != nil {
		return err
	}
	s.publish("card.deleted", boardID, map[string]string{"id": id})
	return nil
}

func (s *Service) SearchCards(ctx context.Context, boardID, query, assignee, status, label string) ([]model.Card, error) {
	return s.store.SearchCards(ctx, boardID, query, assignee, status, label)
}

// --- Dependencies ---

func (s *Service) AddDependency(ctx context.Context, cardID, dependsOnCardID, actor string) error {
	if cardID == dependsOnCardID {
		return fmt.Errorf("card cannot depend on itself")
	}
	dep := &model.CardDependency{
		ID: model.NewID(), CardID: cardID, DependsOnCardID: dependsOnCardID,
	}
	if err := s.store.AddDependency(ctx, dep); err != nil {
		return err
	}
	c, _ := s.store.GetCard(ctx, cardID)
	l, _ := s.store.GetList(ctx, c.ListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	s.logActivity(ctx, cardID, actor, model.ActionDependencyAdded, map[string]string{
		"depends_on": dependsOnCardID,
	})
	s.publish("card.updated", boardID, c)
	return nil
}

func (s *Service) RemoveDependency(ctx context.Context, cardID, dependsOnCardID, actor string) error {
	if err := s.store.RemoveDependency(ctx, cardID, dependsOnCardID); err != nil {
		return err
	}
	c, _ := s.store.GetCard(ctx, cardID)
	l, _ := s.store.GetList(ctx, c.ListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	s.logActivity(ctx, cardID, actor, model.ActionDependencyRemoved, map[string]string{
		"depends_on": dependsOnCardID,
	})
	s.publish("card.updated", boardID, c)
	return nil
}

func (s *Service) GetDependencies(ctx context.Context, cardID string) ([]model.Card, error) {
	return s.store.GetDependencies(ctx, cardID)
}

func (s *Service) GetDependents(ctx context.Context, cardID string) ([]model.Card, error) {
	return s.store.GetDependents(ctx, cardID)
}

// --- Labels ---

func (s *Service) CreateLabel(ctx context.Context, boardID, name, color string) (*model.Label, error) {
	if name == "" {
		return nil, fmt.Errorf("label name is required")
	}
	if color == "" {
		color = "#6b7280"
	}
	l := &model.Label{ID: model.NewID(), BoardID: boardID, Name: name, Color: color}
	if err := s.store.CreateLabel(ctx, l); err != nil {
		return nil, err
	}
	s.publish("label.created", boardID, l)
	return l, nil
}

func (s *Service) ListLabelsByBoard(ctx context.Context, boardID string) ([]model.Label, error) {
	return s.store.ListLabelsByBoard(ctx, boardID)
}

func (s *Service) UpdateLabel(ctx context.Context, id, name, color string) (*model.Label, error) {
	l, err := s.store.GetLabel(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		l.Name = name
	}
	if color != "" {
		l.Color = color
	}
	if err := s.store.UpdateLabel(ctx, l); err != nil {
		return nil, err
	}
	s.publish("label.updated", l.BoardID, l)
	return l, nil
}

func (s *Service) DeleteLabel(ctx context.Context, id string) error {
	l, err := s.store.GetLabel(ctx, id)
	if err != nil {
		return err
	}
	if err := s.store.DeleteLabel(ctx, id); err != nil {
		return err
	}
	s.publish("label.deleted", l.BoardID, map[string]string{"id": id})
	return nil
}

func (s *Service) AddLabelToCard(ctx context.Context, cardID, labelID, actor string) error {
	if err := s.store.AddLabelToCard(ctx, cardID, labelID); err != nil {
		return err
	}
	c, _ := s.store.GetCard(ctx, cardID)
	l, _ := s.store.GetList(ctx, c.ListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	s.logActivity(ctx, cardID, actor, model.ActionLabelAdded, map[string]string{"label_id": labelID})
	s.publish("card.updated", boardID, c)
	return nil
}

func (s *Service) RemoveLabelFromCard(ctx context.Context, cardID, labelID, actor string) error {
	if err := s.store.RemoveLabelFromCard(ctx, cardID, labelID); err != nil {
		return err
	}
	c, _ := s.store.GetCard(ctx, cardID)
	l, _ := s.store.GetList(ctx, c.ListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	s.logActivity(ctx, cardID, actor, model.ActionLabelRemoved, map[string]string{"label_id": labelID})
	s.publish("card.updated", boardID, c)
	return nil
}

// --- Activity ---

func (s *Service) AddComment(ctx context.Context, cardID, actor, text string) error {
	if text == "" {
		return fmt.Errorf("comment text is required")
	}
	s.logActivity(ctx, cardID, actor, model.ActionComment, map[string]string{"text": text})
	c, _ := s.store.GetCard(ctx, cardID)
	l, _ := s.store.GetList(ctx, c.ListID)
	boardID := ""
	if l != nil {
		boardID = l.BoardID
	}
	s.publish("activity.new", boardID, map[string]string{"card_id": cardID, "actor": actor})
	return nil
}

func (s *Service) ListActivityByCard(ctx context.Context, cardID string, limit int) ([]model.ActivityLog, error) {
	return s.store.ListActivityByCard(ctx, cardID, limit)
}

func (s *Service) ListActivityByBoard(ctx context.Context, boardID string, limit int) ([]model.ActivityLog, error) {
	return s.store.ListActivityByBoard(ctx, boardID, limit)
}
