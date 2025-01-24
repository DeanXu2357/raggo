package knowledgebase

import (
	"context"
	"fmt"
	"path/filepath"

	"raggo/src/fsutil"
)

type resourceService struct {
	dataRoot string
	fs       fsutil.FileStore
}

func NewResourceService(dataRoot string, fs fsutil.FileStore) ResourceService {
	return &resourceService{
		dataRoot: dataRoot,
		fs:       fs,
	}
}

func (s *resourceService) List(ctx context.Context, knowledgeBaseID string) ([]Resource, error) {
	// Get the resource directory for the knowledge base
	resourcePath := filepath.Join(s.dataRoot, knowledgeBaseID, "resources")

	// Get all files in the resource directory
	_, _, err := s.fs.GetFileStats(resourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource stats: %w", err)
	}

	// TODO: Convert file stats to Resource objects
	// This will require retrieving metadata from valkey as well
	return nil, fmt.Errorf("not implemented")
}

func (s *resourceService) Create(ctx context.Context, knowledgeBaseID string, file []byte, filename string) (*Resource, error) {
	// TODO:
	// 1. Check file size limits
	// 2. Save file to resources directory
	// 3. Extract text content (if PDF/markdown)
	// 4. Create chunks
	// 5. Store chunks in Weaviate
	// 6. Save metadata in Valkey
	return nil, fmt.Errorf("not implemented")
}

func (s *resourceService) Delete(ctx context.Context, knowledgeBaseID, resourceID string) error {
	// TODO:
	// 1. Delete file from resources directory
	// 2. Delete chunks from Weaviate
	// 3. Delete metadata from Valkey
	return fmt.Errorf("not implemented")
}
