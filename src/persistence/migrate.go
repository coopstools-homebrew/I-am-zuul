package persistence

import (
	"database/sql"
	"embed"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/coopstools-homebrew/I-am-zuul/src/persistence/queries"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db *sql.DB) (*Migrator, error) {
	return &Migrator{db: db}, nil
}

// getMigrations returns the ordered list of migration script filenames
func (m *Migrator) getMigrations() ([]string, error) {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	// Pre-allocate the array to the number of migrations
	migrationFiles := make([]string, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			// Get everything before the first underscore
			parts := strings.Split(entry.Name(), "_")
			if len(parts) < 2 {
				continue // Skip files without underscore
			}

			// Convert prefix to integer
			prefix, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, err
			}

			// Store filename at index matching its prefix
			migrationFiles[prefix] = "migrations/" + entry.Name()
		}
	}

	return migrationFiles, nil
}

// initAndGetVersion ensures the version tracking table exists and returns current version
func (m *Migrator) initAndGetVersion(initialMigration string) (int, error) {
	// Always run the init script first since it's idempotent
	initialMigrationFile, err := migrations.Open(initialMigration)
	if err != nil {
		return 0, errors.Wrap(err, "error opening initial migration file")
	}
	initialMigrationContent, err := io.ReadAll(initialMigrationFile)
	if err != nil {
		return 0, errors.Wrap(err, "error reading initial migration content")
	}
	_, err = m.db.Exec(string(initialMigrationContent))
	if err != nil {
		return 0, errors.Wrapf(err, "error executing initial migration:\n\n%s\n\n", string(initialMigrationContent))
	}

	// Get current version
	var currentVersion int
	err = m.db.QueryRow(queries.GET_VERSION).Scan(&currentVersion)
	if err != nil {
		return 0, errors.Wrap(err, "error getting current version")
	}

	return currentVersion, nil
}

// applyMigrations applies any pending migrations in order
func (m *Migrator) applyMigrations(currentVersion int, migrationFiles []string) error {
	if currentVersion >= len(migrationFiles) {
		log.Printf("db is on version %d, no migrations to apply", currentVersion)
		return nil
	}
	log.Printf("db is on version %d, applying migrations", currentVersion)
	// Execute migration within transaction
	tx, err := m.db.Begin()
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}
	for version, migration := range migrationFiles[1:] {
		err = m.runSingleMigration(tx, version+1, migration)
		if err != nil {
			tx.Rollback()
			return errors.Wrap(err, "error running single migration")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "error committing transaction")
	}
	return nil
}

func (m *Migrator) runSingleMigration(tx *sql.Tx, version int, migration string) error {
	if migration == "" {
		log.Printf("version %d not found: skipping", version)
		return nil
	}

	log.Printf("applying migration to version %d: %s", version, migration)

	migrationFile, err := migrations.Open(migration)
	if err != nil {
		return errors.Wrap(err, "error opening migration file")
	}
	migrationContent, err := io.ReadAll(migrationFile)
	if err != nil {
		return errors.Wrap(err, "error reading migration content")
	}

	_, err = tx.Exec(string(migrationContent))
	if err != nil {
		return errors.Wrap(err, "error executing migration")
	}

	// Update version
	_, err = tx.Exec(queries.INSERT_VERSION, version)
	if err != nil {
		return errors.Wrap(err, "error updating version")
	}
	return nil
}

// Migrate runs all pending migrations
func Migrate(db *sql.DB) error {
	log.Printf("migrating database")
	migrator, err := NewMigrator(db)
	if err != nil {
		return errors.Wrap(err, "error creating migrator")
	}

	migrations, err := migrator.getMigrations()
	if err != nil {
		return errors.Wrap(err, "error getting migrations")
	}

	currentVersion, err := migrator.initAndGetVersion(migrations[0])
	if err != nil {
		return errors.Wrap(err, "error initializing and getting version")
	}

	err = migrator.applyMigrations(currentVersion, migrations)
	if err != nil {
		return errors.Wrap(err, "error applying migrations")
	}
	return nil
}
