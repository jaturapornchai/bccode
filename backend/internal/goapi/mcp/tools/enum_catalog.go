package tools

import (
	"strings"
	"time"
)

// ==================== Structs ====================

type EnumCatalogRequest struct {
	Category string `json:"category"`
	Keyword  string `json:"keyword"`
}

type EnumCatalogResponse struct {
	Enums       []EnumGroup `json:"enums"`
	TotalGroups int         `json:"totalgroups"`
	Categories  []string    `json:"categories"`
	GeneratedAt time.Time   `json:"generatedat"`
}

type EnumGroup struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	GoSource    string      `json:"gosource"`
	Values      []EnumValue `json:"values"`
}

type EnumValue struct {
	Key         interface{} `json:"key"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
}

// ==================== Main Function ====================

func GetEnumCatalog(req EnumCatalogRequest) (*EnumCatalogResponse, error) {
	allEnums := getAllEnums()

	// Filter
	var filtered []EnumGroup
	keyword := strings.ToLower(req.Keyword)
	category := strings.ToLower(req.Category)

	for _, eg := range allEnums {
		if category != "" && strings.ToLower(eg.Category) != category {
			continue
		}
		if keyword != "" {
			nameMatch := strings.Contains(strings.ToLower(eg.Name), keyword)
			descMatch := strings.Contains(strings.ToLower(eg.Description), keyword)
			valueMatch := false
			for _, v := range eg.Values {
				if strings.Contains(strings.ToLower(v.Label), keyword) || strings.Contains(strings.ToLower(v.Description), keyword) {
					valueMatch = true
					break
				}
			}
			if !nameMatch && !descMatch && !valueMatch {
				continue
			}
		}
		filtered = append(filtered, eg)
	}

	// Categories
	categorySet := make(map[string]bool)
	for _, eg := range allEnums {
		categorySet[eg.Category] = true
	}
	categories := make([]string, 0, len(categorySet))
	for c := range categorySet {
		categories = append(categories, c)
	}

	return &EnumCatalogResponse{
		Enums:       filtered,
		TotalGroups: len(filtered),
		Categories:  categories,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Enum Catalog ====================

func getAllEnums() []EnumGroup {
	return []EnumGroup{
		// ===== Transaction Flags =====
		{
			Name:        "transflag",
			Description: "Transaction type flags — ใช้ระบุประเภทเอกสารในทุก transaction",
			Category:    "transaction",
			GoSource:    "internal/goapi/handlers/kafka/constants.go",
			Values: []EnumValue{
				{Key: 6, Label: "Purchase Order", Description: "ใบสั่งซื้อ (PO)"},
				{Key: 12, Label: "Purchase", Description: "ซื้อสินค้า"},
				{Key: 16, Label: "Purchase Return", Description: "ส่งคืนสินค้า (ซื้อ)"},
				{Key: 36, Label: "Sale Order", Description: "ใบสั่งขาย"},
				{Key: 44, Label: "Sale Invoice", Description: "ขายสินค้า (ใบกำกับภาษี)"},
				{Key: 48, Label: "Sale Invoice Return", Description: "รับคืนสินค้า (ขาย)"},
				{Key: 54, Label: "Stock Balance", Description: "ยอดยกมา (opening balance)"},
				{Key: 56, Label: "Stock Pickup Product", Description: "เบิกสินค้า"},
				{Key: 58, Label: "Stock Return Product", Description: "คืนสินค้าจากการเบิก"},
				{Key: 60, Label: "Stock Receive Product", Description: "รับสินค้าเข้าคลัง"},
				{Key: 66, Label: "Stock Adjustment Increase", Description: "ปรับสต็อก (เพิ่ม)"},
				{Key: 68, Label: "Stock Adjustment Decrease", Description: "ปรับสต็อก (ลด)"},
				{Key: 72, Label: "Stock Transfer", Description: "โอนสินค้าระหว่างคลัง"},
				{Key: 310, Label: "Purchase Partial", Description: "รับสินค้าบางส่วน (partial receive)"},
			},
		},

		// ===== Payment Types =====
		{
			Name:        "payment_type",
			Description: "Payment method codes — รหัสช่องทางชำระเงิน",
			Category:    "payment",
			GoSource:    "internal/goapi/handlers/transaction_calculator.go",
			Values: []EnumValue{
				{Key: "01", Label: "Cash", Description: "เงินสด"},
				{Key: "02", Label: "Transfer", Description: "โอนเงิน"},
				{Key: "03", Label: "Credit Card", Description: "บัตรเครดิต"},
				{Key: "04", Label: "Cheque", Description: "เช็ค"},
				{Key: "05", Label: "Coupon", Description: "คูปอง"},
				{Key: "06", Label: "QR Payment", Description: "QR Payment / PromptPay"},
				{Key: "07", Label: "Deposit", Description: "มัดจำ / เงินประกัน"},
			},
		},

		// ===== VAT Types =====
		{
			Name:        "vattype",
			Description: "VAT calculation types — วิธีคำนวณภาษีมูลค่าเพิ่ม",
			Category:    "transaction",
			GoSource:    "internal/goapi/handlers/transaction_calculator.go",
			Values: []EnumValue{
				{Key: 0, Label: "Excluded", Description: "แยกนอก — ภาษีคิดแยกจากราคา"},
				{Key: 1, Label: "Included", Description: "รวมใน — ราคารวมภาษีแล้ว"},
				{Key: 2, Label: "Non-taxable", Description: "ไม่กระทบภาษี — ยกเว้น VAT"},
			},
		},

		// ===== VAT Calculation =====
		{
			Name:        "vat_cal",
			Description: "VAT calculation flag — สถานะการคิดภาษี",
			Category:    "transaction",
			GoSource:    "internal/goapi/handlers/transaction_calculator.go",
			Values: []EnumValue{
				{Key: 0, Label: "Taxable", Description: "มีภาษี (คิด VAT)"},
				{Key: 1, Label: "Tax Exempt", Description: "ยกเว้นภาษี (ไม่คิด VAT)"},
			},
		},

		// ===== Approval Status =====
		{
			Name:        "approval_status",
			Description: "PO approval status — สถานะการอนุมัติใบสั่งซื้อ",
			Category:    "approval",
			GoSource:    "internal/goapi/handlers/approval/handlers.go",
			Values: []EnumValue{
				{Key: "draft", Label: "Draft", Description: "ร่าง — ยังไม่ส่งอนุมัติ"},
				{Key: "pending", Label: "Pending", Description: "รอการอนุมัติ"},
				{Key: "approved", Label: "Approved", Description: "อนุมัติแล้ว"},
				{Key: "rejected", Label: "Rejected", Description: "ไม่อนุมัติ (ปฏิเสธ)"},
				{Key: "auto_approved", Label: "Auto Approved", Description: "อนุมัติอัตโนมัติ (ไม่เข้าเงื่อนไขต้องอนุมัติ)"},
			},
		},

		// ===== Approval Actions =====
		{
			Name:        "approval_action",
			Description: "Approval action types — ประเภทการดำเนินการอนุมัติ",
			Category:    "approval",
			GoSource:    "internal/goapi/handlers/approval/handlers.go",
			Values: []EnumValue{
				{Key: "submit", Label: "Submit", Description: "ส่งเอกสารเพื่อขออนุมัติ"},
				{Key: "approve", Label: "Approve", Description: "อนุมัติ"},
				{Key: "reject", Label: "Reject", Description: "ปฏิเสธ (ไม่อนุมัติ)"},
				{Key: "withdraw", Label: "Withdraw", Description: "ถอนเอกสารกลับ"},
			},
		},

		// ===== Approval Source =====
		{
			Name:        "approval_source",
			Description: "Notification/approval source channel — ช่องทางการแจ้งเตือนและอนุมัติ",
			Category:    "approval",
			GoSource:    "internal/goapi/handlers/approval/notification.go",
			Values: []EnumValue{
				{Key: "line", Label: "LINE OA", Description: "อนุมัติผ่าน LINE"},
				{Key: "email", Label: "Email", Description: "อนุมัติผ่าน Email"},
				{Key: "app", Label: "App", Description: "อนุมัติผ่าน App"},
			},
		},

		// ===== Data History Actions =====
		{
			Name:        "action_type",
			Description: "Data change action types — ประเภทการเปลี่ยนแปลงข้อมูล",
			Category:    "datahistory",
			GoSource:    "internal/goapi/handlers/datahistory/datahistory.go",
			Values: []EnumValue{
				{Key: "create", Label: "Create", Description: "สร้างใหม่"},
				{Key: "update", Label: "Update", Description: "แก้ไข"},
				{Key: "delete", Label: "Delete", Description: "ลบ"},
			},
		},

		// ===== Screen Types =====
		{
			Name:        "screentype",
			Description: "Screen/document types for data history — ประเภทหน้าจอ/เอกสาร",
			Category:    "datahistory",
			GoSource:    "internal/goapi/handlers/datahistory/datahistory.go",
			Values: []EnumValue{
				{Key: "purchaseorder", Label: "Purchase Order", Description: "ใบสั่งซื้อ"},
				{Key: "purchase", Label: "Purchase", Description: "ซื้อสินค้า"},
				{Key: "sale", Label: "Sale", Description: "ขายสินค้า"},
				{Key: "stockadjust", Label: "Stock Adjust", Description: "ปรับสต็อก"},
				{Key: "transfer", Label: "Transfer", Description: "โอนสินค้า"},
			},
		},

		// ===== Import Status =====
		{
			Name:        "import_status",
			Description: "Excel import processing status — สถานะการนำเข้าข้อมูล",
			Category:    "import",
			GoSource:    "internal/goapi/dataimport/xlsx_product.go",
			Values: []EnumValue{
				{Key: 0, Label: "Not Started", Description: "ยังไม่เริ่ม"},
				{Key: 1, Label: "Processing", Description: "กำลังประมวลผล"},
				{Key: 2, Label: "Completed", Description: "เสร็จแล้ว"},
				{Key: 3, Label: "Error", Description: "เกิดข้อผิดพลาด"},
			},
		},

		// ===== Personal Type =====
		{
			Name:        "personaltype",
			Description: "Personal/company type — ประเภทบุคคล/นิติบุคคล (ลูกค้า/เจ้าหนี้/ลูกหนี้)",
			Category:    "master",
			GoSource:    "internal/debtaccount/customer/models/customer.go",
			Values: []EnumValue{
				{Key: 0, Label: "Company", Description: "นิติบุคคล (บริษัท)"},
				{Key: 1, Label: "Individual", Description: "บุคคลธรรมดา"},
			},
		},

		// ===== PDF Orientation =====
		{
			Name:        "pdf_orientation",
			Description: "PDF page orientation — แนวกระดาษ PDF",
			Category:    "pdf",
			GoSource:    "internal/goapi/config/query_config.go",
			Values: []EnumValue{
				{Key: "L", Label: "Landscape", Description: "แนวนอน"},
				{Key: "P", Label: "Portrait", Description: "แนวตั้ง"},
			},
		},

		// ===== PDF Page Size =====
		{
			Name:        "pdf_pagesize",
			Description: "PDF page sizes — ขนาดกระดาษ PDF",
			Category:    "pdf",
			GoSource:    "internal/goapi/config/query_config.go",
			Values: []EnumValue{
				{Key: "A4", Label: "A4", Description: "210 x 297 mm (มาตรฐาน)"},
				{Key: "A5", Label: "A5", Description: "148 x 210 mm"},
				{Key: "Letter", Label: "Letter", Description: "215.9 x 279.4 mm (US)"},
			},
		},

		// ===== Kafka Event Actions =====
		{
			Name:        "kafka_event_action",
			Description: "Kafka event actions — ประเภท event จาก Kafka (topic pattern: when-{entity}-{action})",
			Category:    "kafka",
			GoSource:    "internal/goapi/handlers/kafka/constants.go",
			Values: []EnumValue{
				{Key: "created", Label: "Created", Description: "สร้างเอกสารใหม่"},
				{Key: "updated", Label: "Updated", Description: "แก้ไขเอกสาร"},
				{Key: "deleted", Label: "Deleted", Description: "ลบเอกสาร"},
			},
		},

		// ===== Kafka Entities =====
		{
			Name:        "kafka_entity",
			Description: "Kafka entity names — ชื่อ entity ที่มี Kafka events (topic: when-{entity}-created/updated/deleted)",
			Category:    "kafka",
			GoSource:    "internal/goapi/handlers/kafka/constants.go",
			Values: []EnumValue{
				{Key: "saleinvoice", Label: "Sale Invoice", Description: "ขายสินค้า (transflag: 44)"},
				{Key: "saleinvoicereturn", Label: "Sale Return", Description: "รับคืนสินค้า (transflag: 48)"},
				{Key: "saleorder", Label: "Sale Order", Description: "ใบสั่งขาย (transflag: 36)"},
				{Key: "purchase", Label: "Purchase", Description: "ซื้อสินค้า (transflag: 12)"},
				{Key: "purchaseorder", Label: "Purchase Order", Description: "ใบสั่งซื้อ (transflag: 6)"},
				{Key: "purchasepartial", Label: "Purchase Partial", Description: "รับสินค้าบางส่วน (transflag: 310)"},
				{Key: "purchasereturn", Label: "Purchase Return", Description: "ส่งคืนสินค้า (transflag: 16)"},
				{Key: "stocktransfer", Label: "Stock Transfer", Description: "โอนสินค้า (transflag: 72)"},
				{Key: "stockreceiveproduct", Label: "Stock Receive", Description: "รับสินค้าเข้าคลัง (transflag: 60)"},
				{Key: "stockpickupproduct", Label: "Stock Pickup", Description: "เบิกสินค้า (transflag: 56)"},
				{Key: "stockreturnproduct", Label: "Stock Return", Description: "คืนสินค้าจากเบิก (transflag: 58)"},
				{Key: "stockadjustment", Label: "Stock Adjustment", Description: "ปรับสต็อก (transflag: 66/68)"},
				{Key: "stockbalance", Label: "Stock Balance", Description: "ยอดยกมา (transflag: 54)"},
				{Key: "product-barcode", Label: "Product Barcode", Description: "สินค้า/บาร์โค้ด"},
				{Key: "warehouse", Label: "Warehouse", Description: "คลังสินค้า"},
				{Key: "customer", Label: "Customer", Description: "ลูกค้า"},
				{Key: "creditor", Label: "Creditor", Description: "เจ้าหนี้"},
				{Key: "debtor", Label: "Debtor", Description: "ลูกหนี้"},
			},
		},

		// ===== Circuit Breaker States =====
		{
			Name:        "circuit_breaker_state",
			Description: "Circuit breaker states — สถานะ circuit breaker (database connection)",
			Category:    "system",
			GoSource:    "internal/goapi/mydb/circuit_breaker.go",
			Values: []EnumValue{
				{Key: 0, Label: "CLOSED", Description: "ปกติ — ทำงานได้"},
				{Key: 1, Label: "OPEN", Description: "เปิดวงจร — หยุดส่ง request"},
				{Key: 2, Label: "HALF_OPEN", Description: "ทดลอง — ส่ง request บางส่วน"},
			},
		},

		// ===== LINE OA Token Type =====
		{
			Name:        "lineoa_token_type",
			Description: "LINE OA link token types — ประเภท token สำหรับเชื่อม LINE",
			Category:    "lineoa",
			GoSource:    "internal/goapi/handlers/lineoa/models.go",
			Values: []EnumValue{
				{Key: "employee", Label: "Employee", Description: "token สำหรับเชื่อมพนักงาน"},
				{Key: "user", Label: "User", Description: "token สำหรับเชื่อมผู้ใช้"},
			},
		},

		// ===== Stock Adjustment Type =====
		{
			Name:        "stock_adjustment_type",
			Description: "Stock adjustment types — ประเภทการปรับสต็อก",
			Category:    "stock",
			GoSource:    "internal/goapi/models/mongo-trans-model.go",
			Values: []EnumValue{
				{Key: "INCREASE", Label: "Increase", Description: "เพิ่มสต็อก (transflag: 66)"},
				{Key: "DECREASE", Label: "Decrease", Description: "ลดสต็อก (transflag: 68)"},
			},
		},

		// ===== Purchase Status =====
		{
			Name:        "purchase_compare_status",
			Description: "Purchase order compare status — สถานะเทียบรับสินค้า PO",
			Category:    "purchase",
			GoSource:    "internal/goapi/models/process-doc-purchase-model.go",
			Values: []EnumValue{
				{Key: 0, Label: "Not Compared", Description: "ยังไม่เทียบ"},
				{Key: 1, Label: "Partial", Description: "รับบางส่วน"},
				{Key: 2, Label: "Complete", Description: "รับครบแล้ว"},
				{Key: 3, Label: "Over Received", Description: "รับเกิน"},
			},
		},

		// ===== Setup Config Categories =====
		{
			Name:        "config_category",
			Description: "Bootstrap config categories — หมวดหมู่การตั้งค่าระบบ",
			Category:    "setup",
			GoSource:    "internal/goapi/handlers/setup_config_handler.go",
			Values: []EnumValue{
				{Key: "service_urls", Label: "Service URLs", Description: "URL ของ services ต่างๆ"},
				{Key: "mongodb", Label: "MongoDB", Description: "การตั้งค่า MongoDB"},
				{Key: "postgresql", Label: "PostgreSQL", Description: "การตั้งค่า PostgreSQL"},
				{Key: "clickhouse", Label: "ClickHouse", Description: "การตั้งค่า ClickHouse"},
				{Key: "redis", Label: "Redis", Description: "การตั้งค่า Redis"},
				{Key: "kafka", Label: "Kafka", Description: "การตั้งค่า Kafka"},
				{Key: "integrations", Label: "Integrations", Description: "การตั้งค่า 3rd party integrations"},
			},
		},

		// ===== Chatbot Message Role =====
		{
			Name:        "chatbot_message_role",
			Description: "Chatbot conversation message roles — บทบาทในการสนทนา",
			Category:    "chatbot",
			GoSource:    "internal/goapi/handlers/lineoa/chatbot_models.go",
			Values: []EnumValue{
				{Key: "user", Label: "User", Description: "ข้อความจากผู้ใช้"},
				{Key: "assistant", Label: "Assistant", Description: "ข้อความจาก AI"},
			},
		},

		// ===================================================================
		// ===== MainAPI Enums =====
		// ===================================================================

		// ===== User Role =====
		{
			Name:        "user_role",
			Description: "Shop user roles — บทบาทผู้ใช้ในร้านค้า (MainAPI)",
			Category:    "mainapi",
			GoSource:    "pkg/models/shopuser.go",
			Values: []EnumValue{
				{Key: 0, Label: "User", Description: "ผู้ใช้ทั่วไป (ROLE_USER)"},
				{Key: 1, Label: "Admin", Description: "ผู้ดูแลระบบ (ROLE_ADMIN)"},
				{Key: 2, Label: "Owner", Description: "เจ้าของร้าน (ROLE_OWNER)"},
				{Key: 255, Label: "System", Description: "ระบบ (ROLE_SYSTEM — ใช้ภายใน)"},
			},
		},

		// ===== Product Type =====
		{
			Name:        "product_type",
			Description: "Product item types — ประเภทสินค้า (MainAPI)",
			Category:    "mainapi",
			GoSource:    "internal/microservice/product/models.go",
			Values: []EnumValue{
				{Key: 0, Label: "Normal", Description: "สินค้าปกติ (มีสต็อก)"},
				{Key: 1, Label: "Service", Description: "บริการ (ไม่มีสต็อก)"},
				{Key: 2, Label: "Set/Combo", Description: "ชุดสินค้า/เซ็ต"},
			},
		},

		// ===== Point Type =====
		{
			Name:        "point_type",
			Description: "Point/loyalty types — ประเภทแต้มสะสม (MainAPI)",
			Category:    "mainapi",
			GoSource:    "pkg/models/point.go",
			Values: []EnumValue{
				{Key: 1, Label: "Discount", Description: "แลกส่วนลด (PointTypeDiscount)"},
				{Key: 2, Label: "Cash", Description: "แลกเงินสด (PointTypeCash)"},
			},
		},

		// ===== Year Type =====
		{
			Name:        "yeartype",
			Description: "Calendar year types — ประเภทปีปฏิทิน (MainAPI)",
			Category:    "mainapi",
			GoSource:    "pkg/models/shop.go",
			Values: []EnumValue{
				{Key: 0, Label: "Buddhist Era", Description: "ปีพุทธศักราช (พ.ศ.) — default ไทย"},
				{Key: 1, Label: "Christian Era", Description: "ปีคริสต์ศักราช (ค.ศ.)"},
			},
		},

		// ===== Document Status (general) =====
		{
			Name:        "doc_status",
			Description: "General document status — สถานะเอกสารทั่วไป",
			Category:    "mainapi",
			GoSource:    "pkg/models/common.go",
			Values: []EnumValue{
				{Key: 0, Label: "Active", Description: "เอกสารปกติ (ใช้งาน)"},
				{Key: 1, Label: "Cancelled", Description: "ยกเลิก (iscancel=1)"},
				{Key: 2, Label: "Deleted", Description: "ลบแล้ว (isdelete=1, soft delete)"},
			},
		},

		// ===== Branch Status =====
		{
			Name:        "branch_status",
			Description: "Branch active status — สถานะสาขา (MainAPI)",
			Category:    "mainapi",
			GoSource:    "internal/microservice/branch/models.go",
			Values: []EnumValue{
				{Key: true, Label: "Active", Description: "เปิดใช้งาน"},
				{Key: false, Label: "Inactive", Description: "ปิดใช้งาน (ไม่แสดงในรายการ)"},
			},
		},

		// ===== Shop Status =====
		{
			Name:        "shop_status",
			Description: "Shop status — สถานะร้านค้า (MainAPI)",
			Category:    "mainapi",
			GoSource:    "pkg/models/shop.go",
			Values: []EnumValue{
				{Key: 1, Label: "Active", Description: "เปิดใช้งาน"},
				{Key: 0, Label: "Inactive", Description: "ปิดใช้งาน (ไม่แสดงในรายการ)"},
				{Key: -1, Label: "Suspended", Description: "ถูกระงับ"},
				{Key: -2, Label: "Deleted", Description: "ลบแล้ว (soft delete)"},
			},
		},

		// ===== Currency Code =====
		{
			Name:        "currency_code",
			Description: "Common currency codes — รหัสสกุลเงินที่ใช้บ่อย",
			Category:    "mainapi",
			GoSource:    "internal/microservice/currency/models.go",
			Values: []EnumValue{
				{Key: "THB", Label: "Thai Baht", Description: "บาท (฿) — default"},
				{Key: "USD", Label: "US Dollar", Description: "ดอลลาร์สหรัฐ ($)"},
				{Key: "CNY", Label: "Chinese Yuan", Description: "หยวน (¥)"},
				{Key: "JPY", Label: "Japanese Yen", Description: "เยน (¥)"},
				{Key: "EUR", Label: "Euro", Description: "ยูโร (€)"},
				{Key: "GBP", Label: "British Pound", Description: "ปอนด์ (£)"},
				{Key: "MMK", Label: "Myanmar Kyat", Description: "จ๊าต"},
				{Key: "LAK", Label: "Lao Kip", Description: "กีบ"},
				{Key: "KHR", Label: "Cambodian Riel", Description: "เรียล"},
			},
		},

		// ===== Language Code =====
		{
			Name:        "language_code",
			Description: "Supported language codes — รหัสภาษาที่รองรับ (ใช้ใน NameX)",
			Category:    "mainapi",
			GoSource:    "pkg/models/common.go",
			Values: []EnumValue{
				{Key: "th", Label: "Thai", Description: "ภาษาไทย (default)"},
				{Key: "en", Label: "English", Description: "อังกฤษ"},
				{Key: "cn", Label: "Chinese", Description: "จีน"},
				{Key: "jp", Label: "Japanese", Description: "ญี่ปุ่น"},
				{Key: "mm", Label: "Myanmar", Description: "พม่า"},
			},
		},
	}
}
