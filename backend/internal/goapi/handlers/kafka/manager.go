package kafka

import (
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"smlcloudplatform/internal/goapi/mykafkaconsumer"
)

func StartConsumers() {
	config := NewKafkaConfig()
	if !config.IsValid() {
		logger.Warn("⚠️  KAFKA_SERVER_URL environment variable not set")
		return
	}

	logger.Info("🔌 Attempting to connect to Kafka broker: %s", config.ServerURL)
	logger.Info("📡 Starting Kafka consumers (เริ่มทีละ group เพื่อความเสถียร)...")

	consumers := mykafkaconsumer.NewKafkaConsumer()

	// Start all consumer groups ทีละตัวพร้อม delay เล็กน้อย เพื่อความเสถียร
	logger.Info("📊 เริ่ม Sale Invoice consumers...")
	startSaleInvoiceConsumers(consumers.(*mykafkaconsumer.KafkaConsumer), config.ServerURL)
	time.Sleep(500 * time.Millisecond) // Delay ระหว่าง groups

	logger.Info("📊 เริ่ม Sale Return consumers...")
	startSaleReturnConsumers(consumers.(*mykafkaconsumer.KafkaConsumer), config.ServerURL)
	time.Sleep(500 * time.Millisecond)

	logger.Info("📊 เริ่ม Sale Order consumers...")
	startSaleOrderConsumers(consumers.(*mykafkaconsumer.KafkaConsumer), config.ServerURL)
	time.Sleep(500 * time.Millisecond)

	logger.Info("📊 เริ่ม Inventory consumers...")
	startInventoryConsumers(consumers.(*mykafkaconsumer.KafkaConsumer), config.ServerURL)
	time.Sleep(500 * time.Millisecond)

	logger.Info("📊 เริ่ม Warehouse consumers...")
	startWarehouseConsumers(consumers.(*mykafkaconsumer.KafkaConsumer), config.ServerURL)
	time.Sleep(500 * time.Millisecond)

	logger.Info("📊 เริ่ม Purchase consumers...")
	startPurchaseConsumers(consumers.(*mykafkaconsumer.KafkaConsumer), config.ServerURL)

	logger.Info("✅ All Kafka consumers started successfully")
	logger.Info("📊 Total consumer groups: 6 (sale-invoice, sale-return, sale-order, inventory, warehouse, purchase)")
}

// startSaleInvoiceConsumers - starts sale invoice consumers
func startSaleInvoiceConsumers(consumers *mykafkaconsumer.KafkaConsumer, kafkaServer string) {
	logger.Info("Starting Sale Invoice consumers...")
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_INVOICE_CREATE, CONSUMER_GROUP_SALE_INVOICE, 0, OnConsumeMessageSaleInvoiceCreateOrUpdate)
	time.Sleep(200 * time.Millisecond) // Delay ระหว่าง topics ใน group เดียวกัน
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_INVOICE_UPDATE, CONSUMER_GROUP_SALE_INVOICE, 0, OnConsumeMessageSaleInvoiceCreateOrUpdate)
	time.Sleep(200 * time.Millisecond)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_INVOICE_DELETE, CONSUMER_GROUP_SALE_INVOICE, 0, OnConsumeMessageSaleInvoiceDelete)
}

// startSaleReturnConsumers - starts sale return consumers
func startSaleReturnConsumers(consumers *mykafkaconsumer.KafkaConsumer, kafkaServer string) {
	logger.Info("Starting Sale Return consumers...")
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_INVOICE_RETURN_CREATE, CONSUMER_GROUP_SALE_RETURN, 0, OnConsumeMessageSaleInvoiceReturnCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_INVOICE_RETURN_UPDATE, CONSUMER_GROUP_SALE_RETURN, 0, OnConsumeMessageSaleInvoiceReturnCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_INVOICE_RETURN_DELETE, CONSUMER_GROUP_SALE_RETURN, 0, OnConsumeMessageSaleInvoiceReturnDelete)
}

