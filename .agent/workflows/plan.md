---
description: Chưa biết bắt đầu từ đâu? Lập kế hoạch trước.
---

# /plan - Project Planning Mode

$ARGUMENTS

---

## 🔴 CRITICAL RULES

1. **NO CODE WRITING** - This command creates plan file only
2. **Use project-planner agent** - NOT Antigravity Agent's native Plan mode
3. **Socratic Gate** - Ask clarifying questions before planning
4. **Dynamic Naming** - Plan file named based on task

---

## Task

Use the `project-planner` agent with this context:

```
CONTEXT:
- User Request: $ARGUMENTS
- Mode: PLANNING ONLY (no code)
- Output: .ai/plans/active/PLAN-{task-slug}.md (dynamic naming)

NAMING RULES:
1. Extract 2-3 key words from request
2. Lowercase, hyphen-separated
3. Max 30 characters
4. Example: "e-commerce cart" → PLAN-ecommerce-cart.md

RULES:
1. Follow project-planner.md Phase -1 (Context Check).
2. Follow project-planner.md Phase 0 (Socratic Gate).
3. **Collaboration Protocol**: The plan MUST include a "Review Levels" section as per `.ai/INSTRUCTIONS.md`.
4. **State Management**: The plan MUST include a step to update `.ai/members/{NAME}.md` upon completion of each phase.
5. **Module Ownership**: The plan MUST check `.ai/MODULE_OWNERSHIP.md` for cross-module impacts.
6. Create .ai/plans/active/PLAN-{slug}.md with task breakdown.
6. DO NOT write any code files.
7. REPORT the exact file name created.
```

---

## Expected Output

| Deliverable | Location |
|-------------|----------|
| Project Plan | `.ai/plans/active/PLAN-{task-slug}.md` |
| Task Breakdown | Inside plan file |
| Agent Assignments | Inside plan file |
| Verification Checklist | Phase X in plan file |

---

## After Planning

Tell user:
```
[OK] Plan created: .ai/plans/active/PLAN-{slug}.md

Next steps:
- Review the plan
- Run `/create` to start implementation
- Or modify plan manually
```

---

## Naming Examples

| Request | Plan File |
|---------|-----------|
| `/plan e-commerce site with cart` | `.ai/plans/active/PLAN-ecommerce-cart.md` |
| `/plan mobile app for fitness` | `.ai/plans/active/PLAN-fitness-app.md` |
| `/plan add dark mode feature` | `.ai/plans/active/PLAN-dark-mode.md` |
| `/plan fix authentication bug` | `.ai/plans/active/PLAN-auth-fix.md` |
| `/plan SaaS dashboard` | `.ai/plans/active/PLAN-saas-dashboard.md` |

---

## Usage

```
/plan e-commerce site with cart
/plan mobile app for fitness tracking
/plan SaaS dashboard with analytics
```
