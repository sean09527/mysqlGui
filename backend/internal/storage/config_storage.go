package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"mygui/backend/types"

	_ "github.com/mattn/go-sqlite3"
	"github.com/vrischmann/userdir"
)

// ConfigStorage manages application configuration and connection profiles using SQLite
type ConfigStorage struct {
	db *sql.DB
}

// NewConfigStorage creates a new ConfigStorage instance and initializes the database
func NewConfigStorage() (*ConfigStorage, error) {
	// Get config directory
	configDir := filepath.Join(userdir.GetConfigHome(), "MyGUI")
	dbPath := filepath.Join(configDir, "config.db")

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	cs := &ConfigStorage{db: db}

	// Initialize schema
	if err := cs.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return cs, nil
}

// initSchema creates the necessary tables if they don't exist
func (cs *ConfigStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS connection_profiles (
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
	);

	CREATE TABLE IF NOT EXISTS app_settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS query_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		connection_id TEXT NOT NULL,
		database TEXT,
		sql TEXT NOT NULL,
		execution_time INTEGER,
		rows_affected INTEGER,
		success INTEGER DEFAULT 1
	);

	CREATE INDEX IF NOT EXISTS idx_query_history_connection_id ON query_history(connection_id);
	CREATE INDEX IF NOT EXISTS idx_query_history_timestamp ON query_history(timestamp);

	CREATE TABLE IF NOT EXISTS saved_queries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		sql TEXT NOT NULL,
		description TEXT,
		connection_id TEXT,
		database TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_saved_queries_connection_id ON saved_queries(connection_id);
	CREATE INDEX IF NOT EXISTS idx_saved_queries_name ON saved_queries(name);
	`

	_, err := cs.db.Exec(schema)
	return err
}

// SaveProfile saves or updates a connection profile
func (cs *ConfigStorage) SaveProfile(profile types.ConnectionProfile) error {
	query := `
	INSERT INTO connection_profiles (
		id, name, host, port, username, password, database, charset, timeout,
		ssh_enabled, ssh_host, ssh_port, ssh_username, ssh_password, ssh_key_path,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		host = excluded.host,
		port = excluded.port,
		username = excluded.username,
		password = excluded.password,
		database = excluded.database,
		charset = excluded.charset,
		timeout = excluded.timeout,
		ssh_enabled = excluded.ssh_enabled,
		ssh_host = excluded.ssh_host,
		ssh_port = excluded.ssh_port,
		ssh_username = excluded.ssh_username,
		ssh_password = excluded.ssh_password,
		ssh_key_path = excluded.ssh_key_path,
		updated_at = excluded.updated_at
	`

	sshEnabled := 0
	if profile.SSHEnabled {
		sshEnabled = 1
	}

	now := time.Now()
	// Handle pointer types for timestamps
	if profile.CreatedAt == nil || profile.CreatedAt.IsZero() {
		profile.CreatedAt = &now
	}
	profile.UpdatedAt = &now

	_, err := cs.db.Exec(query,
		profile.ID, profile.Name, profile.Host, profile.Port,
		profile.Username, profile.Password, profile.Database,
		profile.Charset, profile.Timeout,
		sshEnabled, profile.SSHHost, profile.SSHPort,
		profile.SSHUsername, profile.SSHPassword, profile.SSHKeyPath,
		profile.CreatedAt, profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// GetProfile retrieves a connection profile by ID
func (cs *ConfigStorage) GetProfile(id string) (*types.ConnectionProfile, error) {
	query := `
	SELECT id, name, host, port, username, password, database, charset, timeout,
		   ssh_enabled, ssh_host, ssh_port, ssh_username, ssh_password, ssh_key_path,
		   created_at, updated_at
	FROM connection_profiles
	WHERE id = ?
	`

	var profile types.ConnectionProfile
	var sshEnabled int
	var sshHost, sshUsername, sshPassword, sshKeyPath sql.NullString
	var sshPort sql.NullInt64
	var createdAt, updatedAt time.Time

	err := cs.db.QueryRow(query, id).Scan(
		&profile.ID, &profile.Name, &profile.Host, &profile.Port,
		&profile.Username, &profile.Password, &profile.Database,
		&profile.Charset, &profile.Timeout,
		&sshEnabled, &sshHost, &sshPort,
		&sshUsername, &sshPassword, &sshKeyPath,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("profile not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	profile.SSHEnabled = sshEnabled == 1
	if sshHost.Valid {
		profile.SSHHost = sshHost.String
	}
	if sshPort.Valid {
		profile.SSHPort = int(sshPort.Int64)
	}
	if sshUsername.Valid {
		profile.SSHUsername = sshUsername.String
	}
	if sshPassword.Valid {
		profile.SSHPassword = sshPassword.String
	}
	if sshKeyPath.Valid {
		profile.SSHKeyPath = sshKeyPath.String
	}

	// Set pointer timestamps
	profile.CreatedAt = &createdAt
	profile.UpdatedAt = &updatedAt

	return &profile, nil
}

// ListProfiles retrieves all connection profiles
func (cs *ConfigStorage) ListProfiles() ([]types.ConnectionProfile, error) {
	query := `
	SELECT id, name, host, port, username, password, database, charset, timeout,
		   ssh_enabled, ssh_host, ssh_port, ssh_username, ssh_password, ssh_key_path,
		   created_at, updated_at
	FROM connection_profiles
	ORDER BY name
	`

	rows, err := cs.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}
	defer rows.Close()

	// Initialize as empty slice instead of nil slice
	profiles := []types.ConnectionProfile{}

	for rows.Next() {
		var profile types.ConnectionProfile
		var sshEnabled int
		var sshHost, sshUsername, sshPassword, sshKeyPath sql.NullString
		var sshPort sql.NullInt64
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&profile.ID, &profile.Name, &profile.Host, &profile.Port,
			&profile.Username, &profile.Password, &profile.Database,
			&profile.Charset, &profile.Timeout,
			&sshEnabled, &sshHost, &sshPort,
			&sshUsername, &sshPassword, &sshKeyPath,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}

		profile.SSHEnabled = sshEnabled == 1
		if sshHost.Valid {
			profile.SSHHost = sshHost.String
		}
		if sshPort.Valid {
			profile.SSHPort = int(sshPort.Int64)
		}
		if sshUsername.Valid {
			profile.SSHUsername = sshUsername.String
		}
		if sshPassword.Valid {
			profile.SSHPassword = sshPassword.String
		}
		if sshKeyPath.Valid {
			profile.SSHKeyPath = sshKeyPath.String
		}

		// Set pointer timestamps
		profile.CreatedAt = &createdAt
		profile.UpdatedAt = &updatedAt

		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating profiles: %w", err)
	}

	return profiles, nil
}

// DeleteProfile deletes a connection profile by ID
func (cs *ConfigStorage) DeleteProfile(id string) error {
	query := `DELETE FROM connection_profiles WHERE id = ?`

	result, err := cs.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("profile not found: %s", id)
	}

	return nil
}

// SaveSettings saves an application setting
func (cs *ConfigStorage) SaveSettings(key, value string) error {
	query := `
	INSERT INTO app_settings (key, value, updated_at)
	VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET
		value = excluded.value,
		updated_at = excluded.updated_at
	`

	_, err := cs.db.Exec(query, key, value, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save setting: %w", err)
	}

	return nil
}

// GetSettings retrieves an application setting by key
func (cs *ConfigStorage) GetSettings(key string) (string, error) {
	query := `SELECT value FROM app_settings WHERE key = ?`

	var value string
	err := cs.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("setting not found: %s", key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get setting: %w", err)
	}

	return value, nil
}

// Close closes the database connection
func (cs *ConfigStorage) Close() error {
	if cs.db != nil {
		return cs.db.Close()
	}
	return nil
}

// QueryHistoryEntry represents a query history record
type QueryHistoryEntry struct {
	ID            int       `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	ConnectionID  string    `json:"connectionId"`
	Database      string    `json:"database"`
	SQL           string    `json:"sql"`
	ExecutionTime int64     `json:"executionTime"` // milliseconds
	RowsAffected  int64     `json:"rowsAffected"`
	Success       bool      `json:"success"`
}

