package repositories

import (
	"context"
	"server/config"
	"server/internal/database"
	"server/internal/logger"
	"server/internal/models"
	"server/internal/utils"
	"time"

	"github.com/google/uuid"
)

const (
	SESSION_EXPIRY     = 7 * 24 * time.Hour // 7 days
	SESSION_REFRESH    = 5 * 24 * time.Hour // 5 days
	SESSION_CACHE_KEY  = "session:"
	SESSION_ISSUER_KEY = "app_api"
)

type sessionRepository struct {
	db  database.DB
	log logger.Logger
}

func NewSessionRepository(db database.DB) SessionRepository {
	return &sessionRepository{
		db:  db,
		log: logger.New("sessionRepository"),
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.Session, config config.Config) error {
	log := r.log.Function("Create")

	if session.ID != "" {
		return log.ErrMsg("Should not already have a Session ID, not a create")
	}

	if session.UserID == "" {
		return log.ErrMsg("Missing User ID")
	}

	id, _ := uuid.NewV7()
	session.ID = id.String()
	session.ExpiresAt = time.Now().Add(SESSION_EXPIRY)
	session.RefreshAt = time.Now().Add(SESSION_REFRESH)
	
	token, err := utils.GenerateJWTToken(session.UserID, session.ExpiresAt, SESSION_ISSUER_KEY, config)
	if err != nil {
		return log.Err("failed to generate JWT token", err, "userID", session.UserID)
	}

	session.Token = token

	if err := database.NewCacheBuilder(r.db.Cache.Session, session.ID).
		WithHashPattern(SESSION_CACHE_KEY).
		WithSruct(session).
		WithTTL(SESSION_EXPIRY).
		Set(); err != nil {
		return log.Err("failed to set session in cache", err, "session", session)
	}

	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID string) (*models.Session, error) {
	log := r.log.Function("GetByID")
	
	var session models.Session
	
	err := database.NewCacheBuilder(r.db.Cache.Session, sessionID).
		WithHashPattern(SESSION_CACHE_KEY).
		Get(&session)
	if err != nil {
		return nil, log.Err("failed to get session from cache", err, "sessionID", sessionID)
	}

	return &session, nil
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	log := r.log.Function("Delete")
	
	err := database.NewCacheBuilder(r.db.Cache.Session, sessionID).
		WithHashPattern(SESSION_CACHE_KEY).
		Delete()
	if err != nil {
		return log.Err("failed to delete session from cache", err, "sessionID", sessionID)
	}

	return nil
}