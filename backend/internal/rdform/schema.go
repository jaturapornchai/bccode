package rdform

import "strings"

// EditorOption / EditorField - ช่องสำหรับจอแก้ไขแบบ (ไม่มีพิกัด): จอสร้างฟอร์มจากสเปกนี้ทั้งหมด
type EditorOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type EditorField struct {
	Key     string         `json:"key"`
	Type    string         `json:"type"`
	Label   string         `json:"label"`
	Group   string         `json:"group,omitempty"`
	Note    string         `json:"note,omitempty"`
	Options []EditorOption `json:"options,omitempty"`
}

// AttachmentSchema - ใบแนบ: mode "rows" = ตารางรายการ (ระบบแบ่งแผ่นเอง), "sheets" = หนึ่งแผ่นต่อหนึ่งรายการ
type AttachmentSchema struct {
	Code         string        `json:"code"`
	Title        string        `json:"title"`
	Label        string        `json:"label"`
	Mode         string        `json:"mode"`
	RowsPerSheet int           `json:"rowspersheet,omitempty"`
	Columns      []EditorField `json:"columns"`
}

type Schema struct {
	Code       string            `json:"code"`
	Title      string            `json:"title"`
	Fields     []EditorField     `json:"fields"`
	Attachment *AttachmentSchema `json:"attachment,omitempty"`
}

// autoKey - ช่องที่ตัวกรอกเติมเองตามการแบ่งแผ่น ไม่ให้ผู้ใช้กรอก
func autoKey(key string) bool {
	return key == "sheet_no" || key == "sheet_total" || key == "month_name" || key == "seq" ||
		strings.HasPrefix(key, "page_total_")
}

// headerGroups - หมวดหัวใบแนบที่ใช้ค่าเดียวกับแบบ (ผู้ยื่น/งวด/การยื่น/ผู้ลงนาม) ไม่ใช่ค่ารายแผ่น
var headerGroups = map[string]bool{"payer": true, "filing": true, "period": true, "signature": true, "sheet": true}

func editorField(f Field) EditorField {
	e := EditorField{Key: f.Key, Type: f.Type, Label: f.Label, Group: f.Group, Note: f.Note}
	for _, o := range f.Options {
		e.Options = append(e.Options, EditorOption{Value: o.Value, Label: o.Label})
	}
	return e
}

// editorFields - ช่องของแบบสำหรับจอ: รวมคู่ X_baht/X_satang เป็นยอดเงินเดียว X
func editorFields(s *Spec, keep func(Field) bool) []EditorField {
	pairs := map[string]MoneyPair{}
	for _, p := range s.MoneyPairs() {
		pairs[p.Baht.Key], pairs[p.Satang.Key] = p, p
	}
	var out []EditorField
	for _, f := range s.Fields {
		if autoKey(f.Key) || !keep(f) {
			continue
		}
		if p, ok := pairs[f.Key]; ok {
			if f.Key == p.Baht.Key {
				label := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(f.Label), "(บาท)"))
				out = append(out, EditorField{Key: p.Key, Type: "money", Label: label, Group: f.Group})
			}
			continue
		}
		out = append(out, editorField(f))
	}
	return out
}

// SchemaFor - สเปกจอแก้ไขของแบบ (รวมใบแนบ)
func SchemaFor(code string) (*Schema, error) {
	cover, err := Lookup(code)
	if err != nil {
		return nil, err
	}
	out := &Schema{Code: cover.Code, Title: cover.Title, Fields: editorFields(cover, func(Field) bool { return true })}
	attach := Attachment(code)
	if attach == nil {
		return out, nil
	}
	a := &AttachmentSchema{Code: attach.Code, Title: attach.Title}
	if attach.Table != nil {
		a.Mode, a.Label, a.RowsPerSheet = "rows", attach.Table.Label, len(attach.Table.Rows)
		// ช่องเลือกหัวใบแนบที่แบบไม่มี (เช่น ประเภทเงินได้ของ ภ.ง.ด.2) = ค่ารายการ ระบบแยกแผ่นตามค่านี้
		for _, f := range attach.Fields {
			if f.Type == "choice" && cover.Field(f.Key) == nil {
				a.Columns = append(a.Columns, editorField(f))
			}
		}
		grouped := map[string]string{}
		for _, g := range attach.Table.DigitGroups() {
			for _, p := range g.Parts {
				grouped[p.Key] = g.Key
			}
		}
		for _, c := range attach.Table.Columns {
			if g, ok := grouped[c.Key]; ok {
				if c.Key == g+"_d1" {
					label := strings.TrimSpace(strings.TrimSuffix(c.Label, "หลักที่ 1"))
					a.Columns = append(a.Columns, EditorField{Key: g, Type: "digits", Label: label})
				}
				continue
			}
			if !autoKey(c.Key) {
				a.Columns = append(a.Columns, EditorField{Key: c.Key, Type: c.Type, Label: c.Label, Note: c.Note})
			}
		}
	} else {
		a.Mode, a.Label = "sheets", attach.Title
		a.Columns = editorFields(attach, func(f Field) bool { return !headerGroups[f.Group] })
	}
	out.Attachment = a
	return out, nil
}
