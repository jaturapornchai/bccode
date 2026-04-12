package aichat

import (
	"fmt"
	"strings"
)

// ==================== OpenClaw Conventions ====================
//
// Reference: docs.openclaw.ai/gateway/protocol (v3)
//
// Project นี้ไม่ได้ใช้ OpenClaw client/library — เราคุยกับ bcproxy ผ่าน HTTP
// /v1/chat/completions ตามปกติ. แต่เรา "ยืม pattern" ของ OpenClaw มาใช้เพื่อให้
// โครงสร้างข้อมูลในระบบเรา consistent กับมาตรฐาน OpenClaw:
//
// 1. session_key format: agent:{agentId}:{channel}:{type}:{identifier}
// 2. Error envelope: {code, message, details: {code, reason, recommendedNextStep}}
// 3. Untrusted external content wrapping (อยู่ใน untrusted_wrap.go)
//
// เป้าหมาย: ถ้าวันหนึ่งเปลี่ยนไปใช้ OpenClaw gateway จริงๆ — refactor ง่าย
// เพราะ data shape ตรงกันแล้ว

const (
	// agentID — ชื่อ agent หลักของ น้องกุ้ง (BC AI Account chatbot)
	agentID = "nongkung"

	// channelWebChat — channel สำหรับ chat ที่มาจาก Flutter app
	channelWebChat = "webchat"

	// sessionTypeShop — session ผูกกับ shop (ไม่ใช่ direct user)
	sessionTypeShop = "shop"
)

// BuildSessionKey สร้าง session_key รูปแบบ OpenClaw
//
//	agent:nongkung:webchat:shop:{shopID}:{sessionID}
//
// ถ้า sessionID ว่าง → fallback เป็น "anon" (chat แบบไม่มี memory)
func BuildSessionKey(shopID, sessionID string) string {
	if sessionID == "" {
		sessionID = "anon"
	}
	// sanitize: ตัด ":" ออกเพื่อไม่ให้ key พัง
	shopID = strings.ReplaceAll(shopID, ":", "_")
	sessionID = strings.ReplaceAll(sessionID, ":", "_")
	return fmt.Sprintf("agent:%s:%s:%s:%s:%s",
		agentID, channelWebChat, sessionTypeShop, shopID, sessionID)
}

// ==================== OpenClaw Error Envelope ====================
//
// ใช้แทน plain {success:false, message:"..."} เดิม เพื่อให้ client/log
// ได้ structured error ที่เทียบกับ OpenClaw spec ได้

// OpenClawError — error envelope ตาม OpenClaw spec ข้อ 12
type OpenClawError struct {
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Details *OpenClawErrorDetail `json:"details,omitempty"`
}

// OpenClawErrorDetail — รายละเอียดเพิ่มเติม + คำแนะนำว่า client ควรทำอะไรต่อ
type OpenClawErrorDetail struct {
	Code                string `json:"code,omitempty"`                  // เช่น "MODEL_TIMEOUT"
	Reason              string `json:"reason,omitempty"`                // human-readable why
	RecommendedNextStep string `json:"recommendedNextStep,omitempty"`   // hint ให้ client
	CanRetry            bool   `json:"canRetry,omitempty"`              // retry แล้วน่าจะหายไหม
}

// Standard error codes (ตรงกับ OpenClaw spec ที่เราใช้)
const (
	ErrCodeProviderUnavailable = "PROVIDER_UNAVAILABLE"
	ErrCodeModelTimeout        = "MODEL_TIMEOUT"
	ErrCodeInvalidRequest      = "INVALID_REQUEST"
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeNoProvider          = "NO_PROVIDER_CONFIGURED"
	ErrCodeQualityFailed       = "RESPONSE_QUALITY_FAILED"
)

// recommendedNextStep values (ตรงกับ OpenClaw spec)
const (
	NextStepRetry              = "retry"
	NextStepRetryWithBackoff   = "wait_then_retry"
	NextStepConfigureProvider  = "configure_provider"
	NextStepReviewConfig       = "review_auth_configuration"
	NextStepUpdateCredentials  = "update_auth_credentials"
)

// NewOpenClawError สร้าง error envelope แบบเร็ว
func NewOpenClawError(code, message, detailCode, reason, nextStep string, canRetry bool) *OpenClawError {
	err := &OpenClawError{Code: code, Message: message}
	if detailCode != "" || reason != "" || nextStep != "" {
		err.Details = &OpenClawErrorDetail{
			Code:                detailCode,
			Reason:              reason,
			RecommendedNextStep: nextStep,
			CanRetry:            canRetry,
		}
	}
	return err
}

// AsResponseBody แปลง error เป็น map สำหรับ echo c.JSON
func (e *OpenClawError) AsResponseBody() map[string]any {
	return map[string]any{
		"success": false,
		"error":   e,
		// คงไว้สำหรับ frontend เก่าที่อ่าน message ตรงๆ
		"message": e.Message,
	}
}

// classifyAgentError แปลง raw error จาก agent loop → OpenClawError ที่มี hint
// ดูข้อความ error แล้ว map เป็น code ที่เหมาะสม + recommendedNextStep
func classifyAgentError(err error) *OpenClawError {
	if err == nil {
		return NewOpenClawError(ErrCodeInternalError, "unknown error", "", "", "", false)
	}
	msg := err.Error()
	low := strings.ToLower(msg)

	switch {
	case strings.Contains(low, "no ai provider") || strings.Contains(low, "ไม่มี ai provider"):
		return NewOpenClawError(
			ErrCodeNoProvider,
			"ไม่มี AI Provider — กรุณาตั้งค่าก่อน",
			"NO_PROVIDER_CONFIGURED",
			msg,
			NextStepConfigureProvider,
			false,
		)
	case strings.Contains(low, "connection refused") || strings.Contains(low, "ทุก provider ใช้ไม่ได้"):
		return NewOpenClawError(
			ErrCodeProviderUnavailable,
			"AI provider ไม่ตอบสนอง",
			"PROVIDER_UNREACHABLE",
			msg,
			NextStepRetryWithBackoff,
			true,
		)
	case strings.Contains(low, "context deadline") || strings.Contains(low, "timeout"):
		return NewOpenClawError(
			ErrCodeModelTimeout,
			"AI ใช้เวลานานเกินไป",
			"MODEL_TIMEOUT",
			msg,
			NextStepRetry,
			true,
		)
	case strings.Contains(low, "401") || strings.Contains(low, "unauthorized") || strings.Contains(low, "api key"):
		return NewOpenClawError(
			ErrCodeProviderUnavailable,
			"AI provider auth ผิดพลาด",
			"PROVIDER_AUTH_FAILED",
			msg,
			NextStepUpdateCredentials,
			false,
		)
	default:
		return NewOpenClawError(
			ErrCodeInternalError,
			"น้องกุ้งทำงานไม่สำเร็จ",
			"AGENT_LOOP_FAILED",
			msg,
			NextStepRetry,
			true,
		)
	}
}
