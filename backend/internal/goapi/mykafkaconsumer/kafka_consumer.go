package mykafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"

	"github.com/segmentio/kafka-go"
)

type ConsumerHandleFunc func(msg string) error

type IKafkaConsumer interface {
	ConsumeMessage(server string, topic string, groupID string, partition int32, h ConsumerHandleFunc) error
}

// WorkerMetadata tracks worker activity for cleanup
type WorkerMetadata struct {
	channel      chan MessageJob
	lastActivity time.Time
	messageCount int64
	createdAt    time.Time
}

type KafkaConsumer struct {
	// ไม่ใช้ fixed worker count แล้ว เพราะจะสร้าง worker แยกตาม holdingcode
	shopWorkers map[string]*WorkerMetadata // แยก queue ตาม holdingcode with metadata
	mutex       sync.RWMutex               // ป้องกัน race condition

	// Configuration for worker cleanup
	maxIdleTime     time.Duration // Maximum idle time before cleanup (default: 30 minutes)
	maxWorkers      int           // Maximum concurrent workers (default: 1000)
	cleanupInterval time.Duration // Cleanup check interval (default: 5 minutes)

	// Configuration for parallel processing
	workersPerShop int           // Number of parallel workers per shop (default: 3)
	jobTimeout     time.Duration // Timeout for each job (default: 30 seconds)

	// Timeout and retry configuration for Kafka connections
	kafkaTimeout      time.Duration // Overall Kafka operation timeout (default: 30 seconds)
	readTimeout       time.Duration // Read operation timeout (default: 15 seconds)
	dialTimeout       time.Duration // Connection timeout (default: 10 seconds)
	maxRetryAttempts  int           // Maximum retry attempts (default: 5)
	initialRetryDelay time.Duration // Initial retry delay (default: 1 second)
	maxRetryDelay     time.Duration // Maximum retry delay (default: 30 seconds)
}

// MessageJob represents a message to be processed
type MessageJob struct {
	holdingCode string // เพิ่ม holdingCode เพื่อแยก partition
	topic       string
	offset      int64
	partition   int
	value       string
	handler     ConsumerHandleFunc
}

// HoldingCodeExtractor extracts holdingCode from message payload
type HoldingCodeExtractor func(msg string) string

func NewKafkaConsumer() IKafkaConsumer {
	kc := &KafkaConsumer{
		shopWorkers:     make(map[string]*WorkerMetadata),
		mutex:           sync.RWMutex{},
		maxIdleTime:     30 * time.Minute, // Default: cleanup after 30 minutes of inactivity
		maxWorkers:      1000,             // Default: max 1000 concurrent shop workers
		cleanupInterval: 5 * time.Minute,  // Default: check for cleanup every 5 minutes
		workersPerShop:  3,                // Default: 3 parallel workers per shop
		jobTimeout:      30 * time.Second, // Default: 30 second timeout per job

		// Stable configuration - เน้นความเสถียรของการเชื่อมต่อ Kafka (remote broker)
		kafkaTimeout:      60 * time.Second, // 60s timeout สำหรับ remote broker
		readTimeout:       30 * time.Second, // 30s read timeout
		dialTimeout:       15 * time.Second, // 15s connection timeout
		maxRetryAttempts:  3,                // 3 ครั้งพอดี
		initialRetryDelay: 2 * time.Second,  // 2s initial delay
		maxRetryDelay:     30 * time.Second, // 30s maximum delay
	}

	// Start background cleanup goroutine
	go kc.startCleanupRoutine()

	return kc
}

// calculateExponentialBackoff calculates retry delay using exponential backoff with jitter
func (kc *KafkaConsumer) calculateExponentialBackoff(attempt int) time.Duration {
	// Exponential backoff: base_delay * 2^attempt
	backoff := float64(kc.initialRetryDelay) * math.Pow(2, float64(attempt))

	// Add jitter (random factor between 0.5 and 1.5) to prevent thundering herd
	jitter := 0.5 + rand.Float64()
	delay := time.Duration(backoff * jitter)

	// Cap at maximum retry delay
	if delay > kc.maxRetryDelay {
		delay = kc.maxRetryDelay
	}

	return delay
}

func isTransientKafkaError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	transientFragments := []string{
		"connection reset by peer",
		"wsarecv",
		"forcibly closed",
		"broken pipe",
		"closed network connection",
		"eof",
	}
	for _, frag := range transientFragments {
		if strings.Contains(msg, frag) {
			return true
		}
	}
	return false
}

