package config

import (
	"fmt"
	"server/internal/logger"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig_Success(t *testing.T) {
	log := logger.New("test")

	validConfig := Config{
		GeneralVersion:       "1.0.0",
		Environment:          "development",
		ServerPort:           8280,
		DatabaseDbPath:       "data/app.db",
		DatabaseCacheAddress: "localhost",
		DatabaseCachePort:    6379,
		CorsAllowOrigins:     "http://localhost:3010",
		SecuritySalt:         12,
		SecurityPepper:       "test-pepper-value",
		SecurityJwtSecret:    "test-jwt-secret-key",
	}

	err := validateConfig(validConfig, log)

	assert.NoError(t, err)
	// Verify the global config was set
	assert.Equal(t, validConfig, ConfigInstance)
}

func TestValidateConfig_MinimalValidConfig(t *testing.T) {
	log := logger.New("test")

	minimalConfig := Config{
		ServerPort: 1, // Minimum valid port
	}

	err := validateConfig(minimalConfig, log)

	assert.NoError(t, err)
	assert.Equal(t, minimalConfig, ConfigInstance)
}

func TestValidateConfig_TypicalProductionConfig(t *testing.T) {
	log := logger.New("test")

	prodConfig := Config{
		GeneralVersion:       "1.2.3",
		Environment:          "production",
		ServerPort:           80,
		DatabaseDbPath:       "/var/lib/app/database.db",
		DatabaseCacheAddress: "redis.example.com",
		DatabaseCachePort:    6379,
		CorsAllowOrigins:     "https://app.example.com,https://api.example.com",
		SecuritySalt:         16,
		SecurityPepper:       "xytcAjFSNIYlKE48UW1Rwub7iUspR3GVv85lWtfjNe0=",
		SecurityJwtSecret:    "super-secret-production-jwt-key-that-is-very-long",
	}

	err := validateConfig(prodConfig, log)

	assert.NoError(t, err)
	assert.Equal(t, prodConfig, ConfigInstance)
}

func TestValidateConfig_HighPortNumbers(t *testing.T) {
	log := logger.New("test")

	testCases := []struct {
		name string
		port int
	}{
		{"StandardHTTP", 8080},
		{"CustomHigh", 9999},
		{"VeryHigh", 65535}, // Maximum valid port
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				ServerPort: tc.port,
			}

			err := validateConfig(config, log)
			assert.NoError(t, err)
		})
	}
}

func TestValidateConfig_DifferentEnvironments(t *testing.T) {
	log := logger.New("test")

	environments := []string{
		"development",
		"staging",
		"production",
		"test",
		"local",
		"", // Empty environment should be valid
	}

	for _, env := range environments {
		t.Run("env_"+env, func(t *testing.T) {
			config := Config{
				Environment: env,
				ServerPort:  8080,
			}

			err := validateConfig(config, log)
			assert.NoError(t, err)
			assert.Equal(t, env, ConfigInstance.Environment)
		})
	}
}

func TestValidateConfig_DifferentDatabasePaths(t *testing.T) {
	log := logger.New("test")

	dbPaths := []string{
		"data/app.db",
		"/tmp/test.db",
		"./relative/path/db.sqlite",
		"/absolute/path/database.db",
		":memory:", // SQLite in-memory database
		"",         // Empty path should be handled by the application
	}

	for i, path := range dbPaths {
		t.Run("db_path_"+string(rune(i)), func(t *testing.T) {
			config := Config{
				DatabaseDbPath: path,
				ServerPort:     8080,
			}

			err := validateConfig(config, log)
			assert.NoError(t, err)
			assert.Equal(t, path, ConfigInstance.DatabaseDbPath)
		})
	}
}

func TestValidateConfig_CORSOrigins(t *testing.T) {
	log := logger.New("test")

	corsConfigs := []string{
		"http://localhost:3000",
		"https://app.example.com",
		"http://localhost:3000,https://app.example.com",
		"*", // Allow all origins
		"",  // No CORS origins
	}

	for i, cors := range corsConfigs {
		t.Run("cors_"+string(rune(i)), func(t *testing.T) {
			config := Config{
				CorsAllowOrigins: cors,
				ServerPort:       8080,
			}

			err := validateConfig(config, log)
			assert.NoError(t, err)
			assert.Equal(t, cors, ConfigInstance.CorsAllowOrigins)
		})
	}
}

