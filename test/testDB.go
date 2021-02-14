package test

import (
	"database/sql"
	"strings"
	"testing"

	// Postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/config"
)

// SetupTX Sets up a database transaction to be used in tests. DbTidy will
// Rollback the tx after test func returns.
func SetupTX(t *testing.T) (tx *sql.Tx, dbTidy func()) {
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

	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("Unable to begin tx. %v", err)
	}

	dbTidy = func() {
		tx.Rollback()
		db.Close()
	}

	return tx, dbTidy
}

// SetupDB sets up test db. To be used in tests that setup a TX.
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

// CreateArticleTable creates articles table for tests.
func CreateArticleTable(t *testing.T, db article.Executor) {
	t.Helper()
	_, err := db.Exec(`
		create table if not exists articles(
			id bigserial unique primary key,
			art_id varchar unique not null,
			name varchar unique not null,
			stock int default 0
		)
	`)
	if err != nil {
		t.Fatalf("Unable to create articles table. %v", err)
	}
}

// CreateProductTables creates product tables for test.
func CreateProductTables(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		`create table if not exists articles(
			id bigserial unique primary key,
			art_id varchar unique not null,
			name varchar unique not null,
			stock int default 0
		)`,
		`create table if not exists products(
			id bigserial unique primary key,
			barcode varchar unique not null,
			name varchar unique not null
		)`,
		`create table if not exists product_articles(
			id bigserial unique primary key,
			amount int not null,
			product_id bigint not null references products(id),
			article_id bigint not null references articles(id)
		)`,
	}

	for _, s := range stmts {
		_, err := db.Exec(s)
		if err != nil {
			t.Fatalf("Failed to create tables for product tests. %v", err)
		}
	}
}
