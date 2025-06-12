package websockets

import (
	"server/config"
	"server/internal/logger"
	"server/internal/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test pure business logic without cache dependencies

func TestMessage_BusinessLogic(t *testing.T) {
	// Test message creation and validation
	message := Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeMessage,
		Channel:   "test-channel",
		Action:    "test-action",
		UserID:    uuid.New().String(),
		Data:      map[string]any{"key": "value"},
		Timestamp: time.Now(),
	}

	assert.NotEmpty(t, message.ID)
	assert.Equal(t, MessageTypeMessage, message.Type)
	assert.Equal(t, "test-channel", message.Channel)
	assert.Equal(t, "test-action", message.Action)
	assert.NotEmpty(t, message.UserID)
	assert.Equal(t, "value", message.Data["key"])
	assert.False(t, message.Timestamp.IsZero())
}

func TestClient_StatusLogic(t *testing.T) {
	manager := &Manager{
		log: logger.New("test"),
	}

	client := &Client{
		ID:      "test-client",
		UserID:  uuid.Nil,
		Status:  StatusUnauthenticated,
		Manager: manager,
		send:    make(chan Message, 10),
	}

	// Test initial state
	assert.Equal(t, StatusUnauthenticated, client.Status)
	assert.Equal(t, uuid.Nil, client.UserID)

	// Test status transitions
	client.Status = StatusPending
	assert.Equal(t, StatusPending, client.Status)

	client.Status = StatusAuthenticated
	assert.Equal(t, StatusAuthenticated, client.Status)

	testUserID := uuid.New()
	client.UserID = testUserID
	assert.Equal(t, testUserID, client.UserID)
}

