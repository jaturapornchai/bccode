package httpapi

import (
	"context"
	"database/sql/driver"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/doctype"
)

// เคยส่ง args ที่ตัดท้ายสองตัวให้ SELECT docno ทำให้คิวรีล้มเงียบ ๆ แล้วไม่ลบ stock_ledger ของเอกสารที่ถูกลบ
func TestDeleteDocRemovesStockLedgerForBothLookupForms(t *testing.T) {
	const docNo = "PO2610-0001"
	for _, tc := range []struct {
		name       string
		docIDOrNo  string
		selectArgs []driver.Value
		updateArgs []driver.Value
	}{
		{"by docno", docNo,
			[]driver.Value{testCompany, doctype.Purchase, docNo},
			[]driver.Value{testCompany, doctype.Purchase, docNo, testActor, sqlmock.AnyArg()}},
		{"by numeric id", "7",
			[]driver.Value{testCompany, doctype.Purchase, 7, "7"},
			[]driver.Value{testCompany, doctype.Purchase, 7, "7", testActor, sqlmock.AnyArg()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT docno FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND`).
				WithArgs(tc.selectArgs...).
				WillReturnRows(sqlmock.NewRows([]string{"docno"}).AddRow(docNo))
			mock.ExpectExec(`DELETE FROM stock_ledger WHERE businesscode=\$1 AND docno=\$2`).
				WithArgs(testCompany, docNo).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec(`UPDATE doc SET isdelete = true`).
				WithArgs(tc.updateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			if err := NewStore().DeleteDoc(context.Background(), db, testCompany, doctype.Purchase, tc.docIDOrNo, testActor); err != nil {
				t.Fatalf("DeleteDoc(%s) = %v", tc.docIDOrNo, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}
