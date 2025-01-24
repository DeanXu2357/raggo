package v2

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"raggo/src/core/knowledgebase"
)

type searchRequest struct {
	Query       string   `json:"query" binding:"required"`
	ResourceIDs []string `json:"resourceIds"`
	UseHybrid   bool     `json:"useHybrid"` // Whether to use hybrid search
	TimeRange   struct {
		Start *time.Time `json:"start,omitempty"`
		End   *time.Time `json:"end,omitempty"`
	} `json:"timeRange,omitempty"`
}

// Search godoc
// @Summary Search content in a knowledge base
// @Tags search
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID"
// @Param body body searchRequest true "Search parameters"
// @Success 200 {array} knowledgebase.SearchResultChunk
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /knowledge-bases/{id}/search [post]
func (h *Handler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sendError(c, http.StatusBadRequest, err)
		return
	}

	id := c.Param("id")

	// TODO: Apply time range filter to resource IDs
	// This should be done in a future iteration when we support time-based filtering

	// Choose search method based on request
	var results []knowledgebase.SearchResultChunk
	var err error
	if req.UseHybrid {
		results, err = h.searchService.SearchHybrid(c.Request.Context(), id, req.ResourceIDs, req.Query)
	} else {
		results, err = h.searchService.Search(c.Request.Context(), id, req.ResourceIDs, req.Query)
	}
	if err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}

	sendJSON(c, http.StatusOK, results)
}
