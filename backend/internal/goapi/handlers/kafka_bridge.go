package handlers

import (
	"smlcloudplatform/internal/goapi/handlers/kafka"
)

/*
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	KAFKA BRIDGE FUNCTIONS - ฟังก์ชันเชื่อมต่อ Kafka Consumer
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📌 ชื่อระบบ: Kafka Message Consumer Bridge Layer
📌 วัตถุประสงค์: เป็นตัวกลางเชื่อมต่อระหว่าง handlers/kafka.go กับ handlers/kafka/*.go
📌 รูปแบบการทำงาน: Bridge Pattern

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📖 การทำงานของระบบ:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Kafka Message Flow (ขั้นตอนการรับ Message):
   ┌─────────────────┐
   │  Kafka Broker   │ ← Message จาก MongoDB Change Stream
   └────────┬────────┘
            │
   ┌────────▼────────────────────────────────────────────────────────────────┐
   │  mykafkaconsumer/kafka_consumer.go                                      │
   │  - รับ message จาก Kafka topic                                          │
   │  - แยก topic ตามประเภทเอกสาร (Sale, Purchase, Stock, Product, etc.)    │
   └────────┬────────────────────────────────────────────────────────────────┘
            │
   ┌────────▼────────────────────────────────────────────────────────────────┐
   │  handlers/kafka.go (StartConsumers)                                     │
   │  - จัดการ 47 topics                                                     │
   │  - แยกตาม operation (created/updated/deleted)                           │
   │  - เรียก Bridge Functions ในไฟล์นี้                                     │
   └────────┬────────────────────────────────────────────────────────────────┘
            │
   ┌────────▼────────────────────────────────────────────────────────────────┐
   │  handlers/kafka_bridge.go (ไฟล์นี้)                                     │
   │  - รับ message จาก kafka.go                                             │
   │  - เรียก function ที่เหมาะสมใน handlers/kafka/*.go                      │
   │  - ทำหน้าที่เป็น Bridge Layer (ชั้นกลาง)                                │
   └────────┬────────────────────────────────────────────────────────────────┘
            │
   ┌────────▼────────────────────────────────────────────────────────────────┐
   │  handlers/kafka/*.go (ตัวประมวลผลจริง)                                  │
   │  - purchase_order.go, purchase.go, sale_invoice.go, etc.                │
   │  - ประมวลผล JSON message                                                │
   │  - บันทึกข้อมูลลง PostgreSQL และ ClickHouse                             │
   │  - อัพเดท stock queue และ process queue                                │
   └─────────────────────────────────────────────────────────────────────────┘

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 รายการระบบที่รองรับ (47 Topics / 29 Bridge Functions):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

┌─────────────────────────────────────────────────────────────────────────┐
│ 1. ระบบขาย (Sales) - 6 Topics                                           │
├─────────────────────────────────────────────────────────────────────────┤
│ • Sale Invoice (ใบกำกับภาษี/ใบเสร็จรับเงิน)                             │
│   - when-saleinvoice-created/updated → CallSaleInvoiceConsumer          │
│   - when-saleinvoice-deleted → CallSaleInvoiceDeleteConsumer            │
│                                                                           │
│ • Sale Invoice Return (ใบลดหนี้/ใบรับคืนสินค้า)                         │
│   - when-saleinvoicereturn-created/updated → CallSaleReturnConsumer     │
│   - when-saleinvoicereturn-deleted → CallSaleReturnDeleteConsumer       │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ 2. ระบบใบสั่งขาย (Sale Order) - 3 Topics                               │
├─────────────────────────────────────────────────────────────────────────┤
│ • Sale Order (ใบสั่งขาย/ใบเสนอราคา)                                     │
│   - when-saleorder-created/updated → CallSaleOrderConsumer              │
│   - when-saleorder-deleted → CallSaleOrderDeleteConsumer                │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ 3. ระบบจัดซื้อ (Purchase) - 12 Topics                                  │
├─────────────────────────────────────────────────────────────────────────┤
│ • Purchase Order (ใบสั่งซื้อ)                                           │
│   - when-purchaseorder-created/updated → CallPurchaseOrderConsumer      │
│   - when-purchaseorder-deleted → CallPurchaseOrderDeleteConsumer        │
│                                                                           │
│ • Purchase (ใบรับสินค้า)                                                │
│   - when-purchase-created/updated → CallPurchaseConsumer                │
│   - when-purchase-deleted → CallPurchaseDeleteConsumer                  │
│                                                                           │
│ • Purchase Partial (ใบรับสินค้าบางส่วน)                                │
│   - when-purchasepartial-created/updated → CallPurchasePartialConsumer  │
│   - when-purchasepartial-deleted → CallPurchasePartialDeleteConsumer    │
│                                                                           │
│ • Purchase Return (ใบคืนสินค้าซื้อ)                                     │
│   - when-purchasereturn-created/updated → CallPurchaseReturnConsumer    │
│   - when-purchasereturn-deleted → CallPurchaseReturnDeleteConsumer      │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ 4. ระบบสินค้าและคลังสินค้า (Product & Warehouse) - 8 Topics           │
├─────────────────────────────────────────────────────────────────────────┤
│ • Product Barcode (บาร์โค้ดสินค้า)                                      │
│   - when-product-barcode-created/updated → CallInventoryConsumer        │
│   - when-product-barcode-deleted → CallInventoryDeleteConsumer          │
│   - when-product-barcode-bulk-created/updated → CallInventoryBulkConsumer│
│   - when-product-barcode-bulk-deleted → CallInventoryBulkDeleteConsumer │
│                                                                           │
│ • Warehouse (คลังสินค้า/สาขา)                                           │
│   - when-warehouse-created/updated → CallWarehouseConsumer              │
│   - when-warehouse-deleted → CallWarehouseDeleteConsumer                │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ 5. ระบบสต็อก (Stock Management) - 18 Topics                            │
├─────────────────────────────────────────────────────────────────────────┤
│ • Stock Transfer (ใบโอนสินค้า)                                          │
│   - when-stocktransfer-created/updated → CallStockTransferConsumer      │
│   - when-stocktransfer-deleted → CallStockTransferDeleteConsumer        │
│                                                                           │
│ • Stock Receive Product (ใบรับสินค้าเข้าคลัง)                           │
│   - when-stockreceiveproduct-created/updated → CallStockReceiveProductConsumer│
│   - when-stockreceiveproduct-deleted → CallStockReceiveProductDeleteConsumer  │
│                                                                           │
│ • Stock Pickup Product (ใบเบิกสินค้า)                                   │
│   - when-stockpickupproduct-created/updated → CallStockPickupProductCreateOrUpdateConsumer│
│   - when-stockpickupproduct-deleted → CallStockPickupProductDeleteConsumer    │
│                                                                           │
│ • Stock Return Product (ใบคืนสินค้าเข้าคลัง)                            │
│   - when-stockreturnproduct-created/updated → CallStockReturnProductCreateOrUpdateConsumer│
│   - when-stockreturnproduct-deleted → CallStockReturnProductDeleteConsumer    │
│                                                                           │
│ • Stock Adjustment (ใบปรับสต็อก)                                        │
│   - when-stockadjustment-created/updated → CallStockAdjustmentCreateOrUpdateConsumer│
│   - when-stockadjustment-deleted → CallStockAdjustmentDeleteConsumer    │
│                                                                           │
│ • Stock Balance (ใบตรวจนับสต็อก)                                        │
│   - when-stockbalance-created/updated → CallStockBalanceCreateOrUpdateConsumer│
│   - when-stockbalance-deleted → CallStockBalanceDeleteConsumer          │
└─────────────────────────────────────────────────────────────────────────┘

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔧 หลักการทำงาน:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Separation of Concerns (แยกหน้าที่ชัดเจน):
   - kafka.go        → จัดการ Kafka topics และ routing
   - kafka_bridge.go → เชื่อมต่อและส่งต่อ (ไฟล์นี้)
   - kafka/*.go      → ประมวลผลข้อมูลจริง

2. Single Responsibility (หน้าที่เดียว):
   - แต่ละ bridge function มีหน้าที่เพียงเรียก consumer function เดียว
   - ไม่มี business logic ในชั้น bridge

3. Maintainability (บำรุงรักษาง่าย):
   - เพิ่ม/ลด consumer ใหม่ได้ง่าย
   - แก้ไข implementation ได้โดยไม่กระทบ kafka.go

4. Error Handling (จัดการ error):
   - ส่ง error กลับไปยัง kafka.go เพื่อ log
   - ไม่ panic หาก consumer function มีปัญหา

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 ตัวอย่างการใช้งาน:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ใน kafka.go:
	consumers.ConsumeMessage(kafkaServer, "when-saleinvoice-created", "goapi-saleinvoice-consumer", 0,
		func(msg string) error {
			return CallSaleInvoiceConsumer(msg)  ← เรียก bridge function
		})

Bridge function (ไฟล์นี้):
	func CallSaleInvoiceConsumer(msg string) error {
		return kafka.OnConsumeMessageSaleInvoiceCreateOrUpdate(msg)  ← ส่งต่อไปยัง handler
	}

Handler (handlers/kafka/sale_invoice.go):
	func OnConsumeMessageSaleInvoiceCreateOrUpdate(msg string) error {
		// ประมวลผลข้อมูล Sale Invoice
		// 1. Decode JSON
		// 2. Validate data
		// 3. Save to PostgreSQL & ClickHouse
		// 4. Update queues
		return nil
	}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
*/

