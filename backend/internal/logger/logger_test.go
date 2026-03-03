package logger

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mygui/backend/types"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestLogger creates a logger with a temporary database for testing
func setupTestLogger(t *testing.T, level types.LogLevel) (*Logger, string) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create logger with custom paths
	logPath := filepath.Join(tempDir, "test.log")
	dbPath := filepath.Join(tempDir, "test.db")

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logFile.Close()
		t.Fatalf("Failed to open database: %v", err)
	}

	logger := &Logger{
		logFile: logFile,
		db:      db,
		level:   level,
	}

	if err := logger.initSchema(); err != nil {
		logFile.Close()
		db.Close()
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	return logger, tempDir
}

func TestNewLogger(t *testing.T) {
	// This test would normally create a logger in the user's config directory
	// For testing purposes, we'll skip this and use setupTestLogger instead
	t.Skip("Skipping NewLogger test to avoid creating files in user's config directory")
}

func TestLogger_Info(t *testing.T) {
	logger, _ := setupTestLogger(t, types.LogLevelInfo)
	defer logger.Close()

	// Log an info message
	logger.Info("Test info message", map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	})

	// Verify log was written to database
	filter := types.LogFilter{
		Level: func() *types.LogLevel { l := types.LogLevelInfo; return &l }(),
		Limit: 10,
	}

	logs, err := logger.GetLogs(filter)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.Level != types.LogLevelInfo {
		t.Errorf("Expected level INFO, got %s", log.Level)
	}

	if log.Message != "Test info message" {
		t.Errorf("Expected message 'Test info message', got '%s'", log.Message)
	}

	if log.Details["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got '%v'", log.Details["key1"])
	}
}

func TestLogger_Error(t *testing.T) {
	logger, _ := setupTestLogger(t, types.LogLevelInfo)
	defer logger.Close()

	// Log an error message
	testErr := os.ErrNotExist
	logger.Error("Test error message", testErr, map[string]interface{}{
		"operation": "file_read",
	})

	// Verify log was written to database
	filter := types.LogFilter{
		Level: func() *types.LogLevel { l := types.LogLevelError; return &l }(),
		Limit: 10,
	}

	logs, err := logger.GetLogs(filter)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.Level != types.LogLevelError {
		t.Errorf("Expected level ERROR, got %s", log.Level)
	}

	if log.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", log.Message)
	}

	if log.Details["error"] == nil {
		t.Error("Expected error field in details")
	}
}

func TestLogger_Debug(t *testing.T) {
	logger, _ := setupTestLogger(t, types.LogLevelDebug)
	defer logger.Close()

	// Log a debug message
	logger.Debug("Test debug message", map[string]interface{}{
		"debug_info": "test",
	})

	// Verify log was written to database
	filter := types.LogFilter{
		Level: func() *types.LogLevel { l := types.LogLevelDebug; return &l }(),
		Limit: 10,
	}

	logs, err := logger.GetLogs(filter)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.Level != types.LogLevelDebug {
		t.Errorf("Expected level DEBUG, got %s", log.Level)
	}
}

func TestLogger_LogLevelFiltering(t *testing.T) {
	// Create logger with INFO level (should not log DEBUG)
	logger, _ := setupTestLogger(t, types.LogLevelInfo)
	defer logger.Close()

	// Log messages at different levels
	logger.Debug("Debug message", nil)
	logger.Info("Info message", nil)
	logger.Error("Error message", nil, nil)

	// Get all logs
	filter := types.LogFilter{
		Limit: 100,
	}

	logs, err := logger.GetLogs(filter)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	// Should only have INFO and ERROR, not DEBUG
	if len(logs) != 2 {
		t.Fatalf("Expected 2 log entries (INFO and ERROR), got %d", len(logs))
	}

	// Verify DEBUG was not logged
	for _, log := range logs {
		if log.Level == types.LogLevelDebug {
			t.Error("DEBUG message should not have been logged with INFO level")
		}
	}
}

