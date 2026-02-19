# Hướng dẫn làm việc với AI — Dành cho thành viên mới

> **Đọc tài liệu này trong ngày đầu onboarding.**
> Đây là guide thực tế, ngắn gọn — không phải lý thuyết.

---

## TL;DR — 5 điều cần nhớ

1. **Giao tiếp với AI bằng Tiếng Việt** — code và comments viết bằng Tiếng Anh
2. **Dùng `/plan` trước khi làm task lớn** — đừng để AI code ngay
3. **Chạy AI Checklist trước khi commit** — copy từ Mục 4
4. **Khi AI tạo sai pattern** — nhắc AI đọc lại `AGENTS.md` và `PATTERNS.md`
5. **AGENTS.md là luật** — nếu AI làm khác với AGENTS.md, AI đang sai

---

## 1. Setup môi trường (15 phút)

### Bước 1: Cài AI tool

Chọn **một** trong hai:

```bash
# Option A: Claude Code (khuyến nghị cho planning và tasks lớn)
npm install -g @anthropic-ai/claude-code
claude  # trong project folder

# Option B: Cursor (khuyến nghị cho inline code và autocomplete)
# Download tại cursor.com, mở project folder
```

### Bước 2: Verify AI đọc được context

Trong chat với AI, gõ:
```
Đọc AGENTS.md và tóm tắt 3 rule quan trọng nhất của dự án này.
```

Nếu AI trả lời đúng về sqlc, DDD layers, cross-module communication → ✅ Setup OK
Nếu AI trả lời chung chung → ❌ AI chưa đọc được AGENTS.md, kiểm tra file path

### Bước 3: Đọc tài liệu quan trọng (~30 phút)

```bash
# Bắt buộc đọc trước khi bắt đầu code
cat AGENTS.md                      # ~10 phút — rules bắt buộc
cat .ai/context/ARCHITECTURE.md    # ~10 phút — cấu trúc hệ thống
cat .ai/context/PATTERNS.md        # ~10 phút — patterns cần follow
```

---

## 2. Quy trình làm task hàng ngày

### Với task nhỏ (< 30 phút, 1-2 files)

```
Mô tả task cho AI → nhận code → chạy checklist → commit
```

Ví dụ:
```
Thêm method GetByEmail vào UserRepository.
Interface: domain/repository/i_user_repository.go
Implement: infrastructure/persistence/repository/user_repository.go
Dùng sqlc query tên GetUserByEmail đã có trong queries/user.sql
```

### Với task vừa (1-2 giờ, nhiều files)

```
/plan [mô tả task] → review plan → approve → nhận code → checklist → commit
```

### Với task lớn (nửa ngày trở lên, cross-module)

```
/plan [mô tả task] → review plan → approve → implement từng phase → checklist → PR
```

---

## 3. Các lệnh hay dùng nhất

Gõ thẳng vào chat với AI:

```bash
# Tạo module mới
# Trước tiên copy prompt từ .ai/prompts/new-module.md và điền thông tin

# Thêm endpoint
# Copy prompt từ .ai/prompts/add-endpoint.md

# Debug lỗi
/debug
Error: [paste error message]
File: [filepath:line nếu có]
Context: [đang làm gì thì gặp lỗi]

# Review code trước khi tạo PR
/review internal/[module]/... — kiểm tra toàn bộ module

# Viết test cho method
# Copy prompt từ .ai/prompts/write-test.md

# Refactor
/refactor internal/[module]/application/service/[name].go "[vấn đề cần sửa]"
```

---

## 4. AI Output Checklist — Chạy trước mỗi commit

Copy checklist này, điền tên task, tick từng mục:

```
═══════════════════════════════════════
AI OUTPUT CHECKLIST
Task: ___________________________________
Date: ___________________________________
Reviewer: _______________________________
═══════════════════════════════════════

ARCHITECTURE
[ ] Không vi phạm dependency rule (domain không import infrastructure/application)
[ ] Cross-module: dùng Interface + Adapter, không import repo trực tiếp
[ ] Không có import gorm.io/gorm ở bất kỳ đâu

DATABASE
[ ] Chỉ dùng sqlc-generated methods trong repository
[ ] Không có raw SQL string trong service/application layer
[ ] Multi-table writes dùng UnitOfWork (transaction)

GO BEST PRACTICES
[ ] ctx là tham số đầu tiên của mọi I/O function
[ ] Errors wrap bằng %w (không phải %v)
[ ] Không có goroutine thiếu exit condition
[ ] Goroutine detached từ request dùng context.WithoutCancel()
[ ] Goroutine có recover() panic

SECURITY
[ ] Không log password, token, card, PII
[ ] External HTTP calls có timeout
[ ] Không lộ internal error ra HTTP response
[ ] Không hardcode secret/API key

TESTING
[ ] Unit test cho business logic mới
[ ] Mocks là interfaces, không phải concrete types
[ ] Table-driven test cho multiple input cases

GENERAL
[ ] Không có TODO/FIXME không có tracking issue
[ ] Export chỉ những gì cần thiết
[ ] No commented-out code

RESULT: PASS □   FAIL □ (fix issues, re-run checklist)
═══════════════════════════════════════
```

