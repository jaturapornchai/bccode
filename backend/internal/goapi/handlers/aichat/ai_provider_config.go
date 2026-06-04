package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// supportedProviders — รายชื่อ provider ที่รองรับ (custom-* ยอมรับทุกชื่อที่ขึ้นต้นด้วย custom)
var supportedProviders = []string{"openrouter", "groq", "deepseek", "gemini", "ollama", "custom"}

// providerBaseURLs — base URL ของแต่ละ provider (OpenAI-compatible)
var providerBaseURLs = map[string]string{
	"openrouter": "https://openrouter.ai/api/v1/chat/completions",
	"groq":       "https://api.groq.com/openai/v1/chat/completions",
	"deepseek":   "https://api.deepseek.com/v1/chat/completions",
	"ollama":     "http://host.docker.internal:11434/v1/chat/completions",
}

// providerDefaultModels — default model ของแต่ละ provider
var providerDefaultModels = map[string]string{
	"openrouter": "google/gemini-2.0-flash-exp:free",
	"groq":       "llama-3.3-70b-versatile",
	"deepseek":   "deepseek-chat",
	"gemini":     "gemini-2.0-flash",
}

// ---- Request structs ----

type listAIProvidersReq struct {
	HoldingCode string `json:"holding_code"`
}

type saveAIProviderReq struct {
	HoldingCode  string   `json:"holding_code"`
	ProviderName string   `json:"provider_name"`
	APIKey       string   `json:"api_key"`
	BaseURL      string   `json:"base_url"`
	Model        string   `json:"model"`
	Capabilities []string `json:"capabilities"`
	IsActive     bool     `json:"is_active"`
	Priority     int      `json:"priority"`
}

type deleteAIProviderReq struct {
	HoldingCode  string `json:"holding_code"`
	ProviderName string `json:"provider_name"`
}

type testAIProviderReq struct {
	HoldingCode  string `json:"holding_code"`
	ProviderName string `json:"provider_name"`
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url"`
	Model        string `json:"model"`
}

type aiProviderStatusReq struct {
	HoldingCode string `json:"holding_code"`
}

// ---- Handlers ----

// ListAIProviders — POST /api/v1/ai-provider/list
func ListAIProviders(c echo.Context) error {
	var req listAIProvidersReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "invalid request: " + err.Error(),
		})
	}
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "holding_code is required",
		})
	}

	logger.Info("[AIProvider] list holding_code=%s", req.HoldingCode)
	configs, err := getAIProviderConfigs(req.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "ดึงข้อมูลไม่ได้: " + err.Error(),
		})
	}

	// mask API key (แสดงแค่ 8 ตัวแรก)
	masked := make([]map[string]any, 0, len(configs))
	for _, cfg := range configs {
		m := map[string]any{
			"provider_name":  cfg.ProviderName,
			"api_key_masked": maskAPIKey(cfg.APIKey),
			"base_url":       cfg.BaseURL,
			"model":          cfg.Model,
			"capabilities":   cfg.Capabilities,
			"is_active":      cfg.IsActive,
			"priority":       cfg.Priority,
			"last_error":     cfg.LastError,
			"last_error_at":  cfg.LastErrorAt,
			"cooldown_until": cfg.CooldownUntil,
			"created_at":     cfg.CreatedAt,
			"updated_at":     cfg.UpdatedAt,
		}
		masked = append(masked, m)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":             true,
		"providers":           masked,
		"available_providers": supportedProviders,
	})
}

// SaveAIProvider — POST /api/v1/ai-provider/save
func SaveAIProvider(c echo.Context) error {
	var req saveAIProviderReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "invalid request: " + err.Error(),
		})
	}
	if req.HoldingCode == "" || req.ProviderName == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "holding_code and provider_name are required",
		})
	}
	// custom provider ไม่บังคับ api_key (เช่น Ollama, local proxy)
	if !strings.HasPrefix(req.ProviderName, "custom") && req.ProviderName != "ollama" && req.APIKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "api_key is required",
		})
	}
	if !isSupportedProvider(req.ProviderName) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": fmt.Sprintf("provider '%s' ไม่รองรับ รองรับ: %s", req.ProviderName, strings.Join(supportedProviders, ", ")),
		})
	}

	// Custom provider ต้องมี base_url
	if strings.HasPrefix(req.ProviderName, "custom") && req.BaseURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "custom provider ต้องระบุ base_url",
		})
	}

	if req.Model == "" {
		req.Model = providerDefaultModels[req.ProviderName]
	}

	cfg := AIProviderConfig{
		ProviderName: req.ProviderName,
		APIKey:       req.APIKey,
		BaseURL:      req.BaseURL,
		Model:        req.Model,
		Capabilities: req.Capabilities,
		IsActive:     req.IsActive,
		Priority:     req.Priority,
	}

	if err := upsertAIProviderConfig(req.HoldingCode, cfg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "บันทึกไม่ได้: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("บันทึก %s สำเร็จ", req.ProviderName),
	})
}

