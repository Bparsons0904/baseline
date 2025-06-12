package routes

import (
	"net/http/httptest"
	"server/config"
	"server/internal/app"
	"server/internal/database"
	"server/internal/events"
	"server/internal/routes/middleware"
	"server/internal/websockets"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestApp() (*fiber.App, *app.App) {
	testConfig := config.Config{
		SecuritySalt:      12,
		SecurityPepper:    "test-pepper",
		SecurityJwtSecret: "test-jwt-secret",
		GeneralVersion:    "1.0.0",
	}
	config.ConfigInstance = testConfig

	mockDB := database.DB{}
	var mockWsManager *websockets.Manager = nil

	eventBus := events.New(nil, testConfig)
	testApp := &app.App{
		Config:     testConfig,
		Database:   mockDB,
		Websocket:  mockWsManager,
		Middleware: middleware.New(mockDB, eventBus, testConfig, nil, nil),
	}

	fiberApp := fiber.New()

	return fiberApp, testApp
}

func TestRouter_Setup(t *testing.T) {
	fiberApp, testApp := setupTestApp()

	err := Router(fiberApp, testApp)
	assert.NoError(t, err)

	// Test that routes are registered by checking the stack
	routes := fiberApp.GetRoutes()
	assert.NotEmpty(t, routes)

	// Check if health route is registered
	healthRouteFound := false
	wsRouteFound := false

	for _, route := range routes {
		if route.Path == "/api/health" && route.Method == "GET" {
			healthRouteFound = true
		}
		if route.Path == "/ws" && route.Method == "GET" {
			wsRouteFound = true
		}
	}

	assert.True(t, healthRouteFound, "Health route should be registered")
	assert.True(t, wsRouteFound, "WebSocket route should be registered")
}

func TestRouter_WebSocketUpgrade(t *testing.T) {
	fiberApp, testApp := setupTestApp()
	err := Router(fiberApp, testApp)
	require.NoError(t, err)

	// Test WebSocket upgrade required
	req := httptest.NewRequest("GET", "/ws", nil)
	resp, err := fiberApp.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

func TestRouter_WebSocketWithUpgradeHeader_Logic(t *testing.T) {
	// Test just the websocket upgrade logic without the actual handler
	fiberApp := fiber.New()

	// Replicate the middleware logic from setupWebSocketRoute
	fiberApp.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	fiberApp.Get("/ws", func(c *fiber.Ctx) error {
		return c.SendString("WebSocket upgrade detected")
	})

	// Test with WebSocket upgrade headers
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "test-key")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err := fiberApp.Test(req)
	require.NoError(t, err)

	// Should get 200 status for WebSocket upgrade
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSetupWebSocketRoute_MiddlewareLogic(t *testing.T) {
	fiberApp := fiber.New()

	// Test just the middleware part - don't actually call the websocket handler
	fiberApp.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	fiberApp.Get("/ws", func(c *fiber.Ctx) error {
		return c.SendString("WebSocket endpoint registered")
	})

	// Test non-WebSocket request
	req := httptest.NewRequest("GET", "/ws", nil)
	resp, err := fiberApp.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

func TestSetupWebSocketRoute_WithUpgrade(t *testing.T) {
	fiberApp := fiber.New()

	// Test just the middleware part - don't actually call the websocket handler
	fiberApp.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	fiberApp.Get("/ws", func(c *fiber.Ctx) error {
		return c.SendString("WebSocket endpoint with upgrade")
	})

	// Test with WebSocket upgrade
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err := fiberApp.Test(req)
	require.NoError(t, err)

	// Should not be upgrade required error
	assert.NotEqual(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

func TestRouter_WebSocketUpgradeCheck(t *testing.T) {
	fiberApp := fiber.New()

	// Create a test route similar to setupWebSocketRoute
	fiberApp.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	fiberApp.Get("/ws", func(c *fiber.Ctx) error {
		allowed := c.Locals("allowed")
		if allowed != nil && allowed.(bool) {
			return c.SendString("WebSocket allowed")
		}
		return c.SendString("Not allowed")
	})

	// Test without upgrade headers
	req := httptest.NewRequest("GET", "/ws", nil)
	resp, err := fiberApp.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)

	// Test with upgrade headers
	req = httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err = fiberApp.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRouter_RouterFunction_ReturnsNil(t *testing.T) {
	fiberApp, testApp := setupTestApp()

	err := Router(fiberApp, testApp)
	assert.NoError(t, err)
	assert.Nil(t, err)
}

func TestRouter_APIGroup(t *testing.T) {
	fiberApp, testApp := setupTestApp()

	err := Router(fiberApp, testApp)
	require.NoError(t, err)

	// Test that API group routes are accessible
	routes := fiberApp.GetRoutes()
	apiRoutes := make([]fiber.Route, 0)

	for _, route := range routes {
		if len(route.Path) >= 4 && route.Path[:4] == "/api" {
			apiRoutes = append(apiRoutes, route)
		}
	}

	assert.NotEmpty(t, apiRoutes, "API routes should be registered")
}

func TestRouter_UserRouteRegistration(t *testing.T) {
	fiberApp, testApp := setupTestApp()

	err := Router(fiberApp, testApp)
	require.NoError(t, err)

	// Test that user routes are registered
	routes := fiberApp.GetRoutes()
	userLoginFound := false

	for _, route := range routes {
		if route.Path == "/api/users/login" && route.Method == "POST" {
			userLoginFound = true
		}
	}

	assert.True(t, userLoginFound, "User login route should be registered")
}
