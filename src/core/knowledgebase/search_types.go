package knowledgebase

import "context"

// SearchService handles knowledge base search operations
type SearchService interface {
	// Search performs vector similarity search
	Search(ctx context.Context, knowledgeBaseID string, resourceIDs []string, query string) ([]SearchResultChunk, error)

	// SearchHybrid performs hybrid search combining vector similarity and BM25
	SearchHybrid(ctx context.Context, knowledgeBaseID string, resourceIDs []string, query string) ([]SearchResultChunk, error)
}

// SearchResultChunk represents a single chunk result from search
type SearchResultChunk struct {
	Content    string  `json:"content"`
	Summary    string  `json:"summary,omitempty"`
	Score      float64 `json:"score"`
	ResourceID string  `json:"resourceId"`
	Location   struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"location,omitempty"`
}
