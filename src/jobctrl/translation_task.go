package jobctrl

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"raggo/src/chunkctrl"
	"raggo/src/minioctrl"
	"raggo/src/ollama"
	"raggo/src/postgres/chunkctrl"
	"raggo/src/postgres/resourcectrl"
	"raggo/src/postgres/translatedchunkctrl"
	"raggo/src/postgres/translatedresourcectrl"
	"raggo/src/translationflow"
)

const TaskTypeTranslation = "translation"

type TranslationPayload struct {
	SourceLanguage   string `json:"source_language"`
	TargetLanguage   string `json:"target_language"`
	Country          string `json:"country"`
	TargetResourceID string `json:"target_resource_id"`
	UseService       string `json:"use_service"`
	UseModel         string `json:"use_model"`
}

type TranslationTask struct {
	resourceService       *resourcectrl.ResourceService
	chunkService          *chunkctrl.ChunkService
	translatedResourceSvc *translatedresourcectrl.TranslatedResourceService
	translatedChunkSvc    *translatedchunkctrl.TranslatedChunkService
	minioService          *minioctrl.MinioService
	ollamaClient          *ollama.Client
}

func NewTranslationTask(
	resourceService *resourcectrl.ResourceService,
	chunkService *chunkctrl.ChunkService,
	translatedResourceSvc *translatedresourcectrl.TranslatedResourceService,
	translatedChunkSvc *translatedchunkctrl.TranslatedChunkService,
	minioService *minioctrl.MinioService,
	ollamaClient *ollama.Client,
) *TranslationTask {
	return &TranslationTask{
		resourceService:       resourceService,
		chunkService:          chunkService,
		translatedResourceSvc: translatedResourceSvc,
		translatedChunkSvc:    translatedChunkSvc,
		minioService:          minioService,
		ollamaClient:          ollamaClient,
	}
}

