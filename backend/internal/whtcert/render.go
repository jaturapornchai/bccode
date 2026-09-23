package whtcert

import (
	_ "embed"
	"strconv"
	"sync"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/pdftext"
)

//go:embed assets/50tawi-rd.pdf
var officialForm []byte

var (
	prepareOnce sync.Once
	blankForm   []byte // แบบฟอร์มทางการที่ถอดช่องกรอก (AcroForm) ออกแล้ว — ข้อความเราเป็นตัวจริงตัวเดียว
	textShaper  *pdftext.Shaper
	prepareErr  error
)

func prepare() error {
	prepareOnce.Do(func() {
		if textShaper, prepareErr = pdftext.Default(); prepareErr != nil {
			return
		}
		blankForm, prepareErr = pdftext.BlankForm(officialForm)
	})
	return prepareErr
}

// Render - สร้าง PDF ใบ 50 ทวิ: หน้า 1 = ฉบับที่ 1, หน้า 2 = ฉบับที่ 2, หน้า 3 = สำเนาคู่ฉบับ (ถ้าขอ)
// ข้อมูลเหมือนกันทุกหน้าตามประกาศกรมสรรพากร ("2 ฉบับมีข้อความตรงกัน")
func Render(c Certificate) ([]byte, error) {
	n, err := normalize(c)
	if err != nil {
		return nil, err
	}
	if err := prepare(); err != nil {
		return nil, err
	}
	o := pdftext.NewOverlay(textShaper)
	copies := []int{1, 2}
	if n.ArchiveCopy {
		copies = append(copies, 3)
	}
	for _, copyNo := range copies {
		if err := drawCertificate(o, o.NewPage(pageWidth, pageHeight), n, copyNo); err != nil {
			return nil, err
		}
	}
	return stamp(o, len(copies))
}

func drawCertificate(o *pdftext.Overlay, p *pdftext.Page, n normalized, copyNo int) error {
	put := func(f lineField, text, field string) error {
		if text == "" {
			return nil
		}
		size, ok := o.Fit(text, f.maxWidth, fontSize, minFont)
		if !ok {
			return invalid("wht_cert_text_too_long", field)
		}
		o.Text(p, f.x, f.baseline, text, size, f.align)
		return nil
	}

	label := ""
	switch {
	case n.Replacement:
		label = "ใบแทน"
	case copyNo == 3:
		label = "สำเนาคู่ฉบับ"
	}
	if copyNo == 3 && n.Replacement {
		label = "ใบแทน (สำเนาคู่ฉบับ)"
	}
	if err := put(fieldCopyLabel, label, "copy"); err != nil {
		return err
	}
	if c, ok := copyChecks[copyNo]; ok {
		o.Check(p, c.x, c.y)
	}

	for _, item := range []struct {
		f           lineField
		text, field string
	}{
		{fieldBookNo, n.BookNo, "bookno"},
		{fieldRunNo, n.RunNo, "runno"},
		{fieldPayerName, n.Payer.Name, "payer.name"},
		{fieldPayerAddress, n.Payer.Address, "payer.address"},
		{fieldPayeeName, n.Payee.Name, "payee.name"},
		{fieldPayeeAddress, n.Payee.Address, "payee.address"},
		{fieldSequence, n.SequenceNo, "sequenceno"},
		{fieldConditionTxt, n.ConditionNote, "conditionnote"},
	} {
		if err := put(item.f, item.text, item.field); err != nil {
			return err
		}
	}
	digits(o, p, payerID13, n.payerID)
	digits(o, p, payeeID13, n.payeeID)
	digits(o, p, payerID10, n.payerID10)
	digits(o, p, payeeID10, n.payeeID10)
	for _, c := range []point{formChecks[n.Form], conditionChecks[n.Condition]} {
		o.Check(p, c.x, c.y)
	}

	for _, row := range n.incomes {
		baseline := incomeRowBaselines[row.Type]
		if err := put(lineField{x: colDateCenter, baseline: baseline, maxWidth: colDateWidth, align: alignCenter}, row.dateText, "paiddate"); err != nil {
			return err
		}
		if err := amountCells(o, p, baseline, row.amount, row.tax); err != nil {
			return err
		}
		if f, ok := incomeNoteField[row.Type]; ok {
			if err := put(f, row.Note, "incomes.note"); err != nil {
				return err
			}
		}
	}
	if err := amountCells(o, p, totalRowBaseline, n.totalAmount, n.totalTax); err != nil {
		return err
	}
	if err := put(fieldTaxWords, BahtText(n.totalTax), "taxwords"); err != nil {
		return err
	}
	for i, f := range []lineField{fieldFundGPF, fieldFundSSF, fieldFundPVD} {
		if n.funds[i].IsPositive() {
			if err := put(f, moneyText(n.funds[i]), "fund"); err != nil {
				return err
			}
		}
	}
	if err := put(fieldIssueDay, strconv.Itoa(n.issued.Day()), "issuedate"); err != nil {
		return err
	}
	if err := put(fieldIssueMonth, thaiMonths[n.issued.Month()-1], "issuedate"); err != nil {
		return err
	}
	return put(fieldIssueYear, strconv.Itoa(n.issued.Year()+543), "issuedate")
}

// amountCells - ยอดจ่าย|ภาษี แยกบาท (ชิดขวาก่อนเส้นแบ่ง) และสตางค์ (กลางช่องสตางค์)
func amountCells(o *pdftext.Overlay, p *pdftext.Page, baseline float64, amount, tax decimal.Decimal) error {
	for _, col := range []struct {
		v                    decimal.Decimal
		right, satang, width float64
	}{
		{amount, colPayBahtRight, colPaySatang, colPayBahtWidth},
		{tax, colTaxBahtRight, colTaxSatang, colTaxBahtWidth},
	} {
		baht, satang := splitMoney(col.v)
		size, ok := o.Fit(baht, col.width, fontSize, minFont)
		if !ok {
			return invalid("wht_cert_amount_too_large", "amount")
		}
		o.Text(p, col.right, baseline, baht, size, alignRight)
		o.Text(p, col.satang, baseline, satang, fontSize, alignCenter)
	}
	return nil
}

// digits - หนึ่งหลักต่อหนึ่งช่อง กลางช่อง
func digits(o *pdftext.Overlay, p *pdftext.Page, boxes idBoxes, id string) {
	if id == "" {
		return
	}
	i := 0
	for _, g := range boxes.groups {
		cell := (g.right - g.left) / float64(g.count)
		for k := 0; k < g.count && i < len(id); k++ {
			o.Text(p, g.left+cell*(float64(k)+0.5), boxes.baseline, id[i:i+1], fontSize, alignCenter)
			i++
		}
	}
}

// stamp - ต่อแบบฟอร์มเปล่า N หน้า แล้วประทับชั้นข้อความหน้า i ลงหน้า i (multistamp)
func stamp(o *pdftext.Overlay, pages int) ([]byte, error) {
	copies := make([][]byte, pages)
	for i := range copies {
		copies[i] = blankForm
	}
	base, err := pdftext.Merge(copies...)
	if err != nil {
		return nil, err
	}
	return pdftext.Stamp(base, o)
}
