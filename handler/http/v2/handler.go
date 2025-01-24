package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"raggo/src/core/knowledgebase"
)

type Handler struct {
	kbService     knowledgebase.KnowledgeBaseService
	resService    knowledgebase.ResourceService
	searchService knowledgebase.SearchService
	chatService   knowledgebase.ChatService
	sysService    knowledgebase.SystemService
}

func NewHandler(kbService knowledgebase.KnowledgeBaseService, resService knowledgebase.ResourceService, searchService knowledgebase.SearchService, chatService knowledgebase.ChatService, sysService knowledgebase.SystemService) *Handler {
	return &Handler{
		kbService:     kbService,
		resService:    resService,
		searchService: searchService,
		chatService:   chatService,
		sysService:    sysService,
	}
}

// RegisterRoutes registers all v2 API routes
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	v2 := r.Group("/api/v1")

	// Knowledge base routes
	v2.GET("/knowledge-bases", h.ListKnowledgeBases)
	v2.POST("/knowledge-bases", h.CreateKnowledgeBase)
	v2.DELETE("/knowledge-bases/:id", h.DeleteKnowledgeBase)
	v2.GET("/knowledge-bases/:id/summary", h.GetKnowledgeBaseSummary)

	// Resource routes
	v2.GET("/knowledge-bases/:id/resources", h.ListResources)
	v2.POST("/knowledge-bases/:id/resources", h.CreateResource)
	v2.DELETE("/knowledge-bases/:id/resources/:resourceId", h.DeleteResource)

	// Search routes
	v2.POST("/knowledge-bases/:id/search", h.Search)

	// Chat routes
	v2.POST("/chat/completions", h.GenerateCompletion)
	v2.GET("/chat/history", h.GetChatHistory)

	// System routes
	v2.GET("/health", h.CheckHealth)
}

// Common error response structure
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func sendError(c *gin.Context, status int, err error) {
	var code string
	switch {
	case err == knowledgebase.ErrKnowledgeBaseNotFound:
		code = "NOT_FOUND"
		status = http.StatusNotFound
	case err == knowledgebase.ErrResourceLimitExceeded:
		code = "RESOURCE_LIMIT_EXCEEDED"
		status = http.StatusBadRequest
	default:
		code = "INTERNAL_ERROR"
		status = http.StatusInternalServerError
	}

	c.JSON(status, ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

func sendJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}