// DeleteAIProvider — POST /api/v1/ai-provider/delete
func DeleteAIProvider(c echo.Context) error {
	var req deleteAIProviderReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "invalid request: " + err.Error(),
		})
	}
	if req.HoldingCode == "" || req.ProviderName == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "holding_code and provider_name are required",
		})
	}

	if err := deleteAIProviderConfig(req.HoldingCode, req.ProviderName); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "ลบไม่ได้: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("ลบ %s สำเร็จ", req.ProviderName),
	})
}

// TestAIProvider — POST /api/v1/ai-provider/test
func TestAIProvider(c echo.Context) error {
	var req testAIProviderReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "invalid request: " + err.Error(),
		})
	}
	if req.ProviderName == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "provider_name is required",
		})
	}
	if !strings.HasPrefix(req.ProviderName, "custom") && req.ProviderName != "ollama" && req.APIKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "api_key is required",
		})
	}
	if !isSupportedProvider(req.ProviderName) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": fmt.Sprintf("provider '%s' ไม่รองรับ", req.ProviderName),
		})
	}

	model := req.Model
	if model == "" {
		model = providerDefaultModels[req.ProviderName]
	}

	start := time.Now()
	var responseText string
	var testErr error

	ctx, cancel := context.WithTimeout(c.Request().Context(), 120*time.Second)
	defer cancel()

	if req.ProviderName == "gemini" {
		responseText, testErr = testGeminiProvider(ctx, req.APIKey, model)
	} else if strings.HasPrefix(req.ProviderName, "custom") {
		// Custom provider — ใช้ base_url จาก request
		if req.BaseURL == "" {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"message": "custom provider ต้องระบุ base_url",
			})
		}
		chatURL := req.BaseURL
		if !strings.HasSuffix(chatURL, "/chat/completions") {
			chatURL = strings.TrimRight(chatURL, "/") + "/chat/completions"
		}
		responseText, testErr = testOpenAICompatProviderURL(ctx, chatURL, req.APIKey, model)
	} else if req.ProviderName == "ollama" {
		// Ollama — ใช้ base_url จาก request หรือ default
		chatURL := providerBaseURLs["ollama"]
		if req.BaseURL != "" {
			chatURL = strings.TrimRight(req.BaseURL, "/") + "/chat/completions"
		}
		responseText, testErr = testOpenAICompatProviderURL(ctx, chatURL, "", model)
	} else {
		responseText, testErr = testOpenAICompatProvider(ctx, req.ProviderName, req.APIKey, model)
	}

	elapsed := time.Since(start).Milliseconds()

	if testErr != nil {
		logger.Warn("[AIProvider] test failed provider=%s: %v", req.ProviderName, testErr)
		return c.JSON(http.StatusOK, map[string]any{
			"success":          false,
			"message":          fmt.Sprintf("%s ทดสอบไม่ผ่าน: %s", req.ProviderName, testErr.Error()),
			"response_time_ms": elapsed,
		})
	}

	logger.Info("[AIProvider] test passed provider=%s model=%s (%dms)", req.ProviderName, model, elapsed)

	// Auto-detect capabilities
	var chatURL string
	if req.ProviderName == "ollama" {
		chatURL = providerBaseURLs["ollama"]
		if req.BaseURL != "" {
			chatURL = strings.TrimRight(req.BaseURL, "/") + "/chat/completions"
		}
	} else if strings.HasPrefix(req.ProviderName, "custom") {
		chatURL = req.BaseURL
		if !strings.HasSuffix(chatURL, "/chat/completions") {
			chatURL = strings.TrimRight(chatURL, "/") + "/chat/completions"
		}
	} else if url, ok := providerBaseURLs[req.ProviderName]; ok {
		chatURL = url
	}

	var capabilities []string
	if chatURL != "" {
		detectCtx, detectCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer detectCancel()
		capabilities = detectCapabilities(detectCtx, chatURL, req.APIKey, model)
		logger.Info("[AIProvider] detected capabilities for %s/%s: %v", req.ProviderName, model, capabilities)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":          true,
		"message":          fmt.Sprintf("%s ทำงานปกติ", req.ProviderName),
		"response_time_ms": elapsed,
		"model_used":       model,
		"response_preview": truncate(responseText, 200),
		"capabilities":     capabilities,
	})
}

