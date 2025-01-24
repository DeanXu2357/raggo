package knowledgebase

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/weaviate/weaviate/entities/models"

	"raggo/src/fsutil"
	"raggo/src/infrastructure/integrations/ollama"
	"raggo/src/storage/weaviate"
)

type LocalKnowledgeBase struct {
	dataRoot     string           // RAG_DATA_ROOT path
	store        MetadataStore    // Store for knowledge base metadata
	fs           fsutil.FileStore // File system operations
	weaviateSDK  *weaviate.SDK
	ollamaClient *ollama.Client
}

type LocalResourceService struct {
	kb *LocalKnowledgeBase
}

type LocalSearchService struct {
	kb *LocalKnowledgeBase
}

type LocalChatService struct {
	kb *LocalKnowledgeBase
}

type LocalSystemService struct {
	kb *LocalKnowledgeBase
}

// NewV2Service creates a new instance of the v2 knowledge base service
func NewV2Service(dataRoot string, store MetadataStore, weaviateSDK *weaviate.SDK, ollamaClient *ollama.Client) (*LocalKnowledgeBase, error) {
	kb := &LocalKnowledgeBase{
		dataRoot:     dataRoot,
		store:        store,
		fs:           fsutil.NewLocalFileStore(),
		weaviateSDK:  weaviateSDK,
		ollamaClient: ollamaClient,
	}

	// Validate required dependencies
	if err := kb.validateDependencies(); err != nil {
		return nil, fmt.Errorf("failed to validate dependencies: %w", err)
	}

	return kb, nil
}

func (kb *LocalKnowledgeBase) validateDependencies() error {
	if kb.dataRoot == "" {
		return fmt.Errorf("data root path is required")
	}
	if kb.store == nil {
		return fmt.Errorf("metadata store is required")
	}
	if kb.fs == nil {
		return fmt.Errorf("file store is required")
	}
	if kb.weaviateSDK == nil {
		return fmt.Errorf("weaviate SDK is required")
	}
	if kb.ollamaClient == nil {
		return fmt.Errorf("ollama client is required")
	}
	return nil
}

func (kb *LocalKnowledgeBase) initializeWeaviateClass(ctx context.Context, className string) error {
	// Define schema properties for the Weaviate class
	properties := []*models.Property{
		{
			Name:        "content",
			DataType:    []string{"text"},
			Description: "The content of the chunk",
		},
		{
			Name:        "summary",
			DataType:    []string{"text"},
			Description: "Contextual summary of the chunk",
		},
		{
			Name:        "resourceId",
			DataType:    []string{"text"},
			Description: "ID of the source resource",
		},
		{
			Name:        "order",
			DataType:    []string{"int"},
			Description: "Order of the chunk within the resource",
		},
	}

	return kb.weaviateSDK.CreateSchema(ctx, className, properties, "none")
}

func (kb *LocalKnowledgeBase) deleteWeaviateClass(ctx context.Context, className string) error {
	return kb.weaviateSDK.DeleteSchema(ctx, className)
}

// List implements KnowledgeBaseService
func (kb *LocalKnowledgeBase) List(ctx context.Context) ([]KnowledgeBaseDetail, error) {
	return kb.store.ListKnowledgeBases(ctx)
}

// Create implements KnowledgeBaseService
func (kb *LocalKnowledgeBase) Create(ctx context.Context, model *KnowledgeBaseDetail) error {
	// Create knowledge base directory structure
	kbPath := filepath.Join(kb.dataRoot, model.ID)
	resourcePath := filepath.Join(kbPath, "resources")
	if err := kb.fs.MakeDirectory(resourcePath); err != nil {
		return fmt.Errorf("failed to create knowledge base directories: %w", err)
	}

	// Store knowledge base metadata
	if err := kb.store.SaveKnowledgeBase(ctx, model); err != nil {
		return fmt.Errorf("failed to save knowledge base metadata: %w", err)
	}

	// Initialize Weaviate class
	className := fmt.Sprintf("KnowledgeBase_%s", model.ID)
	if err := kb.initializeWeaviateClass(ctx, className); err != nil {
		return fmt.Errorf("failed to initialize weaviate class: %w", err)
	}

	return nil
}

// Delete implements KnowledgeBaseService
func (kb *LocalKnowledgeBase) Delete(ctx context.Context, id string) error {
	// Delete directories
	kbPath := filepath.Join(kb.dataRoot, id)
	if err := kb.fs.RemoveAll(kbPath); err != nil {
		return fmt.Errorf("failed to delete knowledge base directory: %w", err)
	}

	// Delete metadata
	if err := kb.store.DeleteKnowledgeBase(ctx, id); err != nil {
		return fmt.Errorf("failed to delete knowledge base metadata: %w", err)
	}

	// Delete Weaviate class
	className := fmt.Sprintf("KnowledgeBase_%s", id)
	if err := kb.deleteWeaviateClass(ctx, className); err != nil {
		return fmt.Errorf("failed to delete weaviate class: %w", err)
	}

	return nil
}

// GetSummary implements KnowledgeBaseService
func (kb *LocalKnowledgeBase) GetSummary(ctx context.Context, id string) (string, error) {
	// Get knowledge base details
	kbDetail, err := kb.store.GetKnowledgeBase(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get knowledge base details: %w", err)
	}

	// Get resource count and total size
	resourcePath := filepath.Join(kb.dataRoot, id, "resources")
	resourceCount, totalSize, err := kb.fs.GetFileStats(resourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to get resource stats: %w", err)
	}

	// Get vector count from Weaviate
	className := fmt.Sprintf("KnowledgeBase_%s", id)
	vectorCount, err := kb.getVectorCount(ctx, className)
	if err != nil {
		return "", fmt.Errorf("failed to get vector count: %w", err)
	}

	summary := fmt.Sprintf("Knowledge Base: %s\nResources: %d\nTotal Size: %d bytes\nVectors: %d\nEmbedding Model: %s\nReasoning Model: %s",
		kbDetail.Name, resourceCount, totalSize, vectorCount, kbDetail.EmbeddingModel, kbDetail.ReasoningModel)
	return summary, nil
}

// Helper function for Weaviate operations
func (kb *LocalKnowledgeBase) getVectorCount(ctx context.Context, className string) (int, error) {
	// Create a query to count all objects in the class
	result, err := kb.weaviateSDK.QueryVectors(ctx, className, nil, weaviate.QueryConfig{
		Limit: 0, // We only need the count
	})
	if err != nil {
		return 0, fmt.Errorf("failed to query vector count: %w", err)
	}

	return len(result), nil
}

// Helper functions for Weaviate operations
func (kb *LocalKnowledgeBase) getWeaviateClassName(id string) string {
	return fmt.Sprintf("KnowledgeBase_%s", id)
}
