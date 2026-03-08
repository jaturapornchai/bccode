package process

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"

	mypg "smlcloudplatform/internal/goapi/mypg"
)

// ลบข้อมูลในตารางที่เกี่ยวข้องกับเอกสาร
func TruncateProcessTables(shopId string) error {
	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	tableList := []string{"stockwaitprocess", "docwaitprocess"}

	for _, table := range tableList {
		_, err := db.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s", table))
		if err != nil {
			return fmt.Errorf("error truncating %s table: %w", table, err)
		}
		logger.Info("Truncated %s table for shopId %s", table, shopId)
	}
	// insert queue to process
	db, err = mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Info("Database connection error: %v", err)
	}
	_, err = db.ExecContext(context.Background(), "insert into stockwaitprocess (itemcode) (select distinct itemcode from docdetail)")
	if err != nil {
		logger.Info("Database insert error: %v", err)
	}
	_, err = db.ExecContext(context.Background(), "insert into docwaitprocess (docno,transflag) (select distinct docno,transflag from doc)")
	if err != nil {
		logger.Info("Database insert error: %v", err)
	}

	return nil
}

// ลบข้อมูลในตารางที่เกี่ยวข้องกับเอกสาร
func TruncateDocTables(shopId string) error {
	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	tableList := []string{"doc", "docdetail"}

	for _, table := range tableList {
		_, err := db.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s", table))
		if err != nil {
			return fmt.Errorf("error truncating %s table: %w", table, err)
		}
		logger.Info("Truncated %s table for shopId %s", table, shopId)
	}

	return nil
}

// BulkInsertProcessStockCost ฟังก์ชันเฉพาะสำหรับ insert processstockcost
func BulkInsertProcessStockCost(ctx context.Context, db *sql.DB, data any) error {
	// Type assertion to get the actual data
	switch v := data.(type) {
	case []map[string]any:
		if len(v) == 0 {
			return nil
		}

		columns := []string{
			"docdatetime", "docno", "linenumber", "transflag", "itemcode",
			"unitcode", "whcode", "locationcode", "totalqty", "price", "unitstand",
			"unitdivide", "averagecost", "calcamount", "balanceamount", "balanceqty",
			"guid", "unitcost", "docref",
		}

		rows := make([][]any, len(v))
		for i, item := range v {
			rows[i] = []any{
				item["docdatetime"], item["docno"], item["linenumber"], item["transflag"],
				item["itemcode"], item["unitcode"], item["whcode"], item["locationcode"],
				item["totalqty"], item["price"], item["unitstand"], item["unitdivide"],
				item["averagecost"], item["calcamount"], item["balanceamount"],
				item["balanceqty"], item["guid"], item["unitcost"], item["docref"],
			}
		}

		return mypg.BulkInsertWithCopy(ctx, db, "processstockcost", columns, rows)
	default:
		return fmt.Errorf("unsupported data type for BulkInsertProcessStockCost")
	}
}

// BulkInsertProcessStockLot ฟังก์ชันเฉพาะสำหรับ insert processstocklot
func BulkInsertProcessStockLot(ctx context.Context, db *sql.DB, data any) error {
	// Type assertion to get the actual data
	switch v := data.(type) {
	case []map[string]any:
		if len(v) == 0 {
			return nil
		}

		columns := []string{
			"docdatetime", "lotnumber", "docno", "transflag", "itemcode",
			"unitcode", "whcode", "locationcode", "qty", "price", "unitstand",
			"unitdivide", "cost", "balanceamount", "balanceqty", "guidref",
		}

		rows := make([][]any, len(v))
		for i, item := range v {
			rows[i] = []any{
				item["docdatetime"], item["lotnumber"], item["docno"], item["transflag"],
				item["itemcode"], item["unitcode"], item["whcode"], item["locationcode"],
				item["qty"], item["price"], item["unitstand"], item["unitdivide"],
				item["cost"], item["balanceamount"], item["balanceqty"], item["guidref"],
			}
		}

		return mypg.BulkInsertWithCopy(ctx, db, "processstocklot", columns, rows)
	default:
		return fmt.Errorf("unsupported data type for BulkInsertProcessStockLot")
	}
}

// GetTransactionMultiplier is now in myglobal package
// Use myglobal.GetTransactionMultiplier instead