// ============================================================================
// ระบบขาย (Sales Documents)
// ============================================================================

// CallSaleInvoiceConsumer - ประมวลผลใบกำกับภาษี/ใบเสร็จรับเงิน (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-saleinvoice-created, when-saleinvoice-updated
// ส่งต่อไปยัง handlers/kafka/sale_invoice.go
func CallSaleInvoiceConsumer(msg string) error {
	return kafka.OnConsumeMessageSaleInvoiceCreateOrUpdate(msg)
}

// CallSaleInvoiceDeleteConsumer - ประมวลผลการลบใบกำกับภาษี/ใบเสร็จรับเงิน
// รับ JSON message จาก Kafka topic: when-saleinvoice-deleted
// ส่งต่อไปยัง handlers/kafka/sale_invoice.go
func CallSaleInvoiceDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageSaleInvoiceDelete(msg)
}

// CallSaleReturnConsumer - ประมวลผลใบลดหนี้/ใบรับคืนสินค้า (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-saleinvoicereturn-created, when-saleinvoicereturn-updated
// ส่งต่อไปยัง handlers/kafka/sale_return.go
func CallSaleReturnConsumer(msg string) error {
	return kafka.OnConsumeMessageSaleInvoiceReturnCreateOrUpdate(msg)
}

// CallSaleReturnDeleteConsumer - ประมวลผลการลบใบลดหนี้/ใบรับคืนสินค้า
// รับ JSON message จาก Kafka topic: when-saleinvoicereturn-deleted
// ส่งต่อไปยัง handlers/kafka/sale_return.go
func CallSaleReturnDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageSaleInvoiceReturnDelete(msg)
}

