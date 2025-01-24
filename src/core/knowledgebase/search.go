package knowledgebase

import (
	"context"
	"fmt"

	"raggo/src/infrastructure/integrations/ollama"
	"raggo/src/storage/weaviate"
)

type searchService struct {
	weaviateSDK  *weaviate.SDK
	ollamaClient *ollama.Client
}

func NewSearchService(weaviateSDK *weaviate.SDK, ollamaClient *ollama.Client) SearchService {
	return &searchService{
		weaviateSDK:  weaviateSDK,
		ollamaClient: ollamaClient,
	}
}

func (s *searchService) Search(ctx context.Context, knowledgeBaseID string, resourceIDs []string, query string) ([]SearchResultChunk, error) {
	// Return empty slice if no resource IDs are provided
	if len(resourceIDs) == 0 {
		return []SearchResultChunk{}, nil
	}

	// Get query embedding
	className := fmt.Sprintf("KnowledgeBase_%s", knowledgeBaseID)

	// Configure search parameters for vector search
	config := weaviate.QueryConfig{
		Limit: 20, // Maximum results as per specification
	}

	// Get embedding for query
	embedding, err := s.ollamaClient.GetEmbedding(ctx, "nomic-embed-text", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Perform vector search
	results, err := s.weaviateSDK.QueryVectors(ctx, className, embedding, config)
	if err != nil {
		return nil, fmt.Errorf("failed to search weaviate: %w", err)
	}

	return s.processSearchResults(results, resourceIDs), nil
}

// SearchHybrid performs hybrid search using both vector similarity and BM25
func (s *searchService) SearchHybrid(ctx context.Context, knowledgeBaseID string, resourceIDs []string, query string) ([]SearchResultChunk, error) {
	// Return empty slice if no resource IDs are provided
	if len(resourceIDs) == 0 {
		return []SearchResultChunk{}, nil
	}

	className := fmt.Sprintf("KnowledgeBase_%s", knowledgeBaseID)

	// Get embedding for query
	embedding, err := s.ollamaClient.GetEmbedding(ctx, "nomic-embed-text", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Configure hybrid search
	config := weaviate.HybridConfig{
		Query: query,
		Alpha: 0.75, // 75% vector search, 25% BM25
		Limit: 20,   // Maximum results as per specification
	}

	// Perform hybrid search
	results, err := s.weaviateSDK.QueryHybrid(ctx, className, embedding, config)
	if err != nil {
		return nil, fmt.Errorf("failed to search weaviate: %w", err)
	}

	return s.processSearchResults(results, resourceIDs), nil
}

// processSearchResults converts weaviate results to SearchResultChunks and filters by resourceIDs
func (s *searchService) processSearchResults(results []weaviate.QueryResult, resourceIDs []string) []SearchResultChunk {
	chunks := make([]SearchResultChunk, 0, len(results))
	for _, result := range results {
		// Only include results from specified resources
		for _, id := range resourceIDs {
			if result.Properties["resourceId"] == id {
				chunk := SearchResultChunk{
					Content:    result.Properties["content"].(string),
					Summary:    result.Properties["summary"].(string),
					Score:      result.Score,
					ResourceID: result.Properties["resourceId"].(string),
				}
				if start, ok := result.Properties["start"].(float64); ok {
					chunk.Location.Start = int(start)
				}
				if end, ok := result.Properties["end"].(float64); ok {
					chunk.Location.End = int(end)
				}
				chunks = append(chunks, chunk)
				break
			}
		}
	}

	return chunks
}
