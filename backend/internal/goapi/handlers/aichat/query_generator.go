package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
)

// QueryIntent represents AI's interpretation of the question
type QueryIntent struct {
	QueryType    string `json:"query_type"`    // "count", "list", "search", "stat", "general"
	SQL          string `json:"sql"`           // Generated SQL query
	NeedsData    bool   `json:"needs_data"`    // Whether to execute query
	DirectAnswer string `json:"direct_answer"` // If no query needed
	Limit        int    `json:"limit"`         // Result limit
}

// GenerateQueryFromQuestion uses AI to interpret question and generate SQL
func GenerateQueryFromQuestion(ctx context.Context, question string, functionName string) (*QueryIntent, error) {
	ai := aiprovider.GetProvider()

	systemPrompt := `คุณเป็น SQL Query Generator สำหรับระบบ POS
	
**Database Schema:**
ตาราง product:
- itemcode (รหัสสินค้า)
- name0 (ชื่อสินค้า)  
- unitname (หน่วยนับ)

ตาราง productbarcode:
- itemcode (รหัสสินค้า)
- barcode (บาร์โค้ด)

**หน้าที่:**
1. วิเคราะห์คำถามของผู้ใช้
2. สร้าง SQL query ที่เหมาะสม
3. กำหนด query_type: "count" (นับจำนวน), "list" (แสดงรายการ), "search" (ค้นหา), "stat" (สถิติ), "general" (คำถามทั่วไป)

**ตัวอย่าง:**

คำถาม: "มีสินค้ากี่รายการ"
Response:
{
  "query_type": "count",
  "sql": "SELECT COUNT(*) as total FROM product WHERE itemcode IS NOT NULL",
  "needs_data": true,
  "limit": 0
}

คำถาม: "แสดงสินค้า MAKITA 5 รายการแรก"
Response:
{
  "query_type": "list",
  "sql": "SELECT p.itemcode, p.name0, p.unitname, COALESCE(STRING_AGG(DISTINCT pb.barcode, ','), '') as barcodes FROM product p LEFT JOIN productbarcode pb ON p.itemcode = pb.itemcode WHERE UPPER(p.name0) LIKE '%MAKITA%' GROUP BY p.itemcode, p.name0, p.unitname LIMIT 5",
  "needs_data": true,
  "limit": 5
}

คำถาม: "สินค้าไหนมีราคาแพงที่สุด"
Response:
{
  "query_type": "stat",
  "sql": "SELECT p.itemcode, p.name0, p.unitname FROM product p WHERE itemcode IS NOT NULL ORDER BY p.itemcode LIMIT 10",
  "needs_data": true,
  "limit": 10
}

**กฎสำคัญ:**
- ใช้ PostgreSQL syntax
- GROUP BY ต้องรวม p.itemcode, p.name0, p.unitname
- ใช้ COALESCE สำหรับ barcode
- LIMIT สูงสุด 100
- ถ้าคำถามไม่เกี่ยวข้องกับข้อมูล ให้ needs_data = false

**Output Format:** JSON เท่านั้น ไม่ต้องมี markdown หรือคำอธิบาย`

	userPrompt := fmt.Sprintf("คำถาม: %s\n\nสร้าง JSON response:", question)

	resp, err := ai.GenerateContent(ctx, aiprovider.ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.3, // Low temperature for consistent JSON output
		TopP:         0.95,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	content := resp.Text
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	logger.Info("AI Response: %s", content)

	var intent QueryIntent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w\nResponse: %s", err, content)
	}

	return &intent, nil
}
