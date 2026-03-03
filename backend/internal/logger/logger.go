package logger

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mygui/backend/types"

	_ "github.com/mattn/go-sqlite3"
	"github.com/vrischmann/userdir"
)

// Logger handles application logging to both file and SQLite database
type Logger struct {
	logFile *os.File
	db      *sql.DB
	level   types.LogLevel
	mu      sync.Mutex
}

// NewLogger creates a new Logger instance
func NewLogger(level types.LogLevel) (*Logger, error) {
	// Get config directory
	configDir := filepath.Join(userdir.GetConfigHome(), "MyGUI")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open log file
	logPath := filepath.Join(configDir, "app.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Open SQLite database
	dbPath := filepath.Join(configDir, "config.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		logFile.Close()
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger := &Logger{
		logFile: logFile,
		db:      db,
		level:   level,
	}

	// Initialize database schema
	if err := logger.initSchema(); err != nil {
		logFile.Close()
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return logger, nil
}

// initSchema creates the operation_logs table if it doesn't exist
func (l *Logger) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS operation_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		level TEXT NOT NULL,
		operation TEXT NOT NULL,
		message TEXT,
		details TEXT,
		connection_id TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_operation_logs_timestamp ON operation_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_operation_logs_level ON operation_logs(level);
	CREATE INDEX IF NOT EXISTS idx_operation_logs_connection_id ON operation_logs(connection_id);
	`

	_, err := l.db.Exec(schema)
	return err
}

// Info logs an informational message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log(types.LogLevelInfo, "INFO", message, nil, fields)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.log(types.LogLevelError, "ERROR", message, err, fields)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.log(types.LogLevelDebug, "DEBUG", message, nil, fields)
}

// log is the internal logging method
func (l *Logger) log(level types.LogLevel, operation string, message string, err error, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if we should log this level
	if !l.shouldLog(level) {
		return
	}

	// Filter sensitive information
	filteredFields := l.filterSensitiveData(fields)

	// Get connection ID if present
	connectionID := ""
	if cid, ok := filteredFields["connection_id"]; ok {
		if cidStr, ok := cid.(string); ok {
			connectionID = cidStr
		}
	}

	timestamp := time.Now()

	// Write to file
	l.writeToFile(timestamp, level, operation, message, err, filteredFields)

	// Write to database
	l.writeToDatabase(timestamp, level, operation, message, filteredFields, connectionID)
}

// shouldLog checks if the given level should be logged based on the configured level
func (l *Logger) shouldLog(level types.LogLevel) bool {
	levelPriority := map[types.LogLevel]int{
		types.LogLevelDebug: 0,
		types.LogLevelInfo:  1,
		types.LogLevelError: 2,
	}

	return levelPriority[level] >= levelPriority[l.level]
}

// filterSensitiveData removes sensitive information from log fields
func (l *Logger) filterSensitiveData(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}

	sensitiveKeys := []string{
		"password",
		"Password",
		"PASSWORD",
		"private_key",
		"privateKey",
		"PrivateKey",
		"ssh_password",
		"sshPassword",
		"SSHPassword",
		"secret",
		"Secret",
		"SECRET",
		"token",
		"Token",
		"TOKEN",
	}

	filtered := make(map[string]interface{})
	for key, value := range fields {
		// Check if key contains sensitive information
		isSensitive := false
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(strings.ToLower(key), strings.ToLower(sensitiveKey)) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			filtered[key] = "[REDACTED]"
		} else {
			filtered[key] = value
		}
	}

	return filtered
}

// writeToFile writes a log entry to the log file
func (l *Logger) writeToFile(timestamp time.Time, level types.LogLevel, operation string, message string, err error, fields map[string]interface{}) {
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s",
		timestamp.Format("2006-01-02 15:04:05"),
		level,
		operation,
		message,
	)

	if err != nil {
		logLine += fmt.Sprintf(" | error: %v", err)
	}

	if len(fields) > 0 {
		fieldsJSON, _ := json.Marshal(fields)
		logLine += fmt.Sprintf(" | fields: %s", string(fieldsJSON))
	}

	logLine += "\n"

	l.logFile.WriteString(logLine)
}

// writeToDatabase writes a log entry to the SQLite database
func (l *Logger) writeToDatabase(timestamp time.Time, level types.LogLevel, operation string, message string, fields map[string]interface{}, connectionID string) {
	var detailsJSON string
	if len(fields) > 0 {
		detailsBytes, err := json.Marshal(fields)
		if err == nil {
			detailsJSON = string(detailsBytes)
		}
	}

	query := `
	INSERT INTO operation_logs (timestamp, level, operation, message, details, connection_id)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := l.db.Exec(query, timestamp, string(level), operation, message, detailsJSON, connectionID)
	if err != nil {
		// If database write fails, at least log to file
		fmt.Fprintf(l.logFile, "[ERROR] Failed to write to database: %v\n", err)
	}
}

// GetLogs retrieves logs from the database based on the provided filter
func (l *Logger) GetLogs(filter types.LogFilter) ([]types.LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	query := `SELECT id, timestamp, level, operation, message, details, connection_id FROM operation_logs WHERE 1=1`
	args := []interface{}{}

	// Apply filters
	if filter.Level != nil {
		query += ` AND level = ?`
		args = append(args, string(*filter.Level))
	}

	if filter.StartTime != nil {
		query += ` AND timestamp >= ?`
		args = append(args, *filter.StartTime)
	}

	if filter.EndTime != nil {
		query += ` AND timestamp <= ?`
		args = append(args, *filter.EndTime)
	}

	if filter.Keyword != "" {
		query += ` AND (operation LIKE ? OR message LIKE ?)`
		keyword := "%" + filter.Keyword + "%"
		args = append(args, keyword, keyword)
	}

	if filter.ConnectionID != "" {
		query += ` AND connection_id = ?`
		args = append(args, filter.ConnectionID)
	}

	// Order by timestamp descending (newest first)
	query += ` ORDER BY timestamp DESC`

	// Apply limit and offset
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := l.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []types.LogEntry

	for rows.Next() {
		var log types.LogEntry
		var detailsJSON sql.NullString
		var connectionID sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Level,
			&log.Operation,
			&log.Message,
			&detailsJSON,
			&connectionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		// Parse details JSON
		if detailsJSON.Valid && detailsJSON.String != "" {
			var details map[string]interface{}
			if err := json.Unmarshal([]byte(detailsJSON.String), &details); err == nil {
				log.Details = details
			}
		}

		if connectionID.Valid {
			log.ConnectionID = connectionID.String
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating logs: %w", err)
	}

	return logs, nil
}

// ExportLogs exports logs within a time range to a file
func (l *Logger) ExportLogs(startTime, endTime time.Time) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Query logs within time range
	query := `
	SELECT id, timestamp, level, operation, message, details, connection_id
	FROM operation_logs
	WHERE timestamp >= ? AND timestamp <= ?
	ORDER BY timestamp ASC
	`

	rows, err := l.db.Query(query, startTime, endTime)
	if err != nil {
		return "", fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	// Create export file
	configDir := filepath.Join(userdir.GetConfigHome(), "MyGUI")
	exportPath := filepath.Join(configDir, fmt.Sprintf("logs_export_%s.log", time.Now().Format("20060102_150405")))

	exportFile, err := os.Create(exportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create export file: %w", err)
	}
	defer exportFile.Close()

	// Write header
	exportFile.WriteString(fmt.Sprintf("Log Export - %s to %s\n", startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05")))
	exportFile.WriteString(strings.Repeat("=", 80) + "\n\n")

	// Write logs
	count := 0
	for rows.Next() {
		var log types.LogEntry
		var detailsJSON sql.NullString
		var connectionID sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Level,
			&log.Operation,
			&log.Message,
			&detailsJSON,
			&connectionID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to scan log entry: %w", err)
		}

		// Format log entry
		logLine := fmt.Sprintf("[%s] [%s] [%s] %s",
			log.Timestamp.Format("2006-01-02 15:04:05"),
			log.Level,
			log.Operation,
			log.Message,
		)

		if connectionID.Valid && connectionID.String != "" {
			logLine += fmt.Sprintf(" | connection: %s", connectionID.String)
		}

		if detailsJSON.Valid && detailsJSON.String != "" {
			logLine += fmt.Sprintf(" | details: %s", detailsJSON.String)
		}

		logLine += "\n"

		exportFile.WriteString(logLine)
		count++
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating logs: %w", err)
	}

	// Write footer
	exportFile.WriteString(fmt.Sprintf("\nTotal logs exported: %d\n", count))

	return exportPath, nil
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level types.LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Close closes the logger and releases resources
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error

	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close log file: %w", err))
		}
	}

	if l.db != nil {
		if err := l.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing logger: %v", errs)
	}

	return nil
}
