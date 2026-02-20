package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/aellingwood/cielo/internal/api"
	"github.com/aellingwood/cielo/internal/config"
	"github.com/aellingwood/cielo/internal/event"
	"github.com/aellingwood/cielo/internal/mcp"
	"github.com/aellingwood/cielo/internal/service"
	"github.com/aellingwood/cielo/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA journal_mode = WAL")

	if err := store.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	sqliteStore := store.NewSQLiteStore(db)
	bus := event.NewBus()
	svc := service.New(sqliteStore, bus)
	mcpServer := mcp.NewServer(svc)

	app := fiber.New(fiber.Config{
		AppName: "Cielo",
	})

	api.SetupMiddleware(app)
	api.SetupRouter(app, svc, bus, mcpServer)

	// Serve frontend static files if the dist directory exists
	if _, err := os.Stat("web/dist"); err == nil {
		app.Use("/", static.New("web/dist"))

		// SPA fallback: serve index.html for any non-API route
		app.Get("/*", func(c fiber.Ctx) error {
			return c.SendFile("web/dist/index.html")
		})
	}

	log.Printf("Cielo starting on %s", cfg.HTTPAddr)
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
