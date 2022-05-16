package pg

import (
	"database/sql"
	"time"

	"github.com/wcharczuk/go-incr"
)

type schemaVersion struct {
	Version      string
	TimestampUTC time.Time
}

func (sv *schemaVersion) Populate(r *sql.Rows) error {
	return r.Scan(&sv.Version, &sv.TimestampUTC)
}

func (sv *schemaVersion) Values() []any {
	return []any{sv.Version, sv.TimestampUTC}
}

type graph struct {
	ID                 incr.Identifier
	StabilizationNum   uint64
	NumNodes           uint64
	NumNodesRecomputed uint64
	NumNodesChanged    uint64
}

type node struct {
	ID            incr.Identifier
	GraphID       incr.Identifier
	Height        int
	SetAt         uint64
	RecomputedAt  uint64
	ChangedAt     uint64
	Kind          string
	Value         []byte
	Input0        incr.Identifier
	Input1        incr.Identifier
	Input2        incr.Identifier
	Input3        incr.Identifier
	NumRecomputes uint64
	NumChanges    uint64
}
