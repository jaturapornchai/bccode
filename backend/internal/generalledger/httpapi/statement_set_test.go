package httpapi

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"

	gl "smlcloudplatform/internal/generalledger"
)

// ชุดงบการเงินใช้สิทธิ์จอออกแบบงบการเงินเหมือนงบทีละงบ และไม่เปิดผ่าน MCP gl_report (งาน MCP พักไว้)
func TestStatementSetPermissionScreen(t *testing.T) {
	if !canReadReport(map[string]bool{"financial-statement-designer": true}, statementSetReport) || !canReadReport(map[string]bool{"*": true}, statementSetReport) {
		t.Fatal("statement designer cannot read the statement set")
	}
	for _, screen := range []string{"profit-loss", "balance-sheet", "gl-journals", "trial-balance"} {
		if canReadReport(map[string]bool{screen: true}, statementSetReport) {
			t.Fatalf("screen %s must not read the statement set", screen)
		}
	}
	if _, listed := reportScreens[statementSetReport]; listed {
		t.Fatal("statement-set is listed in reportScreens — that opens it to MCP gl_report while MCP work is on hold")
	}
}

func TestStatementSetQueryParams(t *testing.T) {
	base := gl.ReportQuery{FiscalYear: "2569", BranchCode: "B1", Template: "IGNORED"}
	parse := func(raw string) (gl.StatementSetQuery, error) {
		t.Helper()
		values, err := url.ParseQuery(raw)
		if err != nil {
			t.Fatal(err)
		}
		return statementSetQuery(values, base)
	}
	q, err := parse("")
	if err != nil || q.Templates != nil || !q.Notes || q.Report.FiscalYear != "2569" || q.Report.BranchCode != "B1" || q.Report.Template != "" {
		t.Fatalf("defaults = %+v, %v", q, err)
	}
	// templates= ว่าง = หมายเหตุอย่างเดียว ต่างจากไม่ส่ง (ชุดเริ่มต้น)
	if q, err = parse("templates=&notes=true"); err != nil || q.Templates == nil || len(q.Templates) != 0 || !q.Notes {
		t.Fatalf("notes only = %+v, %v", q, err)
	}
	if q, err = parse("templates=PL-1,BS-1,PL-1&notes=false"); err != nil || !reflect.DeepEqual(q.Templates, []string{"PL-1", "BS-1"}) || q.Notes {
		t.Fatalf("given = %+v, %v", q, err)
	}
	if _, err = parse("notes=yes"); err == nil {
		t.Fatal("invalid notes accepted")
	}
	if _, err = parse("templates=" + strings.Repeat("X,", 5) + "A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U"); err == nil {
		t.Fatal("more than 20 templates accepted")
	}
}

// เหตุที่งบหนึ่งคำนวณไม่สำเร็จแปลตามภาษา (key gl_err_*) พร้อมรหัสรูปแบบ; ปัญหาฝั่งระบบไม่ส่งข้อความเทคนิค
func TestStatementSetSectionErrorsAreLocalized(t *testing.T) {
	notFound := &gl.UserError{Code: "statement_set_template_not_found", Status: 400, Field: "templates", Message: "ไม่พบรูปแบบงบการเงินรหัส FS-9 หรือถูกลบไปแล้ว กรุณาปิดแล้วเปิดหน้าพิมพ์ชุดงบใหม่", Args: []string{"FS-9"}}
	status, en := errorPayloadFor(notFound, "en")
	if status != 400 || en.Code != "statement_set_template_not_found" || !strings.Contains(en.Message, "FS-9") || strings.Contains(en.Message, "{0}") || en.MessageTH != notFound.Message {
		t.Fatalf("en payload = %d %+v", status, en)
	}
	if _, th := errorPayloadFor(notFound, "th"); th.Message != notFound.Message {
		t.Fatalf("th payload = %+v", th)
	}

	report := &gl.Report{Rows: []map[string]string{}}
	set := gl.StatementSet{Sections: []gl.StatementSetSection{
		{Code: "BS", Report: report},
		{Code: "EQ", Err: notFound},
		{Code: "CF", Err: errors.New("pq: connection reset by peer")},
	}}
	localizeStatementSet(&set, 42, "en")
	if set.Sequence != 42 || set.Sections[0].Report.Sequence != 42 || set.Sections[0].Error != "" {
		t.Fatalf("ready section = %+v", set.Sections[0])
	}
	if !strings.Contains(set.Sections[1].Error, "FS-9") || strings.ContainsAny(set.Sections[1].Error, "กขค") {
		t.Fatalf("user error section = %q", set.Sections[1].Error)
	}
	if set.Sections[2].Error == "" || strings.Contains(set.Sections[2].Error, "pq:") {
		t.Fatalf("system error leaked = %q", set.Sections[2].Error)
	}
}