func TestJWTTokenParsing(t *testing.T) {
	testConfig := config.Config{
		SecurityJwtSecret: "test-jwt-secret-very-long-key-for-testing",
		SecurityPepper:    "test-pepper",
		SecuritySalt:      12,
	}

	testUserID := uuid.New()

	// Test valid token generation and parsing
	expiresAt := time.Now().Add(time.Hour)
	token, err := utils.GenerateJWTToken(testUserID.String(), expiresAt, "test-issuer", testConfig)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test token parsing
	claims, err := utils.ParseJWTToken(token, testConfig)
	require.NoError(t, err)
	assert.Equal(t, testUserID, claims.UserID)

	// Test invalid token
	_, err = utils.ParseJWTToken("invalid-token", testConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid number of segments")

	// Test empty token
	_, err = utils.ParseJWTToken("", testConfig)
	assert.Error(t, err)
}

func TestMessageRouting_Logic(t *testing.T) {
	manager := &Manager{
		log: logger.New("test"),
	}

	// Test unauthenticated client message blocking
	unauthClient := &Client{
		ID:      "unauth-client",
		Status:  StatusUnauthenticated,
		Manager: manager,
		send:    make(chan Message, 10),
	}

	testMessage := Message{
		Type:    MessageTypeMessage,
		Channel: "user",
		Action:  "test",
	}

	// This should result in an auth failure message being sent
	unauthClient.routeMessage(testMessage)

	// Check that auth failure was sent
	select {
	case failureMsg := <-unauthClient.send:
		assert.Equal(t, MessageTypeAuthFailure, failureMsg.Type)
		assert.Equal(t, "system", failureMsg.Channel)
		assert.Equal(t, "authentication_required", failureMsg.Action)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected auth failure message")
	}

	// Test authenticated client message allowed
	authClient := &Client{
		ID:      "auth-client",
		Status:  StatusAuthenticated,
		UserID:  uuid.New(),
		Manager: manager,
		send:    make(chan Message, 10),
	}

	// This should NOT result in any messages being sent to the channel
	authClient.routeMessage(testMessage)

	// Verify no auth failure messages
	select {
	case <-authClient.send:
		t.Fatal("Unexpected message sent for authenticated client")
	case <-time.After(50 * time.Millisecond):
		// Expected - no messages should be sent
	}
}

func TestAuthResponse_MessageHandling(t *testing.T) {
	testConfig := config.Config{
		SecurityJwtSecret: "test-jwt-secret-very-long-key-for-testing",
	}

	manager := &Manager{
		log:    logger.New("test"),
		config: testConfig,
	}

	client := &Client{
		ID:      "test-client",
		Status:  StatusUnauthenticated,
		Manager: manager,
		send:    make(chan Message, 10),
	}

	// Test invalid token data
	invalidAuthMsg := Message{
		Type: MessageTypeAuthResponse,
		Data: map[string]any{
			"token": 12345, // Invalid type
		},
	}

	client.handleAuthResponse(invalidAuthMsg)

	// Should send auth failure
	select {
	case failureMsg := <-client.send:
		assert.Equal(t, MessageTypeAuthFailure, failureMsg.Type)
		assert.Contains(t, failureMsg.Data["reason"], "Invalid token format")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected auth failure message")
	}

	// Test missing token
	emptyAuthMsg := Message{
		Type: MessageTypeAuthResponse,
		Data: map[string]any{}, // No token
	}

	client.handleAuthResponse(emptyAuthMsg)

	// Should send auth failure
	select {
	case failureMsg := <-client.send:
		assert.Equal(t, MessageTypeAuthFailure, failureMsg.Type)
		assert.Contains(t, failureMsg.Data["reason"], "Invalid token format")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected auth failure message")
	}
}

func TestSendAuthFailure_Logic(t *testing.T) {
	manager := &Manager{
		log: logger.New("test"),
	}

	client := &Client{
		ID:      "test-client",
		Manager: manager,
		send:    make(chan Message, 10),
	}

	reason := "Test failure reason"
	client.sendAuthFailure(reason)

	// Check that auth failure message was created and sent
	select {
	case message := <-client.send:
		assert.Equal(t, MessageTypeAuthFailure, message.Type)
		assert.Equal(t, "system", message.Channel)
		assert.Equal(t, "authentication_failed", message.Action)
		assert.Equal(t, reason, message.Data["reason"])
		assert.NotEmpty(t, message.ID)
		assert.False(t, message.Timestamp.IsZero())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected auth failure message was not sent")
	}
}


func TestHub_ChannelOperations(t *testing.T) {
	hub := &Hub{
		broadcast:  make(chan Message, 5),
		register:   make(chan *Client, 5),
		unregister: make(chan *Client, 5),
	}

	// Test sending and receiving messages
	testMessage := Message{
		ID:   "test-broadcast",
		Type: MessageTypeBroadcast,
	}

	testClient := &Client{
		ID: "test-client",
	}

	// Send to all channels
	hub.broadcast <- testMessage
	hub.register <- testClient
	hub.unregister <- testClient

	// Verify we can receive from all channels
	receivedMessage := <-hub.broadcast
	assert.Equal(t, testMessage.ID, receivedMessage.ID)
	assert.Equal(t, testMessage.Type, receivedMessage.Type)

	receivedRegister := <-hub.register
	assert.Equal(t, testClient.ID, receivedRegister.ID)

	receivedUnregister := <-hub.unregister
	assert.Equal(t, testClient.ID, receivedUnregister.ID)
}

func TestMessage_ComplexData(t *testing.T) {
	complexData := map[string]any{
		"string":  "test string",
		"number":  42,
		"float":   3.14159,
		"boolean": true,
		"array":   []string{"item1", "item2", "item3"},
		"null":    nil,
		"nested": map[string]any{
			"inner_string": "nested value",
			"inner_number": 100,
		},
	}

	message := Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeMessage,
		Channel:   "complex-data",
		Data:      complexData,
		Timestamp: time.Now(),
	}

	// Verify complex data handling
	assert.Equal(t, "test string", message.Data["string"])
	assert.Equal(t, 42, message.Data["number"])
	assert.Equal(t, 3.14159, message.Data["float"])
	assert.Equal(t, true, message.Data["boolean"])
	assert.Len(t, message.Data["array"], 3)
	assert.Nil(t, message.Data["null"])

	nested := message.Data["nested"].(map[string]any)
	assert.Equal(t, "nested value", nested["inner_string"])
	assert.Equal(t, 100, nested["inner_number"])
}

