package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"awesomeProject/internal/models"
)

// AIClient interface defines methods for AI model interactions
type AIClient interface {
	Generate(ctx context.Context, message string, settings models.BotSettings) (string, error)
	Stream(ctx context.Context, message string, settings models.BotSettings) (<-chan string, error)
}

// GPT4Client implements AIClient for GPT-4
type GPT4Client struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// NewGPT4Client creates a new GPT-4 client
func NewGPT4Client(config json.RawMessage) (*GPT4Client, error) {
	var cfg struct {
		APIKey string `json:"api_key"`
		APIURL string `json:"api_url"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse GPT-4 config: %v", err)
	}

	return &GPT4Client{
		apiKey: cfg.APIKey,
		apiURL: cfg.APIURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Generate generates a response using GPT-4
func (c *GPT4Client) Generate(ctx context.Context, message string, settings models.BotSettings) (string, error) {
	reqBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": c.buildSystemPrompt(settings),
			},
			{
				"role":    "user",
				"content": message,
			},
		},
		"max_tokens":   settings.MaxTokens,
		"temperature":  settings.Temperature,
		"stop":        settings.StopSequences,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return result.Choices[0].Message.Content, nil
}

// Stream streams a response using GPT-4
func (c *GPT4Client) Stream(ctx context.Context, message string, settings models.BotSettings) (<-chan string, error) {
	responseChan := make(chan string, 100)

	reqBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": c.buildSystemPrompt(settings),
			},
			{
				"role":    "user",
				"content": message,
			},
		},
		"max_tokens":   settings.MaxTokens,
		"temperature":  settings.Temperature,
		"stop":        settings.StopSequences,
		"stream":      true,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		close(responseChan)
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		close(responseChan)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		close(responseChan)
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	go func() {
		defer resp.Body.Close()
		defer close(responseChan)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					responseChan <- fmt.Sprintf("Error: %v", err)
				}
				return
			}

			if strings.HasPrefix(line, "data: ") {
				line = strings.TrimPrefix(line, "data: ")
				if line == "[DONE]" {
					continue
				}

				var result struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(line), &result); err != nil {
					continue
				}

				if len(result.Choices) > 0 && result.Choices[0].Delta.Content != "" {
					responseChan <- result.Choices[0].Delta.Content
				}
			}
		}
	}()

	return responseChan, nil
}

// buildSystemPrompt builds the system prompt for the AI model
func (c *GPT4Client) buildSystemPrompt(settings models.BotSettings) string {
	var prompt strings.Builder

	// Add custom prompts
	for _, customPrompt := range settings.CustomPrompts {
		prompt.WriteString(customPrompt + "\n")
	}

	// Add restricted topics
	if len(settings.RestrictedTopics) > 0 {
		prompt.WriteString("Please avoid discussing the following topics:\n")
		for _, topic := range settings.RestrictedTopics {
			prompt.WriteString("- " + topic + "\n")
		}
	}

	return prompt.String()
}

// ClaudeClient implements AIClient for Claude
type ClaudeClient struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(config json.RawMessage) (*ClaudeClient, error) {
	var cfg struct {
		APIKey string `json:"api_key"`
		APIURL string `json:"api_url"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse Claude config: %v", err)
	}

	return &ClaudeClient{
		apiKey: cfg.APIKey,
		apiURL: cfg.APIURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Generate generates a response using Claude
func (c *ClaudeClient) Generate(ctx context.Context, message string, settings models.BotSettings) (string, error) {
	// Implementation for Claude's generate method
	// Similar to GPT4Client but using Claude's API
	return "", nil
}

// Stream streams a response using Claude
func (c *ClaudeClient) Stream(ctx context.Context, message string, settings models.BotSettings) (<-chan string, error) {
	// Implementation for Claude's stream method
	// Similar to GPT4Client but using Claude's API
	return nil, nil
} 