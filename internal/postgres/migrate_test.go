package postgres

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mtekmir/warehouse-service/test"
)

func TestParseMigration(t *testing.T) {
	t.Parallel()
	dir := "."
	filename := "V001__some_migration.sql"
	script := "Some sql script"
	file, err := os.Create(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer os.Remove(fmt.Sprintf("%s/%s", dir, filename))
	_, err = file.Write([]byte(script))
	if err != nil {
		t.Error(err.Error())
		return
	}

	m, err := parseMigration(dir, filename)
	expected := migration{version: 1, description: "some migration", filename: filename, script: script}
	if m != expected {
		t.Errorf("Expected: %v\n", expected)
		t.Errorf("Got: %v\n", m)
	}
}

func TestParseMigrations(t *testing.T) {
	t.Parallel()
	dir := "./testmigrations"
	script := "bla bla bla"
	filenames := []string{"V001__some_migration.sql", "V002__another_migration.sql"}
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		os.RemoveAll(dir)
		os.Remove(dir)
	}()

	file, err := os.Create("./testmigrations/somefile.sql")
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = parseMigrations(dir)
	if err == nil {
		t.Error("Invalid file name should return error")
	}
	os.Remove(file.Name())
	for _, f := range filenames {
		file, err := os.Create(fmt.Sprintf("%s/%s", dir, f))
		if err != nil {
			t.Fatal(err.Error())
		}
		file.Write([]byte(script))
	}
	mm, err := parseMigrations(dir)
	m0 := migration{version: 1, description: "some migration", filename: filenames[0], script: script}
	m1 := migration{version: 2, description: "another migration", filename: filenames[1], script: script}

	if mm[0] != m0 {
		t.Errorf("Expected: %v\n", m0)
		t.Errorf("Got: %v\n", mm[0])
	}
	if mm[1] != m1 {
		t.Errorf("Expected: %v\n", m1)
		t.Errorf("Got: %v\n", mm[1])
	}
}

func TestCreateVersTable(t *testing.T) {
	t.Parallel()
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()

	err := createVersTable(db)
	if err != nil {
		t.Fatalf("Error while creating schema_version table. Err: %s", err.Error())
	}

	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM pg_tables
			WHERE  schemaname = $1
			AND    tablename  = 'schema_version'
		)
	`, strings.ToLower(t.Name())).Scan(&tableExists)

	if err != nil {
		t.Errorf("Error while querying for schema_version table. Err: %s", err.Error())
	}
	if !tableExists {
		t.Errorf("schema_version table is not created")
	}
}

func TestExecuteMigration(t *testing.T) {
	t.Parallel()
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()
	err := createVersTable(db)
	if err != nil {
		t.Fatalf("Error while creating schema_version table. Err: %s", err.Error())
	}

	mm, err := parseMigrations("testdata")
	if err != nil {
		t.Fatalf("Error while parsing test migration files. Err: %s", err.Error())
	}
	for _, m := range mm {
		err = executeMigration(db, m)
		if err != nil {
			t.Fatalf("Error while executing. Err: %s", err.Error())
		}
	}
	// check applied migrations
	tablesToCheck := []string{"products", "articles"}
	for _, n := range tablesToCheck {
		var tableExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM pg_tables
				WHERE  schemaname = $1
				AND    tablename  = $2
			)
		`, strings.ToLower(t.Name()), n).Scan(&tableExists)
		if err != nil {
			t.Errorf("Error while querying %s table. Err: %s", n, err.Error())
		}
		if !tableExists {
			t.Errorf("Migration didn't run")
		}
	}

	// Check schema_version table
	rows, err := db.Query("SELECT version, description, filename FROM schema_version")
	if err != nil {
		t.Errorf("Error while querying schema_version table. Err: %s", err.Error())
	}
	defer rows.Close()
	appliedMM := []migration{}
	for rows.Next() {
		var m migration
		if err := rows.Scan(&m.version, &m.description, &m.filename); err != nil {
			t.Errorf("Error while scanning schema_version rows. Err: %s", err.Error())
		}
		appliedMM = append(appliedMM, m)
	}

	for i, a := range appliedMM {
		if a.version != mm[i].version {
			t.Errorf("Expected version to be: %d. Got: %d", mm[i].version, a.version)
		}
		if a.description != mm[i].description {
			t.Errorf("Expected description to be: %s. Got: %s", mm[i].description, a.description)
		}
		if a.filename != mm[i].filename {
			t.Errorf("Expected filename to be: %s. Got: %s", mm[i].filename, a.filename)
		}
	}
}

func TestMigrationsToExecute(t *testing.T) {
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()
	err := createVersTable(db)
	if err != nil {
		t.Fatalf("Error while creating schema_version table. Err: %s", err.Error())
	}

	mm, err := parseMigrations("testdata")
	if err != nil {
		t.Fatalf("Error while parsing test migration files. Err: %s", err.Error())
	}

	mm, err = migrationsToExecute(db, mm)
	if err != nil {
		t.Fatalf("Error while finding migrations to execute. Err: %s", err.Error())
	}
	if len(mm) != 2 {
		t.Errorf("Number of migrations to execute is not correct. Expected 2. Got %d", len(mm))
	}

	_, err = db.Exec(`
		INSERT INTO schema_version (version, description, filename, applied_on)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`, 3, "some migration", "V003_bla.sql")
	if err != nil {
		t.Fatalf("Error while inserting dummy version. Err: %s", err.Error())
	}

	mm, err = migrationsToExecute(db, mm)
	if err != nil {
		t.Fatalf("Error while finding migrations to execute. Err: %s", err.Error())
	}
	if len(mm) != 0 {
		t.Errorf("Number of migrations to execute is not correct. Expected 0. Got %d", len(mm))
	}
}

func TestMigrate(t *testing.T) {
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()
	err := Migrate(db, "nonexistent")
	if err == nil {
		t.Errorf("Expected to get ENOTFOUND. Got nil.")
	}

	err = Migrate(db, "testdata")

	tablesToCheck := []string{"products", "articles"}
	for _, n := range tablesToCheck {
		var tableExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM pg_tables
				WHERE  schemaname = $1
				AND    tablename  = $2
			)
		`, strings.ToLower(t.Name()), n).Scan(&tableExists)
		if err != nil {
			t.Errorf("Error while querying %s table. Err: %s", n, err.Error())
		}
		if !tableExists {
			t.Errorf("Migration didn't run")
		}
	}

	var currentV int
	db.QueryRow("SELECT version FROM schema_version ORDER BY id DESC LIMIT 1").Scan(&currentV)
	if currentV != 2 {
		t.Errorf("Expected schema version to be 2. Got %d", currentV)
	}
	mm, err := parseMigrations("testdata")
	if err != nil {
		t.Fatalf("Error while parsing test migration files. Err: %s", err.Error())
	}

	mm, err = migrationsToExecute(db, mm)
	if err != nil {
		t.Fatalf("Error while finding migrations to execute. Err: %s", err.Error())
	}
	if len(mm) != 0 {
		t.Errorf("Expected migrations to execute to be 0 after running migrations. Got %d", len(mm))
	}
}
