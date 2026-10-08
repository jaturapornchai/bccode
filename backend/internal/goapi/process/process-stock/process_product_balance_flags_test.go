package processstock

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/doctype"
)

// ค้างรับนับจากใบซื้อและใบรับสินค้าจากการซื้อที่อ้างใบสั่งซื้อ ค้างส่งนับจากใบขายที่อ้างใบสั่งขาย
// ใบปรับปรุงสต็อกต้องไม่ถูกนับเป็นการรับของตามใบสั่งซื้อ
func TestPendingQuantitiesUseRegistryDocumentTypes(t *testing.T) {
	recv := regexp.QuoteMeta(fmt.Sprintf("dd.transflag = %d AND d.isclosed = false", doctype.PurchaseOrder)) +
		`[\s\S]*` + regexp.QuoteMeta(fmt.Sprintf("dd.transflag IN (%d, %d)", doctype.Purchase, doctype.GoodsReceive))
	send := regexp.QuoteMeta(fmt.Sprintf("dd.transflag = %d AND d.isclosed = false", doctype.SaleOrder)) +
		`[\s\S]*` + regexp.QuoteMeta(fmt.Sprintf("dd.transflag = %d", doctype.SaleInvoice))
	items := []string{"P-1001"}
	cases := []struct {
		name    string
		pattern string
		args    []driver.Value
		run     func(*sql.DB) (map[string]float64, error)
	}{
		{"pending receive", recv, []driver.Value{"01"}, func(db *sql.DB) (map[string]float64, error) {
			return queryPendingRecvAll(db, "01")
		}},
		{"pending receive by items", recv, []driver.Value{"01", "P-1001", "P-1001"}, func(db *sql.DB) (map[string]float64, error) {
			return queryPendingRecvByItems(db, "01", items)
		}},
		{"pending send", send, []driver.Value{"01"}, func(db *sql.DB) (map[string]float64, error) {
			return queryPendingSendAll(db, "01")
		}},
		{"pending send by items", send, []driver.Value{"01", "P-1001", "P-1001"}, func(db *sql.DB) (map[string]float64, error) {
			return queryPendingSendByItems(db, "01", items)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery(tc.pattern).WithArgs(tc.args...).
				WillReturnRows(sqlmock.NewRows([]string{"itemcode", "pending"}).AddRow("P-1001", 4.0))
			got, err := tc.run(db)
			if err != nil {
				t.Fatal(err)
			}
			if got["P-1001"] != 4 {
				t.Errorf("pending = %v, want 4", got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}
