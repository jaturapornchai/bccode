package process

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"

	mypg "smlcloudplatform/internal/goapi/mypg"
)

// ลบข้อมูลในตารางที่เกี่ยวข้องกับเอกสาร
func TruncateProcessTables(holdingCode string) error {
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	tableList := []string{"stockwaitprocess", "docwaitprocess"}

	for _, table := range tableList {
		_, err := db.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s", table))
		if err != nil {
			return fmt.Errorf("error truncating %s table: %w", table, err)
		}
		logger.Info("Truncated %s table for holdingCode %s", table, holdingCode)
	}
	// insert queue to process
	db, err = mypg.PgSqlFastConnect(holdingCode)
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
func TruncateDocTables(holdingCode string) error {
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	tableList := []string{"doc", "docdetail"}

	for _, table := range tableList {
		_, err := db.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s", table))
		if err != nil {
			return fmt.Errorf("error truncating %s table: %w", table, err)
		}
		logger.Info("Truncated %s table for holdingCode %s", table, holdingCode)
	}

	return nil
}

// GetTransactionMultiplier is now in myglobal package
// Use myglobal.GetTransactionMultiplier instead