// ============================================================================
// ระบบใบสั่งขาย (Sale Order)
// ============================================================================

// CallSaleOrderConsumer - ประมวลผลใบสั่งขาย/ใบเสนอราคา (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-saleorder-created, when-saleorder-updated
// ส่งต่อไปยัง handlers/kafka/sale_order.go
func CallSaleOrderConsumer(msg string) error {
	return kafka.OnConsumeMessageSaleOrderCreateOrUpdate(msg)
}

// CallSaleOrderDeleteConsumer - ประมวลผลการลบใบสั่งขาย/ใบเสนอราคา
// รับ JSON message จาก Kafka topic: when-saleorder-deleted
// ส่งต่อไปยัง handlers/kafka/sale_order.go
func CallSaleOrderDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageSaleOrderDelete(msg)
}

// ============================================================================
// ระบบสินค้าและคลังสินค้า (Product & Warehouse)
// ============================================================================

// CallInventoryConsumer - ประมวลผลบาร์โค้ดสินค้า (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-product-barcode-created, when-product-barcode-updated
// ส่งต่อไปยัง handlers/kafka/inventory.go
func CallInventoryConsumer(msg string) error {
	return kafka.OnConsumeMessageInventoryCreateOrUpdate(msg)
}

// CallInventoryDeleteConsumer - ประมวลผลการลบบาร์โค้ดสินค้า
// รับ JSON message จาก Kafka topic: when-product-barcode-deleted
// ส่งต่อไปยัง handlers/kafka/inventory.go
func CallInventoryDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageInventoryDelete(msg)
}

// CallInventoryBulkConsumer - ประมวลผลบาร์โค้ดสินค้าแบบ bulk (สร้าง/แก้ไขหลายรายการ)
// รับ JSON message จาก Kafka topic: when-product-barcode-bulk-created, when-product-barcode-bulk-updated
// ส่งต่อไปยัง handlers/kafka/inventory.go
func CallInventoryBulkConsumer(msg string) error {
	return kafka.OnConsumeMessageInventoryBulkCreateOrUpdate(msg)
}

