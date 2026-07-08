package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"

	"smlcloudplatform/internal/goapi/mykafkaconsumer"

	"github.com/labstack/echo/v4"
)

// getVersionedGroupID appends version suffix to consumer group ID
// This allows creating new consumer groups for replaying messages
// Example: "goapi-saleinvoice-consumer" + "v1" => "goapi-saleinvoice-consumer-v1"
func getVersionedGroupID(baseGroupID string) string {
	svcConfig := config.NewServiceConfig()
	version := svcConfig.KafkaConsumerGroupVersion()
	return fmt.Sprintf("%s-%s", baseGroupID, version)
}

// StartConsumers - starts all Kafka consumers using the new modular structure
func StartConsumers() {
	kafkaServer := os.Getenv("KAFKA_SERVER_URL")
	if kafkaServer == "" {
		logger.Warn("ไม่ได้กำหนดค่า KAFKA_SERVER_URL environment variable")
		return
	}
	consumers := mykafkaconsumer.NewKafkaConsumer()

	// Start Sale Invoice consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("เกิด Panic ใน Sale Invoice consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleinvoice-created", getVersionedGroupID("goapi-saleinvoice-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผลสร้าง sale invoice: %s", msg)
			// Call actual Sale Invoice consumer function
			return CallSaleInvoiceConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("เกิด Panic ใน Sale Invoice consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleinvoice-updated", getVersionedGroupID("goapi-saleinvoice-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผลอัปเดต sale invoice: %s", msg)
			// Call actual Sale Invoice consumer function
			return CallSaleInvoiceConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("เกิด Panic ใน Sale Invoice consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleinvoice-deleted", getVersionedGroupID("goapi-saleinvoice-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผลลบ sale invoice: %s", msg)
			// Call actual Sale Invoice delete consumer function
			return CallSaleInvoiceDeleteConsumer(msg)
		})
	}()

	// Start Sale Invoice Return consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Sale Invoice Return consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleinvoicereturn-created", getVersionedGroupID("goapi-saleinvoicereturn-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Sale Invoice Return consumer function
			return CallSaleReturnConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Sale Invoice Return consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleinvoicereturn-updated", getVersionedGroupID("goapi-saleinvoicereturn-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Sale Invoice Return consumer function
			return CallSaleReturnConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Sale Invoice Return consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleinvoicereturn-deleted", getVersionedGroupID("goapi-saleinvoicereturn-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Sale Invoice Return delete consumer function
			return CallSaleReturnDeleteConsumer(msg)
		})
	}()

	// Start Inventory consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Inventory consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-barcode-created", getVersionedGroupID("biapi-inventory-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Inventory consumer function
			return CallInventoryConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Inventory consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-barcode-updated", getVersionedGroupID("biapi-inventory-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Inventory consumer function
			return CallInventoryConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Inventory consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-barcode-deleted", getVersionedGroupID("biapi-inventory-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Inventory delete consumer function
			return CallInventoryDeleteConsumer(msg)
		})
	}()

	// Start Product consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-created", getVersionedGroupID("biapi-product-consumer"), 0, func(msg string) error {
			return CallProductConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-updated", getVersionedGroupID("biapi-product-consumer"), 0, func(msg string) error {
			return CallProductConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-deleted", getVersionedGroupID("biapi-product-consumer"), 0, func(msg string) error {
			return CallProductDeleteConsumer(msg)
		})
	}()

	// Start Inventory Bulk consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Inventory Bulk consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-barcode-bulk-created", getVersionedGroupID("biapi-inventory-bulk-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Inventory bulk consumer function
			return CallInventoryBulkConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Inventory Bulk consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-barcode-bulk-updated", getVersionedGroupID("biapi-inventory-bulk-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Inventory bulk consumer function
			return CallInventoryBulkConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Inventory Bulk Delete consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-product-barcode-bulk-deleted", getVersionedGroupID("biapi-inventory-bulk-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk delete")
			// Call actual Inventory bulk delete consumer function
			return CallInventoryBulkDeleteConsumer(msg)
		})
	}()

	// Start Warehouse consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Warehouse consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-warehouse-created", getVersionedGroupID("biapi-warehouse-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Warehouse consumer function
			return CallWarehouseConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Warehouse consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-warehouse-updated", getVersionedGroupID("biapi-warehouse-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Warehouse consumer function
			return CallWarehouseConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Warehouse consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-warehouse-deleted", getVersionedGroupID("biapi-warehouse-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Warehouse delete consumer function
			return CallWarehouseDeleteConsumer(msg)
		})
	}()

	// Start Sale Order consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Sale Order consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleorder-created", getVersionedGroupID("goapi-saleorder-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Sale Order consumer function
			return CallSaleOrderConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Sale Order consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleorder-updated", getVersionedGroupID("goapi-saleorder-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Sale Order consumer function
			return CallSaleOrderConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Sale Order consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-saleorder-deleted", getVersionedGroupID("goapi-saleorder-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Sale Order delete consumer function
			return CallSaleOrderDeleteConsumer(msg)
		})
	}()

	// Start Purchase consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchase-created", getVersionedGroupID("goapi-purchase-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Purchase consumer function
			return CallPurchaseConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchase-updated", getVersionedGroupID("goapi-purchase-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Purchase consumer function
			return CallPurchaseConsumer(msg)
		})
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchase-deleted", getVersionedGroupID("goapi-purchase-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Purchase delete consumer function
			return CallPurchaseDeleteConsumer(msg)
		})
	}()

	// Start Purchase Order consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Order consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaseorder-created", getVersionedGroupID("goapi-purchaseorder-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Purchase Order consumer function
			return CallPurchaseOrderConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Order consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaseorder-updated", getVersionedGroupID("goapi-purchaseorder-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Purchase Order consumer function
			return CallPurchaseOrderConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Order consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaseorder-deleted", getVersionedGroupID("goapi-purchaseorder-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Purchase Order delete consumer function
			return CallPurchaseOrderDeleteConsumer(msg)
		})
	}()

	// Start Purchase Requisition consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Requisition consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaserequisition-created", getVersionedGroupID("goapi-purchaserequisition-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create purchase requisition")
			return CallPurchaseRequisitionConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Requisition consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaserequisition-updated", getVersionedGroupID("goapi-purchaserequisition-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update purchase requisition")
			return CallPurchaseRequisitionConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Requisition consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaserequisition-deleted", getVersionedGroupID("goapi-purchaserequisition-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete purchase requisition")
			return CallPurchaseRequisitionDeleteConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Requisition Bulk consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaserequisition-bulk-created", getVersionedGroupID("goapi-purchaserequisition-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk create purchase requisition")
			return CallPurchaseRequisitionConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Requisition Bulk consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaserequisition-bulk-updated", getVersionedGroupID("goapi-purchaserequisition-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk update purchase requisition")
			return CallPurchaseRequisitionConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Requisition Bulk Delete consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchaserequisition-bulk-deleted", getVersionedGroupID("goapi-purchaserequisition-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk delete purchase requisition")
			return CallPurchaseRequisitionDeleteConsumer(msg)
		})
	}()

	// Start RFQ consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in RFQ consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-rfq-created", getVersionedGroupID("goapi-rfq-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create RFQ")
			return CallRFQConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in RFQ consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-rfq-updated", getVersionedGroupID("goapi-rfq-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update RFQ")
			return CallRFQConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in RFQ consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-rfq-deleted", getVersionedGroupID("goapi-rfq-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete RFQ")
			return CallRFQDeleteConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in RFQ Bulk consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-rfq-bulk-created", getVersionedGroupID("goapi-rfq-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk create RFQ")
			return CallRFQConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in RFQ Bulk consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-rfq-bulk-updated", getVersionedGroupID("goapi-rfq-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk update RFQ")
			return CallRFQConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in RFQ Bulk Delete consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-rfq-bulk-deleted", getVersionedGroupID("goapi-rfq-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk delete RFQ")
			return CallRFQDeleteConsumer(msg)
		})
	}()

	// Start Purchase Partial consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Partial consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchasepartial-created", getVersionedGroupID("goapi-purchasepartial-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Purchase Partial consumer function
			return CallPurchasePartialConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Partial consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchasepartial-updated", getVersionedGroupID("goapi-purchasepartial-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Purchase Partial consumer function
			return CallPurchasePartialConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Partial consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchasepartial-deleted", getVersionedGroupID("goapi-purchasepartial-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Purchase Partial delete consumer function
			return CallPurchasePartialDeleteConsumer(msg)
		})
	}()

	// Start Purchase Return consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Return consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchasereturn-created", getVersionedGroupID("goapi-purchasereturn-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Purchase Return consumer function
			return CallPurchaseReturnConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Return consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchasereturn-updated", getVersionedGroupID("goapi-purchasereturn-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Purchase Return consumer function
			return CallPurchaseReturnConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Purchase Return consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-purchasereturn-deleted", getVersionedGroupID("goapi-purchasereturn-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Purchase Return delete consumer function
			return CallPurchaseReturnDeleteConsumer(msg)
		})
	}()

	// Start Stock Transfer consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Transfer consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stocktransfer-created", getVersionedGroupID("goapi-stocktransfer-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Stock Transfer consumer function
			return CallStockTransferConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Transfer consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stocktransfer-updated", getVersionedGroupID("goapi-stocktransfer-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Stock Transfer consumer function
			return CallStockTransferConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Transfer consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stocktransfer-deleted", getVersionedGroupID("goapi-stocktransfer-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Stock Transfer delete consumer function
			return CallStockTransferDeleteConsumer(msg)
		})
	}()

	// Start Stock Receive Product consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Receive Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockreceiveproduct-created", getVersionedGroupID("goapi-stockreceiveproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Stock Receive Product consumer function
			return CallStockReceiveProductConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Receive Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockreceiveproduct-updated", getVersionedGroupID("goapi-stockreceiveproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Stock Receive Product consumer function
			return CallStockReceiveProductConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Receive Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockreceiveproduct-deleted", getVersionedGroupID("goapi-stockreceiveproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Stock Receive Product delete consumer function
			return CallStockReceiveProductDeleteConsumer(msg)
		})
	}()

	// Start Stock Pickup Product consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Pickup Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockpickupproduct-created", getVersionedGroupID("goapi-stockpickupproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Stock Pickup Product consumer function
			return CallStockPickupProductCreateOrUpdateConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Pickup Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockpickupproduct-updated", getVersionedGroupID("goapi-stockpickupproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Stock Pickup Product consumer function
			return CallStockPickupProductCreateOrUpdateConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Pickup Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockpickupproduct-deleted", getVersionedGroupID("goapi-stockpickupproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Stock Pickup Product delete consumer function
			return CallStockPickupProductDeleteConsumer(msg)
		})
	}()

	// Start Stock Return Product consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Return Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockreturnproduct-created", getVersionedGroupID("goapi-stockreturnproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Stock Return Product consumer function
			return CallStockReturnProductCreateOrUpdateConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Return Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockreturnproduct-updated", getVersionedGroupID("goapi-stockreturnproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Stock Return Product consumer function
			return CallStockReturnProductCreateOrUpdateConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Return Product consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockreturnproduct-deleted", getVersionedGroupID("goapi-stockreturnproduct-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Stock Return Product delete consumer function
			return CallStockReturnProductDeleteConsumer(msg)
		})
	}()

	// Start Stock Adjustment consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Adjustment consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockadjustment-created", getVersionedGroupID("goapi-stockadjustment-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Stock Adjustment consumer function
			return CallStockAdjustmentCreateOrUpdateConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Adjustment consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockadjustment-updated", getVersionedGroupID("goapi-stockadjustment-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Stock Adjustment consumer function
			return CallStockAdjustmentCreateOrUpdateConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Adjustment consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockadjustment-deleted", getVersionedGroupID("goapi-stockadjustment-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Stock Adjustment delete consumer function
			return CallStockAdjustmentDeleteConsumer(msg)
		})
	}()

	// Start Stock Balance consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Balance consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockbalance-created", getVersionedGroupID("goapi-stockbalance-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create $0", msg)
			// Call actual Stock Balance consumer function
			return CallStockBalanceCreateOrUpdateConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Balance consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockbalance-updated", getVersionedGroupID("goapi-stockbalance-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update $0", msg)
			// Call actual Stock Balance consumer function
			return CallStockBalanceCreateOrUpdateConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Stock Balance consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-stockbalance-deleted", getVersionedGroupID("goapi-stockbalance-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete $0", msg)
			// Call actual Stock Balance delete consumer function
			return CallStockBalanceDeleteConsumer(msg)
		})
	}()

	// Start Creditor consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Creditor consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-creditor-created", getVersionedGroupID("goapi-creditor-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create creditor: %s", msg)
			return CallCreditorConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Creditor consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-creditor-updated", getVersionedGroupID("goapi-creditor-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update creditor: %s", msg)
			return CallCreditorConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Creditor consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-creditor-deleted", getVersionedGroupID("goapi-creditor-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete creditor: %s", msg)
			return CallCreditorDeleteConsumer(msg)
		})
	}()

	// Start Creditor Bulk consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Creditor Bulk consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-creditor-bulk-created", getVersionedGroupID("goapi-creditor-bulk-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk create creditor")
			return CallCreditorBulkConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Creditor Bulk Delete consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-creditor-bulk-deleted", getVersionedGroupID("goapi-creditor-bulk-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk delete creditor")
			return CallCreditorBulkDeleteConsumer(msg)
		})
	}()

	// Start Customer consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Customer consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-customer-created", getVersionedGroupID("goapi-customer-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create customer: %s", msg)
			return CallCustomerConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Customer consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-customer-updated", getVersionedGroupID("goapi-customer-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update customer: %s", msg)
			return CallCustomerConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Customer consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-customer-deleted", getVersionedGroupID("goapi-customer-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete customer: %s", msg)
			return CallCustomerDeleteConsumer(msg)
		})
	}()

	// Start Employee consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Employee consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-employee-created", getVersionedGroupID("goapi-employee-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create employee: %s", msg)
			return CallEmployeeConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Employee consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-employee-updated", getVersionedGroupID("goapi-employee-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update employee: %s", msg)
			return CallEmployeeConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Employee consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-employee-deleted", getVersionedGroupID("goapi-employee-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete employee: %s", msg)
			return CallEmployeeDeleteConsumer(msg)
		})
	}()

	// Start Debtor consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Debtor consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-debtor-created", getVersionedGroupID("goapi-debtor-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล create debtor: %s", msg)
			return CallDebtorConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Debtor consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-debtor-updated", getVersionedGroupID("goapi-debtor-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล update debtor: %s", msg)
			return CallDebtorConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Debtor consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-debtor-deleted", getVersionedGroupID("goapi-debtor-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล delete debtor: %s", msg)
			return CallDebtorDeleteConsumer(msg)
		})
	}()

	// Start Debtor Bulk consumers
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Debtor Bulk consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-debtor-bulk-created", getVersionedGroupID("goapi-debtor-bulk-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk create debtor")
			return CallDebtorBulkConsumer(msg)
		})
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("Panic in Debtor Bulk Delete consumer: %v", r)
			}
		}()
		consumers.ConsumeMessage(kafkaServer, "when-debtor-bulk-deleted", getVersionedGroupID("goapi-debtor-bulk-consumer"), 0, func(msg string) error {
			logger.Info("กำลังประมวลผล bulk delete debtor")
			return CallDebtorBulkDeleteConsumer(msg)
		})
	}()

	logger.Info("เริ่มต้น Kafka consumers เรียบร้อย")
}

