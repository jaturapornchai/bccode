package mydb

import (
	"context"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/logger"
)

// PerformanceMetric เก็บข้อมูล performance ของแต่ละ query
type PerformanceMetric struct {
	Query string        `json:"query"`
	Database string        `json:"database"`
	Operation string        `json:"operation"`
	Duration time.Duration `json:"duration"`
	RowsAffected int64        `json:"rows_affected"`
	Status string        `json:"status"` // "success", "error"
	Error string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Context string        `json:"context,omitempty"` // เพิ่มเติมเพื่อระบุ context
}

// QueryPerformanceLogger interface สำหรับ logging performance
type QueryPerformanceLogger interface {
	LogQuery(ctx context.Context, query string, operation string, duration time.Duration, rowsAffected int64, status string, err error)
	LogBatchOperation(ctx context.Context, operation string, totalRows int, successRows int, errorRows int, duration time.Duration, err error)
}

// DatabasePerformanceLogger implementation ของ QueryPerformanceLogger
type DatabasePerformanceLogger struct {
	DatabaseName string
}

// NewDatabasePerformanceLogger สร้าง logger ใหม่
func NewDatabasePerformanceLogger(databaseName string) *DatabasePerformanceLogger {
	return &DatabasePerformanceLogger{
		DatabaseName: databaseName,
	}
}

// LogQuery บันทึก performance ของแต่ละ query
func (pl *DatabasePerformanceLogger) LogQuery(ctx context.Context, query string, operation string, duration time.Duration, rowsAffected int64, status string, err error) {
	metric := PerformanceMetric{
		Query:        pl.sanitizeQuery(query),
		Database:     pl.DatabaseName,
		Operation:    operation,
		Duration:     duration,
		RowsAffected: rowsAffected,
		Status:       status,
		Timestamp:    time.Now(),
		Context:      pl.getContextFromCtx(ctx),
	}
	
	if err != nil {
		metric.Error = err.Error()
	}
	
	// Log ตามระดับ performance
	pl.logPerformance(metric)
}

// LogBatchOperation บันทึก performance ของ batch operation
func (pl *DatabasePerformanceLogger) LogBatchOperation(ctx context.Context, operation string, totalRows int, successRows int, errorRows int, duration time.Duration, err error) {
	status := "success"
	if err != nil || errorRows > 0 {
		status = "error"
	}
	
	metric := PerformanceMetric{
		Query:        fmt.Sprintf("BATCH_%s", operation),
		Database:     pl.DatabaseName,
		Operation:    operation,
		Duration:     duration,
		RowsAffected: int64(successRows),
		Status:       status,
		Timestamp:    time.Now(),
		Context:      pl.getContextFromCtx(ctx),
	}
	
	if err != nil {
		metric.Error = err.Error()
	}
	
	pl.logPerformance(metric)
}

// logPerformance แสดงผล performance ตามระดับ
func (pl *DatabasePerformanceLogger) logPerformance(metric PerformanceMetric) {
	// แบ่งระดับตามเวลาในการ query
	duration := metric.Duration
	
	var level string
	var emoji string
	
	switch {
	case duration < 100*time.Millisecond:
		level = "debug"
		emoji = "⚡"
	case duration < 1*time.Second:
		level = "info"  
		emoji = "🔄"
	case duration < 5*time.Second:
		level = "warn"
		emoji = "🐌"
	default:
		level = "error"
		emoji = "⏰"
	}
	
	// สร้างข้อความ log
	logMessage := fmt.Sprintf("%s [PERF] %s.%s: %s (duration: %v, rows: %d, status: %s)",
		emoji, metric.Database, metric.Operation, metric.Context, duration.Round(time.Millisecond), metric.RowsAffected, metric.Status)
	
	if metric.Error != "" {
		logMessage += fmt.Sprintf(", error: %s", metric.Error)
	}
	
	// แสดงผลตามระดับ
	switch level {
	case "debug":
		logger.Debug(logMessage)
	case "info":
		logger.Info(logMessage)
	case "warn":
		logger.Warn(logMessage)
	case "error":
		logger.Error(logMessage)
	default:
		logger.Info(logMessage)
	}
	
	// Log summary สำหรับ slow query (เกิน 5 วินาที)
	if duration >= 5*time.Second {
		logger.Error("🔍 SLOW QUERY DETECTED - Database: %s, Operation: %s, Duration: %v", 
			metric.Database, metric.Operation, duration.Round(time.Millisecond))
		logger.Error("Query: %s", metric.Query)
		if metric.Error != "" {
			logger.Error("Error: %s", metric.Error)
		}
	}
}

// sanitizeQuery ทำความสะอาด query สำหรับการ log (ตัด parameters ออก)
func (pl *DatabasePerformanceLogger) sanitizeQuery(query string) string {
	// ตัด query ที่ยาวเกินไป
	if len(query) > 500 {
		return query[:500] + "... [TRUNCATED]"
	}
	return query
}

