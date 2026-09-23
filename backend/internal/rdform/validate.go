package rdform

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// เพดานเอกสารหนึ่งแบบ: กันเอกสารใหญ่ผิดปกติ (ใบแนบ ภ.ง.ด. ของ SME หลักสิบถึงหลักร้อยราย)
const (
	MaxRows       = 5000
	MaxSheets     = 500
	maxValueRunes = 300
)

// FieldError - ค่าที่กรอกผิดรูปแบบ: Key = ช่อง, Row = แถว/แผ่นที่ (เริ่ม 1; 0 = ช่องของแบบ)
type FieldError struct {
	Key string
	Row int
}

func (e *FieldError) Error() string {
	if e.Row > 0 {
		return fmt.Sprintf("rdform: invalid %s at row %d", e.Key, e.Row)
	}
	return "rdform: invalid " + e.Key
}

func (e *FieldError) Unwrap() error { return ErrInvalidValue }

type typed struct {
	typ     string
	options map[string]bool
}

func typesOf(fields []Field) map[string]typed {
	out := map[string]typed{}
	for _, f := range fields {
		t := typed{typ: f.Type}
		if f.Type == "choice" {
			t.options = map[string]bool{}
			for _, o := range f.Options {
				t.options[o.Value] = true
			}
		}
		out[f.Key] = t
	}
	return out
}

// Validate - ตรวจค่าตามชนิดช่องก่อนบันทึก/พิมพ์: เงินเป็นทศนิยม, ตัวเลขเป็นตัวเลข, ตัวเลือกอยู่ในแบบ;
// ช่องที่แบบไม่มีถูกปฏิเสธ (กันส่ง key ผิดแล้วเงียบหายจากกระดาษ)
func Validate(code string, doc Document) error {
	cover, err := Lookup(code)
	if err != nil {
		return err
	}
	attach := Attachment(code)
	header := typesOf(cover.Fields)
	for _, p := range cover.MoneyPairs() {
		header[p.Key] = typed{typ: "money"}
	}
	rowTypes := map[string]typed{}
	if attach != nil {
		for k, t := range typesOf(attach.Fields) {
			if _, ok := header[k]; !ok {
				if attach.Table != nil && t.typ == "choice" {
					rowTypes[k] = t // ช่องเลือกหัวใบแนบที่แยกแผ่นตามรายการ (ภ.ง.ด.2 ประเภทเงินได้)
				} else {
					header[k] = t
				}
			}
		}
		if attach.Table != nil {
			for _, c := range attach.Table.Columns {
				rowTypes[c.Key] = typed{typ: c.Type}
			}
			for _, g := range attach.Table.DigitGroups() {
				rowTypes[g.Key] = typed{typ: "digits"}
			}
		} else {
			rowTypes = typesOf(attach.Fields)
		}
	}
	if len(doc.Rows) > MaxRows || len(doc.Sheets) > MaxSheets {
		return &FieldError{Key: "rows"}
	}
	if (len(doc.Rows) > 0 && (attach == nil || attach.Table == nil)) || (len(doc.Sheets) > 0 && (attach == nil || attach.Table != nil)) {
		return &FieldError{Key: "rows"}
	}
	for k, v := range doc.Values {
		if err := checkValue(header, k, v, 0); err != nil {
			return err
		}
	}
	for i, r := range append(append([]map[string]string{}, doc.Rows...), doc.Sheets...) {
		for k, v := range r {
			if err := checkValue(rowTypes, k, v, i+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkValue(types map[string]typed, key, v string, row int) error {
	t, ok := types[key]
	if !ok || utf8.RuneCountInString(v) > maxValueRunes || strings.ContainsAny(v, "\r\n\t") {
		return &FieldError{Key: key, Row: row}
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	switch t.typ {
	case "money":
		if _, _, err := money(v); err != nil {
			return &FieldError{Key: key, Row: row}
		}
	case "choice":
		if !t.options[v] {
			return &FieldError{Key: key, Row: row}
		}
	case "taxid":
		if n := countDigits(v); n != 13 || n != len(strings.NewReplacer("-", "", " ", "").Replace(v)) {
			return &FieldError{Key: key, Row: row}
		}
	}
	return nil
}

func countDigits(v string) int {
	n := 0
	for _, r := range v {
		if r >= '0' && r <= '9' {
			n++
		}
	}
	return n
}
