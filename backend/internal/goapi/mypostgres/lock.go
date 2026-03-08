package mypostgres

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// DistributedLock - โครงสร้าง Distributed Lock ใน PostgreSQL
// แทนที่ Redis SET NX with expiry
type DistributedLock struct {
	db       *sql.DB
	key      string
	owner    string
	expiry   time.Duration
	acquired bool
}

var (
	memoryLockMu             sync.Mutex
	memoryLocks              = make(map[string]*memoryLockEntry)
	distributedLocksDisabled atomic.Bool
)

type memoryLockEntry struct {
	owner      string
	acquiredAt time.Time
	expiresAt  time.Time
}

func usingMemoryFallback() bool {
	return distributedLocksDisabled.Load()
}

func switchToMemoryFallback(err error) {
	if distributedLocksDisabled.CompareAndSwap(false, true) {
		logger.Warn("distributed_locks table unavailable, switching to in-memory fallback: %v", err)
	}
}

func isMissingLockTable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "distributed_locks") && strings.Contains(msg, "does not exist")
}

func cleanupExpiredMemoryLocks() int64 {
	memoryLockMu.Lock()
	defer memoryLockMu.Unlock()
	return cleanupExpiredMemoryLocksLocked()
}

func cleanupExpiredMemoryLocksLocked() int64 {
	now := time.Now()
	var removed int64
	for key, entry := range memoryLocks {
		if entry.expiresAt.Before(now) {
			delete(memoryLocks, key)
			removed++
		}
	}

	if removed > 0 {
		logger.Debug("Cleaned up %d expired in-memory locks", removed)
	}

	return removed
}

func (l *DistributedLock) acquireMemoryLock() error {
	memoryLockMu.Lock()
	defer memoryLockMu.Unlock()

	cleanupExpiredMemoryLocksLocked()

	if entry, exists := memoryLocks[l.key]; exists {
		if entry.expiresAt.After(time.Now()) {
			return fmt.Errorf("lock is already held by another process")
		}
	}

	memoryLocks[l.key] = &memoryLockEntry{
		owner:      l.owner,
		acquiredAt: time.Now(),
		expiresAt:  time.Now().Add(l.expiry),
	}
	l.acquired = true
	logger.Debug("In-memory lock acquired: key=%s, owner=%s", l.key, l.owner)
	return nil
}

func (l *DistributedLock) releaseMemoryLock() error {
	memoryLockMu.Lock()
	defer memoryLockMu.Unlock()

	entry, exists := memoryLocks[l.key]
	if !exists {
		return fmt.Errorf("lock was not held by this process")
	}
	if entry.owner != l.owner {
		return fmt.Errorf("lock was not held by this process")
	}
	delete(memoryLocks, l.key)
	l.acquired = false
	logger.Debug("In-memory lock released: key=%s, owner=%s", l.key, l.owner)
	return nil
}

func (l *DistributedLock) extendMemoryLock(additionalTime time.Duration) error {
	memoryLockMu.Lock()
	defer memoryLockMu.Unlock()

	entry, exists := memoryLocks[l.key]
	if !exists || entry.owner != l.owner {
		return fmt.Errorf("lock was not held by this process")
	}
	entry.expiresAt = time.Now().Add(additionalTime)
	return nil
}

// NewDistributedLock - สร้าง Distributed Lock ใหม่
// แทนที่ myredis.NewDistributedLock
func NewDistributedLock(db *sql.DB, key string, expiry time.Duration) *DistributedLock {
	return &DistributedLock{
		db:     db,
		key:    key,
		owner:  fmt.Sprintf("%d", time.Now().UnixNano()), // Unique owner ID
		expiry: expiry,
	}
}

