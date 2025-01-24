package knowledgebase

import (
	"context"
)

// MetadataStore defines the interface for storing knowledge base metadata and chat history
type MetadataStore interface {
	// Knowledge Base operations
	ListKnowledgeBases(ctx context.Context) ([]KnowledgeBaseDetail, error)
	GetKnowledgeBase(ctx context.Context, id string) (*KnowledgeBaseDetail, error)
	SaveKnowledgeBase(ctx context.Context, detail *KnowledgeBaseDetail) error
	DeleteKnowledgeBase(ctx context.Context, id string) error

	// Chat history operations
	SaveChatMessage(ctx context.Context, msg *ChatMessage) error
	ListChatMessages(ctx context.Context, sessionID string) ([]ChatMessage, error)
}
