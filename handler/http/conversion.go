package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"raggo/src/infrastructure/integrations/unstructured"
	"raggo/src/storage/minioctrl"
	"raggo/src/storage/postgres/chunkctrl"
	"raggo/src/storage/postgres/resourcectrl"
)

type ConversionHandler struct {
	minioService        *minioctrl.MinioService
	resourceBucket      string
	chunkBucket         string
	minioDomain         string
	resourceService     *resourcectrl.ResourceService
	chunkService        *chunkctrl.ChunkService
	unstructuredService *unstructured.UnstructuredService
}

func NewConversionHandler(
	minioService *minioctrl.MinioService,
	resourceBucket string,
	chunkBucket string,
	minioDomain string,
	unstructuredURL string,
	resourceService *resourcectrl.ResourceService,
	chunkService *chunkctrl.ChunkService,
) (*ConversionHandler, error) {
	// Ensure chunk bucket exists
	err := minioService.EnsureBucketExists(context.Background(), chunkBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %v", err)
	}

	unstructuredService := unstructured.NewUnstructuredService(unstructuredURL)

	return &ConversionHandler{
		minioService:        minioService,
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
			err = h.minioService.DeleteObject(c.Request.Context(), parts[0], parts[1])
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
	fileBytes, err := h.minioService.GetObject(c.Request.Context(), h.resourceBucket, objectName)
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
		err = h.minioService.PutObject(
			c.Request.Context(),
			h.chunkBucket,
			chunkName,
			[]byte(element.Text),
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
			i+1, // Use the loop index + 1 as the order
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
