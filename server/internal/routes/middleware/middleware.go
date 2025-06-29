package middleware

import (
	"server/config"
	"server/internal/database"
	"server/internal/events"
	"server/internal/logger"
	"server/internal/repositories"
)

type Middleware struct {
	DB          database.DB
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	Config      config.Config
	log         logger.Logger
	eventBus    *events.EventBus
}

func New(
	db database.DB,
	eventBus *events.EventBus,
	config config.Config,
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
) Middleware {
	log := logger.New("middleware")

	return Middleware{
		DB:          db,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		Config:      config,
		log:         log,
		eventBus:    eventBus,
	}
}
