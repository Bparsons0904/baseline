package utils

import (
	"server/config"
	"server/internal/logger"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID uuid.UUID `json:"userId"`
	jwt.RegisteredClaims
}

func ApplyToken(c *fiber.Ctx, token string) {
	c.Set("X-Auth-Token", token)
}

func GenerateJWTToken(
	userID string,
	// subject string,
	expiresAt time.Time,
	issuer string,
	config config.Config,
) (string, error) {
	log := logger.New("utils").Function("GenerateJWTToken")

	secretKey := config.SecurityJwtSecret
	if secretKey == "" {
		return "", log.ErrMsg("JWT secret key not found in config")
	}

	ID, err := uuid.Parse(userID)
	if err != nil {
		return "", log.Err("failed to parse user id", err)
	}

	claims := TokenClaims{
		ID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			// Subject:   subject,
			ID: uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", log.Err("failed to sign token", err)
	}

	return tokenString, nil
}

func ParseJWTToken(tokenString string, config config.Config) (*TokenClaims, error) {
	log := logger.New("utils").Function("ParseJWTToken")
	secretKey := config.SecurityJwtSecret

	if secretKey == "" {
		return nil, log.ErrMsg("JWT secret key not found in config")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, log.Error("unexpected signing method", "method", token.Header["alg"])
			}
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		return nil, log.Err("failed to parse token", err)
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, log.ErrMsg("invalid token claims")
}
