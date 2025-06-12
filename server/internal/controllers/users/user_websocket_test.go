package userController

import (
	"server/config"
	"server/internal/events"
	"server/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWebSocketManager for testing
type MockWebSocketManager struct {
	mock.Mock
}

func (m *MockWebSocketManager) BroadcastUserLogin(userID string, userData map[string]any) {
	m.Called(userID, userData)
}

func (m *MockWebSocketManager) AssertExpected(t *testing.T) {
	m.AssertExpectations(t)
}

func (m *MockWebSocketManager) AssertCalled(
	t *testing.T,
	methodName string,
	arguments ...interface{},
) {
	m.Mock.AssertCalled(t, methodName, arguments...)
}

func TestUserController_SetWebSocketManager(t *testing.T) {
	config := config.Config{}
	eventBus := &events.EventBus{}
	controller := New(eventBus, nil, nil, config)

	mockWS := &MockWebSocketManager{}
	controller.SetWebSocketManager(mockWS)

	assert.Equal(t, mockWS, controller.wsManager, "WebSocket manager should be set correctly")
}

func TestUserController_BroadcastUserLogin(t *testing.T) {
	config := config.Config{}
	eventBus := &events.EventBus{}
	controller := New(eventBus, nil, nil, config)

	mockWS := &MockWebSocketManager{}
	controller.SetWebSocketManager(mockWS)

	// Create test user
	testUser := models.User{
		BaseModel: models.BaseModel{
			ID: "test-user-123",
		},
		FirstName: "John",
		LastName:  "Doe",
		Login:     "johndoe",
		IsAdmin:   false,
	}

	// Set up expected call with matcher for loginTime
	mockWS.On("BroadcastUserLogin", testUser.ID, mock.MatchedBy(func(userData map[string]any) bool {
		// Check all fields except loginTime
		return userData["userId"] == testUser.ID &&
			userData["firstName"] == testUser.FirstName &&
			userData["lastName"] == testUser.LastName &&
			userData["login"] == testUser.Login &&
			userData["isAdmin"] == testUser.IsAdmin &&
			userData["loginTime"] != nil
	})).Return()

	// Call the method
	controller.broadcastUserLogin(testUser)

	// Verify the call was made
	mockWS.AssertExpected(t)
}

func TestUserController_BroadcastUserLogin_NilWSManager(t *testing.T) {
	config := config.Config{}
	eventBus := &events.EventBus{}
	controller := New(eventBus, nil, nil, config)

	// Don't set WebSocket manager (leave as nil)
	assert.Nil(t, controller.wsManager, "WebSocket manager should be nil initially")

	// Create test user
	testUser := models.User{
		BaseModel: models.BaseModel{
			ID: "test-user-123",
		},
		FirstName: "John",
		LastName:  "Doe",
		Login:     "johndoe",
		IsAdmin:   false,
	}

	// This should not panic even with nil wsManager
	assert.NotPanics(t, func() {
		controller.broadcastUserLogin(testUser)
	}, "Should not panic with nil WebSocket manager")
}

func TestUserController_BroadcastUserLogin_UserData(t *testing.T) {
	config := config.Config{}
	eventBus := &events.EventBus{}
	controller := New(eventBus, nil, nil, config)

	mockWS := &MockWebSocketManager{}
	controller.SetWebSocketManager(mockWS)

	// Create test user with various field types
	testUser := models.User{
		BaseModel: models.BaseModel{
			ID: "user-456",
		},
		FirstName: "Jane",
		LastName:  "Smith",
		Login:     "janesmith",
		IsAdmin:   true, // Test admin user
	}

	// Capture the actual call
	var capturedUserID string
	var capturedUserData map[string]any

	mockWS.On("BroadcastUserLogin", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).
		Run(func(args mock.Arguments) {
			capturedUserID = args.Get(0).(string)
			capturedUserData = args.Get(1).(map[string]any)
		}).
		Return()

	// Call the method
	controller.broadcastUserLogin(testUser)

	// Verify the captured data
	assert.Equal(t, testUser.ID, capturedUserID)
	assert.Equal(t, testUser.ID, capturedUserData["userId"])
	assert.Equal(t, testUser.FirstName, capturedUserData["firstName"])
	assert.Equal(t, testUser.LastName, capturedUserData["lastName"])
	assert.Equal(t, testUser.Login, capturedUserData["login"])
	assert.Equal(t, testUser.IsAdmin, capturedUserData["isAdmin"])

	// Verify loginTime is a recent timestamp
	loginTime, ok := capturedUserData["loginTime"].(int64)
	assert.True(t, ok, "loginTime should be int64")
	assert.True(t, loginTime > 0, "loginTime should be positive")
	assert.True(t, time.Now().Unix()-loginTime < 5, "loginTime should be within last 5 seconds")

	mockWS.AssertExpected(t)
}

func TestUserController_BroadcastUserLogin_EmptyUserFields(t *testing.T) {
	config := config.Config{}
	eventBus := &events.EventBus{}
	controller := New(eventBus, nil, nil, config)

	mockWS := &MockWebSocketManager{}
	controller.SetWebSocketManager(mockWS)

	// Create test user with empty fields
	testUser := models.User{
		BaseModel: models.BaseModel{
			ID: "",
		},
		FirstName: "",
		LastName:  "",
		Login:     "",
		IsAdmin:   false,
	}

	// Capture the actual call
	var capturedUserData map[string]any

	mockWS.On("BroadcastUserLogin", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).
		Run(func(args mock.Arguments) {
			capturedUserData = args.Get(1).(map[string]any)
		}).
		Return()

	// Call the method
	controller.broadcastUserLogin(testUser)

	// Verify empty fields are handled correctly
	assert.Equal(t, "", capturedUserData["userId"])
	assert.Equal(t, "", capturedUserData["firstName"])
	assert.Equal(t, "", capturedUserData["lastName"])
	assert.Equal(t, "", capturedUserData["login"])
	assert.Equal(t, false, capturedUserData["isAdmin"])

	mockWS.AssertExpected(t)
}

func TestUserController_BroadcastUserLogin_SpecialCharacters(t *testing.T) {
	config := config.Config{}
	eventBus := &events.EventBus{}
	controller := New(eventBus, nil, nil, config)

	mockWS := &MockWebSocketManager{}
	controller.SetWebSocketManager(mockWS)

	// Create test user with special characters
	testUser := models.User{
		BaseModel: models.BaseModel{
			ID: "user-特殊字符-🚀",
		},
		FirstName: "José",
		LastName:  "O'Connor",
		Login:     "josé.o'connor@test.com",
		IsAdmin:   false,
	}

	// Capture the actual call
	var capturedUserData map[string]any

	mockWS.On("BroadcastUserLogin", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).
		Run(func(args mock.Arguments) {
			capturedUserData = args.Get(1).(map[string]any)
		}).
		Return()

	// Call the method
	controller.broadcastUserLogin(testUser)

	// Verify special characters are preserved
	assert.Equal(t, testUser.ID, capturedUserData["userId"])
	assert.Equal(t, testUser.FirstName, capturedUserData["firstName"])
	assert.Equal(t, testUser.LastName, capturedUserData["lastName"])
	assert.Equal(t, testUser.Login, capturedUserData["login"])

	mockWS.AssertExpected(t)
}

