// Package rdform กรอกแบบฟอร์มกรมสรรพากร (ภ.ง.ด./ภ.พ./ภ.ธ.) ลงบน PDF ทางการ.
//
// ต้นฉบับแบบ = PDF กรอกได้ของกรมสรรพากร (mydocs/sample) ที่ tools/rdform/rdform_build.py คัดลอกมาไว้ใน
// assets/ พร้อมสเปกพิกัดทุกช่องใน specs/ (generated — ห้ามแก้มือ แก้ที่ tools/rdform/specs แล้ว build ใหม่).
// ตัวกรอกไม่มีตรรกะภาษี: รับค่าที่คำนวณแล้ว (สตริง) แล้ววาดลงช่อง — การคำนวณอยู่ที่ผู้เรียก (handlers).
package rdform

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

//go:embed assets/*.pdf
var assetFS embed.FS

//go:embed specs/*.json
var specFS embed.FS

// Box - ช่องบนแบบ: rect = [x0 y0 x1 y1] pt (มุมล่างซ้ายของหน้า), align = Q ของ AcroForm (0 ซ้าย 1 กลาง 2 ขวา)
type Box struct {
	Page  int          `json:"page"`
	Rect  [4]float64   `json:"rect"`
	Align int          `json:"align"`
	Cells [][2]float64 `json:"cells,omitempty"` // ช่องตัวเลขทีละหลักที่ตรวจเจอจากเส้นบนแบบ
	Comb  int          `json:"comb,omitempty"`  // ไม่เจอเส้น: แบ่งกว้างเท่ากัน N ช่อง
	Split float64      `json:"split,omitempty"` // เส้นแบ่ง บาท|สตางค์
	// BahtEnd - ขอบขวาช่องบาทเมื่อช่องบาทกับช่องสตางค์แยกกันด้วยขีด "–" (ไม่มี = Split)
	BahtEnd float64 `json:"baht_end,omitempty"`
	// Satang - ช่องเงินแบบทีละหลัก: Cells = หลักบาท + Satang ช่องท้ายเป็นสตางค์
	Satang int `json:"satang,omitempty"`
}

type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Box
}

type Field struct {
	Key     string   `json:"key"`
	Type    string   `json:"type"`
	Label   string   `json:"label"`
	Group   string   `json:"group,omitempty"`
	Note    string   `json:"note,omitempty"`
	Options []Option `json:"options,omitempty"`
	Box
}

type Column struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Label string `json:"label"`
	Note  string `json:"note,omitempty"`
}

type Table struct {
	Key     string           `json:"key"`
	Label   string           `json:"label"`
	Columns []Column         `json:"columns"`
	Rows    []map[string]Box `json:"rows"`
}

type Spec struct {
	Code     string       `json:"code"`
	Title    string       `json:"title"`
	Template string       `json:"template"`
	Pages    []int        `json:"pages"`
	Sizes    [][2]float64 `json:"sizes"`
	Fields   []Field      `json:"fields"`
	Table    *Table       `json:"table,omitempty"`
}

// MoneyPair - ยอดเงินหนึ่งยอดที่แบบพิมพ์ช่องบาทกับช่องสตางค์แยกกัน (key_baht + key_satang)
type MoneyPair struct {
	Key          string
	Baht, Satang *Field
}

// MoneyPairs - คู่ช่อง X_baht/X_satang ทั้งหมดของแบบ ตามลำดับบนแบบ
func (s *Spec) MoneyPairs() []MoneyPair {
	var out []MoneyPair
	for i := range s.Fields {
		key, ok := strings.CutSuffix(s.Fields[i].Key, "_baht")
		if !ok {
			continue
		}
		if satang := s.Field(key + "_satang"); satang != nil && s.Field(key) == nil {
			out = append(out, MoneyPair{Key: key, Baht: &s.Fields[i], Satang: satang})
		}
	}
	return out
}

// DigitGroup - เลขหลายหลักที่แบบพิมพ์เป็นช่องแยกทีละหลักคนละช่องกรอก (เช่น สาขาที่ของใบแนบ ภ.พ.30 =
// branch_d1..branch_d5): จอกรอกเห็นเป็นช่องเดียว Key แล้วตัวกรอกแยกหลักให้
type DigitGroup struct {
	Key   string
	Parts []Column
}

// DigitGroups - กลุ่มคอลัมน์ X_d1..X_dN ของตารางใบแนบ
func (t *Table) DigitGroups() []DigitGroup {
	var out []DigitGroup
	index := map[string]int{}
	for _, c := range t.Columns {
		i := strings.LastIndex(c.Key, "_d")
		if i <= 0 || len(c.Key) != i+3 || c.Key[i+2] < '1' || c.Key[i+2] > '9' {
			continue
		}
		key := c.Key[:i]
		if _, ok := index[key]; !ok {
			index[key] = len(out)
			out = append(out, DigitGroup{Key: key})
		}
		out[index[key]].Parts = append(out[index[key]].Parts, c)
	}
	return out
}

// Field - หาช่องตาม key (nil ถ้าไม่มี)
func (s *Spec) Field(key string) *Field {
	for i := range s.Fields {
		if s.Fields[i].Key == key {
			return &s.Fields[i]
		}
	}
	return nil
}

var (
	loadOnce sync.Once
	loadErr  error
	specs    map[string]*Spec
)

func load() error {
	loadOnce.Do(func() {
		entries, err := specFS.ReadDir("specs")
		if err != nil {
			loadErr = err
			return
		}
		specs = make(map[string]*Spec, len(entries))
		for _, e := range entries {
			raw, err := specFS.ReadFile("specs/" + e.Name())
			if err != nil {
				loadErr = err
				return
			}
			var s Spec
			if err := json.Unmarshal(raw, &s); err != nil {
				loadErr = fmt.Errorf("rdform spec %s: %w", e.Name(), err)
				return
			}
			if len(s.Sizes) != len(s.Pages) {
				loadErr = fmt.Errorf("rdform spec %s: %d sizes for %d pages", s.Code, len(s.Sizes), len(s.Pages))
				return
			}
			specs[s.Code] = &s
		}
	})
	return loadErr
}

// Lookup - สเปกตามรหัส (pnd53, pnd53_attach, ...)
func Lookup(code string) (*Spec, error) {
	if err := load(); err != nil {
		return nil, err
	}
	s, ok := specs[code]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownForm, code)
	}
	return s, nil
}

// Forms - รหัสแบบหลักทั้งหมด (ไม่รวมใบแนบ) เรียงตามรหัส
func Forms() ([]string, error) {
	if err := load(); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(specs))
	for code := range specs {
		if !strings.HasSuffix(code, attachSuffix) {
			out = append(out, code)
		}
	}
	sort.Strings(out)
	return out, nil
}

const attachSuffix = "_attach"

// Attachment - ใบแนบของแบบ (nil ถ้าแบบนี้ไม่มีใบแนบ)
func Attachment(code string) *Spec {
	if load() != nil {
		return nil
	}
	return specs[code+attachSuffix]
}

func template(s *Spec) ([]byte, error) {
	return assetFS.ReadFile("assets/" + s.Template)
}
