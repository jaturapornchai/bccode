package approval

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// =====================================================
// Email & LINE Push Notification Service
// =====================================================

// Collection สำหรับเก็บ approval tokens
const ApprovalTokensCollection = "poapprovaltokens"

// ApprovalToken เก็บ token สำหรับอนุมัติผ่าน email/LINE
type ApprovalToken struct {
	ID           primitive.ObjectID `bson:"id,omitempty" json:"guid,omitempty"`
	Token        string             `bson:"token" json:"token"`
	HoldingCode  string             `bson:"holdingcode" json:"holdingcode"`
	DocNo        string             `bson:"docno" json:"docno"`
	GuidFixed    string             `bson:"guidfixed" json:"guidfixed"`
	ApproverCode string             `bson:"approvercode" json:"approvercode"`
	ApproverName string             `bson:"approvername" json:"approvername"`
	Action       string             `bson:"action" json:"action"` // approve, reject
	Used         bool               `bson:"used" json:"used"`
	UsedAt       *time.Time         `bson:"usedat,omitempty" json:"usedat,omitempty"`
	ExpiresAt    time.Time          `bson:"expiresat" json:"expiresat"`
	CreatedAt    time.Time          `bson:"createdat" json:"createdat"`
}

// getBrevoAPIKey ดึง Brevo API Key จาก environment variable
func getBrevoAPIKey() string {
	return os.Getenv("BREVO_API_KEY")
}

// isBrevoConfigured ตรวจสอบว่า Brevo ถูกตั้งค่าหรือไม่
func isBrevoConfigured() bool {
	return getBrevoAPIKey() != ""
}

// BrevoEmailRequest โครงสร้าง request สำหรับ Brevo Email API
type BrevoEmailRequest struct {
	Sender      BrevoContact   `json:"sender"`
	To          []BrevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlcontent"`
}

// BrevoContact ข้อมูลผู้ส่ง/ผู้รับ
type BrevoContact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

