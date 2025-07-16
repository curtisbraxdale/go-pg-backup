# TUI PostgreSQL Backup and Restore Wizard

This project is a Terminal User Interface (TUI) for backing up and restoring PostgreSQL databases. It provides a simple, guided wizard to help users through the process of creating database backups and restoring them.

## Project Summary

The TUI PostgreSQL Backup and Restore Wizard simplifies the `pg_dump` and `psql` commands by wrapping them in an interactive, user-friendly interface. This tool is designed to reduce the complexity of database backup and restore operations, making them more accessible to developers and database administrators who prefer a visual guide over raw command-line interactions.

The wizard prompts for necessary connection details and file paths, validates inputs, and executes the underlying PostgreSQL commands, providing real-time feedback on the process.

## Technologies Used

- **Go**: The application is written in Go, leveraging its strong support for concurrent operations and command-line tool development.
- **Bubble Tea**: The TUI is built using the Bubble Tea framework, a powerful library for creating sophisticated terminal applications.
- **Lipgloss**: Styling of the TUI components is handled by Lipgloss, which provides a fluent and expressive API for terminal styling.
- **Testcontainers for Go**: Integration tests are implemented using Testcontainers, which allows for ephemeral PostgreSQL instances to be spun up in Docker containers for reliable and isolated testing.

## Development Process

This project was developed with the assistance of the Zed editor's agent mode, which facilitated a rapid and iterative development process. The agent's capabilities in code generation, and interactive dialogue were instrumental in building out the features and structure of this application.

## Getting Started

To run this application, you will need to have Go installed on your system, as well as the PostgreSQL client tools (`pg_dump` and `psql`).

1. **Clone the repository:**
   ```sh
   git clone https://github.com/curtisbraxdale/go-pg-backup.git
   cd go-pg-backup
   ```
2. **Run the application:**
   ```sh
   go run main.go
   ```
This will launch the TUI wizard, and you can follow the on-screen prompts to perform a backup or restore operation.