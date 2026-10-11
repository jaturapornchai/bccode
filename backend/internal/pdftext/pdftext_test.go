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

// TestLineContentStream - เส้นตรงต้องออกเป็นคำสั่ง PDF ตรงตัว (ความหนา, จุดเริ่ม, จุดจบ) และปัดทศนิยม 3 ตำแหน่ง
func TestLineContentStream(t *testing.T) {
	s, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	o := NewOverlay(s)
	p := o.NewPage(504.57, 252.28)
	o.Line(p, 10, 20.5, 30.1234, 40, 0.75)
	want := "q 0.75 w 0 J 10 20.5 m 30.123 40 l S Q\n"
	if got := p.content.String(); got != want {
		t.Fatalf("content stream\n got: %q\nwant: %q", got, want)
	}
	out, err := o.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 || string(out[:5]) != "%PDF-" {
		t.Fatalf("overlay with a line must still serialise to a PDF, got %q", out[:min(8, len(out))])
	}
}