func TestValidateConfig_SecurityConfiguration(t *testing.T) {
	log := logger.New("test")

	securityConfigs := []struct {
		name      string
		salt      int
		pepper    string
		jwtSecret string
	}{
		{
			name:      "StandardSecurity",
			salt:      12,
			pepper:    "standard-pepper-value",
			jwtSecret: "standard-jwt-secret",
		},
		{
			name:      "HighSecurity",
			salt:      16,
			pepper:    "very-long-pepper-value-with-special-chars-!@#$%^&*()",
			jwtSecret: "very-long-jwt-secret-with-special-characters-!@#$%^&*()",
		},
		{
			name:      "MinimalSecurity",
			salt:      1,
			pepper:    "p",
			jwtSecret: "j",
		},
	}

	for _, sc := range securityConfigs {
		t.Run(sc.name, func(t *testing.T) {
			config := Config{
				ServerPort:        8080,
				SecuritySalt:      sc.salt,
				SecurityPepper:    sc.pepper,
				SecurityJwtSecret: sc.jwtSecret,
			}

			err := validateConfig(config, log)
			assert.NoError(t, err)
			assert.Equal(t, sc.salt, ConfigInstance.SecuritySalt)
			assert.Equal(t, sc.pepper, ConfigInstance.SecurityPepper)
			assert.Equal(t, sc.jwtSecret, ConfigInstance.SecurityJwtSecret)
		})
	}
}

func TestGetConfig_ReturnsCurrentInstance(t *testing.T) {
	log := logger.New("test")

	// Set a specific config
	testConfig := Config{
		GeneralVersion: "test-version",
		ServerPort:     9999,
		Environment:    "test",
	}

	err := validateConfig(testConfig, log)
	require.NoError(t, err)

	// Test GetConfig returns the same instance
	retrievedConfig := GetConfig()

	assert.Equal(t, testConfig, retrievedConfig)
	assert.Equal(t, "test-version", retrievedConfig.GeneralVersion)
	assert.Equal(t, 9999, retrievedConfig.ServerPort)
	assert.Equal(t, "test", retrievedConfig.Environment)
}

func TestConfig_StructFieldTypes(t *testing.T) {
	// Test that we can create a config with all field types
	config := Config{
		GeneralVersion:       "string-value",
		Environment:          "another-string",
		ServerPort:           12345,        // int
		DatabaseDbPath:       "path/to/db", // string
		DatabaseCacheAddress: "cache-addr", // string
		DatabaseCachePort:    6379,         // int
		CorsAllowOrigins:     "origins",    // string
		SecuritySalt:         42,           // int
		SecurityPepper:       "pepper-str", // string
		SecurityJwtSecret:    "jwt-secret", // string
	}

	// Verify all fields are accessible and have correct types
	assert.IsType(t, "", config.GeneralVersion)
	assert.IsType(t, "", config.Environment)
	assert.IsType(t, 0, config.ServerPort)
	assert.IsType(t, "", config.DatabaseDbPath)
	assert.IsType(t, "", config.DatabaseCacheAddress)
	assert.IsType(t, 0, config.DatabaseCachePort)
	assert.IsType(t, "", config.CorsAllowOrigins)
	assert.IsType(t, 0, config.SecuritySalt)
	assert.IsType(t, "", config.SecurityPepper)
	assert.IsType(t, "", config.SecurityJwtSecret)
}

func TestConfig_DefaultZeroValues(t *testing.T) {
	// Test behavior with zero values
	var config Config

	assert.Equal(t, "", config.GeneralVersion)
	assert.Equal(t, "", config.Environment)
	assert.Equal(t, 0, config.ServerPort)
	assert.Equal(t, "", config.DatabaseDbPath)
	assert.Equal(t, "", config.DatabaseCacheAddress)
	assert.Equal(t, 0, config.DatabaseCachePort)
	assert.Equal(t, "", config.CorsAllowOrigins)
	assert.Equal(t, 0, config.SecuritySalt)
	assert.Equal(t, "", config.SecurityPepper)
	assert.Equal(t, "", config.SecurityJwtSecret)
}

// Negative Test Cases

