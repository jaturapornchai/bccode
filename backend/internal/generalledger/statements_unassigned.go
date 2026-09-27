package generalledger

import (
	"sort"

	"github.com/shopspring/decimal"
)

// บัญชีที่มียอดแต่ไม่อยู่ในบรรทัดใดของงบฐานะการเงิน/งบกำไรขาดทุน — ยอดรวมของงบจะไม่ครบโดยไม่มีใครเห็น จึงแจ้งบนจอ
// (นอกตารางที่พิมพ์). แม่แบบไม่อ้างถึงกัน: บัญชีรายได้/ค่าใช้จ่ายที่อยู่แค่ในงบต้นทุนการผลิต งบกำไรขาดทุนก็ยังไม่ตรงกับบัญชี.
// งบอื่น (ต้นทุนการผลิต กระแสเงินสด ส่วนของผู้ถือหุ้น กำหนดเอง) ไม่ตรวจ — งบกระแสเงินสดมีการตรวจเงินสดปลายงวดแยก

type StatementUnassigned struct {
	Key         string `json:"key"` // amount | prioramount
	FiscalYear  string `json:"fiscalyear"`
	AccountCode string `json:"accountcode"` // or "__current_earnings__"
	AccountName string `json:"accountname"` // "" for the earnings item (frontend renders a key)
	AccountType string `json:"accounttype"`
	Basis       string `json:"basis"`  // closing | movement
	Amount      string `json:"amount"` // ด้านปกติตามหมวดบัญชี ปัดตามทศนิยมของรูปแบบงบ
}

// statementUnassigned: งบฐานะการเงินดูยอดคงเหลือปลายงวดของสินทรัพย์/หนี้สิน/ส่วนของเจ้าของ + กำไร (ขาดทุน) ที่ยังไม่ปิดบัญชี
// เป็นรายการเดียว (ไม่แยกรายบัญชีรายได้/ค่าใช้จ่าย); งบกำไรขาดทุนดูความเคลื่อนไหวของรายได้/ค่าใช้จ่าย. บัญชีที่อยู่แค่ในบรรทัด
// ฐานอื่น (เช่น ยอดต้นงวด) ถือว่ายังไม่อยู่ในงบ; ยอดที่ปัดแล้วเป็นศูนย์ไม่แจ้ง
func statementUnassigned(template Master, balances map[string]statementBalance, key, fiscalYear string, scale int32) []StatementUnassigned {
	var basis string
	switch template.StatementType {
	case "balance_sheet":
		basis = "closing"
	case "pnl":
		basis = "movement"
	default:
		return nil
	}
	assigned := map[string]bool{}
	for _, target := range statementCodeTargets(template) {
		if target.basis != basis {
			continue
		}
		for _, code := range target.codes {
			assigned[code] = true
		}
	}
	if basis == "movement" && assigned[statementCurrentEarnings] {
		return nil
	}
	items := []StatementUnassigned{}
	uncovered := decimal.Zero // กำไร (ขาดทุน) ที่ยังไม่ปิดบัญชีของบัญชีรายได้/ค่าใช้จ่ายที่ไม่อยู่ในบรรทัดใด (เดบิตลบเครดิต)
	for code, balance := range balances {
		if code == statementCurrentEarnings || assigned[code] {
			continue
		}
		result := balance.accountType == "income" || balance.accountType == "expense"
		value := balance.balance
		switch {
		case basis == "movement" && result:
			value = balance.movement
		case basis == "movement":
			continue
		case result:
			uncovered = uncovered.Add(balance.balance)
			continue
		case balance.accountType != "asset" && balance.accountType != "liability" && balance.accountType != "equity":
			continue
		}
		amount := statementSigned(balance, value, "").Round(scale)
		if amount.IsZero() {
			continue
		}
		items = append(items, StatementUnassigned{Key: key, FiscalYear: fiscalYear, AccountCode: code, AccountName: balance.name, AccountType: balance.accountType, Basis: basis, Amount: amount.StringFixed(scale)})
	}
	if basis == "closing" && !assigned[statementCurrentEarnings] {
		// ส่วนของเจ้าของแสดงเครดิตเป็นบวก
		if amount := uncovered.Neg().Round(scale); !amount.IsZero() {
			items = append(items, StatementUnassigned{Key: key, FiscalYear: fiscalYear, AccountCode: statementCurrentEarnings, AccountType: "equity", Basis: basis, Amount: amount.StringFixed(scale)})
		}
	}
	if len(items) == 0 {
		return nil
	}
	sort.Slice(items, func(i, j int) bool { return items[i].AccountCode < items[j].AccountCode })
	return items
}