// sendBrevoEmail ส่ง email ผ่าน Brevo API
func sendBrevoEmail(toEmail, toName, subject, htmlBody string) error {
	apiKey := getBrevoAPIKey()
	if apiKey == "" {
		return fmt.Errorf("Brevo API key not configured")
	}

	// กำหนด sender จาก environment หรือใช้ค่า default
	senderEmail := os.Getenv("BREVO_FROM_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@bcai.cloud"
	}
	senderName := os.Getenv("BREVO_FROM_NAME")
	if senderName == "" {
		senderName = "BC Ai Account"
	}

	emailReq := BrevoEmailRequest{
		Sender: BrevoContact{
			Name:  senderName,
			Email: senderEmail,
		},
		To: []BrevoContact{
			{Name: toName, Email: toEmail},
		},
		Subject:     subject,
		HTMLContent: htmlBody,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// ส่ง request ไปยัง Brevo API
	httpReq, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("api-key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("[Brevo] API error: %s", string(body))
		return fmt.Errorf("Brevo API returned %d: %s", resp.StatusCode, string(body))
	}

	logger.Info("[Brevo] Email sent successfully to %s", toEmail)
	return nil
}

// generateApprovalToken สร้าง token สำหรับอนุมัติ
func generateApprovalToken(holdingCode, docNo, guidFixed, approverCode, approverName, action string) (string, error) {
	if !IsConnected() {
		return "", fmt.Errorf("MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getTokenCollection(ApprovalTokensCollection)
	now := time.Now().UTC() // เก็บเวลาเป็น UTC+0

	// สร้าง unique token (cryptographically random เพื่อป้องกันการเดา)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %v", err)
	}
	token := hex.EncodeToString(randomBytes)

	tokenDoc := ApprovalToken{
		Token:        token,
		HoldingCode:  holdingCode,
		DocNo:        docNo,
		GuidFixed:    guidFixed,
		ApproverCode: approverCode,
		ApproverName: approverName,
		Action:       action,
		Used:         false,
		ExpiresAt:    now.Add(72 * time.Hour), // หมดอายุ 72 ชั่วโมง
		CreatedAt:    now,
	}

	_, err := collection.InsertOne(ctx, tokenDoc)
	if err != nil {
		return "", fmt.Errorf("failed to create approval token: %v", err)
	}

	return token, nil
}

// SendApprovalEmail ส่ง email แจ้งเตือนการอนุมัติ
// isModified = true จะแสดงข้อความว่ามีการแก้ไขเอกสาร
func SendApprovalEmail(toEmail, toName, holdingCode, docNo, totalAmount, purchaseTypeName, createdByName, approveURL, rejectURL, pdfURL string, isModified bool) error {
	// สร้างเนื้อหา email HTML
	subject := fmt.Sprintf("รออนุมัติใบสั่งซื้อ: %s", docNo)
	modifiedBadge := ""
	modifiedNote := "มีใบสั่งซื้อรอการอนุมัติจากคุณ:"

	if isModified {
		subject = fmt.Sprintf("[แก้ไข] รออนุมัติใบสั่งซื้อ: %s", docNo)
		modifiedBadge = `<div style="background-color: #FF9800; color: white; padding: 10px; text-align: center; border-radius: 5px; margin-bottom: 15px; font-weight: bold;">⚠️ เอกสารนี้มีการแก้ไข กรุณาตรวจสอบใหม่</div>`
		modifiedNote = "มีการแก้ไขใบสั่งซื้อ กรุณาตรวจสอบและอนุมัติอีกครั้ง:"
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Sarabun', sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; padding: 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; border-bottom: 2px solid #1976D2; padding-bottom: 20px; margin-bottom: 20px; }
        .header h1 { color: #1976D2; margin: 0; font-size: 24px; }
        .info-table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
        .info-table td { padding: 10px; border-bottom: 1px solid #eee; }
        .info-table td:first-child { color: #666; width: 40%%; }
        .info-table td:last-child { font-weight: bold; }
        .amount { font-size: 28px; color: #4CAF50; text-align: center; margin: 20px 0; }
        .buttons { text-align: center; margin: 30px 0; }
        .btn { display: inline-block; padding: 15px 40px; margin: 5px; border-radius: 5px; text-decoration: none; font-weight: bold; font-size: 16px; }
        .btn-approve { background-color: #4CAF50; color: white; }
        .btn-reject { background-color: #f44336; color: white; }
        .btn-pdf { background-color: #2196F3; color: white; }
        .footer { text-align: center; color: #999; font-size: 12px; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>รออนุมัติใบสั่งซื้อ</h1>
        </div>

        %s

        <p>เรียน คุณ%s,</p>
        <p>%s</p>

        <table class="info-table">
            <tr><td>เลขที่เอกสาร:</td><td>%s</td></tr>
            <tr><td>ประเภทการจัดซื้อ:</td><td>%s</td></tr>
            <tr><td>ผู้สร้างเอกสาร:</td><td>%s</td></tr>
        </table>

        <div class="amount">฿%s</div>

        <div class="buttons">
            <a href="%s" class="btn btn-approve">อนุมัติ</a>
            <a href="%s" class="btn btn-reject">ปฏิเสธ</a>
        </div>

        <div class="buttons">
            <a href="%s" class="btn btn-pdf">ดูรายละเอียด PDF</a>
        </div>

        <div class="footer">
            <p>Email นี้ส่งจากระบบ BC Ai Account</p>
            <p>กรุณาอย่าตอบกลับ email นี้</p>
        </div>
    </div>
</body>
</html>
`, modifiedBadge, toName, modifiedNote, docNo, purchaseTypeName, createdByName, totalAmount, approveURL, rejectURL, pdfURL)

	// ส่งผ่าน Brevo API เท่านั้น
	if !isBrevoConfigured() {
		return fmt.Errorf("Brevo API key not configured (BREVO_API_KEY)")
	}

	err := sendBrevoEmail(toEmail, toName, subject, htmlBody)
	if err != nil {
		logger.Error("[Email] Brevo failed to send to %s: %v", toEmail, err)
		return err
	}

	logger.Info("[Email] Sent via Brevo to %s for PO %s", toEmail, docNo)
	return nil
}

// LINE Push API structures
type LinePushRequest struct {
	To       string    `json:"to"`
	Messages []LineMsg `json:"messages"`
}

type LineMsg struct {
	Type     string      `json:"type"`
	AltText  string      `json:"alttext,omitempty"`
	Contents interface{} `json:"contents,omitempty"`
	Text     string      `json:"text,omitempty"`
}

// LineBubble สำหรับ Flex Message
type LineBubble struct {
	Type   string   `json:"type"`
	Header *LineBox `json:"header,omitempty"`
	Body   *LineBox `json:"body,omitempty"`
	Footer *LineBox `json:"footer,omitempty"`
}

type LineBox struct {
	Type     string        `json:"type"`
	Layout   string        `json:"layout"`
	Contents []interface{} `json:"contents"`
	Spacing  string        `json:"spacing,omitempty"`
}

type LineText struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Weight string `json:"weight,omitempty"`
	Size   string `json:"size,omitempty"`
	Color  string `json:"color,omitempty"`
	Align  string `json:"align,omitempty"`
	Wrap   bool   `json:"wrap,omitempty"`
	Margin string `json:"margin,omitempty"`
}

type LineButton struct {
	Type   string           `json:"type"`
	Style  string           `json:"style,omitempty"`
	Color  string           `json:"color,omitempty"`
	Action LineButtonAction `json:"action"`
	Height string           `json:"height,omitempty"`
	Margin string           `json:"margin,omitempty"`
}

type LineButtonAction struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	URI   string `json:"uri,omitempty"`
}

// =====================================================
// Hardcoded LINE OA Configuration
// =====================================================
const (
	// LINE Channel Access Token (Long-lived) - จาก LINE Developers Console
	hardcodedLineAccessToken = "MF8p1AEMtiL2wtrq8P1R8YzYA2ijLdiBwQCHz2269rxYumiqUFN+aqMxBTEV9ZM0GDLIA+AN8yB75RucLGyYO6YRXzutAiOAE8HXYOnId2ANufHlEugf7CZJ/wQlA7W2R05zD8mIznLzM4OUjkqSTAdB04t89/1O/w1cDnyilFU="
	// LIFF ID - จาก LINE Developers Console
	hardcodedLiffID = "2008792333-tziDpgnJ"
)

// getLineAccessToken ดึง LINE access token (hardcoded)
func getLineAccessToken(holdingCode string) (string, error) {
	// ใช้ค่า hardcoded แทนการดึงจาก database
	logger.Info("[LINE Config] Using hardcoded access token for shop: %s", holdingCode)
	return hardcodedLineAccessToken, nil
}

// getLiffID ดึง LIFF ID (hardcoded)
func getLiffID(holdingCode string) (string, error) {
	// ใช้ค่า hardcoded แทนการดึงจาก database
	logger.Info("[LINE Config] Using hardcoded LIFF ID for shop: %s", holdingCode)
	return hardcodedLiffID, nil
}

// ApprovalNotificationParams พารามิเตอร์สำหรับส่งแจ้งเตือนขออนุมัติ
type ApprovalNotificationParams struct {
	LineUserID       string // LINE User ID ผู้รับ
	HoldingCode      string // รหัสร้าน
	DocNo            string // เลขที่เอกสาร
	DocDatetime      string // วันที่เอกสาร (string format)
	TotalAmount      string // ยอดรวม (formatted string)
	PurchaseTypeName string // ประเภทการจัดซื้อ
	CreatedByName    string // ชื่อผู้สร้างเอกสาร
	CustName         string // ชื่อเจ้าหนี้/ผู้ขาย
	LiffApproveURL   string // URL สำหรับอนุมัติ
	IsModified       bool   // เป็นการแก้ไขหรือไม่
	POComment        string // หมายเหตุจากใบสั่งซื้อ
}

// SendLinePushApproval ส่ง LINE push message สำหรับอนุมัติ
func SendLinePushApproval(lineUserID, holdingCode, docNo, totalAmount, purchaseTypeName, createdByName, liffApproveURL string, isModified bool) error {
	// เรียกใช้ฟังก์ชันใหม่โดยไม่มี docDatetime และ custName
	return SendLinePushApprovalV2(ApprovalNotificationParams{
		LineUserID:       lineUserID,
		HoldingCode:      holdingCode,
		DocNo:            docNo,
		DocDatetime:      "",
		TotalAmount:      totalAmount,
		PurchaseTypeName: purchaseTypeName,
		CreatedByName:    createdByName,
		CustName:         "",
		LiffApproveURL:   liffApproveURL,
		IsModified:       isModified,
	})
}

// SendLinePushApprovalV2 ส่ง LINE push message สำหรับอนุมัติ (version 2 รองรับข้อมูลเพิ่มเติม)
func SendLinePushApprovalV2(params ApprovalNotificationParams) error {
	// Log เพื่อ debug
	logger.Info("[LINE Push] SendLinePushApprovalV2 called - docNo: %s, lineUserID: %s, isModified: %v", params.DocNo, params.LineUserID, params.IsModified)

	accessToken, err := getLineAccessToken(params.HoldingCode)
	if err != nil {
		logger.Error("[LINE Push] Failed to get access token for shop %s: %v", params.HoldingCode, err)
		return err
	}

	// ฟังก์ชันจัดรูปแบบตัวเลขให้มี comma (รับ string เช่น "12345.00" และเพิ่ม comma เป็น "12,345.00")
	formatCurrencyFromString := func(amountStr string) string {
		// แยก integer part และ decimal part
		parts := strings.Split(amountStr, ".")
		intPart := parts[0]
		decPart := "00"
		if len(parts) > 1 {
			decPart = parts[1]
		}

		// เพิ่ม comma ใน integer part
		n := len(intPart)
		if n <= 3 {
			return fmt.Sprintf("%s.%s", intPart, decPart)
		}

		var result []byte
		for i, c := range intPart {
			if i > 0 && (n-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		return fmt.Sprintf("%s.%s", string(result), decPart)
	}

	// Format TotalAmount ให้มี comma
	formattedTotalAmount := formatCurrencyFromString(params.TotalAmount)

	// กำหนดข้อความ header และ badge ตามสถานะการแก้ไข
	headerText := "📋 รออนุมัติใบสั่งซื้อ"
	headerTextColor := "#FFFFFF"
	if params.IsModified {
		headerText = "⚠️ มีการแก้ไขเอกสาร"
		headerTextColor = "#FF9800"
		logger.Info("[LINE Push] Document is MODIFIED - using warning header: %s", headerText)
	} else {
		logger.Info("[LINE Push] Document is NEW - using normal header: %s", headerText)
	}

	// เวลาปัจจุบันในโซน ICT (UTC+7)
	bangkokTime := time.Now().In(time.FixedZone("ICT", 7*60*60))
	sentTimeText := bangkokTime.Format("15:04 น.")
	sentDateText := bangkokTime.Format("02/01/2006")

	// แปลงวันที่เอกสาร
	docDateText := "-"
	docTimeText := ""
	if params.DocDatetime != "" {
		// ลองแปลง format ต่างๆ
		layouts := []string{"2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02T15:04:05Z"}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, params.DocDatetime); err == nil {
				bangkokDocTime := t.In(time.FixedZone("ICT", 7*60*60))
				docDateText = bangkokDocTime.Format("02/01/2006")
				docTimeText = bangkokDocTime.Format("15:04 น.")
				break
			}
		}
	}

	// สร้าง Body contents
	bodyContents := []interface{}{}

	// ถ้ามีการแก้ไข แสดง badge เตือน
	if params.IsModified {
		bodyContents = append(bodyContents, LineText{
			Type:   "text",
			Text:   "กรุณาตรวจสอบและอนุมัติใหม่",
			Size:   "sm",
			Color:  "#FF9800",
			Weight: "bold",
			Align:  "center",
			Margin: "sm",
		})
	}

	// เลขที่เอกสาร
	bodyContents = append(bodyContents, LineText{
		Type:   "text",
		Text:   params.DocNo,
		Weight: "bold",
		Size:   "xl",
		Color:  "#1976D2",
		Align:  "center",
	})

	// วันที่เอกสาร
	docDateDisplay := fmt.Sprintf("📅 วันที่เอกสาร: %s", docDateText)
	if docTimeText != "" {
		docDateDisplay += " " + docTimeText
	}
	bodyContents = append(bodyContents, LineText{
		Type:   "text",
		Text:   docDateDisplay,
		Size:   "sm",
		Color:  "#666666",
		Align:  "center",
		Margin: "sm",
	})

	// ยอดรวม
	bodyContents = append(bodyContents, LineText{
		Type:   "text",
		Text:   fmt.Sprintf("฿%s", formattedTotalAmount),
		Weight: "bold",
		Size:   "xxl",
		Color:  "#4CAF50",
		Align:  "center",
		Margin: "md",
	})

	// Separator
	bodyContents = append(bodyContents, LineText{
		Type:   "text",
		Text:   "─────────────────",
		Size:   "xs",
		Color:  "#CCCCCC",
		Align:  "center",
		Margin: "lg",
	})

	// ประเภทการจัดซื้อ
	bodyContents = append(bodyContents, LineText{
		Type:  "text",
		Text:  fmt.Sprintf("📋 ประเภท: %s", params.PurchaseTypeName),
		Size:  "sm",
		Color: "#333333",
		Wrap:  true,
	})

	// ผู้ขาย/เจ้าหนี้
	if params.CustName != "" {
		bodyContents = append(bodyContents, LineText{
			Type:   "text",
			Text:   fmt.Sprintf("🏢 ผู้ขาย: %s", params.CustName),
			Size:   "sm",
			Color:  "#333333",
			Wrap:   true,
			Margin: "sm",
		})
	}

	// ผู้สร้าง
	bodyContents = append(bodyContents, LineText{
		Type:   "text",
		Text:   fmt.Sprintf("👤 ผู้ขออนุมัติ: %s", params.CreatedByName),
		Size:   "sm",
		Color:  "#333333",
		Margin: "sm",
	})

	// เวลาที่ส่งคำขอ
	bodyContents = append(bodyContents, LineText{
		Type:   "text",
		Text:   fmt.Sprintf("🕐 ส่งเมื่อ: %s %s", sentDateText, sentTimeText),
		Size:   "sm",
		Color:  "#666666",
		Margin: "sm",
	})

	// หมายเหตุจากใบสั่งซื้อ (ถ้ามี)
	if params.POComment != "" {
		bodyContents = append(bodyContents, LineText{
			Type:   "text",
			Text:   "─────────────────",
			Size:   "xs",
			Color:  "#CCCCCC",
			Align:  "center",
			Margin: "lg",
		})
		bodyContents = append(bodyContents, LineText{
			Type:   "text",
			Text:   "💬 หมายเหตุ:",
			Size:   "sm",
			Color:  "#666666",
			Margin: "sm",
		})
		bodyContents = append(bodyContents, LineText{
			Type:   "text",
			Text:   params.POComment,
			Size:   "sm",
			Color:  "#333333",
			Wrap:   true,
			Margin: "sm",
		})
	}

	// สร้าง Flex Message สำหรับแจ้งเตือน
	flexBubble := LineBubble{
		Type: "bubble",
		Header: &LineBox{
			Type:   "box",
			Layout: "vertical",
			Contents: []interface{}{
				LineText{
					Type:   "text",
					Text:   headerText,
					Weight: "bold",
					Size:   "lg",
					Color:  headerTextColor,
					Align:  "center",
				},
			},
		},
		Body: &LineBox{
			Type:     "box",
			Layout:   "vertical",
			Spacing:  "md",
			Contents: bodyContents,
		},
		Footer: &LineBox{
			Type:    "box",
			Layout:  "vertical",
			Spacing: "sm",
			Contents: []interface{}{
				LineButton{
					Type:   "button",
					Style:  "primary",
					Color:  "#4CAF50",
					Height: "md",
					Action: LineButtonAction{
						Type:  "uri",
						Label: "ดูรายละเอียดและอนุมัติ",
						URI:   params.LiffApproveURL,
					},
				},
			},
		},
	}

	pushReq := LinePushRequest{
		To: params.LineUserID,
		Messages: []LineMsg{
			{
				Type:     "flex",
				AltText:  fmt.Sprintf("รออนุมัติใบสั่งซื้อ %s - ฿%s", params.DocNo, formattedTotalAmount),
				Contents: flexBubble,
			},
		},
	}

	jsonData, err := json.Marshal(pushReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// ส่งไปยัง LINE Push API
	httpReq, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/push", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("[LINE Push] API error: %s", string(body))
		return fmt.Errorf("LINE API returned %d: %s", resp.StatusCode, string(body))
	}

	logger.Info("[LINE Push] Sent approval notification to %s for PO %s", params.LineUserID, params.DocNo)
	return nil
}

// OpenedNotificationParams พารามิเตอร์สำหรับส่งแจ้งเตือนการเปิดอ่าน
type OpenedNotificationParams struct {
	HoldingCode      string    // รหัสร้าน
	DocNo            string    // เลขที่เอกสาร
	DocDatetime      time.Time // วันที่เอกสาร
	CreatorCode      string    // รหัสผู้สร้างเอกสาร
	CreatorName      string    // ชื่อผู้สร้างเอกสาร
	OpenerCode       string    // รหัสผู้เปิดอ่าน
	OpenerName       string    // ชื่อผู้เปิดอ่าน
	OpenedFrom       string    // ช่องทางที่เปิดอ่าน (line, email, liff, app)
	PurchaseTypeName string    // ประเภทการจัดซื้อ
	TotalAmount      float64   // ยอดรวม
	CreditorName     string    // ชื่อเจ้าหนี้/ผู้ขาย
}

// SendLineNotifyCreatorOpened ส่ง LINE push message แจ้งผู้สร้างเอกสารว่ามีคนเปิดอ่าน PO แล้ว
// ฟังก์ชันนี้จะค้นหา LINE User ID ของผู้สร้างจาก lineoalinkedaccounts
func SendLineNotifyCreatorOpened(params OpenedNotificationParams) error {
	logger.Info("[LINE Push] SendLineNotifyCreatorOpened called - docNo: %s, creatorCode: %s, openerName: %s, totalAmount: %.2f", params.DocNo, params.CreatorCode, params.OpenerName, params.TotalAmount)

	// ตรวจสอบว่า MongoDB เชื่อมต่ออยู่
	if atlasClient == nil || atlasDB == nil {
		logger.Warn("[LINE Push] MongoDB not connected, skipping notification to creator")
		return fmt.Errorf("MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ค้นหา LINE User ID ของผู้สร้างจาก lineoalinkedaccounts (ใน lineoa database)
	lineoaDB := atlasClient.Database("lineoa")
	linkedAccountsCollection := lineoaDB.Collection("lineoalinkedaccounts")

	var creatorAccount struct {
		LineUserID string `bson:"lineuserid"`
		Status     string `bson:"status"`
	}

	err := linkedAccountsCollection.FindOne(ctx, bson.M{
		"holdingcode":   params.HoldingCode,
		"employee_code": params.CreatorCode,
		"status":        "linked",
	}).Decode(&creatorAccount)

	if err != nil {
		logger.Info("[LINE Push] Creator %s not linked to LINE or not found: %v", params.CreatorCode, err)
		return fmt.Errorf("creator not linked to LINE: %v", err)
	}

	if creatorAccount.LineUserID == "" {
		logger.Info("[LINE Push] Creator %s has no LINE User ID", params.CreatorCode)
		return fmt.Errorf("creator has no LINE User ID")
	}

	// ดึง access token (hardcoded)
	accessToken, err := getLineAccessToken(params.HoldingCode)
	if err != nil {
		logger.Error("[LINE Push] Failed to get access token for shop %s: %v", params.HoldingCode, err)
		return err
	}

	// กำหนดข้อความตามช่องทางที่เปิดอ่าน
	openedFromText := "ระบบ"
	switch params.OpenedFrom {
	case "line":
		openedFromText = "LINE"
	case "email":
		openedFromText = "Email"
	case "liff":
		openedFromText = "LINE LIFF"
	case "app":
		openedFromText = "แอพ"
	}

	// ฟังก์ชันจัดรูปแบบตัวเลขให้มี comma
	formatCurrency := func(amount float64) string {
		// แปลงเป็น string พร้อม comma และทศนิยม 2 ตำแหน่ง
		intPart := int64(amount)
		decPart := int64((amount - float64(intPart)) * 100)

		// เพิ่ม comma ใน integer part
		str := fmt.Sprintf("%d", intPart)
		n := len(str)
		if n <= 3 {
			return fmt.Sprintf("%s.%02d", str, decPart)
		}

		var result []byte
		for i, c := range str {
			if i > 0 && (n-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		return fmt.Sprintf("%s.%02d", string(result), decPart)
	}

	// เวลาปัจจุบันในโซน ICT (UTC+7)
	bangkokTime := time.Now().In(time.FixedZone("ICT", 7*60*60))
	openedTimeText := bangkokTime.Format("15:04 น.")
	openedDateText := bangkokTime.Format("02/01/2006")

	// วันที่เอกสารในโซน ICT (UTC+7)
	docDatetimeBangkok := params.DocDatetime.In(time.FixedZone("ICT", 7*60*60))
	docDateText := docDatetimeBangkok.Format("02/01/2006")
	docTimeText := docDatetimeBangkok.Format("15:04 น.")

	// สร้าง Flex Message แจ้งเตือนการเปิดอ่าน (แบบละเอียด)
	flexBubble := LineBubble{
		Type: "bubble",
		Header: &LineBox{
			Type:   "box",
			Layout: "vertical",
			Contents: []interface{}{
				LineText{
					Type:   "text",
					Text:   "📬 ใบสั่งซื้อถูกเปิดอ่านแล้ว",
					Weight: "bold",
					Size:   "lg",
					Color:  "#FFFFFF",
					Align:  "center",
				},
			},
		},
		Body: &LineBox{
			Type:    "box",
			Layout:  "vertical",
			Spacing: "md",
			Contents: []interface{}{
				// เลขที่เอกสาร
				LineText{
					Type:   "text",
					Text:   params.DocNo,
					Weight: "bold",
					Size:   "xl",
					Color:  "#1976D2",
					Align:  "center",
				},
				// วันที่เอกสาร
				LineText{
					Type:   "text",
					Text:   fmt.Sprintf("📅 วันที่เอกสาร: %s %s", docDateText, docTimeText),
					Size:   "sm",
					Color:  "#666666",
					Align:  "center",
					Margin: "sm",
				},
				// ยอดรวม
				LineText{
					Type:   "text",
					Text:   fmt.Sprintf("฿%s", formatCurrency(params.TotalAmount)),
					Weight: "bold",
					Size:   "xxl",
					Color:  "#4CAF50",
					Align:  "center",
					Margin: "md",
				},
				// Separator
				LineText{
					Type:   "text",
					Text:   "─────────────────",
					Size:   "xs",
					Color:  "#CCCCCC",
					Align:  "center",
					Margin: "lg",
				},
				// ประเภทการจัดซื้อ
				LineText{
					Type:  "text",
					Text:  fmt.Sprintf("📋 ประเภท: %s", params.PurchaseTypeName),
					Size:  "sm",
					Color: "#333333",
					Wrap:  true,
				},
				// เจ้าหนี้/ผู้ขาย
				LineText{
					Type:   "text",
					Text:   fmt.Sprintf("🏢 ผู้ขาย: %s", params.CreditorName),
					Size:   "sm",
					Color:  "#333333",
					Wrap:   true,
					Margin: "sm",
				},
				// Separator
				LineText{
					Type:   "text",
					Text:   "─────────────────",
					Size:   "xs",
					Color:  "#CCCCCC",
					Align:  "center",
					Margin: "lg",
				},
				// ผู้เปิดอ่าน
				LineText{
					Type:   "text",
					Text:   fmt.Sprintf("👤 ผู้เปิดอ่าน: %s", params.OpenerName),
					Size:   "md",
					Color:  "#1976D2",
					Weight: "bold",
					Margin: "sm",
				},
				// ช่องทาง
				LineText{
					Type:   "text",
					Text:   fmt.Sprintf("📱 ช่องทาง: %s", openedFromText),
					Size:   "sm",
					Color:  "#666666",
					Margin: "sm",
				},
				// เวลาที่เปิดอ่าน
				LineText{
					Type:   "text",
					Text:   fmt.Sprintf("🕐 เปิดอ่านเมื่อ: %s %s", openedDateText, openedTimeText),
					Size:   "sm",
					Color:  "#666666",
					Margin: "sm",
				},
			},
		},
	}

	pushReq := LinePushRequest{
		To: creatorAccount.LineUserID,
		Messages: []LineMsg{
			{
				Type:     "flex",
				AltText:  fmt.Sprintf("ใบสั่งซื้อ %s ถูกเปิดอ่านแล้วโดย %s", params.DocNo, params.OpenerName),
				Contents: flexBubble,
			},
		},
	}

	jsonData, err := json.Marshal(pushReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// ส่งไปยัง LINE Push API
	httpReq, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/push", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("[LINE Push] API error: %s", string(body))
		return fmt.Errorf("LINE API returned %d: %s", resp.StatusCode, string(body))
	}

	logger.Info("[LINE Push] Sent 'opened' notification to creator %s (%s) for PO %s", params.CreatorName, params.CreatorCode, params.DocNo)
	return nil
}

// =====================================================
// Enhanced Notification Handler with Real Sending
// =====================================================

// SendRealNotificationsRequest request สำหรับส่งแจ้งเตือนจริง
type SendRealNotificationsRequest struct {
	HoldingCode string `json:"holdingcode"`
	DocNo       string `json:"docno"`
	GuidFixed   string `json:"guidfixed"`
	BaseURL     string `json:"baseurl"` // Base URL สำหรับสร้าง link (e.g., https://erp.example.com)
}

// SendRealApprovalNotificationHandler - ส่งแจ้งเตือนจริงผ่าน Email และ LINE Push
func SendRealApprovalNotificationHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req SendRealNotificationsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.HoldingCode == "" || req.DocNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "holdingcode and docno are required",
		})
	}

	// Base URL fallback
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("APP_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "https://erp.bcai.cloud"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 1. ดึงสถานะการอนุมัติ PO
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	var poStatus POApprovalStatus
	err := statusCollection.FindOne(ctx, bson.M{
		"holdingcode": req.HoldingCode,
		"docno":       req.DocNo,
		"status":      "pending",
	}).Decode(&poStatus)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"message": "ไม่พบ PO ที่รออนุมัติ",
			"sent":    0,
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// 2. ดึงกฎการอนุมัติ
	settingsCollection := getCollection(POApprovalSettingsCollection)
	var setting POApprovalSetting
	err = settingsCollection.FindOne(ctx, bson.M{
		"holdingcode":      req.HoldingCode,
		"purchasetypecode": poStatus.PurchaseTypeCode,
	}).Decode(&setting)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"message": "ไม่พบกฎการอนุมัติ",
			"sent":    0,
		})
	}

	// 3. หาผู้อนุมัติที่เกี่ยวข้อง
	var approvers []ApproverInfo
	for _, rule := range setting.Rules {
		if poStatus.TotalAmount >= rule.MinAmount {
			if rule.MaxAmount == 0 || poStatus.TotalAmount <= rule.MaxAmount {
				if rule.ApprovalLevel == poStatus.RequiredLevel && len(rule.Approvers) > 0 {
					approvers = rule.Approvers
					break
				}
			}
		}
	}

	if len(approvers) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"message": "ไม่พบรายชื่อผู้อนุมัติ",
			"sent":    0,
		})
	}

	// 4. ส่งแจ้งเตือนจริง
	notificationCollection := getCollection(POApprovalNotificationLogCollection)
	now := time.Now().UTC() // เก็บเวลาเป็น UTC+0
	totalAmount := fmt.Sprintf("%.2f", poStatus.TotalAmount)

	// ดึง LIFF ID สำหรับ LINE
	liffID, _ := getLiffID(req.HoldingCode)

	results := []map[string]any{}
	emailSent := 0
	lineSent := 0
	skipped := 0

	for _, approver := range approvers {
		// ตรวจสอบว่าเคยส่งสำเร็จแล้วหรือยัง
		existingFilter := bson.M{
			"holdingcode":  req.HoldingCode,
			"docno":        req.DocNo,
			"approvercode": approver.UserCode,
			"status":       "sent",
		}
		existingCount, _ := notificationCollection.CountDocuments(ctx, existingFilter)
		if existingCount > 0 {
			skipped++
			results = append(results, map[string]any{
				"approver": approver.UserCode,
				"status":   "skipped",
				"reason":   "already_sent",
			})
			continue
		}

		// ส่ง Email ถ้ามี
		if approver.Email != "" {
			// สร้าง approval tokens
			approveToken, _ := generateApprovalToken(req.HoldingCode, req.DocNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "approve")
			rejectToken, _ := generateApprovalToken(req.HoldingCode, req.DocNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "reject")

			approveURL := fmt.Sprintf("%s/api/approval/action?token=%s&action=approve", baseURL, approveToken)
			rejectURL := fmt.Sprintf("%s/api/approval/action?token=%s&action=reject", baseURL, rejectToken)
			pdfURL := fmt.Sprintf("%s/api/po/pdf?holdingcode=%s&docno=%s", baseURL, req.HoldingCode, req.DocNo)

			emailErr := SendApprovalEmail(
				approver.Email,
				approver.UserName,
				req.HoldingCode,
				req.DocNo,
				totalAmount,
				poStatus.PurchaseTypeName,
				poStatus.CreatedByName,
				approveURL,
				rejectURL,
				pdfURL,
				false, // ไม่ใช่การแก้ไข (manual send)
			)

			status := "sent"
			errMsg := ""
			if emailErr != nil {
				status = "failed"
				errMsg = emailErr.Error()
			} else {
				emailSent++
			}

			// บันทึก log
			emailLog := NotificationLog{
				HoldingCode:      req.HoldingCode,
				DocNo:            req.DocNo,
				GuidFixed:        poStatus.GuidFixed,
				ApproverCode:     approver.UserCode,
				ApproverName:     approver.UserName,
				NotificationType: "email",
				RecipientEmail:   approver.Email,
				Status:           status,
				ErrorMessage:     errMsg,
				SentAt:           now,
				CreatedAt:        now,
			}
			notificationCollection.InsertOne(ctx, emailLog)

			results = append(results, map[string]any{
				"approver": approver.UserCode,
				"type":     "email",
				"status":   status,
				"to":       approver.Email,
				"error":    errMsg,
			})
		}

		// ส่ง LINE Push ถ้ามี
		if approver.LineUserID != "" && liffID != "" {
			// สร้าง approval token สำหรับ LIFF
			approveToken, _ := generateApprovalToken(req.HoldingCode, req.DocNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "approve")

			// สร้าง LIFF URL พร้อม token
			liffURL := fmt.Sprintf("https://liff.line.me/%s/approve?token=%s&holdingcode=%s&docno=%s", liffID, approveToken, req.HoldingCode, req.DocNo)

			lineErr := SendLinePushApprovalV2(ApprovalNotificationParams{
				LineUserID:       approver.LineUserID,
				HoldingCode:      req.HoldingCode,
				DocNo:            req.DocNo,
				DocDatetime:      poStatus.DocDatetime,
				TotalAmount:      totalAmount,
				PurchaseTypeName: poStatus.PurchaseTypeName,
				CreatedByName:    poStatus.CreatedByName,
				CustName:         poStatus.CustName,
				LiffApproveURL:   liffURL,
				IsModified:       false, // manual send - ไม่ใช่การแก้ไข
				POComment:        poStatus.LastComment,
			})

			status := "sent"
			errMsg := ""
			if lineErr != nil {
				status = "failed"
				errMsg = lineErr.Error()
			} else {
				lineSent++
			}

			// บันทึก log
			lineLog := NotificationLog{
				HoldingCode:      req.HoldingCode,
				DocNo:            req.DocNo,
				GuidFixed:        poStatus.GuidFixed,
				ApproverCode:     approver.UserCode,
				ApproverName:     approver.UserName,
				NotificationType: "line",
				RecipientLineID:  approver.LineUserID,
				Status:           status,
				ErrorMessage:     errMsg,
				SentAt:           now,
				CreatedAt:        now,
			}
			notificationCollection.InsertOne(ctx, lineLog)

			results = append(results, map[string]any{
				"approver": approver.UserCode,
				"type":     "line",
				"status":   status,
				"to":       approver.LineDisplayName,
				"error":    errMsg,
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":    true,
		"message":    "ส่งแจ้งเตือนเสร็จสิ้น",
		"email_sent": emailSent,
		"line_sent":  lineSent,
		"skipped":    skipped,
		"results":    results,
	})
}

// =====================================================
// Approve/Reject via Token (Email/LIFF Link)
// =====================================================

// ApprovalActionRequest request สำหรับอนุมัติผ่าน token
type ApprovalActionRequest struct {
	Token   string `json:"token" query:"token"`
	Action  string `json:"action" query:"action"` // approve, reject
	Comment string `json:"comment" query:"comment"`
}

// ApproveViaTokenHandler - อนุมัติ/ปฏิเสธ PO ผ่าน token จาก email หรือ LINE LIFF
func ApproveViaTokenHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req ApprovalActionRequest
	// รองรับทั้ง query params และ JSON body
	if err := c.Bind(&req); err != nil {
		// ลอง query params
		req.Token = c.QueryParam("token")
		req.Action = c.QueryParam("action")
		req.Comment = c.QueryParam("comment")
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "token is required",
		})
	}

	if req.Action != "approve" && req.Action != "reject" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "action must be 'approve' or 'reject'",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. ค้นหาและ mark token as used แบบ atomic (ป้องกันใช้ซ้ำพร้อมกัน)
	tokenCollection := getTokenCollection(ApprovalTokensCollection)
	now := time.Now().UTC()

	var tokenDoc ApprovalToken
	tokenFilter := bson.M{
		"token":     req.Token,
		"used":      false,
		"expiresat": bson.M{"$gt": now}, // ยังไม่หมดอายุ
	}
	tokenUpdate := bson.M{
		"$set": bson.M{
			"used":    true,
			"used_at": now,
		},
	}
	// FindOneAndUpdate atomic — ป้องกัน 2 คนใช้ token เดียวกัน
	err := tokenCollection.FindOneAndUpdate(ctx, tokenFilter, tokenUpdate).Decode(&tokenDoc)

	if err == mongo.ErrNoDocuments {
		// ตรวจสอบว่า token มีอยู่จริงหรือไม่ เพื่อแสดงข้อความที่ถูกต้อง
		var checkToken ApprovalToken
		checkErr := tokenCollection.FindOne(ctx, bson.M{"token": req.Token}).Decode(&checkToken)
		if checkErr == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]any{
				"success": false,
				"message": "Token ไม่ถูกต้อง",
			})
		}
		if checkToken.Used {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"message": "Token นี้ถูกใช้งานแล้ว",
			})
		}
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Token หมดอายุแล้ว",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// 2. กำหนด action
	newStatus := ""
	actionText := ""
	if req.Action == "approve" {
		newStatus = "approved"
		actionText = "อนุมัติ"
	} else {
		newStatus = "rejected"
		actionText = "ปฏิเสธ"
	}

	// 3. อัปเดตสถานะ PO แบบ atomic (ป้องกัน approve/reject พร้อมกัน)
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	newHistory := ApprovalHistory{
		Action:       req.Action,
		ActionBy:     tokenDoc.ApproverCode,
		ActionByName: tokenDoc.ApproverName,
		Level:        0, // จะถูกเติมจาก document ที่ match
		Comment:      req.Comment,
		ActionAt:     now,
		Source:       "email",
	}

	// ใช้ FindOneAndUpdate กับ filter status=pending → ป้องกัน race condition
	poFilter := bson.M{
		"holdingcode": tokenDoc.HoldingCode,
		"docno":       tokenDoc.DocNo,
		"status":      "pending",
	}
	poUpdate := bson.M{
		"$set": bson.M{
			"status":       newStatus,
			"last_comment": req.Comment,
			"updatedat":    now,
		},
		"$push": bson.M{
			"history": newHistory,
		},
	}
	// ถ้า approve → อัปเดต current_approved_level ด้วย
	if req.Action == "approve" {
		poUpdate["$set"].(bson.M)["current_approved_level"] = tokenDoc.ApproverCode
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedStatus POApprovalStatus
	err = statusCollection.FindOneAndUpdate(ctx, poFilter, poUpdate, opts).Decode(&updatedStatus)

	if err == mongo.ErrNoDocuments {
		// PO ไม่อยู่ในสถานะ pending — ตรวจสอบสถานะจริง
		var existingPO POApprovalStatus
		findErr := statusCollection.FindOne(ctx, bson.M{
			"holdingcode": tokenDoc.HoldingCode,
			"docno":       tokenDoc.DocNo,
		}).Decode(&existingPO)

		if findErr == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]any{
				"success": false,
				"message": "ไม่พบเอกสารนี้ในระบบ",
			})
		}

		statusText := map[string]string{
			"approved":      "อนุมัติแล้ว",
			"rejected":      "ถูกปฏิเสธแล้ว",
			"auto_approved": "อนุมัติอัตโนมัติแล้ว",
		}
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": fmt.Sprintf("เอกสารนี้%s", statusText[existingPO.Status]),
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to update PO status",
		})
	}

	// อัปเดต current_approved_level จาก required_level (ใน updatedStatus)
	if req.Action == "approve" {
		statusCollection.UpdateOne(ctx, bson.M{
			"holdingcode": tokenDoc.HoldingCode,
			"docno":       tokenDoc.DocNo,
		}, bson.M{
			"$set": bson.M{
				"current_approved_level": updatedStatus.RequiredLevel,
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":       true,
		"message":       fmt.Sprintf("%sเอกสาร %s สำเร็จ", actionText, tokenDoc.DocNo),
		"action":        req.Action,
		"docno":         tokenDoc.DocNo,
		"approver_name": tokenDoc.ApproverName,
	})
}

// GetApprovalTokenInfoHandler - ดึงข้อมูล token สำหรับแสดงใน LIFF
func GetApprovalTokenInfoHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "token is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ค้นหา token
	tokenCollection := getTokenCollection(ApprovalTokensCollection)
	var tokenDoc ApprovalToken
	err := tokenCollection.FindOne(ctx, bson.M{
		"token": token,
	}).Decode(&tokenDoc)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Token ไม่ถูกต้อง",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// ดึงข้อมูล PO
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	var poStatus POApprovalStatus
	err = statusCollection.FindOne(ctx, bson.M{
		"holdingcode": tokenDoc.HoldingCode,
		"docno":       tokenDoc.DocNo,
	}).Decode(&poStatus)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "ไม่พบข้อมูล PO",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"token_valid":        !tokenDoc.Used && time.Now().UTC().Before(tokenDoc.ExpiresAt),
			"token_used":         tokenDoc.Used,
			"token_expired":      time.Now().UTC().After(tokenDoc.ExpiresAt),
			"holdingcode":        tokenDoc.HoldingCode,
			"docno":              tokenDoc.DocNo,
			"totalamount":        poStatus.TotalAmount,
			"purchase_type_name": poStatus.PurchaseTypeName,
			"createdbyname":      poStatus.CreatedByName,
			"status":             poStatus.Status,
			"required_level":     poStatus.RequiredLevel,
			"approver_name":      tokenDoc.ApproverName,
		},
	})
}

// GetEmailStatusHandler - ตรวจสอบสถานะ Email configuration (Brevo)
func GetEmailStatusHandler(c echo.Context) error {
	// ตรวจสอบ Brevo configuration
	brevoConfigured := isBrevoConfigured()

	senderEmail := os.Getenv("BREVO_FROM_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@bcai.cloud"
	}
	senderName := os.Getenv("BREVO_FROM_NAME")
	if senderName == "" {
		senderName = "BC Ai Account"
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":     true,
		"configured":  brevoConfigured,
		"provider":    "brevo",
		"from_email":  senderEmail,
		"from_name":   senderName,
		"has_api_key": brevoConfigured,
	})
}

// GetSMTPStatusHandler - alias for backward compatibility
func GetSMTPStatusHandler(c echo.Context) error {
	return GetEmailStatusHandler(c)
}

// TestEmailRequest request สำหรับทดสอบส่ง email
type TestEmailRequest struct {
	ToEmail string `json:"toemail"`
	ToName  string `json:"toname"`
}

// SendTestEmailHandler - ส่ง email ทดสอบ
func SendTestEmailHandler(c echo.Context) error {
	var req TestEmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ToEmail == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "to_email is required",
		})
	}

	if req.ToName == "" {
		req.ToName = req.ToEmail
	}

	// ตรวจสอบ Brevo configuration
	if !isBrevoConfigured() {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Brevo API key not configured",
		})
	}

	// สร้าง HTML content สำหรับทดสอบ
	subject := "ทดสอบส่ง Email จาก BC Ai Account"
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Sarabun', sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; padding: 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; border-bottom: 2px solid #1976D2; padding-bottom: 20px; margin-bottom: 20px; }
        .header h1 { color: #1976D2; margin: 0; font-size: 24px; }
        .content { text-align: center; padding: 20px; }
        .success-icon { font-size: 64px; color: #4CAF50; }
        .footer { text-align: center; color: #999; font-size: 12px; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ทดสอบระบบส่ง Email</h1>
        </div>
        <div class="content">
            <div class="success-icon">✅</div>
            <h2>ส่ง Email สำเร็จ!</h2>
            <p>เรียน คุณ%s,</p>
            <p>นี่คือ email ทดสอบจากระบบ BC Ai Account</p>
            <p>ระบบส่ง Email ผ่าน Brevo API ทำงานปกติ</p>
        </div>
        <div class="footer">
            <p>Email นี้ส่งจากระบบ BC Ai Account</p>
            <p>เวลาที่ส่ง: %s</p>
        </div>
    </div>
</body>
</html>
`, req.ToName, time.Now().Format("2006-01-02 15:04:05"))

	// ส่ง email
	err := sendBrevoEmail(req.ToEmail, req.ToName, subject, htmlBody)
	if err != nil {
		logger.Error("[TestEmail] Failed to send test email to %s: %v", req.ToEmail, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": fmt.Sprintf("Failed to send email: %v", err),
		})
	}

	logger.Info("[TestEmail] Test email sent successfully to %s", req.ToEmail)
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("ส่ง email ทดสอบไปยัง %s สำเร็จ", req.ToEmail),
	})
}

// =====================================================
// LINE OA Configuration Diagnostic
// =====================================================

// GetLineOAConfigStatusHandler - ตรวจสอบ LINE OA config ของ shop
// ใช้สำหรับ debug ปัญหา LINE notification ไม่ส่ง
func GetLineOAConfigStatusHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	holdingCode := c.QueryParam("holdingcode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "holdingcode is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ค้นหา config จาก lineoaconfigs collection
	collection := atlasDB.Collection("lineoaconfigs")

	var config struct {
		HoldingCode   string `bson:"holdingcode"`
		ChannelID     string `bson:"channelid"`
		AccessToken   string `bson:"accesstoken"`
		LiffID        string `bson:"liffid"`
		IsActive      bool   `bson:"isactive"`
		ChannelSecret string `bson:"channelsecret"`
	}

	err := collection.FindOne(ctx, bson.M{
		"holdingcode": holdingCode,
		"isactive":    true,
	}).Decode(&config)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success":     true,
			"found":       false,
			"message":     "ไม่พบ LINE OA config สำหรับ shop นี้",
			"holdingcode": holdingCode,
			"diagnosis":   "ต้องสร้าง lineoaconfigs document ใน MongoDB",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query failed: " + err.Error(),
		})
	}

	// ตรวจสอบและซ่อน access_token (แสดงเฉพาะ 10 ตัวแรก)
	accessTokenMasked := ""
	if len(config.AccessToken) > 10 {
		accessTokenMasked = config.AccessToken[:10] + "..."
	} else if config.AccessToken != "" {
		accessTokenMasked = "***"
	}

	// วิเคราะห์ปัญหา
	issues := []string{}
	if config.AccessToken == "" {
		issues = append(issues, "access_token ว่าง - LINE Push จะไม่ทำงาน")
	}
	if config.LiffID == "" {
		issues = append(issues, "liff_id ว่าง - LINE Push จะไม่ทำงาน")
	}
	if !config.IsActive {
		issues = append(issues, "isactive = false - config ถูกปิดใช้งาน")
	}

	readyForLine := config.AccessToken != "" && config.LiffID != "" && config.IsActive

	return c.JSON(http.StatusOK, map[string]any{
		"success":              true,
		"found":                true,
		"holdingcode":          config.HoldingCode,
		"channel_id":           config.ChannelID,
		"has_access_token":     config.AccessToken != "",
		"access_token_preview": accessTokenMasked,
		"has_liff_id":          config.LiffID != "",
		"liff_id":              config.LiffID,
		"isactive":             config.IsActive,
		"has_channel_secret":   config.ChannelSecret != "",
		"ready_for_line_push":  readyForLine,
		"issues":               issues,
	})
}

// ListAllLineOAConfigsHandler - แสดงรายการ LINE OA configs ทั้งหมด
// ใช้สำหรับ debug หา holdingcode ที่มี config
func ListAllLineOAConfigsHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("lineoaconfigs")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query failed: " + err.Error(),
		})
	}
	defer cursor.Close(ctx)

	var configs []map[string]any
	for cursor.Next(ctx) {
		var doc struct {
			HoldingCode string `bson:"holdingcode"`
			ChannelID   string `bson:"channelid"`
			LiffID      string `bson:"liffid"`
			IsActive    bool   `bson:"isactive"`
			HasToken    bool
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// ตรวจสอบว่ามี access_token หรือไม่ (ไม่แสดงค่าจริง)
		var rawDoc bson.M
		cursor.Decode(&rawDoc)
		hasToken := false
		if token, ok := rawDoc["access_token"].(string); ok && token != "" {
			hasToken = true
		}

		configs = append(configs, map[string]any{
			"holdingcode":      doc.HoldingCode,
			"channel_id":       doc.ChannelID,
			"liff_id":          doc.LiffID,
			"isactive":         doc.IsActive,
			"has_access_token": hasToken,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"count":   len(configs),
		"configs": configs,
	})
}

// =====================================================
// Test LINE Push Message
// =====================================================

// TestLinePushRequest request สำหรับทดสอบส่ง LINE push
type TestLinePushRequest struct {
	LineUserID  string `json:"lineuserid"`
	UserName    string `json:"username"`
	HoldingCode string `json:"holdingcode"`
}

// TestLinePushHandler - ส่ง LINE push message ทดสอบ
func TestLinePushHandler(c echo.Context) error {
	var req TestLinePushRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.LineUserID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "lineuserid is required",
		})
	}

	if req.HoldingCode == "" {
		req.HoldingCode = "default"
	}

	// ดึง access token (hardcoded)
	accessToken, err := getLineAccessToken(req.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to get access token: " + err.Error(),
		})
	}

	// สร้าง test message
	testMessage := map[string]any{
		"to": req.LineUserID,
		"messages": []map[string]any{
			{
				"type":    "flex",
				"altText": "ทดสอบการแจ้งเตือน LINE",
				"contents": map[string]any{
					"type": "bubble",
					"size": "kilo",
					"header": map[string]any{
						"type":            "box",
						"layout":          "vertical",
						"backgroundColor": "#4CAF50",
						"paddingAll":      "15px",
						"contents": []map[string]any{
							{
								"type":   "text",
								"text":   "✅ ทดสอบสำเร็จ!",
								"color":  "#FFFFFF",
								"size":   "lg",
								"weight": "bold",
							},
						},
					},
					"body": map[string]any{
						"type":       "box",
						"layout":     "vertical",
						"paddingAll": "15px",
						"contents": []map[string]any{
							{
								"type":  "text",
								"text":  fmt.Sprintf("สวัสดี %s", req.UserName),
								"size":  "md",
								"wrap":  true,
								"color": "#333333",
							},
							{
								"type":   "text",
								"text":   "ระบบ LINE Push Message ทำงานปกติ",
								"size":   "sm",
								"wrap":   true,
								"color":  "#666666",
								"margin": "md",
							},
							{
								"type":   "text",
								"text":   fmt.Sprintf("เวลา: %s", time.Now().Format("2006-01-02 15:04:05")),
								"size":   "xs",
								"color":  "#999999",
								"margin": "lg",
							},
						},
					},
				},
			},
		},
	}

	// ส่ง request ไป LINE API
	jsonData, _ := json.Marshal(testMessage)
	httpReq, _ := http.NewRequest("POST", "https://api.line.me/v2/bot/message/push", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Error("[TestLinePush] Failed to send: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to send LINE push: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// อ่าน response
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	if resp.StatusCode != 200 {
		logger.Error("[TestLinePush] LINE API error: %d - %s", resp.StatusCode, bodyString)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success":     false,
			"message":     "LINE API error",
			"status_code": resp.StatusCode,
			"response":    bodyString,
		})
	}

	logger.Info("[TestLinePush] Successfully sent test message to %s (%s)", req.UserName, req.LineUserID)
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("ส่งข้อความทดสอบไปยัง %s สำเร็จ", req.UserName),
	})
}

