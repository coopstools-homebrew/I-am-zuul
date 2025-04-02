package persistence

import (
	"database/sql"
	"embed"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

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
		return nil, errors.Wrap(err, "error reading migrations directory")
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
	stmt, err := m.db.Prepare(queries.GET_VERSION)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to prepare version query")
	}
	defer stmt.Close()

	var currentVersion int
	err = stmt.QueryRow().Scan(&currentVersion)
	if err != nil {
		return 0, errors.Wrap(err, "error getting current version")
	}

	return currentVersion, nil
}

// applyMigrations applies any pending migrations in order
func (m *Migrator) applyMigrations(currentVersion int, migrationFiles []string) error {
	if currentVersion >= len(migrationFiles)-1 {
		log.Printf("db is on version %d, no migrations to apply", currentVersion)
		return nil
	}
	log.Printf("db is on version %d, applying migrations", currentVersion)
	// Execute migration within transaction
	tx, err := m.db.Begin()
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}
	for version, migration := range migrationFiles {
		if version <= currentVersion {
			continue
		}
		err = m.runSingleMigration(tx, version, migration)
		if err != nil {
			tx.Rollback()
			return errors.Wrap(err, "error running single migration")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "error committing transaction")
	}

	// Get all versions and log them
	rows, err := m.db.Query(queries.GET_VERSIONS)
	if err != nil {
		return errors.Wrap(err, "error getting versions")
	}
	defer rows.Close()

	for rows.Next() {
		var version struct {
			Version   int
			CreatedAt time.Time
		}
		err = rows.Scan(&version.Version, &version.CreatedAt)
		if err != nil {
			return errors.Wrap(err, "error scanning version")
		}
		log.Printf("version %d, created at %s", version.Version, version.CreatedAt)
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
	log.Printf("migrated to version %d", version)
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
