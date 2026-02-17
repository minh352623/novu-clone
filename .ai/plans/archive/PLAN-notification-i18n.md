# PLAN: Notification i18n (Multi-Language Support)

## Problem
The Notification Pipeline's template system has **partial i18n infrastructure** but lacks key pieces for production use:
- DB schema has `language_code` on `notification_template_contents` ✅
- `GetByCode` accepts `lang` and `SendRequest` has `Language` ✅
- **No fallback**: if template content doesn't exist for requested language, returns empty body (bug)
- **No CRUD API**: no way to add/list/update template content in multiple languages
- **No language listing**: no endpoint to see which languages a template supports

### Goal
Complete the i18n support so that:
1. Templates can have content in multiple languages (en, vi, ja, etc.)
2. Sending falls back to default language (`en`) when requested language not found
3. API allows managing template content per language
4. Partners can query available languages per template

---

## Review Level: L2 (Proposal)
- Extends existing infrastructure — no breaking changes
- Follows established DDD patterns

---

## Proposed Changes

### Component 1: Language Fallback in Repository

#### [MODIFY] [template.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/infrastructure/persistence/repository/template.repository.go)

Fix `GetByCode`: when content for requested `language` is not found, auto-fallback to `"en"`:

```go
// Current: returns nil content when language not found → bug
// New: try requested lang → fallback to "en" → return error if both missing
```

---

### Component 2: Template Content CRUD

#### [MODIFY] [template.repository.go (interface)](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/domain/repository/template.repository.go)

Add methods:
```go
CreateContent(ctx, templateID uuid.UUID, content *entity.TemplateContent) error
UpdateContent(ctx, content *entity.TemplateContent) error
ListLanguages(ctx, templateID uuid.UUID) ([]string, error)
GetContent(ctx, templateID uuid.UUID, version int, lang string) (*entity.TemplateContent, error)
```

#### [MODIFY] [template.repository.go (impl)](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/infrastructure/persistence/repository/template.repository.go)

Implement the 4 new methods using GORM.

---

### Component 3: Domain Entity

#### [MODIFY] [Template entity](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/domain/entity)

Add `TemplateContent` entity if not already separated (for independent CRUD):
```go
type TemplateContent struct {
    ID           uuid.UUID
    TemplateID   uuid.UUID
    Version      int
    LanguageCode string
    Subject      string
    BodyText     string
    BodyHTML     string
    BodyPush     json.RawMessage
    CreatedAt    time.Time
}
```

---

### Component 4: DTOs and Controller

#### [NEW] [template_content.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/controller/dto/template_content.dto.go)

```go
type CreateTemplateContentRequest struct {
    LanguageCode string `json:"language_code" binding:"required"`
    Subject      string `json:"subject"`
    BodyText     string `json:"body_text"`
    BodyHTML     string `json:"body_html"`
    BodyPush     json.RawMessage `json:"body_push,omitempty"`
}

type TemplateContentResponse struct {
    ID           uuid.UUID `json:"id"`
    LanguageCode string    `json:"language_code"`
    Subject      string    `json:"subject"`
    BodyText     string    `json:"body_text"`
    BodyHTML     string    `json:"body_html"`
    CreatedAt    time.Time `json:"created_at"`
}

type AvailableLanguagesResponse struct {
    Languages []string `json:"languages"`
}
```

#### [MODIFY] Controller + Router

Add endpoints:
| Method | Route | Description |
|--------|-------|-------------|
| `POST` | `/templates/:id/content` | Add content in a new language |
| `PUT` | `/templates/:id/content/:lang` | Update content for a language |
| `GET` | `/templates/:id/languages` | List available languages |

---

### Component 5: TemplateManager Update

#### [MODIFY] [template_manager.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/application/service/template_manager.go)

Add methods for content management:
```go
AddContent(ctx, templateID uuid.UUID, req CreateTemplateContentRequest) error
UpdateContent(ctx, templateID uuid.UUID, lang string, req UpdateTemplateContentRequest) error
ListLanguages(ctx, templateID uuid.UUID) ([]string, error)
```

---

## Verification Plan

### Automated Tests
```bash
go build ./...
go test -race ./internal/notification/...
```
- Unit test language fallback in repository (mock DB)
- Unit test content CRUD
- Integration test: Send with non-existent language falls back to `en`

### Manual Verification
1. Create template with `en` content
2. Add `vi` content via API
3. List languages → should return `["en", "vi"]`
4. Send notification with `language: "vi"` → uses Vietnamese content
5. Send notification with `language: "ja"` → falls back to `en`

---

## Effort Estimate

| Component | Effort |
|-----------|--------|
| Repository fallback fix | ~30 min |
| Entity + Model | ~20 min |
| Repository CRUD (4 methods) | ~1h |
| DTOs | ~20 min |
| Controller + Router | ~45 min |
| TemplateManager service | ~30 min |
| Tests | ~1h |
| **Total** | **~4.5 hours** |

---

## State Management
After completion, update:
- `AI_STATE_MINH.md` — mark Notification i18n complete
- `PROJECT_MAIN_BACKLOG.md` — check off Multi-language support, update Notification progress to ~100%
