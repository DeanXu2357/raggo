package knowledgebasectrl

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	kb "raggo/src/knowledgebasectrl"
)

type KnowledgeBase struct {
	ID             int64  `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	Description    string
	EmbeddingModel string `gorm:"not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type KnowledgeBaseResource struct {
	ID                 int64  `gorm:"primaryKey"`
	KnowledgeBaseID    int64  `gorm:"not null"`
	ChunkID            int64  `gorm:"not null"`
	Title              string `gorm:"not null"`
	ContextDescription string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListKnowledgeBases(ctx context.Context, offset, limit int) ([]kb.KnowledgeBase, error) {
	var bases []KnowledgeBase
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&bases)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list knowledge bases: %v", result.Error)
	}

	// Convert to domain model
	var domainBases []kb.KnowledgeBase
	for _, base := range bases {
		domainBases = append(domainBases, kb.KnowledgeBase{
			ID:             base.ID,
			Name:           base.Name,
			Description:    base.Description,
			EmbeddingModel: base.EmbeddingModel,
			CreatedAt:      base.CreatedAt,
			UpdatedAt:      base.UpdatedAt,
		})
	}

	return domainBases, nil
}

func (r *Repository) ListKnowledgeBaseResources(ctx context.Context, knowledgeBaseID int64, offset, limit int) ([]kb.KnowledgeBaseResource, error) {
	var resources []KnowledgeBaseResource
	result := r.db.WithContext(ctx).
		Where("knowledge_base_id = ?", knowledgeBaseID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list knowledge base resources: %v", result.Error)
	}

	// Convert to domain model
	var domainResources []kb.KnowledgeBaseResource
	for _, resource := range resources {
		domainResources = append(domainResources, kb.KnowledgeBaseResource{
			ID:                 resource.ID,
			KnowledgeBaseID:    resource.KnowledgeBaseID,
			ChunkID:            resource.ChunkID,
			Title:              resource.Title,
			ContextDescription: resource.ContextDescription,
			CreatedAt:          resource.CreatedAt,
			UpdatedAt:          resource.UpdatedAt,
		})
	}

	return domainResources, nil
}

func (r *Repository) CreateKnowledgeBase(ctx context.Context, kb *kb.KnowledgeBase) error {
	base := KnowledgeBase{
		ID:             kb.ID,
		Name:           kb.Name,
		Description:    kb.Description,
		EmbeddingModel: kb.EmbeddingModel,
	}

	result := r.db.WithContext(ctx).Create(&base)
	if result.Error != nil {
		return fmt.Errorf("failed to create knowledge base: %v", result.Error)
	}

	return nil
}

func (r *Repository) AddResource(ctx context.Context, resource *kb.KnowledgeBaseResource) error {
	dbResource := KnowledgeBaseResource{
		ID:                 resource.ID,
		KnowledgeBaseID:    resource.KnowledgeBaseID,
		ChunkID:            resource.ChunkID,
		Title:              resource.Title,
		ContextDescription: resource.ContextDescription,
	}

	result := r.db.WithContext(ctx).Create(&dbResource)
	if result.Error != nil {
		return fmt.Errorf("failed to add resource to knowledge base: %v", result.Error)
	}

	return nil
}