func (task *TranslationTask) HandleTranslationTask(ctx context.Context, payload json.RawMessage) error {
	// decode payload
	var translationPayload TranslationPayload
	if err := json.Unmarshal(payload, &translationPayload); err != nil {
		return fmt.Errorf("failed to unmarshal translation payload: %w", err)
	}

	// find resource
	resourceID, err := strconv.ParseInt(translationPayload.TargetResourceID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid resource ID: %w", err)
	}
	resource, err := task.resourceService.GetByID(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}
	if resource == nil {
		return fmt.Errorf("resource not found: %s", translationPayload.TargetResourceID)
	}

	// Ensure minio buckets exist
	if err := task.minioService.EnsureBucketExists(ctx, minioctrl.TranslatedResourcesBucket); err != nil {
		return fmt.Errorf("failed to ensure translated resources bucket exists: %w", err)
	}
	if err := task.minioService.EnsureBucketExists(ctx, minioctrl.TranslatedChunksBucket); err != nil {
		return fmt.Errorf("failed to ensure translated chunks bucket exists: %w", err)
	}

	// Clean up existing translations
	if err := task.cleanupExistingTranslations(ctx, resourceID); err != nil {
		return fmt.Errorf("failed to cleanup existing translations: %w", err)
	}

	// Create translated resource
	translatedResource, err := task.translatedResourceSvc.Create(
		ctx,
		resource.ID,
		fmt.Sprintf("%s_translated_%s", resource.Filename, translationPayload.TargetLanguage),
		fmt.Sprintf("%s/%s_translated_%s", minioctrl.TranslatedResourcesBucket, resource.MinioURL, translationPayload.TargetLanguage),
		translationPayload.SourceLanguage,
		translationPayload.TargetLanguage,
		translationPayload.Country,
	)
	if err != nil {
		return fmt.Errorf("failed to create translated resource: %w", err)
	}

	// find chunks
	chunks, err := task.chunkService.GetByResourceID(ctx, resource.ID)
	if err != nil {
		return fmt.Errorf("failed to get chunks: %w", err)
	}

	// Create translation flow with ollama provider
	provider := ollama.NewOllamaProvider(task.ollamaClient, translationPayload.UseModel)
	translationFlow := translationflow.NewTranslationFlow(provider)

	// Translate each chunk and collect translations
	var allTranslations []string
	for _, chunk := range chunks {
		// Get chunk content from MinioURL
		bucket, objectName := task.minioService.GetBucketAndObjectFromURL(chunk.MinioURL)
		chunkContent, err := task.minioService.GetObject(ctx, bucket, objectName)
		if err != nil {
			return fmt.Errorf("failed to get chunk content: %w", err)
		}

		// Translate chunk
		translatedContent, err := translationFlow.Translate(
			ctx,
			string(chunkContent),
			translationPayload.SourceLanguage,
			translationPayload.TargetLanguage,
			translationPayload.Country,
		)
		if err != nil {
			return fmt.Errorf("failed to translate chunk %s: %w", chunk.ChunkID, err)
		}

		// Save translated chunk to minio
		translatedObjectName := fmt.Sprintf("%s_translated_%s", objectName, translationPayload.TargetLanguage)
		if err := task.minioService.PutObject(ctx, minioctrl.TranslatedChunksBucket, translatedObjectName, []byte(translatedContent)); err != nil {
			return fmt.Errorf("failed to save translated chunk content: %w", err)
		}

		// Create translated chunk record
		translatedMinioURL := fmt.Sprintf("%s/%s", minioctrl.TranslatedChunksBucket, translatedObjectName)
		_, err = task.translatedChunkSvc.Create(
			ctx,
			translatedResource.ID,
			chunk.ID,
			chunk.ChunkID+"_translated",
			translatedMinioURL,
		)
		if err != nil {
			return fmt.Errorf("failed to save translated chunk record: %w", err)
		}

		// Collect translation for complete document
		allTranslations = append(allTranslations, translatedContent)
	}

	// Combine all translations and save to translated resource
	completeTranslation := strings.Join(allTranslations, "\n")
	bucket, objectName := task.minioService.GetBucketAndObjectFromURL(translatedResource.MinioURL)
	if err := task.minioService.PutObject(ctx, bucket, objectName, []byte(completeTranslation)); err != nil {
		return fmt.Errorf("failed to save complete translated content: %w", err)
	}

	return nil
}

func (task *TranslationTask) cleanupExistingTranslations(ctx context.Context, resourceID int64) error {
	// Get existing translated resources
	translatedResources, err := task.translatedResourceSvc.GetByOriginalID(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get existing translated resources: %w", err)
	}

	for _, tr := range translatedResources {
		// Get translated chunks to clean up
		translatedChunks, err := task.translatedChunkSvc.GetByTranslatedResourceID(ctx, tr.ID)
		if err != nil {
			return fmt.Errorf("failed to get translated chunks: %w", err)
		}

		// Delete translated chunks from minio
		for _, tc := range translatedChunks {
			bucket, objectName := task.minioService.GetBucketAndObjectFromURL(tc.MinioURL)
			if err := task.minioService.DeleteObject(ctx, bucket, objectName); err != nil {
				return fmt.Errorf("failed to delete translated chunk object: %w", err)
			}
		}

		// Delete translated chunks records
		if err := task.translatedChunkSvc.DeleteByTranslatedResourceID(ctx, tr.ID); err != nil {
			return fmt.Errorf("failed to delete translated chunks records: %w", err)
		}

		// Delete translated resource from minio
		bucket, objectName := task.minioService.GetBucketAndObjectFromURL(tr.MinioURL)
		if err := task.minioService.DeleteObject(ctx, bucket, objectName); err != nil {
			return fmt.Errorf("failed to delete translated resource object: %w", err)
		}
	}

	// Delete translated resources records
	if err := task.translatedResourceSvc.DeleteByOriginalID(ctx, resourceID); err != nil {
		return fmt.Errorf("failed to delete translated resources records: %w", err)
	}

	return nil
}
