[![Go Reference](https://pkg.go.dev/badge/github.com/topi314/gomigrate.svg)](https://pkg.go.dev/github.com/topi314/gomigrate)
[![Go Report](https://goreportcard.com/badge/github.com/topi314/gomigrate)](https://goreportcard.com/report/github.com/topi314/gomigrate)
[![Go Version](https://img.shields.io/github/go-mod/go-version/topi314/gomigrate?filename=go.mod)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/topi314/gomigrate/blob/master/LICENSE)
[![GoMigrate Version](https://img.shields.io/github/v/release/topi314/gomigrate?label=release)](https://github.com/topi314/gomigrate/releases/latest)

# gomigrate

GoMigrate is a SQL migration library for Go. It can support multiple databases such as SQLite, PostgreSQL, MySQL, and others.

## Table of Contents

<details>
<summary>Click to expand</summary>

- [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installing](#installing)
- [Usage](#usage)
    - [Create a migrations](#create-a-migrations)
    - [Run migrations](#run-migrations)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

</details>

## Getting Started

### Prerequisites

* Go 1.23 or later
* SQLite, PostgreSQL, MySQL, or other databases
* Go driver for the database you want to use

### Installing

```sh
go get github.com/topi314/gomigrate
```

## Usage

### Create a migrations

Create a new folder named `migrations` and create a file with the following naming convention `VERSION_NAME.DRIVER.sql` where `VERSION` is a number, `NAME` is a name of the migration & `DRIVER` is the name of the database driver this migration is for.
You can omit the `DRIVER` part if you want to use the same migration for all drivers.
As an example: `01_create_users_table.sql`, `01_create_users_table.postgres.sql`, `02_add_email_to_users_table.sql` or `02_add_email_to_users_table.sqlite.sql`.

`01_create_users_table.sql`

```sql
-- create users table
CREATE TABLE users
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL
);

```

`01_create_users_table.postgres.sql`

```sql
-- create users table for PostgreSQL
CREATE TABLE users
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL
);

```

`02_add_email_to_users_table.sql`

```sql
-- add email column to users table
ALTER TABLE users
    ADD COLUMN email VARCHAR;
```

`02_add_email_to_users_table.sqlite.sql`

```sql
-- add email column to users table for SQLite
ALTER TABLE users
    ADD COLUMN email VARCHAR;
```

It should look like this:

```
migrations/
├─ 01_create_users_table.sql
├─ 01_create_users_table.postgres.sql
├─ 02_add_email_to_users_table.sql
├─ 02_add_email_to_users_table.sqlite.sql
```

### Run migrations

Now you can run the migrations in your Go code. Here is an example for SQLite:

```go
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
```

## Examples

You can find examples under

* sqlite: [_example](_examples/sqlite/main.go)
* postgres: [_example](_examples/postgres/main.go)

## Troubleshooting

For help feel free to open an issue.

## Contributing

Contributions are welcomed but for bigger changes please first create an issue to discuss your intentions and ideas.

## License

Distributed under the [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE). See LICENSE for more information.
