package app

import (
	"server/config"
	"server/internal/database"
	"server/internal/events"
	"server/internal/logger"
	"server/internal/repositories"
	"server/internal/routes/middleware"
	"server/internal/websockets"

	adminController "server/internal/controllers/admin"
	userController "server/internal/controllers/users"
)

type App struct {
	Database   database.DB
	Middleware middleware.Middleware
	Websocket  *websockets.Manager
	EventBus   *events.EventBus
	Config     config.Config

	// Repositories
	UserRepo    repositories.UserRepository
	SessionRepo repositories.SessionRepository

	// Controllers
	UserController  *userController.UserController
	AdminController *adminController.AdminController
}

func New() (*App, error) {
	log := logger.New("app").Function("New")

	config, err := config.InitConfig()
	if err != nil {
		return &App{}, log.Err("failed to initialize config", err)
	}

	db, err := database.New(config)
	if err != nil {
		return &App{}, log.Err("failed to create database", err)
	}

	eventBus := events.New(db.Cache.Events, config)

	// Initialize repositories
	userRepo := repositories.New(db)
	sessionRepo := repositories.NewSessionRepository(db)

	// Initialize services with repositories
	middleware := middleware.New(db, eventBus, config, userRepo, sessionRepo)
	userController := userController.New(eventBus, userRepo, sessionRepo, config)
	adminController := adminController.New(eventBus, userRepo, config)

	websocket, err := websockets.New(db, eventBus, config)
	if err != nil {
		return &App{}, log.Err("failed to create websocket manager", err)
	}

	app := &App{
		Database:        db,
		Config:          config,
		Middleware:      middleware,
		UserRepo:        userRepo,
		SessionRepo:     sessionRepo,
		UserController:  userController,
		AdminController: adminController,
		Websocket:       websocket,
		EventBus:        eventBus,
	}

	if err := app.validate(); err != nil {
		return &App{}, log.Err("failed to validate app", err)
	}

	return app, nil
}

func (a *App) validate() error {
	log := logger.New("app").Function("validate")
	if a.Database.SQL == nil {
		return log.ErrMsg("database is nil")
	}

	if a.Config == (config.Config{}) {
		return log.ErrMsg("config is nil")
	}

	nilChecks := []any{
		a.Websocket,
		a.EventBus,
		a.UserController,
		a.Middleware,
		a.UserRepo,
		a.SessionRepo,
	}

	for _, check := range nilChecks {
		if check == nil {
			return log.ErrMsg("nil check failed")
		}
	}

	return nil
}

func (a *App) Close() (err error) {
	if a.EventBus != nil {
		if closeErr := a.EventBus.Close(); closeErr != nil {
			err = closeErr
		}
	}

	if dbErr := a.Database.Close(); dbErr != nil {
		err = dbErr
	}

	return err
}
