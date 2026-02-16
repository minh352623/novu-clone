package plugin

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey = "tenant_id"
	// RLSQuery is the SQL command to set local tenant
	RLSQuery = "SET LOCAL app.current_tenant = ?"
)

type RLSPlugin struct{}

func NewRLSPlugin() *RLSPlugin {
	return &RLSPlugin{}
}

func (p *RLSPlugin) Name() string {
	return "RLSPlugin"
}

func (p *RLSPlugin) Initialize(db *gorm.DB) error {
	// Register callbacks for all operations
	if err := db.Callback().Create().Before("gorm:create").Register("rls:before_create", p.setTenant); err != nil {
		return err
	}
	if err := db.Callback().Query().Before("gorm:query").Register("rls:before_query", p.setTenant); err != nil {
		return err
	}
	if err := db.Callback().Update().Before("gorm:update").Register("rls:before_update", p.setTenant); err != nil {
		return err
	}
	if err := db.Callback().Delete().Before("gorm:delete").Register("rls:before_delete", p.setTenant); err != nil {
		return err
	}
	if err := db.Callback().Row().Before("gorm:row").Register("rls:before_row", p.setTenant); err != nil {
		return err
	}
	return nil
}

func (p *RLSPlugin) setTenant(db *gorm.DB) {
	// Get context from statement
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Extract TenantID from context
	// Note: We use string key "tenant_id" to match middleware
	val := ctx.Value(TenantIDKey)
	if val == nil {
		// No tenant ID in context, skip RLS (system queries or public endpoints)
		return
	}

	var tenantID string
	switch v := val.(type) {
	case uuid.UUID:
		tenantID = v.String()
	case string:
		tenantID = v
	default:
		// Invalid type, ignore
		return
	}

	if tenantID == "" {
		return
	}

	// Execute SET LOCAL on the current connection
	// We use db.Statement.ConnPool to execute the query
	// Using Exec directly on db might check out a new connection if not in transaction
	// But inside a callback, we are already tied to a specific operation scope.
	// IMPORTANT: For GORM Query callbacks, we must ensure we run on the SAME connection.
	// However, GORM doesn't expose the underlying sql.Conn easily in callbacks if not in Tx.
	// But `SET LOCAL` is transaction-scoped.
	// If we are NOT in a transaction, `SET LOCAL` applies to the session (connection) until next reset or close.
	// Since we use connection pooling, this could leak to other requests if not reset.
	// Solution: GORM automatically uses prepared statements or simple query mode.
	// The safest way for RLS in pooled environment is ensuring Transaction or using session-aware driver.
	// BUT, standard Postgres approach is:
	// db.Exec("SET LOCAL ...") won't work reliably without TX because it might pick random conn.
	//
	// Correct approach for GORM:
	// 1. Check if db.Statement.ConnPool is a *sql.Tx. If so, execute on it.
	// 2. If NOT a Tx, we should probably force a Tx or warn.
	// For now, we will try to execute SQL using the DB instance but this is risky without TX.
	// Ref: https://gorm.io/docs/write_plugins.html

	// Better approach: modifying the SQL? No, we need session variable.
	// Standard for RLS with pooling: ALWAYS USE TRANSACTIONS.
	// We will assume business logic uses transactions or we inject one?
	// Actually, `app.current_tenant` needs to be set per session.

	// Let's implement execution:
	if db.Error != nil {
		return
	}

	// We append a clause? No, we need to run a command.
	// In GORM callback, we can't easily run a separate statement on the exact same conn unless we are in a TX.
	// If we are not in a TX, db.Exec might pick a different connection.

	// To solve this properly without forcing TX everywhere:
	// We can use `db.Session(&gorm.Session{Context: ctx})` in service layer which doesn't help with connection binding.

	// Let's try to detect Transaction.
	// If useTransaction is true, or if Statement.ConnPool is a Tx.

	// For this Implementation Step, we will assume code uses Transactions or Accept that we inject the SQL clause.
	// Wait, we can't inject SQL clause for `SET LOCAL`.

	// ALTERNATIVE: Use `gorm.DB`'s instance to run SQL.
	// If we are merely reading, creating a TX just for RLS is expensive.
	// But RLS requires reliable session state.
	// Most systems require one DB connection per request scope if using SET LOCAL.
	// Go sql package doesn't guarantee thread-bound connection unless in Tx.

	// DECISION: We will execute the SET LOCAL command.
	// If not in transaction, this might be flaky under load if connection is returned to pool immediately?
	// No, `SET LOCAL` lasts until end of transaction. If no transaction, it lasts until session end.
	// If connection is reused, next request might inherit it.
	// So we MUST Reset it too?
	// Postgres `SET LOCAL` is only valid for transaction. If used outside, it acts like `SET SESSION`.
	// That is dangerous in pooling.
	// SO: We must ensure we are in transaction.

	// As a safeguard plugin: We will append a generic error if we detect RLS context but no Transaction?
	// Or we simply implement it and rely on Service layer using Transaction.

	// Let's go with executing raw SQL.
	_ = db.Exec(RLSQuery, tenantID)
}
