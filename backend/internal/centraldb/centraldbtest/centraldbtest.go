// Package centraldbtest gives integration tests an isolated central database.
package centraldbtest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"smlcloudplatform/internal/centraldb"
)

// DSN returns the throwaway PostgreSQL used by integration tests.
func DSN() string {
	for _, key := range []string{"BC_GL_TEST_POSTGRES_DSN", "GL_AUTH_TEST_DSN"} {
		if dsn := os.Getenv(key); dsn != "" {
			return dsn
		}
	}
	return ""
}

// New creates a fresh database with the central schema; it is dropped when the test ends.
func New(t testing.TB) *sql.DB {
	t.Helper()
	dsn := DSN()
	if dsn == "" {
		t.Skip("BC_GL_TEST_POSTGRES_DSN not set")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open admin database: %v", err)
	}
	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}
	name := "central_test_" + hex.EncodeToString(suffix)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := admin.ExecContext(ctx, `CREATE DATABASE `+name); err != nil {
		admin.Close()
		t.Fatalf("create test database: %v", err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse test DSN: %v", err)
	}
	parsed.Path = "/" + name
	db, err := sql.Open("postgres", parsed.String())
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer dropCancel()
		_, _ = admin.ExecContext(dropCtx, `DROP DATABASE IF EXISTS `+name+` WITH (FORCE)`)
		admin.Close()
	})
	if err := centraldb.EnsureSchema(ctx, db); err != nil {
		t.Fatalf("ensure central schema: %v", err)
	}
	return db
}

// Exec runs fixture statements and fails the test on error.
func Exec(t testing.TB, db *sql.DB, statement string, args ...interface{}) {
	t.Helper()
	if _, err := db.Exec(statement, args...); err != nil {
		t.Fatalf("fixture %q: %v", statement, err)
	}
}
