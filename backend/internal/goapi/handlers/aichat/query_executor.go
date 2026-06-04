package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mydb"
	"strings"
)

// QueryResult represents the result from database query
type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Count   int             `json:"count"`
}

// ExecuteQuery runs the SQL query and returns results
func ExecuteQuery(ctx context.Context, holdingCode string, sql string, limit int) (*QueryResult, error) {
	db, err := mydb.GetGlobalConnectionFromPool(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	logger.Info("[Query] Executing: %s", sql)

	rows, err := db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Fetch rows
	var results [][]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold each column's value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert []byte to string for display
		for i, val := range values {
			if b, ok := val.([]byte); ok {
				values[i] = string(b)
			}
		}

		results = append(results, values)

		// Safety limit
		if len(results) >= limit && limit > 0 {
			break
		}
		if len(results) >= 100 { // Hard limit
			break
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	logger.Info("[Query] Found %d results", len(results))

	return &QueryResult{
		Columns: columns,
		Rows:    results,
		Count:   len(results),
	}, nil
}

// GenerateAnswerFromData uses AI to generate natural language answer from query results
func GenerateAnswerFromData(ctx context.Context, question string, intent *QueryIntent, queryResult *QueryResult) (string, string, []string, *aiprovider.ChatResponse, error) {
	ai := aiprovider.GetProvider()

	// Format query results as text
	var dataText strings.Builder
	if queryResult != nil && len(queryResult.Rows) > 0 {
		dataText.WriteString("ผลลัพธ์จากฐานข้อมูล:\n\n")

		// Add column headers
		dataText.WriteString("| ")
		for _, col := range queryResult.Columns {
			dataText.WriteString(col)
			dataText.WriteString(" | ")
		}
		dataText.WriteString("\n|")
		for range queryResult.Columns {
			dataText.WriteString("---|")
		}
		dataText.WriteString("\n")

		// Add rows
		for _, row := range queryResult.Rows {
			dataText.WriteString("| ")
			for _, val := range row {
				dataText.WriteString(fmt.Sprintf("%v", val))
				dataText.WriteString(" | ")
			}
			dataText.WriteString("\n")
		}

		dataText.WriteString(fmt.Sprintf("\n(จำนวนข้อมูล: %d รายการ)\n", queryResult.Count))
	} else if intent != nil && !intent.NeedsData {
		dataText.WriteString("คำถามนี้ไม่ต้องการข้อมูลจากฐานข้อมูล\n")
	} else {
		dataText.WriteString("ไม่พบข้อมูล\n")
	}

	systemPrompt := `คุณเป็น AI Assistant สำหรับระบบ POS ของร้านค้า

**หน้าที่:**
1. ตอบคำถามจากข้อมูลที่ได้รับอย่างชัดเจนและเป็นมิตร
2. ใช้ภาษาไทยที่เข้าใจง่าย พร้อม emoji ที่เหมาะสม
3. สร้าง HTML สวยงามพร้อม inline CSS

**HTML Requirements:**
- Background: #fff (สีขาว)
- Padding: 24px, Border-radius: 12px
- Font: system-ui, -apple-system, BlinkMacSystemFont, sans-serif
- Text color: #1f2937 (เทาเข้ม)
- Heading color: #111827 (ดำ)
- Accent color: #3b82f6 (น้ำเงิน)
- Box-shadow: 0 1px 3px rgba(0,0,0,0.1)
- ใช้ inline CSS เท่านั้น (ห้ามใช้ <style> tag แยก)

**Table Styling:**
- border-collapse: collapse
- width: 100%
- th: background #f3f4f6, padding 12px, text-align left
- td: padding 12px, border-bottom 1px solid #e5e7eb
- hover: background #f9fafb

**Badge/Tag:**
- display: inline-block
- padding: 4px 12px
- border-radius: 12px
- font-size: 14px
- success: background #d1fae5, color #065f46
- danger: background #fee2e2, color #991b1b
- info: background #dbeafe, color #1e40af

**Response Format (JSON):**
{
  "answer": "คำตอบแบบ plain text พร้อม emoji",
  "html": "<div style='background:#fff;padding:24px;...'>HTML content with inline CSS</div>",
  "suggested_questions": [
    "คำถามแนะนำข้อ 1",
    "คำถามแนะนำข้อ 2",
    "คำถามแนะนำข้อ 3",
    "คำถามแนะนำข้อ 4",
    "คำถามแนะนำข้อ 5"
  ]
}

**กฎสำคัญ:**
- ตอบตามข้อมูลที่มีเท่านั้น ไม่แต่งเติม
- ถ้าไม่มีข้อมูล บอกตรงๆ ว่าไม่พบ
- คำถามแนะนำต้องเกี่ยวข้องกับบริบท
- Output เป็น JSON เท่านั้น ไม่ต้องมี markdown`

	userPrompt := fmt.Sprintf("**คำถาม:** %s\n\n**ข้อมูล:**\n%s\n\nสร้าง JSON response:", question, dataText.String())

	resp, err := ai.GenerateContent(ctx, aiprovider.ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		TopP:         0.95,
	})
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	content := resp.Text

	// Clean response
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	logger.Info("AI Answer: %s", content[:min(200, len(content))])

	// Parse JSON response
	var aiResponse struct {
		Answer             string   `json:"answer"`
		HTML               string   `json:"html"`
		SuggestedQuestions []string `json:"suggested_questions"`
	}

	// Try to parse JSON
	if err := json.Unmarshal([]byte(content), &aiResponse); err != nil {
		// If JSON parse fails, return plain text
		return content, "", nil, resp, nil
	}

	return aiResponse.Answer, aiResponse.HTML, aiResponse.SuggestedQuestions, resp, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