// getOrCreateShopWorker สร้างหรือดึง worker channel สำหรับ holdingcode นั้นๆ
func (kc *KafkaConsumer) getOrCreateShopWorker(holdingCode string) chan MessageJob {
	// ตรวจสอบว่ามี worker สำหรับ shop นี้แล้วหรือยัง
	kc.mutex.RLock()
	if worker, exists := kc.shopWorkers[holdingCode]; exists {
		kc.mutex.RUnlock()
		// Update last activity time (thread-safe update)
		kc.updateWorkerActivity(holdingCode)
		return worker.channel
	}
	kc.mutex.RUnlock()

	// ยังไม่มี - สร้างใหม่
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	// Double-check (อาจมีคนอื่นสร้างไปแล้วระหว่างรอ lock)
	if worker, exists := kc.shopWorkers[holdingCode]; exists {
		return worker.channel
	}

	// Check if we need to evict old workers (LRU eviction)
	if len(kc.shopWorkers) >= kc.maxWorkers {
		kc.evictLRUWorker()
	}

	// สร้าง worker channel ใหม่สำหรับ shop นี้
	// ⚡ เพิ่ม buffer จาก 100 → 500 เพื่อรองรับ peak traffic
	workerChan := make(chan MessageJob, 500) // buffer 500 messages per shop
	now := time.Now()
	metadata := &WorkerMetadata{
		channel:      workerChan,
		lastActivity: now,
		messageCount: 0,
		createdAt:    now,
	}
	kc.shopWorkers[holdingCode] = metadata

	// เริ่ม worker pool สำหรับ shop นี้ (parallel processing)
	for i := 0; i < kc.workersPerShop; i++ {
		go kc.startShopWorker(holdingCode, workerChan, i+1)
	}

	logger.Info("✨ สร้าง worker pool สำหรับ HoldingCode: %s (%d workers, ร้านค้าทั้งหมด: %d)",
		holdingCode, kc.workersPerShop, len(kc.shopWorkers))
	return workerChan
}

// updateWorkerActivity updates the last activity timestamp for a worker
func (kc *KafkaConsumer) updateWorkerActivity(holdingCode string) {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()
	if worker, exists := kc.shopWorkers[holdingCode]; exists {
		worker.lastActivity = time.Now()
		worker.messageCount++
	}
}

// startShopWorker เริ่มต้น worker สำหรับ shop นั้นๆ (ประมวลผลแบบ parallel with timeout)
func (kc *KafkaConsumer) startShopWorker(holdingCode string, jobChan <-chan MessageJob, workerID int) {
	logger.Info("🔧 Worker #%d เริ่มทำงานสำหรับ HoldingCode: %s (การประมวลผลแบบพาราเรล พร้อม timeout %v)",
		workerID, holdingCode, kc.jobTimeout)

	messageCount := 0
	errorCount := 0
	timeoutCount := 0
	totalDuration := time.Duration(0)

	for job := range jobChan {
		messageCount++
		startTime := time.Now()

		// ⚡ ประมวลผลแบบ PARALLEL with timeout protection
		ctx, cancel := context.WithTimeout(context.Background(), kc.jobTimeout)

		// Create a channel for the result
		done := make(chan error, 1)
		go func() {
			done <- job.handler(job.value)
		}()

		// Wait for either completion or timeout
		select {
		case err := <-done:
			cancel()
			if err != nil {
				errorCount++
				logger.Error("❌ HoldingCode %s Worker #%d ล้มเหลว (topic=%s, offset=%d): %v",
					holdingCode, workerID, job.topic, job.offset, err)
			} else {
				duration := time.Since(startTime)
				totalDuration += duration
				avgDuration := totalDuration / time.Duration(messageCount)

				if messageCount%10 == 0 {
					logger.Info("✅ HoldingCode %s Worker #%d: ประมวลผล=%d, ข้อผิดพลาด=%d, timeout=%d, เฉลี่ย=%.2fs, ครั้งล่าสุด=%.2fs",
						holdingCode, workerID, messageCount, errorCount, timeoutCount, avgDuration.Seconds(), duration.Seconds())
				}
			}
		case <-ctx.Done():
			cancel()
			timeoutCount++
			logger.Error("⏱️ HoldingCode %s Worker #%d TIMEOUT (topic=%s, offset=%d, ระยะเวลา=%v)",
				holdingCode, workerID, job.topic, job.offset, kc.jobTimeout)
		}
	}

	logger.Info("🛑 Worker #%d หยุดทำงานสำหรับ HoldingCode: %s (ประมวลผล=%d, ข้อผิดพลาด=%d, timeout=%d)",
		workerID, holdingCode, messageCount, errorCount, timeoutCount)
}

