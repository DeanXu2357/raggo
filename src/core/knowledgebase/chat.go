package knowledgebase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type chatService struct {
	store       MetadataStore
	searchSvc   SearchService
	llmProvider LLMProvider
}

func NewChatService(store MetadataStore, searchSvc SearchService, llmProvider LLMProvider) ChatService {
	return &chatService{
		store:       store,
		searchSvc:   searchSvc,
		llmProvider: llmProvider,
	}
}

func (s *chatService) GenerateCompletion(ctx context.Context, knowledgeBaseID string, resourceIDs []string, messages []ChatMessage) (*ChatMessage, error) {
	// Get the knowledge base details for model information
	kbDetail, err := s.store.GetKnowledgeBase(ctx, knowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge base details: %w", err)
	}

	// Generate new message ID
	messageID := uuid.New().String()

	// Get relevant context
	lastMessage := messages[len(messages)-1]
	contextChunks, err := s.searchSvc.Search(ctx, knowledgeBaseID, resourceIDs, lastMessage.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	// Format context and messages for the model
	prompt := formatPrompt(messages, contextChunks)

	// Generate completion using LLM
	completion, err := s.llmProvider.Generate(ctx, kbDetail.ReasoningModel, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate completion: %w", err)
	}

	// Create response message
	response := &ChatMessage{
		SessionID: lastMessage.SessionID,
		MessageID: messageID,
		Content:   completion,
		Role:      "assistant",
		CreatedAt: time.Now().UTC(),
	}

	// Store message in history
	if err := s.store.SaveChatMessage(ctx, response); err != nil {
		return nil, fmt.Errorf("failed to save chat message: %w", err)
	}

	return response, nil
}

func (s *chatService) GetHistory(ctx context.Context, sessionID string) ([]ChatMessage, error) {
	return s.store.ListChatMessages(ctx, sessionID)
}

// Helper function to format the prompt with context and history
func formatPrompt(messages []ChatMessage, contextChunks []SearchResultChunk) string {
	// TODO: Implement context and message formatting
	return ""
}
