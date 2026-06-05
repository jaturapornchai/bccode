package tools

import (
	"fmt"
	"strings"
	"time"
)

// ==================== Structs ====================

type ModelSchemaRequest struct {
	Model    string `json:"model"`
	Category string `json:"category"`
	Keyword  string `json:"keyword"`
}

type ModelSchemaResponse struct {
	Models      []ModelDef `json:"models"`
	TotalCount  int        `json:"totalcount"`
	Categories  []string   `json:"categories"`
	GeneratedAt time.Time  `json:"generatedat"`
}

type ModelDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	SourceFile  string     `json:"sourcefile"`
	Fields      []FieldDef `json:"fields"`
	DartClass   string     `json:"dartclass"`
	TSInterface string     `json:"tsinterface"`
}

type FieldDef struct {
	Name        string `json:"name"`
	JSONName    string `json:"jsonname"`
	Type        string `json:"gotype"`
	DartType    string `json:"darttype"`
	TSType      string `json:"tstype"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

// ==================== Main Function ====================

func GetModelSchema(req ModelSchemaRequest) (*ModelSchemaResponse, error) {
	allModels := getAllModels()

	// Auto-generate code for each model
	for i := range allModels {
		allModels[i].DartClass = generateDartClass(allModels[i])
		allModels[i].TSInterface = generateTSInterface(allModels[i])
	}

	// Filter
	var filtered []ModelDef
	keyword := strings.ToLower(req.Keyword)
	category := strings.ToLower(req.Category)
	model := strings.ToLower(req.Model)

	for _, m := range allModels {
		if model != "" && !strings.Contains(strings.ToLower(m.Name), model) {
			continue
		}
		if category != "" && strings.ToLower(m.Category) != category {
			continue
		}
		if keyword != "" {
			nameMatch := strings.Contains(strings.ToLower(m.Name), keyword)
			descMatch := strings.Contains(strings.ToLower(m.Description), keyword)
			if !nameMatch && !descMatch {
				continue
			}
		}
		filtered = append(filtered, m)
	}

	// Categories
	categorySet := make(map[string]bool)
	for _, m := range allModels {
		categorySet[m.Category] = true
	}
	categories := make([]string, 0, len(categorySet))
	for c := range categorySet {
		categories = append(categories, c)
	}

	return &ModelSchemaResponse{
		Models:      filtered,
		TotalCount:  len(filtered),
		Categories:  categories,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Code Generation ====================

func generateDartClass(m ModelDef) string {
	className := m.Name
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("class %s {\n", className))

	// Fields
	for _, f := range m.Fields {
		if f.Required {
			sb.WriteString(fmt.Sprintf("  final %s %s;\n", f.DartType, dartFieldName(f.JSONName)))
		} else {
			sb.WriteString(fmt.Sprintf("  final %s? %s;\n", f.DartType, dartFieldName(f.JSONName)))
		}
	}

	// Constructor
	sb.WriteString(fmt.Sprintf("\n  %s({\n", className))
	for _, f := range m.Fields {
		if f.Required {
			sb.WriteString(fmt.Sprintf("    required this.%s,\n", dartFieldName(f.JSONName)))
		} else {
			sb.WriteString(fmt.Sprintf("    this.%s,\n", dartFieldName(f.JSONName)))
		}
	}
	sb.WriteString("  });\n")

	// fromJson
	sb.WriteString(fmt.Sprintf("\n  factory %s.fromJson(Map<String, dynamic> json) {\n", className))
	sb.WriteString(fmt.Sprintf("    return %s(\n", className))
	for _, f := range m.Fields {
		fieldName := dartFieldName(f.JSONName)
		sb.WriteString(fmt.Sprintf("      %s: json['%s']", fieldName, f.JSONName))
		switch f.DartType {
		case "double":
			sb.WriteString("?.toDouble()")
		case "int":
			sb.WriteString("?.toInt()")
		case "DateTime":
			sb.WriteString(fmt.Sprintf(" != null ? DateTime.parse(json['%s']) : null", f.JSONName))
		}
		sb.WriteString(",\n")
	}
	sb.WriteString("    );\n  }\n")

	// toJson
	sb.WriteString("\n  Map<String, dynamic> toJson() {\n")
	sb.WriteString("    return {\n")
	for _, f := range m.Fields {
		fieldName := dartFieldName(f.JSONName)
		if f.DartType == "DateTime" {
			sb.WriteString(fmt.Sprintf("      '%s': %s?.toIso8601String(),\n", f.JSONName, fieldName))
		} else {
			sb.WriteString(fmt.Sprintf("      '%s': %s,\n", f.JSONName, fieldName))
		}
	}
	sb.WriteString("    };\n  }\n")

	sb.WriteString("}\n")
	return sb.String()
}

func generateTSInterface(m ModelDef) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("interface %s {\n", m.Name))
	for _, f := range m.Fields {
		optional := ""
		if !f.Required {
			optional = "?"
		}
		sb.WriteString(fmt.Sprintf("  %s%s: %s;\n", f.JSONName, optional, f.TSType))
	}
	sb.WriteString("}\n")
	return sb.String()
}

func dartFieldName(jsonName string) string {
	// snake_case → camelCase
	parts := strings.Split(jsonName, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// ==================== Model Catalog ====================

func getAllModels() []ModelDef {
	return []ModelDef{
		// ===== Transaction Header =====
		{
			Name:        "MongoDocModel",
			Description: "Transaction document header (MongoDB) — เอกสารธุรกรรม เช่น ขาย ซื้อ โอน",
			Category:    "transaction",
			SourceFile:  "internal/goapi/models/mongo-trans-model.go",
			Fields: []FieldDef{
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Holding Code"},
				{Name: "BranchId", JSONName: "branchid", Type: "string", DartType: "String", TSType: "string", Description: "Branch ID"},
				{Name: "DocNo", JSONName: "docno", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Document number"},
				{Name: "DocDateTime", JSONName: "docdatetime", Type: "time.Time", DartType: "DateTime", TSType: "Date", Required: true, Description: "Document date/time"},
				{Name: "TransFlag", JSONName: "transflag", Type: "int", DartType: "int", TSType: "number", Required: true, Description: "Transaction type (see transflag enum)"},
				{Name: "VatType", JSONName: "vattype", Type: "int", DartType: "int", TSType: "number", Description: "VAT type (0=excluded, 1=included, 2=non-taxable)"},
				{Name: "CustCode", JSONName: "custcode", Type: "string", DartType: "String", TSType: "string", Description: "Customer/vendor code"},
				{Name: "TotalAmount", JSONName: "totalamount", Type: "float64", DartType: "double", TSType: "number", Description: "Total amount"},
				{Name: "TotalQty", JSONName: "totalqty", Type: "float64", DartType: "double", TSType: "number", Description: "Total quantity"},
				{Name: "Currency", JSONName: "currency", Type: "string", DartType: "String", TSType: "string", Description: "Base currency code"},
				{Name: "DocCurrency", JSONName: "doccurrency", Type: "string", DartType: "String", TSType: "string", Description: "Document currency code"},
				{Name: "ExchangeRate", JSONName: "exchangerate", Type: "float64", DartType: "double", TSType: "number", Description: "Exchange rate"},
				{Name: "TotalAmountDoc", JSONName: "totalamountdoc", Type: "float64", DartType: "double", TSType: "number", Description: "Total in document currency"},
				{Name: "IsCancel", JSONName: "iscancel", Type: "int", DartType: "int", TSType: "number", Description: "0=active, 1=cancelled"},
				{Name: "IsDelete", JSONName: "isdelete", Type: "int", DartType: "int", TSType: "number", Description: "0=active, 1=deleted (soft delete)"},
				{Name: "CreatorCode", JSONName: "creatorcode", Type: "string", DartType: "String", TSType: "string", Description: "Creator employee code"},
				{Name: "CreatorName", JSONName: "creatorname", Type: "string", DartType: "String", TSType: "string", Description: "Creator name"},
				{Name: "CreatedAt", JSONName: "createdat", Type: "time.Time", DartType: "DateTime", TSType: "Date", Description: "Created timestamp"},
			},
		},

		// ===== Transaction Detail =====
		{
			Name:        "MongoDocDetailModel",
			Description: "Transaction line item detail — รายการสินค้าในเอกสาร",
			Category:    "transaction",
			SourceFile:  "internal/goapi/models/mongo-trans-model.go",
			Fields: []FieldDef{
				{Name: "LineNumber", JSONName: "linenumber", Type: "int", DartType: "int", TSType: "number", Required: true, Description: "Line number"},
				{Name: "DocNo", JSONName: "docno", Type: "string", DartType: "String", TSType: "string", Description: "Document number"},
				{Name: "DocDateTime", JSONName: "docdatetime", Type: "time.Time", DartType: "DateTime", TSType: "Date", Description: "Document date"},
				{Name: "Barcode", JSONName: "barcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Product barcode"},
				{Name: "ItemCode", JSONName: "itemcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Product item code"},
				{Name: "UnitCode", JSONName: "unitcode", Type: "string", DartType: "String", TSType: "string", Description: "Unit code"},
				{Name: "WhCode", JSONName: "whcode", Type: "string", DartType: "String", TSType: "string", Description: "Warehouse code"},
				{Name: "LocationCode", JSONName: "locationcode", Type: "string", DartType: "String", TSType: "string", Description: "Location code"},
				{Name: "Qty", JSONName: "qty", Type: "float64", DartType: "double", TSType: "number", Required: true, Description: "Quantity"},
				{Name: "Price", JSONName: "price", Type: "float64", DartType: "double", TSType: "number", Required: true, Description: "Unit price"},
				{Name: "Discount", JSONName: "discount", Type: "string", DartType: "String", TSType: "string", Description: "Discount text (e.g., '10%')"},
				{Name: "TotalValue", JSONName: "totalvalue", Type: "float64", DartType: "double", TSType: "number", Description: "Line total amount"},
				{Name: "SumAmount", JSONName: "sumamount", Type: "float64", DartType: "double", TSType: "number", Description: "Sum amount after discount"},
			},
		},

		// ===== Document (PostgreSQL) =====
		{
			Name:        "DocStruct",
			Description: "Document header (PostgreSQL) — เอกสารในฐานข้อมูล PostgreSQL",
			Category:    "transaction",
			SourceFile:  "internal/goapi/models/process-doc-model.go",
			Fields: []FieldDef{
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "CustCode", JSONName: "custcode", Type: "string", DartType: "String", TSType: "string"},
				{Name: "TransFlag", JSONName: "transflag", Type: "int", DartType: "int", TSType: "number", Required: true},
				{Name: "DocNo", JSONName: "docno", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "DocDateTime", JSONName: "docdatetime", Type: "time.Time", DartType: "DateTime", TSType: "Date", Required: true},
				{Name: "TotalAmount", JSONName: "totalamount", Type: "float64", DartType: "double", TSType: "number"},
				{Name: "Currency", JSONName: "currency", Type: "string", DartType: "String", TSType: "string"},
				{Name: "DocCurrency", JSONName: "doccurrency", Type: "string", DartType: "String", TSType: "string"},
				{Name: "ExchangeRate", JSONName: "exchangerate", Type: "float64", DartType: "double", TSType: "number"},
				{Name: "TotalAmountDoc", JSONName: "totalamountdoc", Type: "float64", DartType: "double", TSType: "number"},
				{Name: "ApprovalStatus", JSONName: "approvalstatus", Type: "string", DartType: "String", TSType: "string", Description: "draft/pending/approved/rejected"},
				{Name: "IsDelete", JSONName: "isdelete", Type: "int", DartType: "int", TSType: "number", Description: "Soft delete flag"},
				{Name: "CreatorCode", JSONName: "creatorcode", Type: "string", DartType: "String", TSType: "string"},
				{Name: "CreatedAt", JSONName: "createdat", Type: "time.Time", DartType: "DateTime", TSType: "Date"},
			},
		},

		// ===== Payment =====
		{
			Name:        "DocPaymentStruct",
			Description: "Payment record — ข้อมูลการชำระเงิน",
			Category:    "transaction",
			SourceFile:  "internal/goapi/models/process-doc-model.go",
			Fields: []FieldDef{
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "DocNo", JSONName: "docno", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "TransFlag", JSONName: "transflag", Type: "int", DartType: "int", TSType: "number"},
				{Name: "DocDateTime", JSONName: "docdatetime", Type: "time.Time", DartType: "DateTime", TSType: "Date"},
				{Name: "ProviderName", JSONName: "providername", Type: "string", DartType: "String", TSType: "string", Description: "Payment provider/method name"},
				{Name: "Amount", JSONName: "amount", Type: "float64", DartType: "double", TSType: "number", Required: true},
				{Name: "Description", JSONName: "description", Type: "string", DartType: "String", TSType: "string"},
			},
		},

		// ===== Product Barcode =====
		{
			Name:        "BarcodeModel",
			Description: "Product barcode — ข้อมูลสินค้า/บาร์โค้ด",
			Category:    "product",
			SourceFile:  "internal/goapi/models/barcode-model.go",
			Fields: []FieldDef{
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "ItemCode", JSONName: "itemcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Product code"},
				{Name: "Barcode", JSONName: "barcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Barcode number"},
				{Name: "Name0", JSONName: "name0", Type: "string", DartType: "String", TSType: "string", Description: "Product name (primary language)"},
				{Name: "Name1", JSONName: "name1", Type: "string", DartType: "String", TSType: "string", Description: "Product name (2nd language)"},
				{Name: "UnitCode", JSONName: "unitcode", Type: "string", DartType: "String", TSType: "string", Description: "Unit code"},
				{Name: "Price1", JSONName: "price1", Type: "float64", DartType: "double", TSType: "number", Description: "Price level 1"},
				{Name: "PriceRetail", JSONName: "priceretail", Type: "float64", DartType: "double", TSType: "number", Description: "Retail price"},
				{Name: "UnitStand", JSONName: "unitstand", Type: "float64", DartType: "double", TSType: "number", Description: "Unit stand (multiplier)"},
				{Name: "UnitDivide", JSONName: "unitdivide", Type: "float64", DartType: "double", TSType: "number", Description: "Unit divide (divisor)"},
				{Name: "IsStock", JSONName: "isstock", Type: "int", DartType: "int", TSType: "number", Description: "1=track stock, 0=no stock"},
				{Name: "ItemType", JSONName: "itemtype", Type: "int", DartType: "int", TSType: "number", Description: "0=Stock, 1=Service, 2=Set, 3=Not Stock"},
				{Name: "MaterialType", JSONName: "materialtype", Type: "int", DartType: "int", TSType: "number", Description: "0=General, 1=Material, 2=Semi-Finished, 3=Set, 4=Agricultural"},
				{Name: "ImageUri", JSONName: "imageuri", Type: "string", DartType: "String", TSType: "string", Description: "Product image URL"},
			},
		},

		// ===== Stock Balance =====
		{
			Name:        "ProductBalanceStruct",
			Description: "Product stock balance — ยอดคงเหลือสินค้า",
			Category:    "stock",
			SourceFile:  "internal/goapi/models/process-model.go",
			Fields: []FieldDef{
				{Name: "ItemCode", JSONName: "itemcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "ItemName", JSONName: "itemname", Type: "string", DartType: "String", TSType: "string"},
				{Name: "WhCode", JSONName: "whcode", Type: "string", DartType: "String", TSType: "string", Description: "Warehouse code"},
				{Name: "LocationCode", JSONName: "locationcode", Type: "string", DartType: "String", TSType: "string", Description: "Location code"},
				{Name: "UnitCode", JSONName: "unitcode", Type: "string", DartType: "String", TSType: "string"},
				{Name: "BalanceQty", JSONName: "balanceqty", Type: "float64", DartType: "double", TSType: "number", Description: "Balance quantity"},
				{Name: "BalanceAmount", JSONName: "balanceamount", Type: "float64", DartType: "double", TSType: "number", Description: "Balance amount (value)"},
				{Name: "BalanceWord", JSONName: "balanceword", Type: "string", DartType: "String", TSType: "string", Description: "Formatted balance (e.g., '1 กล่อง x 2 โหล')"},
				{Name: "IsAutoPacking", JSONName: "isautopacking", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Auto packing enabled"},
			},
		},

		// ===== Customer =====
		{
			Name:        "Customer",
			Description: "Customer master data — ข้อมูลลูกค้า",
			Category:    "master",
			SourceFile:  "internal/debtaccount/customer/models/customer.go",
			Fields: []FieldDef{
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Customer code"},
				{Name: "PersonalType", JSONName: "personaltype", Type: "int", DartType: "int", TSType: "number", Description: "0=company, 1=individual"},
				{Name: "TaxId", JSONName: "taxid", Type: "string", DartType: "String", TSType: "string", Description: "Tax ID (13 digits)"},
				{Name: "Email", JSONName: "email", Type: "string", DartType: "String", TSType: "string"},
				{Name: "CustomerType", JSONName: "customertype", Type: "int", DartType: "int", TSType: "number"},
				{Name: "BranchNumber", JSONName: "branchnumber", Type: "string", DartType: "String", TSType: "string", Description: "Branch number (สำหรับออกใบกำกับภาษี)"},
				{Name: "IsCreditor", JSONName: "iscreditor", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Also a creditor"},
				{Name: "IsDebtor", JSONName: "isdebtor", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Also a debtor"},
				{Name: "CreditDay", JSONName: "creditday", Type: "int", DartType: "int", TSType: "number", Description: "Credit days"},
				{Name: "IsMember", JSONName: "ismember", Type: "bool", DartType: "bool", TSType: "boolean"},
			},
		},

		// ===== Creditor =====
		{
			Name:        "Creditor",
			Description: "Creditor master data — ข้อมูลเจ้าหนี้",
			Category:    "master",
			SourceFile:  "internal/debtaccount/creditor/models/creditor.go",
			Fields: []FieldDef{
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Creditor code"},
				{Name: "PersonalType", JSONName: "personaltype", Type: "int", DartType: "int", TSType: "number", Description: "0=company, 1=individual"},
				{Name: "TaxId", JSONName: "taxid", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Email", JSONName: "email", Type: "string", DartType: "String", TSType: "string"},
				{Name: "BranchNumber", JSONName: "branchnumber", Type: "string", DartType: "String", TSType: "string"},
				{Name: "FundCode", JSONName: "fundcode", Type: "string", DartType: "String", TSType: "string"},
				{Name: "CreditDay", JSONName: "creditday", Type: "int", DartType: "int", TSType: "number"},
				{Name: "IsMember", JSONName: "ismember", Type: "bool", DartType: "bool", TSType: "boolean"},
			},
		},

		// ===== Debtor =====
		{
			Name:        "Debtor",
			Description: "Debtor master data — ข้อมูลลูกหนี้",
			Category:    "master",
			SourceFile:  "internal/debtaccount/debtor/models/debtor.go",
			Fields: []FieldDef{
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Debtor code"},
				{Name: "PersonalType", JSONName: "personaltype", Type: "int", DartType: "int", TSType: "number", Description: "0=company, 1=individual"},
				{Name: "TaxId", JSONName: "taxid", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Email", JSONName: "email", Type: "string", DartType: "String", TSType: "string"},
				{Name: "PointBalance", JSONName: "pointbalance", Type: "float64", DartType: "double", TSType: "number", Description: "Point balance"},
				{Name: "CreditDay", JSONName: "creditday", Type: "int", DartType: "int", TSType: "number"},
				{Name: "PriceLevel", JSONName: "pricelevel", Type: "int", DartType: "int", TSType: "number", Description: "Price level for this debtor"},
			},
		},

		// ===== Image Metadata =====
		{
			Name:        "ImageMetadata",
			Description: "Image metadata — ข้อมูลรูปภาพที่อัพโหลด",
			Category:    "file",
			SourceFile:  "internal/goapi/models/image-model.go",
			Fields: []FieldDef{
				{Name: "ID", JSONName: "id", Type: "ObjectID", DartType: "String", TSType: "string", Required: true},
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "FileName", JSONName: "filename", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "OriginalName", JSONName: "originalname", Type: "string", DartType: "String", TSType: "string"},
				{Name: "ContentType", JSONName: "contenttype", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Size", JSONName: "size", Type: "int64", DartType: "int", TSType: "number"},
				{Name: "Category", JSONName: "category", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Description", JSONName: "description", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Tags", JSONName: "tags", Type: "[]string", DartType: "List<String>", TSType: "string[]"},
				{Name: "UploadedBy", JSONName: "uploadedby", Type: "string", DartType: "String", TSType: "string"},
				{Name: "CreatedAt", JSONName: "createdat", Type: "time.Time", DartType: "DateTime", TSType: "Date"},
			},
		},

		// ===== Attachment Metadata =====
		{
			Name:        "AttachmentMetadata",
			Description: "File attachment metadata — ข้อมูลไฟล์แนบ",
			Category:    "file",
			SourceFile:  "internal/goapi/models/attachment-model.go",
			Fields: []FieldDef{
				{Name: "ID", JSONName: "id", Type: "ObjectID", DartType: "String", TSType: "string", Required: true},
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "ScreenType", JSONName: "screentype", Type: "string", DartType: "String", TSType: "string", Description: "Screen/document type"},
				{Name: "DocNo", JSONName: "docno", Type: "string", DartType: "String", TSType: "string"},
				{Name: "FileName", JSONName: "filename", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "OriginalName", JSONName: "originalname", Type: "string", DartType: "String", TSType: "string"},
				{Name: "ContentType", JSONName: "contenttype", Type: "string", DartType: "String", TSType: "string"},
				{Name: "FileType", JSONName: "filetype", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Size", JSONName: "size", Type: "int64", DartType: "int", TSType: "number"},
				{Name: "Description", JSONName: "description", Type: "string", DartType: "String", TSType: "string"},
				{Name: "CreatedAt", JSONName: "createdat", Type: "time.Time", DartType: "DateTime", TSType: "Date"},
			},
		},

		// ===== PDF History =====
		{
			Name:        "PdfHistory",
			Description: "PDF generation history — ประวัติการสร้าง PDF",
			Category:    "pdf",
			SourceFile:  "internal/goapi/models/pdf-history-model.go",
			Fields: []FieldDef{
				{Name: "ID", JSONName: "id", Type: "ObjectID", DartType: "String", TSType: "string", Required: true},
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "DocNo", JSONName: "docno", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "Title", JSONName: "title", Type: "string", DartType: "String", TSType: "string"},
				{Name: "FileName", JSONName: "filename", Type: "string", DartType: "String", TSType: "string"},
				{Name: "FileSize", JSONName: "filesize", Type: "int64", DartType: "int", TSType: "number"},
				{Name: "PageSize", JSONName: "pagesize", Type: "string", DartType: "String", TSType: "string", Description: "A4, A5, Letter"},
				{Name: "Orientation", JSONName: "orientation", Type: "string", DartType: "String", TSType: "string", Description: "L=landscape, P=portrait"},
				{Name: "PrintedBy", JSONName: "printedby", Type: "string", DartType: "String", TSType: "string"},
				{Name: "PrintedAt", JSONName: "printedat", Type: "time.Time", DartType: "DateTime", TSType: "Date"},
				{Name: "ReprintCount", JSONName: "reprintcount", Type: "int", DartType: "int", TSType: "number"},
				{Name: "TotalAmount", JSONName: "totalamount", Type: "float64", DartType: "double", TSType: "number"},
			},
		},

		// ===== Warehouse =====
		{
			Name:        "Warehouse",
			Description: "Warehouse master data — ข้อมูลคลังสินค้า",
			Category:    "master",
			SourceFile:  "internal/goapi/models/mongo-warehouse-model.go",
			Fields: []FieldDef{
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Warehouse code"},
				{Name: "Names", JSONName: "names", Type: "[]LanguageModel", DartType: "List<Map<String, dynamic>>", TSType: "Array<{code: string, name: string}>", Description: "Warehouse name (multi-language)"},
			},
		},

		// ===== Employee =====
		{
			Name:        "Employee",
			Description: "Employee data — ข้อมูลพนักงาน",
			Category:    "master",
			SourceFile:  "internal/goapi/models/mongo-employee-model.go",
			Fields: []FieldDef{
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Employee code"},
				{Name: "Name", JSONName: "name", Type: "string", DartType: "String", TSType: "string", Description: "Employee name"},
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Position", JSONName: "position", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Department", JSONName: "department", Type: "string", DartType: "String", TSType: "string"},
				{Name: "ApprovalRole", JSONName: "approvalrole", Type: "string", DartType: "String", TSType: "string", Description: "Role for PO approval"},
				{Name: "MaxApprovalAmount", JSONName: "maxapprovalamount", Type: "float64", DartType: "double", TSType: "number", Description: "Max amount this employee can approve"},
			},
		},

		// ===== Query Request =====
		{
			Name:        "ResultFromQueryRequest",
			Description: "Query execution request — request สำหรับ query ข้อมูล",
			Category:    "query",
			SourceFile:  "internal/goapi/models/query-result-model.go",
			Fields: []FieldDef{
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "Query", JSONName: "query", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "SQL query"},
				{Name: "GUID", JSONName: "guid", Type: "string", DartType: "String", TSType: "string", Description: "Optional GUID for result caching"},
			},
		},

		// ===== Purchase Status =====
		{
			Name:        "PurchaseStatusStruct",
			Description: "Purchase order status — สถานะใบสั่งซื้อ (เทียบรับ)",
			Category:    "purchase",
			SourceFile:  "internal/goapi/models/process-doc-purchase-model.go",
			Fields: []FieldDef{
				{Name: "DocNo", JSONName: "docno", Type: "string", DartType: "String", TSType: "string", Required: true},
				{Name: "IsClosed", JSONName: "isclosed", Type: "bool", DartType: "bool", TSType: "boolean", Description: "PO manually closed"},
				{Name: "IsComparedSuccess", JSONName: "iscomparedsuccess", Type: "int", DartType: "int", TSType: "number", Description: "0=not compared, 1=partial, 2=complete, 3=over received"},
			},
		},

		// ===================================================================
		// ===== MainAPI Models (จาก mainapi microservices) =====
		// ===================================================================

		// ===== Shop =====
		{
			Name:        "Shop",
			Description: "Shop master data — ข้อมูลร้านค้า (MainAPI)",
			Category:    "mainapi-master",
			SourceFile:  "pkg/models/shop.go",
			Fields: []FieldDef{
				{Name: "GuidFixed", JSONName: "guidfixed", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Shop GUID (unique identifier)"},
				{Name: "Name1", JSONName: "name1", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Shop name (primary)"},
				{Name: "Name2", JSONName: "name2", Type: "string", DartType: "String", TSType: "string", Description: "Shop name (secondary)"},
				{Name: "TaxId", JSONName: "taxid", Type: "string", DartType: "String", TSType: "string", Description: "Tax ID (13 digits)"},
				{Name: "Phone", JSONName: "phone", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Email", JSONName: "email", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Address", JSONName: "address", Type: "string", DartType: "String", TSType: "string"},
				{Name: "BranchNumber", JSONName: "branchnumber", Type: "string", DartType: "String", TSType: "string", Description: "Branch number (head office = 00000)"},
				{Name: "IsActive", JSONName: "isactive", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Shop active status"},
				{Name: "Settings", JSONName: "settings", Type: "ShopSettings", DartType: "Map<String, dynamic>", TSType: "Record<string, any>", Description: "Shop settings (VAT rate, currency, etc.)"},
				{Name: "CreatedAt", JSONName: "createdat", Type: "time.Time", DartType: "DateTime", TSType: "Date"},
				{Name: "UpdatedAt", JSONName: "updatedat", Type: "time.Time", DartType: "DateTime", TSType: "Date"},
			},
		},

		// ===== Shop User =====
		{
			Name:        "ShopUser",
			Description: "Shop user — ผู้ใช้ในร้านค้า พร้อม role (MainAPI)",
			Category:    "mainapi-master",
			SourceFile:  "pkg/models/shopuser.go",
			Fields: []FieldDef{
				{Name: "Username", JSONName: "username", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Username"},
				{Name: "Name", JSONName: "name", Type: "string", DartType: "String", TSType: "string", Description: "Display name"},
				{Name: "Email", JSONName: "email", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Role", JSONName: "role", Type: "int", DartType: "int", TSType: "number", Required: true, Description: "0=user, 1=admin, 2=owner, 255=system (see user_role enum)"},
				{Name: "HoldingCode", JSONName: "holdingcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Shop GUID"},
				{Name: "IsActive", JSONName: "isactive", Type: "bool", DartType: "bool", TSType: "boolean"},
				{Name: "ApprovalLimit", JSONName: "approvallimit", Type: "float64", DartType: "double", TSType: "number", Description: "Max approval amount"},
				{Name: "ApprovalRole", JSONName: "approvalrole", Type: "string", DartType: "String", TSType: "string", Description: "Approval role name"},
				{Name: "CreatedAt", JSONName: "createdat", Type: "time.Time", DartType: "DateTime", TSType: "Date"},
			},
		},

		// ===== Product (MainAPI) =====
		{
			Name:        "Product",
			Description: "Product master data — ข้อมูลสินค้าหลัก (MainAPI)",
			Category:    "mainapi-product",
			SourceFile:  "internal/microservice/product/models.go",
			Fields: []FieldDef{
				{Name: "GuidFixed", JSONName: "guidfixed", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Product GUID"},
				{Name: "ItemCode", JSONName: "itemcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Item code"},
				{Name: "Names", JSONName: "names", Type: "[]NameX", DartType: "List<Map<String, dynamic>>", TSType: "Array<{code: string, name: string}>", Description: "Product names (multi-language)"},
				{Name: "ItemType", JSONName: "itemtype", Type: "int", DartType: "int", TSType: "number", Description: "0=Stock, 1=Service, 2=Set, 3=Not Stock"},
				{Name: "MaterialType", JSONName: "materialtype", Type: "int", DartType: "int", TSType: "number", Description: "0=General, 1=Material, 2=Semi-Finished, 3=Set, 4=Agricultural"},
				{Name: "GroupCode", JSONName: "groupcode", Type: "string", DartType: "String", TSType: "string", Description: "Product group code"},
				{Name: "CategoryGuid", JSONName: "categoryguid", Type: "string", DartType: "String", TSType: "string", Description: "Category GUID"},
				{Name: "TaxType", JSONName: "taxtype", Type: "int", DartType: "int", TSType: "number", Description: "0=taxable, 1=tax-exempt (see vat_cal enum)"},
				{Name: "IsStock", JSONName: "isstock", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Track stock inventory"},
				{Name: "IsActive", JSONName: "isactive", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Product active status"},
				{Name: "ImageUri", JSONName: "imageuri", Type: "string", DartType: "String", TSType: "string", Description: "Product image URL"},
				{Name: "Description", JSONName: "description", Type: "string", DartType: "String", TSType: "string"},
			},
		},

		// ===== ProductBarcode (MainAPI) =====
		{
			Name:        "ProductBarcode",
			Description: "Product barcode/unit — บาร์โค้ดและหน่วยสินค้า (MainAPI)",
			Category:    "mainapi-product",
			SourceFile:  "internal/microservice/product/models.go",
			Fields: []FieldDef{
				{Name: "Barcode", JSONName: "barcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Barcode number"},
				{Name: "UnitCode", JSONName: "unitcode", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Unit code"},
				{Name: "UnitNames", JSONName: "unitnames", Type: "[]NameX", DartType: "List<Map<String, dynamic>>", TSType: "Array<{code: string, name: string}>", Description: "Unit names (multi-language)"},
				{Name: "StandValue", JSONName: "standvalue", Type: "float64", DartType: "double", TSType: "number", Description: "Unit multiplier (e.g., 1 box = 12 pcs → stand=12)"},
				{Name: "DivideValue", JSONName: "dividevalue", Type: "float64", DartType: "double", TSType: "number", Description: "Unit divisor"},
				{Name: "Price", JSONName: "price", Type: "float64", DartType: "double", TSType: "number", Description: "Selling price"},
				{Name: "AverageCost", JSONName: "averagecost", Type: "float64", DartType: "double", TSType: "number", Description: "Average cost"},
				{Name: "ItemType", JSONName: "itemtype", Type: "int", DartType: "int", TSType: "number", Description: "0=Stock, 1=Service, 2=Set, 3=Not Stock"},
				{Name: "MaterialType", JSONName: "materialtype", Type: "int", DartType: "int", TSType: "number", Description: "0=General, 1=Material, 2=Semi-Finished, 3=Set, 4=Agricultural"},
				{Name: "IsDefault", JSONName: "isdefault", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Default barcode/unit for product"},
			},
		},

		// ===== Branch =====
		{
			Name:        "Branch",
			Description: "Branch master data — ข้อมูลสาขา (MainAPI)",
			Category:    "mainapi-master",
			SourceFile:  "internal/microservice/branch/models.go",
			Fields: []FieldDef{
				{Name: "GuidFixed", JSONName: "guidfixed", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Branch GUID"},
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Branch code"},
				{Name: "Names", JSONName: "names", Type: "[]NameX", DartType: "List<Map<String, dynamic>>", TSType: "Array<{code: string, name: string}>", Description: "Branch names (multi-language)"},
				{Name: "BranchNumber", JSONName: "branchnumber", Type: "string", DartType: "String", TSType: "string", Description: "Revenue department branch number"},
				{Name: "TaxId", JSONName: "taxid", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Phone", JSONName: "phone", Type: "string", DartType: "String", TSType: "string"},
				{Name: "Address", JSONName: "address", Type: "string", DartType: "String", TSType: "string"},
				{Name: "IsActive", JSONName: "isactive", Type: "bool", DartType: "bool", TSType: "boolean"},
				{Name: "POSCount", JSONName: "poscount", Type: "int", DartType: "int", TSType: "number", Description: "Number of POS machines"},
				{Name: "DefaultWarehouse", JSONName: "defaultwarehouse", Type: "string", DartType: "String", TSType: "string", Description: "Default warehouse code"},
			},
		},

		// ===== Warehouse (MainAPI) =====
		{
			Name:        "WarehouseMainAPI",
			Description: "Warehouse master data — ข้อมูลคลังสินค้า (MainAPI, มี location/shelf)",
			Category:    "mainapi-master",
			SourceFile:  "internal/microservice/warehouse/models.go",
			Fields: []FieldDef{
				{Name: "GuidFixed", JSONName: "guidfixed", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Warehouse GUID"},
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Warehouse code"},
				{Name: "Names", JSONName: "names", Type: "[]NameX", DartType: "List<Map<String, dynamic>>", TSType: "Array<{code: string, name: string}>", Description: "Warehouse names (multi-language)"},
				{Name: "BranchCode", JSONName: "branchcode", Type: "string", DartType: "String", TSType: "string", Description: "Linked branch code"},
				{Name: "IsActive", JSONName: "isactive", Type: "bool", DartType: "bool", TSType: "boolean"},
				{Name: "Locations", JSONName: "locations", Type: "[]Location", DartType: "List<Map<String, dynamic>>", TSType: "Array<{code: string, names: NameX[]}>", Description: "Storage locations within warehouse"},
				{Name: "Shelves", JSONName: "shelves", Type: "[]Shelf", DartType: "List<Map<String, dynamic>>", TSType: "Array<{code: string, names: NameX[], locationcode: string}>", Description: "Shelves within locations"},
			},
		},

		// ===== Currency =====
		{
			Name:        "Currency",
			Description: "Currency master data — ข้อมูลสกุลเงิน (MainAPI)",
			Category:    "mainapi-master",
			SourceFile:  "internal/microservice/currency/models.go",
			Fields: []FieldDef{
				{Name: "GuidFixed", JSONName: "guidfixed", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Currency GUID"},
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Currency code (e.g., THB, USD)"},
				{Name: "Names", JSONName: "names", Type: "[]NameX", DartType: "List<Map<String, dynamic>>", TSType: "Array<{code: string, name: string}>", Description: "Currency names (multi-language)"},
				{Name: "Symbol", JSONName: "symbol", Type: "string", DartType: "String", TSType: "string", Description: "Currency symbol (e.g., ฿, $)"},
				{Name: "ExchangeRate", JSONName: "exchangerate", Type: "float64", DartType: "double", TSType: "number", Description: "Exchange rate to base currency"},
				{Name: "IsBase", JSONName: "isbase", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Is base currency (default: THB)"},
				{Name: "DecimalDigit", JSONName: "decimaldigit", Type: "int", DartType: "int", TSType: "number", Description: "Decimal places (default: 2)"},
			},
		},

		// ===== NameX (common sub-model) =====
		{
			Name:        "NameX",
			Description: "Multi-language name — ชื่อหลายภาษา (ใช้ทุก model ที่มี names field)",
			Category:    "common",
			SourceFile:  "pkg/models/common.go",
			Fields: []FieldDef{
				{Name: "Code", JSONName: "code", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Language code (e.g., th, en, cn, jp)"},
				{Name: "Name", JSONName: "name", Type: "string", DartType: "String", TSType: "string", Required: true, Description: "Name in that language"},
				{Name: "IsDefault", JSONName: "isdefault", Type: "bool", DartType: "bool", TSType: "boolean", Description: "Is default language"},
			},
		},
	}
}
