package kafka

import (
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"smlcloudplatform/internal/goapi/mykafkaconsumer"
)

// consumerGroup ผูกหัวข้อข่าวสารเข้ากับตัวจัดการของมัน
//
// เขียนเป็นตารางเพื่อให้เห็นครบในที่เดียวว่าเอกสารประเภทใดถูกรับฟังอยู่บ้าง
// ก่อนหน้านี้การสมัครรับกระจายอยู่ในฟังก์ชันแยก ทำให้เอกสารเคลื่อนไหวสต็อกทั้งหมด
// (โอนคลัง ปรับสต็อก รับ-เบิก-คืนสินค้า ยอดยกมา ซื้อรับ ซื้อคืน) ถูกลืมไปโดยไม่มีใครสังเกต
// ทั้งที่ตัวจัดการเขียนไว้ครบแล้ว เอกสารเหล่านั้นจึงไม่เคยไหลเข้าตาราง docdetail เลย
type consumerGroup struct {
	name             string
	group            string
	createTopic      string
	updateTopic      string
	deleteTopic      string
	onCreateOrUpdate func(string) error
	onDelete         func(string) error
	// movesStock บอกว่าเอกสารประเภทนี้กระทบยอดสต็อกและต้นทุนหรือไม่
	movesStock bool
}

