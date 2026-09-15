package stockengine

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func day(d int) time.Time {
	return time.Date(2026, 3, d, 0, 0, 0, 0, time.UTC)
}

func mv(docNo string, line int, transFlag int, qty, sumAmount float64, t time.Time) Movement {
	return Movement{
		WhCode:      "WH01",
		DocDateTime: t,
		DocNo:       docNo,
		LineNumber:  line,
		TransFlag:   transFlag,
		Qty:         qty,
		UnitStand:   1,
		UnitDivide:  1,
		SumAmount:   sumAmount,
	}
}

func approx(t *testing.T, got, want float64, label string) {
	t.Helper()
	if math.Abs(got-want) > 0.005 {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func TestWeightedAverageBasicFlow(t *testing.T) {
	movements := []Movement{
		mv("PU001", 1, 12, 10, 1000, day(1)), // ซื้อ 10 @100
		mv("PU002", 1, 12, 10, 1400, day(2)), // ซื้อ 10 @140 → เฉลี่ย 120
		mv("SA001", 1, 44, 5, 0, day(3)),     // ขาย 5 ใช้ต้นทุน 120
	}
	res := Calculate(nil, movements, DefaultOptions())

	if len(res.Rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(res.Rows))
	}
	approx(t, res.Rows[0].AvgCost, 100, "avg after first buy")
	approx(t, res.Rows[1].AvgCost, 120, "avg after second buy")
	approx(t, res.Rows[2].UnitCost, 120, "sale unit cost")
	approx(t, res.Rows[2].Amount, 600, "sale amount")
	approx(t, res.Rows[2].BalanceQty, 15, "balance qty")
	approx(t, res.Rows[2].BalanceAmount, 1800, "balance amount")
	approx(t, res.Rows[2].AvgCost, 120, "avg after sale stays")
}

// ทิศทางของทุกประเภทเอกสารต้องตรงกับที่ระบบประกาศไว้ใน myglobal/global.go:28-29
func TestTransFlagDirectionsMatchSystemDefinition(t *testing.T) {
	in := []int{54, 12, 310, 48, 60, 58, 66}
	out := []int{44, 16, 56, 68, 72}

	for _, flag := range in {
		if dir, ok := DirectionOf(flag); !ok || dir != DirectionIn {
			t.Errorf("transflag %d should be inbound, got dir=%d ok=%v", flag, dir, ok)
		}
	}
	for _, flag := range out {
		if dir, ok := DirectionOf(flag); !ok || dir != DirectionOut {
			t.Errorf("transflag %d should be outbound, got dir=%d ok=%v", flag, dir, ok)
		}
	}
	if _, ok := DirectionOf(10); ok {
		t.Error("purchase order (10) must not move stock")
	}
	if _, ok := DirectionOf(866); ok {
		t.Error("cost adjustment (866) is not a stock movement")
	}
}

// ทิศทางต้องมาจากประเภทเอกสาร ไม่ใช่จากเครื่องหมายของจำนวนในเอกสาร
func TestDirectionIgnoresSignOfQuantity(t *testing.T) {
	opt := DefaultOptions()
	positive := Calculate(nil, []Movement{
		mv("PU001", 1, 12, 10, 1000, day(1)),
		mv("AD001", 1, 68, 4, 0, day(2)), // ปรับสต็อกลด ส่งมาเป็นบวก
	}, opt)
	negative := Calculate(nil, []Movement{
		mv("PU001", 1, 12, 10, 1000, day(1)),
		mv("AD001", 1, 68, -4, 0, day(2)), // ปรับสต็อกลด ส่งมาเป็นลบ
	}, opt)

	approx(t, positive.Rows[1].BalanceQty, 6, "reduce with positive qty")
	approx(t, negative.Rows[1].BalanceQty, 6, "reduce with negative qty")
}

// สินค้าคนละคลังต้องมีต้นทุนถัวเฉลี่ยของตัวเอง ไม่ปนกัน
func TestAveragePerWarehouseIsIndependent(t *testing.T) {
	a := mv("PU001", 1, 12, 10, 1000, day(1))
	b := mv("PU002", 1, 12, 10, 3000, day(2))
	b.WhCode = "WH02"
	sale := mv("SA001", 1, 44, 1, 0, day(3))

	res := Calculate(nil, []Movement{a, b, sale}, DefaultOptions())

	approx(t, res.Rows[1].AvgCost, 300, "WH02 average")
	approx(t, res.Rows[2].UnitCost, 100, "sale from WH01 uses WH01 cost")
}

// คิดซ้ำด้วย input เดิมต้องได้ผลเท่าเดิมทุกครั้ง
func TestCalculationIsDeterministic(t *testing.T) {
	movements := []Movement{
		mv("PU001", 1, 12, 10, 1000, day(1)),
		mv("PU002", 1, 12, 5, 700, day(2)),
		mv("SA001", 1, 44, 3, 0, day(2)),
		mv("SA001", 2, 44, 2, 0, day(2)),
		mv("AD001", 1, 66, 1, 150, day(3)),
	}

	first := Calculate(nil, movements, DefaultOptions())
	for round := 0; round < 5; round++ {
		again := Calculate(nil, movements, DefaultOptions())
		for i := range first.Rows {
			if first.Rows[i] != again.Rows[i] {
				t.Fatalf("round %d row %d differs:\n%+v\n%+v", round, i, first.Rows[i], again.Rows[i])
			}
		}
	}
}

// เอกสารที่ลงวันเดียวกันต้องเรียงเป็นใบ ๆ ตาม behindindex แล้วจึงเรียงตามบรรทัด
// ไม่ใช่เอาบรรทัดที่ 1 ของทุกใบมาก่อนบรรทัดที่ 2 แบบตัวคิดต้นทุนเดิม
func TestSortGroupsLinesWithinDocument(t *testing.T) {
	same := day(5)
	movements := []Movement{
		mv("DOC-B", 2, 12, 1, 100, same),
		mv("DOC-A", 2, 12, 1, 100, same),
		mv("DOC-B", 1, 12, 1, 100, same),
		mv("DOC-A", 1, 12, 1, 100, same),
	}
	SortMovements(movements)

	want := []struct {
		doc  string
		line int
	}{{"DOC-A", 1}, {"DOC-A", 2}, {"DOC-B", 1}, {"DOC-B", 2}}
	for i, w := range want {
		if movements[i].DocNo != w.doc || movements[i].LineNumber != w.line {
			t.Fatalf("position %d = %s/%d, want %s/%d", i, movements[i].DocNo, movements[i].LineNumber, w.doc, w.line)
		}
	}
}

// behindindex ต้องมีอำนาจเหนือเลขที่เอกสารเมื่อวันที่เท่ากัน
func TestBehindIndexOverridesDocumentNumber(t *testing.T) {
	same := day(5)
	later := mv("DOC-A", 1, 12, 1, 100, same)
	later.BehindIndex = 2
	earlier := mv("DOC-Z", 1, 12, 1, 100, same)
	earlier.BehindIndex = 1

	movements := []Movement{later, earlier}
	SortMovements(movements)

	if movements[0].DocNo != "DOC-Z" {
		t.Fatalf("first = %s, want DOC-Z (behindindex 1)", movements[0].DocNo)
	}
}

// การเรียงต้องไม่ขึ้นกับลำดับที่ส่งเข้ามา — สลับยังไงก็ต้องได้ผลเดิม
func TestSortIsStableUnderShuffle(t *testing.T) {
	base := []Movement{
		mv("D1", 1, 12, 1, 10, day(1)),
		mv("D1", 2, 12, 1, 10, day(1)),
		mv("D2", 1, 44, 1, 0, day(1)),
		mv("D3", 1, 12, 1, 10, day(2)),
	}
	SortMovements(base)
	reference := Calculate(nil, base, DefaultOptions())

	source := rand.New(rand.NewSource(20260915))
	for attempt := 0; attempt < 20; attempt++ {
		shuffled := make([]Movement, len(base))
		copy(shuffled, base)
		source.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		SortMovements(shuffled)

		got := Calculate(nil, shuffled, DefaultOptions())
		for i := range reference.Rows {
			if reference.Rows[i] != got.Rows[i] {
				t.Fatalf("attempt %d row %d differs after shuffle", attempt, i)
			}
		}
	}
}

// เริ่มจากยอดยกมาของงวดก่อน ต้องได้ผลเท่ากับการไล่คิดตั้งแต่รายการแรก
func TestResumeFromOpeningMatchesFullRecalculation(t *testing.T) {
	history := []Movement{
		mv("PU001", 1, 12, 10, 1000, day(1)),
		mv("PU002", 1, 12, 10, 1400, day(2)),
	}
	later := []Movement{
		mv("SA001", 1, 44, 5, 0, day(3)),
		mv("PU003", 1, 12, 5, 800, day(4)),
	}

	full := Calculate(nil, append(append([]Movement{}, history...), later...), DefaultOptions())
	warmup := Calculate(nil, history, DefaultOptions())

	opening := map[string]Balance{"WH01": {
		Qty:     warmup.Rows[len(warmup.Rows)-1].BalanceQty,
		Amount:  warmup.Rows[len(warmup.Rows)-1].BalanceAmount,
		AvgCost: warmup.Rows[len(warmup.Rows)-1].AvgCost,
	}}
	resumed := Calculate(opening, later, DefaultOptions())

	offset := len(history)
	for i := range resumed.Rows {
		if resumed.Rows[i] != full.Rows[offset+i] {
			t.Fatalf("row %d differs:\nresumed %+v\nfull    %+v", i, resumed.Rows[i], full.Rows[offset+i])
		}
	}
}

// ของหมดคลังต้องไม่เหลือมูลค่าค้าง
func TestZeroQuantityLeavesNoValue(t *testing.T) {
	res := Calculate(nil, []Movement{
		mv("PU001", 1, 12, 3, 1000, day(1)), // ต้นทุนต่อหน่วยหารไม่ลงตัว
		mv("SA001", 1, 44, 3, 0, day(2)),
	}, DefaultOptions())

	last := res.Rows[len(res.Rows)-1]
	approx(t, last.BalanceQty, 0, "balance qty")
	approx(t, last.BalanceAmount, 0, "balance amount")
}

func TestNegativeStockPolicies(t *testing.T) {
	movements := []Movement{
		mv("PU001", 1, 12, 5, 500, day(1)),
		mv("SA001", 1, 44, 8, 0, day(2)), // ขายเกินของที่มี
	}

	allow := Calculate(nil, movements, Options{PointQty: 8, PointAmount: 2, PointCost: 2, Policy: PolicyAllowNegative})
	approx(t, allow.Rows[1].BalanceQty, -3, "allow: qty")
	approx(t, allow.Rows[1].BalanceAmount, -300, "allow: amount stays negative")

	clamp := Calculate(nil, movements, Options{PointQty: 8, PointAmount: 2, PointCost: 2, Policy: PolicyClampValue})
	approx(t, clamp.Rows[1].BalanceQty, -3, "clamp: qty")
	approx(t, clamp.Rows[1].BalanceAmount, 0, "clamp: amount floored at zero")

	reset := Calculate(nil, movements, Options{PointQty: 8, PointAmount: 2, PointCost: 2, Policy: PolicyResetOnNegative})
	approx(t, reset.Rows[1].BalanceAmount, 0, "reset: amount")
	approx(t, reset.Rows[1].AvgCost, 0, "reset: average cost")

	standard := Calculate(nil, []Movement{
		mv("SA001", 1, 44, 2, 0, day(1)), // ขายทั้งที่ยังไม่มีของ
	}, Options{PointQty: 8, PointAmount: 2, PointCost: 2, Policy: PolicyStandardCost,
		StandardCost: map[string]float64{"WH01": 70}})
	approx(t, standard.Rows[0].UnitCost, 70, "standard cost used when empty")
	approx(t, standard.Rows[0].BalanceAmount, -140, "standard cost amount")
}

// หน่วยนับที่ต้องแปลงเป็นหน่วยฐาน เช่น ซื้อเป็นลัง 1 ลัง = 12 ชิ้น
func TestUnitConversion(t *testing.T) {
	m := mv("PU001", 1, 12, 2, 2400, day(1))
	m.UnitStand = 12
	m.UnitDivide = 1

	res := Calculate(nil, []Movement{m}, DefaultOptions())
	approx(t, res.Rows[0].Qty, 24, "converted qty")
	approx(t, res.Rows[0].AvgCost, 100, "cost per base unit")
}

// ขาเข้าที่ไม่ส่งมูลค่ารวมมา ให้ตีจากราคาต่อหน่วย
func TestInboundFallsBackToUnitPrice(t *testing.T) {
	m := mv("PU001", 1, 12, 4, 0, day(1))
	m.PriceExcludeVat = 25

	res := Calculate(nil, []Movement{m}, DefaultOptions())
	approx(t, res.Rows[0].Amount, 100, "amount from unit price")
	approx(t, res.Rows[0].AvgCost, 25, "average cost")
}

func TestClosingBalancePerPeriod(t *testing.T) {
	res := Calculate(nil, []Movement{
		mv("PU001", 1, 12, 10, 1000, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)),
		mv("SA001", 1, 44, 4, 0, time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)),
	}, DefaultOptions())

	january, ok := res.Closing["WH01|2026-01"]
	if !ok {
		t.Fatal("missing january closing balance")
	}
	approx(t, january.Qty, 10, "january qty")

	february, ok := res.Closing["WH01|2026-02"]
	if !ok {
		t.Fatal("missing february closing balance")
	}
	approx(t, february.Qty, 6, "february qty")
	approx(t, february.Amount, 600, "february amount")
}

