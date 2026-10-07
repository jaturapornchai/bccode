package generalledger

import "testing"

// ปุ่ม "ภาษีแยกนอก" / "ภาษีรวมใน" ของกล่องภาษีบนใบสำคัญ (Champ OnTaxAuto) — คำนวณที่ backend ด้วย decimal เท่านั้น
func TestVatCalcModes(t *testing.T) {
	cases := []struct {
		name, mode, amount, rate string
		scale                    int
		base, vat, total         string
	}{
		{"exclusive 100 @7", "exclusive", "100", "7", 2, "100.00", "7.00", "107.00"},
		{"exclusive rounds half up like save", "exclusive", "0.50", "7", 2, "0.50", "0.04", "0.54"}, // 0.035 → 0.04
		{"exclusive rate with decimals", "exclusive", "1000", "7.25", 2, "1000.00", "72.50", "1072.50"},
		{"inclusive 107.00 @7", "inclusive", "107.00", "7", 2, "100.00", "7.00", "107.00"},
		{"inclusive 100.01 @7 adds back exactly", "inclusive", "100.01", "7", 2, "93.47", "6.54", "100.01"},
		{"inclusive 1.07 @7", "inclusive", "1.07", "7", 2, "1.00", "0.07", "1.07"},
		{"inclusive 0.10 @7 boundary", "inclusive", "0.10", "7", 2, "0.09", "0.01", "0.10"}, // 0.0934… → 0.09
		{"rate 0 exclusive", "exclusive", "250.00", "0", 2, "250.00", "0.00", "250.00"},
		{"rate 0 inclusive", "inclusive", "250.00", "0", 2, "250.00", "0.00", "250.00"},
		{"zero amount", "exclusive", "0", "7", 2, "0.00", "0.00", "0.00"},
		{"fiscal year scale 0", "inclusive", "107", "7", 0, "100", "7", "107"},
		{"0.1 + 0.2 style input", "exclusive", "0.30", "10", 2, "0.30", "0.03", "0.33"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := VatCalc(c.mode, c.amount, c.rate, c.scale)
			if err != nil {
				t.Fatalf("VatCalc error: %v", err)
			}
			if string(got.BaseAmount) != c.base || string(got.VatAmount) != c.vat || string(got.TotalAmount) != c.total {
				t.Fatalf("got base=%s vat=%s total=%s; want %s/%s/%s", got.BaseAmount, got.VatAmount, got.TotalAmount, c.base, c.vat, c.total)
			}
			if !got.BaseAmount.Decimal().Add(got.VatAmount.Decimal()).Equal(got.TotalAmount.Decimal()) {
				t.Fatalf("base + vat != total: %s + %s != %s", got.BaseAmount, got.VatAmount, got.TotalAmount)
			}
		})
	}
}

// ผลแยกนอกต้องตรงกับภาษีที่ระบบคำนวณเองตอนบันทึก (ภาษีว่าง = ฐาน × อัตรา ปัดตามทศนิยมปี) ไม่ให้ตัวเลขในกล่องภาษีกับที่บันทึกต่างกัน
func TestVatCalcExclusiveMatchesSaveRounding(t *testing.T) {
	for _, base := range []string{"0.07", "0.14", "1.50", "33.33", "99.99", "12345.67"} {
		got, err := VatCalc("exclusive", base, "7", 2)
		if err != nil {
			t.Fatal(err)
		}
		want := Amount(base).Decimal().Mul(Amount("7").Decimal()).Div(hundred).Round(2)
		if !got.VatAmount.Decimal().Equal(want) {
			t.Fatalf("base %s: calc vat %s, save rounding %s", base, got.VatAmount, want)
		}
	}
}

