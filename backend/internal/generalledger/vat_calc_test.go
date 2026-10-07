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
		scale                                int
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