---

## 5. Xử lý khi AI làm sai

### Tình huống 1: AI tạo code dùng GORM

```
❌ AI tạo: db.Where("id = ?", id).First(&user)

Nói với AI:
"Dự án này không dùng GORM. Đọc lại AGENTS.md ADR-001 và 
.ai/context/PATTERNS.md. Tạo lại theo sqlc pattern."
```

### Tình huống 2: AI vi phạm layer dependency

```
❌ AI import infrastructure trong application service

Nói với AI:
"Application layer không được import infrastructure trực tiếp.
Đọc AGENTS.md section 3.1 và sửa lại để dùng interface injection."
```

### Tình huống 3: AI import trực tiếp repository của module khác

```
❌ AI viết: import "project/internal/user/infrastructure/persistence/repository"

Nói với AI:
"Đây là cross-module dependency violation. Đọc AGENTS.md ADR-003
và .ai/context/PATTERNS.md Pattern 1. Tạo IUserReader port và LocalUserAdapter."
```

### Tình huống 4: AI code ngay không hỏi clarifying questions

```
Nói với AI:
"Trước khi code, đọc .agent/workflows/create.md (hoặc plan.md)
và follow quy trình: hỏi clarifying questions trước, tạo implementation_plan.md,
chờ tôi approve rồi mới implement."
```

### Tình huống 5: AI không biết project context

```
Nói với AI:
"Đọc các file sau rồi trả lời:
1. AGENTS.md
2. .ai/context/ARCHITECTURE.md
3. .ai/context/MODULES.md
Sau đó cho tôi biết bạn đã hiểu project structure chưa."
```

---

## 6. Câu hỏi thường gặp

**Q: Khi nào dùng `/plan` vs code thẳng?**
A: Nếu task cần sửa > 2 files HOẶC liên quan đến module mới HOẶC bạn không chắc chắn cách tiếp cận → dùng `/plan`. Còn lại code thẳng được.

**Q: Tôi có thể dùng Claude.ai (web) thay vì Claude Code không?**
A: Được, nhưng phải paste nội dung AGENTS.md vào đầu mỗi conversation. Claude Code tự động đọc AGENTS.md tốt hơn.

**Q: AI tạo code không compile, phải làm gì?**
A: Paste error vào chat: "Lỗi compile: [error message]. Sửa lại." AI tự hiểu context.

**Q: Team có workflow nào để approve plan của AI không?**
A: AI tạo implementation_plan.md → bạn đọc và comment trực tiếp trong chat → chỉ approve khi plan đúng. Không cần process phức tạp hơn.

**Q: Tôi muốn thêm một pattern mới vào PATTERNS.md, phải làm thế nào?**
A: Tạo PR với thay đổi trong `.ai/context/PATTERNS.md` → team lead review → merge. Sau đó AGENTS.md có thể cần update.

**Q: File `.agent/` có cần commit vào git không?**
A: Có — vì team cần dùng chung workflows. Khác với Antigravity Kit recommendation (họ dùng `.git/info/exclude`), team này commit `.agent/` để đảm bảo nhất quán.

---

## 7. Tham khảo nhanh

| Cần gì | Đọc file nào |
|--------|-------------|
| Hiểu kiến trúc tổng thể | `.ai/context/ARCHITECTURE.md` |
| Hiểu module X làm gì | `.ai/context/MODULES.md` |
| Code pattern chuẩn | `.ai/context/PATTERNS.md` |
| Quyết định kiến trúc đã có | `.ai/context/ADR.md` |
| Rules cho handler files | `.ai/rules/handler.mdc` |
| Rules cho service files | `.ai/rules/service.mdc` |
| Rules cho repository files | `.ai/rules/repository.mdc` |
| Lệnh `/plan` làm gì | `.agent/workflows/plan.md` |
| Lệnh `/create` làm gì | `.agent/workflows/create.md` |
| Lệnh `/debug` làm gì | `.agent/workflows/debug.md` |
| Tạo module mới | `.ai/prompts/new-module.md` |
| Thêm endpoint | `.ai/prompts/add-endpoint.md` |
| Viết test | `.ai/prompts/write-test.md` |
