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

	// Purchase Partial topics (เอกสารรับสินค้าใช้ชุดนี้ ไม่มีหัวข้อ purchasereceive แยกในระบบ)
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

	// Purchase Requisition topics — ต้องตรงกับ internal/transaction/purchaserequisition/config
	TOPIC_PURCHASE_REQUISITION_CREATE = "when-purchaserequisition-created"
	TOPIC_PURCHASE_REQUISITION_UPDATE = "when-purchaserequisition-updated"
	TOPIC_PURCHASE_REQUISITION_DELETE = "when-purchaserequisition-deleted"

	// RFQ topics — ต้องตรงกับ internal/transaction/rfq/config
	TOPIC_RFQ_CREATE = "when-rfq-created"
	TOPIC_RFQ_UPDATE = "when-rfq-updated"
	TOPIC_RFQ_DELETE = "when-rfq-deleted"

	// Debtor Payment (Paid) topics
	TOPIC_DEBTOR_PAYMENT_CREATE = "when-debtor-payment-created"
	TOPIC_DEBTOR_PAYMENT_UPDATE = "when-debtor-payment-updated"
	TOPIC_DEBTOR_PAYMENT_DELETE = "when-debtor-payment-deleted"

	// Creditor Payment (Pay) topics
	TOPIC_CREDITOR_PAYMENT_CREATE = "when-creditor-payment-created"
	TOPIC_CREDITOR_PAYMENT_UPDATE = "when-creditor-payment-updated"
	TOPIC_CREDITOR_PAYMENT_DELETE = "when-creditor-payment-deleted"

	// Receivable Other topics
	TOPIC_RECEIVABLE_OTHER_CREATE = "when-debtor-receivableother-created"
	TOPIC_RECEIVABLE_OTHER_UPDATE = "when-debtor-receivableother-updated"
	TOPIC_RECEIVABLE_OTHER_DELETE = "when-debtor-receivableother-deleted"
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
	CONSUMER_GROUP_DEBTOR_PAYMENT        = "goapi-debtorpayment-consumer"
	CONSUMER_GROUP_CREDITOR_PAYMENT      = "goapi-creditorpayment-consumer"
	CONSUMER_GROUP_RECEIVABLE_OTHER      = "goapi-receivableother-consumer"
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
	TRANS_FLAG_STOCK_ADJUSTMENT_INCREASE = 66  // ปรับสต็อก (เพิ่ม)
	TRANS_FLAG_STOCK_ADJUSTMENT_DECREASE = 68  // ปรับสต็อก (ลด)
	TRANS_FLAG_STOCK_BALANCE             = 54  // ยอดยกมา
	TRANS_FLAG_DEBTOR_PAYMENT            = 50  // รับชำระหนี้ (ลูกหนี้ 239->50)
	TRANS_FLAG_CREDITOR_PAYMENT          = 19  // จ่ายชำระหนี้ (เจ้าหนี้)
	TRANS_FLAG_RECEIVABLE_OTHER          = 99  // ตั้งหนี้อื่น (ลูกหนี้)
)

// Other constants
const (
	BATCH_SIZE = 5000
)
