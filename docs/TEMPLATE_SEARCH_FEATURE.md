# ✅ Email Template Search Feature - Complete!

## 🎯 Implementation Summary

Added search functionality to email template listing endpoint.

## 📦 Changes Made

### 1. Domain Layer ✅
**File:** `internal/settings/domain/repository/email_template.repository.go`

Added `Search` field to `TemplateFilters`:
```go
type TemplateFilters struct {
    Search         *string  // NEW!
    Category       *string
    Status         *int
    IncludeDeleted bool
    Limit          int
    Offset         int
}
```

### 2. Infrastructure Layer ✅
**File:** `internal/settings/infrastructure/persistence/repository/email_template.repository.go`

Implemented search in `GetAll` method:
```go
// Search by template name or template code
if filters.Search != nil && *filters.Search != "" {
    searchPattern := "%" + *filters.Search + "%"
    query = query.Where("template_name ILIKE ? OR template_code ILIKE ?", searchPattern, searchPattern)
}
```

**Features:**
- Case-insensitive search (`ILIKE`)
- Searches in both `template_name` and `template_code`
- Uses wildcard pattern matching

### 3. Controller Layer ✅
**File:** `internal/settings/controller/http/email_template.handler.go`

Added search parameter extraction:
```go
var search *string
if s := ctx.Query("search"); s != "" {
    search = &s
}
```

Updated Swagger documentation:
```go
// @Param search query string false "Search by template name or code"
```

## 🚀 Usage Examples

### Search by Template Name
```bash
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=Welcome'
```

### Search by Template Code
```bash
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=WELCOME_EMAIL'
```

### Combined Filters
```bash
# Search + Category
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=welcome&category=transactional'

# Search + Status
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=reset&status=1'

# Search + Pagination
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=email&limit=10&offset=0'
```

## 📊 Search Behavior

**Case-Insensitive:**
- `search=welcome` matches "Welcome Email", "WELCOME_EMAIL", "welcome-newsletter"

**Partial Match:**
- `search=wel` matches "Welcome", "Farewell"
- `search=pass` matches "Password Reset", "Guest Pass"

**Multi-Field:**
Searches in both fields simultaneously:
- Template Name: "Welcome to TekNix"
- Template Code: "WELCOME_TEKNIX"

## ✅ Build Status

```bash
✅ go build ./...
✅ make swag
```

## 🧪 Test Examples

### Test 1: Find Welcome Templates
```bash
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=welcome'
```

Expected: Returns all templates with "welcome" in name or code

### Test 2: Find by Code Prefix
```bash
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=WELCOME_'
```

Expected: Returns templates with codes starting with "WELCOME_"

### Test 3: Case Insensitive
```bash
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=TEKNIX'
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=teknix'
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates?search=TekNix'
```

Expected: All return same results

## 📝 API Documentation

**Endpoint:** `GET /v1/api/settings/email-templates`

**Query Parameters:**
- `search` (optional) - Search by template name or code
- `category` (optional) - Filter by category
- `status` (optional) - Filter by status (0, 1, 2)
- `limit` (optional) - Results per page (default: 20, max: 100)
- `offset` (optional) - Pagination offset (default: 0)

**Response:**
```json
{
  "templates": [...],
  "total": 10,
  "limit": 20,
  "offset": 0,
  "totalPage": 1,
  "page": 1
}
```

## ✨ Features

✅ Case-insensitive search  
✅ Partial matching  
✅ Multi-field search (name + code)  
✅ Combines with existing filters  
✅ Works with pagination  
✅ Swagger documentation updated  

Implementation complete! 🎉