func (kc *KafkaConsumer) ConsumeMessage(server string, topic string, groupID string, partition int32, h ConsumerHandleFunc) error {
	logger.Info("🚀 เริ่มต้น consumer สำหรับ topic: %s (group: %s)", topic, groupID)

	retryAttempt := 0
	for {
		err := kc.consumeWithRetry(server, topic, groupID, partition, h)
		if err == nil {
			// consumeWithRetry shouldn't return nil, แต่กันไว้เพื่อความปลอดภัย
			retryAttempt = 0
			continue
		}

		retryAttempt++
		if retryAttempt > kc.maxRetryAttempts {
			logger.Warn("Kafka consumer %s ถึงจำนวน retry สูงสุด (%d). รอ %v ก่อนเริ่มใหม่", topic, kc.maxRetryAttempts, kc.maxRetryDelay)
			retryAttempt = 0
			time.Sleep(kc.maxRetryDelay)
			continue
		}

		delay := kc.calculateExponentialBackoff(retryAttempt)
		if isTransientKafkaError(err) {
			logger.Warn("⚡ Kafka consumer %s เจอปัญหาเครือข่าย (attempt=%d): %v. จะลองใหม่ใน %v", topic, retryAttempt, err, delay)
		} else {
			logger.Error("❌ เกิดข้อผิดพลาดใน consume loop สำหรับ topic %s: %v. จะลองใหม่ใน %v", topic, err, delay)
		}
		time.Sleep(delay)
	}
}

func (kc *KafkaConsumer) consumeWithRetry(server string, topic string, groupID string, partition int32, h ConsumerHandleFunc) error {
	logger.Info("🔗 กำลังเชื่อมต่อไปยัง Kafka broker สำหรับ topic: %s (group: %s)...", topic, groupID)

	reader, err := kc.newKafkaReader(server, topic, groupID)
	if err != nil {
		logger.Error("❌ ไม่สามารถเชื่อมต่อ Kafka สำหรับ topic %s: %v", topic, err)
		return fmt.Errorf("error connecting to Kafka for consumer topic %s: %w", topic, err)
	}
	defer reader.Close()

	logger.Success("✅ เชื่อมต่อสำเร็จ! กำลังฟัง topic: %s (group: %s)", topic, groupID)

	// Message processing stats
	messageCount := 0
	lastLogTime := time.Now()
	startTime := time.Now()
	shopStats := make(map[string]int) // track messages per shop

	// Consumer loop - อ่าน messages และส่งไปยัง shop-specific worker
	for {
		// ใช้ timeout เพียง 1 วินาทีเพื่อเริ่มต้นทันที
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		msg, err := reader.ReadMessage(ctx)
		cancel()

		if err != nil {
			// Check if it's a timeout error
			if err == context.DeadlineExceeded {
				// Timeout is not fatal, continue to next iteration
				continue
			} else if strings.Contains(err.Error(), "context deadline exceeded") {
				continue
			} else if strings.Contains(err.Error(), "i/o timeout") {
				logger.Warn("⏱️ Kafka I/O timeout (อาจเป็นเครือข่ายช้า), ลองใหม่: %v", err)
				continue
			}
			if isTransientKafkaError(err) {
				logger.Warn("📉 Kafka connection lost (%v) – จะสร้าง reader ใหม่", err)
			} else {
				logger.Error("❌ ข้อผิดพลาดร้ายแรงในการอ่านข้อความ: %v", err)
			}
			return err
		}

		messageCount++
		msgValue := string(msg.Value)

		// 🔑 Extract HoldingCode from message
		holdingCode := extractHoldingCodeFromMessage(msgValue)
		if holdingCode == "" {
			logger.Warn("⚠️ ไม่สามารถแยก HoldingCode จากข้อความได้ (offset=%d), ข้าม", msg.Offset)
			continue
		}

		// อัปเดต stats
		shopStats[holdingCode]++

		// 🎯 ส่ง message ไปยัง worker ของ shop นั้นๆ (สร้างถ้ายังไม่มี)
		shopWorker := kc.getOrCreateShopWorker(holdingCode)

		job := MessageJob{
			holdingCode: holdingCode,
			topic:       topic,
			offset:      msg.Offset,
			partition:   msg.Partition,
			value:       msgValue,
			handler:     h,
		}

		// ส่ง job ไปยัง shop worker (non-blocking with timeout)
		select {
		case shopWorker <- job:
			// Message sent successfully
			logger.Debug("📤 Message queued for HoldingCode %s (offset=%d, queue_size=%d)",
				holdingCode, msg.Offset, len(shopWorker))
		case <-time.After(5 * time.Second):
			// Worker กำลังยุ่ง, log คำเตือนและส่งต่อ (blocking)
			logger.Warn("⚠️ HoldingCode %s worker กำลังยุ่ง, ข้อความถูกจัดคิวโดยมีความล่าช้า (offset=%d)",
				holdingCode, msg.Offset)
			shopWorker <- job
		}

		// Log stats every 100 messages or every minute
		if messageCount%100 == 0 || time.Since(lastLogTime) > time.Minute {
			avgRate := float64(messageCount) / time.Since(startTime).Seconds()
			activeShops := len(kc.shopWorkers)
			logger.Info("📊 Consumer %s สถิติ: ทั้งหมด=%d, อัตรา=%.2f ข้อความ/วินาที, ร้านค้าที่ใช้งาน=%d",
				topic, messageCount, avgRate, activeShops)

			// แสดง top 5 shops ที่มี message มากที่สุด
			type shopCount struct {
				holdingCode string
				count       int
			}
			var topShops []shopCount
			for sid, count := range shopStats {
				topShops = append(topShops, shopCount{sid, count})
			}
			if len(topShops) > 5 {
				logger.Info("📊 ร้านค้ายอดนิยม: แสดง 5/%d ร้านค้าที่ใช้งาน", len(topShops))
			}

			lastLogTime = time.Now()
		}
	}
}

