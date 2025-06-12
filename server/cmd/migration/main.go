package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"server/cmd/migration/seed"
	"server/config"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
	"strconv"

	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/gorm"
)

const (
	MIGRATION_PATH = "cmd/migration/migrations"
	MIGRATION_DB   = "sqlite3"
)

var MODELS_TO_MIGRATE = []any{
	&User{},
}

func main() {
	log := logger.New("migrations")
	log = log.Function("main")

	config, err := config.InitConfig()
	if err != nil {
		log.Er("failed to initialize config", err)
		os.Exit(1)
	}

	db, err := database.New(config)
	if err != nil {
		log.Er("failed to create database", err)
		os.Exit(1)
	}

	// Get flags from command line
	migrationType := "up"
	if len(os.Args) > 1 {
		migrationType = os.Args[1]
	}

	switch migrationType {
	case "up":
		err = migrateUp(db.SQL, config, log)
	case "down":
		steps := 1
		if len(os.Args) > 2 {
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil {
				log.Er("failed to parse step", err)
				os.Exit(1)
			}
		}
		err = migrateDown(steps, config, log)
	case "seed":
		err = migrateSeed(db.SQL, config, log)
	}

	if err != nil {
		log.Er("failed to run migrations", err)
		os.Exit(1)
	}

	log.Info("Migrations complete")
}

func migrateUp(db *gorm.DB, config config.Config, log logger.Logger) error {
	log = log.Function("migrateUp")
	log.Info("Running migrations up")

	err := runMigrations(config, log, migrate.Up)
	if err != nil {
		return log.Err("failed to run migrations", err)
	}

	err = autoMigrate(db, log)
	if err != nil {
		return log.Err("failed to auto migrate", err)
	}

	return nil
}

func migrateDown(steps int, config config.Config, log logger.Logger) error {
	log = log.Function("migrateDown")
	log.Info("Running migrations down")

	for range steps {
		err := runMigrations(config, log, migrate.Down)
		if err != nil {
			return log.Err("failed to run migrations", err)
		}
	}

	return nil
}

func migrateSeed(db *gorm.DB, config config.Config, log logger.Logger) error {
	log = log.Function("migrateSeed")
	log.Info("Running seed")

	// TODO: Clean DB to get to a new stat before seeding

	if err := migrateUp(db, config, log); err != nil {
		return log.Err("failed to auto migrate", err)
	}

	if err := autoMigrate(db, log); err != nil {
		return log.Err("failed to auto migrate", err)
	}

	log.Info("Seeding database")
	if err := seed.Seed(db, config, log); err != nil {
		return log.Err("failed to seed database", err)
	}

	return nil
}

func autoMigrate(db *gorm.DB, log logger.Logger) error {
	log = log.Function("autoMigrate")

	dbTables := MODELS_TO_MIGRATE

	log.Info("GORM auto-migrating tables", "tables", dbTables)
	err := db.AutoMigrate(dbTables...)
	if err != nil {
		return log.Err("failed to auto migrate", err)
	}

	return nil
}

func runMigrations(
	config config.Config,
	log logger.Logger,
	direction migrate.MigrationDirection,
) error {
	log = log.Function("runMigrations")

	migrations := &migrate.FileMigrationSource{
		Dir: MIGRATION_PATH,
	}

	filename := config.DatabaseDbPath

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return log.Err("failed to create database directory", err)
	}

	db, err := sql.Open(MIGRATION_DB, filename)
	if err != nil {
		return log.Err("failed to open database for migrations", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Er("failed to close database", err)
		}
	}()

	n, err := migrate.Exec(db, MIGRATION_DB, migrations, direction)
	if err != nil {
		return log.Err("failed to run migrations", err)
	}

	if n == 0 {
		log.Info("No migrations to apply")
	} else {
		log.Info("Applied migrations", "migrationCount", n)
	}

	return nil
}
