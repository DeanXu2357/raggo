package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"raggo/src/core/knowledgebase"
)

// CheckHealth godoc
// @Summary Check system health status
// @Tags system
// @Produce json
// @Success 200 {object} knowledgebase.HealthStatus
// @Failure 500 {object} ErrorResponse
// @Router /health [get]
func (h *Handler) CheckHealth(c *gin.Context) {
	status, err := h.sysService.CheckHealth(c.Request.Context())
	if err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}
	sendJSON(c, http.StatusOK, status)
}