func (kc *KafkaConsumer) newKafkaReader(servers string, topic string, groupID string) (*kafka.Reader, error) {
	svcConfig := serviceConfig.NewServiceConfig()

	brokers := []string{svcConfig.KafkaURI()}
	if strings.TrimSpace(servers) != "" {
		// รองรับ server หลายตัวคั่นด้วย comma
		parts := strings.Split(servers, ",")
		custom := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				custom = append(custom, trimmed)
			}
		}
		if len(custom) > 0 {
			brokers = custom
		}
	}

	// Create TLS config if SSL is enabled with enhanced timeout settings
	var dialer *kafka.Dialer
	if svcConfig.KafKaSecurityProtocol() == "SSL" {
		// For SSL/TLS configuration with segmentio/kafka-go, you'll need to implement
		// TLS configuration here if needed
		dialer = &kafka.Dialer{
			Timeout:   kc.dialTimeout, // Connection timeout (ใช้ค่าจาก config: 10s)
			DualStack: true,
			KeepAlive: 60 * time.Second, // ⚡ TCP KeepAlive สำหรับ remote broker
		}
	} else {
		// Create dialer even for non-SSL to configure timeouts
		dialer = &kafka.Dialer{
			Timeout:   kc.dialTimeout, // Connection timeout (15s)
			DualStack: true,
			KeepAlive: 60 * time.Second, // TCP KeepAlive สำหรับ remote broker
		}
	}

	// Stable Kafka reader configuration - เน้นความเสถียร
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		Dialer:      dialer,
		MinBytes:    1,                 // 1 byte minimum - รับ message ทันทีที่มี
		MaxBytes:    10e6,              // 10MB maximum batch size
		MaxWait:     1 * time.Second,   // รอ 1s ก่อน fetch (balance ระหว่างเร็วกับเสถียร)
		StartOffset: kafka.FirstOffset, // ⚠️ อ่านตั้งแต่ต้น เพื่อไม่พลาด message (เสถียร)

		// Stable settings - เน้นความเสถียรของการทำงาน
		QueueCapacity:  100,                    // 100 messages prefetch (พอดี)
		CommitInterval: 5 * time.Second,        // Commit ทุก 5s
		ReadBackoffMin: 500 * time.Millisecond, // 500ms minimum backoff
		ReadBackoffMax: 30 * time.Second,       // 30s max backoff

		// Stable session settings for reliable operation
		SessionTimeout:    30 * time.Second, // 30s session timeout (มาตรฐาน Kafka)
		RebalanceTimeout:  30 * time.Second, // 30s rebalance timeout
		HeartbeatInterval: 3 * time.Second,  // 3s heartbeat (< SessionTimeout/10)

		// Metadata refresh
		PartitionWatchInterval: 30 * time.Second,

		// Error handling - filter out rebalance progress messages (they're normal)
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			msgStr := fmt.Sprintf(msg, args...)
			// Don't log rebalance in progress as errors (it's normal behavior)
			if strings.Contains(msgStr, "Rebalance In Progress") ||
				strings.Contains(msgStr, "rebalancing the group") {
				// Silently ignore rebalance messages
				return
			}
			// Don't log I/O timeout errors (common with remote brokers)
			if strings.Contains(msgStr, "i/o timeout") {
				// Silently ignore I/O timeouts - they're network latency issues
				return
			}
			// Don't log generic timeout errors from remote broker
			if strings.Contains(msgStr, "read tcp") && strings.Contains(msgStr, "timeout") {
				// Silently ignore TCP timeout errors
				return
			}
			// Log other errors normally
			logger.Error("Kafka Reader Error: "+msg, args...)
		}),
	})

	return reader, nil
}

