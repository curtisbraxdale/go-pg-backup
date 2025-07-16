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
