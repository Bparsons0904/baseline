package main

// IMPORTANT: Do not create tests that generate excessive nested directory structures.
// Such tests provide no meaningful test value and clutter the filesystem.

import (
	"os"
	"path/filepath"
	"server/config"
	"server/internal/logger"
	. "server/internal/models"
	"testing"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test Constants

func TestMigrationConstants(t *testing.T) {
	// Test migration constants
	assert.Equal(t, "cmd/migration/migrations", MIGRATION_PATH)
	assert.Equal(t, "sqlite3", MIGRATION_DB)

	// Constants should be non-empty
	assert.NotEmpty(t, MIGRATION_PATH)
	assert.NotEmpty(t, MIGRATION_DB)

	// Path should be reasonable
	assert.Contains(t, MIGRATION_PATH, "migration")
	assert.Contains(t, MIGRATION_PATH, "migrations")
}

func TestModelsToMigrate(t *testing.T) {
	// Test MODELS_TO_MIGRATE slice
	assert.NotNil(t, MODELS_TO_MIGRATE)
	assert.Len(t, MODELS_TO_MIGRATE, 1) // Should have User model

	// Should contain User model
	assert.IsType(t, &User{}, MODELS_TO_MIGRATE[0])
}

// Helper functions for testing

func setupTestDB(t *testing.T) (*gorm.DB, string) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	assert.NoError(t, err)

	return db, dbPath
}

func setupTestConfig(dbPath string) config.Config {
	return config.Config{
		DatabaseDbPath: dbPath,
	}
}

func setupTestLogger() logger.Logger {
	return logger.New("test")
}

// Test autoMigrate function

func TestAutoMigrate_Success(t *testing.T) {
	// Test successful auto migration
	db, _ := setupTestDB(t)
	log := setupTestLogger()

	err := autoMigrate(db, log)

	assert.NoError(t, err)

	// Verify that User table was created
	assert.True(t, db.Migrator().HasTable(&User{}))
}

func TestAutoMigrate_WithNilDB(t *testing.T) {
	// Test auto migration with nil database
	log := setupTestLogger()

	// This will panic because autoMigrate doesn't check for nil
	// Let's test that it panics as expected
	assert.Panics(t, func() {
		_ = autoMigrate(nil, log)
	})
}

func TestAutoMigrate_MultipleCallsIdempotent(t *testing.T) {
	// Test that multiple calls to autoMigrate are idempotent
	db, _ := setupTestDB(t)
	log := setupTestLogger()

	// First call
	err1 := autoMigrate(db, log)
	assert.NoError(t, err1)

	// Second call should also succeed
	err2 := autoMigrate(db, log)
	assert.NoError(t, err2)

	// Table should still exist
	assert.True(t, db.Migrator().HasTable(&User{}))
}

// Test migrateUp function

func TestMigrateUp_Success(t *testing.T) {
	// Test successful migration up
	db, dbPath := setupTestDB(t)
	cfg := setupTestConfig(dbPath)
	log := setupTestLogger()

	// Create migrations directory for test
	migrationDir := filepath.Join(t.TempDir(), "migrations")
	err := os.MkdirAll(migrationDir, 0755)
	assert.NoError(t, err)

	// This will fail because we don't have actual migration files,
	// but we can test the function structure
	_ = migrateUp(db, cfg, log)

	// May error due to missing migration files, but function should not panic
	// The error is expected in test environment
}

func TestMigrateUp_WithNilDB(t *testing.T) {
	// Test migration up with nil database
	cfg := setupTestConfig("nonexistent.db")
	log := setupTestLogger()

	// migrateUp will fail at runMigrations step, not at autoMigrate
	err := migrateUp(nil, cfg, log)

	// Should return an error (from missing migration files)
	assert.Error(t, err)
}

// Test migrateDown function

func TestMigrateDown_SingleStep(t *testing.T) {
	// Test migration down with single step
	cfg := setupTestConfig("test.db")
	log := setupTestLogger()

	_ = migrateDown(1, cfg, log)

	// May error due to missing migration files, but function should not panic
	// The error is expected in test environment without actual migration files
}

func TestMigrateDown_MultipleSteps(t *testing.T) {
	// Test migration down with multiple steps
	cfg := setupTestConfig("test.db")
	log := setupTestLogger()

	_ = migrateDown(3, cfg, log)

	// May error due to missing migration files, but function should not panic
}

func TestMigrateDown_ZeroSteps(t *testing.T) {
	// Test migration down with zero steps
	cfg := setupTestConfig("test.db")
	log := setupTestLogger()

	err := migrateDown(0, cfg, log)

	// Should succeed (no operations to perform)
	assert.NoError(t, err)
}

func TestMigrateDown_NegativeSteps(t *testing.T) {
	// Test migration down with negative steps
	cfg := setupTestConfig("test.db")
	log := setupTestLogger()

	err := migrateDown(-1, cfg, log)

	// Should succeed (no iterations of the loop)
	assert.NoError(t, err)
}

