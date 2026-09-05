package kafka

import (
	"context"
	"database/sql"
	"fmt"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mypg"
)

// Replacement must be atomic: a failed COPY must leave the previous company
// rows intact so Kafka can safely retry the same batch.
func replaceProductBarcodeRows(ctx context.Context, db *sql.DB, products []models.MongoProductBarcodeModel, columns []string, records [][]any) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, product := range products {
		if _, err := tx.ExecContext(ctx, "DELETE FROM productbarcode WHERE holding_code = $1 AND businesscode = $2 AND barcode = $3", product.HoldingCode, product.BusinessCode, product.Barcode); err != nil {
			return fmt.Errorf("delete previous barcode row: %w", err)
		}
	}
	// A PostgreSQL transaction that encountered an error must be rolled back;
	// retrying COPY inside that same transaction cannot recover it.
	if err := mypg.BulkInsertWithCopyConfig(ctx, tx, "productbarcode", columns, records, mypg.BulkInsertConfig{MaxRetries: 1}); err != nil {
		return fmt.Errorf("copy replacement barcode rows: %w", err)
	}
	return tx.Commit()
}
