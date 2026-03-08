package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/segmentio/kafka-go"
)

// Kafka publisher สำหรับ MCP tools — ใช้ segmentio/kafka-go (pure Go, ไม่ต้อง CGO)
// เพื่อให้ MCP tools ทำงานเหมือน frontend → mainapi (MongoDB + Kafka sync)

var (
	kafkaBroker     string
	kafkaWriterPool map[string]*kafka.Writer
	kafkaPoolMu     sync.Mutex
	kafkaInitOnce   sync.Once
)

func initKafka() {
	kafkaInitOnce.Do(func() {
		kafkaBroker = os.Getenv("KAFKA_SERVER_URL")
		if kafkaBroker == "" {
			kafkaBroker = "kafka:29092" // default Docker internal
		}
		kafkaWriterPool = make(map[string]*kafka.Writer)
		logger.Info("[MCP Kafka] initialized broker=%s", kafkaBroker)
	})
}

// getWriter returns a cached writer for the given topic
func getWriter(topic string) *kafka.Writer {
	kafkaPoolMu.Lock()
	defer kafkaPoolMu.Unlock()

	if w, ok := kafkaWriterPool[topic]; ok {
		return w
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(kafkaBroker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		WriteTimeout: 10 * time.Second,
		Async:        false, // sync write เพื่อให้แน่ใจว่าส่งสำเร็จ
	}
	kafkaWriterPool[topic] = w
	return w
}

// publishToKafka ส่ง document ไป Kafka topic (JSON format เหมือน mainapi)
func publishToKafka(ctx context.Context, topic string, doc interface{}) error {
	initKafka()

	if kafkaBroker == "" {
		logger.Warn("[MCP Kafka] ไม่ได้กำหนด KAFKA_SERVER_URL — ข้าม publish")
		return nil
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal document สำหรับ Kafka ล้มเหลว: %w", err)
	}

	writer := getWriter(topic)
	err = writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
	if err != nil {
		logger.Error("[MCP Kafka] publish to %s ล้มเหลว: %v", topic, err)
		return fmt.Errorf("publish to Kafka topic '%s' ล้มเหลว: %w", topic, err)
	}

	logger.Info("[MCP Kafka] published to %s (%d bytes)", topic, len(data))
	return nil
}

// publishToKafkaBatch ส่ง documents หลายรายการไป Kafka topic เป็น JSON array ใน message เดียว
// (เหมือน mainapi ที่ส่ง CreateInBatch — consumer คาดหวัง JSON array)
func publishToKafkaBatch(ctx context.Context, topic string, docs []interface{}) error {
	initKafka()

	if kafkaBroker == "" {
		logger.Warn("[MCP Kafka] ไม่ได้กำหนด KAFKA_SERVER_URL — ข้าม publish")
		return nil
	}

	// ส่งเป็น JSON array ใน message เดียว (ไม่ใช่แยกหลาย messages)
	data, err := json.Marshal(docs)
	if err != nil {
		return fmt.Errorf("marshal documents สำหรับ Kafka ล้มเหลว: %w", err)
	}

	writer := getWriter(topic)
	err = writer.WriteMessages(ctx, kafka.Message{Value: data})
	if err != nil {
		logger.Error("[MCP Kafka] batch publish to %s ล้มเหลว: %v", topic, err)
		return fmt.Errorf("batch publish to Kafka topic '%s' ล้มเหลว: %w", topic, err)
	}

	logger.Info("[MCP Kafka] batch published %d docs (1 message) to %s", len(docs), topic)
	return nil
}

// ==================== Kafka Topic Constants ====================

// Barcode topics (ตรงกับ mainapi config)
const (
	TopicBarcodeCreated     = "when-product-barcode-created"
	TopicBarcodeUpdated     = "when-product-barcode-updated"
	TopicBarcodeDeleted     = "when-product-barcode-deleted"
	TopicBarcodeBulkCreated = "when-product-barcode-bulk-created"
	TopicBarcodeBulkUpdated = "when-product-barcode-bulk-updated"
	TopicBarcodeBulkDeleted = "when-product-barcode-bulk-deleted"
)

// Unit topics (ตรงกับ mainapi config)
const (
	TopicUnitCreated     = "when-product-unit-created"
	TopicUnitUpdated     = "when-product-unit-updated"
	TopicUnitDeleted     = "when-product-unit-deleted"
	TopicUnitBulkCreated = "when-product-unit-bulk-created"
	TopicUnitBulkUpdated = "when-product-unit-bulk-updated"
	TopicUnitBulkDeleted = "when-product-unit-bulk-deleted"
)

// Product Group topics (กลุ่มสินค้า)
const (
	TopicProductGroupCreated     = "when-product-group-created"
	TopicProductGroupUpdated     = "when-product-group-updated"
	TopicProductGroupDeleted     = "when-product-group-deleted"
	TopicProductGroupBulkCreated = "when-product-group-bulk-created"
	TopicProductGroupBulkDeleted = "when-product-group-bulk-deleted"
)

// Product Category topics (หมวดสินค้า)
const (
	TopicProductCategoryCreated     = "when-product-category-created"
	TopicProductCategoryUpdated     = "when-product-category-updated"
	TopicProductCategoryDeleted     = "when-product-category-deleted"
	TopicProductCategoryBulkCreated = "when-product-category-bulk-created"
	TopicProductCategoryBulkDeleted = "when-product-category-bulk-deleted"
)

// Creditor topics (เจ้าหนี้)
const (
	TopicCreditorCreated     = "when-creditor-created"
	TopicCreditorUpdated     = "when-creditor-updated"
	TopicCreditorDeleted     = "when-creditor-deleted"
	TopicCreditorBulkCreated = "when-creditor-bulk-created"
	TopicCreditorBulkDeleted = "when-creditor-bulk-deleted"
)

// Debtor topics (ลูกหนี้)
const (
	TopicDebtorCreated     = "when-debtor-created"
	TopicDebtorUpdated     = "when-debtor-updated"
	TopicDebtorDeleted     = "when-debtor-deleted"
	TopicDebtorBulkCreated = "when-debtor-bulk-created"
	TopicDebtorBulkDeleted = "when-debtor-bulk-deleted"
)
