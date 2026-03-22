package kafka

// Kafka topic constants
const (
	// Sale Invoice topics
	TOPIC_SALE_INVOICE_CREATE = "when-saleinvoice-created"
	TOPIC_SALE_INVOICE_UPDATE = "when-saleinvoice-updated"
	TOPIC_SALE_INVOICE_DELETE = "when-saleinvoice-deleted"

	// Sale Invoice Return topics
	TOPIC_SALE_INVOICE_RETURN_CREATE = "when-saleinvoicereturn-created"
	TOPIC_SALE_INVOICE_RETURN_UPDATE = "when-saleinvoicereturn-updated"
	TOPIC_SALE_INVOICE_RETURN_DELETE = "when-saleinvoicereturn-deleted"

	// Sale Order topics
	TOPIC_SALE_ORDER_CREATE = "when-saleorder-created"
	TOPIC_SALE_ORDER_UPDATE = "when-saleorder-updated"
	TOPIC_SALE_ORDER_DELETE = "when-saleorder-deleted"

	// Inventory/Product topics
	TOPIC_INVENTORY_CREATE = "when-product-barcode-created"
	TOPIC_INVENTORY_UPDATE = "when-product-barcode-updated"
	TOPIC_INVENTORY_DELETE = "when-product-barcode-deleted"
	MQ_TOPIC_BULK_CREATED  = "when-product-barcode-bulk-created"
	MQ_TOPIC_BULK_UPDATED  = "when-product-barcode-bulk-updated"
	MQ_TOPIC_BULK_DELETED  = "when-product-barcode-bulk-deleted"

	// Warehouse topics
	TOPIC_WAREHOUSE_CREATE = "when-warehouse-created"
	TOPIC_WAREHOUSE_UPDATE = "when-warehouse-updated"
	TOPIC_WAREHOUSE_DELETE = "when-warehouse-deleted"

	// Purchase Order topics
	TOPIC_PURCHASE_ORDER_CREATE = "when-purchaseorder-created"
	TOPIC_PURCHASE_ORDER_UPDATE = "when-purchaseorder-updated"
	TOPIC_PURCHASE_ORDER_DELETE = "when-purchaseorder-deleted"

	// Purchase topics
	TOPIC_PURCHASE_CREATE = "when-purchase-created"
	TOPIC_PURCHASE_UPDATE = "when-purchase-updated"
	TOPIC_PURCHASE_DELETE = "when-purchase-deleted"

	// Purchase Receive topics
	TOPIC_PURCHASE_RECEIVE_CREATE = "when-purchasereceive-created"
	TOPIC_PURCHASE_RECEIVE_UPDATE = "when-purchasereceive-updated"
	TOPIC_PURCHASE_RECEIVE_DELETE = "when-purchasereceive-deleted"

	// Purchase Partial topics
	TOPIC_PURCHASE_PARTIAL_CREATE = "when-purchasepartial-created"
	TOPIC_PURCHASE_PARTIAL_UPDATE = "when-purchasepartial-updated"
	TOPIC_PURCHASE_PARTIAL_DELETE = "when-purchasepartial-deleted"
	// Purchase Return topics
	TOPIC_PURCHASE_RETURN_CREATE = "when-purchasereturn-created"
	TOPIC_PURCHASE_RETURN_UPDATE = "when-purchasereturn-updated"
	TOPIC_PURCHASE_RETURN_DELETE = "when-purchasereturn-deleted"
	// Stock Transfer topics
	TOPIC_STOCK_TRANSFER_CREATE = "when-stocktransfer-created"
	TOPIC_STOCK_TRANSFER_UPDATE = "when-stocktransfer-updated"
	TOPIC_STOCK_TRANSFER_DELETE = "when-stocktransfer-deleted"
	// Stock Receive Product topics
	TOPIC_STOCK_RECEIVE_PRODUCT_CREATE = "when-stockreceiveproduct-created"
	TOPIC_STOCK_RECEIVE_PRODUCT_UPDATE = "when-stockreceiveproduct-updated"
	TOPIC_STOCK_RECEIVE_PRODUCT_DELETE = "when-stockreceiveproduct-deleted"
	// Stock Pickup Product topics
	TOPIC_STOCK_PICKUP_PRODUCT_CREATE = "when-stockpickupproduct-created"
	TOPIC_STOCK_PICKUP_PRODUCT_UPDATE = "when-stockpickupproduct-updated"
	TOPIC_STOCK_PICKUP_PRODUCT_DELETE = "when-stockpickupproduct-deleted"

	// Stock Return Product topics
	TOPIC_STOCK_RETURN_PRODUCT_CREATE = "when-stockreturnproduct-created"
	TOPIC_STOCK_RETURN_PRODUCT_UPDATE = "when-stockreturnproduct-updated"
	TOPIC_STOCK_RETURN_PRODUCT_DELETE = "when-stockreturnproduct-deleted"
	// Stock Adjustment topics
	TOPIC_STOCK_ADJUSTMENT_CREATE = "when-stockadjustment-created"
	TOPIC_STOCK_ADJUSTMENT_UPDATE = "when-stockadjustment-updated"
	TOPIC_STOCK_ADJUSTMENT_DELETE = "when-stockadjustment-deleted"

	// Stock Balance topics
	TOPIC_STOCK_BALANCE_CREATE = "when-stockbalance-created"
	TOPIC_STOCK_BALANCE_UPDATE = "when-stockbalance-updated"
	TOPIC_STOCK_BALANCE_DELETE = "when-stockbalance-deleted"
)

