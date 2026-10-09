package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB represents a database for testing, it's designed for running with
// a specific testcase.
type TestDB struct {
	dsn     string
	dbName  string
	adminDB *sql.DB // shared with the TemplateDB; used only to drop this db.

	cleanedUp bool
	mu        sync.RWMutex
}

// Open establishes a GORM database connection to the test database.
func (db *TestDB) Open() (*gorm.DB, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.cleanedUp {
		return nil, errors.New("db already cleaned up")
	}

	return gorm.Open(
		postgres.Open(db.dsn),
		&gorm.Config{
			SkipDefaultTransaction: true,
			Logger:                 logger.Default.LogMode(logger.Silent),
		},
	)
}

// Cleanup drops the test database.
func (db *TestDB) Cleanup(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.cleanedUp {
		return nil
	}

	db.cleanedUp = true

	if _, err := db.adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q WITH (FORCE)", db.dbName)); err != nil {
		return fmt.Errorf("drop db: %w", err)
	}

	return nil
}

// DBName returns the name of the test database.
func (db *TestDB) DBName() string {
	return db.dbName
}
