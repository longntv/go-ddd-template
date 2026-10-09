//go:build integration

package datastore_test

import (
	"context"
	"log"
	"os"
	"testing"

	"gorm.io/gorm"

	"go-ddd-template/internal/testutil"
)

// readDB is shared by *_reader_test.go. Read queries don't affect each other,
// so reader test cases run in parallel against this one database.
//
// *_writer_test.go must not use it: writes would leak between cases. Each
// writer test case calls testutil.InitDB(t) for its own fresh copy instead.
var readDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize the template database (migrations + fixtures, once).
	tmplDB, err := testutil.InitTemplateDB(ctx)
	if err != nil {
		log.Fatalf("cannot init template db: %s", err)
	}

	var closeReadDB func() error
	readDB, _, closeReadDB, err = testutil.InitReadDB(ctx)
	if err != nil {
		log.Fatalf("cannot init read db: %s", err)
	}

	exitVal := m.Run()

	// Clean up the read db and the template db.
	if err := closeReadDB(); err != nil {
		log.Printf("close read db: %s", err)
	}
	if err := tmplDB.Release(ctx); err != nil {
		log.Printf("release template db: %s", err)
	}

	os.Exit(exitVal)
}
