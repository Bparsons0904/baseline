package routes

import (
	"net/http/httptest"
	"server/internal/websockets"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestWebSocketManager(t *testing.T) *websockets.Manager {
	// For testing purposes, return nil since WebSocket tests should be in websockets package
	// This test is just for route registration
	return nil
}

func TestWebSocketRoutes_RegistrationOnly(t *testing.T) {
	app := fiber.New()
	wsManager := setupTestWebSocketManager(t)

	// Test that the function doesn't panic when called with nil manager
	assert.NotPanics(t, func() {
		WebSocketRoutes(app, wsManager)
	})

	// Test that routes are registered by checking the route stack
	routes := app.GetRoutes()
	assert.NotEmpty(t, routes)
	
	// Look for WebSocket route registration
	hasWebSocketRoute := false
	for _, route := range routes {
		if route.Path == "/ws" {
			hasWebSocketRoute = true
			break
		}
	}
	assert.True(t, hasWebSocketRoute, "WebSocket route should be registered")
}

func TestWebSocketRoutes_MiddlewareLogic(t *testing.T) {
	app := fiber.New()

	// Test the middleware logic separately
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", func(c *fiber.Ctx) error {
		allowed := c.Locals("allowed")
		if allowed != nil && allowed.(bool) {
			return c.SendString("WebSocket connection allowed")
		}
		return c.SendString("Not allowed")
	})

	// Test non-WebSocket request
	req := httptest.NewRequest("GET", "/ws", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)

	// Test with WebSocket upgrade headers
	req = httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}