func TestClient_UUIDValidation(t *testing.T) {
	validUUID := uuid.New()
	
	client := &Client{
		ID:     "client-123",
		UserID: validUUID,
	}

	assert.Equal(t, "client-123", client.ID)
	assert.Equal(t, validUUID, client.UserID)
	assert.NotEqual(t, uuid.Nil, client.UserID)

	// Test with nil UUID
	client.UserID = uuid.Nil
	assert.Equal(t, uuid.Nil, client.UserID)
	assert.True(t, client.UserID == uuid.Nil)
}

func TestWebSocketConstants_BusinessRules(t *testing.T) {
	// Test timing constants make business sense
	assert.True(t, PongTimeout > PingInterval, "Pong timeout should be longer than ping interval")
	assert.True(t, WriteTimeout < PingInterval, "Write timeout should be shorter than ping interval")

	// Test size constants are reasonable
	assert.True(t, MaxMessageSize >= 1024, "Max message size should be at least 1KB")
	assert.True(t, SendChannelSize >= 10, "Send channel should have reasonable buffer")

	// Test message types are non-empty strings
	messageTypes := []string{
		MessageTypePing, MessageTypePong, MessageTypeMessage,
		MessageTypeBroadcast, MessageTypeError, MessageTypeUserJoin,
		MessageTypeUserLeave, MessageTypeAuthRequest, MessageTypeAuthResponse,
		MessageTypeAuthSuccess, MessageTypeAuthFailure,
	}

	for _, msgType := range messageTypes {
		assert.NotEmpty(t, msgType, "Message type should not be empty")
		assert.True(t, len(msgType) > 2, "Message type should be descriptive")
	}

	// Test status constants are sequential
	assert.Equal(t, 0, StatusUnauthenticated)
	assert.Equal(t, 1, StatusPending)
	assert.Equal(t, 2, StatusAuthenticated)
	assert.Equal(t, 3, StatusClosed)
}

func TestMessage_EdgeCases(t *testing.T) {
	// Test message with minimal data
	minimalMessage := Message{
		Type: MessageTypePing,
	}

	assert.Equal(t, MessageTypePing, minimalMessage.Type)
	assert.Empty(t, minimalMessage.ID)
	assert.Empty(t, minimalMessage.Channel)
	assert.True(t, minimalMessage.Timestamp.IsZero())

	// Test message with maximum realistic data
	largeData := make(map[string]any)
	for i := 0; i < 100; i++ {
		largeData[uuid.New().String()] = uuid.New().String()
	}

	largeMessage := Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeMessage,
		Channel:   "large-data-channel",
		Action:    "bulk-operation",
		UserID:    uuid.New().String(),
		Data:      largeData,
		Timestamp: time.Now(),
	}

	assert.NotEmpty(t, largeMessage.ID)
	assert.Len(t, largeMessage.Data, 100)
	assert.False(t, largeMessage.Timestamp.IsZero())
}

func TestClient_SendChannelBehavior(t *testing.T) {
	client := &Client{
		send: make(chan Message, 3), // Small buffer for testing
	}

	// Test sending multiple messages
	messages := []Message{
		{ID: "msg1", Type: MessageTypePing},
		{ID: "msg2", Type: MessageTypePong},
		{ID: "msg3", Type: MessageTypeMessage},
	}

	// Send all messages
	for _, msg := range messages {
		client.send <- msg
	}

	// Receive and verify order
	for i, expectedMsg := range messages {
		select {
		case receivedMsg := <-client.send:
			assert.Equal(t, expectedMsg.ID, receivedMsg.ID, "Message %d order mismatch", i)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Failed to receive message %d", i)
		}
	}

	// Verify channel is now empty
	select {
	case <-client.send:
		t.Fatal("Channel should be empty")
	default:
		// Expected - channel is empty
	}
}