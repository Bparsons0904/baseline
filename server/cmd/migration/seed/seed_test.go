package seed

import (
	"server/config"
	"server/internal/logger"
	. "server/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, config.Config) {
	testConfig := config.Config{
		SecuritySalt:   12,
		SecurityPepper: "test-pepper",
	}

	// Set the global config instance for the utils package
	config.ConfigInstance = testConfig

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	return db, testConfig
}

func TestSeed(t *testing.T) {
	tests := []struct {
		name          string
		expectedUsers int
		setupDB       func(*gorm.DB)
	}{
		{
			name:          "seed empty database",
			expectedUsers: 3,
			setupDB:       func(db *gorm.DB) {},
		},
		{
			name:          "seed database with existing user",
			expectedUsers: 3,
			setupDB: func(db *gorm.DB) {
				existingUser := User{
					FirstName: "Bob",
					LastName:  "Parsons",
					Login:     "deadstyle",
					Password:  "existing_password",
					IsAdmin:   true,
				}
				db.Create(&existingUser)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, config := setupTestDB(t)
			log := logger.New("test")

			tt.setupDB(db)

			err := Seed(db, config, log)
			assert.NoError(t, err)

			var userCount int64
			err = db.Model(&User{}).Count(&userCount).Error
			assert.NoError(t, err)
			assert.Equal(t, int64(tt.expectedUsers), userCount)

			var users []User
			err = db.Find(&users).Error
			assert.NoError(t, err)

			expectedLogins := []string{"deadstyle", "bobb", "ada"}
			actualLogins := make([]string, len(users))
			for i, user := range users {
				actualLogins[i] = user.Login
			}

			for _, expectedLogin := range expectedLogins {
				assert.Contains(t, actualLogins, expectedLogin)
			}
		})
	}
}

func TestSeed_UserCreation(t *testing.T) {
	db, config := setupTestDB(t)
	log := logger.New("test")

	err := Seed(db, config, log)
	require.NoError(t, err)

	t.Run("deadstyle user created correctly", func(t *testing.T) {
		var user User
		err := db.First(&user, "login = ?", "deadstyle").Error
		require.NoError(t, err)

		assert.Equal(t, "Bob", user.FirstName)
		assert.Equal(t, "Parsons", user.LastName)
		assert.Equal(t, "deadstyle", user.Login)
		assert.True(t, user.IsAdmin)
		assert.NotEqual(t, "password", user.Password) // Should be hashed
	})

	t.Run("bobb user created correctly", func(t *testing.T) {
		var user User
		err := db.First(&user, "login = ?", "bobb").Error
		require.NoError(t, err)

		assert.Equal(t, "Bob", user.FirstName)
		assert.Equal(t, "Covell", user.LastName)
		assert.Equal(t, "bobb", user.Login)
		assert.True(t, user.IsAdmin)
		assert.NotEqual(t, "password", user.Password) // Should be hashed
	})

	t.Run("ada user created correctly", func(t *testing.T) {
		var user User
		err := db.First(&user, "login = ?", "ada").Error
		require.NoError(t, err)

		assert.Equal(t, "Ada", user.FirstName)
		assert.Equal(t, "Lovelace", user.LastName)
		assert.Equal(t, "ada", user.Login)
		assert.False(t, user.IsAdmin)
		assert.NotEqual(t, "password", user.Password) // Should be hashed
	})
}

func TestSeed_Idempotency(t *testing.T) {
	db, config := setupTestDB(t)
	log := logger.New("test")

	// Run seed twice
	err := Seed(db, config, log)
	require.NoError(t, err)

	err = Seed(db, config, log)
	require.NoError(t, err)

	// Should still only have 3 users
	var userCount int64
	err = db.Model(&User{}).Count(&userCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(3), userCount)
}

func TestSeed_UserCreationError(t *testing.T) {
	db, config := setupTestDB(t)
	log := logger.New("test")

	// Create a user with the same login to cause a constraint violation
	existingUser := User{
		FirstName: "Existing",
		LastName:  "User",
		Login:     "deadstyle",
		Password:  "existing_password",
		IsAdmin:   false,
	}
	err := db.Create(&existingUser).Error
	require.NoError(t, err)

	// Seed should complete without error (it checks for existing users)
	err = Seed(db, config, log)
	assert.NoError(t, err)

	// Verify the existing user wasn't overwritten
	var user User
	err = db.First(&user, "login = ?", "deadstyle").Error
	require.NoError(t, err)
	assert.Equal(t, "Existing", user.FirstName)
	assert.Equal(t, "User", user.LastName)
	assert.False(t, user.IsAdmin)
}
