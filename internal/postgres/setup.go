package postgres

import (
	"database/sql"

	// Postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

// Setup sets up the db runs migrations and returns a func to close it.
func Setup(log *logrus.Logger, dbURL, migrationsPath string) (*sql.DB, func(), error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Connected to DB")

	dbTidy := func() { db.Close() }

	if err := Migrate(log, db, migrationsPath); err != nil {
		return nil, dbTidy, err
	}

	return db, dbTidy, nil
}
