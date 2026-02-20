package api

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/aellingwood/cielo/internal/service"
)

func createCard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		listID := c.Params("listId")
		var body struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Assignee    string `json:"assignee"`
			Priority    string `json:"priority"`
			Position    int    `json:"position"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		card, err := svc.CreateCard(c.Context(), listID, body.Title, body.Description, body.Assignee, body.Priority, "user", body.Position)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(card)
	}
}

func getCard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		card, err := svc.GetCard(c.Context(), id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(card)
	}
}

func updateCard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		var body map[string]any
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		card, err := svc.UpdateCard(c.Context(), id, body, "user")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(card)
	}
}

func deleteCard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		if err := svc.DeleteCard(c.Context(), id, "user"); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(204)
	}
}

func moveCard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		var body struct {
			ListID   string `json:"list_id"`
			Position int    `json:"position"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		card, err := svc.MoveCard(c.Context(), id, body.ListID, body.Position, "user")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(card)
	}
}

func assignCard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		var body struct {
			Assignee string `json:"assignee"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		card, err := svc.AssignCard(c.Context(), id, body.Assignee, "user")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(card)
	}
}

func addDependency(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		var body struct {
			DependsOnCardID string `json:"depends_on_card_id"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		if err := svc.AddDependency(c.Context(), id, body.DependsOnCardID, "user"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(201)
	}
}

func removeDependency(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		depID := c.Params("depId")
		if err := svc.RemoveDependency(c.Context(), id, depID, "user"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(204)
	}
}

func listLabels(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		boardID := c.Params("boardId")
		labels, err := svc.ListLabelsByBoard(c.Context(), boardID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if labels == nil {
			return c.JSON([]any{})
		}
		return c.JSON(labels)
	}
}

func createLabel(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		boardID := c.Params("boardId")
		var body struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		l, err := svc.CreateLabel(c.Context(), boardID, body.Name, body.Color)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(l)
	}
}

func updateLabel(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		var body struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		l, err := svc.UpdateLabel(c.Context(), id, body.Name, body.Color)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(l)
	}
}

func deleteLabel(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		if err := svc.DeleteLabel(c.Context(), id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(204)
	}
}

func addLabelToCard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		cardID := c.Params("id")
		var body struct {
			LabelID string `json:"label_id"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		if err := svc.AddLabelToCard(c.Context(), cardID, body.LabelID, "user"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(201)
	}
}

func removeLabelFromCard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		cardID := c.Params("id")
		labelID := c.Params("labelId")
		if err := svc.RemoveLabelFromCard(c.Context(), cardID, labelID, "user"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(204)
	}
}

func getCardActivity(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		limit, _ := strconv.Atoi(c.Query("limit", "50"))
		entries, err := svc.ListActivityByCard(c.Context(), id, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if entries == nil {
			return c.JSON([]any{})
		}
		return c.JSON(entries)
	}
}

func getBoardActivity(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		boardID := c.Params("boardId")
		limit, _ := strconv.Atoi(c.Query("limit", "50"))
		entries, err := svc.ListActivityByBoard(c.Context(), boardID, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if entries == nil {
			return c.JSON([]any{})
		}
		return c.JSON(entries)
	}
}

func searchCards(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		boardID := c.Params("boardId")
		q := c.Query("q")
		assignee := c.Query("assignee")
		status := c.Query("status")
		label := c.Query("label")
		cards, err := svc.SearchCards(c.Context(), boardID, q, assignee, status, label)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if cards == nil {
			return c.JSON([]any{})
		}
		return c.JSON(cards)
	}
}
