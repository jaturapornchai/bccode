package lineoa

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MaxConversationMessages = 20 // Keep last 20 messages for context
	SessionExpiry           = 30 * time.Minute
)

// WebhookHandler handles LINE webhook events
// POST /api/lineoa/webhook
func WebhookHandler(c echo.Context) error {
	// Get holdingcode from query parameter or header
	holdingCode := c.QueryParam("holdingcode")
	if holdingCode == "" {
		holdingCode = c.Request().Header.Get("X-Shop-ID")
	}

	if holdingCode == "" {
		logger.Warn("[LINE Webhook] Missing holdingcode")
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// Parse webhook body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Error("[LINE Webhook] Failed to read body: %v", err)
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// SECURITY (2026-06-21): verify the LINE webhook signature against the tenant's
	// ChannelSecret BEFORE doing any work. This is an unauthenticated public endpoint, so
	// without HMAC verification anyone could spoof events, inject an arbitrary holdingcode,
	// and amplify goroutine/Mongo/AI cost. Only a request actually signed by LINE with this
	// tenant's secret is processed.
	channelSecret, secErr := getShopChannelSecret(holdingCode)
	if secErr != nil || channelSecret == "" {
		logger.Warn("[LINE Webhook] no channel secret for holding=%s — rejecting", holdingCode)
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSig), []byte(c.Request().Header.Get("X-Line-Signature"))) {
		logger.Warn("[LINE Webhook] signature mismatch for holding=%s — rejecting", holdingCode)
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	var webhookReq LineWebhookRequest
	if err := json.Unmarshal(body, &webhookReq); err != nil {
		logger.Error("[LINE Webhook] Failed to parse body: %v", err)
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// Process events asynchronously
	go processWebhookEvents(holdingCode, webhookReq.Events)

	// Always return 200 OK immediately (LINE requires fast response)
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// processWebhookEvents processes LINE events
func processWebhookEvents(holdingCode string, events []LineEvent) {
	for _, event := range events {
		switch event.Type {
		case "message":
			if event.Message != nil && event.Message.Type == "text" {
				handleTextMessage(holdingCode, event)
			}
		case "postback":
			if event.Postback != nil {
				handlePostback(holdingCode, event)
			}
		case "follow":
			handleFollow(holdingCode, event)
		}
	}
}

// handleTextMessage processes text messages with AI
func handleTextMessage(holdingCode string, event LineEvent) {
	userID := event.Source.UserID
	userMessage := event.Message.Text
	replyToken := event.ReplyToken

	logger.Info("[LINE Chat] Shop: %s, User: %s, Message: %s", holdingCode, userID, userMessage)

	// Get or create conversation session
	session, err := getOrCreateSession(holdingCode, userID)
	if err != nil {
		logger.Error("[LINE Chat] Failed to get session: %v", err)
		sendTextReply(holdingCode, replyToken, "ขออภัยครับ เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง", nil)
		return
	}

	// Check for special commands
	lowerMsg := strings.ToLower(strings.TrimSpace(userMessage))
	if lowerMsg == "เริ่มใหม่" || lowerMsg == "reset" || lowerMsg == "clear" {
		clearSession(holdingCode, userID)
		quickReplies := getMainMenuQuickReplies(holdingCode)
		sendTextReply(holdingCode, replyToken, "เริ่มการสนทนาใหม่แล้วครับ\n\nสามารถถามคำถามได้เลยครับ หรือเลือกเมนูด้านล่าง", quickReplies)
		return
	}

	// Handle "เชื่อมต่อบัญชี" command
	if strings.Contains(lowerMsg, "เชื่อมต่อ") || strings.Contains(lowerMsg, "link") {
		handleLinkAccountCommand(holdingCode, event)
		return
	}

	// Add user message to session
	session.Messages = append(session.Messages, ConversationMessage{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now(),
	})

	// Generate AI response
	aiResponse, suggestedReplies := generateAIResponse(holdingCode, session)

	// Add AI response to session
	session.Messages = append(session.Messages, ConversationMessage{
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now(),
	})

	// Save session
	saveSession(session)

	// Build quick replies
	var quickReplies *LineQuickReply
	if len(suggestedReplies) > 0 {
		quickReplies = buildQuickReplies(suggestedReplies)
	}

	// Send reply
	sendTextReply(holdingCode, replyToken, aiResponse, quickReplies)
}

// handlePostback processes postback events (button clicks)
func handlePostback(holdingCode string, event LineEvent) {
	userID := event.Source.UserID
	data := event.Postback.Data
	replyToken := event.ReplyToken

	logger.Info("[LINE Postback] Shop: %s, User: %s, Data: %s", holdingCode, userID, data)

	// Parse postback data
	parts := strings.Split(data, "=")
	if len(parts) < 2 {
		return
	}

	action := parts[0]
	value := parts[1]

	var response string
	var quickReplies *LineQuickReply

	switch action {
	case "menu":
		switch value {
		case "product":
			response = "คุณเลือกเมนูสินค้า\n\nสามารถถามได้เช่น:\n• สินค้า A เหลือกี่ชิ้น\n• แสดงสินค้าราคาถูกที่สุด 5 รายการ\n• ค้นหาสินค้า MAKITA"
			quickReplies = buildQuickReplies([]string{"สินค้าทั้งหมดกี่รายการ", "สินค้าใกล้หมด", "กลับเมนูหลัก"})
		case "sales":
			response = "คุณเลือกเมนูการขาย\n\nสามารถถามได้เช่น:\n• ยอดขายวันนี้เท่าไหร่\n• สินค้าขายดีที่สุด 10 อันดับ\n• รายงานขายสัปดาห์นี้"
			quickReplies = buildQuickReplies([]string{"ยอดขายวันนี้", "สินค้าขายดี", "กลับเมนูหลัก"})
		case "report":
			response = "คุณเลือกเมนูรายงาน\n\nสามารถถามได้เช่น:\n• สรุปยอดขายเดือนนี้\n• กำไรสุทธิวันนี้\n• รายงานสินค้าคงคลัง"
			quickReplies = buildQuickReplies([]string{"รายงานประจำวัน", "รายงานสต็อก", "กลับเมนูหลัก"})
		case "main":
			response = "กลับมาที่เมนูหลักแล้วครับ\n\nเลือกหัวข้อที่ต้องการ หรือพิมพ์คำถามได้เลยครับ"
			quickReplies = getMainMenuQuickReplies(holdingCode)
		}
	}

	if response != "" {
		sendTextReply(holdingCode, replyToken, response, quickReplies)
	}
}

// handleFollow handles follow events (new friend)
func handleFollow(holdingCode string, event LineEvent) {
	replyToken := event.ReplyToken

	welcomeMessage := "สวัสดีครับ! ยินดีต้อนรับสู่ BC Ai Account 🎉\n\nผมคือ AI ผู้ช่วยของคุณ สามารถช่วยเรื่อง:\n• 📦 ดูข้อมูลสินค้าและสต็อก\n• 💰 ตรวจสอบยอดขาย\n• 📊 ดูรายงานต่างๆ\n\nลองเลือกเมนูด้านล่าง หรือพิมพ์คำถามได้เลยครับ"

	quickReplies := getMainMenuQuickReplies(holdingCode)
	sendTextReply(holdingCode, replyToken, welcomeMessage, quickReplies)
}

// generateAIResponse generates response using AI provider
func generateAIResponse(holdingCode string, session *ConversationSession) (string, []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ai := aiprovider.GetProvider()

	// Build system prompt
	systemPrompt := buildSystemPrompt(holdingCode)

	// Build conversation history as user prompt
	var historyBuilder strings.Builder
	for _, msg := range session.Messages {
		if msg.Role == "user" {
			historyBuilder.WriteString(fmt.Sprintf("ผู้ใช้: %s\n", msg.Content))
		} else {
			historyBuilder.WriteString(fmt.Sprintf("ผู้ช่วย: %s\n", msg.Content))
		}
	}

	resp, err := ai.GenerateContent(ctx, aiprovider.ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   historyBuilder.String(),
		Temperature:  0.7,
		TopP:         0.95,
		MaxTokens:    1024,
	})
	if err != nil {
		logger.Error("[AI] Failed to generate response: %v", err)
		return "ขออภัยครับ ไม่สามารถประมวลผลได้ในขณะนี้ กรุณาลองใหม่อีกครั้ง", nil
	}

	aiText := resp.Text
	if aiText == "" {
		return "ขออภัยครับ ไม่สามารถสร้างคำตอบได้", nil
	}

	// Extract suggested questions from response if AI included them
	suggestedReplies := extractSuggestedReplies(aiText)

	// Clean AI response (remove suggested questions section if present)
	cleanResponse := cleanAIResponse(aiText)

	logger.Info("[AI] Response generated (tokens: %d)", resp.TotalTokens)

	return cleanResponse, suggestedReplies
}

// buildSystemPrompt creates the system prompt for AI
func buildSystemPrompt(holdingCode string) string {
	return fmt.Sprintf(`คุณเป็น AI ผู้ช่วยร้านค้าของระบบ BC Ai Account

**ข้อมูลร้าน:**
- Holding Code: %s

**หน้าที่ของคุณ:**
1. ตอบคำถามเกี่ยวกับสินค้า สต็อก ยอดขาย และรายงานต่างๆ
2. ให้ข้อมูลที่เป็นประโยชน์และตรงประเด็น
3. ถ้าไม่แน่ใจ ให้ถามกลับเพื่อความชัดเจน
4. จำบริบทการสนทนา - ถ้าผู้ใช้ถามต่อเนื่อง ให้อ้างอิงสิ่งที่คุยก่อนหน้า

**กฎสำคัญ:**
- ตอบเป็นภาษาไทยเสมอ
- ตอบกระชับ ไม่เกิน 200 ตัวอักษร
- ใช้ emoji เหมาะสม 1-2 ตัวต่อข้อความ
- ถ้าต้องการข้อมูลจาก database ให้บอกว่า "กำลังค้นหาข้อมูล..."
- สุภาพ เป็นมิตร

**ตัวอย่างการตอบ:**
ผู้ใช้: "สินค้า A เหลือกี่ชิ้น"
AI: "📦 สินค้า A คงเหลือ 50 ชิ้นครับ"

ผู้ใช้: "ยอดขายวันนี้"
AI: "💰 ยอดขายวันนี้ ฿15,000 ครับ (12 บิล)"

**คำแนะนำ:**
ท้ายคำตอบ ให้แนะนำคำถามที่เกี่ยวข้อง 2-3 ข้อ ในรูปแบบ:
[แนะนำ: คำถาม1 | คำถาม2 | คำถาม3]`, holdingCode)
}

// extractSuggestedReplies extracts suggested questions from AI response
func extractSuggestedReplies(text string) []string {
	var suggestions []string

	// Look for [แนะนำ: ...] pattern
	startIdx := strings.Index(text, "[แนะนำ:")
	if startIdx == -1 {
		// Default suggestions
		return []string{"สินค้าทั้งหมด", "ยอดขายวันนี้", "เมนูหลัก"}
	}

	endIdx := strings.Index(text[startIdx:], "]")
	if endIdx == -1 {
		return []string{"สินค้าทั้งหมด", "ยอดขายวันนี้", "เมนูหลัก"}
	}

	suggestPart := text[startIdx+len("[แนะนำ:") : startIdx+endIdx]
	parts := strings.Split(suggestPart, "|")

	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" && len(suggestions) < 4 {
			suggestions = append(suggestions, s)
		}
	}

	if len(suggestions) == 0 {
		return []string{"สินค้าทั้งหมด", "ยอดขายวันนี้", "เมนูหลัก"}
	}

	return suggestions
}

