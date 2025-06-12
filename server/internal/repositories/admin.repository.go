package repositories

import (
	"context"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
)

type adminRepository struct {
	db  database.DB
	log logger.Logger
}

func Newte(db database.DB) AdminRepository {
	return &adminRepository{
		db:  db,
		log: logger.New("adminRepository"),
	}
}

func (r *adminRepository) GetByID(ctx context.Context, message string) (*User, error) {
	log := r.log.Function("SendBroadcast")

	log.Info("Broadcasting user login event", "userID", message, "login", message)
	return &User{}, nil
}