// consumerGroups คือรายการที่ระบบรับฟังทั้งหมด
func consumerGroups() []consumerGroup {
	return []consumerGroup{
		{
			name: "Sale Invoice", group: CONSUMER_GROUP_SALE_INVOICE, movesStock: true,
			createTopic: TOPIC_SALE_INVOICE_CREATE, updateTopic: TOPIC_SALE_INVOICE_UPDATE, deleteTopic: TOPIC_SALE_INVOICE_DELETE,
			onCreateOrUpdate: OnConsumeMessageSaleInvoiceCreateOrUpdate, onDelete: OnConsumeMessageSaleInvoiceDelete,
		},
		{
			name: "Sale Return", group: CONSUMER_GROUP_SALE_RETURN, movesStock: true,
			createTopic: TOPIC_SALE_INVOICE_RETURN_CREATE, updateTopic: TOPIC_SALE_INVOICE_RETURN_UPDATE, deleteTopic: TOPIC_SALE_INVOICE_RETURN_DELETE,
			onCreateOrUpdate: OnConsumeMessageSaleInvoiceReturnCreateOrUpdate, onDelete: OnConsumeMessageSaleInvoiceReturnDelete,
		},
		{
			name: "Sale Order", group: CONSUMER_GROUP_SALE_ORDER,
			createTopic: TOPIC_SALE_ORDER_CREATE, updateTopic: TOPIC_SALE_ORDER_UPDATE, deleteTopic: TOPIC_SALE_ORDER_DELETE,
			onCreateOrUpdate: OnConsumeMessageSaleOrderCreateOrUpdate, onDelete: OnConsumeMessageSaleOrderDelete,
		},
		{
			name: "Purchase", group: CONSUMER_GROUP_PURCHASE, movesStock: true,
			createTopic: TOPIC_PURCHASE_CREATE, updateTopic: TOPIC_PURCHASE_UPDATE, deleteTopic: TOPIC_PURCHASE_DELETE,
			onCreateOrUpdate: OnConsumeMessagePurchaseCreateOrUpdate, onDelete: OnConsumeMessagePurchaseDelete,
		},
		{
			name: "Purchase Order", group: CONSUMER_GROUP_PURCHASE_ORDER,
			createTopic: TOPIC_PURCHASE_ORDER_CREATE, updateTopic: TOPIC_PURCHASE_ORDER_UPDATE, deleteTopic: TOPIC_PURCHASE_ORDER_DELETE,
			onCreateOrUpdate: OnConsumeMessagePurchaseOrderCreateOrUpdate, onDelete: OnConsumeMessagePurchaseOrderDelete,
		},
		{
			name: "Purchase Partial", group: CONSUMER_GROUP_PURCHASE_PARTIAL, movesStock: true,
			createTopic: TOPIC_PURCHASE_PARTIAL_CREATE, updateTopic: TOPIC_PURCHASE_PARTIAL_UPDATE, deleteTopic: TOPIC_PURCHASE_PARTIAL_DELETE,
			onCreateOrUpdate: OnConsumeMessagePurchasePartialCreateOrUpdate, onDelete: OnConsumeMessagePurchasePartialDelete,
		},
		{
			name: "Purchase Return", group: CONSUMER_GROUP_PURCHASE_RETURN, movesStock: true,
			createTopic: TOPIC_PURCHASE_RETURN_CREATE, updateTopic: TOPIC_PURCHASE_RETURN_UPDATE, deleteTopic: TOPIC_PURCHASE_RETURN_DELETE,
			onCreateOrUpdate: OnConsumeMessagePurchaseReturnCreateOrUpdate, onDelete: OnConsumeMessagePurchaseReturnDelete,
		},
		{
			name: "Purchase Requisition", group: CONSUMER_GROUP_PURCHASE_REQUISITION,
			createTopic: TOPIC_PURCHASE_REQUISITION_CREATE, updateTopic: TOPIC_PURCHASE_REQUISITION_UPDATE, deleteTopic: TOPIC_PURCHASE_REQUISITION_DELETE,
			onCreateOrUpdate: OnConsumeMessagePurchaseRequisitionCreateOrUpdate, onDelete: OnConsumeMessagePurchaseRequisitionDelete,
		},
		{
			name: "RFQ", group: CONSUMER_GROUP_RFQ,
			createTopic: TOPIC_RFQ_CREATE, updateTopic: TOPIC_RFQ_UPDATE, deleteTopic: TOPIC_RFQ_DELETE,
			onCreateOrUpdate: OnConsumeMessageRFQCreateOrUpdate, onDelete: OnConsumeMessageRFQDelete,
		},
		{
			name: "Stock Transfer", group: CONSUMER_GROUP_STOCK_TRANSFER, movesStock: true,
			createTopic: TOPIC_STOCK_TRANSFER_CREATE, updateTopic: TOPIC_STOCK_TRANSFER_UPDATE, deleteTopic: TOPIC_STOCK_TRANSFER_DELETE,
			onCreateOrUpdate: OnConsumeMessageStockTransferCreateOrUpdate, onDelete: OnConsumeMessageStockTransferDelete,
		},
		{
			name: "Stock Receive Product", group: CONSUMER_GROUP_STOCK_RECEIVE_PRODUCT, movesStock: true,
			createTopic: TOPIC_STOCK_RECEIVE_PRODUCT_CREATE, updateTopic: TOPIC_STOCK_RECEIVE_PRODUCT_UPDATE, deleteTopic: TOPIC_STOCK_RECEIVE_PRODUCT_DELETE,
			onCreateOrUpdate: OnConsumeMessageStockReceiveProductCreateOrUpdate, onDelete: OnConsumeMessageStockReceiveProductDelete,
		},
		{
			name: "Stock Pickup Product", group: CONSUMER_GROUP_STOCK_PICKUP_PRODUCT, movesStock: true,
			createTopic: TOPIC_STOCK_PICKUP_PRODUCT_CREATE, updateTopic: TOPIC_STOCK_PICKUP_PRODUCT_UPDATE, deleteTopic: TOPIC_STOCK_PICKUP_PRODUCT_DELETE,
			onCreateOrUpdate: OnConsumeMessageStockPickupProductCreateOrUpdate, onDelete: OnConsumeMessageStockPickupProductDelete,
		},
		{
			name: "Stock Return Product", group: CONSUMER_GROUP_STOCK_RETURN_PRODUCT, movesStock: true,
			createTopic: TOPIC_STOCK_RETURN_PRODUCT_CREATE, updateTopic: TOPIC_STOCK_RETURN_PRODUCT_UPDATE, deleteTopic: TOPIC_STOCK_RETURN_PRODUCT_DELETE,
			onCreateOrUpdate: OnConsumeMessageStockReturnProductCreateOrUpdate, onDelete: OnConsumeMessageStockReturnProductDelete,
		},
		{
			name: "Stock Adjustment", group: CONSUMER_GROUP_STOCK_ADJUSTMENT, movesStock: true,
			createTopic: TOPIC_STOCK_ADJUSTMENT_CREATE, updateTopic: TOPIC_STOCK_ADJUSTMENT_UPDATE, deleteTopic: TOPIC_STOCK_ADJUSTMENT_DELETE,
			onCreateOrUpdate: OnConsumeMessageStockAdjustmentCreateOrUpdate, onDelete: OnConsumeMessageStockAdjustmentDelete,
		},
		{
			name: "Stock Balance", group: CONSUMER_GROUP_STOCK_BALANCE, movesStock: true,
			createTopic: TOPIC_STOCK_BALANCE_CREATE, updateTopic: TOPIC_STOCK_BALANCE_UPDATE, deleteTopic: TOPIC_STOCK_BALANCE_DELETE,
			onCreateOrUpdate: OnConsumeMessageStockBalanceCreateOrUpdate, onDelete: OnConsumeMessageStockBalanceDelete,
		},
		{
			name: "Warehouse", group: CONSUMER_GROUP_WAREHOUSE,
			createTopic: TOPIC_WAREHOUSE_CREATE, updateTopic: TOPIC_WAREHOUSE_UPDATE, deleteTopic: TOPIC_WAREHOUSE_DELETE,
			onCreateOrUpdate: OnConsumeMessageWarehouseCreateOrUpdate, onDelete: OnConsumeMessageWarehouseDelete,
		},
		{
			name: "Inventory", group: CONSUMER_GROUP_INVENTORY,
			createTopic: TOPIC_INVENTORY_CREATE, updateTopic: TOPIC_INVENTORY_UPDATE, deleteTopic: TOPIC_INVENTORY_DELETE,
			onCreateOrUpdate: OnConsumeMessageInventoryCreateOrUpdate, onDelete: OnConsumeMessageInventoryDelete,
		},
	}
}