// Acquire - ขอ lock (non-blocking)
// แทนที่ Redis SET NX
func (l *DistributedLock) Acquire(ctx context.Context) error {
	if usingMemoryFallback() {
		return l.acquireMemoryLock()
	}

	if l.db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// ทำความสะอาด expired locks ก่อน
	l.cleanupExpiredLocks(ctx)

	query := `
		INSERT INTO distributed_locks (lock_key, owner, acquired_at, expires_at)
		VALUES ($1, $2, NOW(), NOW() + $3::interval)
		ON CONFLICT (lock_key) DO NOTHING
		RETURNING lock_key
	`

	expiryInterval := fmt.Sprintf("%d seconds", int(l.expiry.Seconds()))

	var lockKey string
	err := l.db.QueryRowContext(ctx, query, l.key, l.owner, expiryInterval).Scan(&lockKey)

	if err == sql.ErrNoRows {
		return fmt.Errorf("lock is already held by another process")
	}

	if err != nil {
		if isMissingLockTable(err) {
			switchToMemoryFallback(err)
			return l.acquireMemoryLock()
		}
		logger.Error("Failed to acquire lock: %v", err)
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	l.acquired = true
	logger.Debug("Lock acquired: key=%s, owner=%s, expiry=%v", l.key, l.owner, l.expiry)
	return nil
}

// AcquireWithRetry - ขอ lock พร้อม retry
// แทนที่ myredis.DistributedLock.AcquireWithRetry
func (l *DistributedLock) AcquireWithRetry(ctx context.Context, maxRetries int, retryDelay time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		err := l.Acquire(ctx)
		if err == nil {
			return nil
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
			// Continue to next retry
		}
	}

	return fmt.Errorf("failed to acquire lock after %d retries", maxRetries)
}

