// Package stockengine คำนวณต้นทุนสต็อกแบบถัวเฉลี่ยถ่วงน้ำหนัก แยกตามคลังสินค้า
//
// ออกแบบตาม ADR docs/kms/decisions/2026-09-15-stock-cost-engine-v2.md
// แยกตรรกะการคำนวณ (ไฟล์นี้) ออกจากการอ่าน/เขียนฐานข้อมูล (store.go) เพื่อให้ทดสอบได้จริง
package stockengine

import (
	"math"
	"sort"
	"time"
)

// Direction ทิศทางการเคลื่อนไหวสต็อก
const (
	DirectionIn  = 1  // รับเข้า
	DirectionOut = -1 // จ่ายออก
)

// NegativePolicy นโยบายเมื่อยอดคงเหลือติดลบ — ตรงกับ x_stock_amount_type ของระบบต้นแบบ
type NegativePolicy int

const (
	// PolicyAllowNegative ปล่อยให้มูลค่าติดลบตามจริง
	PolicyAllowNegative NegativePolicy = 0
	// PolicyClampValue ยอมให้จำนวนติดลบ แต่ตรึงมูลค่าไม่ให้ต่ำกว่าศูนย์
	PolicyClampValue NegativePolicy = 1
	// PolicyResetOnNegative จำนวนติดลบเมื่อไร ปรับทั้งมูลค่าและต้นทุนเป็นศูนย์
	PolicyResetOnNegative NegativePolicy = 3
	// PolicyStandardCost จำนวนติดลบให้ตีมูลค่าด้วยต้นทุนมาตรฐาน/ต้นทุนล่าสุด
	PolicyStandardCost NegativePolicy = 4
)

// ทิศทางของแต่ละประเภทเอกสาร ยึดตามที่ระบบกำหนดไว้ใน myglobal/global.go:28-29
// (ตัวคิดต้นทุนเดิมใน process-stock-calc-cost.go:386-427 ให้ทิศทาง 16, 58, 68 ไม่ตรงกับตารางนี้)
var transFlagDirection = map[int]int{
	54:  DirectionIn,  // ยอดยกมา
	12:  DirectionIn,  // ซื้อสินค้า
	310: DirectionIn,  // รับสินค้า (พาเชียล)
	48:  DirectionIn,  // รับคืนจากลูกค้า
	60:  DirectionIn,  // รับเข้า (สำเร็จรูป)
	58:  DirectionIn,  // รับคืนจากการเบิก
	66:  DirectionIn,  // ปรับสต็อก (เพิ่ม)
	44:  DirectionOut, // ขายสินค้า
	16:  DirectionOut, // ส่งคืนสินค้าให้ผู้ขาย
	56:  DirectionOut, // เบิกออก
	68:  DirectionOut, // ปรับสต็อก (ลด)
	72:  DirectionOut, // โอนย้ายออก
}

// DirectionOf คืนทิศทางของประเภทเอกสาร และบอกว่ารู้จักประเภทนี้หรือไม่
// ใช้ได้กับการคัดกรองระดับเอกสาร ส่วนการคิดต้นทุนรายบรรทัดให้ใช้ DirectionOfMovement
func DirectionOf(transFlag int) (int, bool) {
	dir, ok := transFlagDirection[transFlag]
	return dir, ok
}

// KnownTransFlags คืนประเภทเอกสารทั้งหมดที่เครื่องคิดต้นทุนรู้จักทิศทาง
// ใช้ตรวจว่ารายชื่อประเภทเอกสารที่ระบบนำมาคิด ครอบคลุมครบตามที่เครื่องคิดต้นทุนรู้จัก
func KnownTransFlags() []int {
	flags := make([]int, 0, len(transFlagDirection))
	for flag := range transFlagDirection {
		flags = append(flags, flag)
	}
	sort.Ints(flags)
	return flags
}

// TransFlagStockTransfer คือประเภทเอกสารโอนย้ายสินค้าระหว่างคลัง
const TransFlagStockTransfer = 72

