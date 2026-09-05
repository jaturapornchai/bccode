package kafka

import (
	"encoding/json"
	"fmt"
	"strings"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/utils"
)

// OnConsumeMessageInventoryCreateOrUpdate - handles inventory create/update messages
func OnConsumeMessageInventoryCreateOrUpdate(msg string) error {
	return consumeBarcodeSignals(msg, false)
}

// OnConsumeMessageInventoryDelete - handles inventory delete messages
func OnConsumeMessageInventoryDelete(msg string) error { return consumeBarcodeSignals(msg, false) }

// OnConsumeMessageInventoryBulkCreateOrUpdate - handles bulk inventory create/update messages
func OnConsumeMessageInventoryBulkCreateOrUpdate(msg string) error {
	return consumeBarcodeSignals(msg, true)
}

// OnConsumeMessageInventoryBulkDelete - handles bulk inventory delete messages
func OnConsumeMessageInventoryBulkDelete(msg string) error { return consumeBarcodeSignals(msg, true) }

// OnProductBarcodeCreateUpdateMessageConsume - wrapper function for product barcode processing
func OnProductBarcodeCreateUpdateMessageConsume(msg string) error {
	return SafeConsumerWrapper("PRODUCT_BARCODE_WAREHOUSE", func(msg string) error {
		return ProductBarcodeBuild(msg)
	})(msg)
}

// OnProductBarcodeDeleteMessageConsume - wrapper function for product barcode deletion processing
func OnProductBarcodeDeleteMessageConsume(msg string) error {
	return SafeConsumerWrapper("PRODUCT_BARCODE_WAREHOUSE", func(msg string) error {
		return OnConsumeMessageInventoryDelete(msg)
	})(msg)
}

// OnProductBarcodeBulkDeleteMessageConsume - wrapper function for bulk product barcode deletion processing
func OnProductBarcodeBulkDeleteMessageConsume(msg string) error {
	return SafeConsumerWrapper("PRODUCT_BARCODE_WAREHOUSE", func(msg string) error {
		return OnConsumeMessageInventoryBulkDelete(msg)
	})(msg)
}

// ProductBarcodeBuild - processes single product barcode update
func ProductBarcodeBuild(msg string) error { return consumeBarcodeSignals(msg, false) }

// Legacy entry points use the same primary-source reconciliation as Kafka.
// Keeping their signatures avoids leaving a snapshot-writing bypass for callers.
func ProductBarcodeInsertOrUpdateToPostgreSQL(productData models.MongoProductBarcodeModel) error {
	if err := normalizeProductBarcodeIdentity(&productData, true); err != nil {
		return err
	}
	return reconcileBarcodePayload(productData, false)
}

func ProductBarcodeBulkUpdateWithLogging(productDataList []models.MongoProductBarcodeModel) error {
	if _, _, err := normalizeProductBarcodeBatch(productDataList, true); err != nil {
		return err
	}
	return reconcileBarcodePayload(productDataList, true)
}

func ProductBarcodeDeleteFromPostgreSQL(productData models.MongoProductBarcodeModel) error {
	if err := normalizeProductBarcodeIdentity(&productData, false); err != nil {
		return err
	}
	return reconcileBarcodePayload(productData, false)
}

func ProductBarcodeBulkDeleteWithLogging(productDataList []models.MongoProductBarcodeModel) error {
	if _, _, err := normalizeProductBarcodeBatch(productDataList, false); err != nil {
		return err
	}
	return reconcileBarcodePayload(productDataList, true)
}

func reconcileBarcodePayload(payload interface{}, batch bool) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return consumeBarcodeSignals(string(raw), batch)
}

func normalizeProductBarcodeIdentity(productData *models.MongoProductBarcodeModel, requireItemCode bool) error {
	productData.HoldingCode = strings.TrimSpace(productData.HoldingCode)
	productData.BusinessCode = utils.NormalizeBusinessCode(productData.BusinessCode)
	productData.ItemCode = utils.NormalizeBusinessCode(productData.ItemCode)
	productData.Barcode = utils.NormalizeBusinessCode(productData.Barcode)

	if productData.HoldingCode == "" || productData.BusinessCode == "" || productData.Barcode == "" {
		return fmt.Errorf("holdingcode, businesscode and barcode are required")
	}
	if requireItemCode && productData.ItemCode == "" {
		return fmt.Errorf("itemcode is required for product barcode upsert")
	}
	return nil
}

func normalizeProductBarcodeBatch(productDataList []models.MongoProductBarcodeModel, requireItemCode bool) (string, string, error) {
	if len(productDataList) == 0 {
		return "", "", nil
	}
	for index := range productDataList {
		if err := normalizeProductBarcodeIdentity(&productDataList[index], requireItemCode); err != nil {
			return "", "", fmt.Errorf("barcode %d: %w", index+1, err)
		}
	}
	holdingCode := productDataList[0].HoldingCode
	businessCode := productDataList[0].BusinessCode
	for index := 1; index < len(productDataList); index++ {
		if productDataList[index].HoldingCode != holdingCode || productDataList[index].BusinessCode != businessCode {
			return "", "", fmt.Errorf("barcode %d belongs to a different holding or company", index+1)
		}
	}
	return holdingCode, businessCode, nil
}
