package kafka

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/product/projection"
)

// Structurally unusable payloads must be rejections (acknowledged, never
// retried), while the error text users already see is unchanged.
func TestIdentityFailuresAreRejections(t *testing.T) {
	barcode := models.MongoProductBarcodeModel{HoldingCode: "H", Barcode: "SKU"}
	err := normalizeProductBarcodeIdentity(&barcode, false)
	if !errors.Is(err, projection.ErrRejected) || err.Error() != "projection message rejected: holdingcode, businesscode and barcode are required" {
		t.Fatalf("missing businesscode: %v", err)
	}
	barcode.BusinessCode = "C"
	if err := normalizeProductBarcodeIdentity(&barcode, true); !errors.Is(err, projection.ErrRejected) {
		t.Fatalf("missing itemcode: %v", err)
	}
	product := MongoProductModel{HoldingCode: "H", BusinessCode: "C"}
	if err := normalizeProductSignal(&product); !errors.Is(err, projection.ErrRejected) {
		t.Fatalf("missing product code: %v", err)
	}
	if err := consumeBarcodeSignals("not json", true); !errors.Is(err, projection.ErrRejected) {
		t.Fatalf("unparsable bulk payload: %v", err)
	}
	if err := consumeProductSignal("[]"); !errors.Is(err, projection.ErrRejected) {
		t.Fatalf("unparsable product payload: %v", err)
	}
	if err := consumeBarcodeSignals("[]", true); err != nil {
		t.Fatalf("empty batch is a no-op: %v", err)
	}
}

func TestLockProjectionRowsCapsAdvisoryLocks(t *testing.T) {
	const holding, business = "H", "C"
	small := []string{"A", "B"}
	large := make([]string, maxRowLocksPerTransaction+1)
	for i := range large {
		large[i] = fmt.Sprintf("SKU%04d", i)
	}
	cases := []struct {
		name  string
		codes []string
		setup func(sqlmock.Sqlmock)
	}{
		{"small batch shares the company lock and locks each row", small, func(mock sqlmock.Sqlmock) {
			mock.ExpectExec(regexp.QuoteMeta(projection.CompanySharedSQL)).WithArgs(projection.LockKey("barcode", holding, business, "")).WillReturnResult(sqlmock.NewResult(0, 1))
			for _, code := range small {
				mock.ExpectExec(regexp.QuoteMeta(projection.ExclusiveSQL)).WithArgs(projection.LockKey("barcode", holding, business, code)).WillReturnResult(sqlmock.NewResult(0, 1))
			}
		}},
		{"large batch takes only the exclusive company lock", large, func(mock sqlmock.Sqlmock) {
			mock.ExpectExec(regexp.QuoteMeta(projection.ExclusiveSQL)).WithArgs(projection.LockKey("barcode", holding, business, "")).WillReturnResult(sqlmock.NewResult(0, 1))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			tc.setup(mock)
			mock.ExpectRollback()
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := lockProjectionRows(context.Background(), tx, "barcode", holding, business, tc.codes); err != nil {
				t.Fatal(err)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
