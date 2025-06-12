package routes

import (
	"server/internal/logger"
	"server/internal/websockets"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func WebSocketRoutes(router fiber.Router, wsManager *websockets.Manager) {
	router.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	router.Get("/ws", websocket.New(func(c *websocket.Conn) {
		if wsManager != nil {
			wsManager.HandleWebSocket(c)
		} else {
			if err := c.Close(); err != nil {
				logger.New("Routes").File("websocket.routes.go").Er("failed to close connection", err)
			}
		}
	}))
}
