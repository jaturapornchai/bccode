//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"os"
	"testing"

)

// TestTaxWithholdingReportFromGL ตรวจรายงานภาษีหัก ณ ที่จ่ายกับฐานทดสอบที่มีข้อมูล GL จริง
// (seed ผ่าน backend/cmd/glseed — ใบจ่ายเจ้าหนี้หัก 3% และใบค่าเช่าหัก 5% ลงบัญชี ภ.ง.ด.53)
//
// ตั้งสนามทดสอบ:
//
//	docker run -d --name bc-gl-wht-test-pg18 -e POSTGRES_HOST_AUTH_METHOD=trust \
//	  -e POSTGRES_DB=restore -p 127.0.0.1:55443:5432 postgres:18-alpine
//	(restore ฐาน rungrueng + รัน glseed -apply ตาม docs)
//	BC_GL_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:55443/rungrueng?sslmode=disable'
func TestTaxWithholdingReportFromGL(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to a GL test database with seeded WHT journals")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	// 1) paid ทุกแบบยื่น งวด ม.ค.: ใบจ่ายเจ้าหนี้หัก 3% (3,210/107,000)
	rows, total, basetotal, whttotal, note, err := buildWithholdingRows(ctx, db, "01", 2026, 1, "paid", nil, 200, 0)
	if err != nil {
		t.Fatalf("paid report: %v", err)
	}
	if total != 1 {
		t.Fatalf("paid rows=%d คาด 1 (rows=%+v)", total, rows)
	}
	if basetotal != 107000 || whttotal != 3210 {
		t.Fatalf("paid totals base=%f wht=%f คาด 107000/3210", basetotal, whttotal)
	}
	if note != "" {
		t.Fatalf("ไม่ควรมี note เมื่อพบบัญชี: %s", note)
	}

	pv1, ok := rows[0], rows[0].DocNo == "PV6901-S001"
	if !ok {
		t.Fatalf("แถวแรกไม่ใช่ PV6901-S001: %+v", rows[0])
	}
	if pv1.PartnerCode != "SUPP-TH-001" || pv1.TaxID != "0105558002001" {
		t.Fatalf("PV6901-S001 partner=%q taxid=%q", pv1.PartnerCode, pv1.TaxID)
	}
	if pv1.WhtAmount != 3210 || pv1.BaseAmount != 107000 || pv1.RatePercent != 3 {
		t.Fatalf("PV6901-S001 tax=%f base=%f rate=%f คาด 3210/107000/3", pv1.WhtAmount, pv1.BaseAmount, pv1.RatePercent)
	}

	// 1.1) งวด ก.พ.: ใบค่าเช่าหัก 5% (1,000/20,000)
	_, totalFeb, baseFeb, whtFeb, _, err := buildWithholdingRows(ctx, db, "01", 2026, 2, "paid", nil, 200, 0)
	if err != nil {
		t.Fatalf("paid feb: %v", err)
	}
	if totalFeb != 1 || baseFeb != 20000 || whtFeb != 1000 {
		t.Fatalf("paid feb rows=%d base=%f wht=%f คาด 1/20000/1000", totalFeb, baseFeb, whtFeb)
	}

	// 2) กรองเฉพาะแบบยื่น ภ.ง.ด.53 ต้องได้ใบจ่ายเจ้าหนี้นิติบุคคล
	_, total53, _, wht53, _, err := buildWithholdingRows(ctx, db, "01", 2026, 1, "paid", []string{"53"}, 200, 0)
	if err != nil {
		t.Fatalf("form 53: %v", err)
	}
	if total53 != 1 || wht53 != 3210 {
		t.Fatalf("form 53 rows=%d wht=%f คาด 1/3210", total53, wht53)
	}

	// 3) ภ.ง.ด.3 ในงวดนี้ไม่มีรายการ (แถวร่างจะไม่เข้า GL lines) — คืนว่างโดยไม่ error
	_, total3, _, _, _, err := buildWithholdingRows(ctx, db, "01", 2026, 1, "paid", []string{"3"}, 200, 0)
	if err != nil {
		t.Fatalf("form 3: %v", err)
	}
	if total3 != 0 {
		t.Fatalf("form 3 rows=%d คาด 0", total3)
	}

	// 4) received: ยังไม่มีใบที่เราถูกหักในงวดนี้ — ว่างโดยไม่ error
	_, totalR, _, _, _, err := buildWithholdingRows(ctx, db, "01", 2026, 1, "received", nil, 200, 0)
	if err != nil {
		t.Fatalf("received: %v", err)
	}
	if totalR != 0 {
		t.Fatalf("received rows=%d คาด 0", totalR)
	}

	// 5) แบบยื่นนอก whitelist ต้องถูกปฏิเสธ ไม่ใช่รัน query
	if _, _, _, _, _, err := buildWithholdingRows(ctx, db, "01", 2026, 1, "paid", []string{"99"}, 200, 0); err == nil {
		t.Fatal("ระบบยอมรับแบบยื่นที่ไม่อยู่ใน whitelist")
	}
}
