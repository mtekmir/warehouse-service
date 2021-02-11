package test

import (
	"database/sql"
	"strings"
	"testing"

	// Postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/mtekmir/warehouse-service/internal/config"
)

// SetupDB sets up test db.
func SetupDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	conf, err := config.Parse()
	if err != nil {
		t.Fatalf("Unable to parse config. %v", err)
		return nil, func() {}
	}
	
	db, err := sql.Open("pgx", conf.DBURL)
	if err != nil {
		t.Fatalf("Failed to initialize db. Err: %s", err.Error())
	}

	schemaName := strings.ToLower(t.Name())

	dbTidy := func() {
		_, err := db.Exec("DROP SCHEMA " + schemaName + " CASCADE")
		if err != nil {
			t.Fatalf("Error during db cleanup. Err: %s", err.Error())
		}
		db.Close()
	}

	// create test schema
	_, err = db.Exec("CREATE SCHEMA " + schemaName)
	if err != nil {
		t.Fatalf("Error while creating the schema. Err: %s", err.Error())
	}

	// use schema
	_, err = db.Exec("SET search_path TO " + schemaName)
	if err != nil {
		t.Fatalf("Error while switching to schema. Err: %s", err.Error())
	}

	return db, dbTidy
}