// API endpoints for testing Kafka consumers

// TestSaleInvoiceConsumer - API endpoint to test the sale invoice consumer
func TestSaleInvoiceConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function directly
	logger.Info("Testing sale invoice consumer with message: %s", string(jsonBytes))
	// TODO: Call actual consumer function when available
	// err = kafka.OnConsumeMessageSaleInvoiceCreateOrUpdate(string(jsonBytes))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Sale invoice consumer test completed successfully",
		"note":    "Consumer function placeholder - to be implemented",
	})
}

// TestSaleReturnConsumer - API endpoint to test the sale return consumer
func TestSaleReturnConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function directly
	logger.Info("Testing sale return consumer with message: %s", string(jsonBytes))
	// TODO: Call actual consumer function when available
	// err = kafka.OnConsumeMessageSaleInvoiceReturnCreateOrUpdate(string(jsonBytes))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Sale return consumer test completed successfully",
		"note":    "Consumer function placeholder - to be implemented",
	})
}

// TestInventoryConsumer - API endpoint to test the inventory consumer
func TestInventoryConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function directly
	logger.Info("Testing inventory consumer with message: %s", string(jsonBytes))
	// TODO: Call actual consumer function when available
	// err = kafka.OnConsumeMessageInventoryCreateOrUpdate(string(jsonBytes))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Inventory consumer test completed successfully",
		"note":    "Consumer function placeholder - to be implemented",
	})
}

