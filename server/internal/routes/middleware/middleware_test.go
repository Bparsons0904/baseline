package middleware

import (
	"server/config"
	"server/internal/database"
	"server/internal/events"
	"server/internal/logger"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)


func TestMiddleware_New(t *testing.T) {
	mockDB := database.DB{}
	mockUserRepo := &MockUserRepository{}
	mockSessionRepo := &MockSessionRepository{}
	mockConfig := config.Config{ServerPort: 8080}

	eventBus := &events.EventBus{}
	middleware := New(mockDB, eventBus, mockConfig, mockUserRepo, mockSessionRepo)

	assert.NotNil(t, middleware)
	assert.Equal(t, mockDB, middleware.DB)
	assert.Equal(t, mockUserRepo, middleware.userRepo)
	assert.Equal(t, mockSessionRepo, middleware.sessionRepo)
	assert.Equal(t, mockConfig, middleware.Config)
	assert.NotNil(t, middleware.log)
}

func TestMiddleware_StructCreation(t *testing.T) {
	middleware := Middleware{
		DB:     database.DB{},
		Config: config.Config{ServerPort: 8080},
		log:    logger.New("test"),
	}

	assert.Equal(t, 8080, middleware.Config.ServerPort)
	assert.NotNil(t, middleware.log)
}

func TestMiddleware_FieldTypes(t *testing.T) {
	middleware := Middleware{
		log: logger.New("test"), // Initialize to avoid nil
	}

	// Verify field types
	assert.IsType(t, database.DB{}, middleware.DB)
	assert.IsType(t, config.Config{}, middleware.Config)
	assert.IsType(t, &logger.SlogLogger{}, middleware.log)
}

func TestSessionData_StructCreation(t *testing.T) {
	testUUID := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)

	sessionData := SessionData{
		UserID:    testUUID,
		ExpiresAt: expiresAt,
		UserAgent: "Mozilla/5.0",
	}

	assert.Equal(t, testUUID, sessionData.UserID)
	assert.Equal(t, expiresAt, sessionData.ExpiresAt)
	assert.Equal(t, "Mozilla/5.0", sessionData.UserAgent)
}

func TestSessionData_EmptyValues(t *testing.T) {
	sessionData := SessionData{}

	assert.Equal(t, uuid.Nil, sessionData.UserID)
	assert.True(t, sessionData.ExpiresAt.IsZero())
	assert.Equal(t, "", sessionData.UserAgent)
}

func TestSessionData_FieldTypes(t *testing.T) {
	sessionData := SessionData{}

	assert.IsType(t, uuid.UUID{}, sessionData.UserID)
	assert.IsType(t, time.Time{}, sessionData.ExpiresAt)
	assert.IsType(t, "", sessionData.UserAgent)
}

func TestSessionData_JSONTags(t *testing.T) {
	// Test that struct fields are accessible for JSON serialization
	testUUID := uuid.New()
	now := time.Now()

	sessionData := SessionData{
		UserID:    testUUID,
		ExpiresAt: now,
		UserAgent: "test-agent",
	}

	assert.Equal(t, testUUID, sessionData.UserID)
	assert.Equal(t, now, sessionData.ExpiresAt)
	assert.Equal(t, "test-agent", sessionData.UserAgent)
}

func TestClientTypeConstants(t *testing.T) {
	// Test client type constants
	assert.Equal(t, "flutter", MOBILE_CLIENT_TYPE)
	assert.Equal(t, "solid", WEB_CLIENT_TYPE)
}

func TestMiddleware_BasicAuth_Exists(t *testing.T) {
	middleware := Middleware{
		log: logger.New("test"),
	}

	// Test that BasicAuth method exists and returns a handler function
	handler := middleware.BasicAuth()
	assert.NotNil(t, handler)
}

func TestMiddleware_AuthRequired_Exists(t *testing.T) {
	middleware := Middleware{
		log: logger.New("test"),
	}

	// Test that AuthRequired method exists and returns a handler function
	handler := middleware.AuthRequired()
	assert.NotNil(t, handler)
}

func TestMiddleware_AuthNoContent_Exists(t *testing.T) {
	middleware := Middleware{
		log: logger.New("test"),
	}

	// Test that AuthNoContent method exists and returns a handler function
	handler := middleware.AuthNoContent()
	assert.NotNil(t, handler)
}

func TestSessionData_TimeHandling(t *testing.T) {
	// Test time-related functionality
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	sessionData := SessionData{
		ExpiresAt: future,
	}

	assert.True(t, sessionData.ExpiresAt.After(now))
	assert.False(t, sessionData.ExpiresAt.Before(now))

	sessionData.ExpiresAt = past
	assert.True(t, sessionData.ExpiresAt.Before(now))
	assert.False(t, sessionData.ExpiresAt.After(now))
}

func TestSessionData_UUIDHandling(t *testing.T) {
	// Test UUID handling
	nilUUID := uuid.Nil
	validUUID := uuid.New()

	sessionData := SessionData{UserID: nilUUID}
	assert.Equal(t, uuid.Nil, sessionData.UserID)
	assert.True(t, sessionData.UserID == uuid.Nil)

	sessionData.UserID = validUUID
	assert.NotEqual(t, uuid.Nil, sessionData.UserID)
	assert.False(t, sessionData.UserID == uuid.Nil)
}

// Negative Test Cases

func TestMiddleware_ZeroValues(t *testing.T) {
	// Test middleware with zero values
	middleware := Middleware{}

	assert.Equal(t, database.DB{}, middleware.DB)
	assert.Equal(t, config.Config{}, middleware.Config)
	assert.Nil(t, middleware.log)
}

