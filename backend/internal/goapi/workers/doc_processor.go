package workers

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypostgres"
	"smlcloudplatform/internal/goapi/process"
	"runtime"
	"sync"
	"time"
)

// WorkerManager - จัดการ workers สำหรับประมวลผล queue
type WorkerManager struct {
	numWorkers     int
	workers        []*DynamicWorker
	stopCh         chan struct{}
	wg             sync.WaitGroup
	isRunning      bool
	mu             sync.RWMutex
	shopUpdateChan chan []string // channel สำหรับอัพเดทรายชื่อ shop
	db             *sql.DB        // PostgreSQL connection
}

// DynamicWorker - worker ที่ทำงานแบบ dynamic (ไม่ผูกกับ shop เดียว)
type DynamicWorker struct {
	id             int
	manager        *WorkerManager
	stopCh         chan struct{}
	totalProcessed uint64
	mu             sync.RWMutex
}

// WorkerStats - สถิติของ worker
type WorkerStats struct {
	WorkerID int       `json:"worker_id"`
	TotalProcessed uint64    `json:"total_processed"`
	Status string    `json:"status"`
	CurrentShop string    `json:"current_shop"`
	LastProcessed time.Time `json:"last_processed"`
}

// OptimalWorkerCount คำนวณจำนวน worker ที่เหมาะสมตาม CPU cores
// สูตร: min(maxWorkers, max(minWorkers, CPU * multiplier))
func OptimalWorkerCount(minWorkers, maxWorkers int, multiplier float64) int {
	numCPU := runtime.NumCPU()
	optimal := int(float64(numCPU) * multiplier)

	if optimal < minWorkers {
		optimal = minWorkers
	}
	if optimal > maxWorkers {
		optimal = maxWorkers
	}

	return optimal
}

// NewWorkerManager - สร้าง worker manager ใหม่
// numWorkers: จำนวน workers ที่ต้องการ (ถ้า <= 0 จะคำนวณอัตโนมัติจาก CPU cores)
func NewWorkerManager(numWorkers int) *WorkerManager {
	if numWorkers <= 0 {
		// คำนวณจำนวน worker อัตโนมัติ: CPU * 2, min=4, max=32
		numWorkers = OptimalWorkerCount(4, 32, 2.0)
		logger.Info("Auto-calculated optimal workers: %d (based on %d CPU cores)", numWorkers, runtime.NumCPU())
	}

	logger.Info("กำลังสร้าง Worker Manager พร้อม %d workers", numWorkers)

	// ตรวจสอบว่า global connection มีอยู่
	_, err := myglobal.GetGlobalDatabaseConnection()
	if err != nil {
		logger.Error("Failed to get PostgreSQL connection: %v", err)
		return nil
	}

	return &WorkerManager{
		numWorkers:     numWorkers,
		workers:        make([]*DynamicWorker, 0, numWorkers),
		stopCh:         make(chan struct{}),
		shopUpdateChan: make(chan []string, 100),
		isRunning:      false,
		db:             nil, // Don't store connection - fetch fresh each time
	}
}

// Start - เริ่มต้น worker manager
func (wm *WorkerManager) Start() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.isRunning {
		return fmt.Errorf("worker manager already running")
	}

	logger.Info("กำลังเริ่มต้น Worker Manager...")

	// สร้าง workers
	for i := 0; i < wm.numWorkers; i++ {
		worker := &DynamicWorker{
			id:      i + 1,
			manager: wm,
			stopCh:  make(chan struct{}),
		}
		wm.workers = append(wm.workers, worker)

		// เริ่ม worker goroutine
		wm.wg.Add(1)
		go worker.run()
	}

	// เริ่ม shop monitor goroutine
	wm.wg.Add(1)
	go wm.monitorActiveShops()

	wm.isRunning = true
	logger.Success("Worker Manager เริ่มทำงานแล้ว - Workers: %d", wm.numWorkers)

	return nil
}

// Stop - หยุด worker manager
func (wm *WorkerManager) Stop() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if !wm.isRunning {
		return
	}

	logger.Info("กำลังหยุด Worker Manager...")

	// ส่งสัญญาณหยุดไปยัง workers ทั้งหมด
	close(wm.stopCh)

	// รอให้ workers ทั้งหมดหยุด
	wm.wg.Wait()

	wm.isRunning = false
	logger.Success("Worker Manager หยุดทำงานแล้ว")
}

