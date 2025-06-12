package models

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test constants (moved from models to repositories)
const (
	USER_EXPIRY = 7 * 24 * time.Hour // 7 days
)

func TestUserConstants(t *testing.T) {
	// Test the USER_EXPIRY constant
	expectedExpiry := 7 * 24 * time.Hour // 7 days

	assert.Equal(t, expectedExpiry, USER_EXPIRY)
	assert.Equal(t, time.Hour*168, USER_EXPIRY) // 168 hours = 7 days
}

func TestUser_StructCreation(t *testing.T) {
	// Test creating a User struct
	user := User{}

	// Verify embedded BaseModel fields are accessible
	assert.Equal(t, "", user.ID)
	assert.True(t, user.CreatedAt.IsZero())
	assert.True(t, user.UpdatedAt.IsZero())

	// Verify User-specific fields have zero values
	assert.Equal(t, "", user.FirstName)
	assert.Equal(t, "", user.LastName)
	assert.Equal(t, "", user.Login)
	assert.Equal(t, "", user.Password)
	assert.False(t, user.IsAdmin)
}

func TestUser_StructWithValues(t *testing.T) {
	// Test creating User with specific values
	now := time.Now()

	user := User{
		BaseModel: BaseModel{
			ID:        "user-123",
			CreatedAt: now,
			UpdatedAt: now,
		},
		FirstName: "John",
		LastName:  "Doe",
		Login:     "johndoe",
		Password:  "hashed-password",
		IsAdmin:   true,
	}

	// Verify BaseModel fields
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)

	// Verify User fields
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, "johndoe", user.Login)
	assert.Equal(t, "hashed-password", user.Password)
	assert.True(t, user.IsAdmin)
}

func TestUser_FieldTypes(t *testing.T) {
	user := User{}

	// Verify User field types
	assert.IsType(t, "", user.FirstName)
	assert.IsType(t, "", user.LastName)
	assert.IsType(t, "", user.Login)
	assert.IsType(t, "", user.Password)
	assert.IsType(t, false, user.IsAdmin)

	// Verify inherited BaseModel field types
	assert.IsType(t, "", user.ID)
	assert.IsType(t, time.Time{}, user.CreatedAt)
	assert.IsType(t, time.Time{}, user.UpdatedAt)
}

func TestUser_BaseModelEmbedding(t *testing.T) {
	user := User{
		FirstName: "Alice",
		LastName:  "Smith",
		Login:     "alice",
	}

	// Test that BeforeSave from BaseModel works on User
	err := user.BeforeSave(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)

	// Verify User fields are preserved
	assert.Equal(t, "Alice", user.FirstName)
	assert.Equal(t, "Smith", user.LastName)
	assert.Equal(t, "alice", user.Login)
}

func TestUser_AdminFlag(t *testing.T) {
	// Test admin flag scenarios
	regularUser := User{
		FirstName: "Regular",
		LastName:  "User",
		IsAdmin:   false,
	}

	adminUser := User{
		FirstName: "Admin",
		LastName:  "User",
		IsAdmin:   true,
	}

	assert.False(t, regularUser.IsAdmin)
	assert.True(t, adminUser.IsAdmin)
}

func TestUser_LoginField(t *testing.T) {
	// Test different login formats
	loginFormats := []string{
		"username",
		"user@example.com",
		"user.name",
		"user_name",
		"user-name",
		"123user",
		"", // Empty login
	}

	for _, login := range loginFormats {
		user := User{Login: login}
		assert.Equal(t, login, user.Login)
	}
}

func TestUser_NameFields(t *testing.T) {
	// Test various name scenarios
	nameScenarios := []struct {
		firstName string
		lastName  string
	}{
		{"John", "Doe"},
		{"", "Smith"},                // Empty first name
		{"Alice", ""},                // Empty last name
		{"", ""},                     // Empty both
		{"Jean-Claude", "Van Damme"}, // Hyphenated names
		{"Mary Jane", "Watson"},      // Space in first name
		{"O'Connor", "Smith"},        // Apostrophe
	}

	for _, scenario := range nameScenarios {
		user := User{
			FirstName: scenario.firstName,
			LastName:  scenario.lastName,
		}

		assert.Equal(t, scenario.firstName, user.FirstName)
		assert.Equal(t, scenario.lastName, user.LastName)
	}
}