// AIProviderStatus — POST /api/v1/ai-provider/status
func AIProviderStatus(c echo.Context) error {
	var req aiProviderStatusReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "invalid request: " + err.Error(),
		})
	}
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "holding_code is required",
		})
	}

	configs, err := getAIProviderConfigs(req.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "ดึงข้อมูลไม่ได้: " + err.Error(),
		})
	}

	now := time.Now()
	statuses := make([]map[string]any, 0, len(configs))
	for _, cfg := range configs {
		status := "active"
		if !cfg.IsActive {
			status = "inactive"
		} else if cfg.CooldownUntil != nil && cfg.CooldownUntil.After(now) {
			status = "cooldown"
		}

		statuses = append(statuses, map[string]any{
			"provider_name":  cfg.ProviderName,
			"model":          cfg.Model,
			"is_active":      cfg.IsActive,
			"priority":       cfg.Priority,
			"status":         status,
			"last_error":     cfg.LastError,
			"last_error_at":  cfg.LastErrorAt,
			"cooldown_until": cfg.CooldownUntil,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":  true,
		"statuses": statuses,
	})
}

// ---- helpers ----

func isSupportedProvider(name string) bool {
	// custom provider: ชื่อขึ้นต้นด้วย "custom" (เช่น custom, custom-myproxy, custom-localai)
	if strings.HasPrefix(name, "custom") {
		return true
	}
	for _, p := range supportedProviders {
		if p == name {
			return true
		}
	}
	return false
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:8] + "****"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// testOpenAICompatProvider — ทดสอบ provider แบบ OpenAI-compatible (known providers)
func testOpenAICompatProvider(ctx context.Context, providerName, apiKey, model string) (string, error) {
	baseURL, ok := providerBaseURLs[providerName]
	if !ok {
		return "", fmt.Errorf("ไม่พบ base URL สำหรับ %s", providerName)
	}
	return testOpenAICompatProviderURL(ctx, baseURL, apiKey, model)
}

// testOpenAICompatProviderURL — ทดสอบ OpenAI-compatible endpoint ด้วย URL ที่กำหนด
func testOpenAICompatProviderURL(ctx context.Context, chatURL, apiKey, model string) (string, error) {
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "Say hello in Thai in one short sentence."},
		},
		"max_tokens": 50,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API status %d: %s", resp.StatusCode, string(body))
	}

	var oaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &oaiResp); err != nil {
		return "", fmt.Errorf("parse response error: %w", err)
	}
	if len(oaiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return oaiResp.Choices[0].Message.Content, nil
}

// detectCapabilities — ทดสอบ capabilities ของ model อัตโนมัติ
// ส่ง request แต่ละแบบ แล้วดูว่า model รองรับหรือไม่
func detectCapabilities(ctx context.Context, chatURL, apiKey, model string) []string {
	var caps []string

	// 1. Test tools (function calling)
	toolReq := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "user", "content": "What is 2+2? Use the calculator tool."},
		},
		"tools": []map[string]any{
			{
				"type": "function",
				"function": map[string]any{
					"name":        "calculator",
					"description": "Calculate math",
					"parameters": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"expression": map[string]any{"type": "string"},
						},
					},
				},
			},
		},
		"max_tokens": 100,
	}
	if testRequest(ctx, chatURL, apiKey, toolReq) {
		caps = append(caps, "tools")
	}

	// 2. Test vision (image input)
	visionReq := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": "What color is this?"},
					{"type": "image_url", "image_url": map[string]string{
						"url": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
					}},
				},
			},
		},
		"max_tokens": 50,
	}
	if testRequest(ctx, chatURL, apiKey, visionReq) {
		caps = append(caps, "vision")
	}

	// 3. Test thinking (reasoning) — เช็คจาก model name
	lowerModel := strings.ToLower(model)
	if strings.Contains(lowerModel, "think") || strings.Contains(lowerModel, "reason") ||
		strings.Contains(lowerModel, "deepseek-r1") || strings.Contains(lowerModel, "qwq") {
		caps = append(caps, "thinking")
	}

	return caps
}