// CallInventoryBulkDeleteConsumer - ประมวลผลการลบบาร์โค้ดสินค้าแบบ bulk
// รับ JSON message จาก Kafka topic: when-product-barcode-bulk-deleted
// ส่งต่อไปยัง handlers/kafka/inventory.go
func CallInventoryBulkDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageInventoryBulkDelete(msg)
}

// CallWarehouseConsumer - ประมวลผลคลังสินค้า/สาขา (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-warehouse-created, when-warehouse-updated
// ส่งต่อไปยัง handlers/kafka/warehouse.go
func CallWarehouseConsumer(msg string) error {
	return kafka.OnConsumeMessageWarehouseCreateOrUpdate(msg)
}

// CallWarehouseDeleteConsumer - ประมวลผลการลบคลังสินค้า/สาขา
// รับ JSON message จาก Kafka topic: when-warehouse-deleted
// ส่งต่อไปยัง handlers/kafka/warehouse.go
func CallWarehouseDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageWarehouseDelete(msg)
}

// ============================================================================
// ระบบจัดซื้อ (Purchase Documents)
// ============================================================================

// CallPurchaseOrderConsumer - ประมวลผลใบสั่งซื้อ (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-purchaseorder-created, when-purchaseorder-updated
// ส่งต่อไปยัง handlers/kafka/purchase_order.go
func CallPurchaseOrderConsumer(msg string) error {
	return kafka.OnConsumeMessagePurchaseOrderCreateOrUpdate(msg)
}

// CallPurchaseOrderDeleteConsumer - ประมวลผลการลบใบสั่งซื้อ
// รับ JSON message จาก Kafka topic: when-purchaseorder-deleted
// ส่งต่อไปยัง handlers/kafka/purchase_order.go
func CallPurchaseOrderDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessagePurchaseOrderDelete(msg)
}

// CallPurchaseConsumer - ประมวลผลใบรับสินค้า (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-purchase-created, when-purchase-updated
// ส่งต่อไปยัง handlers/kafka/purchase.go
func CallPurchaseConsumer(msg string) error {
	return kafka.OnConsumeMessagePurchaseCreateOrUpdate(msg)
}

// CallPurchaseDeleteConsumer - ประมวลผลการลบใบรับสินค้า
// รับ JSON message จาก Kafka topic: when-purchase-deleted
// ส่งต่อไปยัง handlers/kafka/purchase.go
func CallPurchaseDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessagePurchaseDelete(msg)
}

// CallPurchasePartialConsumer - ประมวลผลใบรับสินค้าบางส่วน (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-purchasepartial-created, when-purchasepartial-updated
// ส่งต่อไปยัง handlers/kafka/purchase_partial.go
func CallPurchasePartialConsumer(msg string) error {
	return kafka.OnConsumeMessagePurchasePartialCreateOrUpdate(msg)
}

// CallPurchasePartialDeleteConsumer - ประมวลผลการลบใบรับสินค้าบางส่วน
// รับ JSON message จาก Kafka topic: when-purchasepartial-deleted
// ส่งต่อไปยัง handlers/kafka/purchase_partial.go
func CallPurchasePartialDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessagePurchasePartialDelete(msg)
}

// CallPurchaseReturnConsumer - ประมวลผลใบคืนสินค้าซื้อ (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-purchasereturn-created, when-purchasereturn-updated
// ส่งต่อไปยัง handlers/kafka/purchase_return.go
func CallPurchaseReturnConsumer(msg string) error {
	return kafka.OnConsumeMessagePurchaseReturnCreateOrUpdate(msg)
}

// CallPurchaseReturnDeleteConsumer - ประมวลผลการลบใบคืนสินค้าซื้อ
// รับ JSON message จาก Kafka topic: when-purchasereturn-deleted
// ส่งต่อไปยัง handlers/kafka/purchase_return.go
func CallPurchaseReturnDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessagePurchaseReturnDelete(msg)
}

// ============================================================================
// ระบบสต็อก (Stock Management)
// ============================================================================

// CallStockTransferConsumer - ประมวลผลใบโอนสินค้า (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-stocktransfer-created, when-stocktransfer-updated
// ส่งต่อไปยัง handlers/kafka/stock_transfer.go
func CallStockTransferConsumer(msg string) error {
	return kafka.OnConsumeMessageStockTransferCreateOrUpdate(msg)
}

