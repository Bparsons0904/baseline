package middleware

import (
	"context"
	. "server/internal/models"
	"server/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SessionData struct {
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	UserAgent string    `json:"user_agent"`
}

const (
	MOBILE_CLIENT_TYPE = "flutter"
	WEB_CLIENT_TYPE    = "solid"
)

func (m *Middleware) getWebSessionData(c *fiber.Ctx) (Session, error) {
	log := m.log.Function("getWebSessionData")

	sessionID := c.Cookies(SESSION_COOKIE_KEY)
	if sessionID == "" {
		log.Warn("No session cookie found")
		return Session{}, nil
	}

	sessionPtr, err := m.sessionRepo.GetByID(context.Background(), sessionID)
	if err != nil {
		return Session{}, log.Err("failed to get session", err, "sessionID", sessionID)
	}
	session := *sessionPtr

	if session.ExpiresAt.Before(time.Now()) {
		return Session{}, log.ErrMsg("Session expired")
	}

	return session, nil
}

func (m *Middleware) getMobileSessionData(c *fiber.Ctx) (Session, error) {
	log := m.log.Function("getMobileSessionData")
	token := c.Get("Authorization")
	if token == "" {
		return Session{}, log.ErrMsg("No token found")
	}

	claims, err := utils.ParseJWTToken(token, m.Config)
	if err != nil {
		return Session{}, log.Err("failed to parse token", err)
	}

	sessionPtr, err := m.sessionRepo.GetByID(context.Background(), claims.Subject)
	if err != nil {
		return Session{}, log.Err("failed to get session", err)
	}
	session := *sessionPtr

	if session.ExpiresAt.Before(time.Now()) {
		return Session{}, log.ErrMsg("Session expired")
	}

	return session, nil
}

func (m *Middleware) BasicAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := m.log.Function("BasicAuth")

		c.Locals("authenticated", false)
		var session Session
		var err error

		defer func() {
			if err != nil {
				utils.ExpireCookie(c, SESSION_COOKIE_KEY)
				if err := m.sessionRepo.Delete(context.Background(), session.ID); err != nil {
					log.Er("failed to delete session", err, "sessionID", session.ID)
				}
			}
		}()

		clientType := c.Get("X-Client-Type")
		if clientType == "" {
			return log.ErrMsg("No user client type found")
		}

		switch clientType {
		case WEB_CLIENT_TYPE:
			log.Info("Client type is web", "clientType", clientType)
			session, err = m.getWebSessionData(c)
			if err != nil {
				return err
			}
		case MOBILE_CLIENT_TYPE:
			log.Info("Client type is mobile", "clientType", clientType)
			session, err = m.getMobileSessionData(c)
			if err != nil {
				return err
			}
		}

		found := session != (Session{})
		if !found {
			return c.Next()
		}

		if session.RefreshAt.Before(time.Now()) {
			log.Info("Refreshing session", "sessionID", session.ID)
			if err := m.sessionRepo.Create(context.Background(), &session, m.Config); err != nil {
				return log.Err("failed to refresh session", err, "sessionID", session.ID)
			}
			utils.ApplyCookie(c, utils.Cookie{
				Name:    SESSION_COOKIE_KEY,
				Value:   session.ID,
				Expires: session.ExpiresAt,
			})
			utils.ApplyToken(c, session.Token)
		}

		userPtr, err := m.userRepo.GetByID(context.Background(), session.UserID)
		if err != nil {
			return log.Err("failed to get user", err, "userID", session.UserID)
		}
		user := *userPtr

		c.Locals("userID", user.ID)
		c.Locals("user", user)
		c.Locals("session", session)
		c.Locals("authenticated", true)

		return c.Next()
	}
}

func (m *Middleware) AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := m.log.Function("AuthRequired")
		log.Info("AuthRequired")
		if !c.Locals("authenticated").(bool) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}
		return c.Next()
	}
}

func (m *Middleware) AuthNoContent() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := m.log.Function("AuthNoContent")
		log.Info("AuthNoContent")
		if !c.Locals("authenticated").(bool) {
			return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}
		return c.Next()
	}
}
