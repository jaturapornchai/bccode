package taxaddress

import (
	"strings"
	"unicode/utf8"
)

// คำนำหน้าแต่ละส่วนของที่อยู่ = ถ้อยคำบนแบบทางการ (เนื้อหาเอกสาร ไม่ใช่ป้ายบนจอ จึงไม่ผ่าน languages.tsv):
//   - แบบ 50 ทวิ ใต้ช่องที่อยู่ (docs/kms/21 รหัส RD-50TAWI-FORM): "(ให้ระบุ ชื่ออาคาร/หมู่บ้าน ห้องเลขที่ ชั้นที่ เลขที่
//     ตรอก/ซอย หมู่ที่ ถนน ตำบล/แขวง อำเภอ/เขต จังหวัด)" — ลำดับและคำนำหน้าตามข้อความนี้
//   - กรุงเทพมหานครใช้ แขวง/เขต และไม่มีคำว่าจังหวัด: พ.ร.บ.ระเบียบบริหารราชการกรุงเทพมหานคร พ.ศ. 2528 มาตรา 7
//     (แบ่งพื้นที่เป็นเขตและแขวง) และมาตรา 8 (กฎหมายที่อ้างถึงจังหวัด = กรุงเทพมหานคร, อำเภอ = เขต, ตำบล = แขวง)
//     http://www.local.moi.go.th/2009/pdf/p.bangkok2528.pdf (docs/kms/21 รหัส TH-BMA-ACT-2528)
//   - แยก ไม่อยู่ในรายการของแบบ 50 ทวิ แต่เป็นช่องบนหัวแบบ ภ.พ.30/ภ.ง.ด. ("แยก") ที่ขยายความตรอก/ซอย จึงต่อท้ายตรอก/ซอย
const (
	prefixBuilding = "อาคาร"
	prefixVillage  = "หมู่บ้าน"
	prefixRoom     = "ห้องเลขที่"
	prefixFloor    = "ชั้นที่"
	prefixNo       = "เลขที่"
	prefixSoi      = "ซอย"
	prefixLane     = "ตรอก"
	prefixJunction = "แยก"
	prefixMoo      = "หมู่ที่"
	prefixRoad     = "ถนน"
	prefixTambon   = "ตำบล"
	prefixKhwaeng  = "แขวง"
	prefixAmphoe   = "อำเภอ"
	prefixKhet     = "เขต"
	prefixProvince = "จังหวัด"
	bangkok        = "กรุงเทพมหานคร"
)

// bangkokNames - ชื่อที่หมายถึงกรุงเทพมหานคร (ชื่อเต็ม + ชื่อย่อที่ใช้ทั่วไป) หลังตัดคำว่า "จังหวัด"/"จ." ออก
var bangkokNames = map[string]bool{bangkok: true, "กรุงเทพฯ": true, "กรุงเทพ": true, "กทม.": true, "กทม": true}

// IsBangkok - จังหวัดเป็นกรุงเทพมหานคร
func IsBangkok(province string) bool {
	p := strings.TrimSpace(province)
	for _, prefix := range []string{prefixProvince, "จ."} {
		p = strings.TrimSpace(strings.TrimPrefix(p, prefix))
	}
	return bangkokNames[p]
}

// part - คำนำหน้า + ค่า; ไม่เติมซ้ำเมื่อผู้ใช้พิมพ์คำนำหน้า (หรือคำย่อ/คำคู่ของช่องเดียวกัน) มาแล้ว; ค่าว่าง = ""
// ค่าที่ขึ้นต้นด้วยอักษรไทยต่อติดกัน ("ถนนพหลโยธิน", "ถนนเอกชัย") ค่าอื่นเว้นวรรค ("เลขที่ 88/12", "อาคาร Sathorn City")
func part(prefix, value string, already ...string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, typed := range append([]string{prefix}, already...) {
		if strings.HasPrefix(value, typed) {
			return value
		}
	}
	// อักษร/สระไทย U+0E01–U+0E4F (ไม่รวมเลขไทย ๐–๙ ซึ่งเว้นวรรคเหมือนเลขอารบิก)
	if r, _ := utf8.DecodeRuneInString(value); r >= 0x0E01 && r <= 0x0E4F {
		return prefix + value
	}
	return prefix + " " + value
}

// Line - ที่อยู่บรรทัดเดียวสำหรับหนังสือรับรอง 50 ทวิ ตามลำดับคำชี้แจงบนแบบ (ดูคำอธิบายด้านบน) + รหัสไปรษณีย์; ข้ามช่องว่าง
func (a Address) Line() string {
	bkk := IsBangkok(a.Province)
	subdistrictPrefix, districtPrefix := prefixTambon, prefixAmphoe
	if bkk {
		subdistrictPrefix, districtPrefix = prefixKhwaeng, prefixKhet
	}
	province := part(prefixProvince, a.Province, "จ.")
	if bkk {
		province = strings.TrimSpace(a.Province) // ชื่อจังหวัดของกรุงเทพมหานครตามที่ผู้ใช้กรอก ไม่เติม "จังหวัด"
	}
	parts := []string{
		part(prefixBuilding, a.Building),
		part(prefixVillage, a.Village, "มบ."),
		part(prefixRoom, a.Room, "ห้อง"),
		part(prefixFloor, a.Floor, "ชั้น"),
		part(prefixNo, a.No),
		part(prefixSoi, a.Soi, prefixLane, "ซ."),
		part(prefixJunction, a.Junction),
		part(prefixMoo, a.Moo, "หมู่", "ม."),
		part(prefixRoad, a.Road, "ถ."),
		part(subdistrictPrefix, a.Subdistrict, prefixTambon, prefixKhwaeng, "ต."),
		part(districtPrefix, a.District, prefixAmphoe, prefixKhet, "อ."),
		province,
		strings.TrimSpace(a.Postcode),
	}
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}
