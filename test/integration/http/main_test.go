//go:build integration

package http_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/longntv/go-ddd-template/internal/testutil"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Initialize the template database (migrations + fixtures, once).
	tmplDB, err := testutil.InitTemplateDB(context.Background())
	if err != nil {
		log.Fatalf("cannot init template db: %s", err)
	}

	exitVal := m.Run()

	if err := tmplDB.Release(context.Background()); err != nil {
		log.Printf("release template db: %s", err)
	}

	os.Exit(exitVal)
}
