package resourcectrl

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type Resource struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Filename  string    `gorm:"not null" json:"filename"`
	MinioURL  string    `gorm:"not null;column:minio_url" json:"minio_url"` // bucket name + object name
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ResourceService struct {
	db        *gorm.DB
	snowflake *snowflake.Node
}

func NewResourceService(db *gorm.DB) (*ResourceService, error) {
	// Initialize snowflake node
	node, err := snowflake.NewNode(1) // Node number 1
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %v", err)
	}

	return &ResourceService{
		db:        db,
		snowflake: node,
	}, nil
}

func (s *ResourceService) GetByID(ctx context.Context, id int64) (*Resource, error) {
	var resource Resource
	result := s.db.WithContext(ctx).First(&resource, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get resource: %v", result.Error)
	}
	return &resource, nil
}

// List returns a paginated list of resources
func (s *ResourceService) List(ctx context.Context, limit int, offset int) ([]Resource, error) {
	var resources []Resource

	result := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&resources)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list resources: %v", result.Error)
	}

	return resources, nil
}

func (s *ResourceService) Create(ctx context.Context, filename, minioURL string) (*Resource, error) {
	resource := &Resource{
		ID:       s.snowflake.Generate().Int64(),
		Filename: filename,
		MinioURL: minioURL,
	}

	result := s.db.WithContext(ctx).Create(resource)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create resource: %v", result.Error)
	}

	return resource, nil
}
