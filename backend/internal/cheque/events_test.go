package cheque

import (
	"context"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// restoreStates ต้องเขียนทะเบียนเรียงตาม id เหมือน lockDocCheques/ApplyRecalculate ไม่ใช่ตามลำดับที่ RETURNING คืนมา
// — ลำดับล็อกแถวต่างกันระหว่างสองคำขอที่แตะเช็คชุดเดียวกันทำให้ PostgreSQL deadlock (40P01)
func TestRestoreStatesUpdatesChequesInIDOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT DISTINCT ON \(cheque_id\)`).WillReturnRows(sqlmock.NewRows([]string{"cheque_id", "to_status", "book_code", "event_date"}).
		AddRow(chequeA, InDeposited, "SCB01", "2026-10-03").AddRow(chequeB, InOnHand, "", "2026-10-01"))
	for _, id := range []string{chequeA, chequeB} {
		mock.ExpectExec(`UPDATE cheque_registers SET status`).WithArgs("01", id, sqlmock.AnyArg(), sqlmock.AnyArg(), "somchai").
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	// voidDocEvents คืน id ตามลำดับของ RETURNING (ไม่เรียง)
	if err = restoreStates(context.Background(), tx, testScope, []string{chequeB, chequeA}); err != nil {
		t.Fatalf("restoreStates: %v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