func TestLogger_FilterSensitiveData(t *testing.T) {
	logger, _ := setupTestLogger(t, types.LogLevelInfo)
	defer logger.Close()

	// Log message with sensitive data
	logger.Info("Connection created", map[string]interface{}{
		"username":     "testuser",
		"password":     "secret123",
		"ssh_password": "sshsecret",
		"host":         "localhost",
		"private_key":  "/path/to/key",
	})

	// Get logs
	filter := types.LogFilter{
		Limit: 10,
	}

	logs, err := logger.GetLogs(filter)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]

	// Verify sensitive data is redacted
	if log.Details["password"] != "[REDACTED]" {
		t.Errorf("Expected password to be redacted, got '%v'", log.Details["password"])
	}

	if log.Details["ssh_password"] != "[REDACTED]" {
		t.Errorf("Expected ssh_password to be redacted, got '%v'", log.Details["ssh_password"])
	}

	if log.Details["private_key"] != "[REDACTED]" {
		t.Errorf("Expected private_key to be redacted, got '%v'", log.Details["private_key"])
	}

	// Verify non-sensitive data is preserved
	if log.Details["username"] != "testuser" {
		t.Errorf("Expected username='testuser', got '%v'", log.Details["username"])
	}

	if log.Details["host"] != "localhost" {
		t.Errorf("Expected host='localhost', got '%v'", log.Details["host"])
	}
}

func TestLogger_GetLogs_WithFilters(t *testing.T) {
	logger, _ := setupTestLogger(t, types.LogLevelDebug)
	defer logger.Close()

	// Insert multiple log entries
	logger.Info("First info", map[string]interface{}{"connection_id": "conn1"})
	time.Sleep(10 * time.Millisecond)
	logger.Error("First error", nil, map[string]interface{}{"connection_id": "conn1"})
	time.Sleep(10 * time.Millisecond)
	logger.Info("Second info", map[string]interface{}{"connection_id": "conn2"})
	time.Sleep(10 * time.Millisecond)
	logger.Debug("Debug message", map[string]interface{}{"connection_id": "conn1"})

	// Test filter by level
	t.Run("FilterByLevel", func(t *testing.T) {
		filter := types.LogFilter{
			Level: func() *types.LogLevel { l := types.LogLevelInfo; return &l }(),
			Limit: 100,
		}

		logs, err := logger.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if len(logs) != 2 {
			t.Errorf("Expected 2 INFO logs, got %d", len(logs))
		}

		for _, log := range logs {
			if log.Level != types.LogLevelInfo {
				t.Errorf("Expected only INFO logs, got %s", log.Level)
			}
		}
	})

	// Test filter by connection ID
	t.Run("FilterByConnectionID", func(t *testing.T) {
		filter := types.LogFilter{
			ConnectionID: "conn1",
			Limit:        100,
		}

		logs, err := logger.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if len(logs) != 3 {
			t.Errorf("Expected 3 logs for conn1, got %d", len(logs))
		}

		for _, log := range logs {
			if log.ConnectionID != "conn1" {
				t.Errorf("Expected connection_id='conn1', got '%s'", log.ConnectionID)
			}
		}
	})

	// Test filter by keyword
	t.Run("FilterByKeyword", func(t *testing.T) {
		filter := types.LogFilter{
			Keyword: "error",
			Limit:   100,
		}

		logs, err := logger.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 log with keyword 'error', got %d", len(logs))
		}
	})

	// Test limit and offset
	t.Run("LimitAndOffset", func(t *testing.T) {
		filter := types.LogFilter{
			Limit:  2,
			Offset: 1,
		}

		logs, err := logger.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if len(logs) != 2 {
			t.Errorf("Expected 2 logs with limit=2, got %d", len(logs))
		}
	})
}

