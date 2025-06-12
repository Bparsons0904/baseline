package utils

import (
	"server/config"
	"server/internal/logger"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	log := logger.New("utils").File("auth").Function("hashPassword")
	config := config.GetConfig()
	salt := config.SecuritySalt
	pepper := config.SecurityPepper
	if salt <= 0 || pepper == "" {
		return "", log.Error("salt or pepper is empty", "salt", salt, "pepper", pepper)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password+pepper), salt)
	if err != nil {
		return "", log.Err("failed to hash password", err)
	}

	return string(bytes), nil
}
