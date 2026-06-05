package gentranspdf

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// labelCache - cache สำหรับเก็บ language codes ที่โหลดแล้ว
// โครงสร้าง languages.tsv: key\tth\ten\tcn\tja\tkm\tko\tlo\tmy\tvi
var (
	allLanguages    map[string]map[string]string         // code -> lang -> text
	labelCache      = make(map[string]map[string]string) // lang -> code -> text (cache)
	labelCacheLock  sync.RWMutex
	languagesLoaded bool
)

// loadAllLanguages - โหลดไฟล์ languages.tsv (รวมทุกภาษา)
// TSV format: key\tth\ten\tcn\tja\tkm\tko\tlo\tmy\tvi
func loadAllLanguages() {
	labelCacheLock.Lock()
	defer labelCacheLock.Unlock()

	if languagesLoaded {
		return
	}

	filePath := filepath.Join("language", "languages.tsv")
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	// อ่าน header
	if !scanner.Scan() {
		return
	}
	headers := strings.Split(scanner.Text(), "\t")

	allLanguages = make(map[string]map[string]string, 5000)

	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")
		if len(cols) < 2 {
			continue
		}
		key := cols[0]
		langs := make(map[string]string, len(headers)-1)
		for i := 1; i < len(headers) && i < len(cols); i++ {
			if cols[i] != "" {
				langs[headers[i]] = cols[i]
			}
		}
		allLanguages[key] = langs
	}

	languagesLoaded = true
}

// loadLanguageFile - โหลด language file และแปลงเป็น map[code]text สำหรับภาษาที่ระบุ
func loadLanguageFile(langCode string) map[string]string {
	// โหลด languages.tsv ถ้ายังไม่เคยโหลด
	if !languagesLoaded {
		loadAllLanguages()
	}

	// ตรวจสอบ cache ก่อน
	labelCacheLock.RLock()
	if cached, ok := labelCache[langCode]; ok {
		labelCacheLock.RUnlock()
		return cached
	}
	labelCacheLock.RUnlock()

	// แปลง language code เป็น key (ai:cn -> cn, ai:ja -> ja)
	key := langCode
	if strings.HasPrefix(langCode, "ai:") {
		key = strings.TrimPrefix(langCode, "ai:")
	}

	// ดึงข้อมูลจาก allLanguages และแปลงเป็น map[code]text สำหรับภาษานี้
	labelCacheLock.RLock()
	if allLanguages == nil {
		labelCacheLock.RUnlock()
		return nil
	}

	result := make(map[string]string)
	for code, langs := range allLanguages {
		if text, ok := langs[key]; ok && text != "" {
			result[code] = text
		}
	}
	labelCacheLock.RUnlock()

	// บันทึกลง cache
	labelCacheLock.Lock()
	labelCache[langCode] = result
	labelCacheLock.Unlock()

	return result
}

// getLangText - ดึงข้อความจาก language map พร้อม fallback
func getLangText(langMap map[string]string, code string, fallback string) string {
	if langMap == nil {
		return fallback
	}
	if text, ok := langMap[code]; ok && text != "" {
		return text
	}
	return fallback
}

// ThemeColors - สีหลักของธีม
type ThemeColors struct {
	Primary   string `json:"primary"`   // สีหลัก
	Secondary string `json:"secondary"` // สีรอง
	Accent    string `json:"accent"`    // สีเน้น
	Danger    string `json:"danger"`    // สีแจ้งเตือน
	Warning   string `json:"warning"`   // สีเตือน
}

// ThemeHeader - สีส่วนหัวเอกสาร
type ThemeHeader struct {
	TitleColor    string  `json:"titlecolor"`    // สีหัวเรื่อง
	SubtitleColor string  `json:"subtitlecolor"` // สีข้อความรอง
	LineColor     string  `json:"linecolor"`     // สีเส้นคั่น
	LineWidth     float64 `json:"linewidth"`     // ความหนาเส้น
}

// ThemeSection - สีส่วน section
type ThemeSection struct {
	LabelColor      string `json:"labelcolor"`      // สีหัวข้อ
	TextColor       string `json:"textcolor"`       // สีข้อความ
	BackgroundColor string `json:"backgroundcolor"` // สีพื้นหลัง
}

// ThemeTable - สีตาราง
type ThemeTable struct {
	HeaderBgColor   string  `json:"headerbgcolor"`   // สีพื้นหลังหัวตาราง
	HeaderTextColor string  `json:"headertextcolor"` // สีข้อความหัวตาราง
	RowTextColor    string  `json:"rowtextcolor"`    // สีข้อความแถว
	RowBgColor      string  `json:"rowbgcolor"`      // สีพื้นหลังแถว
	RowAltBgColor   string  `json:"rowaltbgcolor"`   // สีพื้นหลังแถวสลับ
	BorderColor     string  `json:"bordercolor"`     // สีเส้นขอบ
	BorderWidth     float64 `json:"borderwidth"`     // ความหนาเส้นขอบ
}

// ThemeSummary - สีส่วนสรุป
type ThemeSummary struct {
	BgColor            string `json:"bgcolor"`            // สีพื้นหลัง
	TextColor          string `json:"textcolor"`          // สีข้อความ
	HighlightBgColor   string `json:"highlightbgcolor"`   // สีพื้นหลังยอดรวม
	HighlightTextColor string `json:"highlighttextcolor"` // สีข้อความยอดรวม
	BorderColor        string `json:"bordercolor"`        // สีเส้นขอบ
}

// ThemeFooter - สีส่วนท้าย
type ThemeFooter struct {
	TextColor string `json:"textcolor"` // สีข้อความ
	LineColor string `json:"linecolor"` // สีเส้น
}

// Theme - โครงสร้างธีมทั้งหมด
type Theme struct {
	Name    string       `json:"name"`    // ชื่อธีม
	Preset  string       `json:"preset"`  // ใช้ preset theme
	Colors  ThemeColors  `json:"colors"`  // สีหลัก
	Header  ThemeHeader  `json:"header"`  // สีส่วนหัว
	Section ThemeSection `json:"section"` // สีส่วน section
	Table   ThemeTable   `json:"table"`   // สีตาราง
	Summary ThemeSummary `json:"summary"` // สีส่วนสรุป
	Footer  ThemeFooter  `json:"footer"`  // สีส่วนท้าย
}

// FontSizes - กำหนดขนาด font แยกตามส่วน
type FontSizes struct {
	Header  int `json:"header"`  // ขนาด font ส่วนหัวเอกสาร (default: 10)
	Detail  int `json:"detail"`  // ขนาด font ตารางรายละเอียด (default: 7)
	Summary int `json:"summary"` // ขนาด font ส่วนสรุป (default: 9)
	Footer  int `json:"footer"`  // ขนาด font ส่วนท้าย (default: 8)
}

// Labels - ข้อความแปลภาษาสำหรับ PDF
type Labels struct {
	// Header
	DocNo string
	Date  string

	// Company & Customer
	SellerInfo   string
	CustomerInfo string
	Code         string
	Phone        string

	// Table Headers
	No        string
	ItemCode  string
	Item      string
	Unit      string
	Qty       string
	UnitPrice string
	Amount    string
	NoItems   string

	// Summary
	Subtotal      string
	Discount      string
	AfterDiscount string
	Tax           string
	GrandTotal    string
	Currency      string
	Notes         string

	// Footer
	PrintedAt   string
	Page        string
	Terms       string
	CompanyName string

	// Signature
	PreparedBy string
	CheckedBy  string
	ApprovedBy string
}

// GetLabels - ดึง labels ตามภาษา (อ่านจาก JSON file ก่อน, fallback เป็นค่า default)
func GetLabels(language string) Labels {
	// โหลด language file
	langMap := loadLanguageFile(language)

	// สร้าง Labels จาก language file หรือ default values
	return Labels{
		// Header
		DocNo: getLangText(langMap, "docno", getDefaultLabel(language, "DocNo")),
		Date:  getLangText(langMap, "docdate", getDefaultLabel(language, "Date")),

		// Company & Customer
		SellerInfo:   getLangText(langMap, "seller_info", getDefaultLabel(language, "SellerInfo")),
		CustomerInfo: getLangText(langMap, "customer_info", getDefaultLabel(language, "CustomerInfo")),
		Code:         getLangText(langMap, "code", getDefaultLabel(language, "Code")),
		Phone:        getLangText(langMap, "tel", getDefaultLabel(language, "Phone")),

		// Table Headers
		No:        getLangText(langMap, "no", getDefaultLabel(language, "No")),
		ItemCode:  getLangText(langMap, "itemcode", getDefaultLabel(language, "ItemCode")),
		Item:      getLangText(langMap, "itemname", getDefaultLabel(language, "Item")),
		Unit:      getLangText(langMap, "unit", getDefaultLabel(language, "Unit")),
		Qty:       getLangText(langMap, "qty", getDefaultLabel(language, "Qty")),
		UnitPrice: getLangText(langMap, "unit_price", getDefaultLabel(language, "UnitPrice")),
		Amount:    getLangText(langMap, "amount", getDefaultLabel(language, "Amount")),
		NoItems:   getLangText(langMap, "no_items", getDefaultLabel(language, "NoItems")),

		// Summary
		Subtotal:      getLangText(langMap, "subtotal", getDefaultLabel(language, "Subtotal")),
		Discount:      getLangText(langMap, "discount", getDefaultLabel(language, "Discount")),
		AfterDiscount: getLangText(langMap, "after_discount", getDefaultLabel(language, "AfterDiscount")),
		Tax:           getLangText(langMap, "vat", getDefaultLabel(language, "Tax")),
		GrandTotal:    getLangText(langMap, "grand_total", getDefaultLabel(language, "GrandTotal")),
		Currency:      getLangText(langMap, "currency", getDefaultLabel(language, "Currency")),
		Notes:         getLangText(langMap, "note", getDefaultLabel(language, "Notes")),

		// Footer
		PrintedAt:   getLangText(langMap, "print", getDefaultLabel(language, "PrintedAt")),
		Page:        getLangText(langMap, "page", getDefaultLabel(language, "Page")),
		Terms:       getLangText(langMap, "terms", getDefaultLabel(language, "Terms")),
		CompanyName: getLangText(langMap, "company_name", getDefaultLabel(language, "CompanyName")),

		// Signature
		PreparedBy: getLangText(langMap, "prepared_by", getDefaultLabel(language, "PreparedBy")),
		CheckedBy:  getLangText(langMap, "checked_by", getDefaultLabel(language, "CheckedBy")),
		ApprovedBy: getLangText(langMap, "approved_by", getDefaultLabel(language, "ApprovedBy")),
	}
}

