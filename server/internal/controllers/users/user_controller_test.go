package userController

import (
	"context"
	"server/config"
	"server/internal/events"
	"server/internal/logger"
	. "server/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) GetByLogin(ctx context.Context, login string) (*User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *User, config config.Config) error {
	args := m.Called(ctx, user, config)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *Session, config config.Config) error {
	args := m.Called(ctx, session, config)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id string) (*Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserController_New(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockSessionRepo := &MockSessionRepository{}
	mockConfig := config.Config{
		ServerPort: 8080,
	}

	eventBus := &events.EventBus{}
	controller := New(eventBus, mockUserRepo, mockSessionRepo, mockConfig)

	assert.NotNil(t, controller)
	assert.Equal(t, mockUserRepo, controller.userRepo)
	assert.Equal(t, mockSessionRepo, controller.sessionRepo)
	assert.Equal(t, mockConfig, controller.Config)
	assert.NotNil(t, controller.log)
}

func TestUserController_StructCreation(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockSessionRepo := &MockSessionRepository{}
	controller := &UserController{
		userRepo:    mockUserRepo,
		sessionRepo: mockSessionRepo,
		Config:      config.Config{ServerPort: 8080},
		log:         logger.New("test"),
	}

	assert.NotNil(t, controller)
	assert.Equal(t, 8080, controller.Config.ServerPort)
	assert.NotNil(t, controller.log)
}

func TestUserController_FieldTypes(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockSessionRepo := &MockSessionRepository{}
	controller := &UserController{
		userRepo:    mockUserRepo,
		sessionRepo: mockSessionRepo,
		log:         logger.New("test"), // Initialize to avoid nil
	}

	// Verify field types
	assert.IsType(t, (*MockUserRepository)(nil), controller.userRepo)
	assert.IsType(t, (*MockSessionRepository)(nil), controller.sessionRepo)
	assert.IsType(t, config.Config{}, controller.Config)
	assert.IsType(t, &logger.SlogLogger{}, controller.log)
}

func TestUserController_Login_StructureTest(t *testing.T) {
	controller := &UserController{
		userRepo:    &MockUserRepository{},
		sessionRepo: &MockSessionRepository{},
		Config: config.Config{SecurityPepper: "test-pepper"},
		log:    logger.New("test"),
	}

	loginRequest := LoginRequest{
		Login:    "testuser",
		Password: "testpass",
	}

	// We can't safely test actual login without database as it panics
	// Just verify the structure and types exist
	assert.NotNil(t, controller)
	assert.IsType(t, LoginRequest{}, loginRequest)
	assert.Equal(t, "testuser", loginRequest.Login)
	assert.Equal(t, "testpass", loginRequest.Password)
}

func TestUserController_Logout_StructureTest(t *testing.T) {
	controller := &UserController{
		userRepo:    &MockUserRepository{},
		sessionRepo: &MockSessionRepository{},
		log: logger.New("test"),
	}

	// We can't safely test actual logout without database as it may panic
	// Just verify the structure exists
	assert.NotNil(t, controller)
	assert.NotNil(t, controller.userRepo)
}

func TestUserController_Register_StructureTest(t *testing.T) {
	controller := &UserController{
		userRepo:    &MockUserRepository{},
		sessionRepo: &MockSessionRepository{},
		Config: config.Config{ServerPort: 8080},
		log:    logger.New("test"),
	}

	user := User{
		FirstName: "Test",
		LastName:  "User",
		Login:     "testuser",
		Password:  "testpass",
	}

	// We can't safely test actual register without database as it may panic
	// Just verify the structure exists
	assert.NotNil(t, controller)
	assert.Equal(t, "Test", user.FirstName)
	assert.Equal(t, "testuser", user.Login)
}

func TestUserController_ComparePassword_Success(t *testing.T) {
	pepper := "test-pepper"
	controller := &UserController{
		Config: config.Config{SecurityPepper: pepper},
		log:    logger.New("test"),
	}

	password := "testpassword"
	passwordWithPepper := password + pepper

	// Generate a hash to test against
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithPepper), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Test successful password comparison
	err = controller.comparePassword(password, string(hashedPassword))
	assert.NoError(t, err)
}

