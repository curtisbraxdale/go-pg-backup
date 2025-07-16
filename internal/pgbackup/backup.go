package pgbackup

import (
	"fmt"
	"os"
	"os/exec"
)

// PreparePgDumpCommand prepares the exec.Cmd for pg_dump but does not run it.
func PreparePgDumpCommand(host string, port int, user, password, dbname, outputPath string) (*exec.Cmd, error) {
	// Check if pg_dump is available
	_, err := exec.LookPath("pg_dump")
	if err != nil {
		return nil, fmt.Errorf("pg_dump not found in system PATH. Please ensure PostgreSQL client tools are installed and in your PATH.")
	}

	args := []string{
		"-h", host,
		"-U", user,
		"-d", dbname,
		"-f", outputPath,
		"-F", "p", // Plain text SQL format, or 'c' for custom, 'd' for directory
	}

	if port != 0 {
		args = append(args, "-p", fmt.Sprintf("%d", port))
	}

	cmd := exec.Command("pg_dump", args...)

	// Set PGPASSWORD environment variable for pg_dump
	if password != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
	}

	return cmd, nil
}

// CreateDestinationDir remains the same
func CreateDestinationDir(dir string) error {
	// ... (same as before)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755) // Read/write/execute for owner, read/execute for others
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		// fmt.Printf("Created destination directory: %s\n", dir) // Removed for TUI context
	} else if err != nil {
		return fmt.Errorf("error checking destination directory %s: %w", dir, err)
	}
	return nil
}
