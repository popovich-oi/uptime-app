package database

import (
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"uptime-app/backend/internal/models"
)

// Open opens a PostgreSQL connection through GORM and verifies that the
// database is reachable before returning it to the application.
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return db, nil
}

// Migrate creates or updates the authentication tables managed by this
// service. It is intentionally kept in the database module so the server
// entrypoint does not know ORM or schema details.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.Monitor{})
}

// Close releases the underlying database/sql connection pool.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return closeSQLDB(sqlDB)
}

var closeSQLDB = func(db *sql.DB) error { return db.Close() }