// Release - ปล่อย lock
// แทนที่ Redis DEL with Lua script
func (l *DistributedLock) Release(ctx context.Context) error {
	if !l.acquired {
		return nil
	}

	if usingMemoryFallback() {
		return l.releaseMemoryLock()
	}

	if l.db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		DELETE FROM distributed_locks
		WHERE lock_key = $1 AND owner = $2
		RETURNING lock_key
	`

	var lockKey string
	err := l.db.QueryRowContext(ctx, query, l.key, l.owner).Scan(&lockKey)

	if err == sql.ErrNoRows {
		logger.Warn("Lock was not held by this process: key=%s, owner=%s", l.key, l.owner)
		return fmt.Errorf("lock was not held by this process")
	}

	if err != nil {
		if isMissingLockTable(err) {
			switchToMemoryFallback(err)
			return l.releaseMemoryLock()
		}
		logger.Error("Failed to release lock: %v", err)
		return fmt.Errorf("failed to release lock: %w", err)
	}

	l.acquired = false
	logger.Debug("Lock released: key=%s, owner=%s", l.key, l.owner)
	return nil
}

// Extend - ขยายระยะเวลา lock
// แทนที่ Redis PEXPIRE with Lua script
func (l *DistributedLock) Extend(ctx context.Context, additionalTime time.Duration) error {
	if !l.acquired {
		return fmt.Errorf("lock is not acquired")
	}

	if usingMemoryFallback() {
		return l.extendMemoryLock(additionalTime)
	}

	if l.db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE distributed_locks
		SET expires_at = NOW() + $1::interval
		WHERE lock_key = $2 AND owner = $3
		RETURNING lock_key
	`

	expiryInterval := fmt.Sprintf("%d seconds", int(additionalTime.Seconds()))

	var lockKey string
	err := l.db.QueryRowContext(ctx, query, expiryInterval, l.key, l.owner).Scan(&lockKey)

	if err == sql.ErrNoRows {
		logger.Warn("Lock was not held by this process: key=%s, owner=%s", l.key, l.owner)
		return fmt.Errorf("lock was not held by this process")
	}

	if err != nil {
		if isMissingLockTable(err) {
			switchToMemoryFallback(err)
			return l.extendMemoryLock(additionalTime)
		}
		logger.Error("Failed to extend lock: %v", err)
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	logger.Debug("Lock extended: key=%s, owner=%s, additional_time=%v", l.key, l.owner, additionalTime)
	return nil
}

// IsAcquired - ตรวจสอบว่า lock ถูกขอแล้วหรือยัง
func (l *DistributedLock) IsAcquired() bool {
	return l.acquired
}

// cleanupExpiredLocks - ทำความสะอาด expired locks
func (l *DistributedLock) cleanupExpiredLocks(ctx context.Context) {
	if usingMemoryFallback() || l.db == nil {
		cleanupExpiredMemoryLocks()
		return
	}

	query := `
		DELETE FROM distributed_locks
		WHERE expires_at < NOW()
	`

	result, err := l.db.ExecContext(ctx, query)
	if err != nil {
		if isMissingLockTable(err) {
			switchToMemoryFallback(err)
			cleanupExpiredMemoryLocks()
			return
		}
		logger.Error("Failed to cleanup expired locks: %v", err)
		return
	}

	deleted, _ := result.RowsAffected()
	if deleted > 0 {
		logger.Debug("Cleaned up %d expired locks", deleted)
	}
}

// LockManager - จัดการ locks
type LockManager struct {
	db *sql.DB
}

// NewLockManager - สร้าง LockManager ใหม่
func NewLockManager(db *sql.DB) *LockManager {
	return &LockManager{db: db}
}

// CleanupExpiredLocks - ทำความสะอาด expired locks ทั้งหมด
func (lm *LockManager) CleanupExpiredLocks(ctx context.Context) (int64, error) {
	if usingMemoryFallback() || lm.db == nil {
		deleted := cleanupExpiredMemoryLocks()
		return deleted, nil
	}

	query := `
		DELETE FROM distributed_locks
		WHERE expires_at < NOW()
	`

	result, err := lm.db.ExecContext(ctx, query)
	if err != nil {
		if isMissingLockTable(err) {
			switchToMemoryFallback(err)
			deleted := cleanupExpiredMemoryLocks()
			return deleted, nil
		}
		logger.Error("Failed to cleanup expired locks: %v", err)
		return 0, fmt.Errorf("failed to cleanup: %w", err)
	}

	deleted, _ := result.RowsAffected()
	if deleted > 0 {
		logger.Info("Cleaned up %d expired locks", deleted)
	}

	return deleted, nil
}

// GetActiveLocks - ดึงรายการ lock ที่กำลัง active
func (lm *LockManager) GetActiveLocks(ctx context.Context) ([]map[string]interface{}, error) {
	if usingMemoryFallback() || lm.db == nil {
		memoryLockMu.Lock()
		defer memoryLockMu.Unlock()

		locks := make([]map[string]interface{}, 0, len(memoryLocks))
		now := time.Now()
		for key, entry := range memoryLocks {
			if entry.expiresAt.After(now) {
				locks = append(locks, map[string]interface{}{
					"lock_key":    key,
					"owner":       entry.owner,
					"acquired_at": entry.acquiredAt,
					"expires_at":  entry.expiresAt,
					"ttl":         time.Until(entry.expiresAt).Seconds(),
				})
			}
		}
		return locks, nil
	}

	query := `
		SELECT lock_key, owner, acquired_at, expires_at
		FROM distributed_locks
		WHERE expires_at > NOW()
		ORDER BY acquired_at DESC
	`

	rows, err := lm.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("Failed to get active locks: %v", err)
		return nil, fmt.Errorf("failed to get active locks: %w", err)
	}
	defer rows.Close()

	locks := make([]map[string]interface{}, 0)
	for rows.Next() {
		var lockKey, owner string
		var acquiredAt, expiresAt time.Time

		if err := rows.Scan(&lockKey, &owner, &acquiredAt, &expiresAt); err != nil {
			logger.Error("Failed to scan lock: %v", err)
			continue
		}

		locks = append(locks, map[string]interface{}{
			"lock_key":    lockKey,
			"owner":       owner,
			"acquired_at": acquiredAt,
			"expires_at":  expiresAt,
			"ttl":         time.Until(expiresAt).Seconds(),
		})
	}

	return locks, nil
}

// GetLockCount - ดึงจำนวน lock ที่กำลัง active
func (lm *LockManager) GetLockCount(ctx context.Context) (int64, error) {
	if usingMemoryFallback() || lm.db == nil {
		memoryLockMu.Lock()
		defer memoryLockMu.Unlock()
		var count int64
		now := time.Now()
		for _, entry := range memoryLocks {
			if entry.expiresAt.After(now) {
				count++
			}
		}
		return count, nil
	}

	query := `
		SELECT COUNT(*)
		FROM distributed_locks
		WHERE expires_at > NOW()
	`

	var count int64
	err := lm.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		logger.Error("Failed to get lock count: %v", err)
		return 0, fmt.Errorf("failed to get lock count: %w", err)
	}

	return count, nil
}

