package translatedresourcectrl

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type TranslatedResource struct {
	ID                 int64     `gorm:"primaryKey" json:"id"`
	OriginalResourceID int64     `gorm:"not null" json:"original_resource_id"`
	Filename           string    `gorm:"not null" json:"filename"`
	MinioURL           string    `gorm:"not null;column:minio_url" json:"minio_url"`
	SourceLanguage     string    `gorm:"not null" json:"source_language"`
	TargetLanguage     string    `gorm:"not null" json:"target_language"`
	Country            string    `gorm:"not null" json:"country"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type TranslatedResourceService struct {
	db        *gorm.DB
	snowflake *snowflake.Node
}

func NewTranslatedResourceService(db *gorm.DB) (*TranslatedResourceService, error) {
	// Initialize snowflake node
	node, err := snowflake.NewNode(3) // Node number 3 for translated resources
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %v", err)
	}

	return &TranslatedResourceService{
		db:        db,
		snowflake: node,
	}, nil
}

func (s *TranslatedResourceService) Create(ctx context.Context, originalResourceID int64, filename, minioURL, sourceLang, targetLang, country string) (*TranslatedResource, error) {
	resource := &TranslatedResource{
		ID:                 s.snowflake.Generate().Int64(),
		OriginalResourceID: originalResourceID,
		Filename:           filename,
		MinioURL:           minioURL,
		SourceLanguage:     sourceLang,
		TargetLanguage:     targetLang,
		Country:            country,
	}

	result := s.db.WithContext(ctx).Create(resource)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create translated resource: %v", result.Error)
	}

	return resource, nil
}

func (s *TranslatedResourceService) GetByOriginalID(ctx context.Context, originalID int64) ([]TranslatedResource, error) {
	var resources []TranslatedResource
	result := s.db.WithContext(ctx).Where("original_resource_id = ?", originalID).Find(&resources)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get translated resources: %v", result.Error)
	}
	return resources, nil
}

func (s *TranslatedResourceService) DeleteByOriginalID(ctx context.Context, originalID int64) error {
	result := s.db.WithContext(ctx).Where("original_resource_id = ?", originalID).Delete(&TranslatedResource{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete translated resources: %v", result.Error)
	}
	return nil
}
