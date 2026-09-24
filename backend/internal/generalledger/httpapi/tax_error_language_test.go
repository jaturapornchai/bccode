package httpapi

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
)

// UAT V13/S21 2026-09-24: ข้อผิดพลาดของรายละเอียดภาษี/กลับรายการ/กระทบยอด ต้องแปลตาม Accept-Language และบอกช่องที่ผิด
// (เดิมเป็น fmt.Errorf ไทยล้วน → invalid_request ภาษาไทยแม้เลือกภาษาอังกฤษ)
func TestGLTaxDetailErrorsTranslateAndKeepField(t *testing.T) {
	err := &gl.UserError{Code: "vat_period_required", Field: "tax_period_month", Message: "ภาษีขายต้องระบุงวดภาษี"}
	status, en := errorPayloadFor(err, "en")
	if status != 409 || en.Code != "vat_period_required" || en.Message != "Output VAT needs a tax period." || en.MessageTH != "ภาษีขายต้องระบุงวดภาษี" || en.Field != "tax_period_month" {
		t.Fatalf("en payload = %d %+v", status, en)
	}
	if _, th := errorPayloadFor(err, "th"); th.Message != "ภาษีขายต้องระบุงวดภาษี" || th.Field != "tax_period_month" {
		t.Fatalf("th payload = %+v", th)
	}
}

// ทุก code ที่ generalledger คืนผ่าน userError/fieldError ต้องมีแถว gl_err_<code> ภาษาอังกฤษจริงใน languages.tsv
// (แถวขาด = จอภาษาอังกฤษค้างเป็นไทย โดยไม่มีใครเห็น) — สแกนจากซอร์สจริง ไม่ใช่รายการที่เขียนมือ
func TestGLUserErrorCodesHaveEnglishRows(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "*.go"))
	if err != nil || len(files) == 0 {
		t.Fatalf("glob generalledger sources: %v", err)
	}
	pattern := regexp.MustCompile(`\b(?:userError|fieldError)\("([a-z0-9_]+)"`)
	seen := map[string]bool{}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range pattern.FindAllStringSubmatch(string(raw), -1) {
			seen[m[1]] = true
		}
	}
	for _, code := range []string{"vat_rate_invalid", "wht_payment_date_invalid", "reverse_reason_required", "reconcile_requires_posted"} {
		if !seen[code] {
			t.Fatalf("scanner missed %s — pattern out of date", code)
		}
	}
	for code := range seen {
		key := "gl_err_" + code
		if text := language.Text(key, "en"); text == key || text == "" {
			t.Errorf("missing English row %s in languages.tsv", key)
		}
	}
}
