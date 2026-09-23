package rdform

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/pdftext"
)

var (
	ErrUnknownForm  = errors.New("rdform: unknown form")
	ErrInvalidValue = errors.New("rdform: invalid value")
)

// Document - ค่าที่จะกรอก: Values = ช่องของแบบ (และช่องหัวใบแนบที่ key ตรงกัน เช่น tax_id, signer_name),
// Rows = รายการในตารางใบแนบ (key = column key) — ตัวกรอกแบ่งแผ่นให้เองตามจำนวนแถวบนกระดาษ,
// Sheets = ใบแนบที่ไม่มีตาราง หนึ่งแผ่นต่อหนึ่งรายการ (เช่น ภ.ธ.40 หนึ่งแผ่นต่อหนึ่งสถานประกอบการ)
type Document struct {
	Values map[string]string   `json:"values"`
	Rows   []map[string]string `json:"rows,omitempty"`
	Sheets []map[string]string `json:"sheets,omitempty"`
}

const (
	maxFont = 11.0
	minFont = 5.0
	pad     = 1.5
)

// Sheet - หนึ่งแผ่นของใบแนบหลังแบ่งแผ่น
type Sheet struct {
	Values map[string]string
	Rows   []map[string]string
}

// Paginate - แบ่งรายการเป็นแผ่นใบแนบ: แยกแผ่นตามช่องเลือกหัวใบแนบที่รายการระบุ (เช่น ประเภทเงินได้ของ ภ.ง.ด.2
// ต้อง "เลือกเพียงข้อเดียว" ต่อแผ่น), เติมลำดับที่ต่อเนื่องทุกแผ่น, แผ่นที่/ในจำนวน และยอดรวมท้ายแผ่น
func Paginate(attach *Spec, doc Document) ([]Sheet, error) {
	if attach == nil {
		return nil, nil
	}
	if attach.Table == nil {
		return fieldSheets(doc), nil
	}
	if len(doc.Rows) == 0 {
		return nil, nil
	}
	perSheet := len(attach.Table.Rows)
	var groupKey string
	for _, f := range attach.Fields {
		if f.Type != "choice" {
			continue
		}
		for _, r := range doc.Rows {
			if strings.TrimSpace(r[f.Key]) != "" {
				groupKey = f.Key
			}
		}
	}
	hasSeq := false
	for _, c := range attach.Table.Columns {
		hasSeq = hasSeq || c.Key == "seq"
	}
	var order []string
	groups := map[string][]map[string]string{}
	for _, r := range doc.Rows {
		g := strings.TrimSpace(r[groupKey])
		if _, ok := groups[g]; !ok {
			order = append(order, g)
		}
		groups[g] = append(groups[g], r)
	}
	var sheets []Sheet
	seq := 0
	for _, g := range order {
		rows := groups[g]
		for start := 0; start < len(rows); start += perSheet {
			end := min(start+perSheet, len(rows))
			sheet := Sheet{Values: map[string]string{}}
			for k, v := range doc.Values {
				sheet.Values[k] = v
			}
			if groupKey != "" {
				sheet.Values[groupKey] = g
			}
			for _, r := range rows[start:end] {
				row := make(map[string]string, len(r))
				for k, v := range r {
					row[k] = v
				}
				seq++
				if hasSeq && strings.TrimSpace(row["seq"]) == "" {
					row["seq"] = strconv.Itoa(seq)
				}
				sheet.Rows = append(sheet.Rows, row)
			}
			sheets = append(sheets, sheet)
		}
	}
	for i := range sheets {
		sheets[i].Values["sheet_no"] = strconv.Itoa(i + 1)
		sheets[i].Values["sheet_total"] = strconv.Itoa(len(sheets))
		for _, f := range attach.Fields {
			if !strings.HasPrefix(f.Key, "page_total_") {
				continue
			}
			sum, err := pageTotal(attach.Table, sheets[i].Rows, strings.TrimPrefix(f.Key, "page_total_"))
			if err != nil {
				return nil, err
			}
			sheets[i].Values[f.Key] = sum
		}
	}
	return sheets, nil
}

