package http

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type PDFHandler struct {
	minioClient *minio.Client
	bucketName  string
	minioDomain string
}

func NewPDFHandler(minioClient *minio.Client, bucketName string, minioDomain string) (*PDFHandler, error) {
	// Ensure bucket exists
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	return &PDFHandler{
		minioClient: minioClient,
		bucketName:  bucketName,
		minioDomain: minioDomain,
	}, nil
}

func (h *PDFHandler) Upload(c *gin.Context) {
	// Get file from request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Validate file type
	if filepath.Ext(header.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
		return
	}

	// Generate unique file name
	id := uuid.New().String()
	objectName := fmt.Sprintf("%s.pdf", id)

	// Upload to MinIO
	_, err = h.minioClient.PutObject(
		context.Background(),
		h.bucketName,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: "application/pdf"},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	// Generate file URL using injected domain
	url := fmt.Sprintf("%s/%s/%s", h.minioDomain, h.bucketName, objectName)

	// Return response according to OpenAPI spec
	c.JSON(http.StatusCreated, gin.H{
		"id":       id,
		"filename": header.Filename,
		"url":      url,
	})
}