// TestWarehouseConsumer - API endpoint to test the warehouse consumer
func TestWarehouseConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function directly
	logger.Info("Testing warehouse consumer with message: %s", string(jsonBytes))
	// TODO: Call actual consumer function when available
	// err = kafka.OnConsumeMessageWarehouseCreateOrUpdate(string(jsonBytes))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Warehouse consumer test completed successfully",
		"note":    "Consumer function placeholder - to be implemented",
	})
}

// TestPurchaseReturnConsumer - API endpoint to test the purchase return consumer
func TestPurchaseReturnConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function directly
	logger.Info("Testing purchase return consumer with message: %s", string(jsonBytes))
	err = CallPurchaseReturnConsumer(string(jsonBytes))
	if err != nil {
		logger.Info("Purchase return consumer test failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Purchase return consumer test failed",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Purchase return consumer test completed successfully",
	})
}

// GetConsumerStatus - API endpoint to get the status of all consumers
func GetConsumerStatus(c echo.Context) error {
	kafkaServer := os.Getenv("KAFKA_SERVER_URL")
	status := map[string]interface{}{
		"kafka_server": kafkaServer,
		"is_valid":     kafkaServer != "",
		"batch_size":   5000,
		"consumer_groups": map[string]string{
			"sale_invoice":         "goapi-saleinvoice-consumer",
			"sale_return":          "goapi-saleinvoicereturn-consumer",
			"sale_order":           "goapi-saleorder-consumer",
			"inventory":            "biapi-inventory-consumer",
			"inventory_bulk":       "biapi-inventory-bulk-consumer",
			"warehouse":            "biapi-warehouse-consumer",
			"purchase":             "goapi-purchase-consumer",
			"purchase_order":       "goapi-purchaseorder-consumer",
			"purchase_partial":     "goapi-purchasepartial-consumer",
			"purchase_return":      "goapi-purchasereturn-consumer",
			"purchase_requisition": "goapi-purchaserequisition-consumer",
			"rfq":                  "goapi-rfq-consumer",
			"creditor":             "goapi-creditor-consumer",
			"customer":             "goapi-customer-consumer",
			"employee":             "goapi-employee-consumer",
			"debtor":               "goapi-debtor-consumer",
		},
		"topics": map[string]interface{}{
			"sale_invoice": []string{
				"when-saleinvoice-created",
				"when-saleinvoice-updated",
				"when-saleinvoice-deleted",
			}, "sale_return": []string{
				"when-saleinvoicereturn-created",
				"when-saleinvoicereturn-updated",
				"when-saleinvoicereturn-deleted",
			},
			"sale_order": []string{
				"when-saleorder-created",
				"when-saleorder-updated",
				"when-saleorder-deleted",
			},
			"inventory": []string{
				"when-product-barcode-created",
				"when-product-barcode-updated",
				"when-product-barcode-deleted",
				"when-product-barcode-bulk-created",
				"when-product-barcode-bulk-updated",
				"when-product-barcode-bulk-deleted",
			},
			"warehouse": []string{
				"when-warehouse-created",
				"when-warehouse-updated",
				"when-warehouse-deleted"}, "purchase": []string{
				"when-purchaseorder-created",
				"when-purchaseorder-updated",
				"when-purchaseorder-deleted",
				"when-purchase-created",
				"when-purchase-updated",
				"when-purchase-deleted",
				"when-purchasepartial-created",
				"when-purchasepartial-updated",
				"when-purchasepartial-deleted",
				"when-purchasereturn-created",
				"when-purchasereturn-updated",
				"when-purchasereturn-deleted",
				"when-purchasereceive-created",
				"when-purchasereceive-updated",
				"when-purchasereceive-deleted",
			},
			"stock_transfer": []string{
				"when-stocktransfer-created",
				"when-stocktransfer-updated",
				"when-stocktransfer-deleted",
			},
			"master_data": []string{
				"when-creditor-created",
				"when-creditor-updated",
				"when-creditor-deleted",
				"when-creditor-bulk-created",
				"when-creditor-bulk-deleted",
				"when-customer-created",
				"when-customer-updated",
				"when-customer-deleted",
				"when-employee-created",
				"when-employee-updated",
				"when-employee-deleted",
				"when-debtor-created",
				"when-debtor-updated",
				"when-debtor-deleted",
				"when-debtor-bulk-created",
				"when-debtor-bulk-deleted",
			},
		},
	}

	return c.JSON(http.StatusOK, status)
}
