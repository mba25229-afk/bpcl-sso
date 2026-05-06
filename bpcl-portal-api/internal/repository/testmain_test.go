package repository_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/bpcl/portal-api/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dbURL := os.Getenv("BPCL_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://bpcl:bpcl@localhost:5433/bpcl_portal?sslmode=disable"
	}

	var err error
	testPool, err = repository.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("test: DB connect failed: %v", err)
	}

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}