// SaveQueryHistory saves a query history entry
func (cs *ConfigStorage) SaveQueryHistory(entry QueryHistoryEntry) error {
	query := `
	INSERT INTO query_history (timestamp, connection_id, database, sql, execution_time, rows_affected, success)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	success := 0
	if entry.Success {
		success = 1
	}

	_, err := cs.db.Exec(query,
		entry.Timestamp,
		entry.ConnectionID,
		entry.Database,
		entry.SQL,
		entry.ExecutionTime,
		entry.RowsAffected,
		success,
	)

	if err != nil {
		return fmt.Errorf("failed to save query history: %w", err)
	}

	return nil
}

// GetQueryHistory retrieves query history for a connection
func (cs *ConfigStorage) GetQueryHistory(connectionID string, limit int) ([]QueryHistoryEntry, error) {
	query := `
	SELECT id, timestamp, connection_id, database, sql, execution_time, rows_affected, success
	FROM query_history
	WHERE connection_id = ?
	ORDER BY timestamp DESC
	LIMIT ?
	`

	rows, err := cs.db.Query(query, connectionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get query history: %w", err)
	}
	defer rows.Close()

	var history []QueryHistoryEntry

	for rows.Next() {
		var entry QueryHistoryEntry
		var success int
		var database sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.ConnectionID,
			&database,
			&entry.SQL,
			&entry.ExecutionTime,
			&entry.RowsAffected,
			&success,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query history: %w", err)
		}

		if database.Valid {
			entry.Database = database.String
		}
		entry.Success = success == 1

		history = append(history, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating query history: %w", err)
	}

	return history, nil
}

// ClearQueryHistory clears query history for a connection
func (cs *ConfigStorage) ClearQueryHistory(connectionID string) error {
	query := `DELETE FROM query_history WHERE connection_id = ?`

	_, err := cs.db.Exec(query, connectionID)
	if err != nil {
		return fmt.Errorf("failed to clear query history: %w", err)
	}

	return nil
}

// SavedQuery represents a saved SQL query
type SavedQuery struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	SQL          string    `json:"sql"`
	Description  string    `json:"description"`
	ConnectionID string    `json:"connectionId"`
	Database     string    `json:"database"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// SaveQuery saves a new SQL query
func (cs *ConfigStorage) SaveQuery(query SavedQuery) (int64, error) {
	sql := `
	INSERT INTO saved_queries (name, sql, description, connection_id, database, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := cs.db.Exec(sql,
		query.Name, query.SQL, query.Description,
		query.ConnectionID, query.Database,
		now, now,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to save query: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetSavedQueries retrieves all saved queries for a connection
func (cs *ConfigStorage) GetSavedQueries(connectionID string) ([]SavedQuery, error) {
	query := `
	SELECT id, name, sql, description, connection_id, database, created_at, updated_at
	FROM saved_queries
	WHERE connection_id = ? OR connection_id = ''
	ORDER BY updated_at DESC
	`

	rows, err := cs.db.Query(query, connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query saved queries: %w", err)
	}
	defer rows.Close()

	var queries []SavedQuery
	for rows.Next() {
		var q SavedQuery
		err := rows.Scan(
			&q.ID, &q.Name, &q.SQL, &q.Description,
			&q.ConnectionID, &q.Database,
			&q.CreatedAt, &q.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan saved query: %w", err)
		}
		queries = append(queries, q)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating saved queries: %w", err)
	}

	return queries, nil
}

// GetSavedQuery retrieves a saved query by ID
func (cs *ConfigStorage) GetSavedQuery(id int64) (*SavedQuery, error) {
	query := `
	SELECT id, name, sql, description, connection_id, database, created_at, updated_at
	FROM saved_queries
	WHERE id = ?
	`

	var q SavedQuery
	err := cs.db.QueryRow(query, id).Scan(
		&q.ID, &q.Name, &q.SQL, &q.Description,
		&q.ConnectionID, &q.Database,
		&q.CreatedAt, &q.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("saved query not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get saved query: %w", err)
	}

	return &q, nil
}

// UpdateSavedQuery updates an existing saved query
func (cs *ConfigStorage) UpdateSavedQuery(query SavedQuery) error {
	sql := `
	UPDATE saved_queries
	SET name = ?, sql = ?, description = ?, database = ?, updated_at = ?
	WHERE id = ?
	`

	now := time.Now()
	_, err := cs.db.Exec(sql,
		query.Name, query.SQL, query.Description,
		query.Database, now, query.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update saved query: %w", err)
	}

	return nil
}

// DeleteSavedQuery deletes a saved query by ID
func (cs *ConfigStorage) DeleteSavedQuery(id int64) error {
	sql := `DELETE FROM saved_queries WHERE id = ?`

	_, err := cs.db.Exec(sql, id)
	if err != nil {
		return fmt.Errorf("failed to delete saved query: %w", err)
	}

	return nil
}
