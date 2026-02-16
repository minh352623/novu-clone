# Plan: Refactor Messages to Closure Table

> **Goal**: Replace the current "Nested Set" model (`lft`, `rgt`) in the `messages` table with the **"Closure Table"** pattern. This optimizes for transactional integrity, atomic updates via SQL, and efficient querying of complex conversation trees, aligning with the "Comprehensive Solution" strategy requested.

## 1. Context & Rationale
- **Current State**: `schema.sql` defines `messages` with `lft, rgt, depth` (Nested Set).
- **Problem**: Nested Set acts like a linked list in terms of updates (insertion requires updating all right-side nodes), which causes lock contention. It's also brittle for partition strategies.
- **Solution**: **Closure Table**.
    - Separate table `message_closure` storing *all* paths.
    - `messages` table becomes clean (only content).
    - Insert cost: O(Depth). Query cost: O(1) JOIN.
    - Supports atomic updates better than Materialized Path or Nested Set.

## 2. Schema Changes

### 2.1. Refactor `messages` Table
- **Remove**: `lft`, `rgt`, `depth`, `parent_id` (optional, but keep `parent_id` as adjacency for quick direct-parent check).
- **Keep**: `id`, `pool_id` (Root container), `content`, `created_at`.

```sql
ALTER TABLE messages DROP COLUMN lft;
ALTER TABLE messages DROP COLUMN rgt;
ALTER TABLE messages DROP COLUMN depth;
-- Keep parent_id for "Immediate Parent" adjacency (useful for lightweight queries)
```

### 2.2. Create `message_closure` Table
This table stores every ancestor-descendant relationship.

```sql
CREATE TABLE message_closure (
    ancestor_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    depth INTEGER NOT NULL, -- Distance between ancestor and descendant
    PRIMARY KEY (ancestor_id, descendant_id)
);

-- Index for retrieving a full thread
CREATE INDEX idx_closure_descendant ON message_closure(descendant_id);
-- Index for retrieving a subtree
CREATE INDEX idx_closure_ancestor ON message_closure(ancestor_id);
```

## 3. Implementation Logic

### 3.1. Insert "Reply" Logic (SQL/Go)
To insert a new message `C` as a child of `B`:
1. **Insert C** into `messages`.
2. **Insert Paths** into `message_closure`:
   ```sql
   INSERT INTO message_closure (ancestor_id, descendant_id, depth)
   SELECT ancestor_id, 'C_uuid', depth + 1
   FROM message_closure
   WHERE descendant_id = 'B_uuid'
   UNION ALL
   SELECT 'C_uuid', 'C_uuid', 0; -- Self-reference
   ```

### 3.2. Query "Full Conversation" (Thread View)
```sql
SELECT m.* 
FROM messages m
JOIN message_closure c ON m.id = c.descendant_id
WHERE c.ancestor_id = 'ROOT_MSG_ID'
ORDER BY c.depth ASC, m.created_at ASC;
```

### 3.3. Update "Reply Count" (Efficient Aggregate)
The request mentioned updating `reply_count` for ancestors.
```sql
UPDATE messages m
SET reply_count = reply_count + 1 -- Assuming we add this denormalized column later
WHERE m.id IN (
    SELECT ancestor_id FROM message_closure WHERE descendant_id = 'NEW_MSG_ID' -- Exclude self if needed, usually depth > 0
    AND ancestor_id != 'NEW_MSG_ID'
);
```

## 4. Execution Steps

### Phase 1: Schema Migration
- [ ] Create Migration `000XX_refactor_messages_closure.sql`.
- [ ] Drop `add_message_node` function (Nested Set logic).
- [ ] Create `message_closure` table.
- [ ] Update `messages` table definition.

### Phase 2: Code Implementation
- [ ] Update `Message` Entity (Remove `Lft`, `Rgt`).
- [ ] Implement `MessageRepository.CreateReply` using the Closure Table INSERT logic.
- [ ] Implement `MessageRepository.GetThread` using JOINs.

## 5. Verification
- **Test**: Insert deeply nested replies (A -> B -> C -> D).
- **Verify**: Querying A as ancestor returns [A, B, C, D] with correct depths.
- **Performance**: Benchmark insert speed vs Nested Set (expect stable O(Depth) insert).

## 6. Agent Assignments
- **Database Architect**: Finalize Schema & Migration.
- **Backend Specialist**: Update Repo Logic.
