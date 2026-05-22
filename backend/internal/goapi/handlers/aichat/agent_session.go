package aichat

import (
	"smlcloudplatform/internal/goapi/aiprovider"
	"sync"
	"time"
)

// ==================== In-Memory Session Store ====================
//
// กฎ: backend ห้ามเขียน DB เลย → session memory เก็บใน RAM อย่างเดียว
// ลบ MongoDB-backed session storage ออกทั้งหมด เปลี่ยนเป็น sync.Map + TTL
//
// สถาปัตยกรรมเดิม: load/save ทุก request → MongoDB
// สถาปัตยกรรมใหม่: load/save → in-memory cache, อายุ 1 ชม. cleanup ทุก 10 นาที
// หมายเหตุ: ข้อมูลหายเมื่อ goapi restart — ยอมรับได้สำหรับ chatbot

const (
	sessionTTL          = 1 * time.Hour
	sessionJanitorEvery = 10 * time.Minute
	maxSessionMessages  = 50
)

// ChatSessionDoc — เก็บ conversation history ใน RAM (ชื่อเดิมเพื่อ backward compat)
type ChatSessionDoc struct {
	SessionID string           `json:"session_id"`
	ShopID string           `json:"shop_id"`
	Messages []SessionMessage `json:"messages"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// SessionMessage — message ใน session (เก็บแค่ user + assistant, ไม่เก็บ tool/system)
type SessionMessage struct {
	Role string    `json:"role"`
	Content string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type sessionEntry struct {
	doc       *ChatSessionDoc
	expiresAt time.Time
}

var (
	sessionStore   = make(map[string]*sessionEntry)
	sessionStoreMu sync.RWMutex
	janitorOnce    sync.Once
)

// startSessionJanitor เปิด goroutine cleanup expired entries — เรียกครั้งเดียวต่อ process
func startSessionJanitor() {
	go func() {
		ticker := time.NewTicker(sessionJanitorEvery)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			sessionStoreMu.Lock()
			for k, e := range sessionStore {
				if now.After(e.expiresAt) {
					delete(sessionStore, k)
				}
			}
			sessionStoreMu.Unlock()
		}
	}()
}

func sessionStoreKey(sessionID, shopID string) string {
	return shopID + "::" + sessionID
}

// loadSession โหลด session จาก in-memory cache
// ไม่พบ → return session ใหม่ (empty)
func loadSession(sessionID, shopID string) (*ChatSessionDoc, error) {
	janitorOnce.Do(startSessionJanitor)

	sessionStoreMu.RLock()
	entry, ok := sessionStore[sessionStoreKey(sessionID, shopID)]
	sessionStoreMu.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return entry.doc, nil
	}
	return &ChatSessionDoc{
		SessionID: sessionID,
		ShopID:    shopID,
		Messages:  []SessionMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// saveSession เก็บ session ลง in-memory cache (ไม่แตะ DB)
func saveSession(doc *ChatSessionDoc) error {
	if doc == nil || doc.SessionID == "" {
		return nil
	}
	if len(doc.Messages) > maxSessionMessages {
		doc.Messages = doc.Messages[len(doc.Messages)-maxSessionMessages:]
	}
	doc.UpdatedAt = time.Now()

	sessionStoreMu.Lock()
	sessionStore[sessionStoreKey(doc.SessionID, doc.ShopID)] = &sessionEntry{
		doc:       doc,
		expiresAt: time.Now().Add(sessionTTL),
	}
	sessionStoreMu.Unlock()
	return nil
}

// sessionToOAIMessages แปลง session history เป็น OAI messages
// จำกัดเป็น 4 messages ล่าสุด (2 turns) เพื่อลด context pollution
func sessionToOAIMessages(session *ChatSessionDoc) []aiprovider.OAIMessage {
	const maxHistory = 4
	start := 0
	if len(session.Messages) > maxHistory {
		start = len(session.Messages) - maxHistory
	}
	var msgs []aiprovider.OAIMessage
	for _, m := range session.Messages[start:] {
		msgs = append(msgs, aiprovider.OAIMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return msgs
}

// appendToSession เพิ่ม message เข้า session (in-memory)
func appendToSession(session *ChatSessionDoc, role, content string) {
	session.Messages = append(session.Messages, SessionMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// clearSession ลบ session ออกจาก in-memory cache
func clearSession(sessionID, shopID string) error {
	sessionStoreMu.Lock()
	delete(sessionStore, sessionStoreKey(sessionID, shopID))
	sessionStoreMu.Unlock()
	return nil
}