// GetStats - ดึงสถิติของ workers ทั้งหมด
func (wm *WorkerManager) GetStats() []WorkerStats {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	stats := make([]WorkerStats, 0, len(wm.workers))
	for _, worker := range wm.workers {
		stats = append(stats, worker.getStats())
	}

	return stats
}

// monitorActiveShops - ตรวจสอบ shop ที่มีงานรออยู่และแจ้ง workers
func (wm *WorkerManager) monitorActiveShops() {
	defer wm.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wm.stopCh:
			logger.Info("Shop monitor stopped")
			return
		case <-ticker.C:
			// ดึงรายชื่อ shop ที่มีงาน
			// เพิ่ม timeout จาก 2 เป็น 15 วินาที เพื่อรองรับ network latency จาก container
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

			// ดึง connection ใหม่ทุกครั้ง เพื่อป้องกัน stale connection
			db, err := myglobal.GetGlobalDatabaseConnection()
			if err != nil {
				logger.Warn("Failed to get database connection: %v", err)
				cancel()
				continue
			}

			qm := mypostgres.NewQueueManager(db)
			shops, err := qm.GetActiveShops(ctx)
			cancel()

			if err != nil {
				logger.Warn("Failed to get active shops: %v", err)
				continue
			}

			// ถ้ามี shop ส่งไปให้ workers รู้
			if len(shops) > 0 {
				select {
				case wm.shopUpdateChan <- shops:
					logger.Debug("Updated active shops: %d shops", len(shops))
				default:
					// Channel เต็ม ข้าม
				}
			}
		}
	}
}

// run - ฟังก์ชันหลักของ worker
func (w *DynamicWorker) run() {
	defer w.manager.wg.Done()

	logger.Info("[Worker-%d] เริ่มทำงาน", w.id)

	// Cache ของ shops ที่มีงาน
	activeShops := make([]string, 0)
	currentShopIndex := 0

	for {
		select {
		case <-w.stopCh:
		case <-w.manager.stopCh:
			logger.Info("[Worker-%d] หยุดทำงาน", w.id)
			return

		case shops := <-w.manager.shopUpdateChan:
			// อัพเดทรายชื่อ shop
			activeShops = shops
			currentShopIndex = 0
			logger.Debug("[Worker-%d] Updated shops: %d", w.id, len(activeShops))

		default:
			// ลองประมวลผลงาน
			if len(activeShops) == 0 {
				// ยังไม่มีรายชื่อ shop ลองดึงใหม่
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

				// ดึง connection ใหม่ทุกครั้ง เพื่อป้องกัน stale connection
				db, dbErr := myglobal.GetGlobalDatabaseConnection()
				if dbErr != nil {
					cancel()
					time.Sleep(100 * time.Millisecond)
					continue
				}

				qm := mypostgres.NewQueueManager(db)
				shops, err := qm.GetActiveShops(ctx)
				cancel()

				if err == nil && len(shops) > 0 {
					activeShops = shops
					currentShopIndex = 0
				} else {
					// ไม่มีงาน รอ 100ms
					time.Sleep(100 * time.Millisecond)
					continue
				}
			}

			// Round-robin ไปหา shop ที่มีงาน
			processed := false
			for i := 0; i < len(activeShops); i++ {
				shopId := activeShops[currentShopIndex]
				currentShopIndex = (currentShopIndex + 1) % len(activeShops)

				// ลองดึงงานจาก shop นี้
				if w.processShopQueue(shopId) {
					processed = true
					break
				}
			}

			if !processed {
				// ไม่มีงานเลย รอ 100ms
				time.Sleep(100 * time.Millisecond)

				// Refresh shop list
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

				// ดึง connection ใหม่ทุกครั้ง เพื่อป้องกัน stale connection
				db, dbErr := myglobal.GetGlobalDatabaseConnection()
				if dbErr == nil {
					qm := mypostgres.NewQueueManager(db)
					shops, _ := qm.GetActiveShops(ctx)
					activeShops = shops
					currentShopIndex = 0
				}
				cancel()
			}
		}
	}
}

