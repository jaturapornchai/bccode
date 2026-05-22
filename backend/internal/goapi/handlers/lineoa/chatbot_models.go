package lineoa

import "time"

// MongoDB Collections for Chatbot
const (
	ConversationCollection = "lineoa_conversations"
)

// ConversationMessage represents a single message in conversation
type ConversationMessage struct {
	Role string    `bson:"role" json:"role"`           // "user" or "assistant"
	Content string    `bson:"content" json:"content"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// ConversationSession stores chat history for context
type ConversationSession struct {
	SessionID string                `bson:"session_id" json:"session_id"`
	ShopID string                `bson:"shop_id" json:"shop_id"`
	LineUserID string                `bson:"line_user_id" json:"line_user_id"`
	DisplayName string                `bson:"display_name" json:"display_name"`
	Messages []ConversationMessage `bson:"messages" json:"messages"`
	Context map[string]string     `bson:"context" json:"context"` // Store state like current topic
	CreatedAt time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time             `bson:"updated_at" json:"updated_at"`
	ExpiresAt time.Time             `bson:"expires_at" json:"expires_at"`
}

// LINE Webhook Types
type LineWebhookRequest struct {
	Events []LineEvent `json:"events"`
}

type LineEvent struct {
	Type string        `json:"type"`
	ReplyToken string        `json:"reply_token"`
	Source LineSource    `json:"source"`
	Timestamp int64         `json:"timestamp"`
	Message *LineMessage  `json:"message,omitempty"`
	Postback *LinePostback `json:"postback,omitempty"`
}

type LineSource struct {
	Type string `json:"type"`
	UserID string `json:"user_id"`
	GroupID string `json:"group_id,omitempty"`
	RoomID string `json:"room_id,omitempty"`
}

type LineMessage struct {
	ID string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type LinePostback struct {
	Data string `json:"data"`
}

// LINE Reply Types
type LineReplyRequest struct {
	ReplyToken string        `json:"reply_token"`
	Messages []interface{} `json:"messages"`
}

type LineTextMessage struct {
	Type string          `json:"type"`
	Text string          `json:"text"`
	QuickReply *LineQuickReply `json:"quick_reply,omitempty"`
}

type LineQuickReply struct {
	Items []LineQuickReplyItem `json:"items"`
}

type LineQuickReplyItem struct {
	Type string          `json:"type"`
	Action LineQuickAction `json:"action"`
}

type LineQuickAction struct {
	Type string `json:"type"`  // "message", "postback", "uri"
	Label string `json:"label"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
	URI string `json:"uri,omitempty"` // For URI action (open LIFF/external URL)
}

// Flex Message Types
type LineFlexMessage struct {
	Type string      `json:"type"`
	AltText string      `json:"alt_text"`
	Contents interface{} `json:"contents"`
}

type FlexBubble struct {
	Type string          `json:"type"`
	Header *FlexBox        `json:"header,omitempty"`
	Body *FlexBox        `json:"body,omitempty"`
	Footer *FlexBox        `json:"footer,omitempty"`
	Styles *FlexBubbleStyle `json:"styles,omitempty"`
}

type FlexBox struct {
	Type string        `json:"type"`
	Layout string        `json:"layout"`
	Contents []interface{} `json:"contents"`
	Spacing string        `json:"spacing,omitempty"`
	Margin string        `json:"margin,omitempty"`
}

type FlexText struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Weight string `json:"weight,omitempty"`
	Size string `json:"size,omitempty"`
	Color string `json:"color,omitempty"`
	Wrap bool   `json:"wrap,omitempty"`
	Margin string `json:"margin,omitempty"`
}

type FlexButton struct {
	Type string          `json:"type"`
	Action LineQuickAction `json:"action"`
	Style string          `json:"style,omitempty"`
	Color string          `json:"color,omitempty"`
	Height string          `json:"height,omitempty"`
}

type FlexBubbleStyle struct {
	Header *FlexBlockStyle `json:"header,omitempty"`
	Body *FlexBlockStyle `json:"body,omitempty"`
	Footer *FlexBlockStyle `json:"footer,omitempty"`
}

type FlexBlockStyle struct {
	BackgroundColor string `json:"background_color,omitempty"`
}

// ChatbotConfig stores AI chatbot settings per shop
type ChatbotConfig struct {
	ShopID string    `bson:"shop_id" json:"shop_id"`
	LineOAConfigGUID string    `bson:"lineoa_config_guid" json:"lineoa_config_guid"`
	IsEnabled bool      `bson:"is_enabled" json:"is_enabled"`
	WelcomeMessage string    `bson:"welcome_message" json:"welcome_message"`
	SystemPrompt string    `bson:"system_prompt" json:"system_prompt"`
	MenuItems []MenuItem `bson:"menu_items" json:"menu_items"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type MenuItem struct {
	Label string `bson:"label" json:"label"`
	Text string `bson:"text" json:"text"`
	Description string `bson:"description" json:"description"`
}
