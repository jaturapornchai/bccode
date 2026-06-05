package lineoa

import "time"

// MongoDB Collections for Chatbot
const (
	ConversationCollection = "lineoaconversations"
)

// ConversationMessage represents a single message in conversation
type ConversationMessage struct {
	Role      string    `bson:"role" json:"role"` // "user" or "assistant"
	Content   string    `bson:"content" json:"content"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// ConversationSession stores chat history for context
type ConversationSession struct {
	SessionID   string                `bson:"sessionid" json:"sessionid"`
	HoldingCode string                `bson:"holdingcode" json:"holdingcode"`
	LineUserID  string                `bson:"lineuserid" json:"lineuserid"`
	DisplayName string                `bson:"displayname" json:"displayname"`
	Messages    []ConversationMessage `bson:"messages" json:"messages"`
	Context     map[string]string     `bson:"context" json:"context"` // Store state like current topic
	CreatedAt   time.Time             `bson:"createdat" json:"createdat"`
	UpdatedAt   time.Time             `bson:"updatedat" json:"updatedat"`
	ExpiresAt   time.Time             `bson:"expiresat" json:"expiresat"`
}

// LINE Webhook Types
type LineWebhookRequest struct {
	Events []LineEvent `json:"events"`
}

type LineEvent struct {
	Type       string        `json:"type"`
	ReplyToken string        `json:"replytoken"`
	Source     LineSource    `json:"source"`
	Timestamp  int64         `json:"timestamp"`
	Message    *LineMessage  `json:"message,omitempty"`
	Postback   *LinePostback `json:"postback,omitempty"`
}

type LineSource struct {
	Type    string `json:"type"`
	UserID  string `json:"userid"`
	GroupID string `json:"groupid,omitempty"`
	RoomID  string `json:"roomid,omitempty"`
}

type LineMessage struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type LinePostback struct {
	Data string `json:"data"`
}

// LINE Reply Types
type LineReplyRequest struct {
	ReplyToken string        `json:"replytoken"`
	Messages   []interface{} `json:"messages"`
}

type LineTextMessage struct {
	Type       string          `json:"type"`
	Text       string          `json:"text"`
	QuickReply *LineQuickReply `json:"quickreply,omitempty"`
}

type LineQuickReply struct {
	Items []LineQuickReplyItem `json:"items"`
}

type LineQuickReplyItem struct {
	Type   string          `json:"type"`
	Action LineQuickAction `json:"action"`
}

type LineQuickAction struct {
	Type  string `json:"type"` // "message", "postback", "uri"
	Label string `json:"label"`
	Text  string `json:"text,omitempty"`
	Data  string `json:"data,omitempty"`
	URI   string `json:"uri,omitempty"` // For URI action (open LIFF/external URL)
}

// Flex Message Types
type LineFlexMessage struct {
	Type     string      `json:"type"`
	AltText  string      `json:"alttext"`
	Contents interface{} `json:"contents"`
}

type FlexBubble struct {
	Type   string           `json:"type"`
	Header *FlexBox         `json:"header,omitempty"`
	Body   *FlexBox         `json:"body,omitempty"`
	Footer *FlexBox         `json:"footer,omitempty"`
	Styles *FlexBubbleStyle `json:"styles,omitempty"`
}

type FlexBox struct {
	Type     string        `json:"type"`
	Layout   string        `json:"layout"`
	Contents []interface{} `json:"contents"`
	Spacing  string        `json:"spacing,omitempty"`
	Margin   string        `json:"margin,omitempty"`
}

type FlexText struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Weight string `json:"weight,omitempty"`
	Size   string `json:"size,omitempty"`
	Color  string `json:"color,omitempty"`
	Wrap   bool   `json:"wrap,omitempty"`
	Margin string `json:"margin,omitempty"`
}

type FlexButton struct {
	Type   string          `json:"type"`
	Action LineQuickAction `json:"action"`
	Style  string          `json:"style,omitempty"`
	Color  string          `json:"color,omitempty"`
	Height string          `json:"height,omitempty"`
}

type FlexBubbleStyle struct {
	Header *FlexBlockStyle `json:"header,omitempty"`
	Body   *FlexBlockStyle `json:"body,omitempty"`
	Footer *FlexBlockStyle `json:"footer,omitempty"`
}

type FlexBlockStyle struct {
	BackgroundColor string `json:"backgroundcolor,omitempty"`
}

// ChatbotConfig stores AI chatbot settings per shop
type ChatbotConfig struct {
	HoldingCode      string     `bson:"holdingcode" json:"holdingcode"`
	LineOAConfigGUID string     `bson:"lineoaconfigguid" json:"lineoaconfigguid"`
	IsEnabled        bool       `bson:"isenabled" json:"isenabled"`
	WelcomeMessage   string     `bson:"welcomemessage" json:"welcomemessage"`
	SystemPrompt     string     `bson:"systemprompt" json:"systemprompt"`
	MenuItems        []MenuItem `bson:"menuitems" json:"menuitems"`
	CreatedAt        time.Time  `bson:"createdat" json:"createdat"`
	UpdatedAt        time.Time  `bson:"updatedat" json:"updatedat"`
}

type MenuItem struct {
	Label       string `bson:"label" json:"label"`
	Text        string `bson:"text" json:"text"`
	Description string `bson:"description" json:"description"`
}