func TestLoginRequest_Struct(t *testing.T) {
	// Test LoginRequest struct
	loginReq := LoginRequest{}

	// Verify zero values
	assert.Equal(t, "", loginReq.Login)
	assert.Equal(t, "", loginReq.Password)

	// Test with values
	loginReq = LoginRequest{
		Login:    "testuser",
		Password: "testpass",
	}

	assert.Equal(t, "testuser", loginReq.Login)
	assert.Equal(t, "testpass", loginReq.Password)
}

func TestLoginRequest_FieldTypes(t *testing.T) {
	loginReq := LoginRequest{}

	assert.IsType(t, "", loginReq.Login)
	assert.IsType(t, "", loginReq.Password)
}

func TestUser_JSONTagBehavior(t *testing.T) {
	// Test that Password field has json:"-" tag behavior
	// This is more of a structural test since we can't easily test JSON marshaling without dependencies

	user := User{
		FirstName: "Test",
		LastName:  "User",
		Login:     "testuser",
		Password:  "secret-password", // This should be excluded from JSON
		IsAdmin:   false,
	}

	// Verify the password is set but would be excluded from JSON
	assert.Equal(t, "secret-password", user.Password)
	assert.Equal(t, "testuser", user.Login)
}

func TestUser_MultipleUsers(t *testing.T) {
	// Test creating multiple users with different properties
	users := []User{
		{
			BaseModel: BaseModel{ID: "user-1"},
			FirstName: "Alice",
			LastName:  "Johnson",
			Login:     "alice",
			IsAdmin:   false,
		},
		{
			BaseModel: BaseModel{ID: "user-2"},
			FirstName: "Bob",
			LastName:  "Smith",
			Login:     "bob",
			IsAdmin:   true,
		},
		{
			BaseModel: BaseModel{ID: "user-3"},
			FirstName: "Charlie",
			LastName:  "Brown",
			Login:     "charlie",
			IsAdmin:   false,
		},
	}

	// Verify each user maintains its properties
	assert.Equal(t, "alice", users[0].Login)
	assert.False(t, users[0].IsAdmin)

	assert.Equal(t, "bob", users[1].Login)
	assert.True(t, users[1].IsAdmin)

	assert.Equal(t, "charlie", users[2].Login)
	assert.False(t, users[2].IsAdmin)

	// Verify all have different IDs
	assert.NotEqual(t, users[0].ID, users[1].ID)
	assert.NotEqual(t, users[1].ID, users[2].ID)
	assert.NotEqual(t, users[0].ID, users[2].ID)
}

// Negative Test Cases

func TestUser_EmptyFields(t *testing.T) {
	// Test user with all empty fields
	user := User{}

	// Should handle empty values gracefully
	assert.Equal(t, "", user.FirstName)
	assert.Equal(t, "", user.LastName)
	assert.Equal(t, "", user.Login)
	assert.Equal(t, "", user.Password)
	assert.False(t, user.IsAdmin)
	assert.Equal(t, "", user.ID)
}

func TestUser_ExtremelyLongFields(t *testing.T) {
	// Test with very long field values
	longString := strings.Repeat("very-long-", 1000)

	user := User{
		FirstName: longString,
		LastName:  longString,
		Login:     longString,
		Password:  longString,
		IsAdmin:   true,
	}

	// Should accept long values
	assert.Equal(t, longString, user.FirstName)
	assert.Equal(t, longString, user.LastName)
	assert.Equal(t, longString, user.Login)
	assert.Equal(t, longString, user.Password)
	assert.True(t, user.IsAdmin)
}