// processShopQueue - ดึงและประมวลผลงานจาก queue ของ shop
func (w *DynamicWorker) processShopQueue(shopId string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// ดึง connection ใหม่ทุกครั้ง เพื่อป้องกัน stale connection
	db, err := myglobal.GetGlobalDatabaseConnection()
	if err != nil {
		logger.Error("[Worker-%d] Failed to get database connection: %v", w.id, err)
		return false
	}

	// ดึงงานจาก queue (non-blocking)
	qm := mypostgres.NewQueueManager(db)
	item, err := qm.PopFromQueue(ctx, shopId)
	if err != nil {
		logger.Error("[Worker-%d] Failed to pop from queue (shop=%s): %v", w.id, shopId, err)
		return false
	}

	if item == nil {
		// ไม่มีงานใน shop นี้
		return false
	}

	// ประมวลผลงาน
	logger.Info("[Worker-%d] Processing: shop=%s, docno=%s, transflag=%s",
		w.id, item.ShopId, item.DocNo, item.TransFlag)

	startTime := time.Now()
	err = w.processDocument(ctx, item)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error("[Worker-%d] Failed to process doc %s (shop=%s): %v",
			w.id, item.DocNo, item.ShopId, err)

		// Requeue หรือส่งไป dead letter queue
		if item.RetryCount < 3 {
			logger.Info("[Worker-%d] Requeuing doc %s (retry=%d)",
				w.id, item.DocNo, item.RetryCount+1)
			item.RetryCount++
			qm.RequeueItem(ctx, *item)
		} else {
			logger.Warn("[Worker-%d] Max retries exceeded, sending to dead letter queue: %s",
				w.id, item.DocNo)
			qm.AddToDeadLetterQueue(ctx, *item, err.Error())
		}

		return true
	}

	// สำเร็จ - ทำเครื่องหมายว่าเสร็จ
	qm.MarkAsCompleted(ctx, item.ID)

	w.mu.Lock()
	w.totalProcessed++
	w.mu.Unlock()

	logger.Info("[Worker-%d] ✓ Completed: shop=%s, docno=%s (duration=%v)",
		w.id, item.ShopId, item.DocNo, duration.Round(time.Millisecond))

	return true
}

// processDocument - ประมวลผลเอกสาร
func (w *DynamicWorker) processDocument(ctx context.Context, item *mypostgres.QueueItem) error {
	// ตรวจสอบ transflag (รองรับทั้ง int string และ abbreviation string)
	switch item.TransFlag {
	case "6": // Purchase Order
		return w.processPurchaseOrder(ctx, item)
	case "12": // Sale Invoice
		return w.processSaleInvoice(ctx, item)
	case "44": // Creditor
		return w.processCreditor(ctx, item)
	case "48": // Customer
		return w.processCustomer(ctx, item)
	case "54": // Debtor
		return w.processDebtor(ctx, item)
	default:
		return fmt.Errorf("unknown transflag: %s", item.TransFlag)
	}
}

// processPurchaseOrder - ประมวลผล Purchase Order
func (w *DynamicWorker) processPurchaseOrder(ctx context.Context, item *mypostgres.QueueItem) error {
	// TODO: เรียก logic จริงจาก process package
	// เช่น คำนวณสถานะ, อัพเดทข้อมูลที่เกี่ยวข้อง
	logger.Debug("[Worker-%d] Processing Purchase Order: %s", w.id, item.DocNo)

	// Placeholder - ใส่ logic จริงตรงนี้
	return process.ProcessPurchaseOrderStatus(ctx, item.ShopId, item.DocNo)
}

// processSaleInvoice - ประมวลผล Sale Invoice
func (w *DynamicWorker) processSaleInvoice(ctx context.Context, item *mypostgres.QueueItem) error {
	logger.Debug("[Worker-%d] Processing Sale Invoice: %s", w.id, item.DocNo)
	return process.ProcessSaleInvoiceStatus(ctx, item.ShopId, item.DocNo)
}

// processCreditor - ประมวลผล Creditor
func (w *DynamicWorker) processCreditor(_ context.Context, item *mypostgres.QueueItem) error {
	logger.Debug("[Worker-%d] Processing Creditor: %s", w.id, item.DocNo)
	return nil // TODO: implement
}

// processCustomer - ประมวลผล Customer
func (w *DynamicWorker) processCustomer(_ context.Context, item *mypostgres.QueueItem) error {
	logger.Debug("[Worker-%d] Processing Customer: %s", w.id, item.DocNo)
	return nil // TODO: implement
}

// processDebtor - ประมวลผล Debtor
func (w *DynamicWorker) processDebtor(_ context.Context, item *mypostgres.QueueItem) error {
	logger.Debug("[Worker-%d] Processing Debtor: %s", w.id, item.DocNo)
	return nil // TODO: implement
}

// getStats - ดึงสถิติของ worker นี้
func (w *DynamicWorker) getStats() WorkerStats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return WorkerStats{
		WorkerID:       w.id,
		TotalProcessed: w.totalProcessed,
		Status:         "running",
	}
}