// ForceReleaseLock - บังคับปล่อย lock (ใช้เฉพาะกรณีฉุกเฉิน)
func (lm *LockManager) ForceReleaseLock(ctx context.Context, lockKey string) error {
	if usingMemoryFallback() || lm.db == nil {
		memoryLockMu.Lock()
		defer memoryLockMu.Unlock()
		if _, exists := memoryLocks[lockKey]; !exists {
			return fmt.Errorf("lock not found: %s", lockKey)
		}
		delete(memoryLocks, lockKey)
		logger.Warn("Force released in-memory lock: key=%s", lockKey)
		return nil
	}

	query := `
		DELETE FROM distributed_locks
		WHERE lock_key = $1
		RETURNING lock_key
	`

	var deletedKey string
	err := lm.db.QueryRowContext(ctx, query, lockKey).Scan(&deletedKey)

	if err == sql.ErrNoRows {
		return fmt.Errorf("lock not found: %s", lockKey)
	}

	if err != nil {
		if isMissingLockTable(err) {
			switchToMemoryFallback(err)
			return lm.ForceReleaseLock(ctx, lockKey)
		}
		logger.Error("Failed to force release lock: %v", err)
		return fmt.Errorf("failed to force release lock: %w", err)
	}

	logger.Warn("Force released lock: key=%s", lockKey)
	return nil
}

// GetLockInfo - ดึงข้อมูล lock
func (lm *LockManager) GetLockInfo(ctx context.Context, lockKey string) (map[string]interface{}, error) {
	if usingMemoryFallback() || lm.db == nil {
		memoryLockMu.Lock()
		defer memoryLockMu.Unlock()
		entry, exists := memoryLocks[lockKey]
		if !exists || entry.expiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("lock not found or expired: %s", lockKey)
		}
		return map[string]interface{}{
			"lock_key":    lockKey,
			"owner":       entry.owner,
			"acquired_at": entry.acquiredAt,
			"expires_at":  entry.expiresAt,
			"ttl":         time.Until(entry.expiresAt).Seconds(),
			"is_active":   entry.expiresAt.After(time.Now()),
		}, nil
	}

	query := `
		SELECT lock_key, owner, acquired_at, expires_at
		FROM distributed_locks
		WHERE lock_key = $1 AND expires_at > NOW()
	`

	var owner string
	var acquiredAt, expiresAt time.Time

	err := lm.db.QueryRowContext(ctx, query, lockKey).Scan(&lockKey, &owner, &acquiredAt, &expiresAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lock not found or expired: %s", lockKey)
	}

	if err != nil {
		if isMissingLockTable(err) {
			switchToMemoryFallback(err)
			return lm.GetLockInfo(ctx, lockKey)
		}
		logger.Error("Failed to get lock info: %v", err)
		return nil, fmt.Errorf("failed to get lock info: %w", err)
	}

	return map[string]interface{}{
		"lock_key":    lockKey,
		"owner":       owner,
		"acquired_at": acquiredAt,
		"expires_at":  expiresAt,
		"ttl":         time.Until(expiresAt).Seconds(),
		"is_active":   expiresAt.After(time.Now()),
	}, nil
}

// StartCleanupRoutine - เริ่มต้น goroutine สำหรับทำความสะอาด expired locks อัตโนมัติ
func (lm *LockManager) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("Started lock cleanup routine (interval: %v)", interval)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping lock cleanup routine")
			return
		case <-ticker.C:
			deleted, err := lm.CleanupExpiredLocks(context.Background())
			if err != nil {
				logger.Error("Lock cleanup error: %v", err)
			} else if deleted > 0 {
				logger.Info("Lock cleanup: removed %d expired locks", deleted)
			}
		}
	}
}

// Backward compatibility functions to match myredis API

// NewDistributedLockCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.NewDistributedLock
// ต้องการ db connection เป็น parameter แรก (แตกต่างจาก Redis ที่ใช้ global client)
func NewDistributedLockCompat(db *sql.DB, key string, expiry time.Duration) *DistributedLock {
	return NewDistributedLock(db, key, expiry)
}
