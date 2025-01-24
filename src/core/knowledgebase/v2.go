package knowledgebase

import (
	"context"
	"errors"
	"time"
)

var (
	ErrKnowledgeBaseNotFound = errors.New("Knowledge base not found")
	ErrResourceNotFound      = errors.New("Resource not found")
	ErrResourceLimitExceeded = errors.New("Resource limit exceeded")
	ErrInvalidRequest        = errors.New("Invalid request")
)

// KnowledgeBaseService defines the interface for knowledge base operations
type KnowledgeBaseService interface {
	List(ctx context.Context) ([]KnowledgeBaseDetail, error)
	Create(ctx context.Context, model *KnowledgeBaseDetail) error
	Delete(ctx context.Context, id string) error
	GetSummary(ctx context.Context, id string) (string, error)
}

// ResourceService defines the interface for resource operations
type ResourceService interface {
	List(ctx context.Context, knowledgeBaseID string) ([]Resource, error)
	Create(ctx context.Context, knowledgeBaseID string, file []byte, filename string) (*Resource, error)
	Delete(ctx context.Context, knowledgeBaseID, resourceID string) error
}

// SearchService defines the interface for search operations
type SearchService interface {
	Search(ctx context.Context, knowledgeBaseID string, resourceIDs []string, query string) ([]SearchResultChunk, error)
}

// ChatService defines the interface for chat operations
type ChatService interface {
	GenerateCompletion(ctx context.Context, knowledgeBaseID string, resourceIDs []string, messages []ChatMessage) (*ChatMessage, error)
	GetHistory(ctx context.Context, sessionID string) ([]ChatMessage, error)
}

// SystemService defines the interface for system operations
type SystemService interface {
	CheckHealth(ctx context.Context) (*HealthStatus, error)
}

// KnowledgeBaseDetail represents a knowledge base
type KnowledgeBaseDetail struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	EmbeddingModel string    `json:"embeddingModel"`
	ReasoningModel string    `json:"reasoningModel"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Resource represents a document resource
type Resource struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

// ChatMessage represents a message in chat history
type ChatMessage struct {
	SessionID string    `json:"sessionId"`
	MessageID string    `json:"messageId"`
	Content   string    `json:"content"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// SearchResultChunk represents a single chunk in search results
type SearchResultChunk struct {
	Content    string  `json:"content"`
	Summary    string  `json:"summary"`
	Score      float64 `json:"score"`
	ResourceID string  `json:"resourceId"`
}

// ComponentStatus represents the status of system components
type ComponentStatus string

const (
	StatusUp   ComponentStatus = "up"
	StatusDown ComponentStatus = "down"
)

// HealthStatus represents system health status
type HealthStatus struct {
	Status     string `json:"status"`
	Components struct {
		Valkey   ComponentStatus `json:"valkey"`
		Weaviate ComponentStatus `json:"weaviate"`
		Ollama   ComponentStatus `json:"ollama"`
	} `json:"components"`
}