func StartConsumers() {
	config := NewKafkaConfig()
	if !config.IsValid() {
		logger.Warn("⚠️  KAFKA_SERVER_URL environment variable not set")
		return
	}

	logger.Info("🔌 Attempting to connect to Kafka broker: %s", config.ServerURL)
	logger.Info("📡 Starting Kafka consumers (เริ่มทีละ group เพื่อความเสถียร)...")

	consumers := mykafkaconsumer.NewKafkaConsumer().(*mykafkaconsumer.KafkaConsumer)

	groups := consumerGroups()
	for _, group := range groups {
		logger.Info("📊 เริ่ม %s consumers...", group.name)
		startGroup(consumers, config.ServerURL, group)
		time.Sleep(500 * time.Millisecond) // เว้นจังหวะระหว่าง group เพื่อความเสถียรตอนเริ่มระบบ
	}

	// สินค้าแบบเป็นชุดใช้กลุ่มแยก เพราะข้อความหนึ่งครอบคลุมหลายรายการและใช้เวลาต่างกันมาก
	//
	// หัวข้อบาร์โค้ดถูกรับฟังโดยกลุ่มสินค้าเท่านั้น เคยมีสำเนาซ้ำในกลุ่มคลังสินค้า
	// ซึ่งทำให้เขียนซ้ำทุกครั้ง และเมื่อบาร์โค้ดตัวหน้าติดขัด ผู้อ่านคลังสินค้าที่ไม่เกี่ยวกันก็ถูกกระทบไปด้วย
	logger.Info("📊 เริ่ม Inventory (bulk) consumers...")
	go consumers.ConsumeMessage(config.ServerURL, MQ_TOPIC_BULK_CREATED, CONSUMER_GROUP_INVENTORY_BULK, 0, OnConsumeMessageInventoryBulkCreateOrUpdate)
	time.Sleep(200 * time.Millisecond)
	go consumers.ConsumeMessage(config.ServerURL, MQ_TOPIC_BULK_UPDATED, CONSUMER_GROUP_INVENTORY_BULK, 0, OnConsumeMessageInventoryBulkCreateOrUpdate)

	logger.Info("✅ All Kafka consumers started successfully")
	logger.Info("📊 Total consumer groups: %d", len(groups)+1)
}

func startGroup(consumers *mykafkaconsumer.KafkaConsumer, server string, group consumerGroup) {
	go consumers.ConsumeMessage(server, group.createTopic, group.group, 0, group.onCreateOrUpdate)
	time.Sleep(200 * time.Millisecond) // เว้นจังหวะระหว่างหัวข้อใน group เดียวกัน
	go consumers.ConsumeMessage(server, group.updateTopic, group.group, 0, group.onCreateOrUpdate)
	time.Sleep(200 * time.Millisecond)
	go consumers.ConsumeMessage(server, group.deleteTopic, group.group, 0, group.onDelete)
}