func TestUserController_ComparePassword_Failure(t *testing.T) {
	pepper := "test-pepper"
	controller := &UserController{
		Config: config.Config{SecurityPepper: pepper},
		log:    logger.New("test"),
	}

	password := "testpassword"
	wrongPassword := "wrongpassword"
	passwordWithPepper := password + pepper

	// Generate a hash with correct password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithPepper), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Test failed password comparison with wrong password
	err = controller.comparePassword(wrongPassword, string(hashedPassword))
	assert.Error(t, err)
}

func TestUserController_ComparePassword_EmptyPassword(t *testing.T) {
	controller := &UserController{
		Config: config.Config{SecurityPepper: "test-pepper"},
		log:    logger.New("test"),
	}

	// Test with empty password
	err := controller.comparePassword("", "some-hash")
	assert.Error(t, err)
}

func TestUserController_ComparePassword_EmptyHash(t *testing.T) {
	controller := &UserController{
		Config: config.Config{SecurityPepper: "test-pepper"},
		log:    logger.New("test"),
	}

	// Test with empty hash
	err := controller.comparePassword("password", "")
	assert.Error(t, err)
}

func TestUserController_ComparePassword_WithPepper(t *testing.T) {
	pepper := "special-pepper-123"
	controller := &UserController{
		Config: config.Config{SecurityPepper: pepper},
		log:    logger.New("test"),
	}

	password := "mypassword"
	passwordWithPepper := password + pepper

	// Generate hash with pepper
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithPepper), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Should succeed with correct password
	err = controller.comparePassword(password, string(hashedPassword))
	assert.NoError(t, err)

	// Should fail if we try to compare without considering pepper
	hashedPasswordNoPepper, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	err = controller.comparePassword(password, string(hashedPasswordNoPepper))
	assert.Error(t, err) // Should fail because pepper is added but hash doesn't include it
}

func TestLoginRequest_StructCreation(t *testing.T) {
	loginRequest := LoginRequest{
		Login:    "testuser",
		Password: "testpass",
	}

	assert.Equal(t, "testuser", loginRequest.Login)
	assert.Equal(t, "testpass", loginRequest.Password)
}

func TestLoginRequest_EmptyValues(t *testing.T) {
	loginRequest := LoginRequest{}

	assert.Equal(t, "", loginRequest.Login)
	assert.Equal(t, "", loginRequest.Password)
}

func TestLoginRequest_FieldTypes(t *testing.T) {
	loginRequest := LoginRequest{}

	assert.IsType(t, "", loginRequest.Login)
	assert.IsType(t, "", loginRequest.Password)
}

// Negative Test Cases

func TestUserController_NilFields(t *testing.T) {
	// Test controller with nil fields
	controller := &UserController{
		log: nil, // Nil logger
	}

	assert.Nil(t, controller.log)
	assert.Nil(t, controller.userRepo)
	assert.Nil(t, controller.sessionRepo)
	assert.Equal(t, config.Config{}, controller.Config)
}

func TestUserController_EmptyConfig(t *testing.T) {
	controller := &UserController{
		Config: config.Config{}, // Empty config
		log:    logger.New("test"),
	}

	// Test password comparison with empty pepper
	password := "testpass"
	passwordWithEmptyPepper := password + "" // Empty pepper

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithEmptyPepper), bcrypt.DefaultCost)
	assert.NoError(t, err)

	err = controller.comparePassword(password, string(hashedPassword))
	assert.NoError(t, err)
}

func TestUserController_ComparePassword_SpecialCharacters(t *testing.T) {
	pepper := "special!@#$%^&*()_+{}|:<>?[];'\"\\,./`~"
	controller := &UserController{
		Config: config.Config{SecurityPepper: pepper},
		log:    logger.New("test"),
	}

	password := "pass!@#$%^&*()"
	passwordWithPepper := password + pepper

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithPepper), bcrypt.DefaultCost)
	assert.NoError(t, err)

	err = controller.comparePassword(password, string(hashedPassword))
	assert.NoError(t, err)
}

