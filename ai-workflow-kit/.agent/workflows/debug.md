# Workflow: /debug

**Trigger:** User types `/debug [problem description]`
**Purpose:** Systematic root cause analysis — find and fix bugs without guessing.

**Rule:** NEVER suggest a fix before identifying root cause. Hypothesize → Investigate → Confirm → Fix.

---

## Step 1 — Gather information

Ask the user to provide:

```
1. Error message / stack trace (exact text, not paraphrase)
2. Steps to reproduce (step by step)
3. Expected behavior vs actual behavior
4. Environment: local / staging / production
5. When did this start? (always / after specific commit / after deploy)
6. Recent changes: last PR merged, last config change
7. Frequency: always / intermittent / under load only
```

Do NOT proceed until you have at least items 1, 2, 3.

---

## Step 2 — Classify the bug

Before hypothesizing, classify:

| Class | Indicators | Investigation approach |
|-------|-----------|----------------------|
| **Logic bug** | Wrong result, off-by-one, wrong condition | Read code path carefully |
| **Concurrency bug** | Intermittent, load-dependent, race detector | Check goroutines, shared state |
| **Integration bug** | Fails on specific env, external service involved | Check config, network, timeouts |
| **Data bug** | Works in dev, fails in prod, specific records | Check DB state, migrations |
| **Memory/GC bug** | OOM, gradual slowdown, goroutine count growing | Check allocations, goroutine leaks |
| **Configuration bug** | Env-specific, after deployment | Check env vars, secrets |

---

## Step 3 — Hypothesize (Top 3, ranked)

List exactly 3 hypotheses. For each:
- State the hypothesis clearly
- Explain the evidence that supports it
- Assign probability: High / Medium / Low
- Describe ONE specific diagnostic step to confirm or eliminate it

Format:
```
Hypothesis 1 [High probability]: [Explain what you think is happening]
Evidence: [Why you think this]
Diagnostic: [Exact diagnostic step — add this log / run this query / check this value]

Hypothesis 2 [Medium]: ...
Hypothesis 3 [Low]: ...
```

---

## Step 4 — Investigate

Guide the user through diagnostics one at a time:
- Provide exact code for log statements to add
- Provide exact DB queries to run
- Provide exact commands (`go test -race`, `pprof`, `GODEBUG=gctrace=1`)

After each diagnostic result: update hypothesis ranking.
If hypothesis is eliminated: move to next.
If confirmed: proceed to Step 5.

---

## Step 5 — Root cause statement

Before fixing, state clearly:
```
Root cause: [Exact explanation of WHY the bug occurs]
Affected code: [File:line range]
Trigger condition: [When exactly does this manifest]
```

---

## Step 6 — Fix

Propose the fix:
1. Show the fix with minimal diff
2. Explain WHY this fix addresses the root cause
3. Identify any edge cases the fix introduces
4. Confirm the fix doesn't violate AGENTS.md rules

---

## Step 7 — Prevent recurrence

After fixing:
1. What test would have caught this? → Write it.
2. Is there a pattern violation that enabled the bug? → Suggest update to AGENTS.md or best practices.
3. Could this bug exist elsewhere in the codebase? → Suggest a grep to check.

---

## Common Go bug patterns to check first

### Goroutine leak checklist
```go
// Q1: Does every goroutine have an exit condition?
go func() {
    for {
        select {
        case <-ctx.Done(): return    // ← must exist
        case work := <-ch: process(work)
        }
    }
}()

// Q2: Is the channel buffered correctly?
// Sending to unbuffered channel with no receiver = goroutine blocks forever

// Q3: Is WaitGroup.Done() called even on error path?
defer wg.Done() // always use defer
```

### Context cancellation
```go
// Q: Is a child context's cancel() being called?
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel() // ← missing cancel() = context leak
```

### sqlc / pgx error handling
```go
// Q: Is pgx.ErrNoRows being checked before other errors?
row, err := q.GetUserByID(ctx, id)
if errors.Is(err, pgx.ErrNoRows) {
    return nil, domain.ErrUserNotFound  // specific case first
}
if err != nil {
    return nil, fmt.Errorf("repo: %w", err)  // general case second
}
```

### Nil pointer dereference
```go
// Q: Is the pointer checked before dereferencing?
// Q: Can the function return nil without error? (anti-pattern — avoid this)
```
