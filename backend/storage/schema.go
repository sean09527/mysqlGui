package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// InitLocalDB initializes the SQLite database for local storage
func InitLocalDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

// createTables creates all necessary tables for local storage
func createTables(db *sql.DB) error {
	schemas := []string{
		// Connection profiles table
		`CREATE TABLE IF NOT EXISTS connection_profiles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			host TEXT NOT NULL,
			port INTEGER NOT NULL,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			database TEXT,
			charset TEXT DEFAULT 'utf8mb4',
			timeout INTEGER DEFAULT 10,
			ssh_enabled INTEGER DEFAULT 0,
			ssh_host TEXT,
			ssh_port INTEGER,
			ssh_username TEXT,
			ssh_password TEXT,
			ssh_key_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Application settings table
		`CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Operation logs table
		`CREATE TABLE IF NOT EXISTS operation_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			level TEXT NOT NULL,
			operation TEXT NOT NULL,
			message TEXT,
			details TEXT,
			connection_id TEXT
		)`,

		// Query history table
		`CREATE TABLE IF NOT EXISTS query_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			connection_id TEXT NOT NULL,
			database TEXT,
			sql TEXT NOT NULL,
			execution_time INTEGER,
			rows_affected INTEGER,
			success INTEGER
		)`,

		// Create indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_operation_logs_timestamp 
		 ON operation_logs(timestamp)`,

		`CREATE INDEX IF NOT EXISTS idx_query_history_timestamp 
		 ON query_history(timestamp)`,

		`CREATE INDEX IF NOT EXISTS idx_query_history_connection 
		 ON query_history(connection_id)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}