// fieldSheets - ใบแนบไม่มีตาราง: ค่าหัวแบบ + ค่าของแผ่นนั้น
func fieldSheets(doc Document) []Sheet {
	sheets := make([]Sheet, 0, len(doc.Sheets))
	for i, sv := range doc.Sheets {
		v := map[string]string{}
		for k, x := range doc.Values {
			v[k] = x
		}
		for k, x := range sv {
			v[k] = x
		}
		v["sheet_no"] = strconv.Itoa(i + 1)
		v["sheet_total"] = strconv.Itoa(len(doc.Sheets))
		sheets = append(sheets, Sheet{Values: v})
	}
	return sheets
}

// pageTotal - รวมทุกคอลัมน์เงินที่ชื่อ = name หรือลงท้าย _name (l1_amount + l2_amount + l3_amount)
func pageTotal(t *Table, rows []map[string]string, name string) (string, error) {
	total := decimal.Zero
	for _, c := range t.Columns {
		if c.Type != "money" || (c.Key != name && !strings.HasSuffix(c.Key, "_"+name)) {
			continue
		}
		for _, r := range rows {
			v, ok, err := money(r[c.Key])
			if err != nil {
				return "", fmt.Errorf("%w: %s", err, c.Key)
			}
			if ok {
				total = total.Add(v)
			}
		}
	}
	return total.StringFixed(2), nil
}

// Render - PDF แบบฟอร์มพร้อมใบแนบทุกแผ่น
func Render(code string, doc Document) ([]byte, error) {
	cover, err := Lookup(code)
	if err != nil {
		return nil, err
	}
	shaper, err := pdftext.Default()
	if err != nil {
		return nil, err
	}
	attach := Attachment(code)
	values := map[string]string{}
	for k, v := range doc.Values {
		values[k] = v
	}
	if m, err := strconv.Atoi(strings.TrimSpace(values["month"])); err == nil && m >= 1 && m <= 12 {
		setDefault(values, "month_name", thaiMonths[m-1]) // ใบแนบบางแบบเขียนชื่อเดือนแทนการติ๊ก
	}
	sheets, err := Paginate(attach, Document{Values: values, Rows: doc.Rows, Sheets: doc.Sheets})
	if err != nil {
		return nil, err
	}
	if len(sheets) > 0 {
		setDefault(values, "has_attachment", "1")
		setDefault(values, "attach_sheets", strconv.Itoa(len(sheets)))
		if len(doc.Rows) > 0 {
			setDefault(values, "attach_payees", strconv.Itoa(len(doc.Rows)))
		}
	}

	o := pdftext.NewOverlay(shaper)
	if err := drawSheet(o, cover, values, nil); err != nil {
		return nil, err
	}
	parts := [][]byte{}
	base, err := blank(cover)
	if err != nil {
		return nil, err
	}
	parts = append(parts, base)
	if len(sheets) > 0 {
		attachBase, err := blank(attach)
		if err != nil {
			return nil, err
		}
		for _, s := range sheets {
			if err := drawSheet(o, attach, s.Values, s.Rows); err != nil {
				return nil, err
			}
			parts = append(parts, attachBase)
		}
	}
	merged := parts[0]
	if len(parts) > 1 {
		if merged, err = pdftext.Merge(parts...); err != nil {
			return nil, err
		}
	}
	return pdftext.Stamp(merged, o)
}

