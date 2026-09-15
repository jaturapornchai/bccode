package stockengine

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"

	"smlcloudplatform/internal/goapi/logger"
)

// manager ดูแลว่าแต่ละฐานข้อมูลของกลุ่มกิจการมี worker ทำงานอยู่ตัวเดียว
//
// worker ถูกเปิดเมื่อมีงานเข้ามาครั้งแรกของกลุ่มกิจการนั้น ไม่ต้องตั้งรายชื่อล่วงหน้า
// และไม่ต้องเปิดค้างไว้สำหรับกลุ่มกิจการที่ยังไม่มีการเคลื่อนไหว
type manager struct {
	mu      sync.Mutex
	running map[string]context.CancelFunc
	ctx     context.Context
	cancel  context.CancelFunc
	flags   []int
}

var defaultManager = &manager{running: map[string]context.CancelFunc{}}

// StartWorkers เปิดระบบ worker ทั้งหมด เรียกครั้งเดียวตอนระบบเริ่มทำงาน
func StartWorkers(parent context.Context, transFlags []int) {
	defaultManager.mu.Lock()
	defer defaultManager.mu.Unlock()

	if defaultManager.ctx != nil {
		return
	}
	defaultManager.ctx, defaultManager.cancel = context.WithCancel(parent)
	defaultManager.flags = transFlags
	logger.Info("🧮 stock engine enabled")
}

// StopWorkers ปิด worker ทั้งหมด
func StopWorkers() {
	defaultManager.mu.Lock()
	defer defaultManager.mu.Unlock()

	if defaultManager.cancel != nil {
		defaultManager.cancel()
	}
	defaultManager.ctx = nil
	defaultManager.cancel = nil
	defaultManager.running = map[string]context.CancelFunc{}
}

// Enabled บอกว่าระบบ worker ถูกเปิดไว้หรือไม่
func Enabled() bool {
	defaultManager.mu.Lock()
	defer defaultManager.mu.Unlock()
	return defaultManager.ctx != nil
}

// WorkerEnabledFromEnv อ่านสวิตช์เปิดปิด worker จาก environment
// ปิดได้ด้วย BCAI_STOCK_WORKER=0 เผื่อกรณีต้องการให้เครื่องนั้นรับแต่ข้อมูลโดยไม่คำนวณ
func WorkerEnabledFromEnv() bool {
	value := strings.TrimSpace(os.Getenv("BCAI_STOCK_WORKER"))
	return value != "0" && !strings.EqualFold(value, "false")
}

// EnsureWorker เปิด worker ของกลุ่มกิจการนี้ถ้ายังไม่มี
// เรียกซ้ำได้ไม่จำกัด ครั้งที่สองเป็นต้นไปไม่ทำอะไร
func EnsureWorker(holdingCode string, db *sql.DB, dsn string) {
	if holdingCode == "" || db == nil {
		return
	}

	defaultManager.mu.Lock()
	defer defaultManager.mu.Unlock()

	if defaultManager.ctx == nil {
		return // ระบบยังไม่เปิด worker
	}
	if _, exists := defaultManager.running[holdingCode]; exists {
		return
	}

	ctx, cancel := context.WithCancel(defaultManager.ctx)
	defaultManager.running[holdingCode] = cancel

	worker := NewWorker(db, dsn, defaultManager.flags)
	worker.Owner = worker.Owner + ":" + holdingCode

	go func() {
		defer func() {
			defaultManager.mu.Lock()
			delete(defaultManager.running, holdingCode)
			defaultManager.mu.Unlock()
		}()

		if err := worker.Run(ctx); err != nil && ctx.Err() == nil {
			logger.Error("stock engine worker for %s stopped: %v", holdingCode, err)
		}
	}()

	logger.Info("🧮 stock engine worker started for holding %s", holdingCode)
}
