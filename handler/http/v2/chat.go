package v2

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"raggo/src/core/knowledgebase"
)

type generateCompletionRequest struct {
	Messages        []knowledgebase.ChatMessage `json:"messages" binding:"required,min=1"`
	KnowledgeBaseID string                      `json:"knowledgeBaseId" binding:"required"`
	ResourceIDs     []string                    `json:"resourceIds" binding:"required"`
}

// GenerateCompletion godoc
// @Summary Generate chat completion
// @Tags chat
// @Accept json
// @Produce json
// @Param body body generateCompletionRequest true "Completion parameters"
// @Success 200 {object} knowledgebase.ChatMessage
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chat/completions [post]
func (h *Handler) GenerateCompletion(c *gin.Context) {
	var req generateCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sendError(c, http.StatusBadRequest, err)
		return
	}

	// Validate that the last message is from the user
	lastMsg := req.Messages[len(req.Messages)-1]
	if lastMsg.Role != "user" {
		sendError(c, http.StatusBadRequest, fmt.Errorf("last message must be from user"))
		return
	}

	completion, err := h.chatService.GenerateCompletion(
		c.Request.Context(),
		req.KnowledgeBaseID,
		req.ResourceIDs,
		req.Messages,
	)
	if err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}

	sendJSON(c, http.StatusOK, completion)
}

// GetChatHistory godoc
// @Summary Get chat history
// @Tags chat
// @Param sessionId query string true "Chat session ID"
// @Produce json
// @Success 200 {array} knowledgebase.ChatMessage
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chat/history [get]
func (h *Handler) GetChatHistory(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		sendError(c, http.StatusBadRequest, fmt.Errorf("sessionId is required"))
		return
	}

	history, err := h.chatService.GetHistory(c.Request.Context(), sessionID)
	if err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}

	sendJSON(c, http.StatusOK, history)
}
