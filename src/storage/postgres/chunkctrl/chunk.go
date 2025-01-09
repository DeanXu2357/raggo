package chunkctrl

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type Chunk struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	ResourceID int64     `gorm:"not null" json:"resource_id"`
	ChunkID    string    `gorm:"not null" json:"chunk_id"`
	MinioURL   string    `gorm:"not null;column:minio_url" json:"minio_url"` // bucket name + object name
	Order      int       `gorm:"not null;column:chunk_order" json:"order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ChunkService struct {
	db        *gorm.DB
	snowflake *snowflake.Node
}

func NewChunkService(db *gorm.DB) (*ChunkService, error) {
	// Initialize snowflake node
	node, err := snowflake.NewNode(2) // Node number 2 for chunks
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %v", err)
	}

	return &ChunkService{
		db:        db,
		snowflake: node,
	}, nil
}

func (s *ChunkService) Create(ctx context.Context, resourceID int64, chunkID string, minioURL string, order int) (*Chunk, error) {
	chunk := &Chunk{
		ID:         s.snowflake.Generate().Int64(),
		ResourceID: resourceID,
		ChunkID:    chunkID,
		MinioURL:   minioURL,
		Order:      order,
	}

	result := s.db.WithContext(ctx).Create(chunk)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create chunk: %v", result.Error)
	}

	return chunk, nil
}

func (s *ChunkService) GetByResourceID(ctx context.Context, resourceID int64) ([]Chunk, error) {
	var chunks []Chunk
	result := s.db.WithContext(ctx).Where("resource_id = ?", resourceID).Find(&chunks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get chunks: %v", result.Error)
	}
	return chunks, nil
}

func (s *ChunkService) DeleteByResourceID(ctx context.Context, resourceID int64) error {
	result := s.db.WithContext(ctx).Where("resource_id = ?", resourceID).Delete(&Chunk{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete chunks: %v", result.Error)
	}
	return nil
}

func (s *ChunkService) GetByID(ctx context.Context, id int64) (*Chunk, error) {
	var chunk Chunk
	result := s.db.WithContext(ctx).First(&chunk, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chunk: %v", result.Error)
	}
	return &chunk, nil
}