// getContextFromCtx ดึง context จาก context.Background()
func (pl *DatabasePerformanceLogger) getContextFromCtx(ctx context.Context) string {
	// สามารถเพิ่ม context value เพิ่มเติมได้ที่นี่
	// เช่น userID, requestID จาก context
	if ctx == nil {
		return "unknown"
	}
	
	// ตัวอย่าง: ดึง requestID จาก context (ถ้ามี)
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	
	return "default"
}

// QueryTimer สำหรับจับเวลา query execution
type QueryTimer struct {
	startTime time.Time
	query     string
	operation string
	database  string
	ctx       context.Context
	logger    *DatabasePerformanceLogger
}

// NewQueryTimer สร้าง timer ใหม่
func NewQueryTimer(ctx context.Context, query string, operation string, database string, logger *DatabasePerformanceLogger) *QueryTimer {
	return &QueryTimer{
		startTime: time.Now(),
		query:     query,
		operation: operation,
		database:  database,
		ctx:       ctx,
		logger:    logger,
	}
}

// End จบการจับเวลาและ log performance
func (qt *QueryTimer) End(rowsAffected int64, status string, err error) {
	duration := time.Since(qt.startTime)
	qt.logger.LogQuery(qt.ctx, qt.query, qt.operation, duration, rowsAffected, status, err)
}

// BatchQueryTimer สำหรับจับเวลา batch operations
type BatchQueryTimer struct {
	startTime   time.Time
	operation   string
	database    string
	ctx         context.Context
	logger      *DatabasePerformanceLogger
	totalRows   int
	successRows int
	errorRows   int
}

// NewBatchQueryTimer สร้าง batch timer ใหม่
func NewBatchQueryTimer(ctx context.Context, operation string, database string, logger *DatabasePerformanceLogger) *BatchQueryTimer {
	return &BatchQueryTimer{
		startTime: time.Now(),
		operation: operation,
		database:  database,
		ctx:       ctx,
		logger:    logger,
	}
}

// AddSuccess เพิ่มจำนวน success rows
func (bqt *BatchQueryTimer) AddSuccess() {
	bqt.successRows++
	bqt.totalRows++
}

// AddError เพิ่มจำนวน error rows  
func (bqt *BatchQueryTimer) AddError() {
	bqt.errorRows++
	bqt.totalRows++
}

// End จบการจับเวลา batch operation และ log performance
func (bqt *BatchQueryTimer) End(err error) {
	duration := time.Since(bqt.startTime)
	bqt.logger.LogBatchOperation(bqt.ctx, bqt.operation, bqt.totalRows, bqt.successRows, bqt.errorRows, duration, err)
}

// DatabasePerformanceStats สถิติ performance ของ database
type DatabasePerformanceStats struct {
	TotalQueries int64                   `json:"total_queries"`
	TotalErrors int64                   `json:"total_errors"`
	AverageDuration time.Duration           `json:"average_duration"`
	SlowQueries []PerformanceMetric     `json:"slow_queries,omitempty"`
	DatabaseName string                  `json:"database_name"`
}

// QueryStatsCollector รวบรวมสถิติ performance
type QueryStatsCollector struct {
	metrics []PerformanceMetric
	mutex   chan struct{} // ใช้แทน sync.Mutex เพื่อหลีกเลี่ยง allocation
}

// NewQueryStatsCollector สร้าง stats collector ใหม่
func NewQueryStatsCollector() *QueryStatsCollector {
	return &QueryStatsCollector{
		metrics: make([]PerformanceMetric, 0),
		mutex:   make(chan struct{}, 1), // buffered channel ขนาด 1
	}
}

// AddMetric เพิ่ม metric ใหม่
func (qsc *QueryStatsCollector) AddMetric(metric PerformanceMetric) {
	qsc.mutex <- struct{}{} // Lock
	qsc.metrics = append(qsc.metrics, metric)
	<-qsc.mutex // Unlock
}

// GetStats ดึงสถิติรวม
func (qsc *QueryStatsCollector) GetStats(databaseName string) DatabasePerformanceStats {
	qsc.mutex <- struct{}{} // Lock
	defer func() {
		<-qsc.mutex // Unlock
	}()
	
	var totalDuration time.Duration
	var totalQueries int64
	var totalErrors int64
	var slowQueries []PerformanceMetric
	
	for _, metric := range qsc.metrics {
		if metric.Database != databaseName {
			continue
		}
		
		totalQueries++
		totalDuration += metric.Duration
		
		if metric.Status == "error" {
			totalErrors++
		}
		
		// เก็บ slow queries (เกิน 1 วินาที)
		if metric.Duration >= 1*time.Second {
			slowQueries = append(slowQueries, metric)
		}
	}
	
	// จำกัดจำนวน slow queries ที่เก็บ (เก็บ 10 อันล่าสุด)
	if len(slowQueries) > 10 {
		slowQueries = slowQueries[len(slowQueries)-10:]
	}
	
	var avgDuration time.Duration
	if totalQueries > 0 {
		avgDuration = time.Duration(int64(totalDuration) / totalQueries)
	}
	
	return DatabasePerformanceStats{
		TotalQueries:    totalQueries,
		TotalErrors:     totalErrors,
		AverageDuration: avgDuration,
		SlowQueries:     slowQueries,
		DatabaseName:    databaseName,
	}
}