// startCleanupRoutine runs periodic cleanup of idle workers
func (kc *KafkaConsumer) startCleanupRoutine() {
	ticker := time.NewTicker(kc.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		kc.cleanupIdleWorkers()
	}
}

// cleanupIdleWorkers removes workers that have been idle for too long
func (kc *KafkaConsumer) cleanupIdleWorkers() {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	now := time.Now()
	removedCount := 0
	totalMessages := int64(0)

	for holdingCode, worker := range kc.shopWorkers {
		idleDuration := now.Sub(worker.lastActivity)

		if idleDuration > kc.maxIdleTime {
			// Close the channel to signal the worker to stop
			close(worker.channel)
			totalMessages += worker.messageCount
			delete(kc.shopWorkers, holdingCode)
			removedCount++

			logger.Info("🧹 ทำความสะอาด idle worker สำหรับ HoldingCode: %s (idle: %v, ข้อความ: %d)",
				holdingCode, idleDuration.Round(time.Second), worker.messageCount)
		}
	}

	if removedCount > 0 {
		logger.Info("🧹 การทำความสะอาดเสร็จสิ้น: ลบ %d idle workers, ประมวลผล %d ข้อความ, เหลือ: %d workers",
			removedCount, totalMessages, len(kc.shopWorkers))
	}
}

// evictLRUWorker removes the least recently used worker (called when maxWorkers is reached)
// NOTE: Must be called with mutex locked
func (kc *KafkaConsumer) evictLRUWorker() {
	if len(kc.shopWorkers) == 0 {
		return
	}

	// Find the worker with oldest lastActivity
	var oldestHoldingCode string
	var oldestTime time.Time
	firstIteration := true

	for holdingCode, worker := range kc.shopWorkers {
		if firstIteration || worker.lastActivity.Before(oldestTime) {
			oldestHoldingCode = holdingCode
			oldestTime = worker.lastActivity
			firstIteration = false
		}
	}

	// Remove the oldest worker
	if oldestHoldingCode != "" {
		worker := kc.shopWorkers[oldestHoldingCode]
		close(worker.channel)
		delete(kc.shopWorkers, oldestHoldingCode)

		logger.Warn("⚠️ LRU eviction: ลบ worker สำหรับ HoldingCode: %s (idle: %v, ข้อความ: %d, max workers: %d)",
			oldestHoldingCode, time.Since(worker.lastActivity).Round(time.Second),
			worker.messageCount, kc.maxWorkers)
	}
}

// extractHoldingCodeFromMessage ดึง holdingCode จาก JSON message
// รองรับหลายรูปแบบ: holdingcode, holdingCode, holdingcode, HoldingCode
// รองรับทั้ง JSON object (single doc) และ JSON array (bulk docs — ดึง holdingCode จากตัวแรก)
func extractHoldingCodeFromMessage(msg string) string {
	// ลอง parse เป็น object ก่อน
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		// ถ้าไม่ใช่ object → ลอง parse เป็น array (bulk messages)
		var arr []map[string]interface{}
		if err2 := json.Unmarshal([]byte(msg), &arr); err2 != nil {
			logger.Warn("ล้มเหลวในการแปลงข้อความเป็น JSON (ทั้ง object และ array): %v", err2)
			return ""
		}
		if len(arr) == 0 {
			logger.Warn("JSON array ว่าง — ไม่สามารถดึง HoldingCode ได้")
			return ""
		}
		data = arr[0] // ใช้ตัวแรกในการดึง HoldingCode
	}

	// ลองหา holdingCode ในหลายรูปแบบ (case-insensitive)
	possibleKeys := []string{"holdingcode", "holdingCode", "holdingcode", "HoldingCode", "HOLDING_CODE", "shop"}

	for _, key := range possibleKeys {
		// ตรวจสอบ key ตรงๆ
		if val, ok := data[key]; ok {
			return fmt.Sprintf("%v", val)
		}

		// ตรวจสอบ key แบบ case-insensitive
		for k, v := range data {
			if strings.EqualFold(k, key) {
				return fmt.Sprintf("%v", v)
			}
		}
	}

	logger.Warn("ไม่พบ HoldingCode ในข้อความ, คีย์ที่มี: %v", getKeys(data))
	return ""
}

// getKeys helper function to list all keys in map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
