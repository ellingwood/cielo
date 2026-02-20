package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/aellingwood/cielo/internal/event"
	"github.com/aellingwood/cielo/internal/mcp"
	"github.com/aellingwood/cielo/internal/service"
)

func SetupRouter(app *fiber.App, svc *service.Service, bus *event.Bus, mcpServer *mcp.Server) {
	api := app.Group("/api/v1")

	api.Get("/boards", listBoards(svc))
	api.Post("/boards", createBoard(svc))
	api.Get("/boards/:id", getBoard(svc))
	api.Put("/boards/:id", updateBoard(svc))
	api.Delete("/boards/:id", deleteBoard(svc))

	api.Post("/boards/:boardId/lists", createList(svc))
	api.Put("/lists/:id", updateList(svc))
	api.Delete("/lists/:id", deleteList(svc))

	api.Post("/lists/:listId/cards", createCard(svc))
	api.Get("/cards/:id", getCard(svc))
	api.Put("/cards/:id", updateCard(svc))
	api.Delete("/cards/:id", deleteCard(svc))
	api.Put("/cards/:id/move", moveCard(svc))
	api.Put("/cards/:id/assign", assignCard(svc))

	api.Post("/cards/:id/dependencies", addDependency(svc))
	api.Delete("/cards/:id/dependencies/:depId", removeDependency(svc))

	api.Get("/boards/:boardId/labels", listLabels(svc))
	api.Post("/boards/:boardId/labels", createLabel(svc))
	api.Put("/labels/:id", updateLabel(svc))
	api.Delete("/labels/:id", deleteLabel(svc))
	api.Post("/cards/:id/labels", addLabelToCard(svc))
	api.Delete("/cards/:id/labels/:labelId", removeLabelFromCard(svc))

	api.Get("/cards/:id/activity", getCardActivity(svc))
	api.Get("/boards/:boardId/activity", getBoardActivity(svc))
	api.Get("/boards/:boardId/search", searchCards(svc))

	api.Get("/boards/:boardId/events", boardSSE(bus))

	app.Post("/mcp", mcpHandler(mcpServer))
}