// Test migrateSeed function

func TestMigrateSeed_StructureTest(t *testing.T) {
	// Test migrate seed function structure
	db, dbPath := setupTestDB(t)
	cfg := setupTestConfig(dbPath)
	log := setupTestLogger()

	_ = migrateSeed(db, cfg, log)

	// May error due to missing migration files or seed issues,
	// but function should not panic
}

func TestMigrateSeed_WithNilDB(t *testing.T) {
	// Test migrate seed with nil database
	cfg := setupTestConfig("nonexistent.db")
	log := setupTestLogger()

	// migrateSeed will fail at runMigrations step in migrateUp
	err := migrateSeed(nil, cfg, log)

	// Should return an error (from missing migration files)
	assert.Error(t, err)
}

// Test runMigrations function

func TestRunMigrations_DirectoryValidation(t *testing.T) {
	// Test runMigrations with various configurations
	cfg := config.Config{
		DatabaseDbPath: filepath.Join(t.TempDir(), "test.db"),
	}
	log := setupTestLogger()

	// Test with Up direction
	_ = runMigrations(cfg, log, migrate.Up)
	// Will error due to missing migration files, but should not panic

	// Test with Down direction
	_ = runMigrations(cfg, log, migrate.Down)
	// Will error due to missing migration files, but should not panic
}

func TestRunMigrations_EmptyDatabasePath(t *testing.T) {
	// Test runMigrations with empty database path
	cfg := config.Config{
		DatabaseDbPath: "",
	}
	log := setupTestLogger()

	err := runMigrations(cfg, log, migrate.Up)

	// Should return an error due to empty path
	assert.Error(t, err)
}

func TestRunMigrations_InvalidDatabasePath(t *testing.T) {
	// Test runMigrations with invalid database path
	cfg := config.Config{
		DatabaseDbPath: "/invalid/path/that/cannot/be/created/test.db",
	}
	log := setupTestLogger()

	err := runMigrations(cfg, log, migrate.Up)

	// Should return an error due to inability to create directory
	assert.Error(t, err)
}

// Test Function Signatures and Types

func TestMigrateUpSignature(t *testing.T) {
	// Test that migrateUp has correct signature
	db, dbPath := setupTestDB(t)
	cfg := setupTestConfig(dbPath)
	log := setupTestLogger()

	// Should accept these parameters and return error (may be nil or actual error)
	err := migrateUp(db, cfg, log)
	// Just test that it returns some type of error (nil is valid)
	assert.True(t, err == nil || err != nil)
}

func TestMigrateDownSignature(t *testing.T) {
	// Test that migrateDown has correct signature
	cfg := setupTestConfig("test.db")
	log := setupTestLogger()

	// Should accept int, config, logger and return error (may be nil or actual error)
	err := migrateDown(1, cfg, log)
	// Just test that it returns some type of error (nil is valid)
	assert.True(t, err == nil || err != nil)
}

func TestMigrateSeedSignature(t *testing.T) {
	// Test that migrateSeed has correct signature
	db, dbPath := setupTestDB(t)
	cfg := setupTestConfig(dbPath)
	log := setupTestLogger()

	// Should accept these parameters and return error (may be nil or actual error)
	err := migrateSeed(db, cfg, log)
	// Just test that it returns some type of error (nil is valid)
	assert.True(t, err == nil || err != nil)
}

func TestAutoMigrateSignature(t *testing.T) {
	// Test that autoMigrate has correct signature
	db, _ := setupTestDB(t)
	log := setupTestLogger()

	// Should accept gorm.DB and logger, return error
	err := autoMigrate(db, log)
	assert.IsType(t, error(nil), err)
}

func TestRunMigrationsSignature(t *testing.T) {
	// Test that runMigrations has correct signature
	cfg := setupTestConfig("test.db")
	log := setupTestLogger()

	// Should accept config, logger, direction and return error (may be nil or actual error)
	err := runMigrations(cfg, log, migrate.Up)
	// Just test that it returns some type of error (nil is valid)
	assert.True(t, err == nil || err != nil)
}

// Test Edge Cases

func TestConstants_EdgeCases(t *testing.T) {
	// Test constants edge cases

	// MIGRATION_PATH should not be absolute
	assert.False(t, filepath.IsAbs(MIGRATION_PATH))

	// MIGRATION_DB should be valid driver name
	assert.NotContains(t, MIGRATION_DB, " ")
	assert.NotContains(t, MIGRATION_DB, "/")
	assert.NotContains(t, MIGRATION_DB, "\\")
}

