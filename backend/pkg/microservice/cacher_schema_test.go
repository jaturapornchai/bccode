package microservice

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// Several services (and parallel test packages) open the cache on the same fresh database at
// once; CREATE TABLE IF NOT EXISTS alone races there ("pg_type_typname_nsp_index").
func TestNewCacherConcurrentOnFreshSchema(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("BC_GL_TEST_POSTGRES_DSN not set")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { admin.Close() })
	ns := fmt.Sprintf("cacher_race_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + ns); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() { _, _ = admin.Exec("DROP SCHEMA " + ns + " CASCADE") })
	// lib/pq passes unknown DSN keys as run-time parameters, so every pooled connection uses ns
	db, err := sql.Open("postgres", dsn+"&search_path="+ns)
	if err != nil {
		t.Fatalf("open schema db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	const openers = 12
	errs := make(chan error, openers)
	var start, done sync.WaitGroup
	start.Add(1)
	for range openers {
		done.Add(1)
		go func() {
			defer done.Done()
			start.Wait()
			_, err := NewCacher(db)
			errs <- err
		}()
	}
	start.Done()
	done.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent NewCacher: %v", err)
		}
	}
}
