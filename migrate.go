package gomigrate

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"
)

const (
	defaultDirectory         = "migrations"
	defaultTableName         = "gomigrate"
	migrationFileExt         = ".sql"
	migrationSeparator       = "_"
	migrationDriverSeparator = "."
)

var (
	// ErrNoDatabase is returned when no database is provided.
	ErrNoDatabase = errors.New("no database provided")

	// ErrNoDriver is returned when no driver is provided.
	ErrNoDriver = errors.New("no driver provided")
)

func wrapMigrateErr(name string, fileName string, version int, err error) error {
	return &MigrateError{
		Name:     name,
		FileName: fileName,
		Version:  version,
		Err:      err,
	}
}

// MigrateError holds additional information about an error that occurred during migration.
type MigrateError struct {
	Name     string
	FileName string
	Version  int
	Err      error
}

// Error returns the error message.
func (e *MigrateError) Error() string {
	return fmt.Sprintf("migration %s failed: %v", e.Name, e.Err)
}

// Unwrap returns the underlying error.
func (e *MigrateError) Unwrap() error {
	return e.Err
}

// Queryer is used to execute SQL queries and transactions.
type Queryer interface {
	// QueryContext executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// ExecContext executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	// BeginTx starts a transaction.
	//
	// The provided context is used until the transaction is committed or rolled back.
	// If the context is canceled, the sql package will roll back
	// the transaction. [Tx.Commit] will return an error if the context provided to
	// BeginTx is canceled.
	//
	// The provided [sql.TxOptions] is optional and may be nil if defaults should be used.
	// If a non-default isolation level is used that the driver doesn't support,
	// an error will be returned.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Migrate reads and executes SQL migrations from the embed.FS to bring the database schema up to date.
// It keeps track of the executed migrations in a table.
// If the database schema is ahead of the migrations, it will return an error.
// Each migration runs in a transaction. If the context is canceled, the transaction for the current migration will be rolled back and it will return an error.
func Migrate(ctx context.Context, db Queryer, newDriver NewDriver, fs embed.FS, opts ...Option) error {
	if db == nil {
		return ErrNoDatabase
	}
	if newDriver == nil {
		return ErrNoDriver
	}

	// Apply options to the default configuration.
	cfg := defaultConfig()
	cfg.apply(opts...)

	// initialize the driver
	driver := newDriver(db, cfg.TableName)

	// Load migrations from the embed.FS and sort them by version.
	migrations, err := loadMigrations(fs, cfg.Directory, driver.Name())
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// If there are no migrations, we should return an error.
	if len(migrations) == 0 {
		return fmt.Errorf("no migrations found in %s", cfg.Directory)
	}

	// create the version table if it does not exist.
	if err = driver.CreateVersionTable(ctx); err != nil {
		return fmt.Errorf("failed to create version table: %w", err)
	}

	// Get the most recent schema version.
	currentVersion, err := driver.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	lastMigration := migrations[len(migrations)-1]
	// If the current version is the same as the last migration version, there is nothing to do.
	if currentVersion == lastMigration.version {
		return nil
	}

	// If the current version is ahead of the last migration version, we should return an error since downgrading is not supported.
	if currentVersion > lastMigration.version {
		return fmt.Errorf("schema version is ahead of migrations: current=%d, latest=%d", currentVersion, lastMigration.version)
	}

	// Execute migrations.
	for _, m := range migrations {
		// Skip migrations that have already been executed.
		if m.version <= currentVersion {
			continue
		}

		if err = execMigration(ctx, db, driver, m, fs); err != nil {
			return wrapMigrateErr(m.name, m.filePath, m.version, err)
		}
	}

	return nil
}

type migration struct {
	name     string
	version  int
	driver   string
	filePath string
}

func loadMigrations(fs embed.FS, dir string, driver string) ([]migration, error) {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory '%s': %w", dir, err)
	}

	var migrations []migration
outer:
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(fileName, migrationFileExt) {
			continue
		}

		mig, err := parseMigrationFileName(dir, fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration file name: %w", err)
		}

		if mig == nil || (mig.driver != "" && mig.driver != driver) {
			continue
		}

		for i, m := range migrations {
			if m.version == mig.version {
				if m.driver == mig.driver {
					return nil, fmt.Errorf("duplicate migration version and driver: version=%d, driver1=%s, driver2=%s", mig.version, m.driver, mig.driver)
				}

				if mig.driver == driver {
					migrations[i] = *mig
				}
				continue outer
			}
		}

		migrations = append(migrations, *mig)
	}

	slices.SortFunc(migrations, func(m1 migration, m2 migration) int {
		return m1.version - m2.version
	})

	return migrations, nil
}

func parseMigrationFileName(dir string, fileName string) (*migration, error) {
	name, ok := strings.CutSuffix(fileName, migrationFileExt)
	if !ok {
		return nil, fmt.Errorf("invalid migration file extension: %s", fileName)
	}

	var driver string
	parts := strings.SplitN(name, migrationDriverSeparator, 2)
	if len(parts) > 1 {
		driver = parts[1]
		name = parts[0]
	}

	parts = strings.SplitN(name, migrationSeparator, 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid migration file name: %s", fileName)
	}

	name = strings.ReplaceAll(parts[1], "_", " ")

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse migration version: %w", err)
	}

	return &migration{
		name:     name,
		version:  version,
		driver:   driver,
		filePath: path.Join(dir, fileName),
	}, nil
}

func execMigration(ctx context.Context, db Queryer, driver Driver, m migration, fs embed.FS) error {
	data, err := fs.ReadFile(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, string(data))
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	if err = driver.AddVersion(ctx, tx, m.version); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to set version: %w", err)
	}

	return tx.Commit()
}