// testRequest — ส่ง request ไป API แล้วเช็คว่าได้ 200 หรือไม่
func testRequest(ctx context.Context, chatURL, apiKey string, reqBody map[string]any) bool {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return false
	}

	testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(testCtx, "POST", chatURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return false
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body) // drain body

	return resp.StatusCode == http.StatusOK
}

// ---- List Models ----

type listModelsReq struct {
	ProviderName string `json:"provider_name"`
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url"`
}

// ListAIModels — POST /api/v1/ai-provider/models
// ดึงรายการ models จาก provider API
func ListAIModels(c echo.Context) error {
	var req listModelsReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "invalid request: " + err.Error(),
		})
	}
	if req.ProviderName == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "provider_name is required",
		})
	}
	if !strings.HasPrefix(req.ProviderName, "custom") && req.ProviderName != "ollama" && req.APIKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "api_key is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()

	var modelsURL string

	switch {
	case req.ProviderName == "gemini":
		return listGeminiModels(c, ctx, req.APIKey)
	case strings.HasPrefix(req.ProviderName, "custom"):
		if req.BaseURL == "" {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"message": "custom provider ต้องระบุ base_url",
			})
		}
		modelsURL = strings.TrimRight(req.BaseURL, "/") + "/models"
	default:
		// known provider — ใช้ base URL ของ provider
		chatURL, ok := providerBaseURLs[req.ProviderName]
		if !ok {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"message": fmt.Sprintf("provider '%s' ไม่รองรับ", req.ProviderName),
			})
		}
		// Ollama: ถ้า user ส่ง base_url มา (เครื่องอื่น) → ใช้ค่านั้น
		if req.ProviderName == "ollama" && req.BaseURL != "" {
			chatURL = strings.TrimRight(req.BaseURL, "/") + "/chat/completions"
		}
		// /chat/completions → /models
		modelsURL = strings.Replace(chatURL, "/chat/completions", "/models", 1)
	}

	// OpenAI-compatible /models endpoint
	httpReq, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false, "message": err.Error(),
		})
	}
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false, "message": fmt.Sprintf("ดึง models ไม่ได้: %v", err),
		})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": fmt.Sprintf("API status %d: %s", resp.StatusCode, truncate(string(body), 200)),
		})
	}

	var modelsResp struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false, "message": "parse error: " + err.Error(),
		})
	}

	models := make([]map[string]string, 0, len(modelsResp.Data))
	for _, m := range modelsResp.Data {
		models = append(models, map[string]string{
			"id":       m.ID,
			"owned_by": m.OwnedBy,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"models":  models,
		"count":   len(models),
	})
}

// listGeminiModels — ดึง models จาก Gemini API
func listGeminiModels(c echo.Context, ctx context.Context, apiKey string) error {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false, "message": err.Error(),
		})
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false, "message": fmt.Sprintf("ดึง models ไม่ได้: %v", err),
		})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": fmt.Sprintf("API status %d: %s", resp.StatusCode, truncate(string(body), 200)),
		})
	}

	var geminiResp struct {
		Models []struct {
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false, "message": "parse error: " + err.Error(),
		})
	}

	models := make([]map[string]string, 0, len(geminiResp.Models))
	for _, m := range geminiResp.Models {
		// Gemini name format: "models/gemini-2.0-flash" → extract "gemini-2.0-flash"
		id := m.Name
		if strings.HasPrefix(id, "models/") {
			id = strings.TrimPrefix(id, "models/")
		}
		models = append(models, map[string]string{
			"id":       id,
			"owned_by": m.DisplayName,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"models":  models,
		"count":   len(models),
	})
}

// testGeminiProvider — ทดสอบ Gemini provider (Google API)
func testGeminiProvider(ctx context.Context, apiKey, model string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": "Say hello in Thai in one short sentence."},
				},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": 50,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("parse response error: %w", err)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}
	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
