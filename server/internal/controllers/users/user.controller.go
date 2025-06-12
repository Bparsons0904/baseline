package userController

import (
	"context"
	"server/config"
	"server/internal/events"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/repositories"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	Config      config.Config
	log         logger.Logger
	wsManager   WebSocketManager
	eventBus    *events.EventBus
}

type WebSocketManager interface {
	BroadcastUserLogin(userID string, userData map[string]any)
}

func New(
	eventBus *events.EventBus,
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	config config.Config,
) *UserController {
	return &UserController{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		Config:      config,
		log:         logger.New("userController"),
		wsManager:   nil,
		eventBus:    eventBus,
	}
}

func (c *UserController) SetWebSocketManager(wsManager WebSocketManager) {
	c.wsManager = wsManager
}

func (c *UserController) Login(
	ctx context.Context,
	loginRequest LoginRequest,
) (user User, session Session, err error) {
	log := c.log.Function("Login")
	userPtr, err := c.userRepo.GetByLogin(ctx, loginRequest.Login)
	if err != nil {
		return
	}
	user = *userPtr

	if err = c.comparePassword(loginRequest.Password, user.Password); err != nil {
		log.Warn("Login failed, password comparison failed", "userID", user.ID)
		return
	}

	session.UserID = user.ID
	if err = c.sessionRepo.Create(ctx, &session, c.Config); err != nil {
		return
	}

	// Broadcast user login event to WebSocket clients
	if c.wsManager != nil {
		go c.broadcastUserLogin(user)
	}

	return
}

func (c *UserController) Logout(sessionID string) (err error) {
	ctx := context.Background()
	if err = c.sessionRepo.Delete(ctx, sessionID); err != nil {
		return
	}
	return
}

// TODO: implement
func (c *UserController) Register(user User) (err error) {
	ctx := context.Background()
	if err = c.userRepo.Create(ctx, &user, c.Config); err != nil {
		return
	}
	return
}

func (c *UserController) comparePassword(password, hashedPassword string) error {
	password = password + c.Config.SecurityPepper
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

// broadcastUserLogin sends a login event to WebSocket clients
func (c *UserController) broadcastUserLogin(user User) {
	log := c.log.Function("broadcastUserLogin")

	userData := map[string]any{
		"userId":    user.ID,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"login":     user.Login,
		"isAdmin":   user.IsAdmin,
		"loginTime": time.Now().Unix(),
	}

	log.Info("Broadcasting user login event", "userID", user.ID, "login", user.Login)
	if c.wsManager != nil {
		c.wsManager.BroadcastUserLogin(user.ID, userData)
	}
}
