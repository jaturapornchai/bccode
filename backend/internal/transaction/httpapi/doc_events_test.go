package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/pkg/microservice"
)

// expectDocEvents คาดการอ่านเอกสารก่อน แล้วอ่าน erp_doc_events เรียงตามลำดับที่บันทึก:
// แถวแรกระบบเขียน (ปรับปรุงสถานะใบรับสินค้า) แถวสองผู้ใช้ยกเลิกโดยพิมพ์เหตุผลให้หน้าตาเหมือน key ของระบบ
func expectDocEvents(mock sqlmock.Sqlmock, flag int, docNo string) {
	expectGetDoc(mock, flag, docNo, true, "", eventGRClosed)
	stamp := time.Date(2026, 10, 9, 3, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT id, businesscode, transflag, docno, action, from_status, to_status, reason, actor, created_at\s+FROM erp_doc_events\s+WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3\s+ORDER BY id ASC`).
		WithArgs(testCompany, flag, docNo).
		WillReturnRows(sqlmock.NewRows([]string{"id", "businesscode", "transflag", "docno", "action", "from_status", "to_status", "reason", "actor", "created_at"}).
			AddRow(1, testCompany, flag, docNo, "adjust_status", "", "closed", eventGRClosed, testActor, stamp).
			AddRow(2, testCompany, flag, docNo, "cancel", "", "cancelled", eventGRClosed, "somchai", stamp.Add(time.Hour)))
}

func TestGetDocEventsReadsDocThenEventsInOrder(t *testing.T) {
	gr := mustDocType(t, "goodsreceive")
	db, mock, _ := newRecordedMock(t)
	expectDocEvents(mock, gr.Flag, "GR6910-0001")

	events, err := NewStore().GetDocEvents(context.Background(), db, testCompany, gr.Flag, "GR6910-0001")
	if err != nil || len(events) != 2 {
		t.Fatalf("GetDocEvents = %+v, %v", events, err)
	}
	// ชั้น store คืนเหตุผลตามที่เก็บ (แปลภาษาที่ handler)
	if events[0].Action != "adjust_status" || events[0].ToStatus != "closed" || events[0].Reason != eventGRClosed ||
		events[1].Action != "cancel" || events[1].ToStatus != "cancelled" || events[1].Actor != "somchai" {
		t.Fatalf("events = %+v", events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetDocEventsMissingDocIsNotFound(t *testing.T) {
	gr := mustDocType(t, "goodsreceive")
	db, mock, _ := newRecordedMock(t)
	mock.ExpectQuery(`FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND NOT COALESCE\(isdelete, false\) AND docno = \$3 LIMIT 1`).
		WithArgs(testCompany, gr.Flag, "GR6910-0099").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	if _, err := NewStore().GetDocEvents(context.Background(), db, testCompany, gr.Flag, "GR6910-0099"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetDocEvents(missing) error = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// GET /transaction/<module>/<id>/events: แปลเฉพาะเหตุผลของเหตุการณ์ที่ระบบเขียน ตามภาษาของคำขอ;
// เหตุผลที่ผู้ใช้พิมพ์ตอนยกเลิกคืนตามเดิม แม้หน้าตาเหมือน key ของระบบ
func TestGetEventsHandlerTranslatesOnlySystemReasons(t *testing.T) {
	gr := mustDocType(t, "goodsreceive")
	for _, tc := range []struct{ query, wantSystem string }{
		{"", "ปิดใบรับสินค้า (ปรับปรุงสถานะใบรับสินค้า)"},
		{"?lang=en", "Goods receipt closed (status adjustment)"},
	} {
		t.Run("lang"+tc.query, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			mock.ExpectBegin()
			mock.ExpectExec(`SELECT pg_advisory_xact_lock`).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectExec(`CREATE TABLE IF NOT EXISTS erp_doc_events`).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectCommit()
			expectDocEvents(mock, gr.Flag, "GR6910-0001")

			h := &Http{
				store: NewStore(),
				session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
					return glhttp.NewSessionAccess(testScope(), "*"), nil
				},
				connect: func(holding string) (*sql.DB, error) {
					if holding != testScope().Holding {
						return nil, fmt.Errorf("unexpected holding %q", holding)
					}
					return db, nil
				},
			}
			request := newPermContext(gr.Key, tc.query)
			request.params["id"] = "GR6910-0001"
			if err := h.getEvents(request); err != nil {
				t.Fatal(err)
			}
			body, ok := request.body.(DocEventsResponse)
			if request.status != http.StatusOK || !ok || !body.Success || len(body.Data) != 2 {
				t.Fatalf("response = %d %+v", request.status, request.body)
			}
			if body.Data[0].Action != "adjust_status" || body.Data[0].Reason != tc.wantSystem {
				t.Errorf("system event reason = %q, want %q", body.Data[0].Reason, tc.wantSystem)
			}
			if body.Data[1].Action != "cancel" || body.Data[1].Reason != eventGRClosed {
				t.Errorf("user-typed cancel reason = %q, want it unchanged (%q)", body.Data[1].Reason, eventGRClosed)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}
