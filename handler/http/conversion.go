package http

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"

	"raggo/src/chunkctrl"
	"raggo/src/resourcectrl"
	"raggo/src/unstructuredctrl"
)

type ConversionHandler struct {
	minioClient         *minio.Client
	resourceBucket      string
	chunkBucket         string
	minioDomain         string
	resourceService     *resourcectrl.ResourceService
	chunkService        *chunkctrl.ChunkService
	unstructuredService *unstructuredctrl.UnstructuredService
}

func NewConversionHandler(
	minioClient *minio.Client,
	resourceBucket string,
	chunkBucket string,
	minioDomain string,
	unstructuredURL string,
	resourceService *resourcectrl.ResourceService,
	chunkService *chunkctrl.ChunkService,
) (*ConversionHandler, error) {
	// Ensure chunk bucket exists
	exists, err := minioClient.BucketExists(context.Background(), chunkBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(context.Background(), chunkBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	unstructuredService := unstructuredctrl.NewUnstructuredService(unstructuredURL)

	return &ConversionHandler{
		minioClient:         minioClient,
		resourceBucket:      resourceBucket,
		chunkBucket:         chunkBucket,
		minioDomain:         minioDomain,
		resourceService:     resourceService,
		chunkService:        chunkService,
		unstructuredService: unstructuredService,
	}, nil
}

type ConversionRequest struct {
	PDFID string `json:"pdfId"`
}

func (h *ConversionHandler) Convert(c *gin.Context) {
	var req ConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Parse ID from string to int64
	var pdfID int64
	_, err := fmt.Sscanf(req.PDFID, "%d", &pdfID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PDF ID format"})
		return
	}

	// Get resource from database
	resource, err := h.resourceService.GetByID(c.Request.Context(), pdfID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource"})
		return
	}
	if resource == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PDF not found"})
		return
	}

	// Check for existing chunks and delete them if found
	existingChunks, err := h.chunkService.GetByResourceID(c.Request.Context(), pdfID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing chunks"})
		return
	}

	if len(existingChunks) > 0 {
		// Delete chunks from MinIO
		for _, chunk := range existingChunks {
			parts := strings.Split(chunk.MinioURL, "/")
			if len(parts) != 2 {
				log.Printf("Invalid MinIO URL format for chunk: %s", chunk.MinioURL)
				continue
			}
			err = h.minioClient.RemoveObject(c.Request.Context(), parts[0], parts[1], minio.RemoveObjectOptions{})
			if err != nil {
				log.Printf("Failed to delete chunk from MinIO: %v", err)
			}
		}

		// Delete chunks from database
		err = h.chunkService.DeleteByResourceID(c.Request.Context(), pdfID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing chunks"})
			return
		}
	}

	// Parse MinIO URL to get bucket and object name
	parts := strings.Split(resource.MinioURL, "/")
	if len(parts) != 2 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid MinIO URL format"})
		return
	}
	objectName := parts[1]

	// Get file from MinIO
	obj, err := h.minioClient.GetObject(c.Request.Context(), h.resourceBucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file from storage"})
		return
	}
	defer obj.Close()

	// Read file into buffer
	fileBytes, err := io.ReadAll(obj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Call unstructured service
	elements, err := h.unstructuredService.ConvertPDFToText(resource.Filename, fileBytes)
	if err != nil {
		log.Printf("Failed to convert PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert PDF"})
		return
	}

	// Store chunks
	var chunks []chunkctrl.Chunk
	for i, element := range elements {
		if element.Text == "" {
			continue
		}

		// Generate chunk filename
		chunkID := fmt.Sprintf("chunk_%d", i+1)
		chunkName := fmt.Sprintf("%s_%s.txt", strings.TrimSuffix(filepath.Base(resource.Filename), filepath.Ext(resource.Filename)), chunkID)

		// Upload chunk to MinIO
		_, err = h.minioClient.PutObject(
			c.Request.Context(),
			h.chunkBucket,
			chunkName,
			strings.NewReader(element.Text),
			int64(len(element.Text)),
			minio.PutObjectOptions{ContentType: "text/plain"},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store chunk"})
			return
		}

		// Create chunk record
		chunk, err := h.chunkService.Create(
			c.Request.Context(),
			resource.ID,
			chunkID,
			fmt.Sprintf("%s/%s", h.chunkBucket, chunkName),
		)
		if err != nil {
			log.Printf("Failed to record chunk: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record chunk"})
			return
		}
		chunks = append(chunks, *chunk)
	}

	c.JSON(http.StatusAccepted, gin.H{
		"jobId":   fmt.Sprintf("conv_%d", resource.ID),
		"status":  "completed",
		"message": fmt.Sprintf("Successfully converted PDF into %d chunks", len(chunks)),
	})
}
