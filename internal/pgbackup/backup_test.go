package pgbackup

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestBackupProcess runs an integration test for the backup process.
func TestBackupProcess(t *testing.T) {
	ctx := context.Background()

	// Define the PostgreSQL container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
	}

	// Create and start the PostgreSQL container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// Get container details
	host, err := pgContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %s", err)
	}

	mappedPort, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get mapped port: %s", err)
	}
	port := mappedPort.Int()

	// Prepare database for testing
	if err := prepareDatabase(t, host, port, "testuser", "testpassword", "testdb"); err != nil {
		t.Fatalf("failed to prepare database: %s", err)
	}

	// Define backup parameters
	backupDir, err := os.MkdirTemp("", "backups")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(backupDir)

	outputPath := filepath.Join(backupDir, "backup.sql")

	// Run the backup
	cmd, err := PreparePgDumpCommand(host, port, "testuser", "testpassword", "testdb", outputPath)
	if err != nil {
		t.Fatalf("failed to prepare pg_dump command: %s", err)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pg_dump failed: %s: %s", err, string(output))
	}

	// Verify the backup
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read backup file: %s", err)
	}

	if len(content) == 0 {
		t.Error("backup file is empty")
	}

	// You can add more specific checks here, e.g., check for table creation SQL
	t.Log("Backup created successfully and is not empty.")
}

// prepareDatabase connects to the database, creates a table, and inserts data.
func prepareDatabase(t *testing.T, host string, port int, user, password, dbname string) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Retry connection
	for i := 0; i < 5; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create a table and insert data
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL
		);
		INSERT INTO test_table (name) VALUES ('test_data_1'), ('test_data_2');
	`)
	if err != nil {
		return fmt.Errorf("failed to create table and insert data: %w", err)
	}

	t.Log("Database prepared successfully.")
	return nil
}
