package aichat

import "smlcloudplatform/internal/goapi/aiprovider"

// agentToolNames — whitelist ของ tools ที่ agent ใช้ได้ (readonly business tools เท่านั้น)
var agentToolNames = map[string]bool{
	"search_products":       true,
	"get_daily_sales":       true,
	"get_sales_by_date_range": true,
	"get_top_selling_products": true,
	"get_sales_by_seller":   true,
	"get_monthly_summary":   true,
	"get_dashboard_kpis":    true,
	"get_business_health":   true,
	"get_profit_analysis":   true,
	"get_accounts_receivable": true,
	"get_accounts_payable":  true,
	"get_cash_flow":         true,
	"get_inventory_value":   true,
	"get_low_stock_alerts":  true,
	"get_dead_stock":        true,
	"get_inventory_turnover": true,
	"get_top_customers":     true,
	"get_customer_growth":   true,
	"get_customer_segments": true,
	"get_yoy_comparison":    true,
	"get_mom_comparison":    true,
	"list_units":            true,
}

// IsAgentTool ตรวจว่า tool อยู่ใน whitelist หรือไม่
func IsAgentTool(name string) bool {
	return agentToolNames[name]
}

// AgentToolDefs returns OpenAI function calling format tool definitions
func AgentToolDefs() []aiprovider.OAITool {
	return []aiprovider.OAITool{
		// Product Search
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "search_products",
				Description: "ค้นหาสินค้าด้วยคำค้น พร้อมข้อมูลราคาและยอดคงเหลือ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":        "string",
							"description": "คำค้นหาสินค้า (ภาษาไทยหรืออังกฤษ)",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนสินค้าสูงสุด (default: 50, max: 200)",
						},
					},
					"required": []string{"keyword"},
				},
			},
		},
		// Sales Tools
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_daily_sales",
				Description: "ดูยอดขายรายวัน — สรุปยอดขาย จำนวนออเดอร์ และรายละเอียดของวันที่ระบุ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"date": map[string]interface{}{
							"type":        "string",
							"description": "วันที่ในรูปแบบ YYYY-MM-DD",
						},
						"branch_code": map[string]interface{}{
							"type":        "string",
							"description": "รหัสสาขา (ไม่ระบุ = ทุกสาขา)",
						},
					},
					"required": []string{"date"},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_sales_by_date_range",
				Description: "ดูยอดขายตามช่วงเวลา — สามารถจัดกลุ่มเป็นรายวัน/สัปดาห์/เดือน",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
						"group_by": map[string]interface{}{
							"type":        "string",
							"description": "จัดกลุ่ม: day, week, month (default: day)",
							"enum":        []string{"day", "week", "month"},
						},
						"branch_code": map[string]interface{}{
							"type":        "string",
							"description": "รหัสสาขา (ไม่ระบุ = ทุกสาขา)",
						},
					},
					"required": []string{"from_date", "to_date"},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_top_selling_products",
				Description: "ดูสินค้าขายดีตามช่วงเวลา",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนสินค้า (default: 10, max: 100)",
						},
						"branch_code": map[string]interface{}{
							"type":        "string",
							"description": "รหัสสาขา (ไม่ระบุ = ทุกสาขา)",
						},
					},
					"required": []string{"from_date", "to_date"},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_sales_by_seller",
				Description: "ดูยอดขายแยกตามพนักงานขาย",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
					},
					"required": []string{"from_date", "to_date"},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_monthly_summary",
				Description: "สรุปยอดขายรายเดือน เทียบกับเดือนก่อนหน้า",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"year": map[string]interface{}{
							"type":        "number",
							"description": "ปี (default: ปีปัจจุบัน)",
						},
						"month": map[string]interface{}{
							"type":        "number",
							"description": "เดือน 1-12 (default: เดือนปัจจุบัน)",
						},
					},
				},
			},
		},
		// Dashboard Tools
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_dashboard_kpis",
				Description: "ดู KPI Dashboard สำหรับผู้บริหาร — ยอดขาย ออเดอร์ กำไร ลูกค้า สินค้าคงเหลือ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"period": map[string]interface{}{
							"type":        "string",
							"description": "ช่วงเวลา: today, this_week, this_month, this_year (default: this_month)",
							"enum":        []string{"today", "this_week", "this_month", "this_year"},
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_business_health",
				Description: "ดูคะแนนสุขภาพธุรกิจ (0-100) พร้อมตัวชี้วัดและคำแนะนำ",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		// Financial Tools
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_profit_analysis",
				Description: "วิเคราะห์กำไร — รายได้ ต้นทุน อัตรากำไร",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_accounts_receivable",
				Description: "ดูลูกหนี้การค้า — ยอดค้างชำระและลูกหนี้รายใหญ่",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_accounts_payable",
				Description: "ดูเจ้าหนี้การค้า — ยอดค้างจ่ายและเจ้าหนี้รายใหญ่",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_cash_flow",
				Description: "วิเคราะห์กระแสเงินสด — รายรับ รายจ่าย ยอดสุทธิ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
					},
				},
			},
		},
		// Inventory Tools
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_inventory_value",
				Description: "ดูมูลค่าสินค้าคงเหลือ แยกตามหมวดหมู่และคลังสินค้า",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"whcode": map[string]interface{}{
							"type":        "string",
							"description": "รหัสคลังสินค้า (ไม่ระบุ = ทุกคลัง)",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_low_stock_alerts",
				Description: "ดูสินค้าที่ใกล้หมด หรือต่ำกว่าระดับขั้นต่ำ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"threshold": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนขั้นต่ำที่ถือว่าใกล้หมด (default: 10)",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนรายการ (default: 50)",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_dead_stock",
				Description: "ดูสินค้าค้างสต็อก ไม่มีความเคลื่อนไหว",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"days_no_movement": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนวันที่ไม่มีเคลื่อนไหว (default: 90)",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนรายการ (default: 50)",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_inventory_turnover",
				Description: "ดูอัตราการหมุนเวียนสินค้า และจำนวนวันที่สินค้าอยู่ในคลัง",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
					},
				},
			},
		},
		// Customer Tools
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_top_customers",
				Description: "ดูลูกค้ารายใหญ่ตามยอดซื้อ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนลูกค้า (default: 10)",
						},
						"sort_by": map[string]interface{}{
							"type":        "string",
							"description": "เรียงตาม: amount, orders, profit (default: amount)",
							"enum":        []string{"amount", "orders", "profit"},
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_customer_growth",
				Description: "ดูการเติบโตของลูกค้า — ลูกค้าใหม่ ลูกค้าที่กลับมาซื้อ",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_customer_segments",
				Description: "วิเคราะห์กลุ่มลูกค้า (RFM: ความถี่ ความใกล้ชิด มูลค่า)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_date": map[string]interface{}{
							"type":        "string",
							"description": "วันเริ่มต้น YYYY-MM-DD",
						},
						"to_date": map[string]interface{}{
							"type":        "string",
							"description": "วันสิ้นสุด YYYY-MM-DD",
						},
					},
				},
			},
		},
		// Comparison Tools
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_yoy_comparison",
				Description: "เปรียบเทียบปีต่อปี — รายได้ ออเดอร์ กำไร",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"year": map[string]interface{}{
							"type":        "number",
							"description": "ปีที่ต้องการเปรียบเทียบ (default: ปีปัจจุบัน)",
						},
						"month": map[string]interface{}{
							"type":        "number",
							"description": "เดือนที่ต้องการ 1-12 (ไม่ระบุ = ทั้งปี)",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "get_mom_comparison",
				Description: "เปรียบเทียบเดือนต่อเดือน พร้อมรายละเอียดรายสัปดาห์",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"year": map[string]interface{}{
							"type":        "number",
							"description": "ปี (default: ปีปัจจุบัน)",
						},
						"month": map[string]interface{}{
							"type":        "number",
							"description": "เดือน 1-12 (default: เดือนปัจจุบัน)",
						},
					},
				},
			},
		},
		// Unit Tools
		{
			Type: "function",
			Function: aiprovider.OAIFunction{
				Name:        "list_units",
				Description: "ดูรายการหน่วยนับสินค้าทั้งหมด (ชิ้น, กล่อง, โหล ฯลฯ)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":        "string",
							"description": "คำค้นหาหน่วยนับ (ไม่ระบุ = ทั้งหมด)",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "จำนวนรายการ (default: 100)",
						},
					},
				},
			},
		},
	}
}
