package knowledgebase

import (
	"context"
)

// LLMProvider defines operations for language model interactions
type LLMProvider interface {
	// GetEmbedding generates embeddings for the given input text
	GetEmbedding(ctx context.Context, model string, input string) ([]float32, error)
	// Generate generates text completion using the specified model
	Generate(ctx context.Context, model string, prompt string) (string, error)
}

// VectorStore defines operations for vector storage and search
type VectorStore interface {
	// QueryVectors performs a vector similarity search
	QueryVectors(ctx context.Context, className string, vector []float32, config QueryConfig) ([]QueryResult, error)
	// QueryHybrid performs a hybrid search combining vector similarity and keyword matching
	QueryHybrid(ctx context.Context, className string, vector []float32, config HybridConfig) ([]QueryResult, error)
}

// QueryConfig defines parameters for vector queries
type QueryConfig struct {
	Fields    []string
	Limit     int
	Certainty float32
}

// HybridConfig defines parameters for hybrid search
type HybridConfig struct {
	Query string
	Alpha float32
	Limit int
}

// QueryResult represents a search result with properties and score
type QueryResult struct {
	Score      float64
	Properties map[string]interface{}
}