// =====================================================
// LIFF PO Approval Handlers
// =====================================================

// PODetailItem รายการสินค้าใน PO
type PODetailItem struct {
	LineNumber int     `json:"linenumber"`
	ItemCode   string  `json:"itemcode"`
	ItemName   string  `json:"itemname"`
	Qty        float64 `json:"qty"`
	UnitName   string  `json:"unitname"`
	Price      float64 `json:"price"`
	SumAmount  float64 `json:"sumamount"`
}

// PODetailsResponse ข้อมูล PO สำหรับแสดงใน LIFF
type PODetailsResponse struct {
	DocNo            string         `json:"docno"`
	DocDatetime      string         `json:"docdatetime"`
	CustCode         string         `json:"custcode"`
	CustName         string         `json:"custname"`
	PurchaseTypeName string         `json:"purchasetypename"`
	TotalAmount      float64        `json:"totalamount"`
	CreatedByName    string         `json:"createdbyname"`
	Status           string         `json:"status"`
	Items            []PODetailItem `json:"items"`
}

// GetPODetailsForLIFFRequest request สำหรับดึงข้อมูล PO
type GetPODetailsForLIFFRequest struct {
	Token       string `json:"token"`
	HoldingCode string `json:"holdingcode"`
	DocNo       string `json:"docno"`
}

