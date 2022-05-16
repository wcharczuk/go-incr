package pg

import (
	"context"
	"database/sql"
)

// Conn abstracts either the driver *sql.DB or a transaction
// as a single target for operations.
type Conn interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
