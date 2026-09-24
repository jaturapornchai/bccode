// Package taxaddress - ที่อยู่สำหรับภาษีของบริษัท (สำนักงานใหญ่) ในทะเบียนบริษัท ใช้เติมหัวแบบยื่นกรมสรรพากร
// (ภ.พ.30 / ภ.ง.ด.) และที่อยู่บนหนังสือรับรอง 50 ทวิ — เก็บแยกช่องเท่ากับ key หัวแบบของ rdform (addr_*) ทีละช่อง
// เพราะหัวแบบของกรมสรรพากรแยกช่อง (Champ เก็บเป็นข้อความเดียว — ADR docs/kms/decisions/2026-09-25-company-tax-address.md)
package taxaddress

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Keys - ช่องที่อยู่บนหัวแบบ rdform ตามลำดับบนหัวแบบ; ใช้เป็นทั้ง json ของ API บริษัทและชื่อคอลัมน์ในตาราง companies
var Keys = []string{"addr_building", "addr_room", "addr_floor", "addr_village", "addr_no", "addr_moo", "addr_soi",
	"addr_junction", "addr_road", "addr_subdistrict", "addr_district", "addr_province", "addr_postcode"}

// PhoneKey - ช่องโทรศัพท์บนหัวแบบ (มีเฉพาะบางแบบ) + คอลัมน์ในตาราง companies
const PhoneKey = "phone"

// PhoneMaxRunes - ความยาวสูงสุดของเบอร์โทรศัพท์ (นับตัวอักษร)
const PhoneMaxRunes = 50

// Address - ที่อยู่แยกช่อง json = key หัวแบบ rdform
type Address struct {
	Building    string `json:"addr_building"`
	Room        string `json:"addr_room"`
	Floor       string `json:"addr_floor"`
	Village     string `json:"addr_village"`
	No          string `json:"addr_no"`
	Moo         string `json:"addr_moo"`
	Soi         string `json:"addr_soi"`
	Junction    string `json:"addr_junction"`
	Road        string `json:"addr_road"`
	Subdistrict string `json:"addr_subdistrict"`
	District    string `json:"addr_district"`
	Province    string `json:"addr_province"`
	Postcode    string `json:"addr_postcode"`
}

// Fields - pointer ของทุกช่องตามลำดับ Keys (ใช้ Scan จากฐานข้อมูลและวนตรวจทีละช่อง)
func (a *Address) Fields() []*string {
	return []*string{&a.Building, &a.Room, &a.Floor, &a.Village, &a.No, &a.Moo, &a.Soi,
		&a.Junction, &a.Road, &a.Subdistrict, &a.District, &a.Province, &a.Postcode}
}

// Values - key หัวแบบ → ค่า เฉพาะช่องที่มีค่า (หลัง trim)
func (a Address) Values() map[string]string {
	values := map[string]string{}
	for i, field := range a.Fields() {
		if v := strings.TrimSpace(*field); v != "" {
			values[Keys[i]] = v
		}
	}
	return values
}

// IsBlank - ไม่มีช่องใดมีค่า
func (a Address) IsBlank() bool {
	return len(a.Values()) == 0
}

// Normalize - ตัดช่องว่างหัวท้ายทุกช่อง (ไม่แก้ข้อความข้างใน)
func (a *Address) Normalize() {
	for _, field := range a.Fields() {
		*field = strings.TrimSpace(*field)
	}
}

// NormalizePhone - ตัดช่องว่างหัวท้าย
func NormalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

// maxRunes - ความยาวสูงสุด (ตัวอักษร) เท่ากับช่องที่อยู่ของไฟล์ยื่นกรมสรรพากร (rdfile/check.go address) เพื่อให้ค่าจากทะเบียน
// ส่งออกไฟล์ได้เลย; แยก (addr_junction) ไม่มีช่องในไฟล์ จึงใช้ 100 เท่าตรอก/ซอยและถนน
var maxRunes = map[string]int{
	"addr_building": 40, "addr_room": 20, "addr_floor": 20, "addr_village": 100, "addr_no": 20, "addr_moo": 20,
	"addr_soi": 100, "addr_junction": 100, "addr_road": 100, "addr_subdistrict": 50, "addr_district": 50, "addr_province": 50,
}

// เหตุผลที่ช่องไม่ผ่าน
const (
	ReasonTooLong  = "too_long"
	ReasonControl  = "control_char"
	ReasonPostcode = "postcode"
)

// FieldError - ช่องที่ไม่ผ่าน: Field = key หัวแบบ (addr_* หรือ phone), Max = ความยาวสูงสุด (เฉพาะ too_long)
type FieldError struct {
	Field  string
	Reason string
	Max    int
}

func (e *FieldError) Error() string {
	if e.Reason == ReasonTooLong {
		return fmt.Sprintf("%s longer than %d characters", e.Field, e.Max)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Reason)
}

// Validate - ตรวจหลัง Normalize ทีละช่องตามลำดับ Keys แล้วโทรศัพท์: ห้ามอักขระควบคุม (ขึ้นบรรทัด/แท็บ — ที่อยู่บน 50 ทวิ
// เป็นบรรทัดเดียว และไฟล์ยื่นเป็นข้อความบรรทัดละรายการ), ยาวไม่เกิน maxRunes, รหัสไปรษณีย์ว่างหรือตัวเลข 5 หลัก — คืนช่องแรกที่ผิด
func Validate(a Address, phone string) error {
	fields := a.Fields()
	for i, key := range Keys {
		value := *fields[i]
		if hasControl(value) {
			return &FieldError{Field: key, Reason: ReasonControl}
		}
		if key == "addr_postcode" {
			if value != "" && !fiveDigits(value) {
				return &FieldError{Field: key, Reason: ReasonPostcode}
			}
			continue
		}
		if max := maxRunes[key]; utf8.RuneCountInString(value) > max {
			return &FieldError{Field: key, Reason: ReasonTooLong, Max: max}
		}
	}
	if hasControl(phone) {
		return &FieldError{Field: PhoneKey, Reason: ReasonControl}
	}
	if utf8.RuneCountInString(phone) > PhoneMaxRunes {
		return &FieldError{Field: PhoneKey, Reason: ReasonTooLong, Max: PhoneMaxRunes}
	}
	return nil
}

// AsFieldError - ดึง FieldError ออกจาก err (nil ถ้าไม่ใช่)
func AsFieldError(err error) *FieldError {
	var fe *FieldError
	if errors.As(err, &fe) {
		return fe
	}
	return nil
}

func hasControl(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func fiveDigits(s string) bool {
	if len(s) != 5 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
