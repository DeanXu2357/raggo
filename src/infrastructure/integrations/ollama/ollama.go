package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"raggo/src/infrastructure/log"
)

const (
	DefaultURL = "http://localhost:11434/api"
)

// EmbeddingRequest represents the request structure for embeddings
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the response structure from embeddings
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// TokenRequest represents the request structure for token counting
type TokenRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// TokenResponse represents the response structure from token counting
type TokenResponse int

// GenerateRequest represents the request structure for model generation
type GenerateRequest struct {
	Model   string                 `json:"model"`
	System  string                 `json:"system"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// ErrTruncated is returned when the response was truncated
type ErrTruncated struct {
	Message string
}

func (e *ErrTruncated) Error() string {
	return e.Message
}

// GenerateResponse represents the response structure from generation
type GenerateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Truncated bool   `json:"truncated,omitempty"`
}

// Client represents an Ollama API client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Ollama API client
func NewClient(baseURL string, c *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultURL
	}

	return &Client{
		httpClient: c,
		baseURL:    baseURL,
	}
}

// CountTokens counts the number of tokens in the given prompt
func (c *Client) CountTokens(ctx context.Context, model, prompt string) (int, error) {
	return len(prompt), nil
}

// GetEmbedding generates an embedding vector for the given text using the specified model
func (c *Client) GetEmbedding(ctx context.Context, model string, text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Model:  model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Convert float64 to float32
	embedding32 := make([]float32, len(result.Embedding))
	for i, v := range result.Embedding {
		embedding32[i] = float32(v)
	}

	return embedding32, nil
}

// GenerateContextDescription generates a concise description of the given text content
func (c *Client) GenerateContextDescription(ctx context.Context, text string) (string, error) {
	system := "You are a text summarizer. Generate a brief, informative description of the given text content."
	description, err := c.Generate(ctx, "llama2", system, text, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate context description: %w", err)
	}

	return description, nil
}

// Generate performs model generation with the given prompt
func (c *Client) Generate(ctx context.Context, model, system, prompt string, options map[string]interface{}) (string, error) {
	reqBody := GenerateRequest{
		Model:   model,
		System:  system,
		Prompt:  prompt,
		Stream:  true,
		Options: options,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/generate", c.baseURL)
	//log.Debug("sending request to ollama",
	//	"url", url,
	//	"model", model,
	//	"options", options,
	//	"prompt_length", len(prompt))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error(err, "failed to make request to ollama")
		return "", fmt.Errorf("error making request: %w", err)
	}

	//log.Debug("received response from ollama",
	//	"status", resp.Status,
	//	"content_length", resp.ContentLength)
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var fullResponse strings.Builder
	var lastResponse string

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if lastResponse != "" {
					return lastResponse, nil
				}
				break
			}
			return "", fmt.Errorf("error reading response: %w", err)
		}

		if len(line) == 0 {
			continue
		}

		//log.Debug("received raw response line", "line", string(line))

		var response GenerateResponse
		if err := json.Unmarshal(line, &response); err != nil {
			log.Error(err, "failed to unmarshal response line", "line", string(line))
			return "", fmt.Errorf("error unmarshaling response: %w", err)
		}

		//log.Debug("received response chunk", "response", response.Response, "done", response.Done)
		fullResponse.WriteString(response.Response)

		if response.Truncated {
			log.Error(fmt.Errorf("response was truncated by the model"), "response was truncated by the model")
			return "", &ErrTruncated{Message: "Response was truncated by the model"}
		}

		if response.Done {
			lastResponse = fullResponse.String()
			//log.Debug("completed response", "full_response", lastResponse)
			if lastResponse != "" {
				return lastResponse, nil
			}
		}
	}

	return "", fmt.Errorf("no response received from Ollama")
}