func TestLogger_GetLogs_WithTimeRange(t *testing.T) {
	logger, _ := setupTestLogger(t, types.LogLevelInfo)
	defer logger.Close()

	startTime := time.Now()

	// Log first message
	logger.Info("Before", nil)
	time.Sleep(100 * time.Millisecond)

	middleTime := time.Now()

	// Log second message
	time.Sleep(100 * time.Millisecond)
	logger.Info("After", nil)

	endTime := time.Now()

	// Test filter by start time
	t.Run("FilterByStartTime", func(t *testing.T) {
		filter := types.LogFilter{
			StartTime: &middleTime,
			Limit:     100,
		}

		logs, err := logger.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 log after middle time, got %d", len(logs))
		}

		if len(logs) > 0 && logs[0].Message != "After" {
			t.Errorf("Expected message 'After', got '%s'", logs[0].Message)
		}
	})

	// Test filter by end time
	t.Run("FilterByEndTime", func(t *testing.T) {
		filter := types.LogFilter{
			EndTime: &middleTime,
			Limit:   100,
		}

		logs, err := logger.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if len(logs) != 1 {
			t.Errorf("Expected 1 log before middle time, got %d", len(logs))
		}

		if len(logs) > 0 && logs[0].Message != "Before" {
			t.Errorf("Expected message 'Before', got '%s'", logs[0].Message)
		}
	})

	// Test filter by time range
	t.Run("FilterByTimeRange", func(t *testing.T) {
		filter := types.LogFilter{
			StartTime: &startTime,
			EndTime:   &endTime,
			Limit:     100,
		}

		logs, err := logger.GetLogs(filter)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if len(logs) != 2 {
			t.Errorf("Expected 2 logs in time range, got %d", len(logs))
		}
	})
}

func TestLogger_ExportLogs(t *testing.T) {
	logger, tempDir := setupTestLogger(t, types.LogLevelInfo)
	defer logger.Close()

	startTime := time.Now().Add(-1 * time.Hour)

	// Log some messages
	logger.Info("Export test 1", map[string]interface{}{"key": "value1"})
	logger.Error("Export test 2", nil, map[string]interface{}{"key": "value2"})
	logger.Info("Export test 3", map[string]interface{}{"key": "value3"})

	endTime := time.Now().Add(1 * time.Hour)

	// Export logs
	exportPath, err := logger.ExportLogs(startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to export logs: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Errorf("Export file does not exist: %s", exportPath)
	}

	// Read export file
	content, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	contentStr := string(content)

	// Verify content contains log messages
	if !contains(contentStr, "Export test 1") {
		t.Error("Export file should contain 'Export test 1'")
	}

	if !contains(contentStr, "Export test 2") {
		t.Error("Export file should contain 'Export test 2'")
	}

	if !contains(contentStr, "Export test 3") {
		t.Error("Export file should contain 'Export test 3'")
	}

	if !contains(contentStr, "Total logs exported: 3") {
		t.Error("Export file should contain total count")
	}

	// Clean up export file
	os.Remove(exportPath)

	// Verify export path is in temp directory (for test)
	if !contains(exportPath, tempDir) {
		t.Logf("Note: Export path is not in temp directory: %s", exportPath)
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger, _ := setupTestLogger(t, types.LogLevelInfo)
	defer logger.Close()

	// Initially set to INFO, DEBUG should not be logged
	logger.Debug("Debug 1", nil)
	logger.Info("Info 1", nil)

	// Change level to DEBUG
	logger.SetLevel(types.LogLevelDebug)

	// Now DEBUG should be logged
	logger.Debug("Debug 2", nil)
	logger.Info("Info 2", nil)

	// Get all logs
	filter := types.LogFilter{
		Limit: 100,
	}

	logs, err := logger.GetLogs(filter)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	// Should have 3 logs: Info 1, Debug 2, Info 2
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}

	// Verify Debug 1 was not logged
	for _, log := range logs {
		if log.Message == "Debug 1" {
			t.Error("Debug 1 should not have been logged")
		}
	}

	// Verify Debug 2 was logged
	foundDebug2 := false
	for _, log := range logs {
		if log.Message == "Debug 2" {
			foundDebug2 = true
			break
		}
	}
	if !foundDebug2 {
		t.Error("Debug 2 should have been logged after level change")
	}
}

func TestLogger_ConcurrentWrites(t *testing.T) {
	logger, _ := setupTestLogger(t, types.LogLevelInfo)
	defer logger.Close()

	// Write logs concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("Concurrent log", map[string]interface{}{
				"goroutine": id,
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all logs were written
	filter := types.LogFilter{
		Limit: 100,
	}

	logs, err := logger.GetLogs(filter)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 10 {
		t.Errorf("Expected 10 logs from concurrent writes, got %d", len(logs))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
