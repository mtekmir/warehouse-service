package postgres

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type migration struct {
	version     int
	description string
	filename    string
	script      string
}

// Migrate runs db migrations. Applied db migrations will be stored in 
// schema_version table. If no new migrations it doesn't do anything.
func Migrate(log *logrus.Logger, d *sql.DB, migrationsPath string) error {
	mm, err := parseMigrations(migrationsPath)
	if err != nil {
		return err
	}

	err = createVersTable(d)
	if err != nil {
		return err
	}

	mm, err = migrationsToExecute(d, mm)
	if err != nil {
		return err
	}
	if len(mm) == 0 {
		log.Println("Db up to date")
		return nil
	}

	for _, m := range mm {
		err = executeMigration(log, d, m)
		if err != nil {
			return err
		}
	}

	return nil
}

func createVersTable(d *sql.DB) error {
	_, err := d.Exec(`
		create table if not exists schema_version(
			id serial unique primary key,
			version int not null,
			description varchar not null,
			filename varchar not null,
			applied_on timestamp
		)
	`)
	return err
}

func migrationsToExecute(d *sql.DB, migrations []migration) ([]migration, error) {
	var version int
	err := d.QueryRow("SELECT version FROM schema_version ORDER BY id DESC LIMIT 1").Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	mm := []migration{}
	for _, m := range migrations {
		if m.version > version {
			mm = append(mm, m)
		}
	}

	return mm, nil
}

func parseMigrations(migrationsPath string) ([]migration, error) {
	ff, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	var migrations []migration
	reg := "^V[0-9]{3}__[^0-9]*.sql$"
	r, _ := regexp.Compile(reg)

	for _, file := range ff {
		if !r.Match([]byte(file.Name())) {
			return nil, errors.New("Migration file names should match " + reg)
		}
		m, err := parseMigration(migrationsPath, file.Name())
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}
	return migrations, nil
}

func parseMigration(migrationsPath, filename string) (migration, error) {
	m := migration{}
	r, _ := regexp.Compile("^V([0-9]*)__([a-zA-Z_]*).sql$")
	parts := r.FindStringSubmatch(filename)

	version, err := strconv.Atoi(parts[1][1:])
	if err != nil {
		return m, nil
	}
	description := strings.Join(strings.Split(parts[2], "_"), " ")

	file, err := os.Open(path.Join(migrationsPath, filename))
	if err != nil {
		return m, err
	}

	script, err := ioutil.ReadAll(file)
	if err != nil {
		return m, err
	}

	m.version = version
	m.description = description
	m.filename = filename
	m.script = string(script)
	return m, nil
}

func executeMigration(log *logrus.Logger, d *sql.DB, m migration) error {
	log.Println("Executing migration", m.filename)
	_, err := d.Exec(m.script)
	if err != nil {
		return err
	}
	_, err = d.Exec(`
		INSERT INTO schema_version (version, description, filename, applied_on)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`, m.version, m.description, m.filename)
	if err != nil {
		return err
	}
	return nil
}
