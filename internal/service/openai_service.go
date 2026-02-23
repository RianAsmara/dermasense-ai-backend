package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
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

// System prompt separated from user message for better instruction following
// and potential prompt caching by the API.
const skinAnalysisSystemPrompt = `You are an expert dermatologist AI assistant. You analyze skin images and return structured JSON diagnoses in BOTH English and Indonesian (Bahasa Indonesia).

Rules:
- "diagnosis" must be a clear, specific skin condition name in English
- "diagnosis_id" must be the same condition name translated to Bahasa Indonesia
- "confidence" must be a number between 0.0 and 1.0 representing your confidence level
- "severity" must be exactly one of: "low", "moderate", or "high"
- "explanation" must be a thorough but accessible explanation in English
- "explanation_id" must be the same explanation translated to Bahasa Indonesia
- "recommendations" must be an array of 3-5 specific, actionable recommendations in English
- "recommendations_id" must be the same recommendations translated to Bahasa Indonesia
- If you cannot identify a skin condition, set diagnosis to "Unidentifiable", diagnosis_id to "Tidak Teridentifikasi", confidence to 0.0, severity to "low", and explain why in both explanation fields
- Always recommend consulting a dermatologist for professional diagnosis`

const skinAnalysisUserPrompt = `Analyze this skin image. Return a JSON object with these exact fields:

{
  "diagnosis": "English name",
  "diagnosis_id": "Nama Bahasa Indonesia",
  "confidence": 0.85,
  "severity": "low",
  "explanation": "English explanation.",
  "explanation_id": "Penjelasan Bahasa Indonesia.",
  "recommendations": ["English rec 1", "English rec 2", "English rec 3"],
  "recommendations_id": ["Rekomendasi ID 1", "Rekomendasi ID 2", "Rekomendasi ID 3"]
}`

// AnalyzeSkinImage sends a skin image to OpenAI Vision API for analysis.
func (s *OpenAIService) AnalyzeSkinImage(ctx context.Context, imageData []byte, mimeType string) (*entity.ScanResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.OpenAI.Timeout)
	defer cancel()

	// Use the actual MIME type from the upload, fallback to jpeg
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	// Encode image as base64 with correct MIME type
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	// Build messages: system prompt (cacheable) + user prompt with image
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: skinAnalysisSystemPrompt,
		},
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: skinAnalysisUserPrompt,
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    imageURL,
						Detail: openai.ImageURLDetailLow, // Low detail = cheaper, sufficient for skin analysis
					},
				},
			},
		},
	}

	// Call OpenAI API with retry (respects context cancellation)
	var result *entity.ScanResult
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		// Check if context is already cancelled before retrying
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return nil, fmt.Errorf("context cancelled after %d attempts: %w", attempt, lastErr)
			}
			return nil, fmt.Errorf("context cancelled: %w", err)
		}

		resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       s.cfg.OpenAI.Model,
			Messages:    messages,
			MaxTokens:   2048,
			Temperature: 0.3,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
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

		// Parse the JSON response (strip markdown wrapper if present)
		content := stripMarkdownJSON(resp.Choices[0].Message.Content)
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

// stripMarkdownJSON removes ```json ... ``` wrappers that LLMs sometimes add.
func stripMarkdownJSON(content string) string {
	content = strings.TrimSpace(content)

	// Remove ```json prefix
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return content
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