var thaiMonths = [12]string{"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}

func setDefault(m map[string]string, k, v string) {
	if strings.TrimSpace(m[k]) == "" {
		m[k] = v
	}
}

var (
	blankMu    sync.Mutex
	blankCache = map[string][]byte{}
)

// blank - แบบเปล่า (ถอดช่องกรอกออก เหลือเฉพาะหน้าแบบ ไม่เอาหน้าคำชี้แจง)
func blank(s *Spec) ([]byte, error) {
	blankMu.Lock()
	defer blankMu.Unlock()
	if b, ok := blankCache[s.Code]; ok {
		return b, nil
	}
	raw, err := template(s)
	if err != nil {
		return nil, err
	}
	b, err := pdftext.BlankForm(raw)
	if err != nil {
		return nil, fmt.Errorf("rdform %s: %w", s.Code, err)
	}
	pages := make([]string, len(s.Pages))
	for i, p := range s.Pages {
		pages[i] = strconv.Itoa(p)
	}
	var out bytes.Buffer
	if err := api.Trim(bytes.NewReader(b), &out, pages, pdftext.Config()); err != nil {
		return nil, fmt.Errorf("rdform %s trim: %w", s.Code, err)
	}
	blankCache[s.Code] = out.Bytes()
	return out.Bytes(), nil
}

// drawSheet - เพิ่มหน้าชั้นข้อความของสเปกหนึ่งชุด (ทุกหน้าใน s.Pages) แล้ววาดทุกช่อง
func drawSheet(o *pdftext.Overlay, s *Spec, values map[string]string, rows []map[string]string) error {
	if err := splitMoneyPairs(s, values); err != nil {
		return err
	}
	pages := map[int]*pdftext.Page{}
	for i, p := range s.Pages {
		pages[p] = o.NewPage(s.Sizes[i][0], s.Sizes[i][1])
	}
	for _, f := range s.Fields {
		v := strings.TrimSpace(values[f.Key])
		if v == "" {
			continue
		}
		if f.Type == "choice" {
			for _, opt := range f.Options {
				if opt.Value == v {
					drawCheck(o, pages[opt.Page], opt.Box)
				}
			}
			continue
		}
		if err := drawValue(o, pages[f.Page], f.Type, f.Box, v); err != nil {
			return fmt.Errorf("%w (%s)", err, f.Key)
		}
	}
	if s.Table == nil {
		return nil
	}
	types := map[string]string{}
	for _, c := range s.Table.Columns {
		types[c.Key] = c.Type
	}
	groups := s.Table.DigitGroups()
	for i, row := range rows {
		if i >= len(s.Table.Rows) {
			break
		}
		row = expandDigitGroups(groups, row)
		for key, box := range s.Table.Rows[i] {
			v := strings.TrimSpace(row[key])
			if v == "" {
				continue
			}
			if err := drawValue(o, pages[box.Page], types[key], box, v); err != nil {
				return fmt.Errorf("%w (%s แถว %d)", err, key, i+1)
			}
		}
	}
	return nil
}

func drawValue(o *pdftext.Overlay, p *pdftext.Page, typ string, b Box, v string) error {
	switch typ {
	case "check":
		if truthy(v) {
			drawCheck(o, p, b)
		}
	case "money":
		return drawMoney(o, p, b, v)
	case "taxid", "digits":
		drawDigits(o, p, b, v, typ == "taxid")
	default:
		drawText(o, p, b, v)
	}
	return nil
}

func truthy(v string) bool {
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y", "on", "x":
		return true
	}
	return false
}

func height(b Box) float64 { return b.Rect[3] - b.Rect[1] }

func fontFor(h float64) float64 { return math.Max(minFont, math.Min(maxFont, h*0.72)) }

// baseline - ตัวอักษรสูงราว 0.62 เท่าของขนาดฟอนต์ วางให้อยู่กลางช่องตามแนวตั้ง
func baseline(b Box, size float64) float64 {
	return b.Rect[1] + math.Max(1, (height(b)-0.62*size)/2)
}

func drawCheck(o *pdftext.Overlay, p *pdftext.Page, b Box) {
	o.Check(p, (b.Rect[0]+b.Rect[2])/2, (b.Rect[1]+b.Rect[3])/2)
}

// drawText - ข้อความบรรทัดเดียว ชิดตาม Q ของช่อง ย่อฟอนต์ให้พอดีความกว้าง
func drawText(o *pdftext.Overlay, p *pdftext.Page, b Box, v string) {
	size, _ := o.Fit(v, b.Rect[2]-b.Rect[0]-2*pad, fontFor(height(b)), minFont)
	x, a := b.Rect[0]+pad, pdftext.AlignLeft
	switch b.Align {
	case 1:
		x, a = (b.Rect[0]+b.Rect[2])/2, pdftext.AlignCenter
	case 2:
		x, a = b.Rect[2]-pad, pdftext.AlignRight
	}
	o.Text(p, x, baseline(b, size), v, size, a)
}