// getDefaultLabel - ดึง default label ตามภาษา (ใช้เมื่อไม่พบใน JSON file)
func getDefaultLabel(language string, labelKey string) string {
	defaults := getDefaultLabelsMap(language)
	if val, ok := defaults[labelKey]; ok {
		return val
	}
	return labelKey
}

// getDefaultLabelsMap - ค่า default labels สำหรับแต่ละภาษา
func getDefaultLabelsMap(language string) map[string]string {
	switch language {
	case "en":
		return map[string]string{
			"DocNo": "No:", "Date": "Date:",
			"SellerInfo": "Seller Information", "CustomerInfo": "Customer Information",
			"Code": "Code:", "Phone": "Tel:",
			"No": "No.", "ItemCode": "Item Code", "Item": "Description",
			"Unit": "Unit", "Qty": "Qty", "UnitPrice": "Unit Price",
			"Amount": "Amount", "NoItems": "No items",
			"Subtotal": "Subtotal", "Discount": "Discount", "AfterDiscount": "After Discount",
			"Tax": "VAT 7%", "GrandTotal": "Grand Total", "Currency": "Baht", "Notes": "Notes",
			"PrintedAt": "Printed:", "Page": "Page",
			"Terms":       "Terms: Payment within 30 days | Price includes VAT",
			"CompanyName": "Sample Company Ltd.",
			"PreparedBy":  "Prepared By", "CheckedBy": "Checked By", "ApprovedBy": "Approved By",
		}
	case "ai:cn":
		return map[string]string{
			"DocNo": "编号:", "Date": "日期:",
			"SellerInfo": "卖方信息", "CustomerInfo": "客户信息",
			"Code": "代码:", "Phone": "电话:",
			"No": "序号", "ItemCode": "商品编码", "Item": "商品名称",
			"Unit": "单位", "Qty": "数量", "UnitPrice": "单价",
			"Amount": "金额", "NoItems": "无商品",
			"Subtotal": "小计", "Discount": "折扣", "AfterDiscount": "折后金额",
			"Tax": "增值税 7%", "GrandTotal": "总计", "Currency": "泰铢", "Notes": "备注",
			"PrintedAt": "打印时间:", "Page": "页",
			"Terms":       "条款: 30天内付款 | 价格含增值税",
			"CompanyName": "示例公司",
			"PreparedBy":  "制单人", "CheckedBy": "审核人", "ApprovedBy": "批准人",
		}
	case "ai:ja":
		return map[string]string{
			"DocNo": "番号:", "Date": "日付:",
			"SellerInfo": "販売者情報", "CustomerInfo": "顧客情報",
			"Code": "コード:", "Phone": "電話:",
			"No": "番号", "ItemCode": "商品コード", "Item": "品名",
			"Unit": "単位", "Qty": "数量", "UnitPrice": "単価",
			"Amount": "金額", "NoItems": "商品なし",
			"Subtotal": "小計", "Discount": "割引", "AfterDiscount": "割引後",
			"Tax": "消費税 7%", "GrandTotal": "合計", "Currency": "バーツ", "Notes": "備考",
			"PrintedAt": "印刷日時:", "Page": "ページ",
			"Terms":       "条件: 30日以内にお支払い | 価格は税込み",
			"CompanyName": "サンプル会社",
			"PreparedBy":  "作成者", "CheckedBy": "確認者", "ApprovedBy": "承認者",
		}
	case "ai:ko":
		return map[string]string{
			"DocNo": "번호:", "Date": "날짜:",
			"SellerInfo": "판매자 정보", "CustomerInfo": "고객 정보",
			"Code": "코드:", "Phone": "전화:",
			"No": "번호", "ItemCode": "상품코드", "Item": "품명",
			"Unit": "단위", "Qty": "수량", "UnitPrice": "단가",
			"Amount": "금액", "NoItems": "상품 없음",
			"Subtotal": "소계", "Discount": "할인", "AfterDiscount": "할인 후",
			"Tax": "부가세 7%", "GrandTotal": "총합계", "Currency": "바트", "Notes": "비고",
			"PrintedAt": "인쇄:", "Page": "페이지",
			"Terms":       "조건: 30일 이내 결제 | 가격은 부가세 포함",
			"CompanyName": "샘플 회사",
			"PreparedBy":  "작성자", "CheckedBy": "확인자", "ApprovedBy": "승인자",
		}
	case "ai:vi":
		return map[string]string{
			"DocNo": "Số:", "Date": "Ngày:",
			"SellerInfo": "Thông tin người bán", "CustomerInfo": "Thông tin khách hàng",
			"Code": "Mã:", "Phone": "ĐT:",
			"No": "STT", "ItemCode": "Mã hàng", "Item": "Tên hàng",
			"Unit": "ĐVT", "Qty": "SL", "UnitPrice": "Đơn giá",
			"Amount": "Thành tiền", "NoItems": "Không có hàng",
			"Subtotal": "Cộng tiền", "Discount": "Chiết khấu", "AfterDiscount": "Sau chiết khấu",
			"Tax": "VAT 7%", "GrandTotal": "Tổng cộng", "Currency": "Baht", "Notes": "Ghi chú",
			"PrintedAt": "In ngày:", "Page": "Trang",
			"Terms":       "Điều khoản: Thanh toán trong 30 ngày | Giá đã bao gồm VAT",
			"CompanyName": "Công ty mẫu",
			"PreparedBy":  "Người lập", "CheckedBy": "Người kiểm tra", "ApprovedBy": "Người duyệt",
		}
	case "ai:lo":
		return map[string]string{
			"DocNo": "ເລກທີ:", "Date": "ວັນທີ:",
			"SellerInfo": "ຂໍ້ມູນຜູ້ຂາຍ", "CustomerInfo": "ຂໍ້ມູນລູກຄ້າ",
			"Code": "ລະຫັດ:", "Phone": "ໂທ:",
			"No": "ລຳດັບ", "ItemCode": "ລະຫັດສິນຄ້າ", "Item": "ລາຍການ",
			"Unit": "ຫົວໜ່ວຍ", "Qty": "ຈຳນວນ", "UnitPrice": "ລາຄາ/ໜ່ວຍ",
			"Amount": "ຈຳນວນເງິນ", "NoItems": "ບໍ່ມີລາຍການສິນຄ້າ",
			"Subtotal": "ລວມເງິນ", "Discount": "ສ່ວນຫຼຸດ", "AfterDiscount": "ຫຼັງຫັກສ່ວນຫຼຸດ",
			"Tax": "ພາສີ 7%", "GrandTotal": "ຍອດລວມທັງໝົດ", "Currency": "ບາດ", "Notes": "ໝາຍເຫດ",
			"PrintedAt": "ພິມ:", "Page": "ໜ້າ",
			"Terms":       "ເງື່ອນໄຂ: ຊຳລະເງິນພາຍໃນ 30 ວັນ | ລາຄາລວມ VAT ແລ້ວ",
			"CompanyName": "ບໍລິສັດ ຕົວຢ່າງ ຈຳກັດ",
			"PreparedBy":  "ຜູ້ຈັດທຳ", "CheckedBy": "ຜູ້ກວດສອບ", "ApprovedBy": "ຜູ້ອະນຸມັດ",
		}
	case "ai:my":
		return map[string]string{
			"DocNo": "နံပါတ်:", "Date": "ရက်စွဲ:",
			"SellerInfo": "ရောင်းသူအချက်အလက်", "CustomerInfo": "ဝယ်သူအချက်အလက်",
			"Code": "ကုဒ်:", "Phone": "ဖုန်း:",
			"No": "စဥ်", "ItemCode": "ကုန်ပစ္စည်းကုဒ်", "Item": "အကြောင်းအရာ",
			"Unit": "ယူနစ်", "Qty": "အရေအတွက်", "UnitPrice": "တစ်ခုစျေး",
			"Amount": "ငွေပမာဏ", "NoItems": "ကုန်ပစ္စည်းမရှိပါ",
			"Subtotal": "စုစုပေါင်း", "Discount": "လျှော့စျေး", "AfterDiscount": "လျှော့ပြီးနောက်",
			"Tax": "အခွန် 7%", "GrandTotal": "စုစုပေါင်းတန်ဖိုး", "Currency": "ဘတ်", "Notes": "မှတ်ချက်",
			"PrintedAt": "ပရင့်:", "Page": "စာမျက်နှာ",
			"Terms":       "စည်းကမ်း: ရက် 30 အတွင်းငွေပေးချေပါ | VAT ပါဝင်ပြီး",
			"CompanyName": "နမူနာကုမ္ပဏီ",
			"PreparedBy":  "ပြုလုပ်သူ", "CheckedBy": "စစ်ဆေးသူ", "ApprovedBy": "အတည်ပြုသူ",
		}
	case "ai:km":
		return map[string]string{
			"DocNo": "លេខ:", "Date": "កាលបរិច្ឆេទ:",
			"SellerInfo": "ព័ត៌មានអ្នកលក់", "CustomerInfo": "ព័ត៌មានអតិថិជន",
			"Code": "កូដ:", "Phone": "ទូរស័ព្ទ:",
			"No": "ល.រ", "ItemCode": "កូដទំនិញ", "Item": "បរិយាយ",
			"Unit": "ឯកតា", "Qty": "បរិមាណ", "UnitPrice": "តម្លៃឯកតា",
			"Amount": "ចំនួនទឹកប្រាក់", "NoItems": "គ្មានទំនិញ",
			"Subtotal": "សរុបរង", "Discount": "បញ្ចុះតម្លៃ", "AfterDiscount": "បន្ទាប់ពីបញ្ចុះតម្លៃ",
			"Tax": "អាករ 7%", "GrandTotal": "សរុបរួម", "Currency": "បាត", "Notes": "កំណត់ចំណាំ",
			"PrintedAt": "បោះពុម្ព:", "Page": "ទំព័រ",
			"Terms":       "លក្ខខណ្ឌ: ទូទាត់ក្នុងរយៈពេល 30 ថ្ងៃ | តម្លៃរួមអាករ",
			"CompanyName": "ក្រុមហ៊ុនគំរូ",
			"PreparedBy":  "អ្នករៀបចំ", "CheckedBy": "អ្នកពិនិត្យ", "ApprovedBy": "អ្នកអនុម័ត",
		}
	default: // "th" - Thai
		return map[string]string{
			"DocNo": "เลขที่:", "Date": "วันที่:",
			"SellerInfo": "ข้อมูลผู้ขาย", "CustomerInfo": "ข้อมูลลูกค้า",
			"Code": "รหัส:", "Phone": "โทร:",
			"No": "ลำดับ", "ItemCode": "รหัสสินค้า", "Item": "รายการ",
			"Unit": "หน่วย", "Qty": "จำนวน", "UnitPrice": "ราคา/หน่วย",
			"Amount": "จำนวนเงิน", "NoItems": "ไม่มีรายการสินค้า",
			"Subtotal": "รวมเงิน", "Discount": "ส่วนลด", "AfterDiscount": "หลังหักส่วนลด",
			"Tax": "ภาษี 7%", "GrandTotal": "ยอดรวมทั้งสิ้น", "Currency": "บาท", "Notes": "หมายเหตุ",
			"PrintedAt": "พิมพ์:", "Page": "หน้า",
			"Terms":       "เงื่อนไข: ชำระเงินภายใน 30 วัน | ราคารวม VAT แล้ว",
			"CompanyName": "บริษัท ตัวอย่าง จำกัด",
			"PreparedBy":  "ผู้จัดทำ", "CheckedBy": "ผู้ตรวจสอบ", "ApprovedBy": "ผู้อนุมัติ",
		}
	}
}

