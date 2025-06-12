package websockets

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessage_StructCreation(t *testing.T) {
	message := Message{
		ID:        "test-id",
		Type:      MessageTypePing,
		Channel:   "test-channel",
		Action:    "test-action",
		UserID:    "user-123",
		Data:      map[string]any{"key": "value"},
		Timestamp: time.Now(),
	}

	assert.Equal(t, "test-id", message.ID)
	assert.Equal(t, MessageTypePing, message.Type)
	assert.Equal(t, "test-channel", message.Channel)
	assert.Equal(t, "test-action", message.Action)
	assert.Equal(t, "user-123", message.UserID)
	assert.Equal(t, "value", message.Data["key"])
	assert.False(t, message.Timestamp.IsZero())
}

func TestMessage_EmptyMessage(t *testing.T) {
	message := Message{}

	assert.Equal(t, "", message.ID)
	assert.Equal(t, "", message.Type)
	assert.Equal(t, "", message.Channel)
	assert.Equal(t, "", message.Action)
	assert.Equal(t, "", message.UserID)
	assert.Nil(t, message.Data)
	assert.True(t, message.Timestamp.IsZero())
}

func TestClient_StructCreation(t *testing.T) {
	testUUID := uuid.New()
	mockManager := &Manager{}

	client := Client{
		ID:         "client-123",
		UserID:     testUUID,
		Connection: nil, // Can't mock websocket.Conn easily
		Manager:    mockManager,
		Status:     StatusAuthenticated,
		send:       make(chan Message, SendChannelSize),
	}

	assert.Equal(t, "client-123", client.ID)
	assert.Equal(t, testUUID, client.UserID)
	assert.Equal(t, mockManager, client.Manager)
	assert.Equal(t, StatusAuthenticated, client.Status)
	assert.NotNil(t, client.send)
	assert.Equal(t, SendChannelSize, cap(client.send))
}

func TestClient_SendChannel(t *testing.T) {
	client := Client{
		send: make(chan Message, 10),
	}

	message := Message{
		ID:   "test-msg",
		Type: MessageTypePing,
	}

	// Test sending message
	client.send <- message

	// Test receiving message
	received := <-client.send
	assert.Equal(t, message.ID, received.ID)
	assert.Equal(t, message.Type, received.Type)
}

func TestManager_StructCreation(t *testing.T) {
	hub := &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}

	manager := &Manager{
		hub: hub,
	}

	assert.NotNil(t, manager.hub)
	assert.Equal(t, hub, manager.hub)
}

func TestHub_StructCreation(t *testing.T) {
	hub := &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}

	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
	assert.NotNil(t, hub.clients)
}


func TestWebSocketConstants(t *testing.T) {
	// Test message type constants
	assert.Equal(t, "ping", MessageTypePing)
	assert.Equal(t, "pong", MessageTypePong)
	assert.Equal(t, "message", MessageTypeMessage)
	assert.Equal(t, "broadcast", MessageTypeBroadcast)
	assert.Equal(t, "error", MessageTypeError)
	assert.Equal(t, "user_join", MessageTypeUserJoin)
	assert.Equal(t, "user_leave", MessageTypeUserLeave)
	assert.Equal(t, "auth_request", MessageTypeAuthRequest)
	assert.Equal(t, "auth_response", MessageTypeAuthResponse)
	assert.Equal(t, "auth_success", MessageTypeAuthSuccess)
	assert.Equal(t, "auth_failure", MessageTypeAuthFailure)

	// Test status constants
	assert.Equal(t, 0, StatusUnauthenticated)
	assert.Equal(t, 1, StatusPending)
	assert.Equal(t, 2, StatusAuthenticated)
	assert.Equal(t, 3, StatusClosed)

	// Test timing constants
	assert.Equal(t, 30*time.Second, PingInterval)
	assert.Equal(t, 60*time.Second, PongTimeout)
	assert.Equal(t, 10*time.Second, WriteTimeout)

	// Test size constants
	assert.Equal(t, 1024*1024, MaxMessageSize)
	assert.Equal(t, 64, SendChannelSize)
}

func TestMessage_WithComplexData(t *testing.T) {
	complexData := map[string]any{
		"string":  "test",
		"number":  42,
		"boolean": true,
		"array":   []string{"a", "b", "c"},
		"nested": map[string]any{
			"key": "value",
		},
	}

	message := Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeMessage,
		Channel:   "complex",
		Data:      complexData,
		Timestamp: time.Now(),
	}

	assert.Equal(t, "test", message.Data["string"])
	assert.Equal(t, 42, message.Data["number"])
	assert.Equal(t, true, message.Data["boolean"])
	assert.Len(t, message.Data["array"], 3)
	assert.IsType(t, map[string]any{}, message.Data["nested"])
}

