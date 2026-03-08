package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// DocumentAnalysisRequest represents the request for document analysis
type DocumentAnalysisRequest struct {
	ShopID       string                 `json:"shop_id" validate:"required"`
	DocumentType string                 `json:"document_type" validate:"required"` // ประเภทเอกสาร
	DocumentData map[string]interface{} `json:"document_data" validate:"required"` // ข้อมูลเอกสาร
}

// DocumentAnalysisResponse represents the response
type DocumentAnalysisResponse struct {
	Success    bool                  `json:"success"`
	Message    string                `json:"message"`
	Analysis   *DocumentAnalysisData `json:"analysis,omitempty"`
	Error      string                `json:"error,omitempty"`
	Timestamp  time.Time             `json:"timestamp"`
	TokenUsage *TokenUsage           `json:"token_usage,omitempty"`
}

// DocumentAnalysisData contains the analysis result
type DocumentAnalysisData struct {
	Strengths       []string `json:"strengths"`        // จุดแข็ง
	Weaknesses      []string `json:"weaknesses"`       // จุดอ่อน
	Recommendations []string `json:"recommendations"`  // คำแนะนำ
	AdditionalCheck []string `json:"additional_check"` // ข้อมูลที่ควรตรวจสอบเพิ่มเติม
	Summary         string   `json:"summary"`          // สรุปภาพรวม
}

