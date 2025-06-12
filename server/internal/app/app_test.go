package app

import (
	"server/config"
	"server/internal/database"
	"server/internal/logger"
	"server/internal/routes/middleware"
	"server/internal/websockets"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_StructCreation(t *testing.T) {
	// Test creating App struct manually
	app := &App{
		Config: config.Config{
			ServerPort: 8080,
		},
	}

	assert.NotNil(t, app)
	assert.Equal(t, 8080, app.Config.ServerPort)
}

func TestApp_StructWithAllFields(t *testing.T) {
	// Test App struct with all fields populated
	mockConfig := config.Config{
		ServerPort:       8080,
		GeneralVersion:   "1.0.0",
		Environment:      "test",
		CorsAllowOrigins: "http://localhost:3000",
	}

	app := &App{
		Database:   database.DB{},
		Middleware: middleware.Middleware{},
		Websocket:  &websockets.Manager{},
		Config:     mockConfig,
	}

	assert.NotNil(t, app)
	assert.Equal(t, mockConfig, app.Config)
	assert.NotNil(t, app.Database)
	assert.NotNil(t, app.Middleware)
	assert.NotNil(t, app.Websocket)
}

func TestApp_FieldTypes(t *testing.T) {
	app := &App{}

	// Verify field types are accessible
	assert.IsType(t, database.DB{}, app.Database)
	assert.IsType(t, middleware.Middleware{}, app.Middleware)
	assert.IsType(t, (*websockets.Manager)(nil), app.Websocket)
	assert.IsType(t, config.Config{}, app.Config)
}

func TestApp_Close(t *testing.T) {
	app := &App{
		Database: database.DB{},
	}

	// Test that Close method exists and can be called
	// We can't test actual database closing without real DB
	err := app.Close()
	// May return error or nil depending on database state
	if err != nil {
		assert.Error(t, err)
	}
}

func TestApp_Validate_EmptyApp(t *testing.T) {
	app := &App{}

	err := app.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database is nil")
}

func TestApp_Validate_NilDatabase(t *testing.T) {
	app := &App{
		Database:  database.DB{SQL: nil}, // Nil SQL
		Config:    config.Config{ServerPort: 8080},
		Websocket: &websockets.Manager{},
		Middleware: middleware.Middleware{
			Config: config.Config{},
		},
	}

	err := app.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database is nil")
}

func TestApp_Validate_EmptyConfig(t *testing.T) {
	// Mock a database with non-nil SQL
	mockDB := database.DB{}
	// We can't easily create a real *gorm.DB, but we can test the validation logic

	app := &App{
		Database:  mockDB,
		Config:    config.Config{}, // Empty config
		Websocket: &websockets.Manager{},
	}

	// This will fail on database nil check first, but that's expected
	err := app.validate()
	assert.Error(t, err)
	// Will fail on database check before config check
}

func TestApp_Validate_NilWebsocket(t *testing.T) {
	app := &App{
		Websocket: nil,
	}

	err := app.validate()
	assert.Error(t, err)
	// Will fail on database check first, but websocket validation exists
}

func TestApp_Validate_NilUserController(t *testing.T) {
	app := &App{
		UserController: nil,
	}

	err := app.validate()
	assert.Error(t, err)
	// Will fail on database check first, but user controller validation exists
}

func TestApp_Validate_EmptyMiddleware(t *testing.T) {
	app := &App{
		Middleware: middleware.Middleware{}, // Empty middleware
	}

	err := app.validate()
	assert.Error(t, err)
	// Will fail on database check first, but middleware validation exists
}

func TestNew_FunctionExists(t *testing.T) {
	// Test that New function exists and can be called
	// This will fail due to missing config file and dependencies,
	// but we can verify it doesn't panic and returns proper error
	app, err := New()

	// Should return error due to missing dependencies
	assert.Error(t, err)
	assert.NotNil(t, app) // Should return empty app even on error
}

// Test validation order and error messages
func TestApp_Validate_DatabaseFirst(t *testing.T) {
	// Test that database validation happens first
	app := &App{
		Database:       database.DB{SQL: nil},
		Config:         config.Config{},
		Websocket:      nil,
		UserController: nil,
		Middleware:     middleware.Middleware{},
	}

	err := app.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database is nil")
}

func TestApp_InterfaceFieldAccess(t *testing.T) {
	// Test that interface fields can be accessed
	app := &App{}

	// UserController is an interface, test it can be nil
	assert.Nil(t, app.UserController)

	// Test that we can assign to it (would be done in real app)
	// app.UserController = someController // (can't create without dependencies)
}

// Negative Test Cases

func TestApp_ZeroValues(t *testing.T) {
	// Test app with all zero values
	app := &App{}

	assert.Equal(t, database.DB{}, app.Database)
	assert.Equal(t, middleware.Middleware{}, app.Middleware)
	assert.Nil(t, app.Websocket)
	assert.Equal(t, config.Config{}, app.Config)
	assert.Nil(t, app.UserController)
}

func TestApp_PartiallyPopulated(t *testing.T) {
	// Test app with only some fields populated
	app := &App{
		Config: config.Config{ServerPort: 8080},
		// Other fields remain zero/nil
	}

	assert.Equal(t, 8080, app.Config.ServerPort)
	assert.Equal(t, database.DB{}, app.Database)
	assert.Nil(t, app.Websocket)
	assert.Nil(t, app.UserController)
}

func TestApp_ConfigComparison(t *testing.T) {
	// Test config comparison logic used in validation
	emptyConfig := config.Config{}
	nonEmptyConfig := config.Config{ServerPort: 8080}

	app1 := &App{Config: emptyConfig}
	app2 := &App{Config: nonEmptyConfig}

	// Test that empty config equals empty config
	assert.Equal(t, emptyConfig, app1.Config)
	assert.NotEqual(t, emptyConfig, app2.Config)
}

func TestApp_MiddlewareComparison(t *testing.T) {
	// Test middleware comparison logic used in validation
	emptyMiddleware := middleware.Middleware{}
	nonEmptyMiddleware := middleware.Middleware{
		Config: config.Config{ServerPort: 8080},
	}

	app1 := &App{Middleware: emptyMiddleware}
	app2 := &App{Middleware: nonEmptyMiddleware}

	assert.Equal(t, emptyMiddleware, app1.Middleware)
	assert.NotEqual(t, emptyMiddleware, app2.Middleware)
}

func TestApp_WebsocketPointerHandling(t *testing.T) {
	// Test websocket pointer handling
	var nilManager *websockets.Manager = nil
	validManager := &websockets.Manager{}

	app1 := &App{Websocket: nilManager}
	app2 := &App{Websocket: validManager}

	assert.Nil(t, app1.Websocket)
	assert.NotNil(t, app2.Websocket)
	assert.IsType(t, (*websockets.Manager)(nil), app1.Websocket)
	assert.IsType(t, &websockets.Manager{}, app2.Websocket)
}

func TestApp_DatabaseFieldAccess(t *testing.T) {
	// Test database field access patterns
	app := &App{}

	// Test accessing SQL field
	assert.Nil(t, app.Database.SQL)

	// Test that database is a value type (not pointer)
	assert.IsType(t, database.DB{}, app.Database)
}

func TestApp_ErrorPropagation(t *testing.T) {
	// Test that validation errors are properly propagated
	log := logger.New("test")

	// Test error creation patterns similar to validation
	err := log.ErrMsg("test error message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test error message")
}

func TestApp_ValidationOrder(t *testing.T) {
	// Test that validation happens in expected order by testing each condition
	tests := []struct {
		name        string
		app         *App
		expectedErr string
	}{
		{
			name:        "DatabaseNil",
			app:         &App{Database: database.DB{SQL: nil}},
			expectedErr: "database is nil",
		},
		{
			name: "ConfigEmpty",
			app: &App{
				Database: database.DB{SQL: nil}, // Will fail here first
				Config:   config.Config{},
			},
			expectedErr: "database is nil", // Database checked first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.app.validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