// Template - โครงสร้างเทมเพลตสำหรับ PDF
// HeaderLayouts: standard, centered, minimal, split, banner, leftAligned, twoColumn
// TableStyles: bordered, striped, minimal, clean, modern, elegant, compact
// SummaryLayouts: right, full, boxed, minimal, twoColumn, highlighted
// FooterStyles: standard, detailed, minimal, branded, withSignature, withTerms
// BorderStyles: none, solid, dashed, double, rounded
// LogoStyles: normal, large, small, circular, withBackground
type Template struct {
	ID                   string  `json:"id"`                   // รหัสเทมเพลต
	Name                 string  `json:"name"`                 // ชื่อเทมเพลต
	HeaderLayout         string  `json:"headerlayout"`         // รูปแบบหัวเอกสาร
	TableStyle           string  `json:"tablestyle"`           // รูปแบบตาราง
	SummaryLayout        string  `json:"summarylayout"`        // รูปแบบส่วนสรุป
	FooterStyle          string  `json:"footerstyle"`          // รูปแบบท้ายเอกสาร
	ShowLogo             bool    `json:"showlogo"`             // แสดงโลโก้
	LogoStyle            string  `json:"logostyle"`            // รูปแบบโลโก้
	ShowWatermark        bool    `json:"showwatermark"`        // แสดงลายน้ำ
	WatermarkText        string  `json:"watermarktext"`        // ข้อความลายน้ำ
	ShowBorder           bool    `json:"showborder"`           // แสดงกรอบ
	BorderStyle          string  `json:"borderstyle"`          // รูปแบบกรอบ
	ShowDocTitle         bool    `json:"showdoctitle"`         // แสดงหัวเอกสาร
	ShowCompanyInfo      bool    `json:"showcompanyinfo"`      // แสดงข้อมูลบริษัท
	ShowCustomerInfo     bool    `json:"showcustomerinfo"`     // แสดงข้อมูลลูกค้า
	ShowPaymentInfo      bool    `json:"showpaymentinfo"`      // แสดงข้อมูลการชำระเงิน
	ShowNotes            bool    `json:"shownotes"`            // แสดงหมายเหตุ
	ShowSignature        bool    `json:"showsignature"`        // แสดงลายเซ็น
	HeaderSpacing        float64 `json:"headerspacing"`        // ระยะห่างหัวเอกสาร
	TableSpacing         float64 `json:"tablespacing"`         // ระยะห่างตาราง
	SummarySpacing       float64 `json:"summaryspacing"`       // ระยะห่างส่วนสรุป
	MarginTop            float64 `json:"margintop"`            // ระยะขอบบน
	MarginBottom         float64 `json:"marginbottom"`         // ระยะขอบล่าง
	MarginLeft           float64 `json:"marginleft"`           // ระยะขอบซ้าย
	MarginRight          float64 `json:"marginright"`          // ระยะขอบขวา
	ShowRowNumber        bool    `json:"showrownumber"`        // แสดงลำดับแถว
	ShowUnitPrice        bool    `json:"showunitprice"`        // แสดงราคาต่อหน่วย
	ShowDiscount         bool    `json:"showdiscount"`         // แสดงส่วนลด
	ShowTax              bool    `json:"showtax"`              // แสดงภาษี
	AlternateRowColor    bool    `json:"alternaterowcolor"`    // สลับสีแถว
	ShowHeaderLine       bool    `json:"showheaderline"`       // แสดงเส้นหัวเอกสาร
	HeaderLineWidth      float64 `json:"headerlinewidth"`      // ความหนาเส้นหัวเอกสาร
	ShowHeaderBackground bool    `json:"showheaderbackground"` // แสดงพื้นหลังหัวตาราง
	ShowSubtotal         bool    `json:"showsubtotal"`         // แสดงยอดรวมย่อย
	ShowTotalDiscount    bool    `json:"showtotaldiscount"`    // แสดงส่วนลดรวม
	ShowTotalTax         bool    `json:"showtotaltax"`         // แสดงภาษีรวม
	HighlightTotal       bool    `json:"highlighttotal"`       // เน้นยอดรวม
}

// TableStyleConfig - การตั้งค่าสไตล์ตาราง
type TableStyleConfig struct {
	HeaderBorder     string  // border สำหรับ header ("1", "TB", "B", "")
	RowBorder        string  // border สำหรับ row ("1", "TB", "B", "")
	HeaderFill       bool    // มี fill color สำหรับ header หรือไม่
	RowFill          bool    // มี fill color สำหรับ row หรือไม่
	HeaderBold       bool    // header ตัวหนาหรือไม่
	BorderWidth      float64 // ความหนาเส้นขอบ
	RowHeight        float64 // ความสูงแถวขั้นต่ำ
	HeaderHeight     float64 // ความสูง header
	CellPadding      float64 // padding ในเซลล์
	ShowTopBorder    bool    // แสดงเส้นบนสุด
	ShowBottomBorder bool    // แสดงเส้นล่างสุด
}

