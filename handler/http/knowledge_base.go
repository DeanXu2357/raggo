package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"raggo/src/knowledgebasectrl"
)

type KnowledgeBaseHandler struct {
	service *knowledgebasectrl.Service
}

func NewKnowledgeBaseHandler(service *knowledgebasectrl.Service) (*KnowledgeBaseHandler, error) {
	return &KnowledgeBaseHandler{
		service: service,
	}, nil
}

// ListKnowledgeBases handles GET /api/v1/knowledge-bases
func (h *KnowledgeBaseHandler) ListKnowledgeBases(c *gin.Context) {
	offset, limit := getPaginationParams(c)

	bases, err := h.service.ListKnowledgeBases(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  bases,
		"offset": offset,
		"limit":  limit,
	})
}

// ListKnowledgeBaseResources handles GET /api/v1/knowledge-bases/:id/resources
func (h *KnowledgeBaseHandler) ListKnowledgeBaseResources(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid knowledge base ID"})
		return
	}

	offset, limit := getPaginationParams(c)

	resources, err := h.service.ListKnowledgeBaseResources(c.Request.Context(), id, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  resources,
		"offset": offset,
		"limit":  limit,
	})
}

// QueryKnowledgeBase handles POST /api/v1/knowledge-bases/:id/query
func (h *KnowledgeBaseHandler) QueryKnowledgeBase(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid knowledge base ID"})
		return
	}

	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := h.service.QueryKnowledgeBase(c.Request.Context(), id, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// AddResourceToKnowledgeBase handles POST /api/v1/knowledge-bases/:id/resources
func (h *KnowledgeBaseHandler) AddResourceToKnowledgeBase(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid knowledge base ID"})
		return
	}

	var req struct {
		ResourceID int64 `json:"resource_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.service.AddResourceToKnowledgeBase(c.Request.Context(), id, req.ResourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func getPaginationParams(c *gin.Context) (offset, limit int) {
	offset, _ = strconv.Atoi(c.Query("offset"))
	limit, _ = strconv.Atoi(c.Query("limit"))

	if limit <= 0 {
		limit = 10 // default limit
	}
	if offset < 0 {
		offset = 0
	}

	return offset, limit
}
