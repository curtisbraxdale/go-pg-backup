package tui

// PgDumpStartedMsg indicates that pg_dump has begun.
type PgDumpStartedMsg struct{}

// PgDumpFinishedMsg indicates that pg_dump has completed, with an error if any.
type PgDumpFinishedMsg struct {
	Err        error
	OutputPath string // Path where the backup was saved
}

// PgDumpProgressMsg can be used to stream output from pg_dump (e.g., stderr for warnings).
type PgDumpProgressMsg string

// PgRestoreStartedMsg indicates that pg_restore has begun.
type PgRestoreStartedMsg struct{}

// PgRestoreFinishedMsg indicates that pg_restore has completed, with an error if any.
type PgRestoreFinishedMsg struct {
	Err error
}

// PgRestoreProgressMsg can be used to stream output from pg_restore (e.g., stderr for warnings).
type PgRestoreProgressMsg string