// GetPODetailsForLIFFHandler - ดึงข้อมูล PO พร้อมรายการสินค้าสำหรับ LIFF
func GetPODetailsForLIFFHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req GetPODetailsForLIFFRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "token is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. ตรวจสอบ token
	tokenCollection := getTokenCollection(ApprovalTokensCollection)
	var tokenDoc ApprovalToken
	err := tokenCollection.FindOne(ctx, bson.M{
		"token": req.Token,
	}).Decode(&tokenDoc)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Token ไม่ถูกต้อง",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// 2. ตรวจสอบว่า token ใช้แล้วหรือหมดอายุ
	if tokenDoc.Used {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Token นี้ถูกใช้งานแล้ว",
		})
	}
	if time.Now().UTC().After(tokenDoc.ExpiresAt) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Token หมดอายุแล้ว",
		})
	}

	// 3. ดึงข้อมูล PO จาก poapprovalstatus
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	var poStatus POApprovalStatus
	err = statusCollection.FindOne(ctx, bson.M{
		"holdingcode": tokenDoc.HoldingCode,
		"docno":       tokenDoc.DocNo,
	}).Decode(&poStatus)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "ไม่พบข้อมูล PO",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// 4. ดึงข้อมูล transaction header และ detail จาก transactiondb
	transactionDB := atlasClient.Database("transactiondb")

	// ดึง header เพื่อเอา custcode, custname, docdatetime
	headerCollection := transactionDB.Collection("transactionheader")
	var transDoc bson.M
	err = headerCollection.FindOne(ctx, bson.M{
		"holdingcode": tokenDoc.HoldingCode,
		"docno":       tokenDoc.DocNo,
	}).Decode(&transDoc)

	custCode := ""
	custName := ""
	docDatetime := poStatus.CreatedAt // ใช้ createdAt เป็น fallback

	if err == nil {
		custCode = getString(transDoc["custcode"])
		// ดึงชื่อลูกค้าจาก custnames
		if custNames, ok := transDoc["custnames"].(bson.A); ok && len(custNames) > 0 {
			if firstName, ok := custNames[0].(bson.M); ok {
				if name, ok := firstName["name"].(string); ok {
					custName = name
				}
			}
		}
		// ดึง docdatetime
		if dt, ok := transDoc["docdatetime"].(primitive.DateTime); ok {
			docDatetime = dt.Time()
		}
	}

	// ดึงรายการสินค้า
	items := []PODetailItem{}
	detailCollection := transactionDB.Collection("transactiondetail")

	cursor, err := detailCollection.Find(ctx, bson.M{
		"holdingcode": tokenDoc.HoldingCode,
		"docno":       tokenDoc.DocNo,
	}, options.Find().SetSort(bson.M{"linenumber": 1}))

	if err == nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				continue
			}

			// ดึงชื่อสินค้าจาก itemnames
			itemName := ""
			if itemNames, ok := doc["itemnames"].(bson.A); ok && len(itemNames) > 0 {
				if firstName, ok := itemNames[0].(bson.M); ok {
					if name, ok := firstName["name"].(string); ok {
						itemName = name
					}
				}
			}

			// ดึงชื่อหน่วยนับจาก unitnames
			unitName := ""
			if unitNames, ok := doc["unitnames"].(bson.A); ok && len(unitNames) > 0 {
				if firstName, ok := unitNames[0].(bson.M); ok {
					if name, ok := firstName["name"].(string); ok {
						unitName = name
					}
				}
			}

			item := PODetailItem{
				LineNumber: getInt(doc["linenumber"]),
				ItemCode:   getString(doc["itemcode"]),
				ItemName:   itemName,
				Qty:        getFloat(doc["qty"]),
				UnitName:   unitName,
				Price:      getFloat(doc["price"]),
				SumAmount:  getFloat(doc["sumamount"]),
			}
			items = append(items, item)
		}
	}

	// 5. สร้าง response
	response := PODetailsResponse{
		DocNo:            poStatus.DocNo,
		DocDatetime:      docDatetime.Format(time.RFC3339),
		CustCode:         custCode,
		CustName:         custName,
		PurchaseTypeName: poStatus.PurchaseTypeName,
		TotalAmount:      poStatus.TotalAmount,
		CreatedByName:    poStatus.CreatedByName,
		Status:           poStatus.Status,
		Items:            items,
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    response,
	})
}

