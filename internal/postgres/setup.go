package postgres

import (
	"database/sql"
	"log"

	// Postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Setup sets up the db runs migrations and returns a func to close it.
func Setup(dbURL, migrationsPath string) (*sql.DB, func(), error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Connected to DB")

	dbTidy := func() { db.Close() }

	if err := Migrate(db, migrationsPath); err != nil {
		return nil, dbTidy, err
	}

	return db, dbTidy, nil
}