// DirectionOfMovement คืนทิศทางของรายการหนึ่งบรรทัด
//
// เอกสารโอนคลังหนึ่งใบมีสองขาเสมอ ขาออกจากคลังต้นทางกับขาเข้าคลังปลายทาง
// ทิศทางของบรรทัดจึงอ่านจาก calcflag ไม่ใช่จากประเภทเอกสาร
// ถ้าดูแต่ประเภทเอกสาร ขาเข้าจะถูกคิดเป็นขาออกด้วย แล้วของจะหายจากบริษัททุกครั้งที่โอนคลัง
func DirectionOfMovement(mv Movement) (int, bool) {
	if mv.TransFlag == TransFlagStockTransfer {
		if mv.CalcFlag < 0 {
			return DirectionOut, true
		}
		return DirectionIn, true
	}
	return DirectionOf(mv.TransFlag)
}

// Movement คือรายการเคลื่อนไหวหนึ่งบรรทัดที่อ่านมาจาก docdetail
type Movement struct {
	WhCode          string
	LocationCode    string
	Barcode         string
	UnitCode        string
	DocRef          string
	DocDateTime     time.Time
	BehindIndex     int
	DocNo           string
	LineNumber      int
	TransFlag       int
	CalcFlag        int     // ทิศทางที่เอกสารกำหนดให้บรรทัดนี้ ใช้กับเอกสารที่มีสองขาในใบเดียว (โอนคลัง)
	Qty             float64 // จำนวนตามหน่วยนับในเอกสาร
	UnitStand       float64
	UnitDivide      float64
	Price           float64
	PriceExcludeVat float64
	SumAmount       float64
}

// LedgerRow คือผลลัพธ์หนึ่งบรรทัดที่จะเขียนลง stock_ledger
//
// เก็บทั้งข้อเท็จจริงตามเอกสาร (DocQty, UnitStand, UnitDivide, Price, DocValue)
// และผลการคิดต้นทุน (Qty, UnitCost, Amount, Balance*) ไว้ในแถวเดียว
// เพื่อให้รายงานอ่านตารางเดียวจบ ไม่ต้องย้อนไป join docdetail อีก
type LedgerRow struct {
	WhCode        string
	LocationCode  string
	Barcode       string
	UnitCode      string
	DocRef        string
	DocDateTime   time.Time
	BehindIndex   int
	DocNo         string
	LineNumber    int
	TransFlag     int
	Direction     int
	DocQty        float64 // จำนวนตามหน่วยนับในเอกสาร ใช้แสดงคู่กับ UnitCode
	UnitStand     float64
	UnitDivide    float64
	Price         float64 // ราคาต่อหน่วยตามเอกสาร
	DocValue      float64 // มูลค่าตามเอกสาร (ยอดขาย/ยอดซื้อของบรรทัดนี้)
	Qty           float64 // จำนวนหน่วยฐานที่ใช้คิดต้นทุน
	UnitCost      float64
	Amount        float64
	BalanceQty    float64
	BalanceAmount float64
	AvgCost       float64
}

// Balance คือยอดคงเหลือของสินค้าในคลังหนึ่ง
type Balance struct {
	Qty      float64
	Amount   float64
	AvgCost  float64
	HasTrans bool
}

// Options ตัวเลือกการคำนวณ
type Options struct {
	PointQty     int // ทศนิยมจำนวน
	PointAmount  int // ทศนิยมมูลค่า
	PointCost    int // ทศนิยมต้นทุน
	Policy       NegativePolicy
	StandardCost map[string]float64 // ต้นทุนมาตรฐานต่อคลัง ใช้เมื่อ PolicyStandardCost
}

// DefaultOptions ค่าเริ่มต้นที่ตรงกับตัวคิดต้นทุนเดิม
func DefaultOptions() Options {
	return Options{PointQty: 8, PointAmount: 2, PointCost: 2, Policy: PolicyAllowNegative}
}

// Result ผลลัพธ์การคำนวณหนึ่งรอบ
type Result struct {
	Rows []LedgerRow
	// Closing คือยอดคงเหลือปลายงวดของแต่ละ (คลัง, งวด) คีย์คือ whcode + "|" + periodkey
	Closing map[string]Balance
}

