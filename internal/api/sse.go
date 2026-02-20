package api

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/aellingwood/cielo/internal/event"
	"github.com/aellingwood/cielo/internal/mcp"
)

func boardSSE(bus *event.Bus) fiber.Handler {
	return func(c fiber.Ctx) error {
		boardID := c.Params("boardId")
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")

		sub := bus.Subscribe(boardID)

		return c.SendStreamWriter(func(w *bufio.Writer) {
			defer bus.Unsubscribe(sub)
			for evt := range sub.Ch {
				data, _ := json.Marshal(evt.Payload)
				fmt.Fprintf(w, "id: %d\n", evt.SeqID)
				fmt.Fprintf(w, "event: %s\n", evt.Type)
				fmt.Fprintf(w, "data: %s\n\n", data)
				if err := w.Flush(); err != nil {
					return
				}
			}
		})
	}
}

func mcpHandler(server *mcp.Server) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req mcp.JSONRPCRequest
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   &mcp.RPCError{Code: -32700, Message: "Parse error"},
			})
		}
		resp := server.HandleRequest(c.Context(), req)
		return c.JSON(resp)
	}
}
