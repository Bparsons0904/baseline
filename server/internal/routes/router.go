package routes

import (
	"server/internal/app"
	"server/internal/logger"
	"server/internal/routes/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Route struct {
	middleware middleware.Middleware
	log        logger.Logger
	router     fiber.Router
}

func Router(router fiber.Router, app *app.App) (err error) {
	setupWebSocketRoute(router, app)

	api := router.Group("/api")
	HealthRoutes(api, app.Config)
	NewUserRoute(*app, api).Register()
	api.Use(app.Middleware.BasicAuth())
	NewAdminRoute(*app, api).Register()

	return nil
}

func setupWebSocketRoute(router fiber.Router, app *app.App) {
	router.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	router.Get("/ws", websocket.New(func(c *websocket.Conn) {
		app.Websocket.HandleWebSocket(c)
	}))
}
