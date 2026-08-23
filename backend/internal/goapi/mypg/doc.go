package mypg

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
)

type execContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func DeleteDocPgSql(ctx context.Context, db *sql.DB, businessCode, docNo string, transFlag int) {
	if err := deleteDocPgSqlExec(ctx, db, businessCode, docNo, transFlag); err != nil {
		logger.Warn("DeleteDocPgSql failed: %v", err)
	}
}

func DeleteDocPgSqlTx(ctx context.Context, tx *sql.Tx, businessCode, docNo string, transFlag int) error {
	return deleteDocPgSqlExec(ctx, tx, businessCode, docNo, transFlag)
}

func deleteDocPgSqlExec(ctx context.Context, exec execContext, businessCode, docNo string, transFlag int) error {
	if businessCode == "" {
		return fmt.Errorf("businesscode is required")
	}
	queries := []string{
		"DELETE FROM doc WHERE businesscode = $1 AND docno = $2 AND transflag = $3",
		"DELETE FROM docdetail WHERE businesscode = $1 AND docno = $2 AND transflag = $3",
	}

	for _, query := range queries {
		if _, err := exec.ExecContext(ctx, query, businessCode, docNo, transFlag); err != nil {
			return fmt.Errorf("execute delete query '%s': %w", query, err)
		}
	}

	if _, err := exec.ExecContext(ctx, "DELETE FROM docpayment WHERE businesscode = $1 AND docno = $2", businessCode, docNo); err != nil {
		return fmt.Errorf("delete docpayment: %w", err)
	}

	return nil
}

// addToDocWaitProcessQueues - adds documents to processing queues
func AddToDocWaitProcessQueues(ctx context.Context, db *sql.DB, docNo string, transFlag int) {

	// Add to doc wait process
	_, err := db.ExecContext(ctx,
		"INSERT INTO docwaitprocess (docno, transflag) VALUES ($1, $2)",
		docNo, transFlag)
	if err != nil {
		logger.Warn("Could not add to doc queue: %v", err)
	}

	// Add to stock wait process for all items in the document
	insertStockWaitQuery := `
		INSERT INTO stockwaitprocess (itemcode)
		SELECT DISTINCT itemcode
		FROM docdetail
		WHERE docno = $1 AND transflag = $2`

	_, err = db.ExecContext(ctx, insertStockWaitQuery, docNo, transFlag)
	if err != nil {
		logger.Warn("Could not add to stock queue: %v", err)
	}
}