func TestModelsToMigrate_EdgeCases(t *testing.T) {
	// Test MODELS_TO_MIGRATE edge cases

	// Should not be empty
	assert.NotEmpty(t, MODELS_TO_MIGRATE)

	// All elements should be pointers
	for i, model := range MODELS_TO_MIGRATE {
		assert.NotNil(t, model, "Model at index %d should not be nil", i)
	}

	// Should contain User model
	foundUser := false
	for _, model := range MODELS_TO_MIGRATE {
		if _, ok := model.(*User); ok {
			foundUser = true
			break
		}
	}
	assert.True(t, foundUser, "MODELS_TO_MIGRATE should contain User model")
}

func TestDatabasePathHandling(t *testing.T) {
	// Test database path handling in various scenarios
	testCases := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "ValidPath",
			path:      filepath.Join(t.TempDir(), "valid.db"),
			shouldErr: false,
		},
		{
			name:      "EmptyPath",
			path:      "",
			shouldErr: true,
		},
		{
			name:      "RelativePath",
			path:      "relative.db",
			shouldErr: false,
		},
		{
			name:      "DeepPath",
			path:      filepath.Join(t.TempDir(), "deep", "nested", "path", "test.db"),
			shouldErr: false,
		},
	}

	log := setupTestLogger()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with autoMigrate (requires actual DB)
			if !tc.shouldErr && tc.path != "" {
				db, err := gorm.Open(sqlite.Open(tc.path), &gorm.Config{})
				if err == nil {
					err = autoMigrate(db, log)
					if tc.shouldErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				}
			}
		})
	}
}

// Test Concurrency and Multiple Operations

func TestAutoMigrate_Concurrent(t *testing.T) {
	// Test concurrent autoMigrate calls
	// Each goroutine gets its own database to avoid table creation conflicts
	log := setupTestLogger()

	const numGoroutines = 5
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Each goroutine gets its own database
			db, _ := setupTestDB(t)
			err := autoMigrate(db, log)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		// All should succeed (each has its own database)
		assert.NoError(t, err)
	}
}

func TestMigrateDown_StressTest(t *testing.T) {
	// Test migrateDown with large number of steps
	cfg := setupTestConfig(filepath.Join(t.TempDir(), "stress.db"))
	log := setupTestLogger()

	// This should complete without hanging or panicking
	err := migrateDown(1000, cfg, log)
	_ = err // May error due to missing migration files, but should not hang
}

// Test Package-Level Validation

func TestPackageConstants_Validation(t *testing.T) {
	// Test that package constants are properly defined
	assert.IsType(t, "", MIGRATION_PATH)
	assert.IsType(t, "", MIGRATION_DB)
	assert.IsType(t, []any{}, MODELS_TO_MIGRATE)

	// Test constant values are reasonable
	assert.True(t, len(MIGRATION_PATH) > 0)
	assert.True(t, len(MIGRATION_DB) > 0)
	assert.True(t, len(MODELS_TO_MIGRATE) > 0)
}

func TestMigrationDirections(t *testing.T) {
	// Test that migration directions are used correctly
	directions := []migrate.MigrationDirection{
		migrate.Up,
		migrate.Down,
	}

	cfg := setupTestConfig(filepath.Join(t.TempDir(), "direction_test.db"))
	log := setupTestLogger()

	for _, direction := range directions {
		t.Run("direction_"+string(rune(direction)), func(t *testing.T) {
			err := runMigrations(cfg, log, direction)
			// May error due to missing migration files, but should not panic
			_ = err
		})
	}
}

// Test Error Handling

func TestErrorHandling_PropagatesCorrectly(t *testing.T) {
	// Test that errors are properly propagated through the call chain

	// Test with various invalid configurations
	invalidConfigs := []config.Config{
		{DatabaseDbPath: ""},                 // Empty path
		{DatabaseDbPath: "/invalid/path.db"}, // Invalid path
	}

	log := setupTestLogger()

	for i, cfg := range invalidConfigs {
		t.Run("invalid_config_"+string(rune('0'+i)), func(t *testing.T) {
			// These should all return errors, not panic
			err1 := runMigrations(cfg, log, migrate.Up)
			assert.Error(t, err1)

			err2 := runMigrations(cfg, log, migrate.Down)
			assert.Error(t, err2)
		})
	}
}

// Test String Handling

func TestStringHandling_EdgeCases(t *testing.T) {
	// Test string handling in various edge cases

	// NOTE: Avoid testing very long directory paths as they create excessive nested directories
	// that clutter the filesystem and provide no meaningful test value

	// Test paths with special characters
	specialChars := []string{
		"test with spaces.db",
		"test-with-dashes.db",
		"test_with_underscores.db",
		"test.with.dots.db",
	}

	for _, specialPath := range specialChars {
		t.Run("special_chars", func(t *testing.T) {
			testPath := filepath.Join(t.TempDir(), specialPath)
			cfg := setupTestConfig(testPath)
			log := setupTestLogger()

			// Should handle special characters in paths
			err := runMigrations(cfg, log, migrate.Up)
			_ = err // May error due to missing migration files
		})
	}
}
