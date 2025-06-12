package websockets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Basic WebSocket package tests - focusing on constants and types rather than complex integration

func TestConstants(t *testing.T) {
	// Test that all message type constants are defined
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
	assert.Equal(t, "broadcast", BROADCAST_CHANNEL)
}

func TestMessageStruct(t *testing.T) {
	// Test Message struct creation and field access
	message := Message{
		ID:      "test-id",
		Type:    MessageTypePing,
		Channel: "test-channel",
		Action:  "test-action",
		UserID:  "test-user",
		Data:    map[string]any{"key": "value"},
	}

	assert.Equal(t, "test-id", message.ID)
	assert.Equal(t, MessageTypePing, message.Type)
	assert.Equal(t, "test-channel", message.Channel)
	assert.Equal(t, "test-action", message.Action)
	assert.Equal(t, "test-user", message.UserID)
	assert.Equal(t, "value", message.Data["key"])
}

func TestClientStruct(t *testing.T) {
	// Test Client struct creation and field access
	client := Client{
		ID:     "test-client",
		Status: StatusUnauthenticated,
	}

	assert.Equal(t, "test-client", client.ID)
	assert.Equal(t, StatusUnauthenticated, client.Status)
}

// Note: Full integration tests with actual WebSocket connections and EventBus
// should be implemented when the architecture is more stable