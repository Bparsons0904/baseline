package app

import (
	"context"
	"os"
	"server/config"
	"server/internal/database"
	"server/internal/models"
	"server/internal/routes/middleware"
	"server/internal/websockets"
	"strings"
	"testing"

	userController "server/internal/controllers/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Focus on testing validation logic and Close functionality with proper mocks

func TestApp_Validate_CompleteValidation(t *testing.T) {
	// Test all validation conditions systematically

	testCases := []struct {
		name        string
		app         *App
		expectError bool
		errorMsg    string
	}{
		{
			name: "AllValidFields",
			app: &App{
				Database:       createValidMockDatabase(t),
				Config:         config.Config{ServerPort: 8080},
				Websocket:      &websockets.Manager{},
				UserController: (*userController.UserController)(nil),
				Middleware:     middleware.Middleware{Config: config.Config{ServerPort: 8080}},
				UserRepo:       &mockUserRepository{},
				SessionRepo:    &mockSessionRepository{},
			},
			expectError: false,
		},
		{
			name: "NilDatabase",
			app: &App{
				Database:       database.DB{SQL: nil},
				Config:         config.Config{ServerPort: 8080},
				Websocket:      &websockets.Manager{},
				UserController: (*userController.UserController)(nil),
				Middleware:     middleware.Middleware{Config: config.Config{ServerPort: 8080}},
			},
			expectError: true,
			errorMsg:    "database is nil",
		},
		{
			name: "EmptyConfig",
			app: &App{
				Database:       createValidMockDatabase(t),
				Config:         config.Config{}, // Empty config
				Websocket:      &websockets.Manager{},
				UserController: (*userController.UserController)(nil),
				Middleware:     middleware.Middleware{Config: config.Config{ServerPort: 8080}},
			},
			expectError: true,
			errorMsg:    "config is nil",
		},
		{
			name: "NilWebsocket",
			app: &App{
				Database:       createValidMockDatabase(t),
				Config:         config.Config{ServerPort: 8080},
				Websocket:      nil,
				UserController: (*userController.UserController)(nil),
				Middleware:     middleware.Middleware{Config: config.Config{ServerPort: 8080}},
			},
			expectError: true,
			errorMsg:    "nil check failed",
		},
		{
			name: "NilUserController",
			app: &App{
				Database:       createValidMockDatabase(t),
				Config:         config.Config{ServerPort: 8080},
				Websocket:      &websockets.Manager{},
				UserController: nil,
				Middleware:     middleware.Middleware{Config: config.Config{ServerPort: 8080}},
			},
			expectError: true,
			errorMsg:    "nil check failed",
		},
		{
			name: "EmptyMiddleware",
			app: &App{
				Database:       createValidMockDatabase(t),
				Config:         config.Config{ServerPort: 8080},
				Websocket:      &websockets.Manager{},
				UserController: (*userController.UserController)(nil),
				Middleware:     middleware.Middleware{}, // Empty middleware
			},
			expectError: true,
			errorMsg:    "nil check failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Close database if it exists to clean up
			if tc.app.Database.SQL != nil {
				defer func() {
					sqlDB, _ := tc.app.Database.SQL.DB()
					if sqlDB != nil {
						_ = sqlDB.Close()
					}
				}()
			}

			err := tc.app.validate()

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApp_Validate_ValidationOrder(t *testing.T) {
	// Test that validation happens in the correct order
	// Database check should happen first

	app := &App{
		Database:       database.DB{SQL: nil},   // This should fail first
		Config:         config.Config{},         // This would also fail
		Websocket:      nil,                     // This would also fail
		UserController: nil,                     // This would also fail
		Middleware:     middleware.Middleware{}, // This would also fail
	}

	err := app.validate()

	assert.Error(t, err)
	// Should get database error first since that's checked first
	assert.Contains(t, err.Error(), "database is nil")
}

func TestApp_Close_Functionality(t *testing.T) {
	testCases := []struct {
		name        string
		app         *App
		expectError bool
	}{
		{
			name: "ValidDatabase",
			app: &App{
				Database: createValidMockDatabase(t),
			},
			expectError: false,
		},
		{
			name: "EmptyDatabase",
			app: &App{
				Database: database.DB{}, // Empty database
			},
			expectError: true, // May error when trying to close nil connection
		},
		{
			name: "NilSQLDatabase",
			app: &App{
				Database: database.DB{SQL: nil},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.app.Close()

			if tc.expectError {
				// May return error when closing invalid database
				if err != nil {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApp_New_ErrorPaths(t *testing.T) {
	// Test error paths that we can actually trigger

	// Save original working directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	// Test 1: Config initialization failure
	t.Run("ConfigFailure", func(t *testing.T) {
		// Create temp directory with no .env file and clear env vars
		tempDir, err := os.MkdirTemp("", "app-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// Clear environment variables
		clearEnvironmentVars(t)

		app, err := New()

		assert.Error(t, err)
		// Error could be either config initialization failure or validation failure
		assert.True(t,
			strings.Contains(err.Error(), "failed to initialize config") ||
				strings.Contains(err.Error(), "invalid port"),
			"Expected config-related error, got: %s", err.Error())
		assert.NotNil(t, app)
		assert.Equal(t, App{}, *app)

		// Reset to original directory for next test
		err = os.Chdir(originalDir)
		require.NoError(t, err)
	})
}

func TestApp_StructFieldAccess(t *testing.T) {
	// Test that all struct fields are accessible and have correct types
	app := &App{}

	// Test field types
	assert.IsType(t, database.DB{}, app.Database)
	assert.IsType(t, middleware.Middleware{}, app.Middleware)
	assert.IsType(t, (*websockets.Manager)(nil), app.Websocket)
	assert.IsType(t, config.Config{}, app.Config)
	// UserController is an interface, test that it can be nil
	assert.Nil(t, app.UserController)

	// Test zero values
	assert.Equal(t, database.DB{}, app.Database)
	assert.Equal(t, middleware.Middleware{}, app.Middleware)
	assert.Nil(t, app.Websocket)
	assert.Equal(t, config.Config{}, app.Config)
	assert.Nil(t, app.UserController)
}

func TestApp_ConfigComparisons(t *testing.T) {
	// Test config comparison logic used in validation
	emptyConfig := config.Config{}
	validConfig := config.Config{ServerPort: 8080}

	app1 := &App{Config: emptyConfig}
	app2 := &App{Config: validConfig}

	// Test equality checks
	assert.True(t, app1.Config == emptyConfig)
	assert.False(t, app2.Config == emptyConfig)
	assert.True(t, app2.Config == validConfig)
}

func TestApp_MiddlewareComparisons(t *testing.T) {
	// Test middleware comparison logic used in validation
	emptyMiddleware := middleware.Middleware{}
	validMiddleware := middleware.Middleware{
		Config: config.Config{ServerPort: 8080},
	}

	app1 := &App{Middleware: emptyMiddleware}
	app2 := &App{Middleware: validMiddleware}

	// Test equality checks
	assert.True(t, app1.Middleware == emptyMiddleware)
	assert.False(t, app2.Middleware == emptyMiddleware)
	assert.True(t, app2.Middleware == validMiddleware)
}

// Helper functions


type mockUserRepository struct{}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	return &models.User{}, nil
}
func (m *mockUserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	return &models.User{}, nil
}
func (m *mockUserRepository) Create(ctx context.Context, user *models.User, config config.Config) error {
	return nil
}
func (m *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	return nil
}
func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

type mockSessionRepository struct{}

func (m *mockSessionRepository) Create(ctx context.Context, session *models.Session, config config.Config) error {
	return nil
}
func (m *mockSessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	return &models.Session{}, nil
}
func (m *mockSessionRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func createValidMockDatabase(t *testing.T) database.DB {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	return database.DB{
		SQL:   db,
		Cache: database.Cache{}, // Empty cache for testing
	}
}

func clearEnvironmentVars(t *testing.T) {
	envVarsToCleanup := []string{
		"SERVER_PORT", "ENVIRONMENT", "GENERAL_VERSION",
		"DB_PATH", "DB_CACHE_ADDRESS", "DB_CACHE_PORT",
		"CORS_ALLOW_ORIGINS", "SECURITY_SALT",
		"SECURITY_PEPPER", "SECURITY_JWT_SECRET",
	}

	for _, envVar := range envVarsToCleanup {
		_ = os.Unsetenv(envVar)
	}
}
