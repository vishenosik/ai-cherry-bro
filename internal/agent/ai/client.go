package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/vishenosik/ai-cherry-bro/internal/entity"
)

type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

type ChatRequest struct {
	Model       string             `json:"model"`
	Messages    []entity.AiMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message entity.AiMessage `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type Provider string

const (
	ProviderDeepSeek Provider = "deepseek"
	ProviderOpenAI   Provider = "openai"
	ProviderMock     Provider = "mock"
)

type Config struct {
	OpenAiApiKey string `env:"OPENAI_API_KEY"`
}

// NewClient создает клиент на основе доступных API ключей
func NewClient() *Client {

	var conf Config
	err := cleanenv.ReadConfig(".env", &conf)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	fmt.Println("✅ Using OpenAI API")
	return NewOpenAIClient(conf.OpenAiApiKey, "gpt-4")
}

// NewOpenAIClient оригинальная реализация для OpenAI
func NewOpenAIClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Call(messages []entity.AiMessage) (*entity.AiResponse, error) {
	request := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   1000,
		Temperature: 0.1,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Парсинг структурированного ответа
	var aiResp entity.AiResponse
	content := chatResp.Choices[0].Message.Content

	// Пытаемся распарсить JSON ответ
	if strings.Contains(content, "{") && strings.Contains(content, "}") {
		jsonStart := strings.Index(content, "{")
		jsonEnd := strings.LastIndex(content, "}") + 1
		if jsonEnd > jsonStart {
			jsonStr := content[jsonStart:jsonEnd]
			if err := json.Unmarshal([]byte(jsonStr), &aiResp); err == nil {
				return &aiResp, nil
			}
		}
	}

	// Fallback: анализируем текстовый ответ
	aiResp = parseTextResponse(content)
	return &aiResp, nil
}

func parseTextResponse(text string) entity.AiResponse {
	// Упрощенный парсинг текстового ответа
	resp := entity.AiResponse{
		Reasoning: text,
		Action:    "analyze",
	}

	text = strings.ToLower(text)

	switch {
	case strings.Contains(text, "click") || strings.Contains(text, "press"):
		resp.Action = "click"
	case strings.Contains(text, "type") || strings.Contains(text, "enter"):
		resp.Action = "type"
	case strings.Contains(text, "navigate") || strings.Contains(text, "go to"):
		resp.Action = "navigate"
	case strings.Contains(text, "complete") || strings.Contains(text, "finished"):
		resp.Action = "complete"
		resp.Completed = true
	}

	return resp
}