// LiffApproveRequest request สำหรับอนุมัติผ่าน LIFF
type LiffApproveRequest struct {
	Token       string `json:"token"`
	Action      string `json:"action"` // approve, reject
	Pin         string `json:"pin"`
	Comment     string `json:"comment"`
	LineUserID  string `json:"lineuserid"`
	HoldingCode string `json:"holdingcode"`
	DocNo       string `json:"docno"`
}

// LiffApproveHandler - อนุมัติ/ปฏิเสธ PO จาก LIFF พร้อม PIN
func LiffApproveHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req LiffApproveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "token is required",
		})
	}

	if req.Action != "approve" && req.Action != "reject" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "action must be 'approve' or 'reject'",
		})
	}

	if len(req.Pin) < 4 {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "PIN must be at least 4 digits",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. ตรวจสอบ token
	tokenCollection := getTokenCollection(ApprovalTokensCollection)
	var tokenDoc ApprovalToken
	err := tokenCollection.FindOne(ctx, bson.M{
		"token": req.Token,
	}).Decode(&tokenDoc)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "Token ไม่ถูกต้อง",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// 2. ตรวจสอบว่า token ใช้แล้วหรือหมดอายุ
	if tokenDoc.Used {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Token นี้ถูกใช้งานแล้ว",
		})
	}
	if time.Now().UTC().After(tokenDoc.ExpiresAt) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Token หมดอายุแล้ว",
		})
	}

	// 3. ตรวจสอบ PIN ของผู้ใช้ (ดึงจาก users collection)
	usersDB := atlasClient.Database("bcai_documents")
	usersCollection := usersDB.Collection("users")

	var userDoc bson.M
	err = usersCollection.FindOne(ctx, bson.M{
		"holdingcode": tokenDoc.HoldingCode,
		"username":    tokenDoc.ApproverCode,
	}).Decode(&userDoc)

	if err == mongo.ErrNoDocuments {
		// ถ้าไม่พบ user ลองดูจาก lineUserId
		if req.LineUserID != "" {
			err = usersCollection.FindOne(ctx, bson.M{
				"holdingcode": tokenDoc.HoldingCode,
				"line_userid": req.LineUserID,
			}).Decode(&userDoc)
		}
	}

	if err == nil {
		// ตรวจสอบ PIN
		storedPin := getString(userDoc["pin"])
		if storedPin != "" && storedPin != req.Pin {
			logger.Warn("[LiffApprove] Invalid PIN for user %s", tokenDoc.ApproverCode)
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"message": "รหัส PIN ไม่ถูกต้อง",
			})
		}
	}
	// ถ้าไม่พบ user หรือไม่มี PIN ให้ผ่านไปก่อน (สำหรับ backward compatibility)

	// 4. ดึงสถานะ PO ปัจจุบัน
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	var poStatus POApprovalStatus
	err = statusCollection.FindOne(ctx, bson.M{
		"holdingcode": tokenDoc.HoldingCode,
		"docno":       tokenDoc.DocNo,
	}).Decode(&poStatus)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "ไม่พบเอกสารนี้ในระบบ",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// 5. ตรวจสอบว่า PO ยังรออนุมัติอยู่หรือไม่
	if poStatus.Status != "pending" {
		statusText := map[string]string{
			"approved":      "อนุมัติแล้ว",
			"rejected":      "ถูกปฏิเสธแล้ว",
			"auto_approved": "อนุมัติอัตโนมัติแล้ว",
		}
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": fmt.Sprintf("เอกสารนี้%s", statusText[poStatus.Status]),
		})
	}

	now := time.Now().UTC() // เก็บเวลาเป็น UTC+0

	// 6. ดำเนินการตาม action
	newStatus := ""
	actionText := ""
	if req.Action == "approve" {
		newStatus = "approved"
		actionText = "อนุมัติ"
	} else {
		newStatus = "rejected"
		actionText = "ปฏิเสธ"
	}

	// เพิ่มประวัติ (source = line เพราะเข้าผ่าน LIFF)
	newHistory := ApprovalHistory{
		Action:       req.Action,
		ActionBy:     tokenDoc.ApproverCode,
		ActionByName: tokenDoc.ApproverName,
		Level:        poStatus.RequiredLevel,
		Comment:      req.Comment,
		ActionAt:     now,
		Source:       "line", // อนุมัติผ่าน LINE LIFF
	}

	history := append(poStatus.History, newHistory)

	// อัปเดตสถานะ PO
	update := bson.M{
		"$set": bson.M{
			"status":                 newStatus,
			"current_approved_level": poStatus.RequiredLevel,
			"history":                history,
			"last_comment":           req.Comment,
			"updatedat":              now,
		},
	}

	_, err = statusCollection.UpdateOne(ctx, bson.M{
		"holdingcode": tokenDoc.HoldingCode,
		"docno":       tokenDoc.DocNo,
	}, update)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to update PO status",
		})
	}

	// 7. Mark token as used
	usedAt := time.Now().UTC() // เก็บเวลาเป็น UTC+0
	tokenCollection.UpdateOne(ctx, bson.M{"token": req.Token}, bson.M{
		"$set": bson.M{
			"used":    true,
			"used_at": &usedAt,
		},
	})

	logger.Info("[LiffApprove] %s PO %s by %s (via LIFF)", actionText, tokenDoc.DocNo, tokenDoc.ApproverName)

	return c.JSON(http.StatusOK, map[string]any{
		"success":       true,
		"message":       fmt.Sprintf("%sเอกสาร %s สำเร็จ", actionText, tokenDoc.DocNo),
		"action":        req.Action,
		"docno":         tokenDoc.DocNo,
		"approver_name": tokenDoc.ApproverName,
	})
}

