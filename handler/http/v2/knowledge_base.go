package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"raggo/src/core/knowledgebase"
)

type createKnowledgeBaseRequest struct {
	Name           string `json:"name" binding:"required"`
	EmbeddingModel string `json:"embeddingModel"`
	ReasoningModel string `json:"reasoningModel"`
}

// ListKnowledgeBases godoc
// @Summary List all knowledge bases
// @Tags knowledge-bases
// @Produce json
// @Success 200 {array} knowledgebase.KnowledgeBaseDetail
// @Failure 500 {object} ErrorResponse
// @Router /knowledge-bases [get]
func (h *Handler) ListKnowledgeBases(c *gin.Context) {
	kbs, err := h.kbService.List(c.Request.Context())
	if err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}
	sendJSON(c, http.StatusOK, kbs)
}

// CreateKnowledgeBase godoc
// @Summary Create a new knowledge base
// @Tags knowledge-bases
// @Accept json
// @Produce json
// @Param body body createKnowledgeBaseRequest true "Knowledge base configuration"
// @Success 201 {object} knowledgebase.KnowledgeBaseDetail
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /knowledge-bases [post]
func (h *Handler) CreateKnowledgeBase(c *gin.Context) {
	var req createKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sendError(c, http.StatusBadRequest, err)
		return
	}

	// Set default models if not provided
	if req.EmbeddingModel == "" {
		req.EmbeddingModel = "nomic-embed-text"
	}
	if req.ReasoningModel == "" {
		req.ReasoningModel = "phi4"
	}

	kb := &knowledgebase.KnowledgeBaseDetail{
		Name:           req.Name,
		EmbeddingModel: req.EmbeddingModel,
		ReasoningModel: req.ReasoningModel,
	}

	if err := h.kbService.Create(c.Request.Context(), kb); err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}

	sendJSON(c, http.StatusCreated, kb)
}

// DeleteKnowledgeBase godoc
// @Summary Delete a knowledge base
// @Tags knowledge-bases
// @Param id path string true "Knowledge base ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /knowledge-bases/{id} [delete]
func (h *Handler) DeleteKnowledgeBase(c *gin.Context) {
	id := c.Param("id")
	if err := h.kbService.Delete(c.Request.Context(), id); err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetKnowledgeBaseSummary godoc
// @Summary Get a knowledge base summary
// @Tags knowledge-bases
// @Param id path string true "Knowledge base ID"
// @Success 200 {string} string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /knowledge-bases/{id}/summary [get]
func (h *Handler) GetKnowledgeBaseSummary(c *gin.Context) {
	id := c.Param("id")
	summary, err := h.kbService.GetSummary(c.Request.Context(), id)
	if err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}
	sendJSON(c, http.StatusOK, gin.H{"summary": summary})
}
