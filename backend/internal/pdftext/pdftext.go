// Package pdftext - กรอกข้อความลงแบบฟอร์ม PDF ทางการ (50 ทวิ, ภ.ง.ด., ภ.พ. ฯลฯ)
//
// ชั้นข้อความโปร่งใส (overlay) หนึ่งไฟล์หลายหน้า — จัดรูปอักษรไทย (สระบน/ล่าง วรรณยุกต์ซ้อน) ด้วย HarfBuzz
// ของ go-text/typesetting แล้ววางทีละ glyph ด้วยพิกัดจริง ฝังฟอนต์ Sarabun ทั้งไฟล์เป็น CIDFontType2 (Identity-H)
// และมี ToUnicode ให้ค้นหา/คัดลอกข้อความใน PDF ได้ จากนั้นประทับ (stamp) ลงแบบฟอร์มที่ถอดช่องกรอกออกแล้ว
package pdftext

import (
	"bytes"
	"compress/zlib"
	_ "embed"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

//go:embed assets/Sarabun-Regular.ttf
var Sarabun []byte

// pdfcpu ห้ามเขียน config ลง $HOME: container รันด้วย appuser ที่ไม่มี home (adduser -H) — ถ้าไม่ปิด
// จะ panic "config problem: mkdir /home/appuser: permission denied" ตอนสร้าง PDF บน production
func init() {
	api.DisableConfigDir()
}

// Config - ตั้งค่า pdfcpu แบบผ่อนปรน (แบบฟอร์มของกรมสรรพากรบางไฟล์ไม่ผ่าน strict validation)
func Config() *model.Configuration {
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	return conf
}

type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

type Shaper struct {
	face   *font.Face
	upem   float64
	ttf    []byte
	shaper shaping.HarfbuzzShaper
}

var (
	defaultOnce   sync.Once
	defaultShaper *Shaper
	defaultErr    error
)

// Default - ตัวจัดรูปอักษรของฟอนต์ Sarabun (สร้างครั้งเดียว ใช้ร่วมทั้งโปรแกรม)
func Default() (*Shaper, error) {
	defaultOnce.Do(func() { defaultShaper, defaultErr = NewShaper(Sarabun) })
	return defaultShaper, defaultErr
}

func NewShaper(ttf []byte) (*Shaper, error) {
	face, err := font.ParseTTF(bytes.NewReader(ttf))
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	return &Shaper{face: face, upem: float64(face.Upem()), ttf: ttf}, nil
}

// Width - ความกว้างของข้อความบรรทัดเดียวที่ขนาด size (pt)
func (s *Shaper) Width(text string, size float64) float64 {
	_, w := s.shape(text, size)
	return w
}

type placedGlyph struct {
	gid  uint16
	x, y float64 // ตำแหน่งจริงบนหน้า (pt)
	text string  // ข้อความของ cluster (ใส่ให้ glyph แรกของ cluster) สำหรับ ToUnicode
}

// shape - จัดข้อความหนึ่งบรรทัดที่ขนาด size คืน glyph (พิกัดสัมพัทธ์จากจุดเริ่ม) และความกว้างรวม
func (s *Shaper) shape(text string, size float64) ([]placedGlyph, float64) {
	runes := []rune(text)
	out := s.shaper.Shape(shaping.Input{
		Text: runes, RunStart: 0, RunEnd: len(runes), Direction: di.DirectionLTR, Face: s.face,
		Size: fixed.Int26_6(size * 64), Script: language.Thai, Language: language.NewLanguage("th"),
	})
	glyphs := make([]placedGlyph, 0, len(out.Glyphs))
	pen := 0.0
	seen := map[int]bool{}
	for _, g := range out.Glyphs {
		pg := placedGlyph{gid: uint16(g.GlyphID), x: pen + fromFixed(g.XOffset), y: fromFixed(g.YOffset)}
		if !seen[g.ClusterIndex] {
			seen[g.ClusterIndex] = true
			end := min(g.ClusterIndex+g.RuneCount, len(runes))
			pg.text = string(runes[g.ClusterIndex:end])
		}
		glyphs = append(glyphs, pg)
		pen += fromFixed(g.XAdvance)
	}
	return glyphs, pen
}

func fromFixed(v fixed.Int26_6) float64 { return float64(v) / 64 }

// Page - คำสั่งวาดของหนึ่งหน้า (ขนาดหน้าต้องตรงกับหน้าแบบฟอร์มที่จะประทับ)
type Page struct {
	W, H    float64
	content bytes.Buffer
}

type Overlay struct {
	shaper  *Shaper
	pages   []*Page
	widths  map[uint16]float64 // advance ที่ 1000 หน่วย สำหรับ /W
	unicode map[uint16]string
}

func NewOverlay(s *Shaper) *Overlay {
	return &Overlay{shaper: s, widths: map[uint16]float64{}, unicode: map[uint16]string{}}
}

func (o *Overlay) NewPage(w, h float64) *Page {
	p := &Page{W: w, H: h}
	o.pages = append(o.pages, p)
	return p
}

// Fit - ขนาดฟอนต์ที่ทำให้ข้อความกว้างไม่เกิน maxWidth (เริ่มที่ size ลดทีละ 0.25pt ถึง minSize)
func (o *Overlay) Fit(text string, maxWidth, size, minSize float64) (float64, bool) {
	for ; size >= minSize; size -= 0.25 {
		if _, w := o.shaper.shape(text, size); w <= maxWidth {
			return size, true
		}
	}
	return minSize, false
}

// Text - วางข้อความ; x คือขอบซ้าย/กึ่งกลาง/ขอบขวาตาม align
func (o *Overlay) Text(p *Page, x, baseline float64, text string, size float64, a Align) {
	glyphs, width := o.shaper.shape(text, size)
	switch a {
	case AlignCenter:
		x -= width / 2
	case AlignRight:
		x -= width
	}
	p.content.WriteString(fmt.Sprintf("BT /F1 %s Tf\n", num(size)))
	for _, g := range glyphs {
		if _, ok := o.widths[g.gid]; !ok {
			o.widths[g.gid] = float64(o.shaper.face.HorizontalAdvance(font.GID(g.gid))) * 1000 / o.shaper.upem
		}
		if g.text != "" {
			if _, ok := o.unicode[g.gid]; !ok {
				o.unicode[g.gid] = g.text
			}
		}
		p.content.WriteString(fmt.Sprintf("1 0 0 1 %s %s Tm <%04X> Tj\n", num(x+g.x), num(baseline+g.y), g.gid))
	}
	p.content.WriteString("ET\n")
}

// Check - เครื่องหมาย ✓ วาดเป็นเส้น (ฟอนต์ไทยทั่วไปไม่มี glyph ✓) ขนาดพอดีช่อง 12pt ที่จุดกึ่งกลาง (x, y)
func (o *Overlay) Check(p *Page, x, y float64) {
	p.content.WriteString(fmt.Sprintf("q 1.3 w 1 J 1 j %s %s m %s %s l %s %s l S Q\n",
		num(x-3.6), num(y+0.2), num(x-1.1), num(y-3.0), num(x+4.0), num(y+4.2)))
}

func num(v float64) string {
	s := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", v), "0"), ".")
	if s == "-0" || s == "" {
		return "0"
	}
	return s
}

// Bytes - เขียน PDF 1.7 หลายหน้า (ขนาดตามแต่ละหน้า) ให้ pdfcpu ใช้เป็นตราประทับทีละหน้า
func (o *Overlay) Bytes() ([]byte, error) {
	var objects [][]byte
	add := func(b []byte) int { objects = append(objects, b); return len(objects) }
	stream := func(dict string, data []byte) []byte {
		var z bytes.Buffer
		w := zlib.NewWriter(&z)
		w.Write(data)
		w.Close()
		return []byte(fmt.Sprintf("<< %s /Filter /FlateDecode /Length %d >>\nstream\n%s\nendstream", dict, z.Len(), z.Bytes()))
	}

	catalog := add(nil)
	pagesObj := add(nil)
	fontFile := add(stream(fmt.Sprintf("/Length1 %d", len(o.shaper.ttf)), o.shaper.ttf))
	descriptor := add([]byte(fmt.Sprintf("<< /Type /FontDescriptor /FontName /BCSarabun /Flags 4 /FontBBox [-437 -401 1287 1097] /ItalicAngle 0 /Ascent 1068 /Descent -232 /CapHeight 700 /StemV 80 /FontFile2 %d 0 R >>", fontFile)))
	gids := make([]int, 0, len(o.widths))
	for g := range o.widths {
		gids = append(gids, int(g))
	}
	sort.Ints(gids)
	var w strings.Builder
	for _, g := range gids {
		w.WriteString(fmt.Sprintf("%d [%s] ", g, num(o.widths[uint16(g)])))
	}
	cid := add([]byte(fmt.Sprintf("<< /Type /Font /Subtype /CIDFontType2 /BaseFont /BCSarabun /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor %d 0 R /CIDToGIDMap /Identity /DW 0 /W [%s] >>", descriptor, w.String())))
	toUnicode := add(stream("", o.toUnicodeCMap(gids)))
	type0 := add([]byte(fmt.Sprintf("<< /Type /Font /Subtype /Type0 /BaseFont /BCSarabun /Encoding /Identity-H /DescendantFonts [%d 0 R] /ToUnicode %d 0 R >>", cid, toUnicode)))

	var kids []string
	for _, p := range o.pages {
		content := add(stream("", p.content.Bytes()))
		page := add([]byte(fmt.Sprintf("<< /Type /Page /Parent %d 0 R /MediaBox [0 0 %s %s] /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>",
			pagesObj, num(p.W), num(p.H), type0, content)))
		kids = append(kids, fmt.Sprintf("%d 0 R", page))
	}
	objects[catalog-1] = []byte(fmt.Sprintf("<< /Type /Catalog /Pages %d 0 R >>", pagesObj))
	objects[pagesObj-1] = []byte(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(kids)))

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, len(objects))
	for i, obj := range objects {
		offsets[i] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, catalog, xref)
	return buf.Bytes(), nil
}

