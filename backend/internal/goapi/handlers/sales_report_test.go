package handlers

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

// TestSalesMoneyIsExactDecimalText - Rule 20: 0.1 + 0.2 ต้องเป็น "0.30" พอดี (float64 ได้ 0.30000000000000004)
func TestSalesMoneyIsExactDecimalText(t *testing.T) {
	amount := decimal.RequireFromString("0.1").Add(decimal.RequireFromString("0.2"))
	data := salesSummaryData(2, amount, decimal.RequireFromString("0.1"), decimal.RequireFromString("0.2"))
	if data["totalamount"] != "0.30" || data["totalcost"] != "0.10" || data["total_profit"] != "0.20" || data["profit_margin"] != "66.67" {
		t.Fatalf("summary = %v", data)
	}

	row := map[string]any{"docno": "IV6909-001", "totalqty": "3.00000000", "totalamount": "0.30000000", "calcamount": "0.10000000", "grossprofit": "0.20000000", "price": "0.10000000", "averagecost": nil}
	if err := formatSalesReportRow(row); err != nil {
		t.Fatal(err)
	}
	if row["totalamount"] != "0.30" || row["calcamount"] != "0.10" || row["grossprofit"] != "0.20" || row["price"] != "0.10" || row["marginpercent"] != "66.67" {
		t.Fatalf("row = %v", row)
	}
	if row["totalqty"] != "3.00000000" || row["averagecost"] != nil {
		t.Fatalf("non-money/NULL columns must stay untouched: %v", row)
	}
}

// TestSalesMoneyRoundingBoundary - ยอดเงินไม่ถูกปัดเลย (อย่างน้อย 2 ตำแหน่ง คงความละเอียดที่บันทึกไว้ เหมือนรายงานอายุหนี้);
// ปัดครั้งเดียวที่ % กำไรขั้นต้น 2 ตำแหน่งแบบครึ่งปัดออกจากศูนย์ (ไม่ใช่ banker's rounding)
func TestSalesMoneyRoundingBoundary(t *testing.T) {
	cases := map[string]string{
		"0.005":       "0.005",
		"0.00500000":  "0.005",
		"0.00499999":  "0.00499999",
		"-0.005":      "-0.005",
		"1.125":       "1.125",
		"2.13500000":  "2.135",
		"53500.00000": "53500.00",
		"0.10000000":  "0.10",
	}
	for in, want := range cases {
		row := map[string]any{"totalamount": in}
		if err := formatSalesReportRow(row); err != nil {
			t.Fatal(err)
		}
		if row["totalamount"] != want {
			t.Errorf("%s -> %v, want %s", in, row["totalamount"], want)
		}
	}

	margins := []struct{ amount, profit, want string }{
		{"200", "0.01", "0.01"},       // 0.005% → ขึ้น
		{"200", "0.00999999", "0.00"}, // 0.004999995% → ลง
		{"3", "1", "33.33"},
		{"1000", "300", "30.00"},
		{"0", "50", "0.00"}, // ไม่มียอดขาย → ไม่มีฐานคิดอัตรา
		{"1000", "-300", "-30.00"},
	}
	for _, m := range margins {
		got := moneyText(salesMarginPercent(decimal.RequireFromString(m.amount), decimal.RequireFromString(m.profit)))
		if got != m.want {
			t.Errorf("margin(%s, %s) = %s, want %s", m.amount, m.profit, got, m.want)
		}
	}
}

// TestSalesReportJSONRoundTripKeepsStrings - ยอดเงินออกเป็น JSON string และอ่านกลับได้ค่าเดิม (ไม่กลายเป็น number)
func TestSalesReportJSONRoundTripKeepsStrings(t *testing.T) {
	row := map[string]any{"totalamount": "53500.00000000", "calcamount": "32000.00000000", "grossprofit": "21500.00000000"}
	if err := formatSalesReportRow(row); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{"data": []any{row}, "summary": salesSummaryData(1, decimal.RequireFromString("53500"), decimal.RequireFromString("32000"), decimal.RequireFromString("21500"))})
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Data    []map[string]any `json:"data"`
		Summary map[string]any   `json:"summary"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.Data[0]
	for key, want := range map[string]string{"totalamount": "53500.00", "calcamount": "32000.00", "grossprofit": "21500.00", "marginpercent": "40.19"} {
		if text, ok := got[key].(string); !ok || text != want {
			t.Errorf("row %s = %#v, want string %q", key, got[key], want)
		}
	}
	for key, want := range map[string]string{"totalamount": "53500.00", "totalcost": "32000.00", "total_profit": "21500.00", "profit_margin": "40.19"} {
		if text, ok := decoded.Summary[key].(string); !ok || text != want {
			t.Errorf("summary %s = %#v, want string %q", key, decoded.Summary[key], want)
		}
	}
}

// TestFormatSalesReportRowRejectsNonDecimal - ค่าที่ไม่ใช่ข้อความทศนิยมจากฐาน ต้องไม่ถูกเดาเป็นตัวเลข
func TestFormatSalesReportRowRejectsNonDecimal(t *testing.T) {
	for _, bad := range []any{53500.5, "abc"} {
		if err := formatSalesReportRow(map[string]any{"totalamount": bad}); err == nil {
			t.Errorf("value %#v accepted", bad)
		}
	}
}

// TestSalesRowsReconcileAfterFormatting - ทุกแถวที่ส่งออก ยอดขาย − ต้นทุน = กำไรขั้นต้น พอดี และยอดรวมท้ายรายงานกระทบกัน
// (เดิมปัดแต่ละคอลัมน์ 2 ตำแหน่งเอง: 1.00 / 0.125 / 0.875 → 1.00 / 0.13 / 0.88 แต่ 1.00 − 0.13 = 0.87)
func TestSalesRowsReconcileAfterFormatting(t *testing.T) {
	rows := []map[string]any{
		{"totalamount": "1.00000000", "calcamount": "0.12500000", "grossprofit": "0.87500000"},
		{"totalamount": "0.30000000", "calcamount": "0.10000000", "grossprofit": "0.20000000"},
		{"totalamount": "107.00000000", "calcamount": "33.33333333", "grossprofit": "73.66666667"},
	}
	var sumAmount, sumCost, sumProfit decimal.Decimal
	for _, row := range rows {
		if err := formatSalesReportRow(row); err != nil {
			t.Fatal(err)
		}
		amount := decimal.RequireFromString(row["totalamount"].(string))
		cost := decimal.RequireFromString(row["calcamount"].(string))
		profit := decimal.RequireFromString(row["grossprofit"].(string))
		if !amount.Sub(cost).Equal(profit) {
			t.Errorf("row %v: %s - %s != %s", row, amount, cost, profit)
		}
		sumAmount, sumCost, sumProfit = sumAmount.Add(amount), sumCost.Add(cost), sumProfit.Add(profit)
	}
	if rows[0]["totalamount"] != "1.00" || rows[0]["calcamount"] != "0.125" || rows[0]["grossprofit"] != "0.875" || rows[0]["marginpercent"] != "87.50" {
		t.Errorf("row 0 = %v", rows[0])
	}
	if !sumAmount.Sub(sumCost).Equal(sumProfit) {
		t.Errorf("totals: %s - %s != %s", sumAmount, sumCost, sumProfit)
	}
}
