package build

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	mypg "smlcloudplatform/internal/goapi/mypg"
)

func processProductByBarcodeBuild(holdingCode string) error {
	logger.Info("Process: Product by Barcode Build")

	pgDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return fmt.Errorf("FastConnect error: %w", err)
	}

	startTime := time.Now()
	logger.Info("Build started at: %v", startTime)
	{
		// query ลบรายการ product ที่ไม่มี ใน product
		deleteProductWithoutBarcodeQuery := `
			DELETE FROM product
			WHERE itemcode NOT IN (SELECT DISTINCT itemcode FROM productbarcode);
		`
		if _, err := pgDB.ExecContext(context.Background(), deleteProductWithoutBarcodeQuery); err != nil {
			return fmt.Errorf("delete product without barcode: %w", err)
		}
	}
	{
		// query สร้างรายการ product จาก productbarcode (เอาเฉพาะหน่วยนับ 1:1)
		insertProductFromBarcodeQuery := `
			INSERT INTO product (itemcode, name0, unitcode, unitname)
				SELECT DISTINCT ON (itemcode) 
					itemcode, 
					name0, 
					unitcode, 
					unitname
				FROM productbarcode 
				WHERE barcoderefunitstand=1 AND barcoderefunitdivide=1 AND itemcode IS NOT NULL and itemcode <> ''
				ORDER BY itemcode, name0, unitcode, unitname  -- กำหนดลำดับที่ต้องการ
			ON CONFLICT (itemcode) 
			DO UPDATE SET 
				name0 = EXCLUDED.name0,
				unitcode = EXCLUDED.unitcode,
				unitname = EXCLUDED.unitname
			WHERE 
				(product.name0, product.unitcode, product.unitname) IS DISTINCT FROM 
				(EXCLUDED.name0, EXCLUDED.unitcode, EXCLUDED.unitname);
		`
		if _, err := pgDB.ExecContext(context.Background(), insertProductFromBarcodeQuery); err != nil {
			return fmt.Errorf("insert product from barcode: %w", err)
		}
	}
	return nil
}
