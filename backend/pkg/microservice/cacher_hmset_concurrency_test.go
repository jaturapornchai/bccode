package microservice

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Two select-holding calls on one session write the same hash at the same time (the workspace
// holdings BFF does this when a screen mounts twice). Rows must be locked in one fixed order,
// otherwise PostgreSQL aborts one writer as a deadlock after deadlock_timeout (1s) and the user
// gets 500/401 (found by `npm run uat`, 2026-10-11).
func TestCacherHMSetConcurrentSameKeyNoDeadlock(t *testing.T) {
	cacher := newTestPostgresCacher(t)
	key := fmt.Sprintf("test:hmset-concurrency:%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = cacher.Del(key) })

	fields := map[string]interface{}{}
	for i := 0; i < 12; i++ {
		fields[fmt.Sprintf("field%02d", i)] = "value"
	}
	if err := cacher.HMSet(key, fields); err != nil {
		t.Fatal(err)
	}

	const writers, rounds = 4, 15
	errs := make(chan error, writers*rounds)
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				if err := cacher.HMSet(key, fields); err != nil {
					errs <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent HMSet on one key failed: %v", err)
	}
}

// Every authenticated request extends its session with Expire while select-holding rewrites the same
// hash with HMSet. Expire must lock rows in the same field order as HMSet; an UPDATE in heap order
// deadlocked with it, PostgreSQL cancelled the Expire after 1s and the request answered 401
// (`npm run uat`, /user screen, 2026-10-11).
func TestCacherExpireConcurrentWithHMSetNoDeadlock(t *testing.T) {
	cacher := newTestPostgresCacher(t)
	key := fmt.Sprintf("test:expire-concurrency:%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = cacher.Del(key) })

	fields := map[string]interface{}{}
	// Insert one field at a time in descending order so heap order is the reverse of HMSet's lock order.
	for i := 11; i >= 0; i-- {
		name := fmt.Sprintf("field%02d", i)
		fields[name] = "value"
		if err := cacher.HMSet(key, map[string]interface{}{name: "value"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := cacher.Expire(key, time.Hour); err != nil {
		t.Fatal(err)
	}

	const workers, rounds = 4, 25
	errs := make(chan error, 2*workers*rounds)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				if err := cacher.HMSet(key, fields); err != nil {
					errs <- fmt.Errorf("HMSet: %w", err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				if err := cacher.Expire(key, time.Hour); err != nil {
					errs <- fmt.Errorf("Expire: %w", err)
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent HMSet/Expire on one key failed: %v", err)
	}
}
