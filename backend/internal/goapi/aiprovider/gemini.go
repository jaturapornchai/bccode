package aiprovider

import (
	"context"
	"smlcloudplatform/internal/goapi/gemini"
)

// geminiProvider wraps gemini.GeminiClient เป็น AIProvider
type geminiProvider struct {
	client *gemini.GeminiClient
}

func newGeminiProvider() AIProvider {
	return &geminiProvider{
		client: gemini.NewGeminiClient(),
	}
}

func (g *geminiProvider) Name() string {
	return "gemini"
}

func (g *geminiProvider) GenerateContent(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// แปลง ChatRequest → GeminiRequest
	geminiReq := gemini.GeminiRequest{
		Contents: []gemini.Content{
			{
				Role: "user",
				Parts: []gemini.Part{
					{Text: req.UserPrompt},
				},
			},
		},
		GenerationConfig: gemini.GenerationConfig{
			Temperature: req.Temperature,
			TopP:        req.TopP,
			TopK:        40,
		},
	}

	if req.MaxTokens > 0 {
		geminiReq.GenerationConfig.MaxOutputTokens = req.MaxTokens
	}

	if req.SystemPrompt != "" {
		geminiReq.SystemInstruction = &gemini.SystemInstruction{
			Parts: []gemini.Part{
				{Text: req.SystemPrompt},
			},
		}
	}

	resp, err := g.client.GenerateContent(ctx, geminiReq)
	if err != nil {
		return nil, err
	}

	text := gemini.ExtractTextFromResponse(resp)

	return &ChatResponse{
		Text:             text,
		PromptTokens:     resp.UsageMetadata.PromptTokenCount,
		CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		Model:            g.client.Model,
	}, nil
}
