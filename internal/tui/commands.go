package tui

import (
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/curtisbraxdale/go-pg-backup/internal/pgbackup"
	"github.com/curtisbraxdale/go-pg-backup/internal/pgrestore"
)

// RunPgDumpCmd prepares and executes the pg_dump command in a goroutine,
// returning a message when it's finished.
func RunPgDumpCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		host := m.inputs[0].Value()
		user := m.inputs[1].Value()
		password := m.inputs[2].Value()
		dbname := m.inputs[3].Value()
		backupDir := m.inputs[4].Value()
		port := 5432 // Default PostgreSQL port

		// Construct a unique filename for the backup
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("%s-backup-%s.sql", dbname, timestamp)
		outputPath := filepath.Join(backupDir, filename)

		// First, create the destination directory if it doesn't exist.
		if err := pgbackup.CreateDestinationDir(backupDir); err != nil {
			return PgDumpFinishedMsg{Err: fmt.Errorf("failed to create backup directory: %w", err)}
		}

		// Prepare the pg_dump command
		cmd, err := pgbackup.PreparePgDumpCommand(host, port, user, password, dbname, outputPath)
		if err != nil {
			return PgDumpFinishedMsg{Err: err}
		}

		// Run the command.
		if output, err := cmd.CombinedOutput(); err != nil {
			return PgDumpFinishedMsg{Err: fmt.Errorf("pg_dump failed: %s: %w", string(output), err)}
		}

		// If we reach here, the backup was successful.
		return PgDumpFinishedMsg{OutputPath: outputPath, Err: nil}
	}
}

// RunPgRestoreCmd prepares and executes the psql command for restoring a database.
func RunPgRestoreCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		host := m.inputs[0].Value()
		user := m.inputs[1].Value()
		password := m.inputs[2].Value()
		dbname := m.inputs[3].Value()
		backupPath := m.inputs[4].Value()
		port := 5432 // Default PostgreSQL port

		// If the user wants to create a new database, do that first.
		if m.restoreNewDB {
			err := pgrestore.CreateNewDB(host, port, user, password, dbname)
			if err != nil {
				return PgRestoreFinishedMsg{Err: fmt.Errorf("failed to create new database: %w", err)}
			}
		}

		// Prepare the pg_restore command
		cmd, err := pgrestore.PreparePgRestoreCommand(host, port, user, password, dbname, backupPath)
		if err != nil {
			return PgRestoreFinishedMsg{Err: err}
		}

		// Run the command
		if output, err := cmd.CombinedOutput(); err != nil {
			return PgRestoreFinishedMsg{Err: fmt.Errorf("pg_restore failed: %s: %w", string(output), err)}
		}

		return PgRestoreFinishedMsg{Err: nil}
	}
}
