package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dermasense/backend/internal/config"
	"github.com/dermasense/backend/internal/domain/entity"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIService handles communication with the OpenAI Vision API.
type OpenAIService struct {
	client *openai.Client
	cfg    *config.Config
}

// NewOpenAIService creates a new OpenAIService.
func NewOpenAIService(cfg *config.Config) *OpenAIService {
	clientCfg := openai.DefaultConfig(cfg.OpenAI.APIKey)
	clientCfg.BaseURL = cfg.OpenAI.BaseURL
	client := openai.NewClientWithConfig(clientCfg)
	return &OpenAIService{
		client: client,
		cfg:    cfg,
	}
}


const skinAnalysisPrompt = `You are an expert dermatologist AI assistant. Analyze the provided skin image and return a structured JSON diagnosis.

You MUST respond with ONLY a valid JSON object in the following exact format, with no additional text:

{
  "diagnosis": "Name of the skin condition detected",
  "confidence": 0.85,
  "severity": "low",
  "explanation": "Detailed medical explanation of the condition, what it means, and what causes it.",
  "recommendations": [
    "First actionable recommendation",
    "Second actionable recommendation",
    "Third actionable recommendation"
  ]
}

Rules:
- "diagnosis" must be a clear, specific skin condition name
- "confidence" must be a number between 0.0 and 1.0 representing your confidence level
- "severity" must be exactly one of: "low", "moderate", or "high"
- "explanation" must be a thorough but accessible explanation
- "recommendations" must be an array of 3-5 specific, actionable recommendations
- If you cannot identify a skin condition, set diagnosis to "Unidentifiable", confidence to 0.0, severity to "low", and explain why in the explanation field
- Always recommend consulting a dermatologist for professional diagnosis

IMPORTANT: Return ONLY the JSON object. No markdown, no code blocks, no additional text.`

// AnalyzeSkinImage sends a skin image to OpenAI Vision API for analysis.
func (s *OpenAIService) AnalyzeSkinImage(ctx context.Context, imageData []byte) (*entity.ScanResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.OpenAI.Timeout)
	defer cancel()

	// Encode image as base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)

	// Build the message with image
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: skinAnalysisPrompt,
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    imageURL,
						Detail: openai.ImageURLDetailHigh,
					},
				},
			},
		},
	}

	// Call OpenAI API with retry
	var result *entity.ScanResult
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       s.cfg.OpenAI.Model,
			Messages:    messages,
			MaxTokens:   1024,
			Temperature: 0.3,
		})

		if err != nil {
			lastErr = fmt.Errorf("openai api error (attempt %d): %w", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		if len(resp.Choices) == 0 {
			lastErr = fmt.Errorf("no response from openai (attempt %d)", attempt+1)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Parse the JSON response
		content := resp.Choices[0].Message.Content
		result, err = parseAnalysisResponse(content)
		if err != nil {
			lastErr = fmt.Errorf("failed to parse response (attempt %d): %w", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		return result, nil
	}

	return nil, fmt.Errorf("all attempts failed: %w", lastErr)
}

// parseAnalysisResponse parses the JSON response from OpenAI.
func parseAnalysisResponse(content string) (*entity.ScanResult, error) {
	var result entity.ScanResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("invalid json response: %w, content: %s", err, content)
	}

	// Validate required fields
	if result.Diagnosis == "" {
		return nil, fmt.Errorf("missing diagnosis in response")
	}
	if result.Confidence < 0 || result.Confidence > 1 {
		result.Confidence = 0.5 // Default to moderate confidence
	}

	// Normalize severity
	switch result.Severity {
	case "low", "moderate", "high":
		// valid
	default:
		result.Severity = "moderate"
	}

	return &result, nil
}
