package build

import (
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RebuildProgress — โครงสร้าง progress event สำหรับ SSE
type RebuildProgress struct {
	JobID      string  `json:"jobid"`
	Step       int     `json:"step"`       // step ปัจจุบัน (1-based)
	TotalSteps int     `json:"totalsteps"` // จำนวน steps ทั้งหมด
	StepName   string  `json:"stepname"`   // ชื่อ step (ภาษาไทย)
	Status     string  `json:"status"`     // "running", "completed", "error"
	Message    string  `json:"message"`    // ข้อความเพิ่มเติม
	Progress   float64 `json:"progress"`   // 0.0 - 1.0
	Detail     string  `json:"detail"`     // รายละเอียดย่อย เช่น "150/500 รายการ"
}

// RebuildJob — เก็บสถานะ job ใน memory
type RebuildJob struct {
	ID          string
	HoldingCode string
	EventsCh    chan RebuildProgress // ส่ง events ให้ SSE handler
	Done        chan struct{}
	CreatedAt   time.Time
}

// In-memory job store
var (
	rebuildJobs   = make(map[string]*RebuildJob)
	rebuildJobsMu sync.RWMutex
)

// CreateJob — สร้าง rebuild job ใหม่ return job
func CreateJob(holdingCode string) *RebuildJob {
	job := &RebuildJob{
		ID:          uuid.New().String(),
		HoldingCode: holdingCode,
		EventsCh:    make(chan RebuildProgress, 200),
		Done:        make(chan struct{}),
		CreatedAt:   time.Now(),
	}

	rebuildJobsMu.Lock()
	rebuildJobs[job.ID] = job
	rebuildJobsMu.Unlock()

	logger.Info("[Rebuild] สร้าง job %s สำหรับ shop %s", job.ID, holdingCode)
	return job
}

// GetJob — ดึง job จาก map
func GetJob(jobID string) *RebuildJob {
	rebuildJobsMu.RLock()
	defer rebuildJobsMu.RUnlock()
	return rebuildJobs[jobID]
}

// RemoveJob — ลบ job หลังเสร็จ (เรียกจาก defer)
func RemoveJob(jobID string) {
	rebuildJobsMu.Lock()
	job, exists := rebuildJobs[jobID]
	if exists {
		close(job.EventsCh)
		close(job.Done)
		delete(rebuildJobs, jobID)
	}
	rebuildJobsMu.Unlock()

	if exists {
		logger.Info("[Rebuild] ลบ job %s สำเร็จ", jobID)
	}
}

// SendProgress — helper ส่ง progress event ไปที่ channel (non-blocking)
func (job *RebuildJob) SendProgress(step, totalSteps int, stepName, status string) {
	job.SendProgressWithMessage(step, totalSteps, stepName, status, "")
}

// SendProgressWithMessage — ส่ง progress event พร้อมข้อความเพิ่มเติม
func (job *RebuildJob) SendProgressWithMessage(step, totalSteps int, stepName, status, message string) {
	progress := float64(step) / float64(totalSteps)
	if status == "completed" {
		progress = 1.0
	}

	event := RebuildProgress{
		JobID:      job.ID,
		Step:       step,
		TotalSteps: totalSteps,
		StepName:   stepName,
		Status:     status,
		Message:    message,
		Progress:   progress,
	}

	// สำหรับ completed/error — ส่งแบบ blocking (ห้าม drop เพราะเป็น event สำคัญ)
	if status == "completed" || status == "error" {
		select {
		case job.EventsCh <- event:
			logger.Info("[Rebuild] ส่ง %s event: step %d/%d — %s", status, step, totalSteps, stepName)
		case <-time.After(10 * time.Second):
			logger.Warn("[Rebuild] ไม่สามารถส่ง %s event ได้ (timeout 10s)", status)
		}
		return
	}

	// Non-blocking send — ถ้า channel เต็ม (ไม่มีคนฟัง) ก็ข้ามไป
	select {
	case job.EventsCh <- event:
		logger.Debug("[Rebuild] ส่ง progress: step %d/%d — %s (%s)", step, totalSteps, stepName, status)
	default:
		logger.Debug("[Rebuild] ข้าม progress event (channel เต็ม): step %d/%d", step, totalSteps)
	}
}

// SendProgressDetail — ส่ง progress event พร้อมรายละเอียดย่อย (เช่น "150/500 รายการ")
func (job *RebuildJob) SendProgressDetail(step, totalSteps int, stepName, detail string) {
	progress := float64(step) / float64(totalSteps)

	event := RebuildProgress{
		JobID:      job.ID,
		Step:       step,
		TotalSteps: totalSteps,
		StepName:   stepName,
		Status:     "running",
		Detail:     detail,
		Progress:   progress,
	}

	// Non-blocking send
	select {
	case job.EventsCh <- event:
	default:
	}
}

// SendError — ส่ง error event
func (job *RebuildJob) SendError(step, totalSteps int, stepName string, err error) {
	job.SendProgressWithMessage(step, totalSteps, stepName, "error", fmt.Sprintf("ล้มเหลว: %v", err))
}

// CleanupOldJobs — ลบ jobs ที่ค้างเกิน 30 นาที (เรียกเป็น periodic task ได้)
func CleanupOldJobs() {
	rebuildJobsMu.Lock()
	defer rebuildJobsMu.Unlock()

	now := time.Now()
	for id, job := range rebuildJobs {
		if now.Sub(job.CreatedAt) > 30*time.Minute {
			close(job.EventsCh)
			close(job.Done)
			delete(rebuildJobs, id)
			logger.Info("[Rebuild] ลบ job ค้าง %s (เกิน 30 นาที)", id)
		}
	}
}
