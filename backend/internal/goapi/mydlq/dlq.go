package mydlq

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"time"
)

// DLQMessage โครงสร้าง message ที่ล้มเหลว
type DLQMessage struct {
	ID             string                 `json:"id"`
	OriginalTopic  string                 `json:"original_topic"`
	OriginalKey    string                 `json:"original_key"`
	OriginalValue  string                 `json:"original_value"`
	Error          string                 `json:"error"`
	RetryCount     int                    `json:"retry_count"`
	FirstFailedAt  time.Time              `json:"first_failed_at"`
	LastFailedAt   time.Time              `json:"last_failed_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// DLQHandler interface สำหรับจัดการ DLQ
type DLQHandler interface {
	Send(ctx context.Context, msg *DLQMessage) error
	Get(ctx context.Context, id string) (*DLQMessage, error)
	List(ctx context.Context, topic string, limit int) ([]*DLQMessage, error)
	Delete(ctx context.Context, id string) error
	Retry(ctx context.Context, id string) error
}

// FileDLQHandler เก็บ DLQ ลงไฟล์ (สำหรับ development/testing)
type FileDLQHandler struct {
	BasePath string
}

// NewFileDLQHandler สร้าง file-based DLQ handler
func NewFileDLQHandler(basePath string) *FileDLQHandler {
	return &FileDLQHandler{
		BasePath: basePath,
	}
}

func (h *FileDLQHandler) Send(ctx context.Context, msg *DLQMessage) error {
	// แปลงเป็น JSON
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}

	// บันทึกลงไฟล์
	filename := fmt.Sprintf("%s/dlq_%s_%s.json", h.BasePath, msg.OriginalTopic, msg.ID)
	logger.Warn("📝 Saving DLQ message to: %s", filename)
	logger.Warn("DLQ Message: %s", string(data))

	// TODO: เขียนลงไฟล์จริง (ถ้าต้องการ)
	// ioutil.WriteFile(filename, data, 0644)

	return nil
}

func (h *FileDLQHandler) Get(ctx context.Context, id string) (*DLQMessage, error) {
	// TODO: อ่านจากไฟล์
	return nil, fmt.Errorf("not implemented")
}

func (h *FileDLQHandler) List(ctx context.Context, topic string, limit int) ([]*DLQMessage, error) {
	// TODO: list จากไฟล์
	return nil, fmt.Errorf("not implemented")
}

func (h *FileDLQHandler) Delete(ctx context.Context, id string) error {
	// TODO: ลบไฟล์
	return fmt.Errorf("not implemented")
}

func (h *FileDLQHandler) Retry(ctx context.Context, id string) error {
	// TODO: อ่านจากไฟล์และส่งกลับไปยัง Kafka
	return fmt.Errorf("not implemented")
}

// LogDLQHandler เขียน DLQ ลง log (simple, ไม่เก็บจริง)
type LogDLQHandler struct{}

// NewLogDLQHandler สร้าง log-based DLQ handler
func NewLogDLQHandler() *LogDLQHandler {
	return &LogDLQHandler{}
}

func (h *LogDLQHandler) Send(ctx context.Context, msg *DLQMessage) error {
	data, _ := json.Marshal(msg)
	logger.Error("🔴 DLQ MESSAGE: %s", string(data))
	logger.Error("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Error("Topic: %s", msg.OriginalTopic)
	logger.Error("Key: %s", msg.OriginalKey)
	logger.Error("Error: %s", msg.Error)
	logger.Error("Retry Count: %d", msg.RetryCount)
	logger.Error("First Failed: %s", msg.FirstFailedAt.Format(time.RFC3339))
	logger.Error("Last Failed: %s", msg.LastFailedAt.Format(time.RFC3339))
	logger.Error("Value: %s", msg.OriginalValue[:min(len(msg.OriginalValue), 500)]) // แสดงแค่ 500 ตัวอักษรแรก
	logger.Error("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

func (h *LogDLQHandler) Get(ctx context.Context, id string) (*DLQMessage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *LogDLQHandler) List(ctx context.Context, topic string, limit int) ([]*DLQMessage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *LogDLQHandler) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (h *LogDLQHandler) Retry(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Global DLQ handler (ใช้ LogDLQHandler เป็น default)
var globalDLQHandler DLQHandler = NewLogDLQHandler()

// SetDLQHandler ตั้งค่า DLQ handler
func SetDLQHandler(handler DLQHandler) {
	globalDLQHandler = handler
}

// GetDLQHandler ดึง DLQ handler ปัจจุบัน
func GetDLQHandler() DLQHandler {
	return globalDLQHandler
}

// SendToDLQ ส่ง message ไปยัง DLQ
func SendToDLQ(ctx context.Context, topic string, key string, value string, err error, retryCount int, metadata map[string]interface{}) error {
	msg := &DLQMessage{
		ID:             fmt.Sprintf("%d", time.Now().UnixNano()),
		OriginalTopic:  topic,
		OriginalKey:    key,
		OriginalValue:  value,
		Error:          err.Error(),
		RetryCount:     retryCount,
		FirstFailedAt:  time.Now(),
		LastFailedAt:   time.Now(),
		Metadata:       metadata,
	}

	return globalDLQHandler.Send(ctx, msg)
}

// QuickSendToDLQ ส่ง message ไปยัง DLQ แบบง่าย
func QuickSendToDLQ(topic string, message string, err error) error {
	return SendToDLQ(context.Background(), topic, "", message, err, 0, nil)
}
