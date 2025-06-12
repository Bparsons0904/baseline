package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test InitConfig function

func TestInitConfig_WithEnvFile_Success(t *testing.T) {
	// Clear all environment variables that might interfere
	clearEnvVars(t)
	
	// Create temporary .env file
	envContent := `GENERAL_VERSION=1.2.3
ENVIRONMENT=test
SERVER_PORT=9999
DB_PATH=/tmp/test.db
DB_CACHE_ADDRESS=localhost
DB_CACHE_PORT=6379
CORS_ALLOW_ORIGINS=http://localhost:3000
SECURITY_SALT=12
SECURITY_PEPPER=test-pepper
SECURITY_JWT_SECRET=test-jwt-secret`

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	config, err := InitConfig()

	assert.NoError(t, err)
	assert.Equal(t, "1.2.3", config.GeneralVersion)
	assert.Equal(t, "test", config.Environment)
	assert.Equal(t, 9999, config.ServerPort)
	assert.Equal(t, "/tmp/test.db", config.DatabaseDbPath)
	assert.Equal(t, "localhost", config.DatabaseCacheAddress)
	assert.Equal(t, 6379, config.DatabaseCachePort)
	assert.Equal(t, "http://localhost:3000", config.CorsAllowOrigins)
	assert.Equal(t, 12, config.SecuritySalt)
	assert.Equal(t, "test-pepper", config.SecurityPepper)
	assert.Equal(t, "test-jwt-secret", config.SecurityJwtSecret)
}

func TestInitConfig_WithEnvFile_MinimalValid(t *testing.T) {
	// Clear all environment variables that might interfere
	clearEnvVars(t)
	
	// Create minimal .env file with just required fields
	envContent := `SERVER_PORT=8080`

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	config, err := InitConfig()

	assert.NoError(t, err)
	assert.Equal(t, 8080, config.ServerPort)
	// Other fields should be zero values
	assert.Equal(t, "", config.GeneralVersion)
	assert.Equal(t, "", config.Environment)
	assert.Equal(t, 0, config.SecuritySalt)
}

func TestInitConfig_WithEnvFile_InvalidPort(t *testing.T) {
	// Clear all environment variables that might interfere
	clearEnvVars(t)
	
	// Create .env file with invalid port
	envContent := `SERVER_PORT=-1`

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	_, err = InitConfig()

	// Should return an error for invalid port
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid port: -1")
	}
}

func TestInitConfig_NoEnvFile_WithEnvVars(t *testing.T) {
	// Ensure no .env file exists
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	// Create a temporary directory without .env file
	tempDir := createTempDir(t)
	defer func() { _ = os.RemoveAll(tempDir) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Set environment variables (viper with AutomaticEnv picks these up)
	err = os.Setenv("SERVER_PORT", "7777")
	require.NoError(t, err)
	defer func() { _ = os.Unsetenv("SERVER_PORT") }()

	config, err := InitConfig()

	assert.NoError(t, err)
	assert.Equal(t, 7777, config.ServerPort)
	// Other fields will be zero values since no .env file and no other env vars set
}

func TestInitConfig_EnvVarsOverrideEnvFile(t *testing.T) {
	// Create .env file
	envContent := `SERVER_PORT=8080
ENVIRONMENT=development`

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	// Set environment variables that should override .env file
	err = os.Setenv("SERVER_PORT", "9090")
	require.NoError(t, err)
	defer func() { _ = os.Unsetenv("SERVER_PORT") }()

	err = os.Setenv("ENVIRONMENT", "production")
	require.NoError(t, err)
	defer func() { _ = os.Unsetenv("ENVIRONMENT") }()

	config, err := InitConfig()

	assert.NoError(t, err)
	// Environment variables should override .env file values
	assert.Equal(t, 9090, config.ServerPort)
	assert.Equal(t, "production", config.Environment)
}

func TestInitConfig_EmptyEnvFile(t *testing.T) {
	// Create empty .env file
	envFile := createTempEnvFile(t, "")
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	_, err = InitConfig()

	// Should fail due to zero port value
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port: 0")
}

func TestInitConfig_MalformedEnvFile(t *testing.T) {
	// Create malformed .env file (but viper can still parse the valid lines)
	envContent := `# This is a comment
SERVER_PORT=8080
ENVIRONMENT=test
# Another comment
GENERAL_VERSION=1.0.0`

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	config, err := InitConfig()

	// Should work since the valid lines can be parsed
	assert.NoError(t, err)
	assert.Equal(t, 8080, config.ServerPort)
	assert.Equal(t, "test", config.Environment)
	assert.Equal(t, "1.0.0", config.GeneralVersion)
}

func TestInitConfig_SpecialCharactersInEnvFile(t *testing.T) {
	// Create .env file with special characters (avoid problematic chars for env parsing)
	envContent := `SERVER_PORT=8080
GENERAL_VERSION=v1.0.0-测试版
ENVIRONMENT=test-with-hyphens
SECURITY_PEPPER="pepper-with-special-chars-safe"
SECURITY_JWT_SECRET="jwt-秘密-with-émojis-🔒"`

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	config, err := InitConfig()

	assert.NoError(t, err)
	assert.Equal(t, 8080, config.ServerPort)
	assert.Equal(t, "v1.0.0-测试版", config.GeneralVersion)
	assert.Equal(t, "test-with-hyphens", config.Environment)
	assert.Equal(t, "pepper-with-special-chars-safe", config.SecurityPepper)
	assert.Equal(t, "jwt-秘密-with-émojis-🔒", config.SecurityJwtSecret)
}

func TestInitConfig_DifferentDataTypes(t *testing.T) {
	// Test parsing of different data types
	envContent := `SERVER_PORT=65535
DB_CACHE_PORT=6379
SECURITY_SALT=16
GENERAL_VERSION=1.0.0
ENVIRONMENT=production
DB_PATH=/absolute/path/to/database.db
DB_CACHE_ADDRESS=redis.example.com
CORS_ALLOW_ORIGINS=https://app1.com,https://app2.com
SECURITY_PEPPER=long-pepper-value
SECURITY_JWT_SECRET=long-jwt-secret-value`

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	config, err := InitConfig()

	assert.NoError(t, err)

	// Test int fields
	assert.Equal(t, 65535, config.ServerPort)
	assert.Equal(t, 6379, config.DatabaseCachePort)
	assert.Equal(t, 16, config.SecuritySalt)

	// Test string fields
	assert.Equal(t, "1.0.0", config.GeneralVersion)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "/absolute/path/to/database.db", config.DatabaseDbPath)
	assert.Equal(t, "redis.example.com", config.DatabaseCacheAddress)
	assert.Equal(t, "https://app1.com,https://app2.com", config.CorsAllowOrigins)
	assert.Equal(t, "long-pepper-value", config.SecurityPepper)
	assert.Equal(t, "long-jwt-secret-value", config.SecurityJwtSecret)
}

