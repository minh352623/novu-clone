package controller

import (
	"encoding/json"
	"net/http"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/controller/dto"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TemplateContentController struct {
	tmplManager *service.TemplateManager
}

func NewTemplateContentController(tmplManager *service.TemplateManager) *TemplateContentController {
	return &TemplateContentController{tmplManager: tmplManager}
}

// AddContent godoc
// @Summary Add template content for a new language
// @Description Add content (subject, body) in a new language for a template
// @Tags Templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param body body dto.CreateTemplateContentRequest true "Content data"
// @Success 201 {object} dto.TemplateContentResponse
// @Failure 400 {object} map[string]string
// @Router /notifications/templates/{id}/content [post]
func (c *TemplateContentController) AddContent(ctx *gin.Context) (interface{}, error) {
	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid template ID", err)
	}

	var req dto.CreateTemplateContentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	content := entity.NewTemplateContent(
		templateID,
		0, // version will be set by service
		req.LanguageCode,
		req.Subject,
		req.BodyText,
		req.BodyHTML,
		req.BodyPush,
	)

	if err := c.tmplManager.AddContent(ctx.Request.Context(), templateID, content); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return dto.ToTemplateContentResponse(content), nil
}

// UpdateContent godoc
// @Summary Update template content for a language
// @Description Update subject and body for a specific language
// @Tags Templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param lang path string true "Language Code"
// @Param body body dto.UpdateTemplateContentRequest true "Content update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /notifications/templates/{id}/content/{lang} [put]
func (c *TemplateContentController) UpdateContent(ctx *gin.Context) (interface{}, error) {
	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid template ID", err)
	}

	lang := ctx.Param("lang")
	if lang == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Language code is required", nil)
	}

	var req dto.UpdateTemplateContentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	// Build partial entity
	update := &entity.TemplateContent{}
	if req.Subject != nil {
		update.Subject = *req.Subject
	}
	if req.BodyText != nil {
		update.BodyText = *req.BodyText
	}
	if req.BodyHTML != nil {
		update.BodyHTML = *req.BodyHTML
	}
	if req.BodyPush != nil {
		update.BodyPush = json.RawMessage(*req.BodyPush)
	}

	if err := c.tmplManager.UpdateContent(ctx.Request.Context(), templateID, lang, update); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return map[string]string{"status": "updated"}, nil
}

// ListLanguages godoc
// @Summary List available languages for a template
// @Description Returns all language codes with content for a template
// @Tags Templates
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} dto.AvailableLanguagesResponse
// @Failure 400 {object} map[string]string
// @Router /notifications/templates/{id}/languages [get]
func (c *TemplateContentController) ListLanguages(ctx *gin.Context) (interface{}, error) {
	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid template ID", err)
	}

	langs, err := c.tmplManager.ListLanguages(ctx.Request.Context(), templateID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	// Get template to include version in response
	tmpl, err := c.tmplManager.GetTemplate(ctx.Request.Context(), uuid.Nil, templateID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	version := 1
	if tmpl != nil {
		version = tmpl.Version
	}

	return &dto.AvailableLanguagesResponse{
		TemplateID: templateID,
		Version:    version,
		Languages:  langs,
	}, nil
}
