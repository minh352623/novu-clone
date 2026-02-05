package http

import (
	"net/http"

	"CONVERDA/internal/r2/service"
	"CONVERDA/internal/r2/service/dto"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
)

type R2Handler struct {
	service service.R2Service
}

func NewR2Handler(service service.R2Service) *R2Handler {
	return &R2Handler{service: service}
}

// UploadFileBase64 godoc
// @Summary Upload files using base64
// @Description Upload multiple files to R2 storage using base64 encoded strings
// @Tags R2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UploadRequest true "Upload Request"
// @Success 200 {object} dto.UploadResponse
// @Failure 400 {object} response.APIError
// @Failure 401 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /r2/upload [post]
func (h *R2Handler) UploadFileBase64(ctx *gin.Context) (res interface{}, err error) {
	var req dto.UploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request", err.Error())
	}

	if len(req.Images) == 0 {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request", "No images provided")
	}

	uploadResponse, err := h.service.UploadFileBase64(ctx, req.Images)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Failed to upload files", err.Error())
	}

	return uploadResponse, nil
}

// DeleteFile godoc
// @Summary Delete a file from R2
// @Description Delete a file from R2 storage by key
// @Tags R2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key query string true "File key to delete"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} response.APIError
// @Failure 401 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /r2/delete [delete]
func (h *R2Handler) DeleteFile(ctx *gin.Context) (res interface{}, err error) {
	key := ctx.Query("key")
	if key == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request", "File key is required")
	}

	err = h.service.DeleteFile(ctx, key)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, "Failed to delete file", err.Error())
	}

	return gin.H{"message": "File deleted successfully"}, nil
}
