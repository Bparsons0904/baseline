package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"server/config"
	"server/internal/database"
	"server/internal/events"
	"server/internal/models"
	"server/internal/utils"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)


func setupAuthMiddlewareTest() (Middleware, config.Config, *MockUserRepository, *MockSessionRepository) {
	testConfig := config.Config{
		SecuritySalt:      12,
		SecurityPepper:    "test-pepper",
		SecurityJwtSecret: "test-jwt-secret-key-for-testing",
	}
	config.ConfigInstance = testConfig

	// Mock database
	mockDB := database.DB{}

	// Mock repositories
	mockUserRepo := &MockUserRepository{}
	mockSessionRepo := &MockSessionRepository{}

	eventBus := &events.EventBus{}
	middleware := New(mockDB, eventBus, testConfig, mockUserRepo, mockSessionRepo)

	return middleware, testConfig, mockUserRepo, mockSessionRepo
}

func TestMiddleware_BasicAuth_NoClientType(t *testing.T) {
	middleware, _, _, _ := setupAuthMiddlewareTest()
	app := fiber.New()

	app.Get("/test", middleware.BasicAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestMiddleware_BasicAuth_WebClient_NoCookie(t *testing.T) {
	middleware, _, _, mockSessionRepo := setupAuthMiddlewareTest()

	// Setup mock to return empty session when no cookie
	mockSessionRepo.On("GetByID", mock.Anything, "").Return((*models.Session)(nil), errors.New("session not found"))
	app := fiber.New()

	app.Get("/test", middleware.BasicAuth(), func(c *fiber.Ctx) error {
		authenticated := c.Locals("authenticated").(bool)
		return c.JSON(fiber.Map{"authenticated": authenticated})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Client-Type", "solid")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.False(t, result["authenticated"].(bool))
}

func TestMiddleware_BasicAuth_MobileClient_NoToken(t *testing.T) {
	middleware, _, _, mockSessionRepo := setupAuthMiddlewareTest()

	// Setup mock to handle session deletion in defer when error occurs
	mockSessionRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
	app := fiber.New()

	app.Get("/test", middleware.BasicAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Client-Type", "flutter")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestMiddleware_BasicAuth_MobileClient_InvalidToken(t *testing.T) {
	middleware, _, _, mockSessionRepo := setupAuthMiddlewareTest()

	// Setup mock to handle session deletion in defer when error occurs
	mockSessionRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
	app := fiber.New()

	app.Get("/test", middleware.BasicAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Client-Type", "flutter")
	req.Header.Set("Authorization", "invalid-token")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestMiddleware_AuthRequired_NotAuthenticated(t *testing.T) {
	middleware, _, _, _ := setupAuthMiddlewareTest()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("authenticated", false)
		return c.Next()
	}, middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "Authentication required", result["error"])
}

func TestMiddleware_AuthRequired_Authenticated(t *testing.T) {
	middleware, _, _, _ := setupAuthMiddlewareTest()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("authenticated", true)
		return c.Next()
	}, middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "success", result["message"])
}

func TestMiddleware_AuthNoContent_NotAuthenticated(t *testing.T) {
	middleware, _, _, _ := setupAuthMiddlewareTest()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("authenticated", false)
		return c.Next()
	}, middleware.AuthNoContent(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestMiddleware_AuthNoContent_Authenticated(t *testing.T) {
	middleware, _, _, _ := setupAuthMiddlewareTest()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("authenticated", true)
		return c.Next()
	}, middleware.AuthNoContent(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMiddleware_getWebSessionData_NoSessionCookie(t *testing.T) {
	middleware, _, _, _ := setupAuthMiddlewareTest()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		session, err := middleware.getWebSessionData(c)
		assert.NoError(t, err)
		assert.Equal(t, models.Session{}, session)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMiddleware_getMobileSessionData_NoAuthHeader(t *testing.T) {
	middleware, _, _, _ := setupAuthMiddlewareTest()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		session, err := middleware.getMobileSessionData(c)
		assert.Error(t, err)
		assert.Equal(t, models.Session{}, session)
		assert.Contains(t, err.Error(), "No token found")
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMiddleware_getMobileSessionData_InvalidToken(t *testing.T) {
	middleware, _, _, _ := setupAuthMiddlewareTest()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		session, err := middleware.getMobileSessionData(c)
		assert.Error(t, err)
		assert.Equal(t, models.Session{}, session)
		assert.Contains(t, err.Error(), "token contains an invalid number of segments")
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "invalid-token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMiddleware_JWT_TokenValidation(t *testing.T) {
	_, testConfig, _, _ := setupAuthMiddlewareTest()

	// Test valid token parsing
	userID := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour)
	token, err := utils.GenerateJWTToken(userID, expiresAt, "test-issuer", testConfig)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Test token parsing
	claims, err := utils.ParseJWTToken(token, testConfig)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID.String())
}

func TestMiddleware_Constants(t *testing.T) {
	assert.Equal(t, "flutter", MOBILE_CLIENT_TYPE)
	assert.Equal(t, "solid", WEB_CLIENT_TYPE)
}

func TestMiddleware_SessionData_Structure(t *testing.T) {
	userID := uuid.New()
	expiresAt := time.Now().Add(time.Hour)
	userAgent := "test-user-agent"

	sessionData := SessionData{
		UserID:    userID,
		ExpiresAt: expiresAt,
		UserAgent: userAgent,
	}

	assert.Equal(t, userID, sessionData.UserID)
	assert.Equal(t, expiresAt, sessionData.ExpiresAt)
	assert.Equal(t, userAgent, sessionData.UserAgent)
}

func TestMiddleware_AuthMiddlewareNew(t *testing.T) {
	testConfig := config.Config{
		SecuritySalt:      12,
		SecurityPepper:    "test-pepper",
		SecurityJwtSecret: "test-secret",
	}

	mockDB := database.DB{}
	mockUserRepo := &MockUserRepository{}
	mockSessionRepo := &MockSessionRepository{}

	eventBus := &events.EventBus{}
	middleware := New(mockDB, eventBus, testConfig, mockUserRepo, mockSessionRepo)

	assert.Equal(t, mockDB, middleware.DB)
	assert.Equal(t, testConfig, middleware.Config)
	assert.NotNil(t, middleware.log)
}
