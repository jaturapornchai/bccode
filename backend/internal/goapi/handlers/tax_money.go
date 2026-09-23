package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/mcptoken"

	"github.com/shopspring/decimal"
)

// ยอดเงินในรายงานภาษีคำนวณด้วย decimal และส่งออกเป็น string ทศนิยม 2 ตำแหน่ง (ห้ามใช้ float/JSON number กับเงิน)

// moneyText - decimal → "1234.50" (ปัดครึ่งขึ้นที่ 2 ตำแหน่ง)
func moneyText(amount decimal.Decimal) string {
	return amount.StringFixed(2)
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