// startSaleOrderConsumers - starts sale order consumers
func startSaleOrderConsumers(consumers *mykafkaconsumer.KafkaConsumer, kafkaServer string) {
	logger.Info("Starting Sale Order consumers...")
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_ORDER_CREATE, CONSUMER_GROUP_SALE_ORDER, 0, OnConsumeMessageSaleOrderCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_ORDER_UPDATE, CONSUMER_GROUP_SALE_ORDER, 0, OnConsumeMessageSaleOrderCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_SALE_ORDER_DELETE, CONSUMER_GROUP_SALE_ORDER, 0, OnConsumeMessageSaleOrderDelete)
}

// startInventoryConsumers - starts inventory consumers
func startInventoryConsumers(consumers *mykafkaconsumer.KafkaConsumer, kafkaServer string) {
	logger.Info("Starting Inventory consumers...")

	// Regular inventory consumers
	go consumers.ConsumeMessage(kafkaServer, TOPIC_INVENTORY_CREATE, CONSUMER_GROUP_INVENTORY, 0, OnConsumeMessageInventoryCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_INVENTORY_UPDATE, CONSUMER_GROUP_INVENTORY, 0, OnConsumeMessageInventoryCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_INVENTORY_DELETE, CONSUMER_GROUP_INVENTORY, 0, OnConsumeMessageInventoryDelete)

	// Bulk inventory consumers
	go consumers.ConsumeMessage(kafkaServer, MQ_TOPIC_BULK_CREATED, CONSUMER_GROUP_INVENTORY_BULK, 0, OnConsumeMessageInventoryBulkCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, MQ_TOPIC_BULK_UPDATED, CONSUMER_GROUP_INVENTORY_BULK, 0, OnConsumeMessageInventoryBulkCreateOrUpdate)
	// Note: Bulk delete is commented out in original code
	// go consumers.ConsumeMessage(kafkaServer, MQ_TOPIC_BULK_DELETED, CONSUMER_GROUP_INVENTORY_BULK, 0, OnConsumeMessageInventoryBulkDelete)

	// Barcode topics are reconciled once, by CONSUMER_GROUP_INVENTORY only. A second
	// copy in CONSUMER_GROUP_WAREHOUSE duplicated every write and let a blocked
	// barcode head rebalance the unrelated warehouse readers on each retry.
}

// startWarehouseConsumers - starts warehouse consumers
func startWarehouseConsumers(consumers *mykafkaconsumer.KafkaConsumer, kafkaServer string) {
	logger.Info("Starting Warehouse consumers...")
	go consumers.ConsumeMessage(kafkaServer, TOPIC_WAREHOUSE_CREATE, CONSUMER_GROUP_WAREHOUSE, 0, OnConsumeMessageWarehouseCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_WAREHOUSE_UPDATE, CONSUMER_GROUP_WAREHOUSE, 0, OnConsumeMessageWarehouseCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_WAREHOUSE_DELETE, CONSUMER_GROUP_WAREHOUSE, 0, OnConsumeMessageWarehouseDelete)
}

// startPurchaseConsumers - starts purchase-related consumers
func startPurchaseConsumers(consumers *mykafkaconsumer.KafkaConsumer, kafkaServer string) {
	logger.Info("Starting Purchase consumers...")

	// Purchase consumers
	go consumers.ConsumeMessage(kafkaServer, TOPIC_PURCHASE_CREATE, CONSUMER_GROUP_PURCHASE, 0, OnConsumeMessagePurchaseCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_PURCHASE_UPDATE, CONSUMER_GROUP_PURCHASE, 0, OnConsumeMessagePurchaseCreateOrUpdate)
	go consumers.ConsumeMessage(kafkaServer, TOPIC_PURCHASE_DELETE, CONSUMER_GROUP_PURCHASE, 0, OnConsumeMessagePurchaseDelete)
}