// expandDigitGroups - ค่า X = "00001" → X_d1..X_d5 ชิดขวา (ไม่ทับค่าที่กรอกแยกหลักมาแล้ว)
func expandDigitGroups(groups []DigitGroup, row map[string]string) map[string]string {
	if len(groups) == 0 {
		return row
	}
	out := make(map[string]string, len(row))
	for k, v := range row {
		out[k] = v
	}
	for _, g := range groups {
		d := []rune(strings.TrimSpace(row[g.Key]))
		if len(d) == 0 || len(d) > len(g.Parts) {
			continue
		}
		offset := len(g.Parts) - len(d)
		for i, r := range d {
			if strings.TrimSpace(out[g.Parts[offset+i].Key]) == "" {
				out[g.Parts[offset+i].Key] = string(r)
			}
		}
	}
	return out
}

// splitMoneyPairs - แบบที่ช่องบาทกับช่องสตางค์เป็นคนละช่อง (X_baht + X_satang): จอแก้ไขส่งยอดเดียว X
// แล้วตัวกรอกแยกให้ — ห้ามให้ผู้ใช้กรอกสองช่องเอง (พิมพ์สตางค์พลาดง่าย)
func splitMoneyPairs(s *Spec, values map[string]string) error {
	for _, pair := range s.MoneyPairs() {
		v, ok, err := money(values[pair.Key])
		if err != nil {
			return fmt.Errorf("%w (%s)", err, pair.Key)
		}
		if !ok || strings.TrimSpace(values[pair.Baht.Key]+values[pair.Satang.Key]) != "" {
			continue
		}
		fixed := v.Abs().StringFixed(2)
		baht := fixed[:len(fixed)-3]
		if pair.Baht.Type != "digits" {
			baht = thousands(baht)
		}
		if v.IsNegative() {
			baht = "-" + baht
		}
		values[pair.Baht.Key], values[pair.Satang.Key] = baht, fixed[len(fixed)-2:]
	}
	return nil
}

// taxIDDashes - ตำแหน่งขีดของเลขประจำตัว 13 หลักแบบ 1-2345-67890-12-3 (ช่อง comb 17)
var taxIDDashes = map[int]bool{1: true, 6: true, 12: true, 15: true}

// drawDigits - ทีละหลักลงช่องที่ตรวจเจอ (หรือช่อง comb เท่ากัน); หลักน้อยกว่าช่อง = ชิดขวา
func drawDigits(o *pdftext.Overlay, p *pdftext.Page, b Box, v string, taxID bool) {
	var d []rune
	for _, r := range v {
		if r >= '0' && r <= '9' {
			d = append(d, r)
		}
	}
	if b.Comb > 0 && len(b.Cells) == 0 && len([]rune(v)) == b.Comb && len(d) < b.Comb {
		// ค่าที่จัดรูปมาเต็มช่องพอดี (เช่น อัตราแลกเปลี่ยน "0032.2076") วางทีละตัวตามตำแหน่ง รวมจุด/ขีด
		w := (b.Rect[2] - b.Rect[0]) / float64(b.Comb)
		size := math.Min(fontFor(height(b)), w/0.62)
		for i, r := range []rune(v) {
			o.Text(p, b.Rect[0]+w*(float64(i)+0.5), baseline(b, size), string(r), size, pdftext.AlignCenter)
		}
		return
	}
	cells := b.Cells
	if len(cells) == 0 && b.Comb > 0 {
		w := (b.Rect[2] - b.Rect[0]) / float64(b.Comb)
		for i := 0; i < b.Comb; i++ {
			if taxID && b.Comb == 17 && len(d) == 13 && taxIDDashes[i] {
				continue // ช่องขีดของแบบ comb 17
			}
			cells = append(cells, [2]float64{b.Rect[0] + w*float64(i), b.Rect[0] + w*float64(i+1)})
		}
	}
	if len(cells) == 0 || len(d) > len(cells) {
		drawText(o, p, b, v)
		return
	}
	size := fontFor(height(b))
	for _, c := range cells {
		size = math.Min(size, (c[1]-c[0])/0.62)
	}
	offset := len(cells) - len(d)
	for i, r := range d {
		c := cells[offset+i]
		o.Text(p, (c[0]+c[1])/2, baseline(b, size), string(r), size, pdftext.AlignCenter)
	}
}