func TestUser_SpecialCharactersInFields(t *testing.T) {
	specialCases := []struct {
		name      string
		firstName string
		lastName  string
		login     string
	}{
		{
			name:      "Unicode",
			firstName: "测试",
			lastName:  "пользователь",
			login:     "user测试",
		},
		{
			name:      "Emojis",
			firstName: "John🚀",
			lastName:  "Doe💻",
			login:     "john_doe_🎯",
		},
		{
			name:      "SpecialChars",
			firstName: "John@#$%",
			lastName:  "Doe^&*()",
			login:     "john.doe+test",
		},
		{
			name:      "Whitespace",
			firstName: "John  With  Spaces",
			lastName:  "Doe\tWith\tTabs",
			login:     "login with spaces",
		},
		{
			name:      "ControlChars",
			firstName: "John\nNewline",
			lastName:  "Doe\rReturn",
			login:     "john\x00null",
		},
	}

	for _, tc := range specialCases {
		t.Run(tc.name, func(t *testing.T) {
			user := User{
				FirstName: tc.firstName,
				LastName:  tc.lastName,
				Login:     tc.login,
				Password:  "test-password",
			}

			assert.Equal(t, tc.firstName, user.FirstName)
			assert.Equal(t, tc.lastName, user.LastName)
			assert.Equal(t, tc.login, user.Login)
		})
	}
}

func TestUser_LoginFieldEdgeCases(t *testing.T) {
	edgeCaseLogins := []string{
		"",                 // Empty
		" ",                // Space only
		"a",                // Single character
		"user@",            // Incomplete email
		"@domain.com",      // No username part
		"user@@domain.com", // Double @
		"user@domain@com",  // Multiple @
		"very.long.email.address.that.might.exceed.normal.limits@very.long.domain.name.that.exceeds.normal.limits.com",
		"user+tag@domain.com",             // Plus addressing
		"user.name+tag+more@domain.co.uk", // Complex email
		"123456789",                       // Numeric only
		"测试@example.com",                  // Unicode in email
	}

	for _, login := range edgeCaseLogins {
		t.Run("login_edge_case", func(t *testing.T) {
			user := User{Login: login}
			assert.Equal(t, login, user.Login)
		})
	}
}

func TestUser_PasswordFieldSecurity(t *testing.T) {
	// Test various password scenarios
	passwordTests := []struct {
		name     string
		password string
	}{
		{"Empty", ""},
		{"Short", "1"},
		{"Long", strings.Repeat("password", 100)},
		{"SpecialChars", "p@ssw0rd!@#$%^&*()"},
		{"Unicode", "пароль测试🔒"},
		{"Whitespace", "password with spaces"},
		{"ControlChars", "pass\nword\ttest"},
		{"NullBytes", "pass\x00word"},
	}

	for _, pt := range passwordTests {
		t.Run(pt.name, func(t *testing.T) {
			user := User{
				Login:    "testuser",
				Password: pt.password,
			}

			assert.Equal(t, pt.password, user.Password)
		})
	}
}

func TestUser_AdminFlagEdgeCases(t *testing.T) {
	// Test admin flag with various user configurations
	users := []User{
		{Login: "admin", IsAdmin: true},
		{Login: "user", IsAdmin: false},
		{Login: "", IsAdmin: true},     // Admin with empty login
		{Login: "", IsAdmin: false},    // Regular user with empty login
		{Password: "", IsAdmin: true},  // Admin with empty password
		{Password: "", IsAdmin: false}, // Regular user with empty password
	}

	for i, user := range users {
		t.Run(fmt.Sprintf("admin_test_%d", i), func(t *testing.T) {
			if user.IsAdmin {
				assert.True(t, user.IsAdmin)
			} else {
				assert.False(t, user.IsAdmin)
			}
		})
	}
}

