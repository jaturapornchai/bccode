package mypg

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
)

// firstLangName ดึงชื่อภาษาแรกจาก []LanguageModel
func firstLangName(names []models.LanguageModel) string {
	if len(names) > 0 {
		return names[0].Name
	}
	return ""
}

// ProductBarcodeUpdate อัพเดทข้อมูลสินค้าในฐานข้อมูล PostgreSQL
func ProductBarcodeUpdate(productData models.MongoProductBarcodeModel) error {
	db, err := PgSqlFastConnect(productData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	ctx := context.Background()

	// สร้าง checksum
	productDataStr := fmt.Sprintf("%v", productData)
	checkSumMongodb := myglobal.CalculateMD5(productDataStr)
	updatePostgreSQL := false

	// เช็คว่ามีข้อมูลอยู่แล้วหรือไม่และ checksum ตรงกันหรือไม่
	query := "SELECT checksum FROM productbarcode WHERE holding_code = $1 AND barcode = $2"
	dataRows, err := QuerySelectAll(db, query, productData.HoldingCode, productData.Barcode)
	if err != nil {
		return fmt.Errorf("failed to check existing product: %w", err)
	}

	if len(dataRows) > 0 {
		if existingChecksum, ok := dataRows[0]["checksum"].(string); !ok || existingChecksum != checkSumMongodb {
			updatePostgreSQL = true
		}
	} else {
		updatePostgreSQL = true
	}

	if updatePostgreSQL {
		// เริ่ม transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		// ลบข้อมูลเก่า
		_, err = tx.ExecContext(ctx, "DELETE FROM productbarcode WHERE holding_code = $1 AND barcode = $2", productData.HoldingCode, productData.Barcode)
		if err != nil {
			return fmt.Errorf("failed to delete existing product: %w", err)
		}

		// เตรียมข้อมูลสำหรับ insert
		productName := firstLangName(productData.Names)
		productGroupName := firstLangName(productData.GroupNames)
		productUnitName := firstLangName(productData.ItemUnitNames)
		brandNames := firstLangName(productData.BrandNames)
		categoryNames := firstLangName(productData.CategoryNames)
		classNames := firstLangName(productData.ClassNames)
		designNames := firstLangName(productData.DesignNames)
		gradeNames := firstLangName(productData.GradeNames)
		modelNames := firstLangName(productData.ModelNames)
		patternNames := firstLangName(productData.PatternNames)
		groupSubOneNames := firstLangName(productData.GroupSubOneNames)
		groupSubTwoNames := firstLangName(productData.GroupSubTwoNames)

		price := 0.0
		priceRetail := 0.0
		if len(productData.Prices) > 0 {
			price = productData.Prices[0].Price
			// หา price keynumber=1 เป็นราคาขายปลีก
			for _, p := range productData.Prices {
				if p.KeyNumber == 1 {
					priceRetail = p.Price
					break
				}
			}
		}

		// barcoderef จาก refbarcodes ตัวแรก
		barcodeRef := ""
		if len(productData.RefBarCodes) > 0 {
			barcodeRef = productData.RefBarCodes[0].Barcode
		}

		// Insert ข้อมูลใหม่
		insertQuery := `
			INSERT INTO productbarcode (
				barcode, barcoderef, itemcode, name0, unitcode, unitname,
				groupcode, groupnames, price1, price_retail,
				barcoderefunitstand, barcoderefunitdivide,
				isstock, itemtype, materialtype, checksum,
				holding_code, guidfixed, imageuri, isusesubbarcodes,
				brandcode, brandnames, categorycode, categorynames,
				classcode, classnames, designcode, designnames,
				gradecode, gradenames, modelcode, modelnames,
				patterncode, patternnames,
				groupsubonecode, groupsubonenames,
				groupsubtwocode, groupsubtwonames
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, $16,
				$17, $18, $19, $20,
				$21, $22, $23, $24,
				$25, $26, $27, $28,
				$29, $30, $31, $32,
				$33, $34,
				$35, $36,
				$37, $38
			)`

		_, err = tx.ExecContext(ctx, insertQuery,
			productData.Barcode, barcodeRef, productData.ItemCode, productName, productData.ItemUnitCode, productUnitName,
			productData.GroupCode, productGroupName, price, priceRetail,
			productData.StandValue, productData.DivideValue,
			boolToStock(productData.IsUseSubBarcodes), productData.ItemType, productData.MaterialType, checkSumMongodb,
			productData.HoldingCode, productData.GuidFixed, productData.ImageUri, productData.IsUseSubBarcodes,
			productData.BrandCode, brandNames, productData.CategoryCode, categoryNames,
			productData.ClassCode, classNames, productData.DesignCode, designNames,
			productData.GradeCode, gradeNames, productData.ModelCode, modelNames,
			productData.PatternCode, patternNames,
			productData.GroupSubOneCode, groupSubOneNames,
			productData.GroupSubTwoCode, groupSubTwoNames,
		)

		if err != nil {
			return fmt.Errorf("failed to insert productbarcode: %w", err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		logger.Success("updated product %s", productData.Barcode)
	}

	return nil
}

// boolToStock แปลง isusesubbarcodes เป็น isstock (1 = มี ref, 0 = ไม่มี)
func boolToStock(isusesubbarcodes bool) int {
	if isusesubbarcodes {
		return 1
	}
	return 0
}
