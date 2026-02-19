# Workflow: /plan

**Trigger:** User types `/plan [task description]`
**Purpose:** Break down complex tasks into a structured, approved implementation plan before writing any code.

---

## When to use /plan

Use this workflow when:
- Task involves more than 2 files being changed
- Task requires cross-module changes
- Task introduces a new pattern or dependency
- Task affects database schema
- Task involves concurrent logic (goroutines, channels)
- You are unsure where to start

Do NOT use for: simple bug fixes, typo corrections, config changes.

---

## Step 1 — Understand the task (Clarify First)

Before planning, ask the user:

1. **Scope:** What is the expected final behavior? How will we know it's done?
2. **Constraints:** Are there performance requirements? Deadlines? Breaking change concerns?
3. **Context:** Does this touch existing code or is it greenfield?
4. **Dependencies:** Does this need changes in other modules?

Wait for answers before proceeding.

---

## Step 2 — Research (Read before write)

Before generating the plan, read:
- `AGENTS.md` — architecture rules and constraints
- `.ai/context/ARCHITECTURE.md` — system overview
- `.ai/context/MODULES.md` — which module owns what
- `.ai/context/ADR.md` — decisions that constrain this task
- `.ai/context/PATTERNS.md` — which patterns apply

---

## Step 3 — Generate Implementation Plan

Create a structured plan in this format:

```markdown
# Implementation Plan: [Task Name]
Date: [YYYY-MM-DD]
Estimated effort: [S / M / L / XL]
Affected layers: [list layers]
Affected modules: [list modules]

## Summary
[1-2 sentences describing what will be built]

## Approach
[Which pattern/strategy will be used and WHY]

## Files to create
| File | Layer | Purpose |
|------|-------|---------|
| internal/[module]/domain/... | Domain | ... |

## Files to modify
| File | Change | Reason |
|------|--------|--------|

## Database changes
- [ ] New table: [table name]
- [ ] New column: [table.column type]
- [ ] New index: [table(col)]
- [ ] New sqlc query: [query name and signature]

## Implementation order
1. Domain entities + interfaces (domain layer)
2. sqlc SQL queries
3. Repository implementation (infrastructure layer)
4. Application service (application layer)
5. DTOs + Handler (interface layer)
6. Module wiring

## Test plan
- Unit tests: [list what to test]
- Manual verification: [curl commands or steps]

## Risk & open questions
- [Uncertainty or potential issue]

## Out of scope
- [What this deliberately does NOT cover]
```

---

## Step 4 — Present and Wait for Approval

Present the plan to the user.
**Do NOT write any implementation code until the user explicitly approves.**

If the user requests changes: update plan → ask for approval again.
If the user approves: proceed to implementation in the exact order from the plan.

---

## Step 5 — Implement (after approval only)

Follow implementation order strictly.
After each major file: briefly summarize what was done.
If you discover something unexpected: STOP → inform user → get direction before continuing.

---

## Step 6 — Verify

After implementation, run checklist from `AGENTS.md` Section 5.
Report results to user with any items needing manual verification.