func TestInitConfig_InvalidIntegerValues(t *testing.T) {
	testCases := []struct {
		name        string
		envContent  string
		expectError bool
	}{
		{
			name:        "NonNumericPort",
			envContent:  "SERVER_PORT=not-a-number",
			expectError: true, // viper should fail to unmarshal
		},
		{
			name:        "FloatPort",
			envContent:  "SERVER_PORT=8080.5",
			expectError: true, // viper should fail to unmarshal
		},
		{
			name:        "NonNumericCachePort",
			envContent:  "SERVER_PORT=8080\nDB_CACHE_PORT=invalid",
			expectError: true,
		},
		{
			name:        "NonNumericSalt",
			envContent:  "SERVER_PORT=8080\nSECURITY_SALT=invalid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envFile := createTempEnvFile(t, tc.envContent)
			defer func() { _ = os.Remove(envFile) }()

			// Change to the directory containing .env file
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() { _ = os.Chdir(originalDir) }()

			err = os.Chdir(filepath.Dir(envFile))
			require.NoError(t, err)

			_, err = InitConfig()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitConfig_LongStrings(t *testing.T) {
	longString := strings.Repeat("a", 1000)

	envContent := "SERVER_PORT=8080\n" +
		"GENERAL_VERSION=" + longString + "\n" +
		"ENVIRONMENT=" + longString + "\n" +
		"SECURITY_PEPPER=" + longString + "\n" +
		"SECURITY_JWT_SECRET=" + longString

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	config, err := InitConfig()

	assert.NoError(t, err)
	assert.Equal(t, 8080, config.ServerPort)
	assert.Equal(t, longString, config.GeneralVersion)
	assert.Equal(t, longString, config.Environment)
	assert.Equal(t, longString, config.SecurityPepper)
	assert.Equal(t, longString, config.SecurityJwtSecret)
}

func TestInitConfig_UpdatesGlobalConfigInstance(t *testing.T) {
	// Store original config
	originalConfig := ConfigInstance

	envContent := `SERVER_PORT=7654
ENVIRONMENT=init-test
GENERAL_VERSION=test-version`

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	config, err := InitConfig()

	assert.NoError(t, err)

	// Verify global ConfigInstance was updated
	assert.Equal(t, config, ConfigInstance)
	assert.Equal(t, 7654, ConfigInstance.ServerPort)
	assert.Equal(t, "init-test", ConfigInstance.Environment)
	assert.Equal(t, "test-version", ConfigInstance.GeneralVersion)

	// Verify it's different from original
	assert.NotEqual(t, originalConfig, ConfigInstance)
}

func TestInitConfig_WithComplexCORSOrigins(t *testing.T) {
	corsOrigins := "http://localhost:3000,https://app.example.com,https://api.example.com:8443,https://admin.example.com"

	envContent := "SERVER_PORT=8080\n" +
		"CORS_ALLOW_ORIGINS=" + corsOrigins

	envFile := createTempEnvFile(t, envContent)
	defer func() { _ = os.Remove(envFile) }()

	// Change to the directory containing .env file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(filepath.Dir(envFile))
	require.NoError(t, err)

	config, err := InitConfig()

	assert.NoError(t, err)
	assert.Equal(t, corsOrigins, config.CorsAllowOrigins)
}

// Helper functions

func clearEnvVars(t *testing.T) {
	// Clear all config-related environment variables
	envVars := []string{
		"GENERAL_VERSION", "ENVIRONMENT", "SERVER_PORT", "DB_PATH",
		"DB_CACHE_ADDRESS", "DB_CACHE_PORT", "CORS_ALLOW_ORIGINS",
		"SECURITY_SALT", "SECURITY_PEPPER", "SECURITY_JWT_SECRET",
	}
	for _, envVar := range envVars {
		_ = os.Unsetenv(envVar)
	}
}

func createTempEnvFile(t *testing.T, content string) string {
	tmpDir := createTempDir(t)
	envFile := filepath.Join(tmpDir, ".env")

	err := os.WriteFile(envFile, []byte(content), 0644)
	require.NoError(t, err)

	return envFile
}

func createTempDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	return tmpDir
}