// businessLocation คือเขตเวลาที่ใช้ตัดสินว่าเอกสารอยู่ในงวดบัญชีใด
//
// ต้องกำหนดให้ชัดเจนที่เดียว ไม่ปล่อยให้ขึ้นกับเขตเวลาของเครื่องหรือของ session ฐานข้อมูล
// เอกสารที่ลงเวลา 1 มีนาคม 00:30 น. ตามเวลาไทย ต้องอยู่ในงวดมีนาคมเสมอ
// ไม่ใช่งวดกุมภาพันธ์เพราะเวลาสากลยังไม่ข้ามวัน
var businessLocation = mustLoadLocation("Asia/Bangkok")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// ระบบที่ไม่มีฐานข้อมูลเขตเวลาให้ใช้ค่าคงที่ +07:00 แทน ผลลัพธ์เท่ากันสำหรับประเทศไทย
		return time.FixedZone("ICT", 7*60*60)
	}
	return loc
}

// SetBusinessLocation กำหนดเขตเวลาที่ใช้ตัดงวดบัญชี ใช้ตอนตั้งค่าระบบเท่านั้น
func SetBusinessLocation(loc *time.Location) {
	if loc != nil {
		businessLocation = loc
	}
}

// PeriodKeyOf คืนรหัสงวดรูปแบบ YYYY-MM ตามเขตเวลาธุรกิจ
//
// คำนวณในโปรแกรมแทนที่จะให้ฐานข้อมูลสร้างให้ เพราะ to_char กับ AT TIME ZONE
// ขึ้นกับฐานข้อมูลเขตเวลาที่แก้ไขได้ PostgreSQL จึงไม่ยอมให้ใช้เป็นคอลัมน์ที่สร้างอัตโนมัติ
func PeriodKeyOf(t time.Time) string {
	return t.In(businessLocation).Format("2006-01")
}

// SortMovements เรียงรายการตามคีย์ที่ให้ผลลัพธ์เหมือนเดิมทุกครั้ง
// ลำดับนี้ต้องตรงกับ ORDER BY ที่ใช้อ่านจากฐานข้อมูล (store.go) และตรงกับระบบต้นแบบ
// ที่เรียง DocDate → BehindIndex → DocNo → LineNumber
func SortMovements(movements []Movement) {
	sort.SliceStable(movements, func(i, j int) bool {
		a, b := movements[i], movements[j]
		if !a.DocDateTime.Equal(b.DocDateTime) {
			return a.DocDateTime.Before(b.DocDateTime)
		}
		if a.BehindIndex != b.BehindIndex {
			return a.BehindIndex < b.BehindIndex
		}
		if a.DocNo != b.DocNo {
			return a.DocNo < b.DocNo
		}
		return a.LineNumber < b.LineNumber
	})
}

