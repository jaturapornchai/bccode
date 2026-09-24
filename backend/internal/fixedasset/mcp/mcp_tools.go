package mcp

import (
	"context"
	"fmt"
	"time"

	fa "smlcloudplatform/internal/fixedasset"
)

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

var Tools = []ToolDefinition{
	{
		Name:        "fa_list_assets",
		Description: "ค้นหาและดูรายการสินทรัพย์ถาวรในระบบ (List fixed assets with filter and pagination)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"q":      map[string]interface{}{"type": "string", "description": "ค้นหาจากรหัส ชื่อ หรือ serial number"},
				"type":   map[string]interface{}{"type": "string", "description": "รหัสประเภทสินทรัพย์ เช่น EQUIPMENT, VEHICLE"},
				"status": map[string]interface{}{"type": "string", "description": "สถานะ: active, disposed, written_off"},
				"page":   map[string]interface{}{"type": "integer", "description": "หมายเลขหน้า"},
				"limit":  map[string]interface{}{"type": "integer", "description": "จำนวนต่อหน้า (default 50)"},
			},
		},
	},
	{
		Name:        "fa_create_asset",
		Description: "บันทึกสร้างสินทรัพย์ใหม่ พร้อมคำนวณตารางค่าเสื่อมราคาอัตโนมัติตามมาตรฐานบัญชีไทย (Create new fixed asset)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"required": []string{"assetcode", "name_th", "cost", "purchasedate"},
			"properties": map[string]interface{}{
				"assetcode":         map[string]interface{}{"type": "string", "description": "รหัสสินทรัพย์"},
				"name_th":           map[string]interface{}{"type": "string", "description": "ชื่อสินทรัพย์ (ภาษาไทย)"},
				"assettypecode":     map[string]interface{}{"type": "string", "description": "รหัสประเภทสินทรัพย์"},
				"cost":              map[string]interface{}{"type": "number", "description": "ราคาทุนสินทรัพย์"},
				"scrapvalue":        map[string]interface{}{"type": "number", "description": "ราคาซาก (default 0)"},
				"usefullifeyears":   map[string]interface{}{"type": "integer", "description": "อายุการใช้งาน (ปี)"},
				"deprecpercent":     map[string]interface{}{"type": "number", "description": "อัตราค่าเสื่อมต่อปี (%)"},
				"purchasedate":      map[string]interface{}{"type": "string", "description": "วันที่ซื้อ (YYYY-MM-DD)"},
				"startcalcdate":     map[string]interface{}{"type": "string", "description": "วันที่เริ่มคำนวณค่าเสื่อม (YYYY-MM-DD)"},
				"firstyearpercent":  map[string]interface{}{"type": "number", "description": "สิทธิพิเศษทางภาษีหักปีแรก (%) เช่น คอมพิวเตอร์ 40%"},
				"assetaccount":      map[string]interface{}{"type": "string", "description": "รหัสบัญชีสินทรัพย์ตามผังบัญชีของบริษัท (เว้นว่าง = ใช้ของประเภทสินทรัพย์)"},
				"accumaccount":      map[string]interface{}{"type": "string", "description": "รหัสบัญชีค่าเสื่อมราคาสะสมตามผังบัญชีของบริษัท (เว้นว่าง = ใช้ของประเภทสินทรัพย์)"},
				"expenseaccount":    map[string]interface{}{"type": "string", "description": "รหัสบัญชีค่าใช้จ่ายค่าเสื่อมราคาตามผังบัญชีของบริษัท (เว้นว่าง = ใช้ของประเภทสินทรัพย์)"},
			},
		},
	},
	{
		Name:        "fa_get_schedule",
		Description: "ดูตารางค่าเสื่อมราคารายงวดของสินทรัพย์ (Get asset depreciation schedule)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"required": []string{"assetcode"},
			"properties": map[string]interface{}{
				"assetcode": map[string]interface{}{"type": "string", "description": "รหัสสินทรัพย์"},
			},
		},
	},
	{
		Name:        "fa_post_depreciation_to_gl",
		Description: "ผ่านรายการค่าเสื่อมราคาประจำงวดเข้าสมุดรายวันทั่วไป GL (Post monthly depreciation to General Ledger)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"required": []string{"fiscalyear", "period"},
			"properties": map[string]interface{}{
				"fiscalyear": map[string]interface{}{"type": "string", "description": "ปีบัญชี เช่น 2026"},
				"period":     map[string]interface{}{"type": "integer", "description": "งวดที่ (1-12)"},
				"date":       map[string]interface{}{"type": "string", "description": "วันที่บันทึกบัญชี (YYYY-MM-DD)"},
				"docno":      map[string]interface{}{"type": "string", "description": "เลขที่ใบสำคัญ ไม่เกิน 30 ตัวอักษร (เว้นว่าง = <รหัสสมุดรายวันทั่วไป>-FA-<ปี>-<งวด>)"},
				"branchcode": map[string]interface{}{"type": "string", "description": "สาขาของใบสำคัญ (จำเป็นเมื่อเข้าระบบระดับบริษัท)"},
			},
		},
	},
	{
		Name:        "fa_dispose_asset",
		Description: "บันทึกการขายหรือตัดจำหน่ายสินทรัพย์ พร้อมคำนวณกำไร/ขาดทุน และลงบัญชี GL อัตโนมัติ (Dispose fixed asset)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"required": []string{"assetcode", "disposaldate"},
			"properties": map[string]interface{}{
				"assetcode":             map[string]interface{}{"type": "string", "description": "รหัสสินทรัพย์"},
				"disposaldate":          map[string]interface{}{"type": "string", "description": "วันที่จำหน่าย (YYYY-MM-DD)"},
				"disposaltype":          map[string]interface{}{"type": "string", "description": "ประเภท: sale, write_off, scrap"},
				"saleprice":             map[string]interface{}{"type": "number", "description": "ราคาขายสุทธิ (ก่อน VAT)"},
				"vatamount":             map[string]interface{}{"type": "number", "description": "ภาษีมูลค่าเพิ่ม (ถ้ามี)"},
				"settlementaccountcode": map[string]interface{}{"type": "string", "description": "รหัสบัญชีรับเงิน (เงินสด/เงินฝาก/ลูกหนี้) ตามผังบัญชีของบริษัท — จำเป็นเมื่อมียอดรับ"},
				"gainlossaccountcode":   map[string]interface{}{"type": "string", "description": "รหัสบัญชีกำไร/ขาดทุนจากการจำหน่าย — จำเป็นเมื่อมีกำไรหรือขาดทุน"},
				"vataccountcode":        map[string]interface{}{"type": "string", "description": "รหัสบัญชีภาษีขาย — จำเป็นเมื่อมีภาษีมูลค่าเพิ่ม"},
				"journaldocno":          map[string]interface{}{"type": "string", "description": "เลขที่ใบสำคัญ ไม่เกิน 30 ตัวอักษร (เว้นว่าง = <รหัสสมุดรายวันทั่วไป>-DISP-<รหัสสินทรัพย์>)"},
				"reason":                map[string]interface{}{"type": "string", "description": "เหตุผลในการจำหน่าย"},
			},
		},
	},
	{
		Name:        "fa_get_schedule_report",
		Description: "ดูรายงานตารางสินทรัพย์และค่าเสื่อมราคา (Fixed Asset Schedule Report)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"fiscalyear": map[string]interface{}{"type": "string", "description": "ปีบัญชี เช่น 2026"},
				"period":     map[string]interface{}{"type": "integer", "description": "งวดสูงสุด (1-12)"},
				"type":       map[string]interface{}{"type": "string", "description": "ประเภทสินทรัพย์"},
			},
		},
	},
}

