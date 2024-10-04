package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/topi314/gomigrate"
)

// Name is the name of the PostgreSQL driver.
const Name = "postgres"

// New returns a new PostgreSQL driver.
func New(db gomigrate.Queryer, tableName string) gomigrate.Driver {
	return &driver{
		db:        db,
		tableName: tableName,
	}
}

type driver struct {
	db        gomigrate.Queryer
	tableName string
}

func (d *driver) Name() string {
	return Name
}

func (d *driver) CreateVersionTable(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (version INT PRIMARY KEY, date TIMESTAMP DEFAULT CURRENT_TIMESTAMP)", d.tableName))
	return err
}

func (d *driver) GetVersion(ctx context.Context) (int, error) {
	raws, err := d.db.QueryContext(ctx, fmt.Sprintf("SELECT version FROM %s ORDER BY version DESC LIMIT 1", d.tableName))
	if err != nil {
		return 0, err
	}
	defer raws.Close()

	if !raws.Next() {
		return 0, nil
	}

	var v int
	if err = raws.Scan(&v); err != nil {
		return 0, err
	}

	return v, nil
}

func (d *driver) AddVersion(ctx context.Context, tx *sql.Tx, version int) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (version) VALUES ($1)", d.tableName), version)
	if err != nil {
		return err
	}
	return nil
}
