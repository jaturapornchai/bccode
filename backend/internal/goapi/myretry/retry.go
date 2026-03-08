package myretry

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"
)

// RetryConfig การตั้งค่า retry
type RetryConfig struct {
	MaxRetries  int           // จำนวนครั้งสูงสุดที่จะ retry
	BaseDelay   time.Duration // เวลารอเริ่มต้น
	MaxDelay    time.Duration // เวลารอสูงสุด
	Multiplier  float64       // ตัวคูณสำหรับ exponential backoff (แนะนำ 2.0)
	OnRetry     func(attempt int, err error)
	IsRetryable func(err error) bool // ฟังก์ชันตรวจสอบว่า error นี้ควร retry หรือไม่
}

// DefaultRetryConfig ค่า default สำหรับ retry
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		OnRetry: func(attempt int, err error) {
			logger.Warn("Retry attempt %d due to error: %v", attempt, err)
		},
		IsRetryable: func(err error) bool {
			// ตัวอย่าง: retry กับ error ที่เป็นปัญหาชั่วคราว
			if err == nil {
				return false
			}
			errMsg := strings.ToLower(err.Error())
			// Retry สำหรับ connection, timeout, temporary errors
			return strings.Contains(errMsg, "connection") ||
				strings.Contains(errMsg, "timeout") ||
				strings.Contains(errMsg, "temporary") ||
				strings.Contains(errMsg, "too many") ||
				strings.Contains(errMsg, "unavailable") ||
				strings.Contains(errMsg, "deadlock")
		},
	}
}

// WithRetry ทำงานซ้ำด้วย exponential backoff
func WithRetry(ctx context.Context, config *RetryConfig, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// ลองทำงาน
		err := fn()

		if err == nil {
			// สำเร็จ
			if attempt > 0 {
				logger.Info("✅ Operation succeeded after %d retries", attempt)
			}
			return nil
		}

		lastErr = err

		// เช็คว่าควร retry หรือไม่
		if !config.IsRetryable(err) {
			logger.Error("❌ Error is not retryable: %v", err)
			return err
		}

		// ถ้าหมด retry แล้ว
		if attempt >= config.MaxRetries {
			logger.Error("❌ Max retries (%d) exceeded", config.MaxRetries)
			break
		}

		// เรียก callback (ถ้ามี)
		if config.OnRetry != nil {
			config.OnRetry(attempt+1, err)
		}

		// คำนวณเวลารอ (exponential backoff)
		power := uint(attempt)
		backoffMultiplier := float64(int64(1) << power) // 2^attempt
		delay := time.Duration(float64(config.BaseDelay) * backoffMultiplier * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		logger.Warn("⏳ Retrying in %v (attempt %d/%d)...", delay, attempt+1, config.MaxRetries)

		// รอก่อน retry
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
			// continue to retry
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// WithRetryAsync ทำงานซ้ำแบบ async และส่งผลลัพธ์ไปยัง channel
func WithRetryAsync(ctx context.Context, config *RetryConfig, fn func() error, resultChan chan<- error) {
	go func() {
		err := WithRetry(ctx, config, fn)
		if resultChan != nil {
			resultChan <- err
		}
	}()
}

// RetryableError สร้าง error ที่บอกว่าควร retry
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError สร้าง RetryableError
func NewRetryableError(err error) error {
	return &RetryableError{Err: err}
}

// NonRetryableError สร้าง error ที่ไม่ควร retry
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return e.Err.Error()
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

// NewNonRetryableError สร้าง NonRetryableError
func NewNonRetryableError(err error) error {
	return &NonRetryableError{Err: err}
}
