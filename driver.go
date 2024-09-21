package gomigrate

import (
	"context"
	"database/sql"
)

// NewDriver is a function that returns a new Driver.
type NewDriver func(db Queryer, tableName string) Driver

// Driver allows gomigrate to work with different databases since.
type Driver interface {
	// CreateVersionTable creates the versioning table if it does not exist.
	CreateVersionTable(ctx context.Context) error

	// GetVersion returns the most recent schema version.
	GetVersion(ctx context.Context) (int, error)

	// AddVersion adds a new schema version to the versioning table.
	AddVersion(ctx context.Context, tx *sql.Tx, version int) error
}
