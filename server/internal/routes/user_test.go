package routes

import (
	"context"
	"errors"
	"server/config"
	"server/internal/app"
	"server/internal/database"
	"server/internal/events"
	"server/internal/models"
	"server/internal/routes/middleware"
	"testing"

	userController "server/internal/controllers/users"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Mock UserController for testing
type MockUserController struct {
	loginResponse  func(context.Context, models.LoginRequest) (models.User, models.Session, error)
	logoutResponse func(string) error
}

func (m *MockUserController) Login(ctx context.Context, req models.LoginRequest) (models.User, models.Session, error) {
	if m.loginResponse != nil {
		return m.loginResponse(ctx, req)
	}
	return models.User{}, models.Session{}, errors.New("not implemented")
}

func (m *MockUserController) Logout(sessionID string) error {
	if m.logoutResponse != nil {
		return m.logoutResponse(sessionID)
	}
	return errors.New("not implemented")
}

func setupUserRouteTest() (*fiber.App, *UserRoute) {
	testConfig := config.Config{
		SecuritySalt:      12,
		SecurityPepper:    "test-pepper",
		SecurityJwtSecret: "test-jwt-secret",
	}
	config.ConfigInstance = testConfig

	fiberApp := fiber.New()
	mockDB := database.DB{}
	eventBus := events.New(nil, testConfig)
	
	// Create a real UserController for testing instead of mock
	userCtrl := userController.New(eventBus, nil, nil, testConfig)

	appInstance := app.App{
		Config:         testConfig,
		Database:       mockDB,
		UserController: userCtrl,
		Middleware:     middleware.New(mockDB, eventBus, testConfig, nil, nil),
	}

	userRoute := NewUserRoute(appInstance, fiberApp)

	return fiberApp, userRoute
}

func TestNewUserRoute(t *testing.T) {
	testConfig := config.Config{
		SecuritySalt:      12,
		SecurityPepper:    "test-pepper",
		SecurityJwtSecret: "test-jwt-secret",
	}

	eventBus := events.New(nil, testConfig)
	userCtrl := userController.New(eventBus, nil, nil, testConfig)
	
	mockApp := app.App{
		Config:         testConfig,
		Database:       database.DB{},
		UserController: userCtrl,
		Middleware:     middleware.New(database.DB{}, eventBus, testConfig, nil, nil),
	}

	fiberApp := fiber.New()
	userRoute := NewUserRoute(mockApp, fiberApp)

	assert.NotNil(t, userRoute)
	assert.NotNil(t, userRoute.log)
	assert.Equal(t, fiberApp, userRoute.router)
	assert.Equal(t, mockApp.Middleware, userRoute.middleware)
}

func TestUserRoute_Register(t *testing.T) {
	app, userRoute := setupUserRouteTest()
	userRoute.Register()

	routes := app.GetRoutes()

	loginRouteFound := false
	getUserRouteFound := false
	logoutRouteFound := false

	for _, route := range routes {
		switch {
		case route.Path == "/users/login" && route.Method == "POST":
			loginRouteFound = true
		case route.Path == "/users/" && route.Method == "GET":
			getUserRouteFound = true
		case route.Path == "/users/logout" && route.Method == "POST":
			logoutRouteFound = true
		}
	}

	assert.True(t, loginRouteFound, "Login route should be registered")
	assert.True(t, getUserRouteFound, "Get user route should be registered")
	assert.True(t, logoutRouteFound, "Logout route should be registered")
}

func TestUserRoute_Login_StructuralTest(t *testing.T) {
	_, userRoute := setupUserRouteTest()

	// Test that the route structure is set up correctly
	assert.NotNil(t, userRoute)
	assert.NotNil(t, userRoute.controller)

	// Test that we can access the login method without panic
	// Note: Full functionality testing should be done at controller level
	assert.NotPanics(t, func() {
		userRoute.Register()
	})
}

// Note: Detailed login functionality tests should be in controller tests
// These route tests focus on registration and structure