// cleanAIResponse removes suggested questions section from response
func cleanAIResponse(text string) string {
	startIdx := strings.Index(text, "[แนะนำ:")
	if startIdx == -1 {
		return strings.TrimSpace(text)
	}

	endIdx := strings.Index(text[startIdx:], "]")
	if endIdx == -1 {
		return strings.TrimSpace(text[:startIdx])
	}

	return strings.TrimSpace(text[:startIdx])
}

// getOrCreateSession gets existing session or creates new one
func getOrCreateSession(holdingCode, userID string) (*ConversationSession, error) {
	if !IsConnected() {
		return nil, fmt.Errorf("MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getCollection(ConversationCollection)
	sessionID := fmt.Sprintf("%s_%s", holdingCode, userID)

	var session ConversationSession
	err := collection.FindOne(ctx, bson.M{
		"sessionid": sessionID,
		"expiresat": bson.M{"$gt": time.Now()},
	}).Decode(&session)

	if err != nil {
		// Create new session
		session = ConversationSession{
			SessionID:   sessionID,
			HoldingCode: holdingCode,
			LineUserID:  userID,
			Messages:    []ConversationMessage{},
			Context:     make(map[string]string),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExpiresAt:   time.Now().Add(SessionExpiry),
		}
	}

	// Trim old messages if too many
	if len(session.Messages) > MaxConversationMessages {
		session.Messages = session.Messages[len(session.Messages)-MaxConversationMessages:]
	}

	return &session, nil
}

// saveSession saves conversation session to MongoDB
func saveSession(session *ConversationSession) error {
	if !IsConnected() {
		return fmt.Errorf("MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getCollection(ConversationCollection)

	session.UpdatedAt = time.Now()
	session.ExpiresAt = time.Now().Add(SessionExpiry)

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx,
		bson.M{"sessionid": session.SessionID},
		bson.M{"$set": session},
		opts,
	)

	return err
}

// clearSession clears conversation history
func clearSession(holdingCode, userID string) error {
	if !IsConnected() {
		return fmt.Errorf("MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getCollection(ConversationCollection)
	sessionID := fmt.Sprintf("%s_%s", holdingCode, userID)

	_, err := collection.DeleteOne(ctx, bson.M{"sessionid": sessionID})
	return err
}

// getMainMenuQuickReplies returns main menu quick reply buttons
func getMainMenuQuickReplies(holdingCode string) *LineQuickReply {
	items := []LineQuickReplyItem{
		{
			Type: "action",
			Action: LineQuickAction{
				Type:  "postback",
				Label: "📦 สินค้า",
				Data:  "menu=product",
			},
		},
		{
			Type: "action",
			Action: LineQuickAction{
				Type:  "postback",
				Label: "💰 การขาย",
				Data:  "menu=sales",
			},
		},
		{
			Type: "action",
			Action: LineQuickAction{
				Type:  "postback",
				Label: "📊 รายงาน",
				Data:  "menu=report",
			},
		},
	}

	// Add LIFF link button if configured
	liffID, err := getShopLiffID(holdingCode)
	if err == nil && liffID != "" {
		liffURL := fmt.Sprintf("https://liff.line.me/%s?holdingcode=%s", liffID, holdingCode)
		items = append(items, LineQuickReplyItem{
			Type: "action",
			Action: LineQuickAction{
				Type:  "uri",
				Label: "🔗 เชื่อมต่อบัญชี",
				URI:   liffURL,
			},
		})
	}

	return &LineQuickReply{Items: items}
}

// buildQuickReplies builds quick reply buttons from suggestions
func buildQuickReplies(suggestions []string) *LineQuickReply {
	if len(suggestions) == 0 {
		return nil
	}

	items := []LineQuickReplyItem{}
	for _, s := range suggestions {
		if len(items) >= 13 { // LINE limit
			break
		}

		label := s
		if len(label) > 20 { // LINE label limit
			label = label[:17] + "..."
		}

		items = append(items, LineQuickReplyItem{
			Type: "action",
			Action: LineQuickAction{
				Type:  "message",
				Label: label,
				Text:  s,
			},
		})
	}

	return &LineQuickReply{Items: items}
}

// sendTextReply sends text reply with optional quick replies
func sendTextReply(holdingCode, replyToken, text string, quickReply *LineQuickReply) error {
	// Get access token for this shop
	accessToken, err := getShopAccessToken(holdingCode)
	if err != nil {
		logger.Error("[LINE Reply] Failed to get access token: %v", err)
		return err
	}

	message := LineTextMessage{
		Type:       "text",
		Text:       text,
		QuickReply: quickReply,
	}

	replyReq := LineReplyRequest{
		ReplyToken: replyToken,
		Messages:   []interface{}{message},
	}

	return sendLineReply(accessToken, replyReq)
}

// getShopAccessToken gets LINE access token for shop
func getShopAccessToken(holdingCode string) (string, error) {
	if !IsConnected() {
		return "", fmt.Errorf("MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getCollection(ConfigCollection)

	var config ConfigDoc
	err := collection.FindOne(ctx, bson.M{
		"holdingcode": holdingCode,
		"isactive":    true,
	}).Decode(&config)

	if err != nil {
		return "", fmt.Errorf("config not found: %v", err)
	}

	if config.AccessToken == "" {
		return "", fmt.Errorf("access token not configured")
	}

	return config.AccessToken, nil
}

// getShopChannelSecret returns the LINE ChannelSecret for a holding, used to verify the
// inbound webhook X-Line-Signature. SECURITY (2026-06-21).
func getShopChannelSecret(holdingCode string) (string, error) {
	if !IsConnected() {
		return "", fmt.Errorf("MongoDB not connected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := getCollection(ConfigCollection)
	var config ConfigDoc
	err := collection.FindOne(ctx, bson.M{
		"holdingcode": holdingCode,
		"isactive":    true,
	}).Decode(&config)
	if err != nil {
		return "", fmt.Errorf("config not found: %v", err)
	}
	return config.ChannelSecret, nil
}

// sendLineReply sends reply to LINE API
func sendLineReply(accessToken string, req LineReplyRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/reply", bytes.NewBuffer(jsonData))
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
		logger.Error("[LINE Reply] API error: %s", string(body))
		return fmt.Errorf("LINE API returned %d", resp.StatusCode)
	}

	logger.Info("[LINE Reply] Message sent successfully")
	return nil
}

// ========== LIFF Button Support ==========

// buildLiffQuickReplyItem creates a quick reply button that opens LIFF
func buildLiffQuickReplyItem(label, liffURL string) LineQuickReplyItem {
	return LineQuickReplyItem{
		Type: "action",
		Action: LineQuickAction{
			Type:  "uri",
			Label: label,
			URI:   liffURL,
		},
	}
}

// getLiffURL generates LIFF URL with parameters
func getLiffURL(holdingCode, liffID string, params map[string]string) string {
	url := fmt.Sprintf("https://liff.line.me/%s?holdingcode=%s", liffID, holdingCode)
	for key, value := range params {
		url += fmt.Sprintf("&%s=%s", key, value)
	}
	return url
}

// sendFlexMessageWithLiffButton sends a Flex message with LIFF button
func sendFlexMessageWithLiffButton(holdingCode, replyToken, title, description, buttonLabel, liffURL string) error {
	accessToken, err := getShopAccessToken(holdingCode)
	if err != nil {
		logger.Error("[LINE Reply] Failed to get access token: %v", err)
		return err
	}

	// Build Flex Bubble with LIFF button
	flexBubble := FlexBubble{
		Type: "bubble",
		Body: &FlexBox{
			Type:   "box",
			Layout: "vertical",
			Contents: []interface{}{
				FlexText{
					Type:   "text",
					Text:   title,
					Weight: "bold",
					Size:   "lg",
					Color:  "#1DB446",
				},
				FlexText{
					Type:   "text",
					Text:   description,
					Size:   "sm",
					Color:  "#666666",
					Wrap:   true,
					Margin: "md",
				},
			},
		},
		Footer: &FlexBox{
			Type:   "box",
			Layout: "vertical",
			Contents: []interface{}{
				FlexButton{
					Type: "button",
					Action: LineQuickAction{
						Type:  "uri",
						Label: buttonLabel,
						URI:   liffURL,
					},
					Style:  "primary",
					Color:  "#1DB446",
					Height: "md",
				},
			},
		},
	}

	flexMessage := LineFlexMessage{
		Type:     "flex",
		AltText:  title,
		Contents: flexBubble,
	}

	replyReq := LineReplyRequest{
		ReplyToken: replyToken,
		Messages:   []interface{}{flexMessage},
	}

	return sendLineReply(accessToken, replyReq)
}

// sendTextWithLiffButton sends text message with LIFF quick reply button
func sendTextWithLiffButton(holdingCode, replyToken, text, buttonLabel, liffURL string, additionalReplies []LineQuickReplyItem) error {
	accessToken, err := getShopAccessToken(holdingCode)
	if err != nil {
		logger.Error("[LINE Reply] Failed to get access token: %v", err)
		return err
	}

	// Build quick replies with LIFF button
	items := []LineQuickReplyItem{
		buildLiffQuickReplyItem(buttonLabel, liffURL),
	}
	items = append(items, additionalReplies...)

	message := LineTextMessage{
		Type: "text",
		Text: text,
		QuickReply: &LineQuickReply{
			Items: items,
		},
	}

	replyReq := LineReplyRequest{
		ReplyToken: replyToken,
		Messages:   []interface{}{message},
	}

	return sendLineReply(accessToken, replyReq)
}

// Example: Handle "เชื่อมต่อบัญชี" command to open LIFF
func handleLinkAccountCommand(holdingCode string, event LineEvent) {
	replyToken := event.ReplyToken
	userID := event.Source.UserID

	// Get LIFF ID from shop config
	liffID, err := getShopLiffID(holdingCode)
	if err != nil || liffID == "" {
		sendTextReply(holdingCode, replyToken, "ขออภัยครับ ยังไม่ได้ตั้งค่า LIFF สำหรับร้านนี้", nil)
		return
	}

	// Generate link token
	linkToken := generateLinkToken(holdingCode, userID)
	if linkToken == "" {
		sendTextReply(holdingCode, replyToken, "ขออภัยครับ ไม่สามารถสร้างลิงก์ได้", nil)
		return
	}

	// Build LIFF URL
	liffURL := fmt.Sprintf("https://liff.line.me/%s?token=%s&holdingcode=%s", liffID, linkToken, holdingCode)

	// Send Flex message with LIFF button
	sendFlexMessageWithLiffButton(
		holdingCode,
		replyToken,
		"🔗 เชื่อมต่อบัญชี",
		"กดปุ่มด้านล่างเพื่อเชื่อมต่อบัญชี LINE กับระบบ",
		"เชื่อมต่อเลย",
		liffURL,
	)
}

// getShopLiffID gets LIFF ID for shop
func getShopLiffID(holdingCode string) (string, error) {
	if !IsConnected() {
		return "", fmt.Errorf("MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getCollection(ConfigCollection)

	var config ConfigDoc
	err := collection.FindOne(ctx, bson.M{
		"holdingcode": holdingCode,
		"isactive":    true,
	}).Decode(&config)

	if err != nil {
		return "", err
	}

	return config.LiffID, nil
}

// generateLinkToken generates a temporary link token
func generateLinkToken(holdingCode, userID string) string {
	if !IsConnected() {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := getCollection(LinkTokenCollection)

	token := fmt.Sprintf("%d", time.Now().UnixNano())
	tokenDoc := LinkTokenDoc{
		Token:       token,
		HoldingCode: holdingCode,
		Username:    userID, // Use LINE user ID as identifier
		TokenType:   "chatbot_link",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	_, err := collection.InsertOne(ctx, tokenDoc)
	if err != nil {
		logger.Error("[Chatbot] Failed to create link token: %v", err)
		return ""
	}

	return token
}
