package generalledger

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Champ 5502 "ผังบัญชีตัวอย่าง" wipes the chart and copies a template; BC's "ใช้ผังบัญชีมาตรฐาน" never deletes:
// accounts action "load-standard" loads the embedded Thai standard chart only into a company that has no account row at all.

//go:embed standard_chart_th.json
var standardChartJSON []byte

var (
	standardChartOnce sync.Once
	standardChartRows []Account
	standardChartErr  error
)

// StandardChart returns the embedded standard chart, validated and ordered parents-first (level, then code).
// It is parsed once; a broken data file shows up as an error here (and in the unit test), never as a half-loaded chart.
func StandardChart() ([]Account, error) {
	standardChartOnce.Do(func() {
		var rows []Account
		if err := json.Unmarshal(standardChartJSON, &rows); err != nil {
			standardChartErr = fmt.Errorf("อ่านไฟล์ผังบัญชีมาตรฐานไม่ได้: %w", err)
			return
		}
		if err := validateStandardChart(rows); err != nil {
			standardChartErr = err
			return
		}
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].Level != rows[j].Level {
				return rows[i].Level < rows[j].Level
			}
			return rows[i].AccountCode < rows[j].AccountCode
		})
		standardChartRows = rows
	})
	return standardChartRows, standardChartErr
}

// validateStandardChart checks the chart as a tree. Codes are only compared for equality and parent links —
// the structure comes from parentaccountcode and level, never from the digits of a code.
func validateStandardChart(chart []Account) error {
	if len(chart) == 0 {
		return fmt.Errorf("ผังบัญชีมาตรฐานว่างเปล่า")
	}
	byCode := make(map[string]*Account, len(chart))
	thaiNames := make(map[string]string, len(chart))
	children := make(map[string]int, len(chart))
	for i := range chart {
		a := &chart[i]
		if err := a.Validate(); err != nil {
			return fmt.Errorf("บัญชี %s: %w", a.AccountCode, err)
		}
		if a.AccountCode != NormalizeCode(a.AccountCode) || a.ParentAccountCode != NormalizeCode(a.ParentAccountCode) {
			return fmt.Errorf("บัญชี %s: รหัสต้องไม่มีช่องว่างหัวท้าย", a.AccountCode)
		}
		if _, dup := byCode[a.AccountCode]; dup {
			return fmt.Errorf("รหัสบัญชี %s ซ้ำในผังมาตรฐาน", a.AccountCode)
		}
		byCode[a.AccountCode] = a
		// Same comparison as checkDuplicateAccountName: Thai name, trimmed, NFC, case-insensitive.
		key := strings.ToLower(NormalizeCode(a.ThaiName()))
		if other, dup := thaiNames[key]; dup {
			return fmt.Errorf("ชื่อบัญชี “%s” ซ้ำกันระหว่างรหัส %s และ %s", a.ThaiName(), other, a.AccountCode)
		}
		thaiNames[key] = a.AccountCode
	}
	for i := range chart {
		a := &chart[i]
		if a.ParentAccountCode == "" {
			if a.Level != 1 {
				return fmt.Errorf("บัญชี %s: บัญชีที่ไม่มีบัญชีแม่ต้องเป็นระดับ 1", a.AccountCode)
			}
			continue
		}
		parent, ok := byCode[a.ParentAccountCode]
		if !ok {
			return fmt.Errorf("บัญชี %s: ไม่พบบัญชีแม่ %s", a.AccountCode, a.ParentAccountCode)
		}
		if parent.Level != a.Level-1 {
			return fmt.Errorf("บัญชี %s: ระดับ %d แต่บัญชีแม่ %s เป็นระดับ %d", a.AccountCode, a.Level, parent.AccountCode, parent.Level)
		}
		if parent.AllowPosting {
			return fmt.Errorf("บัญชี %s: บัญชีแม่ %s ต้องเป็นบัญชีหัวข้อ (ลงรายการไม่ได้)", a.AccountCode, parent.AccountCode)
		}
		if a.IsActive && !parent.IsActive {
			return fmt.Errorf("บัญชี %s: บัญชีแม่ %s ต้องเปิดใช้งาน", a.AccountCode, parent.AccountCode)
		}
		children[parent.AccountCode]++
	}
	for i := range chart {
		a := &chart[i]
		switch hasChildren := children[a.AccountCode] > 0; {
		case hasChildren && a.AllowPosting:
			return fmt.Errorf("บัญชี %s: มีบัญชีย่อยแล้วต้องเป็นบัญชีหัวข้อ (ลงรายการไม่ได้)", a.AccountCode)
		case !hasChildren && !a.AllowPosting:
			return fmt.Errorf("บัญชี %s: บัญชีที่ไม่มีบัญชีย่อยต้องลงรายการได้", a.AccountCode)
		}
		// Every parent link and level was checked above, so this walk ends at the level-1 root.
		root := a
		for root.ParentAccountCode != "" {
			root = byCode[root.ParentAccountCode]
		}
		if a.AccountType != root.AccountType {
			return fmt.Errorf("บัญชี %s: หมวด %s ต้องตรงกับบัญชีระดับ 1 %s (%s)", a.AccountCode, a.AccountType, root.AccountCode, root.AccountType)
		}
	}
	return nil
}

// standardChartChanges turns the chart into one new account record per row (fresh identity each), parents first.
func standardChartChanges(scope Scope, now time.Time, chart []Account) ([]Change, error) {
	changes := make([]Change, 0, len(chart))
	for _, row := range chart {
		acc := row
		acc.Identity = newIdentity(scope, now)
		data, err := json.Marshal(acc)
		if err != nil {
			return nil, err
		}
		changes = append(changes, Change{Kind: "accounts", ID: acc.ID, Code: acc.AccountCode, Payload: string(data)})
	}
	return changes, nil
}

// loadStandardChart answers accounts action "load-standard" inside the command's transaction. Soft-deleted
// accounts count as existing: their codes would still collide, and a company that once had a chart is not "new".
func (s *PostgresStore) loadStandardChart(ctx context.Context, tx *sql.Tx, scope Scope, now time.Time) ([]Change, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='accounts')`, scope.Company).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, userError("standard_chart_not_empty", "ใช้ผังบัญชีมาตรฐานได้เฉพาะบริษัทที่ยังไม่มีผังบัญชี บริษัทนี้มีบัญชีอยู่แล้ว กรุณาเพิ่มหรือแก้ไขบัญชีในผังบัญชีเดิม")
	}
	chart, err := StandardChart()
	if err != nil {
		return nil, err
	}
	return standardChartChanges(scope, now, chart)
}
