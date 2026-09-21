// คำสั่ง glseed เติมข้อมูลหลักฐานลูกหนี้ เจ้าหนี้ ธนาคาร งบประมาณ การเชื่อมบัญชี
// กลุ่มบัญชีสินค้า และใบสำคัญที่ผ่านบัญชีแล้ว ให้ครบทุกจอของระบบบัญชีแยกประเภท
// ทุกการเขียนผ่าน service layer เดียวกับ backend (store.Execute) ไม่ยิง SQL ตรง
//
// ใช้งาน: GLSEED_DSN='postgres://...' go run ./cmd/glseed -company 01 -fiscal 2569
// ตรวจแผนก่อนเขียนจริง: เพิ่ม -apply (ค่าเริ่มต้น dry-run พิมพ์แผนอย่างเดียว)
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"

	gl "smlcloudplatform/internal/generalledger"
)

type seedAccount struct {
	Code     string `json:"accountcode"`
	Name     string `json:"-"`
	Level    int    `json:"level"`
	Posting  bool   `json:"allowposting"`
	Type     string `json:"accounttype"`
	IsActive bool   `json:"isactive"`
	Parent   string `json:"parentaccountcode"`
}

type nameEntry struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func main() {
	company := flag.String("company", "01", "รหัสบริษัทในฐาน Holding")
	branch := flag.String("branch", "00000", "รหัสสาขา")
	holding := flag.String("holding", "rungrueng", "รหัส Holding (สำหรับ connector)")
	fiscal := flag.String("fiscal", "2569", "ปีบัญชี")
	actor := flag.String("actor", "seed-gl-complete-screens", "ผู้บันทึกใน audit")
	apply := flag.Bool("apply", false, "เขียนจริง (ค่าเริ่มต้น dry-run)")
	flag.Parse()

	dsn := os.Getenv("GLSEED_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "กรุณาตั้ง GLSEED_DSN")
		os.Exit(2)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fatal("เชื่อมต่อฐานไม่ได้: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fatal("ping ฐานไม่ได้: %v", err)
	}

	// โหลดผังบัญชีจริงของบริษัท เพื่อให้ทุกใบใหม่ใช้บัญชีที่มีอยู่แล้วเท่านั้น
	rows, err := db.Query(`SELECT payload FROM gl_records WHERE company=$1 AND kind='accounts'`, *company)
	if err != nil {
		fatal("อ่านผังบัญชีไม่ได้: %v", err)
	}
	accounts := map[string]*seedAccount{}
	byName := map[string][]*seedAccount{}
	for rows.Next() {
		var raw []byte
		var a seedAccount
		if err := rows.Scan(&raw); err != nil {
			fatal("%v", err)
		}
		if err := json.Unmarshal(raw, &a); err != nil {
			fatal("payload บัญชีพัง: %v", err)
		}
		var names []nameEntry
		_ = json.Unmarshal(raw, &struct {
			Names *[]nameEntry `json:"names"`
		}{Names: &names})
		if len(names) > 0 {
			a.Name = names[0].Name
		}
		accounts[a.Code] = &a
		byName[a.Name] = append(byName[a.Name], &a)
	}
	rows.Close()
	if len(accounts) == 0 {
		fatal("ไม่พบผังบัญชีของบริษัท %s", *company)
	}

	acc := func(code string) string {
		a, ok := accounts[code]
		if !ok || !a.Posting || !a.IsActive {
			fatal("บัญชี %s ไม่พร้อมลงรายการ (มี=%v)", code, ok)
		}
		return code
	}
	// หาบัญชีจากชื่อ (ต้องเจอตัวเดียวเท่านั้น เพื่อไม่ให้ seed เดาเอง)
	findByName := func(keyword, wantType string) string {
		var hits []*seedAccount
		for name, list := range byName {
			if !strings.Contains(name, keyword) {
				continue
			}
			for _, a := range list {
				if a.Posting && a.IsActive && a.Type == wantType {
					hits = append(hits, a)
				}
			}
		}
		if len(hits) != 1 {
			fatal("ค้นหาบัญชีด้วยคำว่า %q พบ %d บัญชี ต้องเจอตัวเดียว: %v", keyword, len(hits), func() []string {
				codes := []string{}
				for _, h := range hits {
					codes = append(codes, h.Code+" "+h.Name)
				}
				return codes
			}())
		}
		return hits[0].Code
	}

	ctx := context.Background()
	pg := gl.NewPostgres(func(string) (*sql.DB, error) { return db, nil })
	store := gl.NewPostgresStore(pg)
	scope := gl.Scope{Holding: *holding, Company: *company, Branch: *branch, Actor: *actor}

	// ตรวจว่า docno นี้มีอยู่แล้วหรือไม่ เพื่อให้รันซ้ำได้อย่างปลอดภัย (idempotent)
	exists := func(docno string) bool {
		var n int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gl_records WHERE company=$1 AND kind='journals' AND payload->>'docno'=$2`, *company, docno).Scan(&n); err != nil {
			fatal("ตรวจ docno ไม่ได้: %v", err)
		}
		return n > 0
	}

	// ค้นบัญชีตามชื่อแบบยืดหยุ่น (คืน "" ถ้าไม่พบ) และใช้สร้างบัญชีใหม่เมื่อจำเป็น
	quietFind := func(keywords ...string) string {
		for _, keyword := range keywords {
			for name, list := range byName {
				if !strings.Contains(name, keyword) {
					continue
				}
				for _, a := range list {
					if a.Posting && a.IsActive {
						return a.Code
					}
				}
			}
		}
		return ""
	}

	// บัญชีเฉพาะทางที่แผนต้องใช้ หาจากชื่อจริงในผังบัญชี
	cash := acc("1111")
	bank := acc("1121")
	ar := acc("1131")
	goods := acc("1141")
	inputVAT := acc("1151")
	ap := acc("2111")
	outputVAT := acc("2131")
	capital := acc("3111")
	salesVAT := acc("4111")
	wht := quietFind("ภาษีหัก ณ ที่จ่าย", "ภาษีเงินได้หัก ณ ที่จ่าย")
	rent := findByName("ค่าเช่า", "expense")
	cogs := acc("5121")
	stationery := acc("5331")

	// ถ้าผังบัญชียังไม่มีบัญชีค้างนำส่งภาษีหัก ณ ที่จ่าย ให้สร้างใหม่ใต้กลุ่มเดียวกับภาษีขาย
	if wht == "" {
		vatAcct, ok := accounts[outputVAT]
		if !ok || vatAcct.Parent == "" {
			fatal("ไม่พบกลุ่มบัญชีแม่ของภาษีขาย %s สำหรับสร้างบัญชีภาษีหัก ณ ที่จ่าย", outputVAT)
		}
		whtCode := ""
		for n := 1; n <= 9; n++ {
			candidate := outputVAT[:2] + fmt.Sprintf("%d1", n) + "1"
			if candidate != outputVAT {
				if _, taken := accounts[candidate]; !taken {
					whtCode = candidate
					break
				}
			}
		}
		if whtCode == "" {
			fatal("ไม่พบรหัสบัญชีว่างสำหรับบัญชีภาษีหัก ณ ที่จ่าย")
		}
		fmt.Printf("แผน: สร้างบัญชีภาษีหัก ณ ที่จ่ายค้างนำส่ง %s ใต้ %s\n", whtCode, vatAcct.Parent)
		if *apply {
			cmd := gl.Command{Resource: "accounts", Action: "create", RequestID: "seed-gl-screens-acc-" + whtCode,
				Account: &gl.Account{AccountCode: whtCode, Names: []gl.Name{{Code: "th", Name: "ภาษีเงินได้หัก ณ ที่จ่ายค้างนำส่ง"}}, AccountType: "liability", NormalBalance: "credit", AllowPosting: true, IsActive: true, Level: vatAcct.Level, ParentAccountCode: vatAcct.Parent}}
			if _, err := store.Execute(ctx, scope, cmd); err != nil {
				fatal("สร้างบัญชีภาษีหัก ณ ที่จ่ายไม่ได้: %v", err)
			}
			accounts[whtCode] = &seedAccount{Code: whtCode, Level: vatAcct.Level, Posting: true, Type: "liability", IsActive: true, Parent: vatAcct.Parent}
		}
		wht = whtCode
	}

	createJournal := func(docno, date, book, desc, kind string, lines []gl.Line, details *gl.JournalDetails) (string, string) {
		// สมดุลก่อนส่งเสมอ
		var dr, cr decimal.Decimal
		for _, l := range lines {
			d, _ := decimal.NewFromString(string(l.Debit))
			c, _ := decimal.NewFromString(string(l.Credit))
			dr = dr.Add(d)
			cr = cr.Add(c)
		}
		if !dr.Equal(cr) || dr.IsZero() {
			fatal("ใบ %s ไม่สมดุล DR=%s CR=%s", docno, dr, cr)
		}
		if exists(docno) {
			fmt.Printf("ข้าม %s (มีอยู่แล้ว)\n", docno)
			return "", "skipped"
		}
		fmt.Printf("แผน: %s %s [%s] %s — %d บรรทัด DR=CR=%s\n", docno, date, book, desc, len(lines), dr)
		if !*apply {
			return "", "dry-run"
		}
		cmd := gl.Command{Resource: "journals", Action: "create", RequestID: "seed-gl-screens-" + docno + "-create",
			Journal: &gl.Journal{DocNo: docno, Date: date, BookCode: book, BranchCode: *branch, FiscalYear: *fiscal, Description: desc, Kind: kind, Lines: lines, Details: details}}
		res, err := store.Execute(ctx, scope, cmd)
		if err != nil {
			fatal("สร้าง %s ไม่ได้: %v", docno, err)
		}
		if kind != "" && kind != "draft-hold" {
			post := gl.Command{Resource: "journals", Action: "post", RequestID: "seed-gl-screens-" + docno + "-post", ID: res.ID, Version: res.Version, Reason: "ผ่านรายการตามแผนข้อมูลเริ่มต้น"}
			if _, err := store.Execute(ctx, scope, post); err != nil {
				fatal("ผ่านรายการ %s ไม่ได้: %v", docno, err)
			}
		}
		return res.ID, "created"
	}

	L := func(code, desc, dr, cr string) gl.Line {
		return gl.Line{AccountCode: code, Description: desc, Debit: gl.Amount(dr), Credit: gl.Amount(cr)}
	}

	fmt.Printf("=== แผนเติมข้อมูล บริษัท %s สาขา %s ปี %s (apply=%v) ===\n", *company, *branch, *fiscal, *apply)
	fmt.Printf("บัญชีที่ใช้: เงินสด=%s ธนาคาร=%s ลูกหนี้=%s สินค้า=%s ภาษีซื้อ=%s เจ้าหนี้=%s ภาษีขาย=%s ทุน=%s รายได้=%s\n",
		cash, bank, ar, goods, inputVAT, ap, outputVAT, capital, salesVAT)
	fmt.Printf("บัญชีค้นชื่อ: ภาษีหัก=%s ค่าเช่า=%s ต้นทุนขาย=%s วัสดุสำนักงาน=%s\n", wht, rent, cogs, stationery)

	create := func(docno, date, book, desc, kind string, lines []gl.Line, details *gl.JournalDetails) {
		_, _ = createJournal(docno, date, book, desc, kind, lines, details)
	}

	partnerCustA := gl.SubledgerPartner{Code: "CUST-TH-001", Name: "บริษัท โครงการก่อสร้างรุ่งเรืองพัฒน์ จำกัด", TaxID: "0105558001001", IsCustomer: true, IsActive: true}
	partnerCustB := gl.SubledgerPartner{Code: "CUST-TH-002", Name: "ร้านวัสดุบ้านแข็งแรง", TaxID: "0105558001002", IsCustomer: true, IsActive: true}
	partnerSuppA := gl.SubledgerPartner{Code: "SUPP-TH-001", Name: "บริษัท ปูนกรุงไทย จำกัด", TaxID: "0105558002001", IsSupplier: true, IsActive: true}
	partnerSuppB := gl.SubledgerPartner{Code: "SUPP-TH-002", Name: "บริษัท เหล็กไทยพัฒนา จำกัด", TaxID: "0105558002002", IsSupplier: true, IsActive: true}
	kbank := gl.SubledgerBankAccount{Code: "BANK-KBANK", BankName: "ธนาคารกสิกรไทย", AccountNumber: "123-4-56789-0", AccountName: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", GLAccountCode: bank, Currency: "THB", IsActive: true}

	// 1) ยอดยกมาต้นงวด
	create("JV6901-S001", "2026-01-01", "JV", "บันทึกยอดยกมาต้นงวด บัญชีแยกประเภท ปี 2569", "opening", []gl.Line{
		L(cash, "ยอดยกมา เงินสดในมือ", "80000.00", "0.00"),
		L(bank, "ยอดยกมา เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย", "1650000.00", "0.00"),
		L(goods, "ยอดยกมา สินค้าสำเร็จรูป (ปูนซีเมนต์ เหล็กเส้น สีทาอาคาร)", "920000.00", "0.00"),
		L(capital, "ยอดยกมา ทุนเรือนหุ้น - หุ้นสามัญ", "0.00", "2650000.00"),
	}, nil)

	// 2) ซื้อปูนซีเมนต์เชื่อ + เอกสาร AP
	create("SV6901-S001", "2026-01-05", "SV", "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. 800 ถุง เป็นเงินเชื่อ", "manual", []gl.Line{
		L(goods, "รับสินค้าปูนซีเมนต์", "100000.00", "0.00"),
		L(inputVAT, "ภาษีซื้อ 7%", "7000.00", "0.00"),
		L(ap, "เจ้าหนี้การค้า บริษัท ปูนกรุงไทย จำกัด", "0.00", "107000.00"),
	}, &gl.JournalDetails{
		Partners:    []gl.SubledgerPartner{partnerSuppA},
		Documents:   []gl.SubledgerDocument{{ID: "AP-S001", Ledger: "ap", PartnerCode: partnerSuppA.Code, DocumentNo: "PT-2569-0112", Date: "2026-01-05", DueDate: "2026-02-04", BranchCode: *branch, Kind: 1, Side: 1, Amount: gl.Amount("107000.00"), Currency: "THB", ControlAccountCode: ap}},
		Allocations: []gl.SubledgerAllocation{{ID: "ALLOC-S001", Ledger: "ap", DocumentID: "AP-S001", LineNumber: 3, Amount: gl.Amount("107000.00")}},
	})

	// 3) ซื้อเหล็กเส้นเชื่อ + เอกสาร AP
	create("SV6901-S002", "2026-01-08", "SV", "ซื้อเหล็กเส้นกลม SZ12 120 เส้น เป็นเงินเชื่อ", "manual", []gl.Line{
		L(goods, "รับสินค้าเหล็กเส้น", "200000.00", "0.00"),
		L(inputVAT, "ภาษีซื้อ 7%", "14000.00", "0.00"),
		L(ap, "เจ้าหนี้การค้า บริษัท เหล็กไทยพัฒนา จำกัด", "0.00", "214000.00"),
	}, &gl.JournalDetails{
		Partners:    []gl.SubledgerPartner{partnerSuppB},
		Documents:   []gl.SubledgerDocument{{ID: "AP-S002", Ledger: "ap", PartnerCode: partnerSuppB.Code, DocumentNo: "ST-2569-0088", Date: "2026-01-08", DueDate: "2026-02-07", BranchCode: *branch, Kind: 1, Side: 1, Amount: gl.Amount("214000.00"), Currency: "THB", ControlAccountCode: ap}},
		Allocations: []gl.SubledgerAllocation{{ID: "ALLOC-S002", Ledger: "ap", DocumentID: "AP-S002", LineNumber: 3, Amount: gl.Amount("214000.00")}},
	})

	// 4) ขายเชื่อโครงการ A + เอกสาร AR
	create("UV6901-S001", "2026-01-12", "UV", "ขายวัสดุก่อสร้าง ปูนและสีทาอาคาร โครงการรุ่งเรืองพัฒน์ เป็นเงินเชื่อ", "manual", []gl.Line{
		L(ar, "ลูกหนี้การค้า บริษัท โครงการก่อสร้างรุ่งเรืองพัฒน์ จำกัด", "267500.00", "0.00"),
		L(salesVAT, "รายได้จากการขายสินค้า (ก่อนภาษี)", "0.00", "250000.00"),
		L(outputVAT, "ภาษีขาย 7%", "0.00", "17500.00"),
	}, &gl.JournalDetails{
		Partners:    []gl.SubledgerPartner{partnerCustA},
		Documents:   []gl.SubledgerDocument{{ID: "AR-S001", Ledger: "ar", PartnerCode: partnerCustA.Code, DocumentNo: "INV-2569-0201", Date: "2026-01-12", DueDate: "2026-02-11", BranchCode: *branch, Kind: 1, Side: 1, Amount: gl.Amount("267500.00"), Currency: "THB", ControlAccountCode: ar}},
		Allocations: []gl.SubledgerAllocation{{ID: "ALLOC-S003", Ledger: "ar", DocumentID: "AR-S001", LineNumber: 1, Amount: gl.Amount("267500.00")}},
	})

	// 5) ขายเชื่อร้านวัสดุ + เอกสาร AR
	create("UV6901-S002", "2026-01-16", "UV", "ขายเหล็กเส้นและปูนซีเมนต์ ร้านวัสดุบ้านแข็งแรง เป็นเงินเชื่อ", "manual", []gl.Line{
		L(ar, "ลูกหนี้การค้า ร้านวัสดุบ้านแข็งแรง", "160500.00", "0.00"),
		L(salesVAT, "รายได้จากการขายสินค้า (ก่อนภาษี)", "0.00", "150000.00"),
		L(outputVAT, "ภาษีขาย 7%", "0.00", "10500.00"),
	}, &gl.JournalDetails{
		Partners:    []gl.SubledgerPartner{partnerCustB},
		Documents:   []gl.SubledgerDocument{{ID: "AR-S002", Ledger: "ar", PartnerCode: partnerCustB.Code, DocumentNo: "INV-2569-0215", Date: "2026-01-16", DueDate: "2026-02-15", BranchCode: *branch, Kind: 1, Side: 1, Amount: gl.Amount("160500.00"), Currency: "THB", ControlAccountCode: ar}},
		Allocations: []gl.SubledgerAllocation{{ID: "ALLOC-S004", Ledger: "ar", DocumentID: "AR-S002", LineNumber: 1, Amount: gl.Amount("160500.00")}},
	})

	// 6) รับชำระบางส่วน 150,000 จากโครงการ A (โอนกสิกร) + settlement
	rvS001ID, rvS001Status := createJournal("RV6901-S001", "2026-01-20", "RV", "รับชำระบางส่วนใบ INV-2569-0201 โอนเข้าธนาคารกสิกรไทย", "manual", []gl.Line{
		L(bank, "รับเงินโอนเข้าธนาคารกสิกรไทย", "150000.00", "0.00"),
		L(ar, "ตัดลูกหนี้การค้าบางส่วน", "0.00", "150000.00"),
	}, &gl.JournalDetails{
		Partners:     []gl.SubledgerPartner{partnerCustA},
		BankAccounts: []gl.SubledgerBankAccount{kbank},
		BankLines:    []gl.SubledgerBankLine{{LineNumber: 1, BankAccountCode: "BANK-KBANK", Direction: 1}},
		Documents:    []gl.SubledgerDocument{{ID: "AR-SREC1", Ledger: "ar", PartnerCode: partnerCustA.Code, DocumentNo: "REC-2569-0051", Date: "2026-01-20", BranchCode: *branch, Kind: 2, Side: 2, Amount: gl.Amount("150000.00"), Currency: "THB", ControlAccountCode: ar}},
		Allocations:  []gl.SubledgerAllocation{{ID: "ALLOC-S005", Ledger: "ar", DocumentID: "AR-SREC1", LineNumber: 2, Amount: gl.Amount("150000.00")}},
		Settlements:  []gl.SubledgerSettlement{{ID: "SETTLE-S001", Ledger: "ar", PartnerCode: partnerCustA.Code, DebtDocumentID: "AR-S001", PaymentDocumentID: "AR-SREC1", Date: "2026-01-20", Amount: gl.Amount("150000.00")}},
	})
	if rvS001Status == "created" && *apply {
		recon := gl.Command{Resource: "journals", Action: "reconcile", RequestID: "seed-gl-screens-RV6901-S001-recon", ID: rvS001ID, Version: 2, Reason: "จับคู่หลักฐานรายการเงินเข้าจากธนาคาร",
			Journal: &gl.Journal{Details: &gl.JournalDetails{
				StatementLines: []gl.SubledgerStatementLine{{ID: "STMT-KB-001", BankAccountCode: "BANK-KBANK", SourceKey: "kbank-2569-01-20-000118", Date: "2026-01-20", Reference: "FT256901A0018", Description: "โอนเข้าจาก บจก.โครงการก่อสร้างรุ่งเรืองพัฒน์", Direction: 1, Amount: gl.Amount("150000.00")}},
				Matches:        []gl.SubledgerMatch{{ID: "MATCH-S001", StatementLineID: "STMT-KB-001", LineNumber: 1, Amount: gl.Amount("150000.00")}},
			}}}
		if _, err := store.Execute(ctx, scope, recon); err != nil {
			fatal("จับคู่ Statement RV-S001 ไม่ได้: %v", err)
		}
		fmt.Println("แผน: reconcile RV6901-S001 — Statement เข้า 150,000 จับคู่สำเร็จ")
	}

	// 7) จ่ายเจ้าหนี้ปูนกรุงไทย พร้อมหัก ณ ที่จ่าย 3% (ถ้าผังมีบัญชี WHT)
	var pvLines []gl.Line
	var pvDetails *gl.JournalDetails
	{
		if wht != "" {
			pvLines = []gl.Line{
				L(ap, "ตัดเจ้าหนี้การค้า บริษัท ปูนกรุงไทย จำกัด", "107000.00", "0.00"),
				L(bank, "จ่ายโอนออกจากธนาคารกสิกรไทย", "0.00", "103790.00"),
				L(wht, "ภาษีหัก ณ ที่จ่าย 3% ค้างนำส่ง", "0.00", "3210.00"),
			}
		} else {
			pvLines = []gl.Line{
				L(ap, "ตัดเจ้าหนี้การค้า บริษัท ปูนกรุงไทย จำกัด", "107000.00", "0.00"),
				L(bank, "จ่ายโอนออกจากธนาคารกสิกรไทย", "0.00", "107000.00"),
			}
		}
		pvDetails = &gl.JournalDetails{
			Partners:     []gl.SubledgerPartner{partnerSuppA},
			BankAccounts: []gl.SubledgerBankAccount{kbank},
			BankLines:    []gl.SubledgerBankLine{{LineNumber: 2, BankAccountCode: "BANK-KBANK", Direction: 2}},
			Documents:    []gl.SubledgerDocument{{ID: "AP-SPAY1", Ledger: "ap", PartnerCode: partnerSuppA.Code, DocumentNo: "PAY-2569-0021", Date: "2026-01-22", BranchCode: *branch, Kind: 2, Side: 2, Amount: gl.Amount("107000.00"), Currency: "THB", ControlAccountCode: ap}},
			Allocations:  []gl.SubledgerAllocation{{ID: "ALLOC-S006", Ledger: "ap", DocumentID: "AP-SPAY1", LineNumber: 1, Amount: gl.Amount("107000.00")}},
			Settlements:  []gl.SubledgerSettlement{{ID: "SETTLE-S002", Ledger: "ap", PartnerCode: partnerSuppA.Code, DebtDocumentID: "AP-S001", PaymentDocumentID: "AP-SPAY1", Date: "2026-01-22", Amount: gl.Amount("107000.00")}},
		}
	}
	pvS001ID, pvS001Status := createJournal("PV6901-S001", "2026-01-22", "PV", "จ่ายชำระหนี้ บจก.ปูนกรุงไทย ตามบิล PT-2569-0112 พร้อมหักภาษี ณ ที่จ่าย 3%", "manual", pvLines, pvDetails)
	if pvS001Status == "created" && *apply {
		recon := gl.Command{Resource: "journals", Action: "reconcile", RequestID: "seed-gl-screens-PV6901-S001-recon", ID: pvS001ID, Version: 2, Reason: "จับคู่หลักฐานรายการเงินออกจากธนาคาร",
			Journal: &gl.Journal{Details: &gl.JournalDetails{
				StatementLines: []gl.SubledgerStatementLine{{ID: "STMT-KB-002", BankAccountCode: "BANK-KBANK", SourceKey: "kbank-2569-01-22-000127", Date: "2026-01-22", Reference: "FT256901A0027", Description: "โอนออก ชำระ บจก.ปูนกรุงไทย", Direction: 2, Amount: gl.Amount("103790.00")}},
				Matches:        []gl.SubledgerMatch{{ID: "MATCH-S002", StatementLineID: "STMT-KB-002", LineNumber: 2, Amount: gl.Amount("103790.00")}},
			}}}
		if _, err := store.Execute(ctx, scope, recon); err != nil {
			fatal("จับคู่ Statement PV-S001 ไม่ได้: %v", err)
		}
		fmt.Println("แผน: reconcile PV6901-S001 — Statement ออก 103,790 จับคู่สำเร็จ")
	}

	// 8) ขายสดรับโอน
	create("UV6901-S003", "2026-02-03", "UV", "ขายวัสดุก่อสร้าง รับชำระโอนเข้าธนาคารกสิกรไทย", "manual", []gl.Line{
		L(bank, "รับเงินโอนเข้าบัญชีธนาคาร", "32100.00", "0.00"),
		L(salesVAT, "รายได้จากการขายสินค้า (ก่อนภาษี)", "0.00", "30000.00"),
		L(outputVAT, "ภาษีขาย 7%", "0.00", "2100.00"),
	}, &gl.JournalDetails{
		BankAccounts: []gl.SubledgerBankAccount{kbank},
		BankLines:    []gl.SubledgerBankLine{{LineNumber: 1, BankAccountCode: "BANK-KBANK", Direction: 1}},
	})

	// 9) จ่ายค่าเช่าหน้าร้าน + WHT 5%
	create("PV6901-S002", "2026-02-05", "PV", "จ่ายค่าเช่าหน้าร้านและคลังสินค้า ประจำเดือนมกราคม 2569 หักภาษี ณ ที่จ่าย 5%", "manual", func() []gl.Line {
		lines := []gl.Line{L(rent, "ค่าเช่าหน้าร้านและคลังสินค้า", "20000.00", "0.00"), L(bank, "จ่ายโอนออกจากธนาคารกสิกรไทย", "0.00", "19000.00")}
		if wht != "" {
			lines = append(lines, L(wht, "ภาษีหัก ณ ที่จ่าย 5% ค้างนำส่ง", "0.00", "1000.00"))
		} else {
			lines[1] = L(bank, "จ่ายโอนออกจากธนาคารกสิกรไทย", "0.00", "20000.00")
		}
		return lines
	}(), nil)

	// 10) จ่ายค่าเครื่องเขียนเงินสด
	create("PV6901-S003", "2026-02-08", "PV", "จ่ายค่าเครื่องเขียน แบบพิมพ์ และวัสดุสำนักงาน เงินสด", "manual", []gl.Line{
		L(stationery, "ค่าเครื่องเขียนและวัสดุสำนักงาน", "1500.00", "0.00"),
		L(cash, "จ่ายด้วยเงินสดในมือ", "0.00", "1500.00"),
	}, nil)

	// 11) ตัดต้นทุนขาย
	create("JV6901-S002", "2026-02-28", "JV", "ตัดต้นทุนขายวัสดุก่อสร้างประจำเดือนมกราคม 2569", "manual", []gl.Line{
		L(cogs, "ต้นทุนขายวัสดุก่อสร้าง", "240000.00", "0.00"),
		L(goods, "ตัดสินค้าสำเร็จรูปออกจากคลัง", "0.00", "240000.00"),
	}, nil)

	// 12) รับชำระเต็มจำนวนจากร้านวัสดุ + settlement เต็ม
	create("RV6901-S002", "2026-02-10", "RV", "รับชำระหนี้เต็มจำนวนใบ INV-2569-0215 โอนเข้าธนาคารกสิกรไทย", "manual", []gl.Line{
		L(bank, "รับเงินโอนเข้าธนาคารกสิกรไทย", "160500.00", "0.00"),
		L(ar, "ล้างลูกหนี้การค้า ร้านวัสดุบ้านแข็งแรง", "0.00", "160500.00"),
	}, &gl.JournalDetails{
		Partners:     []gl.SubledgerPartner{partnerCustB},
		BankAccounts: []gl.SubledgerBankAccount{kbank},
		BankLines:    []gl.SubledgerBankLine{{LineNumber: 1, BankAccountCode: "BANK-KBANK", Direction: 1}},
		Documents:    []gl.SubledgerDocument{{ID: "AR-SREC2", Ledger: "ar", PartnerCode: partnerCustB.Code, DocumentNo: "REC-2569-0064", Date: "2026-02-10", BranchCode: *branch, Kind: 2, Side: 2, Amount: gl.Amount("160500.00"), Currency: "THB", ControlAccountCode: ar}},
		Allocations:  []gl.SubledgerAllocation{{ID: "ALLOC-S007", Ledger: "ar", DocumentID: "AR-SREC2", LineNumber: 2, Amount: gl.Amount("160500.00")}},
		Settlements:  []gl.SubledgerSettlement{{ID: "SETTLE-S003", Ledger: "ar", PartnerCode: partnerCustB.Code, DebtDocumentID: "AR-S002", PaymentDocumentID: "AR-SREC2", Date: "2026-02-10", Amount: gl.Amount("160500.00")}},
	})

	// 13) งบประมาณประจำปี 2569 (ผ่าน masters)
	budgets := []struct{ code, name, acct, amount string }{
		{"BG-2569-SALES", "งบรายได้จากการขายสินค้า ปี 2569", salesVAT, "6000000.00"},
		{"BG-2569-COGS", "งบต้นทุนขายวัสดุก่อสร้าง ปี 2569", cogs, "3600000.00"},
		{"BG-2569-RENT", "งบค่าเช่าหน้าร้านและคลังสินค้า ปี 2569", rent, "240000.00"},
		{"BG-2569-STATIONERY", "งบค่าเครื่องเขียนและวัสดุสำนักงาน ปี 2569", stationery, "36000.00"},
	}
	for _, b := range budgets {
		fmt.Printf("แผน: งบประมาณ %s → %s (%s)\n", b.code, b.acct, b.amount)
		if !*apply {
			continue
		}
		var n int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gl_records WHERE company=$1 AND kind='budgets' AND payload->>'code'=$2`, *company, b.code).Scan(&n); err != nil {
			fatal("ตรวจงบประมาณไม่ได้: %v", err)
		}
		if n > 0 {
			continue
		}
		cmd := gl.Command{Resource: "budgets", Action: "create", RequestID: "seed-gl-screens-" + b.code,
			Master: &gl.Master{Kind: "budgets", Code: b.code, Name: b.name, IsActive: true, AccountCode: b.acct, FiscalYear: *fiscal, StartDate: "2026-01-01", EndDate: "2026-12-31", Amount: gl.Amount(b.amount)}}
		if _, err := store.Execute(ctx, scope, cmd); err != nil {
			fatal("สร้างงบประมาณ %s ไม่ได้: %v", b.code, err)
		}
	}

	// 14) รูปแบบการเชื่อมบัญชีตามสมุดรายวัน
	mappings := []struct {
		code, name, book string
		rules            []gl.MappingRule
	}{
		{"MAP-SV", "รูปแบบการเชื่อมบัญชี สมุดรายวันซื้อเชื่อ", "SV", []gl.MappingRule{
			{AccountCode: goods, Side: "debit", Source: "สินค้า"},
			{AccountCode: inputVAT, Side: "debit", Source: "ภาษีซื้อ"},
			{AccountCode: ap, Side: "credit", Source: "ยอดรวมเอกสาร"},
		}},
		{"MAP-UV", "รูปแบบการเชื่อมบัญชี สมุดรายวันขายเชื่อ", "UV", []gl.MappingRule{
			{AccountCode: ar, Side: "debit", Source: "ยอดรวมเอกสาร"},
			{AccountCode: salesVAT, Side: "credit", Source: "รายได้"},
			{AccountCode: outputVAT, Side: "credit", Source: "ภาษีขาย"},
		}},
		{"MAP-RV", "รูปแบบการเชื่อมบัญชี สมุดรายวันรับเงิน", "RV", []gl.MappingRule{
			{AccountCode: bank, Side: "debit", Source: "ยอดรับจริง"},
			{AccountCode: ar, Side: "credit", Source: "ยอดตัดลูกหนี้"},
		}},
	}
	for _, m := range mappings {
		fmt.Printf("แผน: รูปแบบการเชื่อม %s (%s) %d กฎ\n", m.code, m.book, len(m.rules))
		if !*apply {
			continue
		}
		var n int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gl_records WHERE company=$1 AND kind='mappings' AND payload->>'code'=$2`, *company, m.code).Scan(&n); err != nil {
			fatal("ตรวจรูปแบบการเชื่อมไม่ได้: %v", err)
		}
		if n > 0 {
			continue
		}
		cmd := gl.Command{Resource: "mappings", Action: "create", RequestID: "seed-gl-screens-" + m.code,
			Master: &gl.Master{Kind: "mappings", Code: m.code, Name: m.name, IsActive: true, BookCode: m.book, Rules: m.rules}}
		if _, err := store.Execute(ctx, scope, cmd); err != nil {
			fatal("สร้างรูปแบบการเชื่อม %s ไม่ได้: %v", m.code, err)
		}
	}

	// 15) กลุ่มบัญชีสินค้า
	groups := []struct{ code, name, item, cost, revenue string }{
		{"PAG-CEMENT", "กลุ่มบัญชีสินค้า ปูนซีเมนต์", goods, cogs, salesVAT},
		{"PAG-STEEL", "กลุ่มบัญชีสินค้า เหล็กเส้น", goods, cogs, salesVAT},
		{"PAG-PAINT", "กลุ่มบัญชีสินค้า สีทาอาคาร", goods, cogs, salesVAT},
	}
	for _, g := range groups {
		fmt.Printf("แผน: กลุ่มบัญชีสินค้า %s\n", g.code)
		if !*apply {
			continue
		}
		var n int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gl_records WHERE company=$1 AND kind='product-account-groups' AND payload->>'code'=$2`, *company, g.code).Scan(&n); err != nil {
			fatal("ตรวจกลุ่มบัญชีสินค้าไม่ได้: %v", err)
		}
		if n > 0 {
			continue
		}
		cmd := gl.Command{Resource: "product-account-groups", Action: "create", RequestID: "seed-gl-screens-" + g.code,
			Master: &gl.Master{Kind: "product-account-groups", Code: g.code, Name: g.name, IsActive: true, ItemAccount: g.item, CostAccount: g.cost, RevenueAccount: g.revenue}}
		if _, err := store.Execute(ctx, scope, cmd); err != nil {
			fatal("สร้างกลุ่มบัญชีสินค้า %s ไม่ได้: %v", g.code, err)
		}
	}

	fmt.Printf("=== จบแผนเติมข้อมูล (apply=%v) ===\n", *apply)
	if !*apply {
		fmt.Println("dry-run เท่านั้น ไม่มีการเขียนข้อมูล — รันซ้ำพร้อม -apply เพื่อเขียนจริง")
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "glseed: "+format+"\n", args...)
	os.Exit(1)
}