func TestValidateConfig_InvalidServerPort_Zero(t *testing.T) {
	log := logger.New("test")

	invalidConfig := Config{
		ServerPort: 0, // Invalid: zero port
	}

	err := validateConfig(invalidConfig, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port: 0")
}

func TestValidateConfig_InvalidServerPort_Negative(t *testing.T) {
	log := logger.New("test")

	negativePortConfigs := []int{
		-1,
		-100,
		-8080,
		-65535,
	}

	for _, port := range negativePortConfigs {
		t.Run(fmt.Sprintf("port_%d", port), func(t *testing.T) {
			invalidConfig := Config{
				ServerPort: port,
			}

			err := validateConfig(invalidConfig, log)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("invalid port: %d", port))
		})
	}
}

func TestValidateConfig_PortBoundaryValues(t *testing.T) {
	log := logger.New("test")

	boundaryTests := []struct {
		name        string
		port        int
		shouldError bool
	}{
		{"NegativeOne", -1, true},
		{"Zero", 0, true},
		{"ValidOne", 1, false},
		{"ValidMax", 65535, false},
		{"OverMax", 65536, false},    // This might be valid in Go but could cause issues in real usage
		{"VeryLarge", 999999, false}, // Go allows this but OS might reject
	}

	for _, test := range boundaryTests {
		t.Run(test.name, func(t *testing.T) {
			config := Config{
				ServerPort: test.port,
			}

			err := validateConfig(config, log)

			if test.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfig_GlobalStateModification(t *testing.T) {
	log := logger.New("test")

	// Test that failed validation doesn't modify global state
	originalConfig := ConfigInstance

	invalidConfig := Config{
		ServerPort: -1, // Invalid
	}

	err := validateConfig(invalidConfig, log)

	assert.Error(t, err)
	// Global config should remain unchanged after failed validation
	assert.Equal(t, originalConfig, ConfigInstance)
}

func TestValidateConfig_EmptyConfigStruct(t *testing.T) {
	log := logger.New("test")

	emptyConfig := Config{} // All zero values

	err := validateConfig(emptyConfig, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port: 0")
}

func TestValidateConfig_PartiallyInvalidConfig(t *testing.T) {
	log := logger.New("test")

	// Config with some valid fields but invalid port
	partialConfig := Config{
		GeneralVersion:    "1.0.0",
		Environment:       "production",
		ServerPort:        -8080, // Invalid
		DatabaseDbPath:    "/valid/path/database.db",
		SecurityJwtSecret: "valid-secret",
	}

	err := validateConfig(partialConfig, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port: -8080")
}

func TestValidateConfig_SpecialPortValues(t *testing.T) {
	log := logger.New("test")

	specialPorts := []struct {
		name        string
		port        int
		description string
	}{
		{"ReservedHTTP", 80, "Standard HTTP port"},
		{"ReservedHTTPS", 443, "Standard HTTPS port"},
		{"ReservedSSH", 22, "SSH port"},
		{"ReservedFTP", 21, "FTP port"},
		{"PrivilegedLow", 1, "Lowest non-reserved port"},
		{"PrivilegedHigh", 1023, "Highest privileged port"},
		{"UnprivilegedLow", 1024, "Lowest unprivileged port"},
		{"EphemeralLow", 32768, "Start of ephemeral range"},
		{"EphemeralHigh", 65535, "End of port range"},
	}

	for _, sp := range specialPorts {
		t.Run(sp.name, func(t *testing.T) {
			config := Config{
				ServerPort: sp.port,
			}

			err := validateConfig(config, log)

			// All positive ports should be valid from validation perspective
			assert.NoError(t, err, "Port %d (%s) should be valid", sp.port, sp.description)
		})
	}
}

func TestGetConfig_AfterFailedValidation(t *testing.T) {
	log := logger.New("test")

	// Set a known good config first
	goodConfig := Config{
		ServerPort:  8080,
		Environment: "test",
	}
	err := validateConfig(goodConfig, log)
	require.NoError(t, err)

	// Verify GetConfig returns the good config
	retrieved := GetConfig()
	assert.Equal(t, goodConfig, retrieved)

	// Try to validate a bad config
	badConfig := Config{
		ServerPort: -1,
	}
	err = validateConfig(badConfig, log)
	assert.Error(t, err)

	// GetConfig should still return the previous good config
	stillRetrieved := GetConfig()
	assert.Equal(t, goodConfig, stillRetrieved)
	assert.NotEqual(t, badConfig, stillRetrieved)
}

// Edge Case Tests

func TestConfig_ExtremeStringValues(t *testing.T) {
	log := logger.New("test")

	// Test with extremely long strings
	veryLongString := strings.Repeat("a", 10000)

	config := Config{
		ServerPort:           8080, // Valid port
		GeneralVersion:       veryLongString,
		Environment:          veryLongString,
		DatabaseDbPath:       veryLongString,
		DatabaseCacheAddress: veryLongString,
		CorsAllowOrigins:     veryLongString,
		SecurityPepper:       veryLongString,
		SecurityJwtSecret:    veryLongString,
	}

	err := validateConfig(config, log)
	assert.NoError(t, err) // Should be valid despite long strings

	// Verify the config was stored
	assert.Equal(t, config, ConfigInstance)
}

func TestConfig_UnicodeAndSpecialCharacters(t *testing.T) {
	log := logger.New("test")

	unicodeConfig := Config{
		ServerPort:           8080,
		GeneralVersion:       "v1.0.0-测试版",
		Environment:          "тест", // Cyrillic
		DatabaseDbPath:       "/path/with/émojis/🚀/database.db",
		DatabaseCacheAddress: "café.example.com",
		CorsAllowOrigins:     "https://测试.example.com,https://тест.com",
		SecurityPepper:       "🔒secure🔑pepper🛡️",
		SecurityJwtSecret:    "jwt-secret-with-特殊字符",
	}

	err := validateConfig(unicodeConfig, log)
	assert.NoError(t, err)
	assert.Equal(t, unicodeConfig, ConfigInstance)
}

func TestConfig_ControlCharactersAndWhitespace(t *testing.T) {
	log := logger.New("test")

	// Test with various whitespace and control characters
	controlCharConfig := Config{
		ServerPort:           8080,
		GeneralVersion:       "v1.0.0\n\t\r",
		Environment:          " production ",
		DatabaseDbPath:       "/path/with\nnewlines/db.sqlite",
		DatabaseCacheAddress: "\t\tredis.example.com\t\t",
		CorsAllowOrigins:     "http://localhost:3000\n,https://app.com\r\n",
		SecurityPepper:       "pepper\x00with\x01control\x02chars",
		SecurityJwtSecret:    "jwt\tsecret\nwith\rwhitespace",
	}

	err := validateConfig(controlCharConfig, log)
	assert.NoError(t, err) // Validation should succeed
}

func TestConfig_NilLogger(t *testing.T) {
	// Test validateConfig with nil logger (edge case)
	config := Config{
		ServerPort: -1, // Invalid port
	}

	// This should panic or fail gracefully
	defer func() {
		if r := recover(); r != nil {
			// Panic is expected with nil logger
			assert.Contains(t, fmt.Sprintf("%v", r), "runtime error")
		}
	}()

	err := validateConfig(config, nil)

	// If no panic, should still return error
	assert.Error(t, err)
}

func TestConfig_ConcurrentValidation(t *testing.T) {
	log := logger.New("test")

	// Test concurrent validation calls
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			config := Config{
				ServerPort:  8080 + id, // Different ports
				Environment: fmt.Sprintf("test-%d", id),
			}
			err := validateConfig(config, log)
			results <- err
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Global config should be one of the valid configs
	finalConfig := GetConfig()
	assert.True(t, finalConfig.ServerPort >= 8080 && finalConfig.ServerPort < 8090)
}

func TestConfig_ValidationOrderDependency(t *testing.T) {
	log := logger.New("test")

	// Test that validation works regardless of field order
	configs := []Config{
		{ServerPort: 8080, Environment: "test1"},
		{Environment: "test2", ServerPort: 8081},
		{DatabaseDbPath: "/path", ServerPort: 8082, Environment: "test3"},
	}

	for i, config := range configs {
		t.Run(fmt.Sprintf("order_%d", i), func(t *testing.T) {
			err := validateConfig(config, log)
			assert.NoError(t, err)
			assert.Equal(t, config, GetConfig())
		})
	}
}