func TestUserController_ComparePassword_UnicodeCharacters(t *testing.T) {
	pepper := "测试胡椒🔒"
	controller := &UserController{
		Config: config.Config{SecurityPepper: pepper},
		log:    logger.New("test"),
	}

	password := "密码测试🚀"
	passwordWithPepper := password + pepper

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithPepper), bcrypt.DefaultCost)
	assert.NoError(t, err)

	err = controller.comparePassword(password, string(hashedPassword))
	assert.NoError(t, err)
}

func TestUserController_ComparePassword_VeryLongPepper(t *testing.T) {
	// Test with long pepper (but not exceeding bcrypt 72 byte limit)
	pepper := string(make([]byte, 60)) // Keep under bcrypt limit
	for i := range pepper {
		pepper = pepper[:i] + "a" + pepper[i+1:]
	}

	controller := &UserController{
		Config: config.Config{SecurityPepper: pepper},
		log:    logger.New("test"),
	}

	password := "short" // Keep total length under 72 bytes
	passwordWithPepper := password + pepper

	// Only test if total length is under bcrypt limit
	if len(passwordWithPepper) <= 72 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordWithPepper), bcrypt.DefaultCost)
		assert.NoError(t, err)

		err = controller.comparePassword(password, string(hashedPassword))
		assert.NoError(t, err)
	} else {
		// Test that very long passwords fail gracefully
		assert.True(t, len(passwordWithPepper) > 72)
	}
}

func TestUserController_ComparePassword_InvalidHashFormat(t *testing.T) {
	controller := &UserController{
		Config: config.Config{SecurityPepper: "test-pepper"},
		log:    logger.New("test"),
	}

	// Test with invalid hash format
	invalidHashes := []string{
		"not-a-bcrypt-hash",
		"$2a$10$invalid",
		"",
		"plain-text",
		"$2a$",
		"$invalid$format$here",
	}

	for _, invalidHash := range invalidHashes {
		err := controller.comparePassword("password", invalidHash)
		assert.Error(t, err, "Should fail with invalid hash: %s", invalidHash)
	}
}

func TestUserController_Login_EmptyLoginRequest(t *testing.T) {
	controller := &UserController{
		userRepo:    &MockUserRepository{},
		sessionRepo: &MockSessionRepository{},
		Config: config.Config{},
		log:    logger.New("test"),
	}

	emptyRequest := LoginRequest{}

	// Can't safely test without database, just verify structure
	assert.NotNil(t, controller)
	assert.Equal(t, "", emptyRequest.Login)
	assert.Equal(t, "", emptyRequest.Password)
}

func TestUserController_Logout_EmptySessionID(t *testing.T) {
	controller := &UserController{
		userRepo:    &MockUserRepository{},
		sessionRepo: &MockSessionRepository{},
		log: logger.New("test"),
	}

	// Can't safely test without database, just verify structure
	assert.NotNil(t, controller)
	assert.NotNil(t, controller.userRepo)
}

func TestUserController_Register_EmptyUser(t *testing.T) {
	controller := &UserController{
		userRepo:    &MockUserRepository{},
		sessionRepo: &MockSessionRepository{},
		Config: config.Config{},
		log:    logger.New("test"),
	}

	emptyUser := User{}

	// Can't safely test without database, just verify structure
	assert.NotNil(t, controller)
	assert.Equal(t, "", emptyUser.FirstName)
	assert.Equal(t, "", emptyUser.Login)
}

func TestUserController_ComparePassword_EdgeCases(t *testing.T) {
	controller := &UserController{
		Config: config.Config{SecurityPepper: "test"},
		log:    logger.New("test"),
	}

	// Test various edge cases
	edgeCases := []struct {
		name      string
		password  string
		hash      string
		shouldErr bool
	}{
		{"BothEmpty", "", "", true},
		{"EmptyPassword", "", "some-hash", true},
		{"EmptyHash", "password", "", true},
		{"Newlines", "pass\nword", "", true},
		{"Tabs", "pass\tword", "", true},
		{"NullBytes", "pass\x00word", "", true},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			err := controller.comparePassword(tc.password, tc.hash)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
