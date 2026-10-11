package report

import "testing"

func TestChartRowsFollowChampLayout(t *testing.T) {
	records := []masterRecord{
		{code: "10000", name1: "สินทรัพย์", name2: "Assets", group: "asset", level: "1"},
		{code: "11110", name1: "เงินสด", name2: "Cash", group: "asset", level: "4", posting: true},
		{code: "21110", name1: "เจ้าหนี้การค้า", name2: "เจ้าหนี้การค้า", group: "liability", level: "4", posting: true},
	}
	out := chartRows(records, "both", "th")
	if len(out.Rows) != 5 {
		t.Fatalf("rows = %v", out.Rows)
	}
	first, second, closeAsset := out.Rows[0], out.Rows[1], out.Rows[2]
	if first["accountgroup"] != "สินทรัพย์" || first[BoldKey] != "true" || first["accountname"] != "สินทรัพย์ / Assets" {
		t.Fatalf("first row = %v", first)
	}
	if second["accountgroup"] != "" || second[BoldKey] != "" || second[KindKey] != KindDetail {
		t.Fatalf("second row = %v", second)
	}
	if closeAsset[KindKey] != KindGroupTotal || closeAsset[LabelKey] != "รวมสินทรัพย์ 2 รายการ" {
		t.Fatalf("group close = %v", closeAsset)
	}
	if out.Rows[3]["accountname"] != "เจ้าหนี้การค้า" {
		t.Fatalf("same names print once: %v", out.Rows[3])
	}
	if out.Totals["count"] != "3" || out.Totals["header_count"] != "1" || out.Totals["detail_count"] != "2" {
		t.Fatalf("totals = %v", out.Totals)
	}
	if len(out.Footer) != 2 || out.Footer[0].LabelKey != "report_master_control_count" || out.Footer[1].Args[0] != "2" {
		t.Fatalf("footer = %+v", out.Footer)
	}
	if got := chartRows(records, "name2", "th").Rows[0]["accountname"]; got != "Assets" {
		t.Fatalf("name2 = %s", got)
	}
}

func TestRegisterRowsAndPaging(t *testing.T) {
	out := registerRows([]masterRecord{{code: "JV", name1: "สมุดรายวันทั่วไป", title: "ใบสำคัญทั่วไป"}, {code: "SV", name1: "สมุดรายวันขาย"}}, true, "th")
	if len(out.Columns) != 3 || out.Rows[0]["titlename"] != "ใบสำคัญทั่วไป" || out.Totals["count"] != "2" || out.Footer[0].Args[0] != "2" {
		t.Fatalf("register = %+v", out)
	}
	if len(registerRows(nil, false, "th").Columns) != 2 {
		t.Fatal("only journal books print the voucher title")
	}
	rows := []map[string]string{{"code": "1"}, {"code": "2"}, {"code": "3"}}
	if got := pageSlice(rows, 2, 2); len(got) != 1 || got[0]["code"] != "3" {
		t.Fatalf("page 2 = %v", got)
	}
	if got := pageSlice(rows, 3, 2); len(got) != 0 {
		t.Fatalf("past the end = %v", got)
	}
}

func TestGroupRange(t *testing.T) {
	if got, failure := groupRange("", "", true); failure != nil || len(got) != 5 {
		t.Fatalf("default range = %v %v", got, failure)
	}
	if got, failure := groupRange("liability", "equity", true); failure != nil || len(got) != 2 || got[0] != "liability" {
		t.Fatalf("range = %v %v", got, failure)
	}
	_, failure := groupRange("expense", "asset", true)
	wantFailure(t, failure, "REPORT_PARAM_INVALID", "group_from")
	if got, _ := groupRange("expense", "asset", false); len(got) != 0 {
		t.Fatal("registers have no group filter")
	}
}
