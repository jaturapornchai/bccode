package kafka

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/lib/pq"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"smlcloudplatform/internal/goapi/models"
)

func TestReplaceProductBarcodeRowsTransaction(t *testing.T) {
	for _, mode := range []string{"success", "delete failure", "copy failure", "commit failure"} {
		t.Run(mode, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			deletion := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM productbarcode WHERE holding_code = $1 AND businesscode = $2 AND barcode = $3")).WithArgs("H", "A", "SKU")
			if mode == "delete failure" {
				deletion.WillReturnError(errors.New("delete rejected"))
				mock.ExpectRollback()
			} else {
				deletion.WillReturnResult(sqlmock.NewResult(0, 1))
				copyStmt := mock.ExpectPrepare(regexp.QuoteMeta(pq.CopyIn("productbarcode", "holding_code", "businesscode", "barcode", "price1"))).WillBeClosed()
				copyStmt.ExpectExec().WithArgs("H", "A", "SKU", "0.3000").WillReturnResult(sqlmock.NewResult(0, 1))
				flush := copyStmt.ExpectExec().WithArgs()
				if mode == "copy failure" {
					flush.WillReturnError(errors.New("copy rejected"))
					mock.ExpectRollback()
				} else {
					flush.WillReturnResult(sqlmock.NewResult(0, 1))
					if mode == "commit failure" {
						mock.ExpectCommit().WillReturnError(errors.New("commit rejected"))
					} else {
						mock.ExpectCommit()
					}
				}
			}
			err = replaceProductBarcodeRows(context.Background(), db, []models.MongoProductBarcodeModel{{HoldingCode: "H", BusinessCode: "A", Barcode: "SKU"}}, []string{"holding_code", "businesscode", "barcode", "price1"}, [][]any{{"H", "A", "SKU", "0.3000"}})
			if (mode == "success") != (err == nil) {
				t.Fatalf("wrong outcome: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