func TestVatCalcRejectsBadInput(t *testing.T) {
	cases := []struct {
		name, mode, amount, rate, code, field string
		scale                                 int
	}{
		{"unknown mode", "gross", "100", "7", "vat_calc_mode_invalid", "mode", 2},
		{"empty mode", "", "100", "7", "vat_calc_mode_invalid", "mode", 2},
		{"text amount", "exclusive", "abc", "7", "vat_calc_amount_invalid", "amount", 2},
		{"empty amount", "exclusive", "", "7", "vat_calc_amount_invalid", "amount", 2},
		{"negative amount", "inclusive", "-1", "7", "vat_calc_amount_invalid", "amount", 2},
		{"thousand separator", "exclusive", "1,000", "7", "vat_calc_amount_invalid", "amount", 2},
		{"more decimals than year", "exclusive", "1.234", "7", "vat_calc_amount_invalid", "amount", 2},
		{"too large", "exclusive", "100000000000000", "7", "vat_calc_amount_invalid", "amount", 2},
		{"total too large", "exclusive", "99999999999999", "7", "vat_calc_amount_invalid", "amount", 2},
		{"empty rate (no default 7%)", "exclusive", "100", "", "vat_rate_invalid", "vat_rate", 2},
		{"rate over 100", "exclusive", "100", "100.01", "vat_rate_invalid", "vat_rate", 2},
		{"negative rate", "exclusive", "100", "-7", "vat_rate_invalid", "vat_rate", 2},
		{"rate 3 decimals", "exclusive", "100", "7.001", "vat_rate_invalid", "vat_rate", 2},
		{"from tax rate 0 (divide by rate)", "from_tax", "7.00", "0", "vat_calc_rate_zero", "vat_rate", 2},
		{"from tax rate 0.00", "from_tax", "7.00", "0.00", "vat_calc_rate_zero", "vat_rate", 2},
		{"from tax negative rate", "from_tax", "7.00", "-7", "vat_rate_invalid", "vat_rate", 2},
		{"from tax empty rate (no default 7%)", "from_tax", "7.00", "", "vat_rate_invalid", "vat_rate", 2},
		{"from tax rate over 100", "from_tax", "7.00", "100.01", "vat_rate_invalid", "vat_rate", 2},
		{"from tax negative tax", "from_tax", "-7", "7", "vat_calc_amount_invalid", "amount", 2},
		{"from tax more decimals than year", "from_tax", "0.001", "7", "vat_calc_amount_invalid", "amount", 2},
		{"from tax base too large", "from_tax", "99999999999.99", "0.01", "vat_calc_amount_invalid", "amount", 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := VatCalc(c.mode, c.amount, c.rate, c.scale)
			user, ok := AsUserError(err)
			if !ok || user.Code != c.code || user.Field != c.field {
				t.Fatalf("err = %v; want code %s field %s", err, c.code, c.field)
			}
		})
	}
}

// กรอกภาษีแต่จำนวนเงินว่าง (Champ OnTaxAuto glfrmjournal.cpp:1691-1701): ฐาน = ภาษี × 100 ÷ อัตรา ปัดครึ่งขึ้น, ภาษีคงตามที่กรอก
func TestVatCalcFromTax(t *testing.T) {
	cases := []struct {
		name, amount, rate string
		scale              int
		base, vat, total   string
	}{
		{"7.00 @7", "7.00", "7", 2, "100.00", "7.00", "107.00"},
		{"integer input", "7", "7", 2, "100.00", "7.00", "107.00"},
		{"one satang @7", "0.01", "7", 2, "0.14", "0.01", "0.15"},      // 0.142857… → 0.14
		{"10.70 @7", "10.70", "7", 2, "152.86", "10.70", "163.56"},     // 152.857142… → 152.86
		{"tie rounds half up", "0.01", "8", 2, "0.13", "0.01", "0.14"}, // 0.125 → 0.13
		{"rate with decimals", "72.50", "7.25", 2, "1000.00", "72.50", "1072.50"},
		{"rate 100", "5.55", "100", 2, "5.55", "5.55", "11.10"},
		{"zero tax", "0", "7", 2, "0.00", "0.00", "0.00"},
		{"fiscal year scale 0", "7", "7", 0, "100", "7", "107"},
		{"scale 0 tie rounds half up", "1", "8", 0, "13", "1", "14"}, // 12.5 → 13
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := VatCalc("from_tax", c.amount, c.rate, c.scale)
			if err != nil {
				t.Fatalf("VatCalc error: %v", err)
			}
			if string(got.BaseAmount) != c.base || string(got.VatAmount) != c.vat || string(got.TotalAmount) != c.total {
				t.Fatalf("got base=%s vat=%s total=%s; want %s/%s/%s", got.BaseAmount, got.VatAmount, got.TotalAmount, c.base, c.vat, c.total)
			}
			if !got.BaseAmount.Decimal().Add(got.VatAmount.Decimal()).Equal(got.TotalAmount.Decimal()) {
				t.Fatalf("base + vat != total: %s + %s != %s", got.BaseAmount, got.VatAmount, got.TotalAmount)
			}
		})
	}
}

// ฐานที่หาได้จากภาษีต้องคำนวณภาษีแยกนอกกลับได้ยอดภาษีเดิม (อัตรา < 100% คลาดไม่ถึงครึ่งสตางค์) — กดปุ่มซ้ำแล้วภาษีที่กรอกไม่เปลี่ยน
func TestVatCalcFromTaxRoundTripsExclusive(t *testing.T) {
	for _, rate := range []string{"7", "7.25", "10", "3", "0.5"} {
		for _, tax := range []string{"0.01", "0.05", "0.99", "10.70", "33.33", "700.00", "12345.67"} {
			derived, err := VatCalc("from_tax", tax, rate, 2)
			if err != nil {
				t.Fatal(err)
			}
			back, err := VatCalc("exclusive", string(derived.BaseAmount), rate, 2)
			if err != nil {
				t.Fatal(err)
			}
			if !back.VatAmount.Decimal().Equal(Amount(tax).Decimal()) {
				t.Fatalf("rate %s tax %s: base %s gives vat %s", rate, tax, derived.BaseAmount, back.VatAmount)
			}
		}
	}
}
