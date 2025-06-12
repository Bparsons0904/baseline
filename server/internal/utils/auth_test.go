package utils

import (
	"server/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthTestConfig() {
	testConfig := config.Config{
		SecuritySalt:   12,
		SecurityPepper: "test-pepper-for-auth",
	}
	config.ConfigInstance = testConfig
}

func TestHashPassword(t *testing.T) {
	setupAuthTestConfig()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password that exceeds bcrypt limit",
			password: "this-is-a-very-long-password-with-many-characters-and-symbols!@#$%^&*()_+more-text-to-exceed-72-bytes-limit-after-adding-pepper",
			wantErr:  true,
		},
		{
			name:     "special characters",
			password: "p@ssw0rd!#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hashedPassword)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashedPassword)
				assert.NotEqual(t, tt.password, hashedPassword)

				// Verify it's a valid bcrypt hash
				verifyErr := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(tt.password+"test-pepper-for-auth"))
				assert.NoError(t, verifyErr, "hashed password should be valid bcrypt hash with pepper")
			}
		})
	}
}

func TestHashPassword_WithDifferentPasswords(t *testing.T) {
	setupAuthTestConfig()

	password1 := "password1"
	password2 := "password2"

	hash1, err1 := HashPassword(password1)
	require.NoError(t, err1)

	hash2, err2 := HashPassword(password2)
	require.NoError(t, err2)

	// Different passwords should produce different hashes
	assert.NotEqual(t, hash1, hash2)
}

func TestHashPassword_SamePasswordDifferentHashes(t *testing.T) {
	setupAuthTestConfig()

	password := "samepassword"

	hash1, err1 := HashPassword(password)
	require.NoError(t, err1)

	hash2, err2 := HashPassword(password)
	require.NoError(t, err2)

	// Same password should produce different hashes due to salt randomization
	assert.NotEqual(t, hash1, hash2)

	// But both should be valid
	verifyErr1 := bcrypt.CompareHashAndPassword([]byte(hash1), []byte(password+"test-pepper-for-auth"))
	assert.NoError(t, verifyErr1)

	verifyErr2 := bcrypt.CompareHashAndPassword([]byte(hash2), []byte(password+"test-pepper-for-auth"))
	assert.NoError(t, verifyErr2)
}

func TestHashPassword_NoConfig(t *testing.T) {
	// Save original config
	originalConfig := config.ConfigInstance

	// Set empty config
	config.ConfigInstance = config.Config{}

	hashedPassword, err := HashPassword("password")
	assert.Error(t, err)
	assert.Empty(t, hashedPassword)
	assert.Contains(t, err.Error(), "salt or pepper is empty")

	// Restore original config
	config.ConfigInstance = originalConfig
}

func TestHashPassword_NoSalt(t *testing.T) {
	// Save original config
	originalConfig := config.ConfigInstance

	// Set config with no salt
	config.ConfigInstance = config.Config{
		SecuritySalt:   0,
		SecurityPepper: "test-pepper",
	}

	hashedPassword, err := HashPassword("password")
	assert.Error(t, err)
	assert.Empty(t, hashedPassword)
	assert.Contains(t, err.Error(), "salt or pepper is empty")

	// Restore original config
	config.ConfigInstance = originalConfig
}

func TestHashPassword_NoPepper(t *testing.T) {
	// Save original config
	originalConfig := config.ConfigInstance

	// Set config with no pepper
	config.ConfigInstance = config.Config{
		SecuritySalt:   12,
		SecurityPepper: "",
	}

	hashedPassword, err := HashPassword("password")
	assert.Error(t, err)
	assert.Empty(t, hashedPassword)
	assert.Contains(t, err.Error(), "salt or pepper is empty")

	// Restore original config
	config.ConfigInstance = originalConfig
}

func TestHashPassword_WeakSalt(t *testing.T) {
	// Save original config
	originalConfig := config.ConfigInstance

	// Set config with weak salt (but still valid - bcrypt accepts 4-31)
	config.ConfigInstance = config.Config{
		SecuritySalt:   4, // Minimum valid bcrypt cost
		SecurityPepper: "test-pepper",
	}

	hashedPassword, err := HashPassword("password")
	assert.NoError(t, err, "bcrypt should accept cost of 4")
	assert.NotEmpty(t, hashedPassword)

	// Restore original config
	config.ConfigInstance = originalConfig
}

func TestHashPassword_BcryptLimits(t *testing.T) {
	// Save original config
	originalConfig := config.ConfigInstance

	// Test with bcrypt cost too high
	config.ConfigInstance = config.Config{
		SecuritySalt:   32, // Too high for bcrypt (max is 31)
		SecurityPepper: "test-pepper",
	}

	hashedPassword, err := HashPassword("password")
	assert.Error(t, err)
	assert.Empty(t, hashedPassword)

	// Restore original config
	config.ConfigInstance = originalConfig
}

func TestHashPassword_PepperIncludedInHash(t *testing.T) {
	setupAuthTestConfig()

	password := "testpassword"
	pepper := "test-pepper-for-auth"

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)

	// Verify the hash was created with password + pepper
	verifyWithPepper := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+pepper))
	assert.NoError(t, verifyWithPepper, "hash should include pepper")

	// Verify the hash fails without pepper
	verifyWithoutPepper := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.Error(t, verifyWithoutPepper, "hash should fail without pepper")
}

func TestHashPassword_RealisticScenarios(t *testing.T) {
	setupAuthTestConfig()

	// Test realistic password scenarios
	realisticPasswords := []string{
		"MySecurePassword123!",
		"user@domain.com",
		"P@ssw0rd",
		"ThisIsALongPassphraseWithSpaces And Symbols!",
		"简单密码", // Unicode characters
		"🔒🔑💻",  // Emoji
	}

	for _, password := range realisticPasswords {
		t.Run("realistic_password_"+password, func(t *testing.T) {
			hashedPassword, err := HashPassword(password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hashedPassword)
			assert.NotEqual(t, password, hashedPassword)

			// Verify it's a proper bcrypt hash
			verifyErr := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+"test-pepper-for-auth"))
			assert.NoError(t, verifyErr)
		})
	}
}
