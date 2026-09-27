package httpapi

import (
	"context"
	"log"
	"net/url"

	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/pkg/microservice"
)

// statementSetReport GET /gl/v2/reports/statement-set — ชุดงบการเงินในคำขอเดียว (gl.StatementSet, จอพิมพ์ชุดงบการเงิน).
// สิทธิ์เท่ากับ reports/statement (จอออกแบบงบการเงิน — canReadReport). ตั้งใจไม่ใส่ใน reportScreens: MCP gl_report เปิดเฉพาะ
// รายงานใน reportScreens และงาน MCP พักไว้ตามคำสั่งลุงจืด 2026-09-25 — ชุดงบจึงยังไม่เปิดผ่าน MCP
const statementSetReport = "statement-set"

// statementSetQuery อ่านพารามิเตอร์ของชุดงบ: templates ไม่ส่ง = ชุดเริ่มต้น, ส่งค่าว่าง = ไม่พิมพ์งบ (หมายเหตุอย่างเดียว);
// notes ไม่ส่ง = รวมหมายเหตุ. ตัวกรองอื่นใช้ของ reports/statement (report) — template เดี่ยวไม่มีผลกับชุด
func statementSetQuery(values url.Values, report gl.ReportQuery) (gl.StatementSetQuery, error) {
	report.Template = ""
	q := gl.StatementSetQuery{Report: report}
	if values.Has("templates") {
		codes, err := gl.ParseStatementSetTemplates(values.Get("templates"))
		if err != nil {
			return q, err
		}
		q.Templates = codes
	}
	notes, err := gl.ParseStatementSetNotes(values.Get("notes"))
	if err != nil {
		return q, err
	}
	q.Notes = notes
	return q, nil
}

// localizeStatementSet ใส่เลขลำดับข้อมูลที่อ่าน (sequence) และแปลงเหตุที่คำนวณงบไม่สำเร็จเป็นข้อความตามภาษาที่ผู้ใช้เลือก
// ด้วยสัญญาข้อผิดพลาดเดียวกับทั้งคำขอ (errorPayloadFor) — ปัญหาฝั่งระบบบันทึก log ไว้ ไม่ส่งข้อความเทคนิคให้ผู้ใช้
func localizeStatementSet(set *gl.StatementSet, version int64, lang string) {
	set.Sequence = version
	for i := range set.Sections {
		section := &set.Sections[i]
		if section.Err == nil {
			if section.Report != nil {
				section.Report.Sequence = version
			}
			continue
		}
		status, payload := errorPayloadFor(section.Err, lang)
		if status >= 500 {
			log.Printf("[gl/v2] statement set section %s failed: %v", section.Code, section.Err)
		}
		section.Error = payload.Message
	}
}

func (h *Http) statementSet(ctx context.Context, request microservice.IContext, scope gl.Scope, report gl.ReportQuery, version int64) error {
	q, err := statementSetQuery(request.Request().URL.Query(), report)
	if err != nil {
		return failure(request, err)
	}
	set, err := h.pg.StatementSet(ctx, scope, q)
	if err != nil {
		return failure(request, err)
	}
	if err = h.endRead(ctx, scope, version); err != nil {
		return failure(request, err)
	}
	localizeStatementSet(&set, version, getRequestLanguage(request))
	return response(request, set)
}
