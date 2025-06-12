package seed

import (
	"server/config"
	"server/internal/logger"
	. "server/internal/models"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB, config config.Config, log logger.Logger) error {
	log = log.Function("seed")
	users := []User{
		{
			FirstName: "John",
			LastName:  "Doe",
			Login:     "johndoe",
			Password:  "password",
			IsAdmin:   true,
		},
		{
			FirstName: "Jane",
			LastName:  "Doe",
			Login:     "janedoe",
			Password:  "password",
			IsAdmin:   true,
		},
		{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Login:     "ada",
			Password:  "password",
			IsAdmin:   false,
		},
		{
			FirstName: "Grace",
			LastName:  "Hopper",
			Login:     "grace",
			Password:  "password",
			IsAdmin:   false,
		},
	}

	for _, user := range users {
		var existingUser User
		if err := db.First(&existingUser, "login = ?", user.Login).Error; err == nil {
			log.Info("User already exists", "user", user)
			continue
		}
		log.Info("Seeding user", "user", user)
		if err := db.Create(&user).Error; err != nil {
			log.Er("failed to create user", err, "user", user)
		}
	}

	return nil
}
