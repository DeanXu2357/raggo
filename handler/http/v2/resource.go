package v2

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"raggo/src/core/knowledgebase"
)

// ListResources godoc
// @Summary List resources in a knowledge base
// @Tags resources
// @Param id path string true "Knowledge base ID"
// @Produce json
// @Success 200 {array} knowledgebase.Resource
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /knowledge-bases/{id}/resources [get]
func (h *Handler) ListResources(c *gin.Context) {
	id := c.Param("id")
	resources, err := h.resService.List(c.Request.Context(), id)
	if err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}
	sendJSON(c, http.StatusOK, resources)
}

// CreateResource godoc
// @Summary Add a resource to a knowledge base
// @Tags resources
// @Accept multipart/form-data
// @Param id path string true "Knowledge base ID"
// @Param file formData file true "Resource file"
// @Produce json
// @Success 201 {object} knowledgebase.Resource
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /knowledge-bases/{id}/resources [post]
func (h *Handler) CreateResource(c *gin.Context) {
	id := c.Param("id")

	// Get file from form data
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		sendError(c, http.StatusBadRequest, fmt.Errorf("file upload required: %w", err))
		return
	}
	defer file.Close()

	// Read file contents
	data, err := io.ReadAll(file)
	if err != nil {
		sendError(c, http.StatusInternalServerError, fmt.Errorf("failed to read file: %w", err))
		return
	}

	// Create resource
	resource, err := h.resService.Create(c.Request.Context(), id, data, header.Filename)
	if err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}

	sendJSON(c, http.StatusCreated, resource)
}

// DeleteResource godoc
// @Summary Remove a resource from a knowledge base
// @Tags resources
// @Param id path string true "Knowledge base ID"
// @Param resourceId path string true "Resource ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /knowledge-bases/{id}/resources/{resourceId} [delete]
func (h *Handler) DeleteResource(c *gin.Context) {
	kbID := c.Param("id")
	resourceID := c.Param("resourceId")

	if err := h.resService.Delete(c.Request.Context(), kbID, resourceID); err != nil {
		sendError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}
