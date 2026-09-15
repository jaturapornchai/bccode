package stockengine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/lib/pq"
)

// fallbackInterval ความถี่ของการตรวจคิวเอง เผื่อสัญญาณปลุกหายระหว่างที่การเชื่อมต่อหลุด
// ระบบต้นแบบวน timer ทุกหนึ่งวินาทีตลอดเวลา ที่นี่ปกติจะถูกปลุกด้วยสัญญาณ
// รอบตรวจเองจึงเป็นเพียงตาข่ายกันพลาด ไม่ใช่กลไกหลัก
const fallbackInterval = 15 * time.Second

// Worker คำนวณต้นทุนสินค้าที่ค้างอยู่ในคิวของฐานข้อมูลหนึ่งฐาน
type Worker struct {
	DB           *sql.DB
	DSN          string
	Owner        string
	BatchSize    int
	TransFlags   []int
	Options      Options
	PollInterval time.Duration
}

// NewWorker สร้าง worker พร้อมค่าเริ่มต้นที่ใช้งานได้ทันที
func NewWorker(db *sql.DB, dsn string, transFlags []int) *Worker {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "worker"
	}
	return &Worker{
		DB:           db,
		DSN:          dsn,
		Owner:        fmt.Sprintf("%s-%d", host, os.Getpid()),
		BatchSize:    16,
		TransFlags:   transFlags,
		Options:      DefaultOptions(),
		PollInterval: fallbackInterval,
	}
}

// Run ทำงานจนกว่า context จะถูกยกเลิก
//
// ถูกปลุกด้วยสัญญาณจากฐานข้อมูลเมื่อมีงานใหม่ และมีรอบตรวจเองเป็นตาข่ายกันพลาด
func (w *Worker) Run(ctx context.Context) error {
	if w.BatchSize <= 0 {
		w.BatchSize = 16
	}
	if w.PollInterval <= 0 {
		w.PollInterval = fallbackInterval
	}

	wake := w.listen(ctx)

	ticker := time.NewTicker(w.PollInterval)
	defer ticker.Stop()

	logger.Info("🧮 stock engine worker started (owner=%s, batch=%d)", w.Owner, w.BatchSize)

	for {
		// ทำงานที่ค้างอยู่ให้หมดก่อนกลับไปรอสัญญาณ
		for {
			done, err := w.DrainOnce(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				logger.Error("stock engine drain failed: %v", err)
				break
			}
			if done == 0 {
				break
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		case <-wake:
		}
	}
}

// DrainOnce หยิบงานหนึ่งชุดมาทำ แล้วคืนจำนวนงานที่หยิบได้
func (w *Worker) DrainOnce(ctx context.Context) (int, error) {
	items, err := ClaimDirty(ctx, w.DB, w.Owner, w.BatchSize)
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		w.process(ctx, item)
	}
	return len(items), nil
}

// process คำนวณงานหนึ่งชิ้น งานที่ล้มเหลวจะถูกปล่อยกลับเข้าคิวเสมอ ไม่หายเงียบ
func (w *Worker) process(ctx context.Context, item DirtyItem) {
	scope := Scope{BusinessCode: item.BusinessCode, ItemCode: item.ItemCode, From: item.FromDate}

	rows, err := Recalculate(ctx, w.DB, scope, w.TransFlags, w.Options)
	if err == nil {
		logger.Debug("stock engine recalculated %s/%s: %d rows", item.BusinessCode, item.ItemCode, rows)
		return
	}

	// มีตัวอื่นถืองานนี้อยู่ ปล่อยคืนทันทีโดยไม่นับเป็นความล้มเหลว
	if errors.Is(err, ErrLockBusy) {
		if releaseErr := ReleaseDirty(ctx, w.DB, item, nil); releaseErr != nil {
			logger.Error("stock engine release failed for %s/%s: %v", item.BusinessCode, item.ItemCode, releaseErr)
		}
		return
	}

	logger.Error("stock engine failed for %s/%s (attempt %d): %v", item.BusinessCode, item.ItemCode, item.Attempts, err)

	if item.Attempts >= MaxAttempts {
		if deadErr := MoveToDeadLetter(ctx, w.DB, item, err); deadErr != nil {
			logger.Error("stock engine dead letter failed for %s/%s: %v", item.BusinessCode, item.ItemCode, deadErr)
		}
		return
	}
	if releaseErr := ReleaseDirty(ctx, w.DB, item, err); releaseErr != nil {
		logger.Error("stock engine release failed for %s/%s: %v", item.BusinessCode, item.ItemCode, releaseErr)
	}
}

// listen เปิดช่องรับสัญญาณปลุกจากฐานข้อมูล
// ถ้าเปิดไม่ได้ worker ยังทำงานต่อได้ด้วยรอบตรวจเอง เพียงแต่ตอบสนองช้าลง
func (w *Worker) listen(ctx context.Context) <-chan struct{} {
	wake := make(chan struct{}, 1)
	if w.DSN == "" {
		logger.Warn("stock engine: no DSN provided, falling back to periodic polling every %s", w.PollInterval)
		return wake
	}

	listener := pq.NewListener(w.DSN, 2*time.Second, time.Minute, func(_ pq.ListenerEventType, err error) {
		if err != nil {
			logger.Warn("stock engine listener event: %v", err)
		}
	})
	if err := listener.Listen(notifyChannel); err != nil {
		logger.Warn("stock engine: cannot listen on %s (%v), falling back to periodic polling", notifyChannel, err)
		_ = listener.Close()
		return wake
	}

	go func() {
		defer listener.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case <-listener.Notify:
				select {
				case wake <- struct{}{}:
				default: // มีสัญญาณค้างอยู่แล้ว ไม่ต้องซ้อน
				}
			}
		}
	}()

	return wake
}
