package valkey

import (
	"context"
	"fmt"
	"time"

	vk "github.com/valkey-io/valkey-go"
	"raggo/src/core/knowledgebase"
)

const (
	kbPrefix   = "kb:"   // Knowledge base key prefix
	chatPrefix = "chat:" // Chat message key prefix
)

type valkeyStore struct {
	client *vk.Client
}

// NewValkeyStore creates a new MetadataStore instance backed by valkey
func NewValkeyStore(config vk.Config) (knowledgebase.MetadataStore, error) {
	client, err := vk.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey client: %w", err)
	}

	return &valkeyStore{
		client: client,
	}, nil
}

// ListKnowledgeBases returns all knowledge bases
func (s *valkeyStore) ListKnowledgeBases(ctx context.Context) ([]knowledgebase.KnowledgeBaseDetail, error) {
	// Get all keys with kb prefix
	keys, err := s.client.Keys().Pattern(kbPrefix + "*").Do(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list knowledge base keys: %w", err)
	}

	result := make([]knowledgebase.KnowledgeBaseDetail, 0, len(keys))
	for _, key := range keys {
		kb, err := s.getKnowledgeBaseFromHash(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get knowledge base data for key %s: %w", key, err)
		}
		result = append(result, *kb)
	}

	return result, nil
}

// GetKnowledgeBase returns a specific knowledge base by ID
func (s *valkeyStore) GetKnowledgeBase(ctx context.Context, id string) (*knowledgebase.KnowledgeBaseDetail, error) {
	return s.getKnowledgeBaseFromHash(ctx, kbPrefix+id)
}

// SaveKnowledgeBase stores a knowledge base as a hash
func (s *valkeyStore) SaveKnowledgeBase(ctx context.Context, detail *knowledgebase.KnowledgeBaseDetail) error {
	key := kbPrefix + detail.ID

	// Store each field in the hash
	err := s.client.HSet().
		Key(key).
		Field("id", detail.ID).
		Field("name", detail.Name).
		Field("embedding_model", detail.EmbeddingModel).
		Field("reasoning_model", detail.ReasoningModel).
		Field("created_at", detail.CreatedAt.Format(time.RFC3339)).
		Do(ctx).Error()

	if err != nil {
		return fmt.Errorf("failed to save knowledge base: %w", err)
	}

	return nil
}

// DeleteKnowledgeBase removes a knowledge base
func (s *valkeyStore) DeleteKnowledgeBase(ctx context.Context, id string) error {
	err := s.client.Delete().Key(kbPrefix + id).Do(ctx).Error()
	if err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}
	return nil
}

// SaveChatMessage stores a chat message as a hash
func (s *valkeyStore) SaveChatMessage(ctx context.Context, msg *knowledgebase.ChatMessage) error {
	key := fmt.Sprintf("%s%s:%s", chatPrefix, msg.SessionID, msg.MessageID)

	err := s.client.HSet().
		Key(key).
		Field("session_id", msg.SessionID).
		Field("message_id", msg.MessageID).
		Field("content", msg.Content).
		Field("role", msg.Role).
		Field("created_at", msg.CreatedAt.Format(time.RFC3339)).
		Do(ctx).Error()

	if err != nil {
		return fmt.Errorf("failed to save chat message: %w", err)
	}

	return nil
}

// ListChatMessages returns all messages for a given session
func (s *valkeyStore) ListChatMessages(ctx context.Context, sessionID string) ([]knowledgebase.ChatMessage, error) {
	pattern := fmt.Sprintf("%s%s:*", chatPrefix, sessionID)
	keys, err := s.client.Keys().Pattern(pattern).Do(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list chat message keys: %w", err)
	}

	messages := make([]knowledgebase.ChatMessage, 0, len(keys))
	for _, key := range keys {
		msg, err := s.getChatMessageFromHash(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat message data for key %s: %w", key, err)
		}
		messages = append(messages, *msg)
	}

	return messages, nil
}

// Helper function to get knowledge base from hash
func (s *valkeyStore) getKnowledgeBaseFromHash(ctx context.Context, key string) (*knowledgebase.KnowledgeBaseDetail, error) {
	cmd := s.client.HGetAll().Key(key)
	fields, err := cmd.Do(ctx).Result()
	if err != nil {
		if err == vk.Nil {
			return nil, knowledgebase.ErrKnowledgeBaseNotFound
		}
		return nil, fmt.Errorf("failed to get knowledge base fields: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, fields["created_at"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return &knowledgebase.KnowledgeBaseDetail{
		ID:             fields["id"],
		Name:           fields["name"],
		EmbeddingModel: fields["embedding_model"],
		ReasoningModel: fields["reasoning_model"],
		CreatedAt:      createdAt,
	}, nil
}

// Helper function to get chat message from hash
func (s *valkeyStore) getChatMessageFromHash(ctx context.Context, key string) (*knowledgebase.ChatMessage, error) {
	cmd := s.client.HGetAll().Key(key)
	fields, err := cmd.Do(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get chat message fields: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, fields["created_at"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return &knowledgebase.ChatMessage{
		SessionID: fields["session_id"],
		MessageID: fields["message_id"],
		Content:   fields["content"],
		Role:      fields["role"],
		CreatedAt: createdAt,
	}, nil
}
