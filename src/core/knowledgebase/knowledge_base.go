package knowledgebase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/weaviate/weaviate/entities/models"

	"raggo/src/infrastructure/integrations/ollama"
	"raggo/src/storage/minioctrl"
	"raggo/src/storage/postgres/chunkctrl"
	"raggo/src/storage/postgres/resourcectrl"
	"raggo/src/storage/weaviate"
)

type KnowledgeBase struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	EmbeddingModel string    `json:"embedding_model"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type KnowledgeBaseResource struct {
	ID                 int64     `json:"id"`
	KnowledgeBaseID    int64     `json:"knowledge_base_id"`
	ResourceID         int64     `json:"resource_id"`
	ChunkID            int64     `json:"chunk_id"`
	Title              string    `json:"title"`
	ContextDescription string    `json:"context_description"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// PostgresRepository defines the interface for PostgreSQL operations
type PostgresRepository interface {
	ListKnowledgeBases(ctx context.Context, offset, limit int) ([]KnowledgeBase, error)
	ListKnowledgeBaseResources(ctx context.Context, knowledgeBaseID int64, offset, limit int) ([]KnowledgeBaseResource, error)
	CreateKnowledgeBase(ctx context.Context, kb *KnowledgeBase) error
	AddResource(ctx context.Context, resource *KnowledgeBaseResource) error
	GetKnowledgeBase(ctx context.Context, id int64) (*KnowledgeBase, error)
}

// Service coordinates operations between PostgreSQL, Weaviate, Ollama and MinIO
type Service struct {
	postgresRepo    PostgresRepository
	snowflake       *snowflake.Node
	weaviateSDK     *weaviate.SDK
	ollamaClient    *ollama.Client
	minioService    *minioctrl.MinioService
	resourceService *resourcectrl.ResourceService
	chunkService    *chunkctrl.ChunkService
}

func NewService(postgresRepo PostgresRepository, weaviateSDK *weaviate.SDK, ollamaClient *ollama.Client, minioService *minioctrl.MinioService, resourceService *resourcectrl.ResourceService, chunkService *chunkctrl.ChunkService) (*Service, error) {
	// Initialize snowflake node
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %v", err)
	}

	return &Service{
		postgresRepo:    postgresRepo,
		snowflake:       node,
		weaviateSDK:     weaviateSDK,
		ollamaClient:    ollamaClient,
		minioService:    minioService,
		resourceService: resourceService,
		chunkService:    chunkService,
	}, nil
}

// ListKnowledgeBases returns a paginated list of knowledge bases
func (s *Service) ListKnowledgeBases(ctx context.Context, offset, limit int) ([]KnowledgeBase, error) {
	return s.postgresRepo.ListKnowledgeBases(ctx, offset, limit)
}

// ListKnowledgeBaseResources returns a paginated list of resources in a knowledge base
func (s *Service) ListKnowledgeBaseResources(ctx context.Context, knowledgeBaseID int64, offset, limit int) ([]KnowledgeBaseResource, error) {
	return s.postgresRepo.ListKnowledgeBaseResources(ctx, knowledgeBaseID, offset, limit)
}

// getWeaviateClassName returns the Weaviate class name for a knowledge base
func getWeaviateClassName(knowledgeBaseID int64) string {
	return fmt.Sprintf("KnowledgeBaseResource_%d", knowledgeBaseID)
}

