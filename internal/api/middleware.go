package api

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func SetupMiddleware(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Accept"},
	}))

	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Printf("%s %s %d %s", c.Method(), c.Path(), c.Response().StatusCode(), time.Since(start))
		return err
	})
}
