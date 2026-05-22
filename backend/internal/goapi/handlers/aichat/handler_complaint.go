package aichat

import (
	"net/http"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"github.com/labstack/echo/v4"
)

// SubmitComplaintRequest — request ร้องเรียน AI จาก frontend
type SubmitComplaintRequest struct {
	ShopID string `json:"shop_id" validate:"required"`
	ModelID string `json:"model_id" validate:"required"`
	Category string `json:"category" validate:"required"` // wrong_answer, gibberish, wrong_language, refused, hallucination, too_short, irrelevant
	Question string `json:"question,omitempty"`
	Answer string `json:"answer,omitempty"`
}

// SubmitComplaint — POST /api/v1/ai-provider/complaint
// ส่งร้องเรียน model ไปยัง bcproxyai
func SubmitComplaint(c echo.Context) error {
	var req SubmitComplaintRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false, "message": "Invalid request", "error": err.Error(),
		})
	}
	if req.ShopID == "" || req.ModelID == "" || req.Category == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false, "message": "shop_id, model_id, and category are required",
		})
	}

	// หา base URL ของ custom provider (bcproxyai)
	configs, err := getAIProviderConfigs(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false, "message": "ไม่สามารถดึง config ได้", "error": err.Error(),
		})
	}

	var baseURL string
	for _, cfg := range configs {
		if cfg.BaseURL != "" {
			baseURL = cfg.BaseURL
			break
		}
	}

	if baseURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false, "message": "ไม่พบ base URL สำหรับร้องเรียน (ต้องใช้ bcproxyai)",
		})
	}

	complaint := aiprovider.BCProxyComplaint{
		ModelID:          req.ModelID,
		Category:         req.Category,
		Reason:           "manual complaint from user",
		UserMessage:      req.Question,
		AssistantMessage: req.Answer,
		Source:           "api",
	}

	if err := aiprovider.SendBCProxyComplaint(baseURL, complaint); err != nil {
		logger.Error("[Complaint] failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false, "message": "ร้องเรียนไม่สำเร็จ", "error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":   true,
		"message":   "ร้องเรียนสำเร็จ — ระบบจะตรวจสอบและปรับปรุง AI",
		"timestamp": time.Now(),
	})
}
