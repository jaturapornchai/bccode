package httpapi

import (
	"context"
	"net/http"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	gl "smlcloudplatform/internal/generalledger"
)

func intPtr(v int) *int { return &v }

// วันเครดิตไม่ส่งมา = ของเจ้าหนี้ในทะเบียน; วันครบกำหนด = วันที่เอกสาร (ตามเวลาไทย) + วันเครดิต (Champ ChmFrmGoodsPurchasePage1.cpp:813-825)
func TestResolveDueDateUsesCreditorCreditDays(t *testing.T) {
	db, mock, _ := newRecordedMock(t)
	mock.ExpectQuery(`FROM gl_subledger_partners WHERE company = \$1 AND code = \$2`).WithArgs(testCompany, "S0012").
		WillReturnRows(sqlmock.NewRows([]string{"code", "title", "name", "is_supplier", "is_active", "credit_days"}).
			AddRow("S0012", "บริษัท", "เหล็กสยาม จำกัด", true, true, 30))
	doc := &ErpTransactionDoc{CustCode: " S0012 "}
	due, err := resolveDueDate(context.Background(), db, testCompany, mustDocType(t, "purchase"), doc, testDocDate)
	if err != nil {
		t.Fatal(err)
	}
	if !due.Valid || businessDay(due.Time) != "2026-11-07" || doc.DueDate != "2026-11-07" || doc.CreditDay == nil || *doc.CreditDay != 30 {
		t.Fatalf("due = %v %q credit %v; want 2026-11-07 and 30 days", due, doc.DueDate, doc.CreditDay)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// วันเครดิตและวันครบกำหนดที่ผู้ใช้กรอกมาใช้ตามนั้น ไม่อ่านทะเบียนซ้ำ
func TestResolveDueDateKeepsGivenValues(t *testing.T) {
	db, mock, _ := newRecordedMock(t)
	doc := &ErpTransactionDoc{CustCode: "S0012", CreditDay: intPtr(0), DueDate: "2026-12-15"}
	due, err := resolveDueDate(context.Background(), db, testCompany, mustDocType(t, "apotherdebt"), doc, testDocDate)
	if err != nil {
		t.Fatal(err)
	}
	if businessDay(due.Time) != "2026-12-15" || *doc.CreditDay != 0 {
		t.Fatalf("due = %v credit %d; want the given 2026-12-15 and 0", due, *doc.CreditDay)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestResolveDueDateRejectsBadValues(t *testing.T) {
	for _, tc := range []struct {
		name, code, field string
		doc               ErpTransactionDoc
	}{
		{"negative credit days", "erp_credit_days_invalid", "creditday", ErpTransactionDoc{CreditDay: intPtr(-1)}},
		{"credit days over max", "erp_credit_days_invalid", "creditday", ErpTransactionDoc{CreditDay: intPtr(gl.CreditDaysMax + 1)}},
		{"impossible date", "erp_due_date_invalid", "duedate", ErpTransactionDoc{CreditDay: intPtr(30), DueDate: "2026-02-30"}},
		{"due before document", "erp_due_before_doc", "duedate", ErpTransactionDoc{CreditDay: intPtr(30), DueDate: "2026-10-07"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, _, _ := newRecordedMock(t)
			doc := tc.doc
			_, err := resolveDueDate(context.Background(), db, testCompany, mustDocType(t, "creditorbeginningbalance"), &doc, testDocDate)
			user, ok := gl.AsUserError(err)
			if !ok || user.Code != tc.code || user.Field != tc.field || user.Status != http.StatusBadRequest {
				t.Fatalf("resolveDueDate error = %#v, want %s on %s (400)", err, tc.code, tc.field)
			}
		})
	}
}

// เอกสารที่ไม่มีช่องวันครบกำหนดล้างค่าที่ส่งมา และไม่แตะ perioddatetime เดิม (NULL)
func TestResolveDueDateIgnoredForOtherTypes(t *testing.T) {
	db, mock, _ := newRecordedMock(t)
	doc := &ErpTransactionDoc{CustCode: "C101", CreditDay: intPtr(15), DueDate: "2026-10-30"}
	due, err := resolveDueDate(context.Background(), db, testCompany, mustDocType(t, "sale-invoice"), doc, testDocDate)
	if err != nil || due.Valid || doc.CreditDay != nil || doc.DueDate != "" {
		t.Fatalf("due = %v err %v doc %v %q; want cleared", due, err, doc.CreditDay, doc.DueDate)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
