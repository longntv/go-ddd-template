// Package testutil provides database helpers for integration tests.
//
// Lifecycle (see internal/infrastructure/datastore/main_test.go):
//
//	TestMain     -> InitTemplateDB: migrations + fixtures applied once per package
//	reader tests -> InitReadDB: one shared clone, safe for parallel read-only tests
//	writer tests -> InitDB(t): a fresh clone per test, dropped on t.Cleanup
package testutil

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"gorm.io/gorm"
)

// getTemplateDB initializes a TemplateDB instance using the configuration
// from the environment and returns it.
var getTemplateDB = sync.OnceValues(func() (*TemplateDB, error) {
	_, currentFilename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("runtime.Caller error")
	}
	root := filepath.Join(filepath.Dir(currentFilename), "../..")

	env := LoadEnv()
	return NewTemplateDB(&TemplateDBConfig{
		DBHost:       env.DBHost,
		DBPort:       env.DBPort,
		DBUser:       env.DBUser,
		DBPass:       env.DBPass,
		DBNamePrefix: env.DBName,

		FixturesDir:   filepath.Join(root, "testdata/fixtures"),
		MigrationsDir: filepath.Join(root, "database/migrations"),
	}), nil
})

// InitDB creates a fresh test database cloned from the template and returns
// a [gorm.DB] for it along with its name. The database is dropped when the
// test finishes.
func InitDB(t *testing.T) (*gorm.DB, string) {
	t.Helper()

	tmplDB, err := getTemplateDB()
	if err != nil {
		t.Fatalf("get template db: %s", err)
	}

	if err := tmplDB.Init(context.Background()); err != nil {
		t.Fatalf("init template db: %s", err)
	}

	db, err := tmplDB.NewTestDB(context.Background())
	if err != nil {
		t.Fatalf("new test db: %s", err)
	}

	gormDB, err := db.Open()
	if err != nil {
		t.Fatalf("new gorm db: %s", err)
	}

	t.Cleanup(func() {
		sqlDB, err := gormDB.DB()
		if err != nil {
			t.Errorf("get gorm underlying db: %s", err)
			return
		}

		if err := sqlDB.Close(); err != nil {
			t.Errorf("close gorm underlying db: %s", err)
		}

		if err := db.Cleanup(context.Background()); err != nil {
			t.Errorf("clean up db: %s", err)
		}
	})

	return gormDB, db.DBName()
}

// InitReadDB creates a test database meant to be shared by read-only tests
// and returns it with a function that closes and drops it.
func InitReadDB(ctx context.Context) (gormDB *gorm.DB, dbName string, dbCloseFunc func() error, err error) {
	tmplDB, err := getTemplateDB()
	if err != nil {
		return nil, "", nil, fmt.Errorf("get template db: %w", err)
	}

	if err := tmplDB.Init(ctx); err != nil {
		return nil, "", nil, fmt.Errorf("init template db: %w", err)
	}

	db, err := tmplDB.NewTestDB(ctx)
	if err != nil {
		return nil, "", nil, fmt.Errorf("new test db: %w", err)
	}

	gormDB, err = db.Open()
	if err != nil {
		return nil, "", nil, fmt.Errorf("new gorm db: %w", err)
	}

	dbCloseFunc = func() error {
		sqlDB, err := gormDB.DB()
		if err != nil {
			return fmt.Errorf("get gorm underlying db: %w", err)
		}

		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("close gorm underlying db: %w", err)
		}

		if err := db.Cleanup(ctx); err != nil {
			return fmt.Errorf("clean up db: %w", err)
		}

		return nil
	}

	return gormDB, db.DBName(), dbCloseFunc, nil
}

// InitTemplateDB initializes the template database and acquires a reference
// to it. Call Release on the result when the test binary finishes.
func InitTemplateDB(ctx context.Context) (*TemplateDB, error) {
	tmplDB, err := getTemplateDB()
	if err != nil {
		return nil, fmt.Errorf("get template db: %w", err)
	}

	if err := tmplDB.Init(ctx); err != nil {
		return nil, fmt.Errorf("init template db: %w", err)
	}

	return tmplDB.Acquire()
}
