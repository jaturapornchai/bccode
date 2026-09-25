package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/mcptoken"
	branchmodels "smlcloudplatform/internal/organization/branch/models"
	"smlcloudplatform/internal/taxaddress"

	"github.com/shopspring/decimal"
)

// ยอดเงินในรายงานภาษีคำนวณด้วย decimal และส่งออกเป็น string ทศนิยม 2 ตำแหน่ง (ห้ามใช้ float/JSON number กับเงิน)

// moneyText - decimal → "1234.50" (ปัดครึ่งขึ้นที่ 2 ตำแหน่ง)
func moneyText(amount decimal.Decimal) string {
	return amount.StringFixed(2)
}

// CompanyHeader - ข้อมูลผู้ประกอบการจริงจากทะเบียนบริษัท สำหรับหัวแบบ ภ.พ.30 / ภ.ง.ด. / 50 ทวิ
// ที่อยู่สำหรับภาษีเป็นของสำนักงานใหญ่ (ทะเบียนยังไม่มีเลขสาขา/ที่อยู่สาขา จึงไม่ส่งค่าเดา — หัวแบบใช้ 00000 ที่แก้ได้);
// AddressLine = ที่อยู่บรรทัดเดียวสำหรับ 50 ทวิ (taxaddress.Address.Line) — ว่างเมื่อทะเบียนยังไม่มีที่อยู่
type CompanyHeader struct {
	Code        string              `json:"code"`
	Name        string              `json:"name"`
	TaxID       string              `json:"taxid"`
	Address     *taxaddress.Address `json:"address,omitempty"`
	Phone       string              `json:"phone,omitempty"`
	AddressLine string              `json:"addressline,omitempty"`
}

// registryAddress - ที่อยู่สำหรับภาษีจากทะเบียน (ว่างเมื่อไม่มี)
func (h CompanyHeader) registryAddress() taxaddress.Address {
	if h.Address == nil {
		return taxaddress.Address{}
	}
	return *h.Address
}

// loadCompanyHeader - อ่านชื่อ/เลขผู้เสียภาษี/ที่อยู่สำหรับภาษี/โทรศัพท์ของบริษัทที่ผู้ใช้เลือกจากฐานข้อมูลควบคุมกลาง
func loadCompanyHeader(ctx context.Context, holdingCode, businessCode string) (CompanyHeader, error) {
	db, err := mypg.PgSqlFastConnect(mcptoken.ControlDatabase)
	if err != nil {
		return CompanyHeader{Code: businessCode}, fmt.Errorf("connect control database: %w", err)
	}
	return queryCompanyHeader(ctx, db, holdingCode, businessCode)
}

// queryCompanyHeader - เนื้อของ loadCompanyHeader บนฐานควบคุมกลางที่ส่งมา (เทส integration เรียกตรง); ไม่พบบริษัท = หัวว่าง
func queryCompanyHeader(ctx context.Context, db *sql.DB, holdingCode, businessCode string) (CompanyHeader, error) {
	header := CompanyHeader{Code: businessCode}
	var address taxaddress.Address
	dest := []interface{}{&header.Name, &header.TaxID}
	for _, field := range address.Fields() {
		dest = append(dest, field)
	}
	err := db.QueryRowContext(ctx, `SELECT name, COALESCE(tax_id, ''), `+strings.Join(taxaddress.Keys, ", ")+`, phone
FROM companies WHERE holding_code = $1 AND code = $2`, holdingCode, businessCode).Scan(append(dest, &header.Phone)...)
	if err == sql.ErrNoRows {
		return header, nil
	}
	if err != nil {
		return header, fmt.Errorf("load company: %w", err)
	}
	address.Normalize()
	header.Phone = taxaddress.NormalizePhone(header.Phone)
	if !address.IsBlank() {
		header.Address = &address
		header.AddressLine = address.Line()
	}
	return header, nil
}

// TaxEstablishment - สถานประกอบการของบริษัท (สาขาในทะเบียน, เลขสาขา 5 หลักตาม ภ.พ.20) ให้เลือกทำรายงานภาษีซื้อ/ขาย
// เป็นรายสถานประกอบการ (ม.87 วรรคสาม, ประกาศฯ VAT ฉบับที่ 89 ข้อ 5) และพิมพ์หัวรายงาน สำนักงานใหญ่/สาขาที่
type TaxEstablishment struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	IsHeadOffice bool   `json:"isheadoffice"`
}

// loadTaxEstablishments - สาขาทั้งหมดของบริษัทจากฐานข้อมูลควบคุมกลาง เรียงตามเลขสาขา (รวมสาขาที่ปิดแล้ว เพราะงวดเก่ายังมีรายการ)
func loadTaxEstablishments(ctx context.Context, holdingCode, businessCode string) ([]TaxEstablishment, error) {
	db, err := mypg.PgSqlFastConnect(mcptoken.ControlDatabase)
	if err != nil {
		return nil, fmt.Errorf("connect control database: %w", err)
	}
	return queryTaxEstablishments(ctx, db, holdingCode, businessCode)
}

// queryTaxEstablishments - เนื้อของ loadTaxEstablishments บนฐานควบคุมกลางที่ส่งมา; สำนักงานใหญ่ = เลขสาขา 00000 ตามกรมสรรพากร
func queryTaxEstablishments(ctx context.Context, db *sql.DB, holdingCode, businessCode string) ([]TaxEstablishment, error) {
	rows, err := db.QueryContext(ctx, `SELECT code, name FROM branches WHERE holding_code = $1 AND company_code = $2 ORDER BY code`, holdingCode, businessCode)
	if err != nil {
		return nil, fmt.Errorf("load branches: %w", err)
	}
	defer rows.Close()
	out := []TaxEstablishment{}
	for rows.Next() {
		var e TaxEstablishment
		if err := rows.Scan(&e.Code, &e.Name); err != nil {
			return nil, fmt.Errorf("scan branch: %w", err)
		}
		e.IsHeadOffice = branchmodels.IsThaiHeadOfficeBranchCode(e.Code)
		out = append(out, e)
	}
	return out, rows.Err()
}