// AddResourceToKnowledgeBase implements the business logic for adding a resource to a knowledge base
func (s *Service) AddResourceToKnowledgeBase(ctx context.Context, knowledgeBaseID int64, resourceID int64) error {
	// Get resource metadata
	resource, err := s.resourceService.GetByID(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get resource: %v", err)
	}
	if resource == nil {
		return fmt.Errorf("resource not found: %d", resourceID)
	}

	// Get chunks for the resource
	chunks, err := s.chunkService.GetByResourceID(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get chunks: %v", err)
	}

	// Get knowledge base to determine embedding model
	kb, err := s.postgresRepo.GetKnowledgeBase(ctx, knowledgeBaseID)
	if err != nil {
		return fmt.Errorf("failed to get knowledge base: %v", err)
	}
	if kb == nil {
		return fmt.Errorf("knowledge base not found: %d", knowledgeBaseID)
	}

	className := getWeaviateClassName(knowledgeBaseID)
	if s.weaviateSDK != nil {
		// Ensure schema exists in Weaviate
		if err = s.ensureWeaviateSchema(ctx, className); err != nil {
			return fmt.Errorf("failed to ensure Weaviate schema: %v", err)
		}
	}

	// Process each chunk
	for _, chunk := range chunks {
		// Get chunk content from MinIO
		// MinioURL format is "bucket/objectKey"
		parts := strings.Split(chunk.MinioURL, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid minio URL format: %s", chunk.MinioURL)
		}
		content, err := s.minioService.GetObject(ctx, parts[0], parts[1])
		if err != nil {
			return fmt.Errorf("failed to get chunk content: %v", err)
		}

		// Generate embeddings and context description using Ollama
		embedding, err := s.ollamaClient.GetEmbedding(ctx, kb.EmbeddingModel, string(content))
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %v", err)
		}

		description, err := s.ollamaClient.GenerateContextDescription(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to generate context description: %v", err)
		}

		// Create knowledge base resource
		kbResource := &KnowledgeBaseResource{
			ID:                 s.snowflake.Generate().Int64(),
			KnowledgeBaseID:    knowledgeBaseID,
			ResourceID:         resourceID,
			ChunkID:            chunk.ID,
			Title:              fmt.Sprintf("%s - Part %d", resource.Filename, chunk.Order),
			ContextDescription: description,
		}

		// Store in PostgreSQL
		if err = s.postgresRepo.AddResource(ctx, kbResource); err != nil {
			return fmt.Errorf("failed to add resource to knowledge base: %v", err)
		}

		// Store vector in Weaviate
		if s.weaviateSDK != nil {
			// Convert properties to map for Weaviate
			props := map[string]interface{}{
				"knowledgeBaseId": kbResource.KnowledgeBaseID,
				"resourceId":      kbResource.ResourceID,
				"chunkId":         kbResource.ChunkID,
				"title":           kbResource.Title,
				"description":     kbResource.ContextDescription,
			}

			// Create object with vector
			vectorObj := weaviate.VectorObject{
				Vector:     embedding,
				Properties: props,
			}

			if err = s.weaviateSDK.AddVector(ctx, className, vectorObj); err != nil {
				return fmt.Errorf("failed to store vector: %v", err)
			}
		}
	}

	return nil
}

// ensureWeaviateSchema ensures the required schema exists in Weaviate
func (s *Service) ensureWeaviateSchema(ctx context.Context, className string) error {
	properties := []*models.Property{
		{
			Name:     "knowledgeBaseId",
			DataType: []string{"int"},
		},
		{
			Name:     "resourceId",
			DataType: []string{"int"},
		},
		{
			Name:     "chunkId",
			DataType: []string{"int"},
		},
		{
			Name:     "title",
			DataType: []string{"string"},
		},
		{
			Name:     "description",
			DataType: []string{"text"},
		},
	}

	err := s.weaviateSDK.CreateSchema(ctx, className, properties, "none")
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}
	return nil
}

// QueryResult represents a single result from the knowledge base query
type QueryResult struct {
	ChunkID     int64   `json:"chunk_id"`
	ResourceID  int64   `json:"resource_id"`
	Score       float64 `json:"score"`
	Content     string  `json:"content"`
	Description string  `json:"description"`
	MinioURL    string  `json:"minio_url"`
}

// QueryKnowledgeBase implements the RAG query logic
func (s *Service) QueryKnowledgeBase(ctx context.Context, knowledgeBaseID int64, query string) ([]QueryResult, error) {
	if s.weaviateSDK == nil {
		return nil, fmt.Errorf("weaviate is not configured")
	}

	// Get knowledge base to determine embedding model
	kb, err := s.postgresRepo.GetKnowledgeBase(ctx, knowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge base: %v", err)
	}
	if kb == nil {
		return nil, fmt.Errorf("knowledge base not found: %d", knowledgeBaseID)
	}

	// Get query embedding
	embedding, err := s.ollamaClient.GetEmbedding(ctx, kb.EmbeddingModel, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %v", err)
	}

	// Search similar vectors in Weaviate
	className := getWeaviateClassName(knowledgeBaseID)
	config := weaviate.QueryConfig{
		Fields:    []string{"chunkId", "description"},
		Limit:     5,
		Certainty: 0.7, // Minimum similarity threshold
	}

	results, err := s.weaviateSDK.QueryVectors(ctx, className, embedding, config)
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %v", err)
	}

	// Get chunk content from MinIO for each result
	var queryResults []QueryResult
	for _, result := range results {
		chunkID := int64(result.Properties["chunkId"].(float64))
		description := result.Properties["description"].(string)

		chunk, err := s.chunkService.GetByID(ctx, chunkID)
		if err != nil {
			continue
		}

		parts := strings.Split(chunk.MinioURL, "/")
		if len(parts) != 2 {
			continue
		}

		content, err := s.minioService.GetObject(ctx, parts[0], parts[1])
		if err != nil {
			continue
		}

		queryResults = append(queryResults, QueryResult{
			ChunkID:     chunkID,
			ResourceID:  chunk.ResourceID,
			Score:       result.Score,
			Content:     string(content),
			Description: description,
			MinioURL:    chunk.MinioURL,
		})
	}

	if len(queryResults) == 0 {
		return nil, fmt.Errorf("no relevant context found")
	}

	return queryResults, nil
}