// GetTableStyleConfig - ดึงการตั้งค่า Table Style ตามชื่อ (ปรับให้ประหยัดกระดาษ)
func GetTableStyleConfig(styleName string) TableStyleConfig {
	configs := map[string]TableStyleConfig{
		"bordered": {
			HeaderBorder:     "1",
			RowBorder:        "1",
			HeaderFill:       true,
			RowFill:          true,
			HeaderBold:       true,
			BorderWidth:      0.1,
			RowHeight:        4.5,
			HeaderHeight:     5.0,
			CellPadding:      0.3,
			ShowTopBorder:    true,
			ShowBottomBorder: true,
		},
		"striped": {
			HeaderBorder:     "1",
			RowBorder:        "1",
			HeaderFill:       true,
			RowFill:          true,
			HeaderBold:       true,
			BorderWidth:      0.1,
			RowHeight:        4.5,
			HeaderHeight:     5.0,
			CellPadding:      0.3,
			ShowTopBorder:    true,
			ShowBottomBorder: true,
		},
		"minimal": {
			HeaderBorder:     "TB",
			RowBorder:        "",
			HeaderFill:       false,
			RowFill:          false,
			HeaderBold:       true,
			BorderWidth:      0.2,
			RowHeight:        4.0,
			HeaderHeight:     4.5,
			CellPadding:      0.2,
			ShowTopBorder:    true,
			ShowBottomBorder: true,
		},
		"clean": {
			HeaderBorder:     "TB",
			RowBorder:        "B",
			HeaderFill:       true,
			RowFill:          false,
			HeaderBold:       true,
			BorderWidth:      0.1,
			RowHeight:        4.0,
			HeaderHeight:     4.5,
			CellPadding:      0.2,
			ShowTopBorder:    true,
			ShowBottomBorder: true,
		},
		"modern": {
			HeaderBorder:     "1",
			RowBorder:        "LR",
			HeaderFill:       true,
			RowFill:          true,
			HeaderBold:       true,
			BorderWidth:      0.05,
			RowHeight:        4.5,
			HeaderHeight:     5.0,
			CellPadding:      0.3,
			ShowTopBorder:    false,
			ShowBottomBorder: true,
		},
		"elegant": {
			HeaderBorder:     "TB",
			RowBorder:        "B",
			HeaderFill:       false,
			RowFill:          false,
			HeaderBold:       true,
			BorderWidth:      0.15,
			RowHeight:        4.5,
			HeaderHeight:     5.0,
			CellPadding:      0.3,
			ShowTopBorder:    true,
			ShowBottomBorder: true,
		},
		"compact": {
			HeaderBorder:     "1",
			RowBorder:        "1",
			HeaderFill:       true,
			RowFill:          true,
			HeaderBold:       true,
			BorderWidth:      0.05,
			RowHeight:        3.5,
			HeaderHeight:     4.0,
			CellPadding:      0.1,
			ShowTopBorder:    true,
			ShowBottomBorder: true,
		},
	}

	if config, ok := configs[styleName]; ok {
		return config
	}
	return configs["bordered"] // default
}

// GenPDFPayload - payload สำหรับการสร้าง PDF จาก MongoDB document
type GenPDFPayload struct {
	HoldingCode string     `json:"holdingcode"`
	Collection  string     `json:"collection"`
	DocNo       string     `json:"docno"` // เลขที่เอกสาร
	Title       string     `json:"title"`
	PageSize    string     `json:"pagesize"`    // A4, A3, Letter, Legal
	Orientation string     `json:"orientation"` // P (Portrait), L (Landscape)
	FontSize    int        `json:"fontsize"`    // ขนาดฟอนต์พื้นฐาน (8-16) - ใช้เป็น fallback
	FontSizes   *FontSizes `json:"fontsizes"`   // ขนาด font แยกตามส่วน
	FontFamily  string     `json:"fontfamily"`  // ชื่อ font family (Sarabun, Kanit, Prompt, etc.) - default: GoNotoCurrent
	LineSpacing float64    `json:"linespacing"` // ระยะห่างระหว่างบรรทัด (1.0 = ปกติ, 1.5 = 1.5 เท่า, 2.0 = 2 เท่า) - default: 1.0
	ColorMode   *bool      `json:"colormode"`   // true = สี, false = ขาวดำ (default: true)
	DateFormat  string     `json:"dateformat"`  // รูปแบบวันที่: DD/MM/YYYY, DD/MM/BBBB, DD MMM YY, DD MMMM BBBB, etc.
	Language    string     `json:"language"`    // ภาษา: th, en, ai:cn, ai:ja, ai:ko, ai:lo, ai:km, ai:my, ai:vi
	ThemeName   string     `json:"themename"`   // ชื่อ preset theme (modern-blue, professional, etc.)
	TemplateID  string     `json:"templateid"`  // รหัส preset template (standard, modern, etc.)
	Theme       *Theme     `json:"theme"`       // ธีมแบบ custom (ถ้าต้องการกำหนดเอง)
	Template    *Template  `json:"template"`    // เทมเพลตแบบ custom (ถ้าต้องการกำหนดเอง)

	// Multi-currency display option
	ShowDualCurrency *bool `json:"showdualcurrency"` // true = แสดง 2 สกุลเงิน (default), false = แสดงเฉพาะ doc currency

	// Preview mode - รับข้อมูลจาก payload โดยตรง (ไม่ดึงจาก DB)
	IsPreview bool                   `json:"ispreview"` // true = แสดง watermark "Preview"
	Document  map[string]interface{} `json:"document"`  // ข้อมูลเอกสารสำหรับ preview (กรณียังไม่ save)

	// User info for history tracking
	PrintedBy string `json:"printedby"` // ผู้พิมพ์ (optional)
}

