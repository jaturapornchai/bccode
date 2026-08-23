package processstock

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestProcessProductBalanceUpdateByItemsRequiresCompanyScope(t *testing.T) {
	if err := ProcessProductBalanceUpdateByItems(nil, "", "COMPANY-A", []string{"ITEM-1"}); err == nil {
		t.Fatal("expected missing holdingcode to fail closed")
	}
	if err := ProcessProductBalanceUpdateByItems(nil, "HOLDING-A", "", []string{"ITEM-1"}); err == nil {
		t.Fatal("expected missing businesscode to fail closed")
	}
}

func TestBatchUpdateProductCompanyKeepsSameItemSeparateByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := `UPDATE product SET[\s\S]*WHERE holding_code = \$8 AND businesscode = \$9 AND itemcode IN \(\$1\)`
	item := []productQtyInfo{{ItemCode: "ITEM-1", BalanceQty: 1, PendingRecvQty: 2, PendingSendQty: 3}}

	for _, scope := range []struct{ holding, business string }{
		{"HOLDING-A", "COMPANY-A"},
		{"HOLDING-A", "COMPANY-B"},
		{"HOLDING-B", "COMPANY-A"},
		{"HOLDING-B", "COMPANY-B"},
	} {
		mock.ExpectExec(query).
			WithArgs("ITEM-1", float64(1), sqlmock.AnyArg(), float64(2), sqlmock.AnyArg(), float64(3), sqlmock.AnyArg(), scope.holding, scope.business).
			WillReturnResult(sqlmock.NewResult(0, 1))

		updated, err := batchUpdateProductCompany(db, scope.holding, scope.business, item, nil, nil)
		if err != nil {
			t.Fatalf("scope %s/%s: %v", scope.holding, scope.business, err)
		}
		if updated != 1 {
			t.Fatalf("scope %s/%s: updated=%d", scope.holding, scope.business, updated)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