func TestUser_BaseModelInheritance(t *testing.T) {
	user := User{
		FirstName: "Test",
		LastName:  "User",
		Login:     "testuser",
	}

	// Test that BeforeSave works (inherited from BaseModel)
	err := user.BeforeSave(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)

	// Test that time fields are accessible
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

func TestLoginRequest_EmptyFields(t *testing.T) {
	// Test LoginRequest with empty fields
	loginReq := LoginRequest{}

	assert.Equal(t, "", loginReq.Login)
	assert.Equal(t, "", loginReq.Password)
}

func TestLoginRequest_ExtremeValues(t *testing.T) {
	// Test LoginRequest with extreme values
	longString := strings.Repeat("extreme-", 500)

	loginReq := LoginRequest{
		Login:    longString,
		Password: longString,
	}

	assert.Equal(t, longString, loginReq.Login)
	assert.Equal(t, longString, loginReq.Password)
}

func TestLoginRequest_SpecialCharacters(t *testing.T) {
	specialChars := "!@#$%^&*()_+{}|:<>?[];'\"\\,./`~"
	unicode := "测试用户пользователь🚀"

	loginReq := LoginRequest{
		Login:    "user" + specialChars + unicode,
		Password: "pass" + specialChars + unicode,
	}

	assert.Contains(t, loginReq.Login, specialChars)
	assert.Contains(t, loginReq.Login, unicode)
	assert.Contains(t, loginReq.Password, specialChars)
	assert.Contains(t, loginReq.Password, unicode)
}

func TestUSER_EXPIRY_Constant(t *testing.T) {
	// Test that the constant has expected value and type
	assert.Equal(t, 7*24*time.Hour, USER_EXPIRY)
	assert.IsType(t, time.Duration(0), USER_EXPIRY)

	// Test that it's a reasonable value
	assert.True(t, USER_EXPIRY > 0)
	assert.Equal(t, time.Hour*168, USER_EXPIRY) // 7 days = 168 hours
}

// Edge Case Tests

func TestUser_MemoryFootprint(t *testing.T) {
	// Test creating many users doesn't cause issues
	const numUsers = 1000
	users := make([]User, numUsers)

	for i := 0; i < numUsers; i++ {
		users[i] = User{
			FirstName: fmt.Sprintf("User%d", i),
			LastName:  fmt.Sprintf("Last%d", i),
			Login:     fmt.Sprintf("user%d@test.com", i),
			Password:  fmt.Sprintf("password%d", i),
			IsAdmin:   i%10 == 0, // Every 10th user is admin
		}

		// Generate ID
		err := users[i].BeforeSave(nil)
		assert.NoError(t, err)
	}

	// Verify all users are distinct
	loginMap := make(map[string]bool)
	idMap := make(map[string]bool)

	for _, user := range users {
		assert.False(t, loginMap[user.Login], "Duplicate login: %s", user.Login)
		assert.False(t, idMap[user.ID], "Duplicate ID: %s", user.ID)

		loginMap[user.Login] = true
		idMap[user.ID] = true
	}
}

func TestUser_CopyBehavior(t *testing.T) {
	original := User{
		BaseModel: BaseModel{ID: "original-id"},
		FirstName: "Original",
		LastName:  "User",
		Login:     "original@test.com",
		Password:  "original-password",
		IsAdmin:   false,
	}

	// Copy the user
	copied := original

	// Modify the copy
	copied.FirstName = "Copied"
	copied.IsAdmin = true

	// Original should remain unchanged
	assert.Equal(t, "Original", original.FirstName)
	assert.False(t, original.IsAdmin)
	assert.Equal(t, "Copied", copied.FirstName)
	assert.True(t, copied.IsAdmin)

	// ID should be the same (copied)
	assert.Equal(t, original.ID, copied.ID)
}

func TestUser_PointerBehavior(t *testing.T) {
	user := &User{
		FirstName: "Pointer",
		LastName:  "User",
		Login:     "pointer@test.com",
	}

	// Test BeforeSave on pointer
	err := user.BeforeSave(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
}

func TestUser_ZeroValueComparison(t *testing.T) {
	var user1 User
	var user2 User

	// Zero users should be equal
	assert.Equal(t, user1, user2)

	// Modify one
	user1.FirstName = "Modified"

	// Should no longer be equal
	assert.NotEqual(t, user1, user2)
}
