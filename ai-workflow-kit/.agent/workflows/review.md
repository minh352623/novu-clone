# Workflow: /review

**Trigger:** User types `/review [target]`
**Examples:**
- `/review internal/notification/application/service/notification_service.go`
- `/review "PR #42 — add notification module"`
- `/review "security audit on auth handler"`

---

## Review types

| Command suffix | Focus |
|----------------|-------|
| (default) | Architecture + code quality + best practices |
| `security` | Security vulnerabilities (OWASP Top 10, Go-specific) |
| `performance` | DB queries, allocations, goroutine usage |
| `architecture` | Layer violations, coupling, DDD compliance |
| `test` | Test quality, coverage, missing cases |

---

## Step 1 — Understand scope

Ask the user:
1. What is the context of this code? (new feature / bug fix / refactor)
2. Any specific concerns you already have?
3. Which review type? (or default = all)

---

## Step 2 — Architecture Review

Check in this order:

**Layer dependency violations:**
```
Does Domain layer import Application, Infrastructure, or Interface? → VIOLATION
Does Application layer import Infrastructure directly? → VIOLATION
Does Interface layer contain business logic? → VIOLATION
Does any module import another module's repository directly? → VIOLATION
```

**DDD compliance:**
- Are domain entities pure Go structs (no framework tags in entity layer)?
- Are repository interfaces defined in domain, not in infrastructure?
- Are use cases in application service, not in handler?
- Is business logic in domain/application, not in infrastructure?

**Cross-module communication:**
- Does any service import another module's concrete type? → Must use Interface + Adapter

---

## Step 3 — Code Quality Review

Check each function/method for:

**Error handling:**
```go
// ❌ Error swallowed
result, _ := someFunc()

// ❌ Wrong wrap verb
return fmt.Errorf("failed: %v", err)  // loses error chain

// ❌ HTTP exception in service
return http.StatusBadRequest, err     // wrong layer

// ✅ Correct
return nil, fmt.Errorf("service.Method id=%s: %w", id, err)
```

**Context propagation:**
```go
// ❌ Missing ctx
func (s *Service) DoWork() error { ... }

// ❌ Background inside function
func (s *Service) DoWork(ctx context.Context) error {
    db.Query(context.Background(), ...)  // ignores caller's deadline
}

// ✅ Correct
func (s *Service) DoWork(ctx context.Context) error {
    db.Query(ctx, ...)
}
```

**Goroutine safety:**
- Does every goroutine have an exit condition?
- Is shared state protected (mutex, channel, sync.Map)?
- Are goroutines detached from request context when they should be?

**Resource cleanup:**
- Are HTTP response bodies closed? (`defer resp.Body.Close()`)
- Are DB rows closed? (`defer rows.Close()`)
- Are file handles closed?
- Is context cancel called? (`defer cancel()`)

---

## Step 4 — Security Review (if requested or default)

Check for:

| Issue | What to look for |
|-------|-----------------|
| SQL injection | Raw SQL string concatenation in any layer |
| Sensitive data in logs | password, token, card, ssn in log calls |
| Missing auth | Handler without `@UseGuards` / auth middleware |
| Internal error exposure | `c.JSON(500, gin.H{"error": err.Error()})` leaks stack |
| Missing timeout | External HTTP call without `timeout` config |
| Hardcoded secrets | API keys, passwords in source code |
| Missing input validation | Handler binding without validation pipe |
| Insecure direct object reference | User accesses resource by ID without ownership check |

---

## Step 5 — Performance Review (if requested or default)

Check for:

**Database:**
- N+1 query: query inside a loop → must batch or use JOIN
- Missing index: filter/sort on non-indexed column
- SELECT * when only specific columns needed
- Missing pagination on list endpoints

**Memory:**
- Large allocations in hot paths (use sync.Pool for buffers)
- Unnecessary copies of large structs (use pointer)
- String concatenation in loop (use strings.Builder)

**Goroutines:**
- Creating goroutine per request without pool → use worker pool for high throughput
- Channel unbuffered where buffered would suffice

---

## Step 6 — Output format

For each issue found, report using this format:

```
[SEVERITY] File:line — Issue description
Current code:
  [show current code]
Suggested fix:
  [show fix]
Why: [brief explanation]
```

Severity levels:
- **[CRITICAL]** — Security vulnerability, data loss risk, production outage risk
- **[HIGH]** — Architecture violation, layer breach, goroutine leak
- **[MEDIUM]** — Best practice violation, missing error handling
- **[LOW]** — Code style, naming, minor improvement

End with a summary:
```
Review Summary
==============
Critical: X
High: Y
Medium: Z
Low: W

Overall: APPROVE / REQUEST CHANGES / BLOCK
```
