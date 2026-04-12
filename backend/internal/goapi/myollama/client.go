package myollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"io"
	"net/http"
	"time"
)

const (
	// OllamaURL - URL ของ Ollama service
	OllamaURL = "http://ollama:11434/api/embed"

	// ModelName - nomic-embed-text (768 dim, supports Thai)
	ModelName = "nomic-embed-text"
)

var httpClient *http.Client

func init() {
	// สร้าง HTTP client สำหรับเรียก Ollama service
	httpClient = &http.Client{
		Timeout: 60 * time.Second, // E5 อาจช้ากว่า FastEmbed
	}
}

// EmbedRequest - Ollama API request
type EmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// EmbedResponse - Ollama API response
type EmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Model      string      `json:"model"`
}

// GenerateEmbeddings - สร้าง embeddings จาก Ollama E5
func GenerateEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	embeddings := make([][]float32, 0, len(texts))

	// Ollama embed API รับทีละ 1 text (ไม่รองรับ batch แบบ array)
	// ต้อง loop เรียกทีละตัว
	for i, text := range texts {
		reqBody := EmbedRequest{
			Model: ModelName,
			Input: text,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %v", err)
		}

		resp, err := httpClient.Post(OllamaURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to call ollama (text %d/%d): %v", i+1, len(texts), err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("ollama error %d (text %d/%d): %s", resp.StatusCode, i+1, len(texts), string(body))
		}

		var embedResp EmbedResponse
		if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
			return nil, fmt.Errorf("failed to decode response (text %d/%d): %v", i+1, len(texts), err)
		}

		if len(embedResp.Embeddings) == 0 {
			return nil, fmt.Errorf("no embedding returned (text %d/%d)", i+1, len(texts))
		}

		embeddings = append(embeddings, embedResp.Embeddings[0])
	}

	logger.Info("✅ Generated %d embeddings using Ollama E5 (model: %s)", len(embeddings), ModelName)
	return embeddings, nil
}

// GenerateSingleEmbedding - สร้าง embedding สำหรับ text เดียว
func GenerateSingleEmbedding(text string) ([]float32, error) {
	embeddings, err := GenerateEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return embeddings[0], nil
}