// GetPresetTheme - ดึง preset theme ตามชื่อ
func GetPresetTheme(name string) Theme {
	presets := map[string]Theme{
		"modern-blue": {
			Name:    "modern-blue",
			Colors:  ThemeColors{Primary: "#2980B9", Secondary: "#3498DB", Accent: "#27AE60", Danger: "#E74C3C", Warning: "#F39C12"},
			Header:  ThemeHeader{TitleColor: "#2980B9", SubtitleColor: "#646464", LineColor: "#2980B9", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#2980B9", TextColor: "#3C3C3C", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#2980B9", HeaderTextColor: "#FFFFFF", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#F8FAFC", BorderColor: "#DCDCDC", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#F8FAFC", TextColor: "#000000", HighlightBgColor: "#2980B9", HighlightTextColor: "#FFFFFF", BorderColor: "#DCDCDC"},
			Footer:  ThemeFooter{TextColor: "#969696", LineColor: "#DCDCDC"},
		},
		"professional": {
			Name:    "professional",
			Colors:  ThemeColors{Primary: "#34495E", Secondary: "#5D6D7E", Accent: "#1ABC9C", Danger: "#E74C3C", Warning: "#F39C12"},
			Header:  ThemeHeader{TitleColor: "#34495E", SubtitleColor: "#7F8C8D", LineColor: "#34495E", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#34495E", TextColor: "#2C3E50", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#34495E", HeaderTextColor: "#FFFFFF", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#F4F6F7", BorderColor: "#BDC3C7", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#F4F6F7", TextColor: "#000000", HighlightBgColor: "#34495E", HighlightTextColor: "#FFFFFF", BorderColor: "#BDC3C7"},
			Footer:  ThemeFooter{TextColor: "#95A5A6", LineColor: "#BDC3C7"},
		},
		"nature": {
			Name:    "nature",
			Colors:  ThemeColors{Primary: "#27AE60", Secondary: "#2ECC71", Accent: "#3498DB", Danger: "#E74C3C", Warning: "#F39C12"},
			Header:  ThemeHeader{TitleColor: "#27AE60", SubtitleColor: "#7F8C8D", LineColor: "#27AE60", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#27AE60", TextColor: "#2C3E50", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#27AE60", HeaderTextColor: "#FFFFFF", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#F0FFF4", BorderColor: "#A9DFBF", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#F0FFF4", TextColor: "#000000", HighlightBgColor: "#27AE60", HighlightTextColor: "#FFFFFF", BorderColor: "#A9DFBF"},
			Footer:  ThemeFooter{TextColor: "#95A5A6", LineColor: "#A9DFBF"},
		},
		"warm": {
			Name:    "warm",
			Colors:  ThemeColors{Primary: "#E67E22", Secondary: "#F39C12", Accent: "#27AE60", Danger: "#E74C3C", Warning: "#F1C40F"},
			Header:  ThemeHeader{TitleColor: "#E67E22", SubtitleColor: "#7F8C8D", LineColor: "#E67E22", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#E67E22", TextColor: "#2C3E50", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#E67E22", HeaderTextColor: "#FFFFFF", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#FEF9E7", BorderColor: "#F5CBA7", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#FEF9E7", TextColor: "#000000", HighlightBgColor: "#E67E22", HighlightTextColor: "#FFFFFF", BorderColor: "#F5CBA7"},
			Footer:  ThemeFooter{TextColor: "#95A5A6", LineColor: "#F5CBA7"},
		},
		"classic-red": {
			Name:    "classic-red",
			Colors:  ThemeColors{Primary: "#C0392B", Secondary: "#E74C3C", Accent: "#27AE60", Danger: "#922B21", Warning: "#F39C12"},
			Header:  ThemeHeader{TitleColor: "#C0392B", SubtitleColor: "#7F8C8D", LineColor: "#C0392B", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#C0392B", TextColor: "#2C3E50", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#C0392B", HeaderTextColor: "#FFFFFF", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#FDEDEC", BorderColor: "#F5B7B1", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#FDEDEC", TextColor: "#000000", HighlightBgColor: "#C0392B", HighlightTextColor: "#FFFFFF", BorderColor: "#F5B7B1"},
			Footer:  ThemeFooter{TextColor: "#95A5A6", LineColor: "#F5B7B1"},
		},
		"purple": {
			Name:    "purple",
			Colors:  ThemeColors{Primary: "#8E44AD", Secondary: "#9B59B6", Accent: "#3498DB", Danger: "#E74C3C", Warning: "#F39C12"},
			Header:  ThemeHeader{TitleColor: "#8E44AD", SubtitleColor: "#7F8C8D", LineColor: "#8E44AD", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#8E44AD", TextColor: "#2C3E50", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#8E44AD", HeaderTextColor: "#FFFFFF", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#F5EEF8", BorderColor: "#D7BDE2", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#F5EEF8", TextColor: "#000000", HighlightBgColor: "#8E44AD", HighlightTextColor: "#FFFFFF", BorderColor: "#D7BDE2"},
			Footer:  ThemeFooter{TextColor: "#95A5A6", LineColor: "#D7BDE2"},
		},
		"teal": {
			Name:    "teal",
			Colors:  ThemeColors{Primary: "#16A085", Secondary: "#1ABC9C", Accent: "#3498DB", Danger: "#E74C3C", Warning: "#F39C12"},
			Header:  ThemeHeader{TitleColor: "#16A085", SubtitleColor: "#7F8C8D", LineColor: "#16A085", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#16A085", TextColor: "#2C3E50", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#16A085", HeaderTextColor: "#FFFFFF", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#E8F8F5", BorderColor: "#A3E4D7", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#E8F8F5", TextColor: "#000000", HighlightBgColor: "#16A085", HighlightTextColor: "#FFFFFF", BorderColor: "#A3E4D7"},
			Footer:  ThemeFooter{TextColor: "#95A5A6", LineColor: "#A3E4D7"},
		},
		"dark": {
			Name:    "dark",
			Colors:  ThemeColors{Primary: "#2C3E50", Secondary: "#34495E", Accent: "#1ABC9C", Danger: "#E74C3C", Warning: "#F39C12"},
			Header:  ThemeHeader{TitleColor: "#2C3E50", SubtitleColor: "#7F8C8D", LineColor: "#2C3E50", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#2C3E50", TextColor: "#2C3E50", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#2C3E50", HeaderTextColor: "#FFFFFF", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#EBEDEF", BorderColor: "#ABB2B9", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#EBEDEF", TextColor: "#000000", HighlightBgColor: "#2C3E50", HighlightTextColor: "#FFFFFF", BorderColor: "#ABB2B9"},
			Footer:  ThemeFooter{TextColor: "#95A5A6", LineColor: "#ABB2B9"},
		},
	}

	if theme, ok := presets[name]; ok {
		return theme
	}
	return presets["modern-blue"] // default
}

// GetEffectiveTheme - ดึง theme ที่จะใช้งานจริง (รวม preset + custom)
func GetEffectiveTheme(payload GenPDFPayload) Theme {
	// ถ้า colormode = false (ส่งมาเป็น false โดยเฉพาะ) ใช้ theme ขาวดำ
	// ถ้า colormode = nil หรือ true ใช้ theme สี (default เป็นสี)
	if payload.ColorMode != nil && !*payload.ColorMode {
		return Theme{
			Name:    "monochrome",
			Colors:  ThemeColors{Primary: "#000000", Secondary: "#333333", Accent: "#000000", Danger: "#000000", Warning: "#000000"},
			Header:  ThemeHeader{TitleColor: "#000000", SubtitleColor: "#646464", LineColor: "#000000", LineWidth: 0.5},
			Section: ThemeSection{LabelColor: "#000000", TextColor: "#000000", BackgroundColor: "#FFFFFF"},
			Table:   ThemeTable{HeaderBgColor: "#FFFFFF", HeaderTextColor: "#000000", RowTextColor: "#000000", RowBgColor: "#FFFFFF", RowAltBgColor: "#FFFFFF", BorderColor: "#000000", BorderWidth: 0.1},
			Summary: ThemeSummary{BgColor: "#FFFFFF", TextColor: "#000000", HighlightBgColor: "#FFFFFF", HighlightTextColor: "#000000", BorderColor: "#000000"},
			Footer:  ThemeFooter{TextColor: "#646464", LineColor: "#000000"},
		}
	}

	// ถ้าระบุ themeName โดยตรง ใช้ preset นั้น (รูปแบบที่ frontend ส่งมา)
	if payload.ThemeName != "" {
		return GetPresetTheme(payload.ThemeName)
	}

	// ถ้าไม่มี theme ใช้ default
	if payload.Theme == nil {
		return GetPresetTheme("modern-blue")
	}

	// ถ้าระบุ preset ใน Theme object ใช้ preset นั้น
	if payload.Theme.Preset != "" {
		return GetPresetTheme(payload.Theme.Preset)
	}

	// ใช้ theme ที่กำหนดเอง แต่ต้อง merge กับ default
	theme := GetPresetTheme("modern-blue")

	// Override ด้วยค่าที่ส่งมา
	if payload.Theme.Colors.Primary != "" {
		theme.Colors.Primary = payload.Theme.Colors.Primary
	}
	if payload.Theme.Colors.Secondary != "" {
		theme.Colors.Secondary = payload.Theme.Colors.Secondary
	}
	if payload.Theme.Header.TitleColor != "" {
		theme.Header.TitleColor = payload.Theme.Header.TitleColor
	}
	if payload.Theme.Header.SubtitleColor != "" {
		theme.Header.SubtitleColor = payload.Theme.Header.SubtitleColor
	}
	if payload.Theme.Header.LineColor != "" {
		theme.Header.LineColor = payload.Theme.Header.LineColor
	}
	if payload.Theme.Table.HeaderBgColor != "" {
		theme.Table.HeaderBgColor = payload.Theme.Table.HeaderBgColor
	}
	if payload.Theme.Table.HeaderTextColor != "" {
		theme.Table.HeaderTextColor = payload.Theme.Table.HeaderTextColor
	}
	if payload.Theme.Table.RowAltBgColor != "" {
		theme.Table.RowAltBgColor = payload.Theme.Table.RowAltBgColor
	}
	if payload.Theme.Table.BorderColor != "" {
		theme.Table.BorderColor = payload.Theme.Table.BorderColor
	}
	if payload.Theme.Summary.HighlightBgColor != "" {
		theme.Summary.HighlightBgColor = payload.Theme.Summary.HighlightBgColor
	}
	if payload.Theme.Summary.HighlightTextColor != "" {
		theme.Summary.HighlightTextColor = payload.Theme.Summary.HighlightTextColor
	}

	return theme
}

// GetEffectiveFontSizes - ดึงขนาด font ที่จะใช้งานจริง
func GetEffectiveFontSizes(payload GenPDFPayload) FontSizes {
	// Default font sizes
	defaults := FontSizes{
		Header:  10,
		Detail:  7,
		Summary: 9,
		Footer:  8,
	}

	// ถ้าไม่มี FontSizes ใช้ค่าจาก FontSize เดิม หรือ default
	if payload.FontSizes == nil {
		if payload.FontSize > 0 {
			// ใช้ FontSize เป็นฐานคำนวณ
			base := payload.FontSize
			return FontSizes{
				Header:  base + 1,
				Detail:  int(float64(base) * 0.75),
				Summary: base,
				Footer:  base - 1,
			}
		}
		return defaults
	}

	// ใช้ค่าที่กำหนดมา ถ้าไม่มีให้ใช้ default
	result := defaults
	if payload.FontSizes.Header > 0 {
		result.Header = payload.FontSizes.Header
	}
	if payload.FontSizes.Detail > 0 {
		result.Detail = payload.FontSizes.Detail
	}
	if payload.FontSizes.Summary > 0 {
		result.Summary = payload.FontSizes.Summary
	}
	if payload.FontSizes.Footer > 0 {
		result.Footer = payload.FontSizes.Footer
	}

	return result
}

// GetPresetTemplate - ดึง preset template ตามรหัส
func GetPresetTemplate(id string) Template {
	presets := map[string]Template{
		"standard": {
			ID: "standard", Name: "มาตรฐาน",
			HeaderLayout: "standard", TableStyle: "striped", SummaryLayout: "right", FooterStyle: "standard",
			ShowLogo: false, LogoStyle: "normal", ShowWatermark: false, WatermarkText: "",
			ShowBorder: false, BorderStyle: "none",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: false, ShowNotes: false, ShowSignature: false,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: true, ShowHeaderLine: true, HeaderLineWidth: 0.3, ShowHeaderBackground: true,
			ShowSubtotal: true, ShowTotalDiscount: true, ShowTotalTax: true, HighlightTotal: true,
		},
		"modern": {
			ID: "modern", Name: "ทันสมัย",
			HeaderLayout: "banner", TableStyle: "modern", SummaryLayout: "boxed", FooterStyle: "branded",
			ShowLogo: false, LogoStyle: "normal", ShowWatermark: false, WatermarkText: "",
			ShowBorder: false, BorderStyle: "none",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: false, ShowNotes: false, ShowSignature: false,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: true, ShowHeaderLine: false, HeaderLineWidth: 0.3, ShowHeaderBackground: true,
			ShowSubtotal: true, ShowTotalDiscount: true, ShowTotalTax: true, HighlightTotal: true,
		},
		"professional": {
			ID: "professional", Name: "มืออาชีพ",
			HeaderLayout: "split", TableStyle: "elegant", SummaryLayout: "right", FooterStyle: "detailed",
			ShowLogo: false, LogoStyle: "normal", ShowWatermark: false, WatermarkText: "",
			ShowBorder: true, BorderStyle: "solid",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: true, ShowNotes: true, ShowSignature: true,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: true, ShowHeaderLine: true, HeaderLineWidth: 0.5, ShowHeaderBackground: true,
			ShowSubtotal: true, ShowTotalDiscount: true, ShowTotalTax: true, HighlightTotal: true,
		},
		"compact": {
			ID: "compact", Name: "กะทัดรัด",
			HeaderLayout: "minimal", TableStyle: "compact", SummaryLayout: "minimal", FooterStyle: "minimal",
			ShowLogo: false, LogoStyle: "small", ShowWatermark: false, WatermarkText: "",
			ShowBorder: false, BorderStyle: "none",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: false, ShowNotes: false, ShowSignature: false,
			HeaderSpacing: 2, TableSpacing: 2, SummarySpacing: 2,
			MarginTop: 8, MarginBottom: 8, MarginLeft: 8, MarginRight: 8,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: false, ShowTax: false,
			AlternateRowColor: false, ShowHeaderLine: true, HeaderLineWidth: 0.2, ShowHeaderBackground: true,
			ShowSubtotal: false, ShowTotalDiscount: false, ShowTotalTax: false, HighlightTotal: true,
		},
		"simple": {
			ID: "simple", Name: "เรียบง่าย",
			HeaderLayout: "leftAligned", TableStyle: "clean", SummaryLayout: "right", FooterStyle: "minimal",
			ShowLogo: false, LogoStyle: "normal", ShowWatermark: false, WatermarkText: "",
			ShowBorder: false, BorderStyle: "none",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: false, ShowNotes: false, ShowSignature: false,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: false, ShowHeaderLine: true, HeaderLineWidth: 0.3, ShowHeaderBackground: false,
			ShowSubtotal: true, ShowTotalDiscount: true, ShowTotalTax: true, HighlightTotal: false,
		},
		"classic": {
			ID: "classic", Name: "คลาสสิค",
			HeaderLayout: "centered", TableStyle: "bordered", SummaryLayout: "full", FooterStyle: "standard",
			ShowLogo: false, LogoStyle: "normal", ShowWatermark: false, WatermarkText: "",
			ShowBorder: true, BorderStyle: "double",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: true, ShowNotes: true, ShowSignature: true,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: true, ShowHeaderLine: true, HeaderLineWidth: 0.4, ShowHeaderBackground: true,
			ShowSubtotal: true, ShowTotalDiscount: true, ShowTotalTax: true, HighlightTotal: true,
		},
		"premium": {
			ID: "premium", Name: "พรีเมียม",
			HeaderLayout: "banner", TableStyle: "elegant", SummaryLayout: "highlighted", FooterStyle: "withTerms",
			ShowLogo: false, LogoStyle: "large", ShowWatermark: false, WatermarkText: "",
			ShowBorder: true, BorderStyle: "rounded",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: true, ShowNotes: true, ShowSignature: true,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: true, ShowHeaderLine: true, HeaderLineWidth: 0.5, ShowHeaderBackground: true,
			ShowSubtotal: true, ShowTotalDiscount: true, ShowTotalTax: true, HighlightTotal: true,
		},
		"minimalist": {
			ID: "minimalist", Name: "มินิมอล",
			HeaderLayout: "minimal", TableStyle: "minimal", SummaryLayout: "minimal", FooterStyle: "minimal",
			ShowLogo: false, LogoStyle: "small", ShowWatermark: false, WatermarkText: "",
			ShowBorder: false, BorderStyle: "none",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: false, ShowNotes: false, ShowSignature: false,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: false, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: false, ShowHeaderLine: true, HeaderLineWidth: 0.2, ShowHeaderBackground: false,
			ShowSubtotal: false, ShowTotalDiscount: false, ShowTotalTax: false, HighlightTotal: true,
		},
		"corporate": {
			ID: "corporate", Name: "องค์กร",
			HeaderLayout: "split", TableStyle: "bordered", SummaryLayout: "twoColumn", FooterStyle: "detailed",
			ShowLogo: false, LogoStyle: "normal", ShowWatermark: false, WatermarkText: "",
			ShowBorder: true, BorderStyle: "solid",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: true, ShowNotes: true, ShowSignature: true,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: true, ShowHeaderLine: true, HeaderLineWidth: 0.4, ShowHeaderBackground: true,
			ShowSubtotal: true, ShowTotalDiscount: true, ShowTotalTax: true, HighlightTotal: true,
		},
		"creative": {
			ID: "creative", Name: "สร้างสรรค์",
			HeaderLayout: "banner", TableStyle: "modern", SummaryLayout: "highlighted", FooterStyle: "branded",
			ShowLogo: false, LogoStyle: "circular", ShowWatermark: false, WatermarkText: "",
			ShowBorder: true, BorderStyle: "dashed",
			ShowDocTitle: true, ShowCompanyInfo: true, ShowCustomerInfo: true, ShowPaymentInfo: false, ShowNotes: true, ShowSignature: false,
			HeaderSpacing: 3, TableSpacing: 3, SummarySpacing: 3,
			MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
			ShowRowNumber: true, ShowUnitPrice: true, ShowDiscount: true, ShowTax: true,
			AlternateRowColor: true, ShowHeaderLine: false, HeaderLineWidth: 0.3, ShowHeaderBackground: true,
			ShowSubtotal: true, ShowTotalDiscount: true, ShowTotalTax: true, HighlightTotal: true,
		},
	}

	if template, ok := presets[id]; ok {
		return template
	}
	return presets["standard"] // default
}

// GetEffectiveTemplate - ดึงเทมเพลตที่จะใช้งานจริง
func GetEffectiveTemplate(payload GenPDFPayload) Template {
	// ถ้าระบุ templateId โดยตรง ใช้ preset นั้น (รูปแบบที่ frontend ส่งมา)
	if payload.TemplateID != "" {
		return GetPresetTemplate(payload.TemplateID)
	}

	// ถ้าไม่มี Template ใช้ default
	if payload.Template == nil {
		return GetPresetTemplate("standard")
	}

	// ถ้าระบุ ID ใน Template object ใช้ preset นั้น
	if payload.Template.ID != "" {
		// ตรวจสอบว่าเป็น preset หรือไม่
		preset := GetPresetTemplate(payload.Template.ID)
		if preset.ID != "" {
			// เริ่มจาก preset แล้ว override ด้วยค่าที่กำหนดมา
			return mergeTemplate(preset, payload.Template)
		}
	}

	// ใช้ default template แล้ว merge กับ custom
	defaults := GetPresetTemplate("standard")
	return mergeTemplate(defaults, payload.Template)
}

// mergeTemplate - merge template values (custom overrides preset)
func mergeTemplate(base Template, custom *Template) Template {
	if custom == nil {
		return base
	}

	result := base

	// Override string fields ถ้ามีค่า
	if custom.ID != "" {
		result.ID = custom.ID
	}
	if custom.Name != "" {
		result.Name = custom.Name
	}
	if custom.HeaderLayout != "" {
		result.HeaderLayout = custom.HeaderLayout
	}
	if custom.TableStyle != "" {
		result.TableStyle = custom.TableStyle
	}
	if custom.SummaryLayout != "" {
		result.SummaryLayout = custom.SummaryLayout
	}
	if custom.FooterStyle != "" {
		result.FooterStyle = custom.FooterStyle
	}
	if custom.LogoStyle != "" {
		result.LogoStyle = custom.LogoStyle
	}
	if custom.WatermarkText != "" {
		result.WatermarkText = custom.WatermarkText
	}
	if custom.BorderStyle != "" {
		result.BorderStyle = custom.BorderStyle
	}

	// Override float64 fields ถ้ามีค่า > 0
	if custom.HeaderSpacing > 0 {
		result.HeaderSpacing = custom.HeaderSpacing
	}
	if custom.TableSpacing > 0 {
		result.TableSpacing = custom.TableSpacing
	}
	if custom.SummarySpacing > 0 {
		result.SummarySpacing = custom.SummarySpacing
	}
	if custom.MarginTop > 0 {
		result.MarginTop = custom.MarginTop
	}
	if custom.MarginBottom > 0 {
		result.MarginBottom = custom.MarginBottom
	}
	if custom.MarginLeft > 0 {
		result.MarginLeft = custom.MarginLeft
	}
	if custom.MarginRight > 0 {
		result.MarginRight = custom.MarginRight
	}
	if custom.HeaderLineWidth > 0 {
		result.HeaderLineWidth = custom.HeaderLineWidth
	}

	// Override bool fields - ใช้ค่าจาก custom โดยตรง
	result.ShowLogo = custom.ShowLogo
	result.ShowWatermark = custom.ShowWatermark
	result.ShowBorder = custom.ShowBorder
	result.ShowDocTitle = custom.ShowDocTitle
	result.ShowCompanyInfo = custom.ShowCompanyInfo
	result.ShowCustomerInfo = custom.ShowCustomerInfo
	result.ShowPaymentInfo = custom.ShowPaymentInfo
	result.ShowNotes = custom.ShowNotes
	result.ShowSignature = custom.ShowSignature
	result.ShowRowNumber = custom.ShowRowNumber
	result.ShowUnitPrice = custom.ShowUnitPrice
	result.ShowDiscount = custom.ShowDiscount
	result.ShowTax = custom.ShowTax
	result.AlternateRowColor = custom.AlternateRowColor
	result.ShowHeaderLine = custom.ShowHeaderLine
	result.ShowHeaderBackground = custom.ShowHeaderBackground
	result.ShowSubtotal = custom.ShowSubtotal
	result.ShowTotalDiscount = custom.ShowTotalDiscount
	result.ShowTotalTax = custom.ShowTotalTax
	result.HighlightTotal = custom.HighlightTotal

	return result
}

// HexToRGB - แปลง hex color เป็น RGB
func HexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// WrapThaiText - ตัดคำภาษาไทยและภาษาอื่นให้พอดีกับความกว้าง
// ใช้การวัดความกว้างจริงของข้อความแทนการนับตัวอักษร
func WrapThaiText(pdf *gofpdf.Fpdf, text string, maxWidth float64) []string {
	if text == "" {
		return []string{""}
	}

	// ถ้าข้อความสั้นพอ ไม่ต้องตัด
	if pdf.GetStringWidth(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	runes := []rune(text)
	currentLine := ""

	for i := 0; i < len(runes); i++ {
		char := string(runes[i])
		testLine := currentLine + char

		// วัดความกว้างของบรรทัดทดสอบ
		if pdf.GetStringWidth(testLine) <= maxWidth {
			currentLine = testLine
		} else {
			// บรรทัดเต็มแล้ว ให้หาจุดตัดที่เหมาะสม
			if currentLine != "" {
				// หาจุดตัดที่ดี (เว้นวรรค หรือ ตัวอักษรพิเศษ)
				breakPoint := findThaiBreakPoint(currentLine, char)
				if breakPoint > 0 && breakPoint < len([]rune(currentLine)) {
					// ตัดที่จุดเหมาะสม
					lineRunes := []rune(currentLine)
					lines = append(lines, string(lineRunes[:breakPoint]))
					currentLine = string(lineRunes[breakPoint:]) + char
				} else {
					lines = append(lines, currentLine)
					currentLine = char
				}
			} else {
				currentLine = char
			}
		}
	}

	// เพิ่มบรรทัดสุดท้าย
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// findThaiBreakPoint - หาจุดตัดคำที่เหมาะสมสำหรับภาษาไทย
func findThaiBreakPoint(text string, nextChar string) int {
	runes := []rune(text)
	length := len(runes)

	// หาจุดตัดจากท้ายมาหน้า
	for i := length - 1; i > 0; i-- {
		char := runes[i]
		prevChar := runes[i-1]

		// ตัดหลังเว้นวรรค
		if char == ' ' {
			return i + 1
		}

		// ตัดหลังตัวเลข ถ้าตัวถัดไปไม่ใช่ตัวเลข
		if isDigit(prevChar) && !isDigit(char) {
			return i
		}

		// ตัดหลังเครื่องหมาย / - . ,
		if prevChar == '/' || prevChar == '-' || prevChar == '.' || prevChar == ',' {
			return i
		}

		// สำหรับภาษาไทย - ตัดก่อนพยัญชนะต้น (ไม่ตัดสระลอย สระบน สระล่าง วรรณยุกต์)
		if isThaiConsonant(char) && !isThaiUpperVowelOrTone(char) && i > 1 {
			// ตรวจสอบว่าตัวก่อนหน้าไม่ใช่สระนำ
			if !isThaiLeadingVowel(prevChar) {
				return i
			}
		}
	}

	return 0 // ไม่พบจุดตัดที่ดี
}

// isDigit - ตรวจสอบว่าเป็นตัวเลขหรือไม่
func isDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= '๐' && r <= '๙')
}

// isThaiConsonant - ตรวจสอบว่าเป็นพยัญชนะไทยหรือไม่
func isThaiConsonant(r rune) bool {
	return r >= 'ก' && r <= 'ฮ'
}

// isThaiUpperVowelOrTone - ตรวจสอบว่าเป็นสระบน สระล่าง หรือวรรณยุกต์หรือไม่
func isThaiUpperVowelOrTone(r rune) bool {
	// สระบน: ิ ี ึ ื ็ ั
	// สระล่าง: ุ ู
	// วรรณยุกต์: ่ ้ ๊ ๋
	// ทัณฑฆาต: ์
	return (r >= 0x0E31 && r <= 0x0E3A) || (r >= 0x0E47 && r <= 0x0E4E)
}

// isThaiLeadingVowel - ตรวจสอบว่าเป็นสระนำหรือไม่ (เ แ โ ใ ไ)
func isThaiLeadingVowel(r rune) bool {
	return r == 'เ' || r == 'แ' || r == 'โ' || r == 'ใ' || r == 'ไ'
}

// Helper functions

func GetStringValue(doc map[string]interface{}, key string) string {
	if val, ok := doc[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func GetFloatValue(doc map[string]interface{}, key string) float64 {
	if val, ok := doc[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func GetIntValue(doc map[string]interface{}, key string) int {
	if val, ok := doc[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func FormatDateValue(val interface{}) string {
	return FormatDateWithFormat(val, "DD/MM/YYYY")
}

// FormatDateWithFormat - แปลงวันที่ตามรูปแบบที่กำหนด
// ค.ศ. (YY/YYYY) = English month names (Jan, January)
// พ.ศ. (BB/BBBB) = Thai month names (ม.ค., มกราคม)
func FormatDateWithFormat(val interface{}, format string) string {
	if val == nil {
		return ""
	}

	var t time.Time
	valid := false

	switch v := val.(type) {
	case primitive.DateTime:
		t = v.Time()
		valid = t.Year() > 1970
	case time.Time:
		t = v
		valid = v.Year() > 1970
	case map[string]interface{}:
		if dateVal, ok := v["$date"]; ok {
			return FormatDateWithFormat(dateVal, format)
		}
	}

	if !valid {
		return ""
	}

	// แปลงตามรูปแบบ
	buddhist2 := (t.Year() + 543) % 100 // พ.ศ. 2 หลัก
	buddhist4 := t.Year() + 543         // พ.ศ. 4 หลัก
	year2 := t.Year() % 100             // ค.ศ. 2 หลัก

	switch format {
	// ค.ศ. 4 หลัก (English month names)
	case "DD-MM-YYYY":
		return t.Format("02-01-2006")
	case "YYYY-MM-DD":
		return t.Format("2006-01-02")
	case "DD MMM YYYY":
		return fmt.Sprintf("%d %s %d", t.Day(), getEnglishShortMonth(int(t.Month())), t.Year())
	case "DD MMMM YYYY":
		return fmt.Sprintf("%d %s %d", t.Day(), getEnglishFullMonth(int(t.Month())), t.Year())
	case "MM/DD/YYYY": // US format
		return t.Format("01/02/2006")

	// ค.ศ. 2 หลัก (English month names)
	case "DD/MM/YY":
		return fmt.Sprintf("%02d/%02d/%02d", t.Day(), t.Month(), year2)
	case "DD-MM-YY":
		return fmt.Sprintf("%02d-%02d-%02d", t.Day(), t.Month(), year2)
	case "DD MMM YY":
		return fmt.Sprintf("%d %s %02d", t.Day(), getEnglishShortMonth(int(t.Month())), year2)
	case "DD MMMM YY":
		return fmt.Sprintf("%d %s %02d", t.Day(), getEnglishFullMonth(int(t.Month())), year2)
	case "MM/DD/YY": // US format
		return fmt.Sprintf("%02d/%02d/%02d", t.Month(), t.Day(), year2)

	// พ.ศ. 4 หลัก (Thai month names)
	case "DD/MM/BBBB":
		return fmt.Sprintf("%02d/%02d/%d", t.Day(), t.Month(), buddhist4)
	case "DD-MM-BBBB":
		return fmt.Sprintf("%02d-%02d-%d", t.Day(), t.Month(), buddhist4)
	case "DD MMMM BBBB":
		return fmt.Sprintf("%d %s %d", t.Day(), getThaiFullMonth(int(t.Month())), buddhist4)
	case "DD MMM BBBB":
		return fmt.Sprintf("%d %s %d", t.Day(), getThaiShortMonth(int(t.Month())), buddhist4)
	case "MM/DD/BBBB": // US format
		return fmt.Sprintf("%02d/%02d/%d", t.Month(), t.Day(), buddhist4)

	// พ.ศ. 2 หลัก (Thai month names)
	case "DD/MM/BB":
		return fmt.Sprintf("%02d/%02d/%02d", t.Day(), t.Month(), buddhist2)
	case "DD-MM-BB":
		return fmt.Sprintf("%02d-%02d-%02d", t.Day(), t.Month(), buddhist2)
	case "DD MMM BB":
		return fmt.Sprintf("%d %s %02d", t.Day(), getThaiShortMonth(int(t.Month())), buddhist2)
	case "DD MMMM BB":
		return fmt.Sprintf("%d %s %02d", t.Day(), getThaiFullMonth(int(t.Month())), buddhist2)
	case "MM/DD/BB": // US format
		return fmt.Sprintf("%02d/%02d/%02d", t.Month(), t.Day(), buddhist2)

	default: // "DD/MM/YYYY"
		return t.Format("02/01/2006")
	}
}

// getThaiShortMonth - ชื่อเดือนไทยแบบย่อ
func getThaiShortMonth(month int) string {
	months := []string{"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}
	if month >= 1 && month <= 12 {
		return months[month]
	}
	return ""
}

// getThaiFullMonth - ชื่อเดือนไทยแบบเต็ม
func getThaiFullMonth(month int) string {
	months := []string{"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน", "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}
	if month >= 1 && month <= 12 {
		return months[month]
	}
	return ""
}

// getEnglishShortMonth - ชื่อเดือนอังกฤษแบบย่อ
func getEnglishShortMonth(month int) string {
	months := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	if month >= 1 && month <= 12 {
		return months[month]
	}
	return ""
}

// getEnglishFullMonth - ชื่อเดือนอังกฤษแบบเต็ม
func getEnglishFullMonth(month int) string {
	months := []string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	if month >= 1 && month <= 12 {
		return months[month]
	}
	return ""
}

// GetCurrentDateTime - ดึงวันที่เวลาปัจจุบัน
func GetCurrentDateTime() string {
	return time.Now().Format("02/01/2006 15:04")
}

// GetCurrentDateTimeWithFormat - ดึงวันที่เวลาปัจจุบันตามรูปแบบที่กำหนด
func GetCurrentDateTimeWithFormat(format string) string {
	now := time.Now()
	dateStr := FormatDateWithFormat(now, format)
	return fmt.Sprintf("%s %s", dateStr, now.Format("15:04"))
}

func FormatNumber(val float64, decimals int) string {
	if decimals == 0 {
		return fmt.Sprintf("%.0f", val)
	}
	// Format with 2 decimals and add thousand separators
	formatted := fmt.Sprintf("%.2f", val)
	// Add thousand separators manually
	parts := strings.Split(formatted, ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = "." + parts[1]
	}

	// Add commas to integer part
	negative := false
	if len(intPart) > 0 && intPart[0] == '-' {
		negative = true
		intPart = intPart[1:]
	}

	var result strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}

	if negative {
		return "-" + result.String() + decPart
	}
	return result.String() + decPart
}

// NumberToThaiText - แปลงตัวเลขเป็นข้อความภาษาไทย
func NumberToThaiText(amount float64) string {
	if amount == 0 {
		return "ศูนย์บาทถ้วน"
	}

	// Simple implementation - can be expanded for full Thai number conversion
	intPart := int64(amount)
	decPart := int64((amount - float64(intPart)) * 100)

	result := FormatThaiNumber(intPart) + "บาท"
	if decPart > 0 {
		result += FormatThaiNumber(decPart) + "สตางค์"
	} else {
		result += "ถ้วน"
	}

	return result
}

// FormatThaiNumber - แปลงตัวเลขเป็นคำอ่านภาษาไทย
func FormatThaiNumber(n int64) string {
	if n == 0 {
		return "ศูนย์"
	}

	digits := []string{"", "หนึ่ง", "สอง", "สาม", "สี่", "ห้า", "หก", "เจ็ด", "แปด", "เก้า"}
	positions := []string{"", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน", "ล้าน"}

	if n < 10 {
		return digits[n]
	}

	var result strings.Builder
	str := fmt.Sprintf("%d", n)
	length := len(str)

	for i, c := range str {
		digit := int(c - '0')
		pos := length - i - 1

		if pos >= 6 {
			// Handle millions
			millionPart := n / 1000000
			result.WriteString(FormatThaiNumber(millionPart))
			result.WriteString("ล้าน")
			remainder := n % 1000000
			if remainder > 0 {
				result.WriteString(FormatThaiNumber(remainder))
			}
			return result.String()
		}

		if digit == 0 {
			continue
		}

		if pos == 1 && digit == 2 {
			result.WriteString("ยี่")
		} else if pos == 1 && digit == 1 {
			// Skip "หนึ่ง" for tens position
		} else if pos == 0 && digit == 1 && length > 1 {
			result.WriteString("เอ็ด")
		} else {
			result.WriteString(digits[digit])
		}

		if pos > 0 && digit != 0 {
			result.WriteString(positions[pos])
		}
	}

	return result.String()
}

// RegisterFonts - ลงทะเบียน fonts สำหรับ PDF (backward compatible)
// language - รหัสภาษา (ไม่ได้ใช้แล้ว แต่เก็บไว้เพื่อ backward compatibility)
func RegisterFonts(pdf *gofpdf.Fpdf, language string) string {
	return RegisterFontsWithPreferred(pdf, language, "")
}

// RegisterFontsWithPreferred - ลงทะเบียน fonts โดยระบุ preferred font ได้
// preferredFont - ชื่อ font family ที่ต้องการ
// รายการ fonts ที่มีในระบบ: GoNotoCurrent, NotoSansThai, NotoSansThai-Condensed, NotoSansThai-SemiCondensed
func RegisterFontsWithPreferred(pdf *gofpdf.Fpdf, language string, preferredFont string) string {
	const fontDir = "fonts"

	type fontCandidate struct {
		family  string
		regular string
		bold    string
	}

	// Helper function to register a font
	registerFont := func(cand fontCandidate) bool {
		fullPath := filepath.Join(fontDir, cand.regular)
		if _, err := os.Stat(fullPath); err == nil {
			pdf.AddUTF8Font(cand.family, "", cand.regular)
			boldPath := filepath.Join(fontDir, cand.bold)
			if _, err := os.Stat(boldPath); err == nil {
				pdf.AddUTF8Font(cand.family, "B", cand.bold)
			}
			fmt.Printf("[PDF Font] Registered font: %s (regular: %s, bold: %s)\n", cand.family, cand.regular, cand.bold)
			return true
		}
		fmt.Printf("[PDF Font] Font file not found: %s\n", fullPath)
		return false
	}

	// รายการ fonts ที่มีจริงในระบบ (ไฟล์อยู่ใน fonts directory)
	availableFonts := map[string]fontCandidate{
		// Universal Font - รองรับทุกภาษารวมไทย จีน ญี่ปุ่น เกาหลี ลาว พม่า เขมร
		"GoNotoCurrent": {"GoNotoCurrent", "GoNotoCurrent-Regular.ttf", "GoNotoCurrent-Bold.ttf"},
		// Google Thai Fonts - ยอดนิยม
		"Sarabun": {"Sarabun", "Sarabun-Regular.ttf", "Sarabun-Bold.ttf"},
		"Kanit":   {"Kanit", "Kanit-Regular.ttf", "Kanit-Bold.ttf"},
		"Prompt":  {"Prompt", "Prompt-Regular.ttf", "Prompt-Bold.ttf"},
		"Mitr":    {"Mitr", "Mitr-Regular.ttf", "Mitr-Bold.ttf"},
		// Thai Fonts - NotoSansThai variants
		"NotoSansThai":               {"NotoSansThai", "NotoSansThai-Regular.ttf", "NotoSansThai-Bold.ttf"},
		"NotoSansThai-Light":         {"NotoSansThaiLight", "NotoSansThai-Light.ttf", "NotoSansThai-Medium.ttf"},
		"NotoSansThai-Condensed":     {"NotoSansThaiCondensed", "NotoSansThai_Condensed-Regular.ttf", "NotoSansThai_Condensed-Bold.ttf"},
		"NotoSansThai-SemiCondensed": {"NotoSansThaiSemiCondensed", "NotoSansThai_SemiCondensed-Regular.ttf", "NotoSansThai_SemiCondensed-Bold.ttf"},
		// CJK Fonts
		"NotoSansCJKsc": {"NotoSansCJKsc", "NotoSansCJKsc-Regular.ttf", "NotoSansCJKsc-Bold.ttf"},
		"NotoSansCJKjp": {"NotoSansCJKjp", "NotoSansCJKjp-Regular.ttf", "NotoSansCJKjp-Bold.ttf"},
		"NotoSansCJKkr": {"NotoSansCJKkr", "NotoSansCJKkr-Regular.ttf", "NotoSansCJKkr-Bold.ttf"},
		// Other Asian Fonts
		"NotoSansKhmer":   {"NotoSansKhmer", "NotoSansKhmer-Regular.ttf", "NotoSansKhmer-Bold.ttf"},
		"NotoSansLao":     {"NotoSansLao", "NotoSansLao-Regular.ttf", "NotoSansLao-Bold.ttf"},
		"NotoSansMyanmar": {"NotoSansMyanmar", "NotoSansMyanmar-Regular.ttf", "NotoSansMyanmar-Bold.ttf"},
	}

	// ถ้าระบุ preferred font และมีใน map ให้ลองโหลดก่อน
	fmt.Printf("[PDF Font] Requested font: '%s', language: '%s'\n", preferredFont, language)
	if preferredFont != "" {
		if font, ok := availableFonts[preferredFont]; ok {
			if registerFont(font) {
				fmt.Printf("[PDF Font] Using preferred font: %s\n", font.family)
				return font.family
			}
		} else {
			fmt.Printf("[PDF Font] Font '%s' not in availableFonts map\n", preferredFont)
		}
	}

	// ใช้ Universal Font เป็น fallback - รองรับทุกภาษาในไฟล์เดียว
	fmt.Printf("[PDF Font] Falling back to universal fonts\n")
	universalFonts := []fontCandidate{
		{"GoNotoCurrent", "GoNotoCurrent-Regular.ttf", "GoNotoCurrent-Bold.ttf"},
		{"NotoSansCJKsc", "NotoSansCJKsc-Regular.ttf", "NotoSansCJKsc-Bold.ttf"},
		{"NotoSansThai", "NotoSansThai-Regular.ttf", "NotoSansThai-Bold.ttf"},
	}

	for _, font := range universalFonts {
		if registerFont(font) {
			fmt.Printf("[PDF Font] Using fallback font: %s\n", font.family)
			return font.family
		}
	}

	fmt.Printf("[PDF Font] No font found! Returning empty string\n")
	return ""
}

// SavePDF - บันทึก PDF ไปยัง temp directory
func SavePDF(pdf *gofpdf.Fpdf, guid string) (string, error) {
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("doc_%s_%s.pdf", guid, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(tempDir, filename)

	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return "", err
	}

	return filePath, nil
}
