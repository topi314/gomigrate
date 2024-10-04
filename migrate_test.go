package gomigrate

import (
	"embed"
	"fmt"
	"testing"
)

//go:embed testdata
var testMigrations embed.FS

func TestLoadMigrations(t *testing.T) {
	migrations, err := loadMigrations(testMigrations, "testdata", "postgres")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got: %d", len(migrations))
	}

	expected := []migration{
		{
			name:     "initial",
			version:  1,
			driver:   "",
			filePath: "testdata/1_initial.sql",
		},
		{
			name:     "some changes",
			version:  2,
			driver:   "postgres",
			filePath: "testdata/2_some_changes.postgres.sql",
		},
	}

	for i, mig := range migrations {
		if mig.name != expected[i].name {
			t.Errorf("expected name: %s, got: %s", expected[i].name, mig.name)
		}

		if mig.version != expected[i].version {
			t.Errorf("expected version: %d, got: %d", expected[i].version, mig.version)
		}

		if mig.driver != expected[i].driver {
			t.Errorf("expected driver: %s, got: %s", expected[i].driver, mig.driver)
		}

		if mig.filePath != expected[i].filePath {
			t.Errorf("expected filePath: %s, got: %s", expected[i].filePath, mig.filePath)
		}
	}
}

func TestParseMigrationFileName(t *testing.T) {
	data := []struct {
		dir      string
		fileName string
		expected *migration
		err      error
	}{
		{
			dir:      "migrations",
			fileName: "1_create_table.sql",
			expected: &migration{
				name:     "create table",
				version:  1,
				driver:   "",
				filePath: "migrations/1_create_table.sql",
			},
			err: nil,
		},
		{
			dir:      "migrations",
			fileName: "2_add_column_to_table.sql",
			expected: &migration{
				name:     "add column to table",
				version:  2,
				driver:   "",
				filePath: "migrations/2_add_column_to_table.sql",
			},
			err: nil,
		},
		{
			dir:      "migrations",
			fileName: "3_add_column_to_table_and_index.postgres.sql",
			expected: &migration{
				name:     "add column to table and index",
				version:  3,
				driver:   "postgres",
				filePath: "migrations/3_add_column_to_table_and_index.postgres.sql",
			},
			err: nil,
		},
		{
			dir:      "migrations",
			fileName: "4_add_column_to_table_and_index_2.sqlite.sql",
			expected: &migration{
				name:     "add column to table and index 2",
				version:  4,
				driver:   "sqlite",
				filePath: "migrations/4_add_column_to_table_and_index_2.sqlite.sql",
			},
			err: nil,
		},
		{
			dir:      "migrations",
			fileName: "5_add_column_to_table_and_index_3.sqll",
			expected: nil,
			err:      fmt.Errorf("invalid migration file extension: 5_add_column_to_table_and_index_3.sqll"),
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("Case_%d", i), func(t *testing.T) {
			mig, err := parseMigrationFileName(d.dir, d.fileName)
			if err != nil {
				if d.err == nil {
					t.Errorf("unexpected error: %s", err)
				} else if err.Error() != d.err.Error() {
					t.Errorf("expected error: %s, got: %s", d.err, err)
				}
			}

			if mig == nil {
				if d.expected != nil {
					t.Errorf("expected migration: %+v, got: %+v", d.expected, mig)
				}
				return
			}

			if mig.name != d.expected.name {
				t.Errorf("expected name: %s, got: %s", d.expected.name, mig.name)
			}

			if mig.version != d.expected.version {
				t.Errorf("expected version: %d, got: %d", d.expected.version, mig.version)
			}

			if mig.driver != d.expected.driver {
				t.Errorf("expected driver: %s, got: %s", d.expected.driver, mig.driver)
			}

			if mig.filePath != d.expected.filePath {
				t.Errorf("expected filePath: %s, got: %s", d.expected.filePath, mig.filePath)
			}
		})
	}
}
