package generalledger

import (
	"errors"
	"testing"
)

func TestDocNoPrefixChampFormat(t *testing.T) {
	cases := []struct {
		format, book, date, prefix string
		digits                     int
	}{
		{"", "JV", "2026-10-07", "JV6910-", 4},
		{"YYMM-@#####", "PV", "2026-10-07", "6910-PV", 5},
		{"@@-yymmdd###", "SV", "2026-01-05", "SV-690105", 3},
		{"GL#", "JV", "2026-12-31", "GL", 1},
		{"@YYMM-##-#", "JV", "2026-10-07", "JV6910-", 3},
	}
	for _, c := range cases {
		prefix, digits, err := docNoPrefix(c.format, c.book, c.date)
		if err != nil || prefix != c.prefix || digits != c.digits {
			t.Errorf("docNoPrefix(%q,%q,%q) = %q,%d,%v; want %q,%d", c.format, c.book, c.date, prefix, digits, err, c.prefix, c.digits)
		}
	}
}

func TestDocNoPrefixRejectsBadInput(t *testing.T) {
	if _, _, err := docNoPrefix("JV-YYMM", "JV", "2026-10-07"); !errors.Is(err, errDocFormatInvalid) {
		t.Fatalf("format without # must fail, got %v", err)
	}
	if _, _, err := docNoPrefix("", "JV", "2026-13-01"); err == nil {
		t.Fatal("invalid voucher date must fail")
	}
	if err := validateDocFormat("@YYMMDD-#############"); !errors.Is(err, errDocFormatInvalid) {
		t.Fatalf("format over 20 characters must fail, got %v", err)
	}
	if err := validateDocFormat(""); err != nil {
		t.Fatalf("blank format means the default, got %v", err)
	}
}