// CallStockTransferDeleteConsumer - ประมวลผลการลบใบโอนสินค้า
// รับ JSON message จาก Kafka topic: when-stocktransfer-deleted
// ส่งต่อไปยัง handlers/kafka/stock_transfer.go
func CallStockTransferDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageStockTransferDelete(msg)
}

// CallStockReceiveProductConsumer - ประมวลผลใบรับสินค้าเข้าคลัง (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-stockreceiveproduct-created, when-stockreceiveproduct-updated
// ส่งต่อไปยัง handlers/kafka/stock_receive_product.go
func CallStockReceiveProductConsumer(msg string) error {
	return kafka.OnConsumeMessageStockReceiveProductCreateOrUpdate(msg)
}

// CallStockReceiveProductDeleteConsumer - ประมวลผลการลบใบรับสินค้าเข้าคลัง
// รับ JSON message จาก Kafka topic: when-stockreceiveproduct-deleted
// ส่งต่อไปยัง handlers/kafka/stock_receive_product.go
func CallStockReceiveProductDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageStockReceiveProductDelete(msg)
}

// CallStockPickupProductCreateOrUpdateConsumer - ประมวลผลใบเบิกสินค้า (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-stockpickupproduct-created, when-stockpickupproduct-updated
// ส่งต่อไปยัง handlers/kafka/stock_pickup_product.go
func CallStockPickupProductCreateOrUpdateConsumer(msg string) error {
	return kafka.OnConsumeMessageStockPickupProductCreateOrUpdate(msg)
}

// CallStockPickupProductDeleteConsumer - ประมวลผลการลบใบเบิกสินค้า
// รับ JSON message จาก Kafka topic: when-stockpickupproduct-deleted
// ส่งต่อไปยัง handlers/kafka/stock_pickup_product.go
func CallStockPickupProductDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageStockPickupProductDelete(msg)
}

// CallStockReturnProductCreateOrUpdateConsumer - ประมวลผลใบคืนสินค้าเข้าคลัง (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-stockreturnproduct-created, when-stockreturnproduct-updated
// ส่งต่อไปยัง handlers/kafka/stock_return_product.go
func CallStockReturnProductCreateOrUpdateConsumer(msg string) error {
	return kafka.OnConsumeMessageStockReturnProductCreateOrUpdate(msg)
}

// CallStockReturnProductDeleteConsumer - ประมวลผลการลบใบคืนสินค้าเข้าคลัง
// รับ JSON message จาก Kafka topic: when-stockreturnproduct-deleted
// ส่งต่อไปยัง handlers/kafka/stock_return_product.go
func CallStockReturnProductDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageStockReturnProductDelete(msg)
}

// CallStockAdjustmentCreateOrUpdateConsumer - ประมวลผลใบปรับสต็อก (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-stockadjustment-created, when-stockadjustment-updated
// ส่งต่อไปยัง handlers/kafka/stock_adjustment.go
func CallStockAdjustmentCreateOrUpdateConsumer(msg string) error {
	return kafka.OnConsumeMessageStockAdjustmentCreateOrUpdate(msg)
}

// CallStockAdjustmentDeleteConsumer - ประมวลผลการลบใบปรับสต็อก
// รับ JSON message จาก Kafka topic: when-stockadjustment-deleted
// ส่งต่อไปยัง handlers/kafka/stock_adjustment.go
func CallStockAdjustmentDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageStockAdjustmentDelete(msg)
}

// CallStockBalanceCreateOrUpdateConsumer - ประมวลผลใบตรวจนับสต็อก (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-stockbalance-created, when-stockbalance-updated
// ส่งต่อไปยัง handlers/kafka/stock_balance.go
func CallStockBalanceCreateOrUpdateConsumer(msg string) error {
	return kafka.OnConsumeMessageStockBalanceCreateOrUpdate(msg)
}

// CallStockBalanceDeleteConsumer - ประมวลผลการลบใบตรวจนับสต็อก
// รับ JSON message จาก Kafka topic: when-stockbalance-deleted
// ส่งต่อไปยัง handlers/kafka/stock_balance.go
func CallStockBalanceDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageStockBalanceDelete(msg)
}

// ============================================================================
// ระบบข้อมูลหลัก (Master Data)
// ============================================================================

