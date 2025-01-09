package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	jobctrl "raggo/src/infrastructure/job"
)

type TranslationRequest struct {
	TextID         string `json:"textId" binding:"required"`
	SourceLanguage string `json:"sourceLanguage" binding:"required"`
	TargetLanguage string `json:"targetLanguage" binding:"required"`
	Country        string `json:"country" binding:"required"`
	ModelProvider  string `json:"modelProvider" binding:"required,oneof=ollama"`
	Model          string `json:"model" binding:"required"`
}

type TranslationHandler struct {
	jobService *jobctrl.JobService
}

func NewTranslationHandler(jobService *jobctrl.JobService) (*TranslationHandler, error) {
	return &TranslationHandler{
		jobService: jobService,
	}, nil
}

func (h *TranslationHandler) Translate(c *gin.Context) {
	var req TranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create translation job payload
	payload := jobctrl.TranslationPayload{
		SourceLanguage:   req.SourceLanguage,
		TargetLanguage:   req.TargetLanguage,
		Country:          req.Country,
		TargetResourceID: req.TextID,
		UseService:       req.ModelProvider,
		UseModel:         req.Model,
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal job payload"})
		return
	}

	// Create and enqueue job
	job, err := h.jobService.EnqueueJob(c.Request.Context(), jobctrl.TaskTypeTranslation, payloadBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue translation job"})
		return
	}

	// Return job response
	c.JSON(http.StatusAccepted, gin.H{
		"jobId":   strconv.Itoa(job.ID),
		"status":  "accepted",
		"message": "Translation job created successfully",
	})
}