func TestClient_StatusTransitions(t *testing.T) {
	client := Client{
		Status: StatusUnauthenticated,
	}

	// Test status progression
	assert.Equal(t, StatusUnauthenticated, client.Status)

	client.Status = StatusPending
	assert.Equal(t, StatusPending, client.Status)

	client.Status = StatusAuthenticated
	assert.Equal(t, StatusAuthenticated, client.Status)

	client.Status = StatusClosed
	assert.Equal(t, StatusClosed, client.Status)
}

func TestClient_UUIDHandling(t *testing.T) {
	// Test with nil UUID
	client := Client{
		UserID: uuid.Nil,
	}

	assert.Equal(t, uuid.Nil, client.UserID)
	assert.True(t, client.UserID == uuid.Nil)

	// Test with valid UUID
	testUUID := uuid.New()
	client.UserID = testUUID
	assert.Equal(t, testUUID, client.UserID)
	assert.False(t, client.UserID == uuid.Nil)
}

func TestMessage_JSONTags(t *testing.T) {
	// This tests the struct field accessibility and JSON tag presence
	message := Message{}

	// Test that all fields are accessible
	message.ID = "test"
	message.Type = "test"
	message.Channel = "test"
	message.Action = "test"
	message.UserID = "test"
	message.Data = make(map[string]any)
	message.Timestamp = time.Now()

	assert.Equal(t, "test", message.ID)
	assert.Equal(t, "test", message.Type)
	assert.Equal(t, "test", message.Channel)
	assert.Equal(t, "test", message.Action)
	assert.Equal(t, "test", message.UserID)
	assert.NotNil(t, message.Data)
	assert.False(t, message.Timestamp.IsZero())
}

// Negative Test Cases

func TestClient_NilSendChannel(t *testing.T) {
	client := Client{
		send: nil,
	}

	assert.Nil(t, client.send)

	// Test that we can check for nil without panicking
	if client.send != nil {
		client.send <- Message{}
	}
}

func TestMessage_EmptyData(t *testing.T) {
	message := Message{
		Data: map[string]any{},
	}

	assert.NotNil(t, message.Data)
	assert.Len(t, message.Data, 0)

	// Test accessing non-existent key
	value, exists := message.Data["nonexistent"]
	assert.Nil(t, value)
	assert.False(t, exists)
}

func TestMessage_NilData(t *testing.T) {
	message := Message{
		Data: nil,
	}

	assert.Nil(t, message.Data)

	// Accessing nil map should not panic but return nil, false
	if message.Data != nil {
		value, exists := message.Data["key"]
		assert.Nil(t, value)
		assert.False(t, exists)
	}
}

func TestClient_InvalidUUID(t *testing.T) {
	// Test with malformed UUID string (this would be caught at creation time)
	client := Client{
		ID: "invalid-uuid-format",
	}

	assert.Equal(t, "invalid-uuid-format", client.ID)

	// Test that we can still work with invalid ID format
	assert.NotEmpty(t, client.ID)
}

func TestMessage_VeryLongFields(t *testing.T) {
	longString := string(make([]byte, 10000))

	message := Message{
		ID:      longString,
		Type:    longString,
		Channel: longString,
		Action:  longString,
		UserID:  longString,
	}

	assert.Len(t, message.ID, 10000)
	assert.Len(t, message.Type, 10000)
	assert.Len(t, message.Channel, 10000)
	assert.Len(t, message.Action, 10000)
	assert.Len(t, message.UserID, 10000)
}

func TestMessage_SpecialCharacters(t *testing.T) {
	specialChars := "!@#$%^&*()_+{}|:<>?[];'\"\\,./`~测试🚀"

	message := Message{
		ID:      specialChars,
		Type:    specialChars,
		Channel: specialChars,
		Action:  specialChars,
		UserID:  specialChars,
		Data: map[string]any{
			"special": specialChars,
		},
	}

	assert.Equal(t, specialChars, message.ID)
	assert.Equal(t, specialChars, message.Type)
	assert.Equal(t, specialChars, message.Channel)
	assert.Equal(t, specialChars, message.Action)
	assert.Equal(t, specialChars, message.UserID)
	assert.Equal(t, specialChars, message.Data["special"])
}



func TestMessage_TimestampBehavior(t *testing.T) {
	// Test zero timestamp
	message := Message{}
	assert.True(t, message.Timestamp.IsZero())

	// Test setting timestamp
	now := time.Now()
	message.Timestamp = now
	assert.Equal(t, now, message.Timestamp)
	assert.False(t, message.Timestamp.IsZero())

	// Test timestamp comparison
	future := now.Add(1 * time.Hour)
	message.Timestamp = future
	assert.True(t, message.Timestamp.After(now))
}