// Consumer group constants
const (
	CONSUMER_GROUP_SALE_INVOICE          = "goapi-saleinvoice-consumer"
	CONSUMER_GROUP_SALE_RETURN           = "goapi-saleinvoicereturn-consumer"
	CONSUMER_GROUP_SALE_ORDER            = "goapi-saleorder-consumer"
	CONSUMER_GROUP_INVENTORY             = "biapi-inventory-consumer"
	CONSUMER_GROUP_INVENTORY_BULK        = "biapi-inventory-bulk-consumer"
	CONSUMER_GROUP_WAREHOUSE             = "biapi-warehouse-consumer"
	CONSUMER_GROUP_PURCHASE              = "goapi-purchase-consumer"
	CONSUMER_GROUP_PURCHASE_ORDER        = "goapi-purchaseorder-consumer"
	CONSUMER_GROUP_PURCHASE_PARTIAL      = "goapi-purchasepartial-consumer"
	CONSUMER_GROUP_PURCHASE_RETURN       = "goapi-purchasereturn-consumer"
	CONSUMER_GROUP_PURCHASE_REQUISITION  = "goapi-purchaserequisition-consumer"
	CONSUMER_GROUP_RFQ                   = "goapi-rfq-consumer"
	CONSUMER_GROUP_STOCK_TRANSFER        = "goapi-stocktransfer-consumer"
	CONSUMER_GROUP_STOCK_RECEIVE_PRODUCT = "goapi-stockreceiveproduct-consumer"
	CONSUMER_GROUP_STOCK_PICKUP_PRODUCT  = "goapi-stockpickupproduct-consumer"
	CONSUMER_GROUP_STOCK_RETURN_PRODUCT  = "goapi-stockreturnproduct-consumer"
	CONSUMER_GROUP_STOCK_ADJUSTMENT      = "goapi-stockadjustment-consumer"
	CONSUMER_GROUP_STOCK_BALANCE         = "goapi-stockbalance-consumer"
)

// Transaction flags
const (
	TRANS_FLAG_SALE_INVOICE              = 44
	TRANS_FLAG_SALE_INVOICE_RETURN       = 48
	TRANS_FLAG_SALE_ORDER                = 36
	TRANS_FLAG_PURCHASE                  = 12
	TRANS_FLAG_PURCHASE_ORDER            = 6
	TRANS_FLAG_PURCHASE_PARTIAL          = 310
	TRANS_FLAG_PURCHASE_RETURN           = 16
	TRANS_FLAG_PURCHASE_REQUISITION      = 21
	TRANS_FLAG_RFQ                       = 22
	TRANS_FLAG_STOCK_TRANSFER            = 72
	TRANS_FLAG_STOCK_RECEIVE_PRODUCT     = 60
	TRANS_FLAG_STOCK_PICKUP_PRODUCT      = 56
	TRANS_FLAG_STOCK_RETURN_PRODUCT      = 58
	TRANS_FLAG_STOCK_ADJUSTMENT_INCREASE = 66 // ปรับสต็อก (เพิ่ม)
	TRANS_FLAG_STOCK_ADJUSTMENT_DECREASE = 68 // ปรับสต็อก (ลด)
	TRANS_FLAG_STOCK_BALANCE             = 54 // ยอดยกมา
)

// Other constants
const (
	BATCH_SIZE = 5000
)
