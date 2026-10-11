package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"smlcloudplatform/internal/cheque"
	"smlcloudplatform/internal/doctype"
)

// รูปแบบลงบัญชี (mappings) ของเอกสาร ERP: หนึ่งประเภทเอกสารมีรูปแบบเดียว ระบุสมุดรายวันและรายการเดบิต/เครดิต
// โดยแต่ละรายการอ้างยอดจากเอกสารด้วยคำสั่งเดียวในสามค่านี้ — ระบบไม่เดาบัญชีเองจากรหัสบัญชี
const (
	MappingSourceTotal = "total" // ยอดรวมเอกสาร (รวมภาษี)
	MappingSourceGoods = "goods" // มูลค่าสินค้า/บริการก่อนภาษี
	MappingSourceVat   = "vat"   // ภาษีมูลค่าเพิ่ม
)

// แหล่งยอดของเอกสารเปลี่ยนสถานะเช็ค (cheque.MappingSourceDocs) ใช้แทนสามค่าข้างบน: ยอดเช็ครวม = ยอดสุทธิ + ค่าธรรมเนียม
const (
	MappingSourceChequeTotal = "cheque_total" // ยอดเช็ครวมของเอกสาร
	MappingSourceChequeNet   = "cheque_net"   // ยอดเช็คหลังหักค่าธรรมเนียม
	MappingSourceChequeFee   = "cheque_fee"   // ค่าธรรมเนียม/ค่าใช้จ่ายของเช็ค
)

// MappingAccountFromBankBook: บัญชีของรายการมาจากสมุดเงินฝากในหัวเอกสารเช็ค — Champ ใช้บัญชีของสมุดแทนบัญชีในแม่แบบ
// (gltlsprocess.cpp:7802-7824); accountcode ของรายการเป็นบัญชีสำรองเมื่อเอกสารไม่มีสมุด (ยกเลิกเช็ครับในมือ)
const MappingAccountFromBankBook = "bank_book"

var (
	erpMappingSources    = []string{MappingSourceTotal, MappingSourceGoods, MappingSourceVat}
	chequeMappingSources = []string{MappingSourceChequeTotal, MappingSourceChequeNet, MappingSourceChequeFee}
)

const (
	MappingSideDebit  = "debit"
	MappingSideCredit = "credit"
)

// MappingForSourceDoc returns the active posting pattern of an ERP document type (doctype.PostingKeys) or a cheque
// status document (cheque.MappingSourceDocs). found is
// false when the company has not set one up (or switched it off) — the document then stays unposted rather than
// posting to accounts the system had to guess.
func MappingForSourceDoc(ctx context.Context, tx *sql.Tx, scope Scope, sourceDoc string) (Master, bool, error) {
	var payload []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='mappings' AND payload->>'sourcedoc'=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY code LIMIT 1`, scope.Company, sourceDoc).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return Master{}, false, nil
	}
	if err != nil {
		return Master{}, false, err
	}
	var m Master
	if err = json.Unmarshal(payload, &m); err != nil {
		return Master{}, false, err
	}
	return m, m.IsActive, nil
}

// validateMasterMapping ตรวจรูปแบบลงบัญชีก่อนบันทึก; ชนิดอื่นไม่เก็บ sourcedoc
// selfID คือ id ของรูปแบบที่กำลังแก้ ("" = สร้างใหม่) เพื่อไม่ให้ชนกับตัวเองตอนตรวจว่าซ้ำ
func (s *PostgresStore) validateMasterMapping(ctx context.Context, tx *sql.Tx, scope Scope, m *Master, selfID string) error {
	if m.Kind != "mappings" {
		m.SourceDoc = ""
		return nil
	}
	m.SourceDoc = strings.TrimSpace(m.SourceDoc)
	chequeDoc := contains(cheque.MappingSourceDocs(), m.SourceDoc)
	if !chequeDoc && !contains(doctype.PostingKeys(), m.SourceDoc) {
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
		r.AccountFrom = strings.ToLower(strings.TrimSpace(r.AccountFrom))
		if r.Side != MappingSideDebit && r.Side != MappingSideCredit {
			return mappingRuleError("mapping_rule_side_invalid", i, "รายการที่ %d ต้องเลือกด้านเดบิตหรือเครดิต")
		}
		// เอกสารแต่ละตระกูลใช้แหล่งยอดของตัวเองเท่านั้น (ERP: total/goods/vat, เช็ค: cheque_*)
		if chequeDoc && !contains(chequeMappingSources, r.Source) {
			return mappingRuleError("mapping_rule_cheque_source_invalid", i, "รายการที่ %d ต้องเลือกยอดจากเอกสารเช็ค (ยอดเช็ค ยอดสุทธิ หรือค่าธรรมเนียม)")
		}
		if !chequeDoc && !contains(erpMappingSources, r.Source) {
			return mappingRuleError("mapping_rule_source_invalid", i, "รายการที่ %d ต้องเลือกยอดจากเอกสาร (ยอดรวม มูลค่าสินค้า หรือภาษี)")
		}
		switch {
		case r.AccountFrom == "":
		case r.AccountFrom == MappingAccountFromBankBook && chequeDoc:
			if r.AccountCode == "" {
				continue // ไม่มีบัญชีสำรอง: ใช้บัญชีของสมุดเท่านั้น
			}
		default:
			return mappingRuleError("mapping_rule_account_from_invalid", i, "รายการที่ %d เลือกบัญชีจากสมุดเงินฝากได้เฉพาะรูปแบบลงบัญชีของเอกสารเช็ค")
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
// ยอดรวมเท่ากับมูลค่าสินค้า + ภาษี (เช็ค: ยอดเช็ค = ยอดสุทธิ + ค่าธรรมเนียม) จึงเปรียบเทียบสัมประสิทธิ์ของสองส่วนนี้ของแต่ละด้าน
func mappingBalances(rules []MappingRule) bool {
	var debit, credit [2]int // [มูลค่าสินค้า/ยอดสุทธิ, ภาษี/ค่าธรรมเนียม]
	for _, r := range rules {
		side := &credit
		if r.Side == MappingSideDebit {
			side = &debit
		}
		switch r.Source {
		case MappingSourceTotal, MappingSourceChequeTotal:
			side[0]++
			side[1]++
		case MappingSourceGoods, MappingSourceChequeNet:
			side[0]++
		case MappingSourceVat, MappingSourceChequeFee:
			side[1]++
		}
	}
	return debit == credit && debit != [2]int{}
}
