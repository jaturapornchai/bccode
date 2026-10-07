package generalledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// กลุ่มผังบัญชี (Champ 5497 BCAccGroup): แถวงบชนิด account อ้างกลุ่มแทนการเลือกบัญชีทีละตัว (Champ ACCGRP → ACC(รายการของกลุ่ม))
// สมาชิกเก็บที่ตัวกลุ่ม เป็นรายการ "รหัสเดียว" หรือ "ช่วง from–to" (Champ "1:2,4,5"); บัญชีหนึ่งอยู่ได้หลายกลุ่ม
// ช่วงเทียบรหัสเป็นตัวอักษรตามลำดับ byte (COLLATE "C") เหมือน Champ ProcessWord BETWEEN — ผังบัญชีเป็นผู้ตัดสินว่ามีบัญชีใดในช่วง
// ตอนคำนวณงบ resolveStatementGroups แปลงกลุ่มเป็นรหัสบัญชีจริงที่จุดเดียว แล้วตัวตรวจ/ตัวคำนวณอื่นอ่าน AccountCodes ตามเดิม

// AccountEntry หนึ่งรายการของกลุ่ม: To ว่าง = รหัสเดียว (From), มี To = ทุกบัญชีที่รหัสอยู่ระหว่าง From ถึง To รวมปลายทั้งสอง
type AccountEntry struct {
	From string `json:"from" bson:"from"`
	To   string `json:"to,omitempty" bson:"to,omitempty"`
}

