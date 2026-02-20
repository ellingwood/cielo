package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/aellingwood/cielo/internal/service"
)

func listBoards(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		boards, err := svc.ListBoards(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if boards == nil {
			return c.JSON([]any{})
		}
		return c.JSON(boards)
	}
}

func createBoard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		var body struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		b, err := svc.CreateBoard(c.Context(), body.Name, body.Description, "user")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(b)
	}
}

func getBoard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		b, err := svc.GetBoard(c.Context(), id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		lists, err := svc.ListListsByBoard(c.Context(), id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if lists == nil {
			return c.JSON(fiber.Map{
				"id": b.ID, "name": b.Name, "description": b.Description,
				"created_at": b.CreatedAt, "updated_at": b.UpdatedAt, "lists": []any{},
			})
		}
		return c.JSON(fiber.Map{
			"id":          b.ID,
			"name":        b.Name,
			"description": b.Description,
			"created_at":  b.CreatedAt,
			"updated_at":  b.UpdatedAt,
			"lists":       lists,
		})
	}
}

func updateBoard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		var body struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		b, err := svc.UpdateBoard(c.Context(), id, body.Name, body.Description)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(b)
	}
}

func deleteBoard(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		if err := svc.DeleteBoard(c.Context(), id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(204)
	}
}