// Calculate คำนวณต้นทุนถัวเฉลี่ยถ่วงน้ำหนักแยกตามคลัง
//
// opening คือยอดตั้งต้นของแต่ละคลัง (มาจาก stock_period_balance ของงวดก่อนหน้า)
// คลังที่ไม่มีใน opening ถือว่าเริ่มจากศูนย์
//
// ผู้เรียกต้องส่ง movements ที่เรียงแล้วด้วย SortMovements หรืออ่านมาตามลำดับเดียวกัน
func Calculate(opening map[string]Balance, movements []Movement, opt Options) Result {
	state := make(map[string]Balance, len(opening)+4)
	for wh, bal := range opening {
		state[wh] = Balance{Qty: bal.Qty, Amount: bal.Amount, AvgCost: bal.AvgCost}
	}

	rows := make([]LedgerRow, 0, len(movements))
	closing := make(map[string]Balance, len(state)+4)

	// ต้นทุนของขาออกในเอกสารโอนคลัง เก็บไว้ให้ขาเข้าของใบเดียวกันใช้ต่อ
	// การโอนของภายในบริษัทต้องไม่ทำให้มูลค่าสต็อกรวมเปลี่ยน ปลายทางจึงรับของด้วยต้นทุนของต้นทาง
	transferCost := map[string]float64{}

	// งวดของยอดตั้งต้น ใช้เป็นฐานให้คลังที่ไม่มีรายการในรอบนี้
	for _, mv := range movements {
		bal := state[mv.WhCode]

		qty := normalizeQty(mv, opt.PointQty)
		docQty := documentQty(mv, opt.PointQty)
		direction, known := DirectionOfMovement(mv)
		if !known {
			// ประเภทเอกสารที่ไม่รู้จักทิศทาง ไม่นำมาคิดสต็อก ปลอดภัยกว่าการเดา
			continue
		}

		unitCost, amount := valueOf(mv, qty, docQty, direction, bal, opt)
		if mv.TransFlag == TransFlagStockTransfer {
			unitCost, amount = transferValue(mv, qty, direction, unitCost, amount, transferCost, opt)
		}

		signedQty := qty
		if direction == DirectionOut {
			signedQty = -qty
		}
		signedAmount := amount
		if direction == DirectionOut {
			signedAmount = -amount
		}

		bal.Qty = myRound(bal.Qty+signedQty, opt.PointQty)
		bal.Amount = myRound(bal.Amount+signedAmount, opt.PointAmount)
		bal.HasTrans = true

		applyPolicy(&bal, opt)

		if bal.Qty > 0 {
			bal.AvgCost = myRound(bal.Amount/bal.Qty, opt.PointCost)
		}
		// ยอดคงเหลือไม่เป็นบวก ให้คงต้นทุนถัวเฉลี่ยเดิมไว้ใช้กับรายการถัดไป

		state[mv.WhCode] = bal

		rows = append(rows, LedgerRow{
			WhCode:        mv.WhCode,
			LocationCode:  mv.LocationCode,
			Barcode:       mv.Barcode,
			UnitCode:      mv.UnitCode,
			DocRef:        mv.DocRef,
			DocDateTime:   mv.DocDateTime,
			BehindIndex:   mv.BehindIndex,
			DocNo:         mv.DocNo,
			LineNumber:    mv.LineNumber,
			TransFlag:     mv.TransFlag,
			Direction:     direction,
			DocQty:        docQty,
			UnitStand:     mv.UnitStand,
			UnitDivide:    mv.UnitDivide,
			Price:         math.Abs(mv.Price),
			DocValue:      documentValue(mv, docQty, opt),
			Qty:           qty,
			UnitCost:      unitCost,
			Amount:        amount,
			BalanceQty:    bal.Qty,
			BalanceAmount: bal.Amount,
			AvgCost:       bal.AvgCost,
		})

		closing[mv.WhCode+"|"+PeriodKeyOf(mv.DocDateTime)] = bal
	}

	return Result{Rows: rows, Closing: closing}
}

// normalizeQty แปลงจำนวนตามหน่วยนับในเอกสารให้เป็นหน่วยฐาน และคืนค่าเป็นจำนวนบวกเสมอ
// ทิศทางถูกตัดสินจากประเภทเอกสารอย่างเดียว ไม่ใช่จากเครื่องหมายของจำนวน
func normalizeQty(mv Movement, pointQty int) float64 {
	factor := 1.0
	if mv.UnitDivide != 0 {
		factor = mv.UnitStand / mv.UnitDivide
	}
	return math.Abs(myRound(mv.Qty*factor, pointQty))
}

// documentQty คืนจำนวนตามหน่วยนับในเอกสาร เป็นค่าบวกเสมอ
func documentQty(mv Movement, pointQty int) float64 {
	return math.Abs(myRound(mv.Qty, pointQty))
}

// documentValue คืนมูลค่าตามเอกสารของบรรทัดนี้ (ยอดขายสำหรับขาออก ยอดซื้อสำหรับขาเข้า)
//
// ราคาต่อหน่วยในเอกสารเป็นราคาต่อ "หน่วยนับในเอกสาร" จึงต้องคูณกับจำนวนตามเอกสาร
// ไม่ใช่จำนวนหน่วยฐาน มิฉะนั้นสินค้าที่ขายเป็นลัง (1 ลัง = 12 ชิ้น) จะได้มูลค่าเกินจริง 12 เท่า
func documentValue(mv Movement, docQty float64, opt Options) float64 {
	value := math.Abs(mv.SumAmount)
	if value == 0 && mv.Price != 0 {
		value = docQty * math.Abs(mv.Price)
	}
	return myRound(value, opt.PointAmount)
}