// normalizeAccountEntries ตัด/NFC รหัส ตรวจรูปแบบรหัส ช่วงต้องไม่กลับด้าน (ปลายช่วงเท่าต้นช่วง = รหัสเดียว) และตัดรายการซ้ำ
func normalizeAccountEntries(entries []AccountEntry) ([]AccountEntry, error) {
	out := make([]AccountEntry, 0, len(entries))
	seen := map[AccountEntry]bool{}
	for i, entry := range entries {
		from, to := NormalizeCode(entry.From), NormalizeCode(entry.To)
		if err := checkCode(from, fmt.Sprintf("accountentries[%d].from", i), "รหัสบัญชี", codeMaxRunes); err != nil {
			return nil, err
		}
		if to != "" {
			if err := checkCode(to, fmt.Sprintf("accountentries[%d].to", i), "รหัสบัญชีปลายช่วง", codeMaxRunes); err != nil {
				return nil, err
			}
		}
		if to == from {
			to = ""
		}
		// เทียบ byte ของ UTF-8 ที่ถูกต้อง = ลำดับ COLLATE "C" ในฐานข้อมูล
		if to != "" && from > to {
			e := fieldError("account_group_entry_reversed", fmt.Sprintf("accountentries[%d].to", i), fmt.Sprintf("ช่วงรหัสบัญชี %s ถึง %s: รหัสเริ่มต้นต้องไม่มากกว่ารหัสสิ้นสุด", from, to))
			e.Args = []string{from, to}
			return nil, e
		}
		item := AccountEntry{From: from, To: to}
		if seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

// validateMasterAccountGroup: รายการรหัสบัญชีมีเฉพาะข้อมูลหลักชนิด account-groups (ชนิดอื่นไม่เก็บ); รหัสเดียวต้องเป็นบัญชีที่ลงรายการได้
// และยังไม่ถูกลบ (ปลายช่วงไม่ตรวจ — ช่วงใช้ผังบัญชีปัจจุบันตอนคำนวณงบ)
func (s *PostgresStore) validateMasterAccountGroup(ctx context.Context, tx *sql.Tx, scope Scope, m *Master) error {
	if m.Kind != "account-groups" {
		m.AccountEntries = nil
		return nil
	}
	entries, err := normalizeAccountEntries(m.AccountEntries)
	if err != nil {
		return err
	}
	m.AccountEntries = entries
	singles := []Line{}
	for _, entry := range entries {
		if entry.To == "" {
			singles = append(singles, Line{AccountCode: entry.From})
		}
	}
	if len(singles) == 0 {
		return nil
	}
	accounts, err := loadLineAccounts(ctx, tx, scope.Company, singles)
	if err != nil {
		return err
	}
	bad := []string{}
	for _, line := range singles {
		if account, ok := accounts[line.AccountCode]; !ok || !account.AllowPosting {
			bad = append(bad, line.AccountCode)
		}
	}
	if len(bad) == 0 {
		return nil
	}
	joined := strings.Join(bad, ", ")
	e := fieldError("account_group_accounts_invalid", "accountentries", fmt.Sprintf("บัญชี %s ไม่มีในผังบัญชี ถูกลบแล้ว หรือเป็นบัญชีหัวข้อที่บันทึกบัญชีไม่ได้ — เลือกบัญชีที่ลงรายการได้ หรือใช้ช่วงรหัสแทน", joined))
	e.Args = []string{joined}
	return e
}

// resolveStatementGroups: แถวชนิด account ที่อ้างกลุ่ม ได้ AccountCodes = บัญชีของกลุ่ม ณ ตอนนี้ (แทนรายการเดิมของแถว) — จุดเดียวที่แปลงกลุ่ม
// ผลคือสำเนาของแม่แบบ (ไม่แก้ที่เก็บ); คืนข้อความเตือนสำหรับกลุ่มที่หายหรือไม่มีบัญชีเลย เพราะแถวนั้นจะได้ยอดศูนย์โดยไม่มีใครรู้
func resolveStatementGroups(ctx context.Context, tx *sql.Tx, company string, template Master) (Master, []string, error) {
	wanted := []string{}
	seen := map[string]bool{}
	for _, row := range template.Rows {
		if row.RowType == "account" && row.AccountGroup != "" && !seen[row.AccountGroup] {
			seen[row.AccountGroup] = true
			wanted = append(wanted, row.AccountGroup)
		}
	}
	if len(wanted) == 0 {
		return template, nil, nil
	}
	// ดึงกลุ่มและบัญชีสมาชิกทุกกลุ่มในคำสั่งเดียว (LEFT JOIN ให้กลุ่มที่ไม่มีสมาชิกยังถูกนับว่ามีอยู่)
	found, err := tx.QueryContext(ctx, `SELECT g.code, m.code FROM gl_records g
    LEFT JOIN LATERAL (
      SELECT DISTINCT a.code FROM jsonb_to_recordset(COALESCE(g.payload->'accountentries','[]'::jsonb)) AS e("from" text, "to" text)
      JOIN gl_records a ON a.company=g.company AND a.kind='accounts' AND NOT COALESCE((a.payload->>'isdeleted')::boolean,false) AND COALESCE((a.payload->>'allowposting')::boolean,false)
        AND (CASE WHEN COALESCE(e."to",'')='' THEN a.code=e."from" ELSE a.code COLLATE "C" BETWEEN e."from" AND e."to" END)
    ) m ON true
    WHERE g.company=$1 AND g.kind='account-groups' AND g.code=ANY($2) AND NOT COALESCE((g.payload->>'isdeleted')::boolean,false)
    ORDER BY g.code, m.code COLLATE "C"`, company, pq.Array(wanted))
	if err != nil {
		return template, nil, err
	}
	defer found.Close()
	exists := map[string]bool{}
	members := map[string][]string{}
	for found.Next() {
		var code string
		var member sql.NullString
		if err = found.Scan(&code, &member); err != nil {
			return template, nil, err
		}
		exists[code] = true
		if member.Valid {
			members[code] = append(members[code], member.String)
		}
	}
	if err = found.Err(); err != nil {
		return template, nil, err
	}
	rows := append([]StatementRow(nil), template.Rows...)
	warnings := []string{}
	for i := range rows {
		row := &rows[i]
		if row.RowType != "account" || row.AccountGroup == "" {
			continue
		}
		if !exists[row.AccountGroup] {
			warnings = append(warnings, fmt.Sprintf("บรรทัด %d “%s”: ไม่พบกลุ่มผังบัญชี “%s” (ถูกลบแล้ว) จึงไม่มียอด — เลือกกลุ่มใหม่ในรูปแบบงบ", row.RowNo, row.Title, row.AccountGroup))
		} else if len(members[row.AccountGroup]) == 0 {
			warnings = append(warnings, fmt.Sprintf("บรรทัด %d “%s”: กลุ่มผังบัญชี “%s” ไม่มีบัญชีที่ลงรายการได้ตามรายการที่ระบุ จึงไม่มียอด", row.RowNo, row.Title, row.AccountGroup))
		}
		row.AccountCodes = members[row.AccountGroup]
	}
	template.Rows = rows
	return template, warnings, nil
}

// validateStatementTemplateGroups: ตอนบันทึกแม่แบบ แถว account ที่อ้างกลุ่มต้องมีกลุ่มนั้นอยู่จริง และไม่เลือกบัญชีเองซ้อนกับกลุ่ม
// (แถวชนิดอื่นไม่ถูกอ่านตอนคำนวณ จึงไม่ตรวจ เหมือนรหัสบัญชีที่ค้างในบรรทัดชนิดอื่น)
func validateStatementTemplateGroups(ctx context.Context, tx *sql.Tx, company string, m Master) error {
	wanted := []string{}
	seen := map[string]bool{}
	for _, row := range m.Rows {
		if row.RowType == "account" && row.AccountGroup != "" && !seen[row.AccountGroup] {
			seen[row.AccountGroup] = true
			wanted = append(wanted, row.AccountGroup)
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='account-groups' AND code=ANY($2) AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, pq.Array(wanted))
	if err != nil {
		return err
	}
	defer rows.Close()
	exists := map[string]bool{}
	for rows.Next() {
		var code string
		if err = rows.Scan(&code); err != nil {
			return err
		}
		exists[code] = true
	}
	if err = rows.Err(); err != nil {
		return err
	}
	for i, row := range m.Rows {
		if row.RowType != "account" || row.AccountGroup == "" {
			continue
		}
		field := fmt.Sprintf("rows[%d].accountgroup", i)
		if len(row.AccountCodes) > 0 {
			e := fieldError("statement_template_row_group_and_accounts", field, fmt.Sprintf("บรรทัด %d “%s”: เลือกทั้งกลุ่มผังบัญชีและบัญชีเองไม่ได้ — เลือกอย่างใดอย่างหนึ่ง", row.RowNo, row.Title))
			e.Args = []string{fmt.Sprint(row.RowNo), row.Title}
			return e
		}
		if !exists[row.AccountGroup] {
			e := fieldError("statement_template_row_group_missing", field, fmt.Sprintf("บรรทัด %d “%s”: ไม่พบกลุ่มผังบัญชี “%s” — เลือกกลุ่มที่มีอยู่ในบันทึกกลุ่มผังบัญชี", row.RowNo, row.Title, row.AccountGroup))
			e.Args = []string{fmt.Sprint(row.RowNo), row.Title, row.AccountGroup}
			return e
		}
	}
	return nil
}

// accountGroupInUse: มีรูปแบบงบการเงินที่ยังใช้อยู่อ้างกลุ่มนี้ในแถว account หรือไม่ (ลบ/เปลี่ยนรหัสกลุ่มแล้วแถวนั้นจะไม่มียอด)
func accountGroupInUse(ctx context.Context, tx *sql.Tx, company, code string) (bool, error) {
	var used bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records t WHERE t.company=$1 AND t.kind='statement-templates' AND NOT COALESCE((t.payload->>'isdeleted')::boolean,false)
      AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(t.payload->'rows','[]'::jsonb)) r WHERE r->>'rowtype'='account' AND r->>'accountgroup'=$2))`, company, code).Scan(&used)
	return used, err
}

// guardAccountGroupDelete / guardAccountGroupRecode: กลุ่มที่รูปแบบงบอ้างอยู่ ลบหรือเปลี่ยนรหัสไม่ได้ (ชื่อและรายการบัญชีแก้ได้เสมอ)
func guardAccountGroupDelete(ctx context.Context, tx *sql.Tx, company string, old Master) error {
	used, err := accountGroupInUse(ctx, tx, company, old.Code)
	if err != nil || !used {
		return err
	}
	e := conflictError("account_group_in_use_delete", "code", fmt.Sprintf("กลุ่มผังบัญชี “%s” ถูกใช้ในรูปแบบงบการเงินอยู่ ลบไม่ได้ ให้แก้รูปแบบงบไม่ให้อ้างกลุ่มนี้ก่อน", old.Code))
	e.Args = []string{old.Code}
	return e
}

func guardAccountGroupRecode(ctx context.Context, tx *sql.Tx, company string, old, next Master) error {
	if next.Code == old.Code {
		return nil
	}
	used, err := accountGroupInUse(ctx, tx, company, old.Code)
	if err != nil || !used {
		return err
	}
	e := conflictError("account_group_in_use_code", "code", fmt.Sprintf("กลุ่มผังบัญชี “%s” ถูกใช้ในรูปแบบงบการเงินอยู่ เปลี่ยนรหัสไม่ได้ ถ้าต้องการรหัสใหม่ ให้สร้างกลุ่มใหม่แล้วแก้รูปแบบงบให้อ้างกลุ่มใหม่", old.Code))
	e.Args = []string{old.Code}
	return e
}
