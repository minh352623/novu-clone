# AI Workflow Kit — Go DDD Projects

> **Drop-in AI configuration** cho mọi Go DDD project.
> Biến AI coding assistant thành một senior engineer hiểu architecture của team.

---

## Cài đặt

```bash
# Copy toàn bộ kit vào root của project
cp -r ai-workflow-kit/. your-project/

# Điền thông tin vào AGENTS.md
# - Thay [Project Name], [Domain]  
# - Cập nhật Module Catalog (Section 7)
# - Cập nhật .ai/context/ARCHITECTURE.md với stack thực tế
# - Cập nhật .ai/context/MODULES.md với modules đã có
```

---

## Cấu trúc

```
.
├── AGENTS.md                    # ⭐ Constitution — AI đọc đầu tiên
│
├── .agent/                      # AI library (workflows, agents, skills)
│   ├── workflows/
│   │   ├── plan.md              # /plan — lập kế hoạch trước khi code
│   │   ├── create.md            # /create — tạo module/endpoint mới
│   │   ├── debug.md             # /debug — debug có hệ thống
│   │   ├── review.md            # /review — code review
│   │   └── refactor.md          # /refactor — refactor an toàn
│   ├── agents/
│   │   ├── backend-go.md        # Go DDD specialist
│   │   ├── security.md          # Security auditor
│   │   └── qa.md                # Testing specialist
│   └── skills/
│       ├── ddd-golang.md        # DDD patterns in Go
│       ├── sqlc-patterns.md     # sqlc query & repository patterns
│       └── error-handling.md    # Go error handling patterns
│
├── .ai/                         # Project-specific context
│   ├── rules/                   # Per-file-type rules (Cursor/Windsurf)
│   │   ├── handler.mdc
│   │   ├── service.mdc
│   │   ├── repository.mdc
│   │   └── dto.mdc
│   ├── context/                 # Project knowledge base
│   │   ├── ARCHITECTURE.md      # System overview + stack
│   │   ├── ADR.md               # Architecture Decision Records
│   │   ├── PATTERNS.md          # Approved code patterns
│   │   └── MODULES.md           # Module catalog + dependencies
│   └── prompts/                 # Reusable prompt templates
│       ├── new-module.md
│       ├── add-endpoint.md
│       └── write-test.md
│
└── docs/
    └── onboarding/
        └── AI_GUIDE.md          # Hướng dẫn cho thành viên mới
```

---

## Quick Start cho team member mới

1. Setup tool: `npm install -g @anthropic-ai/claude-code`
2. Đọc: `AGENTS.md` + `.ai/context/ARCHITECTURE.md` + `.ai/context/PATTERNS.md`
3. Đọc: `docs/onboarding/AI_GUIDE.md` cho hướng dẫn hàng ngày
4. Test: Hỏi AI "Tóm tắt 3 rules quan trọng nhất của dự án này"

---

## Checklist khi customize cho project mới

- [ ] Cập nhật `AGENTS.md` section 1 (Project Overview, stack)
- [ ] Cập nhật `AGENTS.md` section 7 (Module Catalog)
- [ ] Cập nhật `.ai/context/ARCHITECTURE.md` (stack thực tế, modules)
- [ ] Cập nhật `.ai/context/MODULES.md` (tất cả modules đang có)
- [ ] Cập nhật `.ai/context/ADR.md` (thêm ADRs nếu dùng tech khác)
- [ ] Cập nhật `.ai/rules/*.mdc` nếu framework khác GIN
