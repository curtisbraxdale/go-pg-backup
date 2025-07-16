package pgrestore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/curtisbraxdale/go-pg-backup/internal/pgbackup"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestRestoreProcess runs an integration test for the restore process.
func TestRestoreProcess(t *testing.T) {
	ctx := context.Background()

	// Define the PostgreSQL container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb", // Initial DB
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
	user := "testuser"
	password := "testpassword"
	sourceDbName := "testdb"
	restoredDbName := "restoredb"

	// Prepare the source database with some data
	if err := prepareSourceDatabase(t, host, port, user, password, sourceDbName); err != nil {
		t.Fatalf("failed to prepare source database: %s", err)
	}

	// Create a backup of the source database
	backupDir, err := os.MkdirTemp("", "backups")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(backupDir)
	backupPath := filepath.Join(backupDir, "backup.sql")

	backupCmd, err := pgbackup.PreparePgDumpCommand(host, port, user, password, sourceDbName, backupPath)
	if err != nil {
		t.Fatalf("failed to prepare pg_dump command: %s", err)
	}
	if output, err := backupCmd.CombinedOutput(); err != nil {
		t.Fatalf("pg_dump failed: %s: %s", err, string(output))
	}

	// Create a new database to restore into
	if err := CreateNewDB(host, port, user, password, restoredDbName); err != nil {
		t.Fatalf("failed to create new database: %s", err)
	}

	// Run the restore
	restoreCmd, err := PreparePgRestoreCommand(host, port, user, password, restoredDbName, backupPath)
	if err != nil {
		t.Fatalf("failed to prepare psql command: %s", err)
	}
	if output, err := restoreCmd.CombinedOutput(); err != nil {
		t.Fatalf("psql failed: %s: %s", err, string(output))
	}

	// Verify the restored database
	if err := verifyRestoredDatabase(t, host, port, user, password, restoredDbName); err != nil {
		t.Fatalf("failed to verify restored database: %s", err)
	}

	t.Log("Restore process completed and verified successfully.")
}

// prepareSourceDatabase connects to the source database, creates a table, and inserts data.
func prepareSourceDatabase(t *testing.T, host string, port int, user, password, dbname string) error {
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
		CREATE TABLE restore_test_table (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL
		);
		INSERT INTO restore_test_table (name) VALUES ('restored_data_1'), ('restored_data_2');
	`)
	if err != nil {
		return fmt.Errorf("failed to create table and insert data: %w", err)
	}

	t.Log("Source database prepared successfully.")
	return nil
}

// verifyRestoredDatabase connects to the restored database and checks for the restored data.
func verifyRestoredDatabase(t *testing.T, host string, port int, user, password, dbname string) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to restored database: %w", err)
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
		return fmt.Errorf("failed to ping restored database: %w", err)
	}

	// Check if the table exists and has the correct data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM restore_test_table WHERE name LIKE 'restored_data_%'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query restored table: %w", err)
	}

	if count != 2 {
		return fmt.Errorf("expected 2 rows in restored table, but found %d", count)
	}

	t.Log("Restored database verified successfully.")
	return nil
}
