package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ai_assistant/internal/service"

	"github.com/spf13/viper"
)

// OpenAIEmbedding implements EmbeddingService using OpenAI API.
type OpenAIEmbedding struct {
	apiKey  string
	model   string
	url     string
	dim     int
	client  *http.Client
}

// NewOpenAIEmbedding creates a new OpenAI embedding client.
func NewOpenAIEmbedding(cfg *viper.Viper) *OpenAIEmbedding {
	return &OpenAIEmbedding{
		apiKey: cfg.GetString("openai.api_key"),
		model:  cfg.GetString("openai.embedding_model"),
		url:    cfg.GetString("openai.embedding_url"),
		dim:    cfg.GetInt("openai.embedding_dimensions"),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// EmbedTexts generates embeddings for a list of texts.
func (c *OpenAIEmbedding) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("openai api key not configured")
	}

	reqBody := map[string]interface{}{
		"model": c.model,
		"input": texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api error: status %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = d.Embedding
	}

	return embeddings, nil
}

// Ensure OpenAIEmbedding implements EmbeddingService
var _ service.EmbeddingService = (*OpenAIEmbedding)(nil)
