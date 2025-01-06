package knowledgebasectrl

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"

	"raggo/src/minioctrl"
	"raggo/src/ollama"
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
}

// Service coordinates operations between PostgreSQL, Weaviate, Ollama and MinIO
type Service struct {
	postgresRepo   PostgresRepository
	snowflake      *snowflake.Node
	weaviateClient *weaviate.Client // Make it a pointer so it can be nil
	ollamaClient   *ollama.Client
	minioService   *minioctrl.MinioService
}

func NewService(postgresRepo PostgresRepository, weaviateClient *weaviate.Client, ollamaClient *ollama.Client, minioService *minioctrl.MinioService) (*Service, error) {
	// Initialize snowflake node
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %v", err)
	}

	return &Service{
		postgresRepo:   postgresRepo,
		snowflake:      node,
		weaviateClient: weaviateClient,
		ollamaClient:   ollamaClient,
		minioService:   minioService,
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

// AddResourceToKnowledgeBase implements the business logic for adding a resource to a knowledge base
func (s *Service) AddResourceToKnowledgeBase(ctx context.Context, knowledgeBaseID int64, resourceID int64) error {
	// TODO: Implementation will include:
	// 1. Get resource metadata from PostgreSQL
	// resource, err := s.resourceRepo.GetResource(ctx, resourceID)
	// if err != nil { return err }

	// 2. Get content from MinIO
	// content, err := s.minioService.GetObject(ctx, resource.Bucket, resource.ObjectKey)
	// if err != nil { return err }

	// 3. Generate embeddings and context description using Ollama
	// embedding, err := s.ollamaClient.GetEmbedding(ctx, string(content))
	// if err != nil { return err }
	// description, err := s.ollamaClient.GenerateContextDescription(ctx, string(content))
	// if err != nil { return err }

	// 4. Store in PostgreSQL and Weaviate
	// kbResource := &KnowledgeBaseResource{
	//   ID: s.snowflake.Generate().Int64(),
	//   KnowledgeBaseID: knowledgeBaseID,
	//   ResourceID: resourceID,
	//   Title: resource.Title,
	//   ContextDescription: description,
	// }
	// err = s.postgresRepo.AddResource(ctx, kbResource)
	// if err != nil { return err }
	// err = s.weaviateClient.AddVector(ctx, kbResource.ID, embedding)
	// if err != nil { return err }

	return fmt.Errorf("not implemented")
}

// QueryKnowledgeBase implements the RAG query logic
func (s *Service) QueryKnowledgeBase(ctx context.Context, knowledgeBaseID int64, query string) (string, error) {
	// TODO: Implementation will include:
	// 1. Get query embedding
	// embedding, err := s.ollamaClient.GetEmbedding(ctx, query)
	// if err != nil { return "", err }

	// 2. Search similar vectors in Weaviate
	// results, err := s.weaviateClient.SearchSimilar(ctx, embedding, 5)
	// if err != nil { return "", err }

	// 3. Get resource content from MinIO
	// var contexts []string
	// for _, result := range results {
	//   resource, err := s.postgresRepo.GetKnowledgeBaseResource(ctx, result.ID)
	//   if err != nil { continue }
	//   content, err := s.minioService.GetObject(ctx, resource.Bucket, resource.ObjectKey)
	//   if err != nil { continue }
	//   contexts = append(contexts, string(content))
	// }

	// 4. Use Ollama to generate final response
	// response, err := s.ollamaClient.GenerateResponse(ctx, query, contexts)
	// if err != nil { return "", err }

	return "", fmt.Errorf("not implemented")
}
