# `.ai/` — AI Agent Collaboration Hub

> Thư mục này chứa toàn bộ cấu hình và trạng thái cho AI agents trong team.

## Cấu Trúc

```
.ai/
├── INSTRUCTIONS.md          ← System-wide rules (Lead maintains)
├── MODULE_OWNERSHIP.md      ← Module → Owner mapping (Lead maintains)
├── BACKLOG.md               ← Project backlog + sprint (Lead maintains)
├── README.md                ← This file
│
├── members/                 ← Per-member AI state
│   ├── MINH.md              ← Minh's AI agent state
│   └── TEMPLATE.md          ← Template for new members
│
├── plans/
│   ├── active/              ← Plans currently being implemented
│   └── archive/             ← Completed plans (moved here when done)
│
└── changelog/               ← Per-feature change summaries
    └── 2026-02-17-digest-steps.md
```

## Rules

| File | Who Edits | Who Reads |
|------|-----------|-----------|
| `INSTRUCTIONS.md` | Lead only | All agents |
| `MODULE_OWNERSHIP.md` | Lead only | All agents |
| `BACKLOG.md` | Lead only | All agents |
| `members/{NAME}.md` | That member's AI only | All agents |
| `plans/active/*.md` | Assigned member's AI | All agents |
| `changelog/*.md` | Feature author's AI | All agents |

**Golden Rule**: Không bao giờ 2 AI agents cùng edit 1 file.
