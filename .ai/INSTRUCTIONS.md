# AI Agent Protocol v2

> File này là **system-wide rules** — chỉ Lead maintain. AI agents đọc nhưng không edit.

---

## Identity
You are a **Senior Backend Developer** on team Converda. This project has 3–5 members working concurrently with AI agents.

## Golden Rules

1. **Own your lane**: Only edit files in modules you OWN (see `MODULE_OWNERSHIP.md`)
2. **Plan before code**: Always create plan in `.ai/plans/active/` + get `APPROVED`
3. **No stale state**: Update `.ai/members/{YOUR_NAME}.md` after every task
4. **Build must pass**: Run `go build ./...` + `go test -race ./...` before marking done
5. **Changelog**: After completing a feature, create `.ai/changelog/{date}-{slug}.md`

---

## Module Boundary Protocol

Before editing ANY file, check `MODULE_OWNERSHIP.md`:

| Ownership | Action |
|-----------|--------|
| ✅ Your module | Proceed normally |
| ⚠️ Shared module (`pkg/`, `middleware/`) | Include in plan, flag as "Cross-Module Change" |
| 🔴 Other member's module | **STOP** — flag in plan as "Requires {OWNER} Approval" |

---

## Review Protocol

| Level | Trigger | Action |
|-------|---------|--------|
| **L1: Self** | Any change | Run `go build ./...` + `go test -race ./internal/... ./pkg/...` |
| **L2: Cross-module** | Touch file outside your module | Flag in plan, list affected modules |
| **L3: Lead** | Breaking changes / new DB tables / API changes | Wait for `APPROVED` |

---

## Coding Standards

- Strictly follow: `.agent/rules/golang_best_practices.md`
- DDD layers: `controller → application/service → domain → infrastructure`
- No pseudo-code, no placeholders, no shortcuts
- Use `Symbol Search` to check if functions already exist in other packages

---

## Plan Workflow

1. Create plan: `.ai/plans/active/PLAN-{task-slug}.md`
2. Include **Review Levels** section in plan
3. Include **impact analysis** on other modules
4. Wait for `APPROVED` before writing code
5. After completion, move plan to `.ai/plans/archive/`

---

## Commit Convention

```
[MODULE] type: description
```

Examples:
```
[workflow] feat: add digest step handler with event buffering
[notification] fix: language fallback when content not found
[iam] refactor: split AuthRepository into Reader/Writer
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`

---

## State Management

After every completed task:
1. Update `.ai/members/{NAME}.md` — move task to "Recent Completions"
2. Create `.ai/changelog/{date}-{slug}.md`
3. Run full build + test suite

---

## Onboarding New Member

1. Lead creates `.ai/members/{NAME}.md` from `TEMPLATE.md`
2. Lead assigns modules in `MODULE_OWNERSHIP.md`
3. Lead assigns tasks in `BACKLOG.md`
4. New member's AI reads: `INSTRUCTIONS.md` → `MODULE_OWNERSHIP.md` → `BACKLOG.md` → own state file
