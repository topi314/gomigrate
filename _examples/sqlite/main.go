package main

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"

	"github.com/topi314/gomigrate"
	"github.com/topi314/gomigrate/drivers/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	// open database
	db, err := sql.Open("sqlite", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// run migrations
	if err = gomigrate.Migrate(ctx, db, sqlite.New, migrations,
		gomigrate.WithDirectory("migrations"), // set directory for migrations
		gomigrate.WithTableName("gomigrate"),  // set custom table name for migrations
		gomigrate.WithLogger(slog.Default()),  // set custom logger
	); err != nil {
		log.Fatal(err)
	}

	// run your code
}
