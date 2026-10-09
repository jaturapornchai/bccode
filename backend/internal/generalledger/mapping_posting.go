package generalledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"smlcloudplatform/internal/doctype"
)

// รูปแบบลงบัญชี (mappings) ของเอกสาร ERP: หนึ่งประเภทเอกสารมีรูปแบบเดียว ระบุสมุดรายวันและรายการเดบิต/เครดิต
// โดยแต่ละรายการอ้างยอดจากเอกสารด้วยคำสั่งเดียวในสามค่านี้ — ระบบไม่เดาบัญชีเองจากรหัสบัญชี
const (
	MappingSourceTotal = "total" // ยอดรวมเอกสาร (รวมภาษี)
	MappingSourceGoods = "goods" // มูลค่าสินค้า/บริการก่อนภาษี
	MappingSourceVat   = "vat"   // ภาษีมูลค่าเพิ่ม
)

const (
	MappingSideDebit  = "debit"
	MappingSideCredit = "credit"
)

// validateMasterMapping ตรวจรูปแบบลงบัญชีก่อนบันทึก; ชนิดอื่นไม่เก็บ sourcedoc
// selfID คือ id ของรูปแบบที่กำลังแก้ ("" = สร้างใหม่) เพื่อไม่ให้ชนกับตัวเองตอนตรวจว่าซ้ำ
func (s *PostgresStore) validateMasterMapping(ctx context.Context, tx *sql.Tx, scope Scope, m *Master, selfID string) error {
	if m.Kind != "mappings" {
		m.SourceDoc = ""
		return nil
	}
	m.SourceDoc = strings.TrimSpace(m.SourceDoc)
	if !contains(doctype.PostingKeys(), m.SourceDoc) {
		return fieldError("mapping_source_doc_invalid", "sourcedoc", "กรุณาเลือกประเภทเอกสารที่ต้องการลงบัญชีจากรายการ")
	}
	var existing string
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='mappings' AND payload->>'sourcedoc'=$2 AND id<>$3 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY code LIMIT 1`, scope.Company, m.SourceDoc, selfID).Scan(&existing)
	if err == nil {
		e := conflictError("mapping_source_doc_duplicate", "sourcedoc", fmt.Sprintf("ประเภทเอกสารนี้มีรูปแบบลงบัญชีอยู่แล้ว (%s) กรุณาแก้ไขรูปแบบเดิมแทนการสร้างใหม่", existing))
		e.Args = []string{existing}
		return e
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	book, found, err := loadJournalBook(ctx, tx, scope.Company, strings.TrimSpace(m.BookCode))
	if err != nil {
		return err
	}
	if !found || !book.IsActive {
		return fieldError("mapping_book_invalid", "bookcode", "ไม่พบสมุดรายวันที่เลือก หรือสมุดรายวันถูกปิดใช้งาน กรุณาเลือกสมุดรายวันจากรายการ")
	}
	m.BookCode = book.Code
	if len(m.Rules) < 2 {
		return fieldError("mapping_rules_required", "rules", "รูปแบบลงบัญชีต้องมีอย่างน้อย 1 รายการเดบิตและ 1 รายการเครดิต")
	}
	lines := make([]Line, 0, len(m.Rules))
	for i := range m.Rules {
		r := &m.Rules[i]
		r.AccountCode = NormalizeCode(r.AccountCode)
		r.Side = strings.ToLower(strings.TrimSpace(r.Side))
		r.Source = strings.ToLower(strings.TrimSpace(r.Source))
		if r.Side != MappingSideDebit && r.Side != MappingSideCredit {
			return mappingRuleError("mapping_rule_side_invalid", i, "รายการที่ %d ต้องเลือกด้านเดบิตหรือเครดิต")
		}
		if !contains([]string{MappingSourceTotal, MappingSourceGoods, MappingSourceVat}, r.Source) {
			return mappingRuleError("mapping_rule_source_invalid", i, "รายการที่ %d ต้องเลือกยอดจากเอกสาร (ยอดรวม มูลค่าสินค้า หรือภาษี)")
		}
		lines = append(lines, Line{AccountCode: r.AccountCode})
	}
	accounts, err := loadLineAccounts(ctx, tx, scope.Company, lines)
	if err != nil {
		return err
	}
	bad := []string{}
	for _, l := range lines {
		if a, ok := accounts[l.AccountCode]; !ok || !a.AllowPosting {
			bad = append(bad, l.AccountCode)
		}
	}
	if len(bad) > 0 {
		joined := strings.Join(bad, ", ")
		e := fieldError("mapping_account_invalid", "rules", fmt.Sprintf("บัญชี %s ไม่มีในผังบัญชี ถูกลบแล้ว หรือเป็นบัญชีหัวข้อที่บันทึกบัญชีไม่ได้ — เลือกบัญชีที่ลงรายการได้", joined))
		e.Args = []string{joined}
		return e
	}
	if !mappingBalances(m.Rules) {
		return fieldError("mapping_unbalanced", "rules", "ยอดเดบิตไม่เท่ากับยอดเครดิต กรุณาปรับรายการให้ผลรวมสองด้านเท่ากัน เช่น เดบิตยอดรวม = เครดิตมูลค่าสินค้า + เครดิตภาษี")
	}
	return nil
}

func mappingRuleError(code string, index int, format string) *UserError {
	e := fieldError(code, "rules", fmt.Sprintf(format, index+1))
	e.Args = []string{fmt.Sprint(index + 1)}
	return e
}

// mappingBalances ตรวจว่าทุกเอกสารที่ลงตามรูปแบบนี้ได้ยอดเดบิต = เครดิต ไม่ว่ายอดเอกสารเท่าใด
// ยอดรวมเท่ากับมูลค่าสินค้า + ภาษี จึงเปรียบเทียบสัมประสิทธิ์ของสองส่วนนี้ของแต่ละด้าน
func mappingBalances(rules []MappingRule) bool {
	var debit, credit [2]int // [มูลค่าสินค้า, ภาษี]
	for _, r := range rules {
		side := &credit
		if r.Side == MappingSideDebit {
			side = &debit
		}
		switch r.Source {
		case MappingSourceTotal:
			side[0]++
			side[1]++
		case MappingSourceGoods:
			side[0]++
		case MappingSourceVat:
			side[1]++
		}
	}
	return debit == credit && debit != [2]int{}
}
