package websockets

import (
	"log/slog"
	"server/config"
	"server/internal/database"
	"server/internal/events"
	"server/internal/logger"
	"server/internal/utils"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

const (
	MessageTypePing         = "ping"
	MessageTypePong         = "pong"
	MessageTypeMessage      = "message"
	MessageTypeBroadcast    = "broadcast"
	MessageTypeError        = "error"
	MessageTypeUserJoin     = "user_join"
	MessageTypeUserLeave    = "user_leave"
	MessageTypeAuthRequest  = "auth_request"
	MessageTypeAuthResponse = "auth_response"
	MessageTypeAuthSuccess  = "auth_success"
	MessageTypeAuthFailure  = "auth_failure"
	PingInterval            = 30 * time.Second
	PongTimeout             = 60 * time.Second
	WriteTimeout            = 10 * time.Second
	MaxMessageSize          = 1024 * 1024 // 1 MB
	SendChannelSize         = 64
	// Channels
	BROADCAST_CHANNEL = "broadcast"
)

type Message struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Channel   string         `json:"channel,omitempty"`
	Action    string         `json:"action,omitempty"`
	UserID    string         `json:"userId,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type Client struct {
	ID         string
	UserID     uuid.UUID
	Connection *websocket.Conn
	Manager    *Manager
	Status     int
	send       chan Message
}

type Manager struct {
	hub      *Hub
	db       database.DB
	config   config.Config
	log      logger.Logger
	eventBus *events.EventBus
}

func New(db database.DB, eventBus *events.EventBus, config config.Config) (*Manager, error) {
	log := logger.New("websockets")

	manager := &Manager{
		hub: &Hub{
			broadcast:  make(chan Message),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			clients:    make(map[string]*Client),
		},
		db:       db,
		config:   config,
		log:      log,
		eventBus: eventBus,
	}

	log.Function("New").Info("Starting websocket hub")
	go manager.hub.run(manager)

	go manager.subscribeToBroadcastEvents()

	return manager, nil
}

func (m *Manager) HandleWebSocket(c *websocket.Conn) {
	log := m.log.Function("HandleWebSocket")
	clientID := uuid.New().String()

	client := &Client{
		ID:         clientID,
		UserID:     uuid.Nil,
		Connection: c,
		Manager:    m,
		Status:     StatusUnauthenticated,
		send:       make(chan Message, SendChannelSize),
	}

	authRequest := Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeAuthRequest,
		Channel:   "system",
		Action:    "authenticate",
		Timestamp: time.Now(),
	}

	if err := c.WriteJSON(authRequest); err != nil {
		log.Er("failed to send auth request", err)
		if err := c.Close(); err != nil {
			log.Er("failed to close connection", err)
		}
		return
	}

	log.Info("Auth request sent to client", "clientID", clientID)
	m.hub.register <- client
	defer func() {
		log.Info("Client disconnected in the defer", "clientID", clientID)
		m.hub.unregister <- client
		if err := c.Close(); err != nil {
			log.Er("failed to close connection", err)
		}
	}()

	go client.readPump()
	client.writePump()
}

func (m *Manager) BroadcastMessage(message Message) {
	log := m.log.Function("BroadcastMessage")
	log.Info("Broadcasting message from ", "messageID", message.ID)

	select {
	case m.hub.broadcast <- message:
		log.Info("Message sent to broadcast channel", "messageID", message.ID)
	default:
		log.Warn("Broadcast channel is full, dropping message", "messageID", message.ID)
	}
}

func (m *Manager) BroadcastUserLogin(userID string, userData map[string]any) {
	log := m.log.Function("BroadcastUserLogin")

	message := Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeUserJoin,
		Channel:   "system",
		Action:    "user_login",
		UserID:    userID,
		Data:      userData,
		Timestamp: time.Now(),
	}

	log.Info("Broadcasting user login", "userID", userID, "messageID", message.ID)

	select {
	case m.hub.broadcast <- message:
		log.Info("User login message sent to broadcast channel", "userID", userID)
	default:
		log.Warn("Broadcast channel is full, dropping user login message", "userID", userID)
	}
}

func (c *Client) readPump() {
	log := c.Manager.log.Function("readPump")
	defer func() {
		c.Manager.hub.unregister <- c
		_ = c.Connection.Close()
	}()

	c.Connection.SetReadLimit(MaxMessageSize)
	if err := c.Connection.SetReadDeadline(time.Now().Add(PongTimeout)); err != nil {
		log.Er("failed to set read deadline", err, "clientID", c.ID)
	}
	c.Connection.SetPongHandler(func(string) error {
		if err := c.Connection.SetReadDeadline(time.Now().Add(PongTimeout)); err != nil {
			log.Er("failed to set read deadline in pong handler", err, "clientID", c.ID)
		}
		return nil
	})

	for {
		var message Message
		err := c.Connection.ReadJSON(&message)
		log.Info("Read message", "clientID", c.ID, "message", message)
		if err != nil {
			log.Er("failed to read message", err)
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Er("Unexpected close error", err, "clientID", c.ID)
			}
			break
		}

		message.ID = uuid.New().String()
		message.Timestamp = time.Now()

		c.routeMessage(message)
	}
}

func (c *Client) routeMessage(message Message) {
	log := c.Manager.log.Function("routeMessage")

	if message.Type == MessageTypeAuthResponse {
		c.handleAuthResponse(message)
		return
	}

	if c.Status == StatusUnauthenticated {
		log.Warn(
			"Blocking message from unauthenticated client",
			"clientID",
			c.ID,
			"messageType",
			message.Type,
		)
		authFailure := Message{
			ID:        uuid.New().String(),
			Type:      MessageTypeAuthFailure,
			Channel:   "system",
			Action:    "authentication_required",
			Data:      map[string]any{"reason": "Authentication required"},
			Timestamp: time.Now(),
		}
		c.send <- authFailure
		return
	}

	switch message.Channel {
	case "system":
		slog.Info("System message", "messageID", message.ID, "clientID", c.ID, "message", message)
	case "user":
		slog.Info("User message", "messageID", message.ID, "clientID", c.ID, "message", message)
	}
}

func (c *Client) handleAuthResponse(message Message) {
	log := c.Manager.log.Function("handleAuthResponse")

	if c.Status != StatusUnauthenticated {
		log.Warn("Auth response from already authenticated client", "clientID", c.ID)
		return
	}

	token, ok := message.Data["token"].(string)
	if !ok || token == "" {
		log.Warn("Invalid token in auth response", "clientID", c.ID)
		c.sendAuthFailure("Invalid token format")
		return
	}

	tokenClaims, err := utils.ParseJWTToken(token, c.Manager.config)
	if err != nil {
		log.Er("failed to parse token", err, "clientID", c.ID)
		c.sendAuthFailure("Invalid token")
		return
	}

	c.UserID = tokenClaims.UserID
	c.Status = StatusAuthenticated

	log.Info("Client authenticated successfully", "clientID", c.ID, "userID", c.UserID)

	c.Manager.promoteClientToAuthenticated(c)

	authSuccess := Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeAuthSuccess,
		Channel:   "system",
		Action:    "authenticated",
		Data:      map[string]any{"userId": c.UserID.String()},
		Timestamp: time.Now(),
	}

	c.send <- authSuccess
}

func (c *Client) sendAuthFailure(reason string) {
	log := c.Manager.log.Function("sendAuthFailure")

	authFailure := Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeAuthFailure,
		Channel:   "system",
		Action:    "authentication_failed",
		Data:      map[string]any{"reason": reason},
		Timestamp: time.Now(),
	}

	c.send <- authFailure

	log.Info("Auth failure sent, closing connection", "clientID", c.ID, "reason", reason)

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = c.Connection.Close()
	}()
}

func (c *Client) writePump() {
	log := c.Manager.log.Function("writePump")

	ticker := time.NewTicker(PingInterval)
	defer func() {
		ticker.Stop()
		_ = c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.Connection.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
				log.Er("failed to set write deadline", err, "clientID", c.ID)
			}
			if !ok {
				log.Info("Channel closed", "clientID", c.ID)
				_ = c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Connection.WriteJSON(message); err != nil {
				log.Er("WebSocket write error", err, "clientID", c.ID, "message", message)
				return
			}

		case <-ticker.C:
			log.Debug("Sending ping", "clientID", c.ID)
			if err := c.Connection.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
				log.Er("failed to set write deadline for ping", err, "clientID", c.ID)
			}
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (m *Manager) subscribeToBroadcastEvents() {
	log := m.log.Function("subscribeToBroadcastEvents")
	log.Info("Starting broadcast events subscription")

	err := m.eventBus.Subscribe(BROADCAST_CHANNEL, func(event events.Event) error {
		log.Info(
			"Received broadcast event",
			"eventID",
			event.ID,
			"eventType",
			event.Type,
			"data",
			event.Data,
		)

		m.sendToAuthenticatedClients(Message{
			ID:        uuid.New().String(),
			Type:      MessageTypeBroadcast,
			Channel:   "system",
			Action:    "broadcast",
			Data:      event.Data,
			Timestamp: time.Now(),
		})
		return nil
	})
	if err != nil {
		log.Er("Failed to subscribe to broadcast events", err)
	}
}

func (m *Manager) sendToAuthenticatedClients(message Message) {
	log := m.log.Function("sendToAuthenticatedClients")
	
	sent := 0
	for _, client := range m.hub.clients {
		if client.Status == StatusAuthenticated {
			select {
			case client.send <- message:
				sent++
			default:
				log.Warn("Client send channel full, dropping message", "clientID", client.ID)
			}
		}
	}
	
	log.Info("Message sent to authenticated clients", "messageID", message.ID, "clientCount", sent)
}
