package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/mcptoken"

	"github.com/shopspring/decimal"
)

// ยอดเงินในรายงานภาษีคำนวณด้วย decimal และส่งออกเป็น string ทศนิยม 2 ตำแหน่ง (ห้ามใช้ float/JSON number กับเงิน)

// moneySQL - แปลงคอลัมน์ยอดเงิน (ตาราง ERP บางตารางยังเป็น double precision) เป็น numeric ปัด 2 ตำแหน่งใน SQL
// ชื่อคอลัมน์ต้องเป็นค่าคงที่ในโค้ดเท่านั้น ห้ามรับจาก input
func moneySQL(column string) string {
	return fmt.Sprintf("ROUND(COALESCE(%s, 0)::numeric, 2)", column)
}

// moneyText - decimal → "1234.50" (ปัดครึ่งขึ้นที่ 2 ตำแหน่ง)
func moneyText(amount decimal.Decimal) string {
	return amount.StringFixed(2)
}

var moneyInputPattern = regexp.MustCompile(`^[0-9]{1,13}(\.[0-9]{1,2})?$`)

// parseMoneyInput - ยอดเงินที่ผู้ใช้กรอก: ว่าง = 0, ต้องไม่ติดลบและไม่เกิน 2 ตำแหน่ง
func parseMoneyInput(raw string) (decimal.Decimal, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return decimal.Zero, true
	}
	if !moneyInputPattern.MatchString(raw) {
		return decimal.Zero, false
	}
	amount, err := decimal.NewFromString(raw)
	return amount, err == nil
}

// CompanyHeader - ข้อมูลผู้ประกอบการจริงจากทะเบียนบริษัท สำหรับหัวแบบ ภ.พ.30 / ภ.ง.ด. / 50 ทวิ
// ทะเบียนบริษัทยังไม่มีที่อยู่และเลขสาขา จึงไม่ส่งค่าเดา — หน้าจอต้องแสดงว่ายังไม่ระบุ
type CompanyHeader struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	TaxID string `json:"taxid"`
}

// loadCompanyHeader - อ่านชื่อ/เลขผู้เสียภาษีของบริษัทที่ผู้ใช้เลือกจากฐานข้อมูลควบคุมกลาง
func loadCompanyHeader(ctx context.Context, holdingCode, businessCode string) (CompanyHeader, error) {
	header := CompanyHeader{Code: businessCode}
	db, err := mypg.PgSqlFastConnect(mcptoken.ControlDatabase)
	if err != nil {
		return header, fmt.Errorf("connect control database: %w", err)
	}
	err = db.QueryRowContext(ctx, `SELECT name, COALESCE(tax_id, '') FROM companies WHERE holding_code = $1 AND code = $2`,
		holdingCode, businessCode).Scan(&header.Name, &header.TaxID)
	if err == sql.ErrNoRows {
		return header, nil
	}
	if err != nil {
		return header, fmt.Errorf("load company: %w", err)
	}
	return header, nil
}

var (
	thaiDigits = []string{"ศูนย์", "หนึ่ง", "สอง", "สาม", "สี่", "ห้า", "หก", "เจ็ด", "แปด", "เก้า"}
	thaiUnits  = []string{"", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน"}
)

// thaiBahtText - จำนวนเงินเป็นตัวอักษรไทย เช่น 1234.50 → "หนึ่งพันสองร้อยสามสิบสี่บาทห้าสิบสตางค์"
// กติกาเดียวกับ frontend/src/lib/thai-baht-text.ts (เอ็ด / ยี่สิบ / สิบ / ล้านซ้อน)
func thaiBahtText(amount decimal.Decimal) string {
	fixed := amount.Abs().StringFixed(2)
	intPart, satang, _ := strings.Cut(fixed, ".")
	if intPart == "0" && satang == "00" {
		return "ศูนย์บาทถ้วน"
	}

	text := ""
	if intPart != "0" {
		var groups []string
		for rest := intPart; rest != ""; {
			take := min(6, len(rest))
			groups = append([]string{rest[len(rest)-take:]}, groups...)
			rest = rest[:len(rest)-take]
		}
		multiDigit := len(intPart) > 1 || intPart > "1"
		for i, group := range groups {
			groupText := thaiDigitGroup(group, i == len(groups)-1 && multiDigit)
			if groupText != "" {
				text += groupText + strings.Repeat("ล้าน", len(groups)-i-1)
			}
		}
		text += "บาท"
	}

	if satang == "00" {
		text += "ถ้วน"
	} else {
		text += thaiDigitGroup(satang, satang > "01") + "สตางค์"
	}
	if amount.Sign() < 0 {
		return "ลบ" + text
	}
	return text
}

// thaiDigitGroup - อ่านตัวเลขไม่เกิน 6 หลัก (หนึ่งรอบล้าน)
func thaiDigitGroup(digits string, unitOneIsEt bool) string {
	result := ""
	for i, ch := range digits {
		digit := int(ch - '0')
		pos := len(digits) - i - 1
		switch {
		case digit == 0:
		case pos == 0 && digit == 1 && unitOneIsEt:
			result += "เอ็ด"
		case pos == 1 && digit == 2:
			result += "ยี่สิบ"
		case pos == 1 && digit == 1:
			result += "สิบ"
		default:
			result += thaiDigits[digit] + thaiUnits[pos]
		}
	}
	return result
}
