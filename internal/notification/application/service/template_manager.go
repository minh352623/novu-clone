package service

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
)

type TemplateManager struct {
	repo       repository.TemplateRepository
	layoutRepo repository.NotificationLayoutRepository
}

func NewTemplateManager(repo repository.TemplateRepository, layoutRepo repository.NotificationLayoutRepository) *TemplateManager {
	return &TemplateManager{
		repo:       repo,
		layoutRepo: layoutRepo,
	}
}

func (tm *TemplateManager) GetTemplate(ctx context.Context, envID uuid.UUID, id uuid.UUID) (*entity.Template, error) {
	return tm.repo.GetByID(ctx, id)
}

func (tm *TemplateManager) GetAndCompile(ctx context.Context, envID uuid.UUID, code, lang string, data map[string]interface{}) (string, string, error) {
	// 1. Fetch Template
	tmpl, err := tm.repo.GetByCode(ctx, envID, code, lang)
	if err != nil {
		return "", "", fmt.Errorf("failed to get template %s: %w", code, err)
	}

	if !tmpl.IsActive {
		return "", "", fmt.Errorf("template %s is inactive", code)
	}

	// 2. Compile Subject
	subject, err := tm.compileString(tmpl.Subject, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to compile subject: %w", err)
	}

	// 3. Compile Body
	body, err := tm.compileString(tmpl.Body, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to compile body: %w", err)
	}

	// 4. Wrap with Layout if applicable
	wrappedBody, err := tm.wrapWithLayout(ctx, envID, tmpl, body, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to wrap with layout: %w", err)
	}

	return subject, wrappedBody, nil
}

func (tm *TemplateManager) wrapWithLayout(ctx context.Context, envID uuid.UUID, tmpl *entity.Template, compiledBody string, data map[string]interface{}) (string, error) {
	var layout *entity.NotificationLayout
	var err error

	if tmpl.LayoutID != nil {
		layout, err = tm.layoutRepo.GetByID(ctx, *tmpl.LayoutID)
		if err != nil {
			return "", err
		}
	} else {
		// Try default layout for environment
		layout, err = tm.layoutRepo.GetDefault(ctx, envID)
		if err != nil {
			return "", err
		}
	}

	if layout == nil || layout.ContentHTML == "" {
		return compiledBody, nil
	}

	// Prepare data for layout, adding Content
	layoutData := make(map[string]interface{})
	for k, v := range data {
		layoutData[k] = v
	}
	layoutData["Content"] = compiledBody

	// Compile Layout
	return tm.compileString(layout.ContentHTML, layoutData)
}

func (tm *TemplateManager) compileString(tmplStr string, data map[string]interface{}) (string, error) {
	t, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
