package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/topi314/gomigrate"
	"github.com/topi314/gomigrate/drivers/postgres"
)

var (
	user   = os.Getenv("POSTGRES_USER")
	pass   = os.Getenv("POSTGRES_PASSWORD")
	host   = os.Getenv("POSTGRES_HOST")
	port   = os.Getenv("POSTGRES_PORT")
	dbname = os.Getenv("POSTGRES_DB")
	secure = os.Getenv("POSTGRES_SECURE")
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	// open database
	db, err := sql.Open("pgx", fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s", user, pass, host, port, dbname, secure))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// run migrations
	if err = gomigrate.Migrate(ctx, db, postgres.New, migrations,
		gomigrate.WithDirectory("migrations"), // set directory for migrations
		gomigrate.WithTableName("gomigrate"),  // set custom table name for migrations
		gomigrate.WithLogger(slog.Default()),  // set custom logger
	); err != nil {
		log.Fatal(err)
	}

	// run your code
}
