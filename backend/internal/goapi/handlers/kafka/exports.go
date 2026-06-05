package kafka

// This file serves as the main entry point for all Kafka-related functionality
// It exports the main functions that should be used by other packages

// Re-export the main consumer starter function
var StartKafkaConsumers = StartConsumers

// Re-export configuration
var NewKafkaConfiguration = NewKafkaConfig

// Re-export API handlers for testing
var (
	TestSaleInvoiceHandler     = TestSaleInvoiceConsumer
	TestSaleReturnHandler      = TestSaleReturnConsumer
	TestSaleOrderHandler       = TestSaleOrderConsumer
	TestInventoryHandler       = TestInventoryConsumer
	TestWarehouseHandler       = TestWarehouseConsumer
	TestPurchaseHandler        = TestPurchaseConsumer
	TestPurchaseOrderHandler   = TestPurchaseOrderConsumer
	TestPurchasePartialHandler = TestPurchasePartialConsumer
	GetStatusHandler           = GetConsumerStatus
)

// Re-export consumer functions for direct testing
var (
	ProcessSaleInvoiceMessage               = OnConsumeMessageSaleInvoiceCreateOrUpdate
	ProcessSaleInvoiceDeleteMessage         = OnConsumeMessageSaleInvoiceDelete
	ProcessSaleReturnMessage                = OnConsumeMessageSaleInvoiceReturnCreateOrUpdate
	ProcessSaleReturnDeleteMessage          = OnConsumeMessageSaleInvoiceReturnDelete
	ProcessSaleOrderMessage                 = OnConsumeMessageSaleOrderCreateOrUpdate
	ProcessSaleOrderDeleteMessage           = OnConsumeMessageSaleOrderDelete
	ProcessInventoryMessage                 = OnConsumeMessageInventoryCreateOrUpdate
	ProcessInventoryDeleteMessage           = OnConsumeMessageInventoryDelete
	ProcessInventoryBulkMessage             = OnConsumeMessageInventoryBulkCreateOrUpdate
	ProcessWarehouseMessage                 = OnConsumeMessageWarehouseCreateOrUpdate
	ProcessWarehouseDeleteMessage           = OnConsumeMessageWarehouseDelete
	ProcessPurchaseMessage                  = OnConsumeMessagePurchaseCreateOrUpdate
	ProcessPurchaseDeleteMessage            = OnConsumeMessagePurchaseDelete
	ProcessPurchaseOrderMessage             = OnConsumeMessagePurchaseOrderCreateOrUpdate
	ProcessPurchaseOrderDeleteMessage       = OnConsumeMessagePurchaseOrderDelete
	ProcessPurchasePartialMessage           = OnConsumeMessagePurchasePartialCreateOrUpdate
	ProcessPurchasePartialDeleteMessage     = OnConsumeMessagePurchasePartialDelete
	ProcessPurchaseReturnMessage            = OnConsumeMessagePurchaseReturnCreateOrUpdate
	ProcessPurchaseReturnDeleteMessage      = OnConsumeMessagePurchaseReturnDelete
	ProcessStockTransferMessage             = OnConsumeMessageStockTransferCreateOrUpdate
	ProcessStockTransferDeleteMessage       = OnConsumeMessageStockTransferDelete
	ProcessPurchaseRequisitionMessage       = OnConsumeMessagePurchaseRequisitionCreateOrUpdate
	ProcessPurchaseRequisitionDeleteMessage = OnConsumeMessagePurchaseRequisitionDelete
	ProcessRFQMessage                       = OnConsumeMessageRFQCreateOrUpdate
	ProcessRFQDeleteMessage                 = OnConsumeMessageRFQDelete
)

// Re-export utility functions
var (
	SafeWrapper = SafeConsumerWrapper
	ValidateMsg = ValidateMessage
	LogMsg      = LogMessageInfo
)

// Re-export constants for external use
const (
	// Topic constants
	SaleInvoiceCreateTopic = TOPIC_SALE_INVOICE_CREATE
	SaleInvoiceUpdateTopic = TOPIC_SALE_INVOICE_UPDATE
	SaleInvoiceDeleteTopic = TOPIC_SALE_INVOICE_DELETE
	SaleReturnCreateTopic  = TOPIC_SALE_INVOICE_RETURN_CREATE
	SaleReturnUpdateTopic  = TOPIC_SALE_INVOICE_RETURN_UPDATE
	SaleReturnDeleteTopic  = TOPIC_SALE_INVOICE_RETURN_DELETE

	SaleOrderCreateTopic = TOPIC_SALE_ORDER_CREATE
	SaleOrderUpdateTopic = TOPIC_SALE_ORDER_UPDATE
	SaleOrderDeleteTopic = TOPIC_SALE_ORDER_DELETE

	InventoryCreateTopic = TOPIC_INVENTORY_CREATE
	InventoryUpdateTopic = TOPIC_INVENTORY_UPDATE
	InventoryDeleteTopic = TOPIC_INVENTORY_DELETE

	WarehouseCreateTopic = TOPIC_WAREHOUSE_CREATE
	WarehouseUpdateTopic = TOPIC_WAREHOUSE_UPDATE
	WarehouseDeleteTopic = TOPIC_WAREHOUSE_DELETE
	// Consumer group constants
	SaleInvoiceConsumerGroup = CONSUMER_GROUP_SALE_INVOICE
	SaleReturnConsumerGroup  = CONSUMER_GROUP_SALE_RETURN
	SaleOrderConsumerGroup   = CONSUMER_GROUP_SALE_ORDER
	InventoryConsumerGroup   = CONSUMER_GROUP_INVENTORY
	WarehouseConsumerGroup   = CONSUMER_GROUP_WAREHOUSE

	// Transaction flags
	SaleInvoiceTransFlag = TRANS_FLAG_SALE_INVOICE
	SaleReturnTransFlag  = TRANS_FLAG_SALE_INVOICE_RETURN
	SaleOrderTransFlag   = TRANS_FLAG_SALE_ORDER

	// Batch size
	DefaultBatchSize = BATCH_SIZE
)
