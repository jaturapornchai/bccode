package pdftext

import "testing"

// TestThaiShaping - สระบน + วรรณยุกต์ต้องไม่ทับกัน: "ปั้ม" ต้องได้ glyph ครบและ cluster แรกมีข้อความสำหรับ ToUnicode
func TestThaiShaping(t *testing.T) {
	s, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	glyphs, width := s.shape("ปั้ม", 11)
	if len(glyphs) < 4 || width <= 0 {
		t.Fatalf("glyphs=%d width=%v", len(glyphs), width)
	}
	if glyphs[0].text == "" {
		t.Fatal("first glyph missing cluster text for ToUnicode")
	}
}
