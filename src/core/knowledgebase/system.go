package knowledgebase

import (
	"context"

	"raggo/src/storage/weaviate"
)

type systemService struct {
	kb *LocalKnowledgeBase
}

func NewSystemService(kb *LocalKnowledgeBase) SystemService {
	return &systemService{
		kb: kb,
	}
}

func (s *systemService) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Status: "healthy",
		Components: struct {
			Valkey   ComponentStatus `json:"valkey"`
			Weaviate ComponentStatus `json:"weaviate"`
			Ollama   ComponentStatus `json:"ollama"`
		}{
			Valkey:   StatusDown,
			Weaviate: StatusDown,
			Ollama:   StatusDown,
		},
	}

	// Check Valkey
	if _, err := s.kb.store.ListKnowledgeBases(ctx); err == nil {
		status.Components.Valkey = StatusUp
	}

	// Check Weaviate
	if _, err := s.kb.weaviateSDK.QueryVectors(ctx, "TestClass", nil, weaviate.QueryConfig{Limit: 1}); err == nil {
		status.Components.Weaviate = StatusUp
	}

	// Check Ollama
	if _, err := s.kb.ollamaClient.Models(ctx); err == nil {
		status.Components.Ollama = StatusUp
	}

	// If any component is down, mark system as unhealthy
	if status.Components.Valkey == StatusDown ||
		status.Components.Weaviate == StatusDown ||
		status.Components.Ollama == StatusDown {
		status.Status = "unhealthy"
	}

	return status, nil
}