// toUnicodeCMap - glyph → ข้อความ (glyph ที่เป็นเครื่องหมายต่อจาก cluster ไม่มีข้อความของตัวเอง)
func (o *Overlay) toUnicodeCMap(gids []int) []byte {
	var b strings.Builder
	b.WriteString("/CIDInit /ProcSet findresource begin\n12 dict begin\nbegincmap\n/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n/CMapName /Adobe-Identity-UCS def\n/CMapType 2 def\n1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n")
	var entries []string
	for _, g := range gids {
		text, ok := o.unicode[uint16(g)]
		if !ok {
			continue
		}
		var hex strings.Builder
		for _, u := range utf16Units(text) {
			fmt.Fprintf(&hex, "%04X", u)
		}
		entries = append(entries, fmt.Sprintf("<%04X> <%s>", g, hex.String()))
	}
	for len(entries) > 0 {
		chunk := entries[:min(100, len(entries))]
		entries = entries[len(chunk):]
		fmt.Fprintf(&b, "%d beginbfchar\n%s\nendbfchar\n", len(chunk), strings.Join(chunk, "\n"))
	}
	b.WriteString("endcmap\nCMapName currentdict /CMap defineresource pop\nend\nend\n")
	return []byte(b.String())
}

func utf16Units(s string) []uint16 {
	var out []uint16
	for _, r := range s {
		if r >= 0x10000 {
			r -= 0x10000
			out = append(out, uint16(0xD800+(r>>10)), uint16(0xDC00+(r&0x3FF)))
			continue
		}
		out = append(out, uint16(r))
	}
	return out
}

