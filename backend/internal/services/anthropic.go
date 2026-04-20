package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	anthropicAPIURL  = "https://api.anthropic.com/v1/messages"
	anthropicModel   = "claude-opus-4-5"
	anthropicVersion = "2023-06-01"
	maxTokens        = 1024
)

// AnthropicClient wraps the Anthropic Messages REST API.
type AnthropicClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewAnthropicClient creates an AnthropicClient.
// If apiKey is empty, Enabled() returns false and Generate() returns an error.
func NewAnthropicClient(apiKey string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Enabled reports whether a real API key is configured.
func (c *AnthropicClient) Enabled() bool {
	return c.apiKey != ""
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Generate sends systemPrompt + userPrompt to Claude and returns the generated text.
func (c *AnthropicClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not configured")
	}

	reqBody := anthropicRequest{
		Model:     anthropicModel,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Anthropic API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Anthropic response: %w", err)
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return "", fmt.Errorf("failed to decode Anthropic response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("Anthropic API error (%s): %s", apiResp.Error.Type, apiResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API returned HTTP %d", resp.StatusCode)
	}

	for _, block := range apiResp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("Anthropic response contained no text content")
}
