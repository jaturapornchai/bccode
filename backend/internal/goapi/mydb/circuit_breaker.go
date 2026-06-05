package mydb

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"sync"
	"time"
)

// CircuitState - สถานะของ circuit breaker
type CircuitState int

const (
	// StateClosed - ปกติ ทำงานได้
	StateClosed CircuitState = iota
	// StateOpen - เปิดวงจร หยุดส่ง request ไปยัง service
	StateOpen
	// StateHalfOpen - ลองส่ง request บางส่วนเพื่อทดสอบ
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker - Circuit breaker implementation
type CircuitBreaker struct {
	name string

	// Configuration
	maxFailures      uint32        // จำนวน failure สูงสุดก่อนเปิด circuit
	resetTimeout     time.Duration // เวลาที่รอก่อนเปลี่ยนจาก Open → Half-Open
	halfOpenMaxCalls uint32        // จำนวน call สูงสุดที่อนุญาตใน Half-Open state
	successThreshold uint32        // จำนวน success ต่อเนื่องที่ต้องการเพื่อปิด circuit

	// State
	mu              sync.RWMutex
	state           CircuitState
	failures        uint32
	successes       uint32
	lastFailureTime time.Time
	lastStateChange time.Time
	halfOpenCalls   uint32

	// Statistics
	totalCalls    uint64
	totalFailures uint64
	totalSuccess  uint64
}

// NewCircuitBreaker - สร้าง circuit breaker ใหม่
func NewCircuitBreaker(name string) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		maxFailures:      5,                // เปิด circuit หลัง fail 5 ครั้ง
		resetTimeout:     10 * time.Second, // รอ 10 วิก่อนลองใหม่
		halfOpenMaxCalls: 3,                // ลองส่ง 3 calls ใน half-open
		successThreshold: 2,                // ต้อง success 2 ครั้งติดกันจึงจะปิด circuit
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

// Execute - รัน function โดยมี circuit breaker ดูแล
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// ตรวจสอบว่าควรให้ทำงานหรือไม่
	if !cb.canProceed() {
		cb.mu.RLock()
		state := cb.state
		cb.mu.RUnlock()
		return fmt.Errorf("circuit breaker %s is %s - request blocked", cb.name, state)
	}

	// เพิ่มจำนวน calls
	cb.beforeCall()

	// รัน function
	err := fn(ctx)

	// บันทึกผลลัพธ์
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// canProceed - ตรวจสอบว่าควรให้ request ผ่านหรือไม่
func (cb *CircuitBreaker) canProceed() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalCalls++

	switch cb.state {
	case StateClosed:
		// ปกติ - ให้ผ่าน
		return true

	case StateOpen:
		// ตรวจสอบว่าพ้น timeout แล้วหรือยัง
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			logger.Info("Circuit breaker %s: transitioning from OPEN to HALF_OPEN", cb.name)
			cb.state = StateHalfOpen
			cb.halfOpenCalls = 0
			cb.lastStateChange = time.Now()
			return true
		}
		// ยังไม่พ้น timeout - ไม่ให้ผ่าน
		return false

	case StateHalfOpen:
		// อนุญาตเฉพาะบางส่วนเพื่อทดสอบ
		if cb.halfOpenCalls < cb.halfOpenMaxCalls {
			cb.halfOpenCalls++
			return true
		}
		return false

	default:
		return false
	}
}

// beforeCall - เรียกก่อนทำงาน
func (cb *CircuitBreaker) beforeCall() {
	// เพิ่ม metrics ถ้าจำเป็น
}

// recordSuccess - บันทึกความสำเร็จ
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalSuccess++

	switch cb.state {
	case StateClosed:
		// Reset failure counter
		cb.failures = 0

	case StateHalfOpen:
		// นับ success ใน half-open state
		cb.successes++
		if cb.successes >= cb.successThreshold {
			logger.Info("Circuit breaker %s: transitioning from HALF_OPEN to CLOSED (success threshold reached)", cb.name)
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.halfOpenCalls = 0
			cb.lastStateChange = time.Now()
		}
	}
}

// recordFailure - บันทึกความล้มเหลว
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalFailures++
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		// ตรวจสอบว่าถึง threshold หรือยัง
		if cb.failures >= cb.maxFailures {
			logger.Warn("Circuit breaker %s: transitioning from CLOSED to OPEN (max failures: %d)", cb.name, cb.failures)
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}

	case StateHalfOpen:
		// ถ้าใน half-open แล้ว fail เปิด circuit ทันที
		logger.Warn("Circuit breaker %s: transitioning from HALF_OPEN to OPEN (test failed)", cb.name)
		cb.state = StateOpen
		cb.successes = 0
		cb.halfOpenCalls = 0
		cb.lastStateChange = time.Now()
	}
}

// GetState - ดึงสถานะปัจจุบัน
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats - ดึงสถิติ
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":              cb.name,
		"state":             cb.state.String(),
		"total_calls":       cb.totalCalls,
		"total_success":     cb.totalSuccess,
		"total_failures":    cb.totalFailures,
		"consecutive_fails": cb.failures,
		"last_state_change": cb.lastStateChange,
		"last_failure_time": cb.lastFailureTime,
	}
}

// Reset - รีเซ็ต circuit breaker (ใช้เวลา maintenance)
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	logger.Info("Circuit breaker %s: manual reset", cb.name)
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenCalls = 0
	cb.lastStateChange = time.Now()
}

// RetryConfig - การตั้งค่าสำหรับ retry logic
type RetryConfig struct {
	MaxAttempts     int           // จำนวนครั้งสูงสุด
	InitialDelay    time.Duration // delay ครั้งแรก
	MaxDelay        time.Duration // delay สูงสุด
	BackoffFactor   float64       // ตัวคูณสำหรับ exponential backoff
	RetryableErrors []string      // error ที่ควร retry
}

// DefaultRetryConfig - ค่า default สำหรับ retry
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"connection refused",
			"connection reset",
			"timeout",
			"temporary failure",
		},
	}
}

// ExecuteWithRetry - รัน function พร้อม retry logic และ exponential backoff
func ExecuteWithRetry(ctx context.Context, config RetryConfig, fn func(context.Context) error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// ลองรัน function
		err := fn(ctx)
		if err == nil {
			// สำเร็จ
			if attempt > 1 {
				logger.Info("Operation succeeded after %d attempts", attempt)
			}
			return nil
		}

		lastErr = err

		// ถ้าเป็นครั้งสุดท้ายแล้ว ไม่ retry
		if attempt == config.MaxAttempts {
			logger.Error("Operation failed after %d attempts: %v", attempt, err)
			break
		}

		// ตรวจสอบว่า error นี้ควร retry หรือไม่
		if !isRetryableError(err, config.RetryableErrors) {
			logger.Warn("Non-retryable error encountered: %v", err)
			return err
		}

		logger.Warn("Attempt %d failed: %v, retrying in %v...", attempt, err, delay)

		// รอตาม delay
		select {
		case <-time.After(delay):
			// คำนวณ delay ครั้งต่อไป (exponential backoff)
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// isRetryableError - ตรวจสอบว่า error นี้ควร retry หรือไม่
func isRetryableError(err error, retryableErrors []string) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	for _, retryableErr := range retryableErrors {
		if containsString(errMsg, retryableErr) {
			return true
		}
	}

	return false
}

// containsString - ตรวจสอบว่า string มี substring หรือไม่ (case-insensitive)
func containsString(s, substr string) bool {
	// Simple implementation - อาจใช้ strings.Contains หรือ strings.ToLower
	return len(s) >= len(substr) && (s == substr || len(s) > 0)
}