// BlankForm - ถอดช่องกรอก (AcroForm) ออกจากแบบฟอร์มทางการ ให้ข้อความของเราเป็นตัวจริงตัวเดียว
// (ไม่มีช่องว่างของโปรแกรมอ่าน PDF ทับข้อความ และพิมพ์ออกมาเหมือนกันทุกเครื่อง)
func BlankForm(pdf []byte) ([]byte, error) {
	fields, err := api.FormFields(bytes.NewReader(pdf), Config())
	if err != nil {
		return nil, fmt.Errorf("read form fields: %w", err)
	}
	if len(fields) == 0 {
		return pdf, nil
	}
	ids := make([]string, 0, len(fields))
	for _, f := range fields {
		ids = append(ids, f.ID)
	}
	var out bytes.Buffer
	if err := api.RemoveFormFields(bytes.NewReader(pdf), &out, ids, Config()); err != nil {
		return nil, fmt.Errorf("remove form fields: %w", err)
	}
	return out.Bytes(), nil
}

// Stamp - ประทับชั้นข้อความหน้า i ลงหน้า i ของ base (base ต้องมีจำนวนหน้าเท่ากับ overlay)
func Stamp(base []byte, o *Overlay) ([]byte, error) {
	overlayPDF, err := o.Bytes()
	if err != nil {
		return nil, err
	}
	wm, err := api.PDFWatermarkForReadSeeker(bytes.NewReader(overlayPDF), 0, "scalefactor:1 abs, pos:bl, off:0 0, rot:0", true, false, types.POINTS)
	if err != nil {
		return nil, fmt.Errorf("load text layer: %w", err)
	}
	var out bytes.Buffer
	if err := api.AddWatermarks(bytes.NewReader(base), &out, nil, wm, Config()); err != nil {
		return nil, fmt.Errorf("stamp text layer: %w", err)
	}
	return out.Bytes(), nil
}

// Merge - ต่อ PDF หลายไฟล์ตามลำดับ (หน้าซ้ำได้ — ส่งไฟล์เดียวกันหลายครั้ง)
func Merge(parts ...[]byte) ([]byte, error) {
	if len(parts) == 1 {
		return parts[0], nil
	}
	readers := make([]io.ReadSeeker, len(parts))
	for i, p := range parts {
		readers[i] = bytes.NewReader(p)
	}
	var out bytes.Buffer
	if err := api.MergeRaw(readers, &out, false, Config()); err != nil {
		return nil, fmt.Errorf("merge pages: %w", err)
	}
	return out.Bytes(), nil
}