// AnalyzeDocument handles document analysis requests
// POST /api/v1/chatbot/analyze-document
func AnalyzeDocument(c echo.Context) error {
	var req DocumentAnalysisRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, DocumentAnalysisResponse{
			Success:   false,
			Message:   "Invalid request format",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
	}

	// Validate request
	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, DocumentAnalysisResponse{
			Success:   false,
			Message:   "shop_id is required",
			Timestamp: time.Now(),
		})
	}

	if req.DocumentType == "" {
		return c.JSON(http.StatusBadRequest, DocumentAnalysisResponse{
			Success:   false,
			Message:   "document_type is required",
			Timestamp: time.Now(),
		})
	}

	if req.DocumentData == nil {
		return c.JSON(http.StatusBadRequest, DocumentAnalysisResponse{
			Success:   false,
			Message:   "document_data is required",
			Timestamp: time.Now(),
		})
	}

	ctx := context.Background()
	logger.Info("[AnalyzeDocument] Document Type: %s (shop: %s)", req.DocumentType, req.ShopID)

	// Generate analysis using AI
	analysis, usage, err := generateDocumentAnalysis(ctx, req.DocumentType, req.DocumentData)
	if err != nil {
		logger.Error("Failed to analyze document: %v", err)
		return c.JSON(http.StatusInternalServerError, DocumentAnalysisResponse{
			Success:   false,
			Message:   "Failed to analyze document",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
	}

	return c.JSON(http.StatusOK, DocumentAnalysisResponse{
		Success:    true,
		Message:    "Document analyzed successfully",
		Analysis:   analysis,
		TokenUsage: usage,
		Timestamp:  time.Now(),
	})
}

// generateDocumentAnalysis uses AI to analyze document
func generateDocumentAnalysis(ctx context.Context, docType string, docData map[string]interface{}) (*DocumentAnalysisData, *TokenUsage, error) {
	ai := aiprovider.GetProvider()

	// สร้าง system prompt ตามประเภทเอกสาร
	systemPrompt := getDocumentAnalysisPrompt(docType)

	// สร้าง user prompt จากข้อมูลเอกสาร
	userPrompt := fmt.Sprintf(`วิเคราะห์เอกสาร %s นี้:

ข้อมูลเอกสาร:
%v

กรุณาวิเคราะห์และตอบกลับเป็น JSON ตามรูปแบบที่กำหนด`, docType, formatDocumentData(docData))

	resp, err := ai.GenerateContent(ctx, aiprovider.ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    2048,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("AI generation failed: %v", err)
	}

	// Parse response
	analysis := parseAnalysisResponse(resp.Text)

	usage := &TokenUsage{
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
		TotalTokens:      resp.TotalTokens,
		CostUSD:          0,
		CostTHB:          0,
		Model:            resp.Model,
	}

	return analysis, usage, nil
}

// getDocumentAnalysisPrompt returns the system prompt based on document type
func getDocumentAnalysisPrompt(docType string) string {
	basePrompt := `คุณเป็นผู้เชี่ยวชาญด้านการวิเคราะห์เอกสารทางธุรกิจ

**หน้าที่:**
วิเคราะห์เอกสารที่ได้รับและให้ข้อมูลดังนี้:
1. จุดแข็งของเอกสาร (strengths)
2. จุดอ่อน/ข้อควรระวัง (weaknesses)
3. คำแนะนำในการปรับปรุง (recommendations)
4. ข้อมูลที่ควรตรวจสอบเพิ่มเติม (additional_check)
5. สรุปภาพรวม (summary)

**Output Format:** JSON เท่านั้น ตามรูปแบบ:
{
  "strengths": ["จุดแข็ง 1", "จุดแข็ง 2"],
  "weaknesses": ["จุดอ่อน 1", "จุดอ่อน 2"],
  "recommendations": ["คำแนะนำ 1", "คำแนะนำ 2"],
  "additional_check": ["ตรวจสอบ 1", "ตรวจสอบ 2"],
  "summary": "สรุปภาพรวม"
}
`

	// เพิ่ม context ตามประเภทเอกสาร
	switch docType {
	case "purchase_order", "ใบสั่งซื้อ", "PO":
		return basePrompt + `
**ประเภทเอกสาร: ใบสั่งซื้อ (Purchase Order)**
ตรวจสอบ:
- ความครบถ้วนของข้อมูลเจ้าหนี้
- ราคาสินค้าเทียบกับราคาตลาด
- ส่วนลดที่ได้รับ
- เงื่อนไขการชำระเงิน
- วันที่ส่งมอบสินค้า`

	case "sales_order", "ใบสั่งขาย", "SO":
		return basePrompt + `
**ประเภทเอกสาร: ใบสั่งขาย (Sales Order)**
ตรวจสอบ:
- ความครบถ้วนของข้อมูลลูกค้า
- ราคาขายเทียบกับต้นทุน
- ส่วนลดที่ให้ลูกค้า (ไม่ควรเกิน margin)
- วงเงินเครดิตของลูกค้า
- สต็อกสินค้าคงเหลือ`

	case "invoice", "ใบแจ้งหนี้", "IV":
		return basePrompt + `
**ประเภทเอกสาร: ใบแจ้งหนี้ (Invoice)**
ตรวจสอบ:
- ความถูกต้องของข้อมูลภาษี
- การคำนวณยอดรวม
- ข้อมูลลูกค้าครบถ้วน
- เลขที่เอกสารไม่ซ้ำ
- วันครบกำหนดชำระ`

	case "receipt", "ใบเสร็จรับเงิน", "RC":
		return basePrompt + `
**ประเภทเอกสาร: ใบเสร็จรับเงิน (Receipt)**
ตรวจสอบ:
- ยอดรับชำระตรงกับใบแจ้งหนี้
- วิธีการชำระเงิน
- ข้อมูลผู้รับชำระ
- เลขที่เอกสารอ้างอิง`

	case "goods_receive", "ใบรับสินค้า", "GR":
		return basePrompt + `
**ประเภทเอกสาร: ใบรับสินค้า (Goods Receive)**
ตรวจสอบ:
- จำนวนสินค้าตรงกับใบสั่งซื้อ
- คุณภาพสินค้า
- เอกสารอ้างอิง (PO)
- ข้อมูลคลังสินค้าที่รับ`

	case "goods_issue", "ใบส่งสินค้า", "GI":
		return basePrompt + `
**ประเภทเอกสาร: ใบส่งสินค้า (Goods Issue)**
ตรวจสอบ:
- สต็อกเพียงพอสำหรับการส่ง
- ข้อมูลที่อยู่จัดส่ง
- เอกสารอ้างอิง (SO/IV)
- น้ำหนักและขนาดบรรจุ`

	case "transfer", "ใบโอนย้ายสินค้า", "TR":
		return basePrompt + `
**ประเภทเอกสาร: ใบโอนย้ายสินค้า (Stock Transfer)**
ตรวจสอบ:
- คลังต้นทางมีสต็อกเพียงพอ
- คลังปลายทางถูกต้อง
- เหตุผลในการโอนย้าย
- ต้นทุนสินค้า`

	case "adjustment", "ใบปรับปรุงสต็อก", "ADJ":
		return basePrompt + `
**ประเภทเอกสาร: ใบปรับปรุงสต็อก (Stock Adjustment)**
ตรวจสอบ:
- เหตุผลในการปรับปรุง
- จำนวนที่ปรับปรุง
- มูลค่าที่ได้รับผลกระทบ
- ผู้อนุมัติการปรับปรุง`

	case "quotation", "ใบเสนอราคา", "QT":
		return basePrompt + `
**ประเภทเอกสาร: ใบเสนอราคา (Quotation)**
ตรวจสอบ:
- ราคาและส่วนลดเหมาะสม
- เงื่อนไขการชำระเงิน
- ระยะเวลาที่ใบเสนอราคามีผล
- ข้อมูลลูกค้าครบถ้วน`

	default:
		return basePrompt + `
**ประเภทเอกสาร: ` + docType + `**
ตรวจสอบ:
- ความครบถ้วนของข้อมูล
- ความถูกต้องของการคำนวณ
- เอกสารอ้างอิง
- ข้อมูลผู้เกี่ยวข้อง`
	}
}

// formatDocumentData formats document data for the prompt
func formatDocumentData(data map[string]interface{}) string {
	result := ""
	for key, value := range data {
		result += fmt.Sprintf("- %s: %v\n", key, value)
	}
	return result
}

// parseAnalysisResponse parses AI response to DocumentAnalysisData
func parseAnalysisResponse(response string) *DocumentAnalysisData {
	analysis := &DocumentAnalysisData{
		Strengths:       []string{},
		Weaknesses:      []string{},
		Recommendations: []string{},
		AdditionalCheck: []string{},
		Summary:         "",
	}

	// ลบ markdown code block ถ้ามี (```json ... ```)
	jsonStr := response
	if start := strings.Index(response, "```json"); start != -1 {
		jsonStr = response[start+7:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	} else if start := strings.Index(response, "```"); start != -1 {
		jsonStr = response[start+3:]
		if end := strings.Index(jsonStr, "```"); end != -1 {
			jsonStr = jsonStr[:end]
		}
	}
	jsonStr = strings.TrimSpace(jsonStr)

	// Parse JSON response
	var parsed struct {
		Strengths       []string `json:"strengths"`
		Weaknesses      []string `json:"weaknesses"`
		Recommendations []string `json:"recommendations"`
		AdditionalCheck []string `json:"additional_check"`
		Summary         string   `json:"summary"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		logger.Error("[parseAnalysisResponse] Failed to parse JSON: %v", err)
		analysis.Summary = response // ถ้า parse ไม่ได้ ใส่ raw response
		return analysis
	}

	analysis.Strengths = parsed.Strengths
	analysis.Weaknesses = parsed.Weaknesses
	analysis.Recommendations = parsed.Recommendations
	analysis.AdditionalCheck = parsed.AdditionalCheck
	analysis.Summary = parsed.Summary

	return analysis
}
