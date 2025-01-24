package weaviate

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

// HybridConfig contains configuration for hybrid search
type HybridConfig struct {
	Query     string  // Text query for BM25
	Alpha     float32 // Weight for vector search (default: 0.75)
	Fields    []string
	Limit     int
	Distance  float64
	Certainty float64
}

// DefaultHybridConfig returns default configuration for hybrid search
func DefaultHybridConfig(query string) HybridConfig {
	return HybridConfig{
		Query: query,
		Alpha: 0.75, // 75% vector search, 25% BM25
		Limit: DefaultQueryLimit,
	}
}

// QueryHybrid performs hybrid search combining vector similarity and BM25
func (w *SDK) QueryHybrid(ctx context.Context, className string, vector []float32, config HybridConfig) ([]QueryResult, error) {
	// Convert string fields to GraphQL fields
	fields := make([]graphql.Field, len(config.Fields))
	for i, field := range config.Fields {
		fields[i] = graphql.Field{Name: field}
	}
	// Add _additional field for metadata
	fields = append(fields, graphql.Field{Name: "_additional { id distance certainty score }"})

	// Build hybrid search arguments
	hybridBuilder := w.client.GraphQL().HybridArgumentBuilder().
		WithVector(vector).
		WithQuery(config.Query).
		WithAlpha(config.Alpha)

	if config.Distance > 0 {
		hybridBuilder.WithDistance(float32(config.Distance))
	}
	if config.Certainty > 0 {
		hybridBuilder.WithCertainty(float32(config.Certainty))
	}

	limit := config.Limit
	if limit <= 0 {
		limit = DefaultQueryLimit
	}

	// Execute query
	result, err := w.client.GraphQL().Get().
		WithClassName(className).
		WithFields(fields...).
		WithHybrid(hybridBuilder).
		WithLimit(limit).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %v", err)
	}

	// Parse results
	var queryResults []QueryResult
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if objects, ok := data[className].([]interface{}); ok {
			for _, obj := range objects {
				if objMap, ok := obj.(map[string]interface{}); ok {
					additional := objMap["_additional"].(map[string]interface{})

					// Create properties map excluding _additional
					properties := make(map[string]interface{})
					for k, v := range objMap {
						if k != "_additional" {
							properties[k] = v
						}
					}

					queryResults = append(queryResults, QueryResult{
						ID:         additional["id"].(string),
						Score:      additional["score"].(float64), // Use hybrid score instead of distance
						Properties: properties,
					})
				}
			}
		}
	}

	return queryResults, nil
}
