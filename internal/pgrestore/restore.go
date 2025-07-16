package pgrestore

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	_ "github.com/lib/pq"
)

// PreparePgRestoreCommand prepares the exec.Cmd for psql to restore a database.
func PreparePgRestoreCommand(host string, port int, user, password, dbname, backupPath string) (*exec.Cmd, error) {
	_, err := exec.LookPath("psql")
	if err != nil {
		return nil, fmt.Errorf("psql not found in system PATH. Please ensure PostgreSQL client tools are installed and in your PATH.")
	}

	args := []string{
		"-h", host,
		"-U", user,
		"-d", dbname,
		"-f", backupPath,
	}

	if port != 0 {
		args = append(args, "-p", fmt.Sprintf("%d", port))
	}

	cmd := exec.Command("psql", args...)

	if password != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
	}

	return cmd, nil
}

// CreateNewDB creates a new database.
func CreateNewDB(host string, port int, user, password, dbname string) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		return fmt.Errorf("failed to create new database: %w", err)
	}

	return nil
}