func TestUnknownTransFlagIsSkipped(t *testing.T) {
	res := Calculate(nil, []Movement{
		mv("PU001", 1, 12, 5, 500, day(1)),
		mv("SO001", 1, 14, 3, 300, day(2)), // ใบสั่งขาย ไม่กระทบสต็อก
	}, DefaultOptions())

	if len(res.Rows) != 1 {
		t.Fatalf("rows = %d, want 1 (sales order must not move stock)", len(res.Rows))
	}
}

// งวดบัญชีต้องตัดตามเวลาไทย เอกสารเที่ยงคืนครึ่งของวันที่ 1 ต้องอยู่งวดใหม่
// ไม่ใช่งวดก่อนหน้าเพราะเวลาสากลยังไม่ข้ามเดือน
// (เดิมเคยให้ฐานข้อมูลคำนวณด้วย to_char ซึ่ง PostgreSQL ปฏิเสธเพราะผลขึ้นกับเขตเวลาของ session)
func TestPeriodKeyUsesBusinessTimezone(t *testing.T) {
	bangkok := time.FixedZone("ICT", 7*60*60)

	justAfterMidnight := time.Date(2026, 3, 1, 0, 30, 0, 0, bangkok)
	if got := PeriodKeyOf(justAfterMidnight); got != "2026-03" {
		t.Errorf("period of 1 March 00:30 ICT = %s, want 2026-03", got)
	}

	// เวลาเดียวกันแต่ส่งมาเป็น UTC ต้องได้งวดเดียวกัน
	if got := PeriodKeyOf(justAfterMidnight.UTC()); got != "2026-03" {
		t.Errorf("same instant expressed in UTC = %s, want 2026-03", got)
	}

	lateFebruary := time.Date(2026, 2, 28, 23, 30, 0, 0, bangkok)
	if got := PeriodKeyOf(lateFebruary); got != "2026-02" {
		t.Errorf("period of 28 February 23:30 ICT = %s, want 2026-02", got)
	}
}