// valueOf คืนต้นทุนต่อหน่วยและมูลค่ารวมของรายการหนึ่ง
// ขาเข้าใช้มูลค่าจากเอกสาร ขาออกใช้ต้นทุนถัวเฉลี่ยขณะนั้น
func valueOf(mv Movement, qty, docQty float64, direction int, bal Balance, opt Options) (unitCost, amount float64) {
	if direction == DirectionOut {
		unitCost = bal.AvgCost
		if bal.Qty <= 0 && opt.Policy == PolicyStandardCost {
			if std, ok := opt.StandardCost[mv.WhCode]; ok && std > 0 {
				unitCost = std
			}
		}
		return unitCost, myRound(qty*unitCost, opt.PointAmount)
	}

	amount = math.Abs(mv.SumAmount)
	if amount == 0 && mv.PriceExcludeVat != 0 {
		amount = myRound(docQty*math.Abs(mv.PriceExcludeVat), opt.PointAmount)
	}
	if amount == 0 && mv.Price != 0 {
		amount = myRound(docQty*math.Abs(mv.Price), opt.PointAmount)
	}
	amount = myRound(amount, opt.PointAmount)

	if qty > 0 {
		unitCost = myRound(amount/qty, opt.PointCost)
	}
	return unitCost, amount
}

// transferValue ตีมูลค่าของบรรทัดในเอกสารโอนคลัง
//
// ขาออกบันทึกต้นทุนที่ใช้ไว้ให้ขาเข้าของใบเดียวกันหยิบไปใช้ ขาเข้าจึงรับของด้วยต้นทุนเดิม
// ไม่ใช่ราคาในเอกสาร (ใบโอนมักไม่มีราคา ถ้าใช้ราคาในเอกสารของจะกลายเป็นต้นทุนศูนย์)
// ขาเข้าที่หาขาออกคู่กันไม่เจอ ให้ใช้มูลค่าที่ตีไว้ตามปกติ ดีกว่าปล่อยเป็นศูนย์เงียบ ๆ
func transferValue(mv Movement, qty float64, direction int, unitCost, amount float64, transferCost map[string]float64, opt Options) (float64, float64) {
	key := mv.DocNo + "|" + mv.Barcode
	if direction == DirectionOut {
		transferCost[key] = unitCost
		return unitCost, amount
	}
	sourceCost, ok := transferCost[key]
	if !ok {
		return unitCost, amount
	}
	return sourceCost, myRound(qty*sourceCost, opt.PointAmount)
}

// applyPolicy ปรับยอดตามนโยบายสต็อกติดลบ
func applyPolicy(bal *Balance, opt Options) {
	// ยอดที่เกือบศูนย์ให้ถือเป็นศูนย์ ไม่ทิ้งมูลค่าค้างไว้เมื่อของหมด
	if math.Abs(bal.Qty) < 1e-6 {
		bal.Qty = 0
		bal.Amount = 0
		return
	}
	if math.Abs(bal.Amount) < 1e-4 {
		bal.Amount = 0
	}

	if bal.Qty >= 0 {
		return
	}

	switch opt.Policy {
	case PolicyClampValue:
		if bal.Amount < 0 {
			bal.Amount = 0
		}
	case PolicyResetOnNegative:
		bal.Amount = 0
		bal.AvgCost = 0
	case PolicyStandardCost, PolicyAllowNegative:
		// ปล่อยตามจริง ต้นทุนที่ใช้ถูกเลือกไว้แล้วตอนตีมูลค่าขาออก
	}
}

func myRound(value float64, precision int) float64 {
	multiplier := math.Pow10(precision)
	return math.Round(value*multiplier) / multiplier
}
