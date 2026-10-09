package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/google/uuid"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
)

// TemplateDBConfig represents a config for creating TemplateDB.
type TemplateDBConfig struct {
	DBHost       string
	DBPort       int
	DBUser       string
	DBPass       string
	DBNamePrefix string

	FixturesDir   string
	MigrationsDir string
}

// TemplateDB is a database with the schema and fixtures applied once per
// test binary. Each test gets its own copy through Postgres'
// CREATE DATABASE ... TEMPLATE, which is much faster than re-running
// migrations and fixtures.
type TemplateDB struct {
	cfg *TemplateDBConfig

	adminDB *sql.DB // connection to the "postgres" maintenance database.
	dbName  string  // name of the template database.

	initialized bool
	mu          sync.Mutex

	nRef int // number of references to this template db.
}

// NewTemplateDB creates a TemplateDB. Call Init before use.
func NewTemplateDB(cfg *TemplateDBConfig) *TemplateDB {
	return &TemplateDB{cfg: cfg}
}

func (db *TemplateDB) dsn(dbName string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		db.cfg.DBHost, db.cfg.DBPort, db.cfg.DBUser, db.cfg.DBPass, dbName,
	)
}

// Init creates the template database, applies migrations and loads fixtures
// if it hasn't been initialized yet.
func (db *TemplateDB) Init(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.initialized {
		return nil
	}

	var err error
	db.adminDB, err = sql.Open("pgx", db.dsn("postgres"))
	if err != nil {
		return fmt.Errorf("open admin conn: %w", err)
	}
	if err := db.adminDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres at %s:%d (is `make docker/up` running?): %w", db.cfg.DBHost, db.cfg.DBPort, err)
	}

	id := uuid.New()
	db.dbName = fmt.Sprintf("%s_template_%x", db.cfg.DBNamePrefix, id[:])
	if _, err := db.adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %q", db.dbName)); err != nil {
		return fmt.Errorf("create template db: %w", err)
	}

	// The template must have no open connections when it is cloned, so the
	// setup connection is closed before returning.
	tmplConn, err := sql.Open("pgx", db.dsn(db.dbName))
	if err != nil {
		return fmt.Errorf("open template conn: %w", err)
	}
	defer tmplConn.Close()

	if err := applyMigrations(ctx, tmplConn, db.cfg.MigrationsDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	if err := loadFixtures(ctx, tmplConn, db.cfg.FixturesDir); err != nil {
		return fmt.Errorf("load fixtures: %w", err)
	}

	db.initialized = true

	return nil
}

// applyMigrations runs every *.up.sql file in dir in lexical order.
func applyMigrations(ctx context.Context, conn *sql.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no *.up.sql files in %s", dir)
	}
	sort.Strings(files)

	for _, file := range files {
		stmt, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		if _, err := conn.ExecContext(ctx, string(stmt)); err != nil {
			return fmt.Errorf("exec %s: %w", file, err)
		}
	}

	return nil
}

// NewTestDB creates a new test database cloned from the template database.
func (db *TemplateDB) NewTestDB(ctx context.Context) (*TestDB, error) {
	// Serialize clones: Postgres refuses to copy a template that another
	// CREATE DATABASE is reading at the same moment.
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.initialized {
		return nil, errors.New("db is not initialized")
	}

	id := uuid.New()
	dbName := fmt.Sprintf("%s_%x", db.cfg.DBNamePrefix, id[:])

	stmt := fmt.Sprintf("CREATE DATABASE %q TEMPLATE %q", dbName, db.dbName)
	if _, err := db.adminDB.ExecContext(ctx, stmt); err != nil {
		return nil, fmt.Errorf("clone template db: %w", err)
	}

	return &TestDB{
		dsn:     db.dsn(dbName),
		dbName:  dbName,
		adminDB: db.adminDB,
	}, nil
}

// cleanup drops the template database and closes the admin connection.
func (db *TemplateDB) cleanup(ctx context.Context) error {
	if !db.initialized {
		return nil
	}

	db.initialized = false

	if _, err := db.adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q WITH (FORCE)", db.dbName)); err != nil {
		return fmt.Errorf("drop template db: %w", err)
	}

	if err := db.adminDB.Close(); err != nil {
		return fmt.Errorf("admin conn close: %w", err)
	}

	return nil
}

// Acquire acquires a reference to the template database.
func (db *TemplateDB) Acquire() (*TemplateDB, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.initialized {
		return nil, errors.New("template db is not initialized")
	}

	db.nRef++
	log.Printf("acquire template db, nRef = %d", db.nRef)

	return db, nil
}

// Release releases a reference to the template database and drops it when
// the last reference is released.
func (db *TemplateDB) Release(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.initialized {
		return errors.New("template db is not initialized")
	}

	if db.nRef == 0 {
		return nil
	}

	db.nRef--
	log.Printf("release template db, nRef = %d", db.nRef)
	if db.nRef > 0 {
		return nil
	}

	return db.cleanup(ctx)
}