func TestMiddleware_NilLogger(t *testing.T) {
	middleware := Middleware{
		DB:     database.DB{},
		Config: config.Config{},
		log:    nil,
	}

	assert.Nil(t, middleware.log)

	// Test that we can still access other fields
	assert.NotNil(t, middleware.DB)
	assert.NotNil(t, middleware.Config)
}

func TestSessionData_ExtremeValues(t *testing.T) {
	// Test with extreme time values
	veryFarFuture := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	veryFarPast := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

	sessionData := SessionData{
		ExpiresAt: veryFarFuture,
	}
	assert.Equal(t, veryFarFuture, sessionData.ExpiresAt)

	sessionData.ExpiresAt = veryFarPast
	assert.Equal(t, veryFarPast, sessionData.ExpiresAt)
}

func TestSessionData_SpecialCharacterUserAgent(t *testing.T) {
	// Test with special characters in user agent
	specialUserAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"User-Agent with special chars !@#$%^&*()",
		"测试用户代理🚀",
		"Agent\nwith\nnewlines",
		"Agent\twith\ttabs",
		"Agent with\x00null\x01bytes",
		"", // Empty user agent
	}

	for _, userAgent := range specialUserAgents {
		sessionData := SessionData{
			UserAgent: userAgent,
		}
		assert.Equal(t, userAgent, sessionData.UserAgent)
	}
}

func TestSessionData_LongUserAgent(t *testing.T) {
	// Test with very long user agent string
	longUserAgent := make([]byte, 10000)
	for i := range longUserAgent {
		longUserAgent[i] = 'a'
	}

	sessionData := SessionData{
		UserAgent: string(longUserAgent),
	}

	assert.Len(t, sessionData.UserAgent, 10000)
	assert.Equal(t, string(longUserAgent), sessionData.UserAgent)
}

func TestClientTypeConstants_Immutability(t *testing.T) {
	// Test that constants maintain their values
	originalMobile := MOBILE_CLIENT_TYPE
	originalWeb := WEB_CLIENT_TYPE

	// Constants should remain the same
	assert.Equal(t, "flutter", originalMobile)
	assert.Equal(t, "solid", originalWeb)
	assert.Equal(t, originalMobile, MOBILE_CLIENT_TYPE)
	assert.Equal(t, originalWeb, WEB_CLIENT_TYPE)
}

func TestSessionData_CopyBehavior(t *testing.T) {
	// Test copying behavior
	original := SessionData{
		UserID:    uuid.New(),
		ExpiresAt: time.Now(),
		UserAgent: "original-agent",
	}

	// Copy the struct
	copied := original

	// Modify the copy
	copied.UserAgent = "modified-agent"

	// Original should remain unchanged
	assert.Equal(t, "original-agent", original.UserAgent)
	assert.Equal(t, "modified-agent", copied.UserAgent)

	// UUIDs and times should be the same (copied)
	assert.Equal(t, original.UserID, copied.UserID)
	assert.Equal(t, original.ExpiresAt, copied.ExpiresAt)
}

func TestSessionData_Comparison(t *testing.T) {
	// Test struct comparison
	testUUID := uuid.New()
	testTime := time.Now()

	sessionData1 := SessionData{
		UserID:    testUUID,
		ExpiresAt: testTime,
		UserAgent: "same-agent",
	}

	sessionData2 := SessionData{
		UserID:    testUUID,
		ExpiresAt: testTime,
		UserAgent: "same-agent",
	}

	sessionData3 := SessionData{
		UserID:    testUUID,
		ExpiresAt: testTime,
		UserAgent: "different-agent",
	}

	// Should be equal
	assert.Equal(t, sessionData1, sessionData2)

	// Should not be equal
	assert.NotEqual(t, sessionData1, sessionData3)
}

func TestMiddleware_ConfigAccess(t *testing.T) {
	// Test accessing config fields
	config := config.Config{
		ServerPort:     8080,
		SecurityPepper: "test-pepper",
		SecuritySalt:   12,
	}

	middleware := Middleware{
		Config: config,
	}

	assert.Equal(t, 8080, middleware.Config.ServerPort)
	assert.Equal(t, "test-pepper", middleware.Config.SecurityPepper)
	assert.Equal(t, 12, middleware.Config.SecuritySalt)
}

func TestMiddleware_DatabaseAccess(t *testing.T) {
	// Test accessing database fields
	db := database.DB{}

	middleware := Middleware{
		DB: db,
	}

	assert.Equal(t, db, middleware.DB)
	assert.IsType(t, database.DB{}, middleware.DB)
}

func TestSessionData_ZeroTimeComparison(t *testing.T) {
	// Test zero time handling
	zeroTime := time.Time{}
	sessionData := SessionData{
		ExpiresAt: zeroTime,
	}

	assert.True(t, sessionData.ExpiresAt.IsZero())
	assert.True(t, sessionData.ExpiresAt.Before(time.Now()))
}

func TestSessionData_TimeZoneHandling(t *testing.T) {
	// Test different time zones
	utc := time.Now().UTC()
	local := time.Now().Local()

	sessionData := SessionData{ExpiresAt: utc}
	assert.Equal(t, utc, sessionData.ExpiresAt)

	sessionData.ExpiresAt = local
	assert.Equal(t, local, sessionData.ExpiresAt)

	// They should represent different times (unless local is UTC)
	if time.Local != time.UTC {
		assert.NotEqual(t, utc, local)
	}
}