// Helper functions for safe type conversion
func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

func getFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}

// ResendApprovalNotificationHandler - ส่งคำเตือนผู้อนุมัติซ้ำ (บังคับส่งใหม่แม้ส่งแล้ว)
// ใช้สำหรับกรณีที่ผู้สร้างเอกสารต้องการกดส่งคำเตือนอีกครั้ง
func ResendApprovalNotificationHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req SendRealNotificationsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.HoldingCode == "" || req.DocNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "holdingcode and docno are required",
		})
	}

	// Base URL fallback
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("APP_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "https://erp.bcai.cloud"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 1. ดึงสถานะการอนุมัติ PO
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	var poStatus POApprovalStatus
	err := statusCollection.FindOne(ctx, bson.M{
		"holdingcode": req.HoldingCode,
		"docno":       req.DocNo,
		"status":      "pending",
	}).Decode(&poStatus)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": "ไม่พบ PO ที่รออนุมัติ หรืออนุมัติ/ปฏิเสธไปแล้ว",
			"sent":    0,
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed: " + err.Error(),
		})
	}

	// 2. ดึงกฎการอนุมัติ
	settingsCollection := getCollection(POApprovalSettingsCollection)
	var setting POApprovalSetting
	err = settingsCollection.FindOne(ctx, bson.M{
		"holdingcode":      req.HoldingCode,
		"purchasetypecode": poStatus.PurchaseTypeCode,
	}).Decode(&setting)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": "ไม่พบกฎการอนุมัติสำหรับประเภทการจัดซื้อนี้",
			"sent":    0,
		})
	}

	// 3. หาผู้อนุมัติที่เกี่ยวข้อง
	var approvers []ApproverInfo
	for _, rule := range setting.Rules {
		if poStatus.TotalAmount >= rule.MinAmount {
			if rule.MaxAmount == 0 || poStatus.TotalAmount <= rule.MaxAmount {
				if rule.ApprovalLevel == poStatus.RequiredLevel && len(rule.Approvers) > 0 {
					approvers = rule.Approvers
					break
				}
			}
		}
	}

	if len(approvers) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": "ไม่พบรายชื่อผู้อนุมัติ",
			"sent":    0,
		})
	}

	// 4. ส่งคำเตือนซ้ำ (ไม่ตรวจสอบว่าส่งแล้วหรือยัง)
	notificationCollection := getCollection(POApprovalNotificationLogCollection)
	now := time.Now().UTC()
	totalAmount := fmt.Sprintf("%.2f", poStatus.TotalAmount)

	// ดึง LIFF ID สำหรับ LINE
	liffID, _ := getLiffID(req.HoldingCode)

	results := []map[string]any{}
	emailSent := 0
	lineSent := 0

	for _, approver := range approvers {
		// ส่ง Email ถ้ามี
		if approver.Email != "" {
			// สร้าง approval tokens
			approveToken, _ := generateApprovalToken(req.HoldingCode, req.DocNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "approve")
			rejectToken, _ := generateApprovalToken(req.HoldingCode, req.DocNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "reject")

			approveURL := fmt.Sprintf("%s/api/approval/action?token=%s&action=approve", baseURL, approveToken)
			rejectURL := fmt.Sprintf("%s/api/approval/action?token=%s&action=reject", baseURL, rejectToken)
			pdfURL := fmt.Sprintf("%s/api/po/pdf?holdingcode=%s&docno=%s", baseURL, req.HoldingCode, req.DocNo)

			// ใช้ subject ที่บอกว่าเป็น reminder
			emailErr := sendBrevoEmail(
				approver.Email,
				approver.UserName,
				fmt.Sprintf("[คำเตือน] รออนุมัติใบสั่งซื้อ: %s", req.DocNo),
				generateApprovalEmailHTML(req.DocNo, totalAmount, poStatus.PurchaseTypeName, poStatus.CreatedByName, approveURL, rejectURL, pdfURL, true),
			)

			if emailErr != nil {
				logger.Error("[Resend] Failed to send reminder email to %s: %v", approver.Email, emailErr)
				results = append(results, map[string]any{
					"approver": approver.UserCode,
					"type":     "email",
					"status":   "failed",
					"error":    emailErr.Error(),
				})
			} else {
				// บันทึก log การส่ง (type = reminder)
				_, _ = notificationCollection.InsertOne(ctx, bson.M{
					"holdingcode":       req.HoldingCode,
					"docno":             req.DocNo,
					"guidfixed":         req.GuidFixed,
					"approvercode":      approver.UserCode,
					"approver_name":     approver.UserName,
					"notification_type": "email_reminder",
					"recipient_email":   approver.Email,
					"status":            "sent",
					"sent_at":           now,
					"createdat":         now,
				})
				emailSent++
				results = append(results, map[string]any{
					"approver": approver.UserCode,
					"type":     "email",
					"status":   "sent",
				})
			}
		}

		// ส่ง LINE Push ถ้ามี
		if approver.LineUserID != "" && liffID != "" {
			// สร้าง approval token สำหรับ LIFF
			liffToken, _ := generateApprovalToken(req.HoldingCode, req.DocNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "liff")
			// URL ต้องมี /approve path และ holdingcode, docno parameters
			liffApproveURL := fmt.Sprintf("https://liff.line.me/%s/approve?token=%s&holdingcode=%s&docno=%s", liffID, liffToken, req.HoldingCode, req.DocNo)

			// ส่ง LINE Push พร้อมข้อความ reminder
			lineErr := SendLinePushApprovalV2(ApprovalNotificationParams{
				LineUserID:       approver.LineUserID,
				HoldingCode:      req.HoldingCode,
				DocNo:            req.DocNo,
				DocDatetime:      poStatus.DocDatetime,
				TotalAmount:      totalAmount,
				PurchaseTypeName: poStatus.PurchaseTypeName,
				CreatedByName:    poStatus.CreatedByName,
				CustName:         poStatus.CustName,
				LiffApproveURL:   liffApproveURL,
				IsModified:       true, // ใช้ IsModified=true เพื่อแสดงข้อความเตือน
				POComment:        poStatus.LastComment,
			})

			if lineErr != nil {
				logger.Error("[Resend] Failed to send LINE reminder to %s: %v", approver.LineUserID, lineErr)
				results = append(results, map[string]any{
					"approver": approver.UserCode,
					"type":     "line",
					"status":   "failed",
					"error":    lineErr.Error(),
				})
			} else {
				// บันทึก log การส่ง (type = reminder)
				_, _ = notificationCollection.InsertOne(ctx, bson.M{
					"holdingcode":       req.HoldingCode,
					"docno":             req.DocNo,
					"guidfixed":         req.GuidFixed,
					"approvercode":      approver.UserCode,
					"approver_name":     approver.UserName,
					"notification_type": "line_reminder",
					"recipient_line_id": approver.LineUserID,
					"status":            "sent",
					"sent_at":           now,
					"createdat":         now,
				})
				lineSent++
				results = append(results, map[string]any{
					"approver": approver.UserCode,
					"type":     "line",
					"status":   "sent",
				})
			}
		}
	}

	totalSent := emailSent + lineSent
	return c.JSON(http.StatusOK, map[string]any{
		"success":    true,
		"message":    fmt.Sprintf("ส่งคำเตือนสำเร็จ %d ครั้ง (Email: %d, LINE: %d)", totalSent, emailSent, lineSent),
		"sent":       totalSent,
		"email_sent": emailSent,
		"line_sent":  lineSent,
		"details":    results,
	})
}

