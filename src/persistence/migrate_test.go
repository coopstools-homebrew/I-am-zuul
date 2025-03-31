package persistence_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/coopstools-homebrew/I-am-zuul/src/config"
	"github.com/coopstools-homebrew/I-am-zuul/src/persistence"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB     *sql.DB
	testConfig *config.Config
)

func TestMain(m *testing.M) {
	// Set up the test environment
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "test",
			},
			WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())
			}),
		},
		Started: true,
	})
	if err != nil {
		fmt.Printf("Failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer pgContainer.Terminate(ctx)

	// Get the container's host and port
	host, err := pgContainer.Host(ctx)
	if err != nil {
		fmt.Printf("Failed to get container host: %v\n", err)
		os.Exit(1)
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Printf("Failed to get container port: %v\n", err)
		os.Exit(1)
	}

	// Create test config
	testConfig = &config.Config{
		DatabaseURL: fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port()),
	}

	// Connect to the test database
	testDB, err = sql.Open("postgres", testConfig.DatabaseURL)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	// Run migrations
	err = persistence.Migrate(testDB)
	if err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Run the tests
	code := m.Run()

	os.Exit(code)
}

func TestAddUsers(t *testing.T) {

	// Insert version
	_, err := testDB.Exec(
		`INSERT INTO users (id, org_id, avatar_url, email) VALUES ($1, $2, $3, $4)`,
		"123", "123", "https://example.com/avatar.png", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert version: %v", err)
	}

	// Retrieve version
	var user struct {
		ID        int32
		OrgID     string
		AvatarURL string
		Email     string
	}
	err = testDB.QueryRow(
		`SELECT id, org_id, avatar_url, email FROM users ORDER BY created_at DESC LIMIT 1`).
		Scan(&user.ID, &user.OrgID, &user.AvatarURL, &user.Email)
	if err != nil {
		t.Fatalf("Failed to retrieve version: %v", err)
	}

	assert.Equal(t, user.ID, int32(123))
	assert.Equal(t, user.OrgID, "123")
	assert.Equal(t, user.AvatarURL, "https://example.com/avatar.png")
	assert.Equal(t, user.Email, "test@example.com")
}
