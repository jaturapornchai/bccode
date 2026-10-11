package generalledger

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// sourcePeriodMock เปิดธุรกรรมบน sqlmock แล้วคาดการตรวจงวดของวันที่หนึ่งวัน: ปีบัญชีที่ครอบคลุม → สถานะปีบัญชี → งวดล็อก
func sourcePeriodMock(t *testing.T, date string, yearFound, locked bool) (*sql.Tx, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	codeQuery := mock.ExpectQuery(`SELECT code FROM gl_records WHERE company=\$1 AND kind='fiscal-years'`).WithArgs("01", date)
	if !yearFound {
		codeQuery.WillReturnError(sql.ErrNoRows)
	} else {
		codeQuery.WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("2569"))
		mock.ExpectQuery(`FROM gl_records WHERE company=\$1 AND kind='fiscal-years' AND code=\$2`).WithArgs("01", "2569").
			WillReturnRows(sqlmock.NewRows([]string{"payload"}).AddRow([]byte(`{"code":"2569","isactive":true,"closed":false}`)))
		expectPeriodLocked(mock, date, locked)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	return tx, mock
}

func expectPeriodLocked(mock sqlmock.Sqlmock, date string, locked bool) {
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM gl_records WHERE company=\$1 AND kind='periods'`).WithArgs("01", date).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(locked))
}

func requireSourcePeriodLocked(t *testing.T, err error, date string) {
	t.Helper()
	user, ok := AsUserError(err)
	if !ok || user.Code != codeSourcePeriodLocked || len(user.Args) != 1 || user.Args[0] != date {
		t.Fatalf("error = %v, want %s with the date", err, codeSourcePeriodLocked)
	}
	if user.Field != "date" || user.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("field %q status %d, want date / 400", user.Field, user.HTTPStatus())
	}
	if !strings.Contains(user.Message, date) || !strings.Contains(user.Message, "ถูกล็อก") || strings.Contains(user.Message, "{0}") {
		t.Errorf("message must be the Thai text with the date filled in, got %q", user.Message)
	}
}

func TestCheckSourceDateOpen(t *testing.T) {
	ctx := context.Background()
	scope := Scope{Holding: "rungrueng", Company: "01", Actor: "demo"}

	t.Run("open period returns the fiscal year", func(t *testing.T) {
		tx, mock := sourcePeriodMock(t, "2026-10-08", true, false)
		year, err := CheckSourceDateOpen(ctx, tx, scope, "2026-10-08")
		if err != nil || year != "2569" {
			t.Fatalf("CheckSourceDateOpen = %q, %v; want 2569", year, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("locked period is a coded user error", func(t *testing.T) {
		tx, mock := sourcePeriodMock(t, "2026-10-08", true, true)
		_, err := CheckSourceDateOpen(ctx, tx, scope, "2026-10-08")
		requireSourcePeriodLocked(t, err, "2026-10-08")
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("missing fiscal year stops before the period check", func(t *testing.T) {
		tx, mock := sourcePeriodMock(t, "2031-01-05", false, false)
		_, err := CheckSourceDateOpen(ctx, tx, scope, "2031-01-05")
		if user, ok := AsUserError(err); !ok || user.Code != codeSourceFiscalYearNotFound {
			t.Fatalf("error = %v, want %s", err, codeSourceFiscalYearNotFound)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

// งวดล็อกของสมุดรายวันใช้ code เดียวกับเอกสารต้นทาง (แปลได้ทุกภาษา) แทนข้อความไทยดิบ
func TestCheckOpenDateReturnsCodedPeriodLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	expectPeriodLocked(mock, "2026-09-30", true)
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = (&PostgresStore{}).checkOpenDate(context.Background(), tx, Scope{Company: "01"}, "2026-09-30")
	requireSourcePeriodLocked(t, err, "2026-09-30")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