// generateApprovalEmailHTML สร้าง HTML สำหรับ email อนุมัติ
func generateApprovalEmailHTML(docNo, totalAmount, purchaseTypeName, createdByName, approveURL, rejectURL, pdfURL string, isReminder bool) string {
	reminderBadge := ""
	if isReminder {
		reminderBadge = `<div style="background-color: #FF9800; color: white; padding: 10px; text-align: center; border-radius: 5px; margin-bottom: 15px; font-weight: bold;">⏰ นี่คือคำเตือน - กรุณาดำเนินการอนุมัติ</div>`
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>รออนุมัติใบสั่งซื้อ</title>
</head>
<body style="font-family: 'Sarabun', sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    %s
    <div style="background-color: #f5f5f5; padding: 20px; border-radius: 10px;">
        <h2 style="color: #1976D2; text-align: center;">📋 รออนุมัติใบสั่งซื้อ</h2>
        <p style="text-align: center; font-size: 24px; font-weight: bold; color: #333;">%s</p>
        <table style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
            <tr>
                <td style="padding: 10px; border-bottom: 1px solid #ddd;">ยอดรวม:</td>
                <td style="padding: 10px; border-bottom: 1px solid #ddd; text-align: right; font-weight: bold; color: #4CAF50;">฿%s</td>
            </tr>
            <tr>
                <td style="padding: 10px; border-bottom: 1px solid #ddd;">ประเภท:</td>
                <td style="padding: 10px; border-bottom: 1px solid #ddd; text-align: right;">%s</td>
            </tr>
            <tr>
                <td style="padding: 10px; border-bottom: 1px solid #ddd;">ผู้สร้าง:</td>
                <td style="padding: 10px; border-bottom: 1px solid #ddd; text-align: right;">%s</td>
            </tr>
        </table>
        <div style="text-align: center; margin-top: 30px;">
            <a href="%s" style="display: inline-block; padding: 15px 30px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 5px;">✅ อนุมัติ</a>
            <a href="%s" style="display: inline-block; padding: 15px 30px; background-color: #f44336; color: white; text-decoration: none; border-radius: 5px; margin: 5px;">❌ ไม่อนุมัติ</a>
        </div>
        <p style="text-align: center; margin-top: 20px; font-size: 12px; color: #666;">
            <a href="%s" style="color: #1976D2;">ดู PDF เอกสาร</a>
        </p>
    </div>
</body>
</html>
`, reminderBadge, docNo, totalAmount, purchaseTypeName, createdByName, approveURL, rejectURL, pdfURL)
}
