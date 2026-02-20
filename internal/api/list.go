package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/aellingwood/cielo/internal/service"
)

func createList(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		boardID := c.Params("boardId")
		var body struct {
			Name     string `json:"name"`
			Position int    `json:"position"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		l, err := svc.CreateList(c.Context(), boardID, body.Name, body.Position, "user")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(l)
	}
}

func updateList(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		var body struct {
			Name     string `json:"name"`
			Position int    `json:"position"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		l, err := svc.UpdateList(c.Context(), id, body.Name, body.Position)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(l)
	}
}

func deleteList(svc *service.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		if err := svc.DeleteList(c.Context(), id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(204)
	}
}