// CallCreditorConsumer - ประมวลผลเจ้าหนี้ (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-creditor-created, when-creditor-updated
// ส่งต่อไปยัง handlers/kafka/creditor.go
func CallCreditorConsumer(msg string) error {
	return kafka.OnConsumeMessageCreditorCreateOrUpdate(msg)
}

// CallCreditorDeleteConsumer - ประมวลผลการลบเจ้าหนี้
// รับ JSON message จาก Kafka topic: when-creditor-deleted
// ส่งต่อไปยัง handlers/kafka/creditor.go
func CallCreditorDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageCreditorDelete(msg)
}

// CallCreditorBulkConsumer - ประมวลผลเจ้าหนี้แบบ bulk (สร้าง/แก้ไขหลายรายการ)
// รับ JSON array จาก Kafka topic: when-creditor-bulk-created
func CallCreditorBulkConsumer(msg string) error {
	return kafka.OnConsumeMessageCreditorBulkCreateOrUpdate(msg)
}

// CallCreditorBulkDeleteConsumer - ประมวลผลการลบเจ้าหนี้แบบ bulk
// รับ JSON array จาก Kafka topic: when-creditor-bulk-deleted
func CallCreditorBulkDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageCreditorBulkDelete(msg)
}

// CallCustomerConsumer - ประมวลผลลูกค้า (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-customer-created, when-customer-updated
// ส่งต่อไปยัง handlers/kafka/customer.go
func CallCustomerConsumer(msg string) error {
	return kafka.OnConsumeMessageCustomerCreateOrUpdate(msg)
}

// CallCustomerDeleteConsumer - ประมวลผลการลบลูกค้า
// รับ JSON message จาก Kafka topic: when-customer-deleted
// ส่งต่อไปยัง handlers/kafka/customer.go
func CallCustomerDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageCustomerDelete(msg)
}

// CallEmployeeConsumer - ประมวลผลพนักงาน (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-employee-created, when-employee-updated
// ส่งต่อไปยัง handlers/kafka/employee.go
func CallEmployeeConsumer(msg string) error {
	return kafka.OnConsumeMessageEmployeeCreateOrUpdate(msg)
}

// CallEmployeeDeleteConsumer - ประมวลผลการลบพนักงาน
// รับ JSON message จาก Kafka topic: when-employee-deleted
// ส่งต่อไปยัง handlers/kafka/employee.go
func CallEmployeeDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageEmployeeDelete(msg)
}

// CallDebtorConsumer - ประมวลผลลูกหนี้ (สร้าง/แก้ไข)
// รับ JSON message จาก Kafka topic: when-debtor-created, when-debtor-updated
// ส่งต่อไปยัง handlers/kafka/debtor.go
func CallDebtorConsumer(msg string) error {
	return kafka.OnConsumeMessageDebtorCreateOrUpdate(msg)
}

// CallDebtorDeleteConsumer - ประมวลผลการลบลูกหนี้
// รับ JSON message จาก Kafka topic: when-debtor-deleted
// ส่งต่อไปยัง handlers/kafka/debtor.go
func CallDebtorDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageDebtorDelete(msg)
}

// CallDebtorBulkConsumer - ประมวลผลลูกหนี้แบบ bulk (สร้าง/แก้ไขหลายรายการ)
// รับ JSON array จาก Kafka topic: when-debtor-bulk-created
func CallDebtorBulkConsumer(msg string) error {
	return kafka.OnConsumeMessageDebtorBulkCreateOrUpdate(msg)
}

// CallDebtorBulkDeleteConsumer - ประมวลผลการลบลูกหนี้แบบ bulk
// รับ JSON array จาก Kafka topic: when-debtor-bulk-deleted
func CallDebtorBulkDeleteConsumer(msg string) error {
	return kafka.OnConsumeMessageDebtorBulkDelete(msg)
}

// ============================================================================
// Startup Function
// ============================================================================

// StartKafkaConsumersWithActualImplementation - เริ่มต้น Kafka consumers ทั้งหมด
// เรียกใช้จาก cmd/main.go เพื่อเริ่มรับ messages จาก Kafka
// ฟังก์ชันนี้จะเรียก kafka.StartConsumers() ซึ่งจะเริ่ม goroutines สำหรับแต่ละ topic
func StartKafkaConsumersWithActualImplementation() {
	kafka.StartConsumers()
}