type MCPHandler struct {
	store    *fa.Store
	poster   *fa.GLPoster
	reporter *fa.Reporter
}

func NewMCPHandler(store *fa.Store, poster *fa.GLPoster, reporter *fa.Reporter) *MCPHandler {
	return &MCPHandler{
		store:    store,
		poster:   poster,
		reporter: reporter,
	}
}

func (h *MCPHandler) HandleToolCall(ctx context.Context, scope fa.Scope, toolName string, args map[string]interface{}) (interface{}, error) {
	now := time.Now().UTC()
	switch toolName {
	case "fa_list_assets":
		q, _ := args["q"].(string)
		tCode, _ := args["type"].(string)
		st, _ := args["status"].(string)
		page := 1
		limit := 50
		if p, ok := args["page"].(float64); ok && p > 0 {
			page = int(p)
		}
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = int(l)
		}
		items, total, err := h.store.ListAssets(ctx, scope, q, tCode, st, page, limit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"items": items, "total": total}, nil

	case "fa_create_asset":
		code, _ := args["assetcode"].(string)
		nameTh, _ := args["name_th"].(string)
		typeCode, _ := args["assettypecode"].(string)
		costNum, _ := args["cost"].(float64)
		scrapNum, _ := args["scrapvalue"].(float64)
		usefulYears, _ := args["usefullifeyears"].(float64)
		deprecPct, _ := args["deprecpercent"].(float64)
		purchaseDate, _ := args["purchasedate"].(string)
		startCalcDate, _ := args["startcalcdate"].(string)
		firstYearPct, _ := args["firstyearpercent"].(float64)
		assetAcc, _ := args["assetaccount"].(string)
		accumAcc, _ := args["accumaccount"].(string)
		expenseAcc, _ := args["expenseaccount"].(string)

		asset := fa.Asset{
			AssetCode:                code,
			Names:                    []fa.Name{{Code: "th", Name: nameTh}},
			AssetTypeCode:            typeCode,
			Cost:                     fa.AmountFromFloat(costNum),
			ScrapValue:               fa.AmountFromFloat(scrapNum),
			UsefulLifeYears:          int(usefulYears),
			DeprecPercent:            fa.AmountFromFloat(deprecPct),
			PurchaseDate:             purchaseDate,
			StartCalcDate:            startCalcDate,
			FirstYearPercent:         fa.AmountFromFloat(firstYearPct),
			AssetAccountCode:         assetAcc,
			AccumDeprecAccountCode:   accumAcc,
			DeprecExpenseAccountCode: expenseAcc,
			Status:                   "active",
		}

		created, err := h.store.CreateAsset(ctx, scope, asset, now)
		if err != nil {
			return nil, err
		}
		return created, nil

	case "fa_get_schedule":
		code, _ := args["assetcode"].(string)
		return h.store.GetAssetDepreciationSchedule(ctx, scope, code)

	case "fa_post_depreciation_to_gl":
		year, _ := args["fiscalyear"].(string)
		periodNum, _ := args["period"].(float64)
		date, _ := args["date"].(string)
		docNo, _ := args["docno"].(string)
		branch, _ := args["branchcode"].(string)
		return h.poster.PostDepreciation(ctx, scope, year, int(periodNum), date, docNo, branch, now)

	case "fa_dispose_asset":
		code, _ := args["assetcode"].(string)
		date, _ := args["disposaldate"].(string)
		dType, _ := args["disposaltype"].(string)
		price, _ := args["saleprice"].(float64)
		vat, _ := args["vatamount"].(float64)
		settlementAcc, _ := args["settlementaccountcode"].(string)
		gainLossAcc, _ := args["gainlossaccountcode"].(string)
		vatAcc, _ := args["vataccountcode"].(string)
		journalDocNo, _ := args["journaldocno"].(string)
		reason, _ := args["reason"].(string)

		disp := fa.AssetDisposal{
			AssetCode:             code,
			DisposalDate:          date,
			DisposalType:          dType,
			SalePrice:             fa.AmountFromFloat(price),
			VatAmount:             fa.AmountFromFloat(vat),
			SettlementAccountCode: settlementAcc,
			GainLossAccountCode:   gainLossAcc,
			VatAccountCode:        vatAcc,
			JournalDocNo:          journalDocNo,
			Reason:                reason,
		}
		resDisp, resJournal, err := h.poster.DisposeAsset(ctx, scope, disp, now)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"disposal": resDisp, "journal": resJournal}, nil

	case "fa_get_schedule_report":
		year, _ := args["fiscalyear"].(string)
		periodNum, _ := args["period"].(float64)
		tCode, _ := args["type"].(string)
		return h.reporter.GetAssetScheduleReport(ctx, scope, year, int(periodNum), tCode)
	}

	return nil, fmt.Errorf("unknown tool: %s", toolName)
}
