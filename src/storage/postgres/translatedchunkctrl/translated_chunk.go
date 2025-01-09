package translatedchunkctrl

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type TranslatedChunk struct {
	ID                   int64     `gorm:"primaryKey" json:"id"`
	TranslatedResourceID int64     `gorm:"not null" json:"translated_resource_id"`
	OriginalChunkID      int64     `gorm:"not null" json:"original_chunk_id"`
	ChunkID              string    `gorm:"not null" json:"chunk_id"`
	MinioURL             string    `gorm:"not null;column:minio_url" json:"minio_url"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type TranslatedChunkService struct {
	db        *gorm.DB
	snowflake *snowflake.Node
}

func NewTranslatedChunkService(db *gorm.DB) (*TranslatedChunkService, error) {
	// Initialize snowflake node
	node, err := snowflake.NewNode(4) // Node number 4 for translated chunks
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %v", err)
	}

	return &TranslatedChunkService{
		db:        db,
		snowflake: node,
	}, nil
}

func (s *TranslatedChunkService) Create(ctx context.Context, translatedResourceID, originalChunkID int64, chunkID, minioURL string) (*TranslatedChunk, error) {
	chunk := &TranslatedChunk{
		ID:                   s.snowflake.Generate().Int64(),
		TranslatedResourceID: translatedResourceID,
		OriginalChunkID:      originalChunkID,
		ChunkID:              chunkID,
		MinioURL:             minioURL,
	}

	result := s.db.WithContext(ctx).Create(chunk)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create translated chunk: %v", result.Error)
	}

	return chunk, nil
}

func (s *TranslatedChunkService) GetByTranslatedResourceID(ctx context.Context, translatedResourceID int64) ([]TranslatedChunk, error) {
	var chunks []TranslatedChunk
	result := s.db.WithContext(ctx).Where("translated_resource_id = ?", translatedResourceID).Find(&chunks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get translated chunks: %v", result.Error)
	}
	return chunks, nil
}

func (s *TranslatedChunkService) DeleteByTranslatedResourceID(ctx context.Context, translatedResourceID int64) error {
	result := s.db.WithContext(ctx).Where("translated_resource_id = ?", translatedResourceID).Delete(&TranslatedChunk{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete translated chunks: %v", result.Error)
	}
	return nil
}

func (s *TranslatedChunkService) DeleteByOriginalChunkIDs(ctx context.Context, originalChunkIDs []int64) error {
	if len(originalChunkIDs) == 0 {
		return nil
	}

	result := s.db.WithContext(ctx).Where("original_chunk_id IN ?", originalChunkIDs).Delete(&TranslatedChunk{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete translated chunks: %v", result.Error)
	}
	return nil
}