// money - ทศนิยมจากสตริง ("1234.5", "1,234.50", "-12") — ห้าม float (กฎตัวเลขบัญชี)
func money(v string) (decimal.Decimal, bool, error) {
	v = strings.ReplaceAll(strings.TrimSpace(v), ",", "")
	if v == "" {
		return decimal.Zero, false, nil
	}
	d, err := decimal.NewFromString(v)
	if err != nil {
		return decimal.Zero, false, fmt.Errorf("%w: จำนวนเงิน %q", ErrInvalidValue, v)
	}
	return d, true, nil
}

// thousands - ใส่จุลภาคหลักพัน ("1234567" → "1,234,567")
func thousands(s string) string {
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	var out []byte
	for i := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

// drawMoneyCells - ช่องเงินทีละหลัก: หลักบาท (ไม่มีจุลภาค) ชิดขวาในช่องบาท, สตางค์ 2 หลักในช่องสตางค์;
// false = หลักบาทเกินจำนวนช่อง
func drawMoneyCells(o *pdftext.Overlay, p *pdftext.Page, b Box, baht, satang string) bool {
	digits := strings.ReplaceAll(baht, ",", "")
	bahtCells, satangCells := b.Cells[:len(b.Cells)-b.Satang], b.Cells[len(b.Cells)-b.Satang:]
	if len(digits) > len(bahtCells) || len(satang) != len(satangCells) {
		return false
	}
	size := fontFor(height(b))
	for _, c := range b.Cells {
		size = math.Min(size, (c[1]-c[0])/0.62)
	}
	put := func(cells [][2]float64, text string) {
		offset := len(cells) - len(text)
		for i := range text {
			c := cells[offset+i]
			o.Text(p, (c[0]+c[1])/2, baseline(b, size), text[i:i+1], size, pdftext.AlignCenter)
		}
	}
	put(bahtCells, digits)
	put(satangCells, satang)
	return true
}

// drawMoney - บาทชิดขวาก่อนเส้นแบ่ง สตางค์กลางช่องสตางค์; ไม่มีเส้นแบ่ง = "1,234.56" ชิดขวา
func drawMoney(o *pdftext.Overlay, p *pdftext.Page, b Box, v string) error {
	d, ok, err := money(v)
	if err != nil || !ok {
		return err
	}
	fixed := d.Abs().StringFixed(2)
	baht, satang := fixed[:len(fixed)-3], fixed[len(fixed)-2:]
	baht = thousands(baht)
	if d.IsNegative() {
		baht = "-" + baht
	}
	size := fontFor(height(b))
	if b.Satang > 0 && len(b.Cells) > b.Satang {
		if drawMoneyCells(o, p, b, baht, satang) {
			return nil
		}
		b.Split = b.Cells[len(b.Cells)-b.Satang][0] // หลักเกินช่อง: วางแบบบาท|สตางค์แทน
	}
	if b.Split == 0 {
		text := baht + "." + satang
		s, _ := o.Fit(text, b.Rect[2]-b.Rect[0]-2*pad, size, minFont)
		o.Text(p, b.Rect[2]-pad, baseline(b, s), text, s, pdftext.AlignRight)
		return nil
	}
	bahtEnd := b.Split
	if b.BahtEnd > 0 {
		bahtEnd = b.BahtEnd
	}
	s, _ := o.Fit(baht, bahtEnd-b.Rect[0]-2*pad, size, minFont)
	o.Text(p, bahtEnd-pad, baseline(b, s), baht, s, pdftext.AlignRight)
	o.Text(p, (b.Split+b.Rect[2])/2, baseline(b, size), satang, math.Min(size, (b.Rect[2]-b.Split)/1.2), pdftext.AlignCenter)
	return nil
}

// ThaiMonth - ชื่อเดือนภาษาไทยเต็ม (1-12) สำหรับช่อง "เดือน" ที่แบบให้เขียนเป็นตัวอักษร
func ThaiMonth(m int) string {
	if m < 1 || m > 12 {
		return ""
	}
	return thaiMonths[m-1]
}
