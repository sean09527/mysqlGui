# Logger Component

## Overview

The Logger component provides comprehensive logging functionality for the MySQL Management Tool. It supports multiple log levels, writes to both file and SQLite database, filters sensitive information, and provides flexible log querying and export capabilities.

## Features

- **Multiple Log Levels**: INFO, ERROR, DEBUG
- **Dual Storage**: Logs are written to both a file and SQLite database for redundancy
- **Sensitive Data Filtering**: Automatically redacts passwords, private keys, and other sensitive information
- **Flexible Querying**: Filter logs by level, time range, keyword, and connection ID
- **Log Export**: Export logs within a time range to a file
- **Thread-Safe**: Safe for concurrent use from multiple goroutines
- **Configurable Log Level**: Change the logging level at runtime

## Usage

### Creating a Logger

```go
import (
    "mygui/backend/internal/logger"
    "mygui/backend/types"
)

// Create a new logger with INFO level
log, err := logger.NewLogger(types.LogLevelInfo)
if err != nil {
    // Handle error
}
defer log.Close()
```

### Logging Messages

```go
// Log an informational message
log.Info("Database connection established", map[string]interface{}{
    "connection_id": "conn-123",
    "host": "localhost",
    "database": "mydb",
})

// Log an error
err := someOperation()
if err != nil {
    log.Error("Operation failed", err, map[string]interface{}{
        "operation": "data_insert",
        "table": "users",
    })
}

// Log a debug message
log.Debug("Query execution details", map[string]interface{}{
    "sql": "SELECT * FROM users",
    "duration_ms": 45,
})
```

### Querying Logs

```go
// Get all ERROR logs
filter := types.LogFilter{
    Level: func() *types.LogLevel { l := types.LogLevelError; return &l }(),
    Limit: 100,
}

logs, err := log.GetLogs(filter)
if err != nil {
    // Handle error
}

// Get logs for a specific connection
filter = types.LogFilter{
    ConnectionID: "conn-123",
    Limit: 50,
}

logs, err = log.GetLogs(filter)

// Get logs within a time range
startTime := time.Now().Add(-24 * time.Hour)
endTime := time.Now()

filter = types.LogFilter{
    StartTime: &startTime,
    EndTime: &endTime,
    Limit: 1000,
}

logs, err = log.GetLogs(filter)

// Search logs by keyword
filter = types.LogFilter{
    Keyword: "connection",
    Limit: 100,
}

logs, err = log.GetLogs(filter)
```

### Exporting Logs

```go
// Export logs within a time range
startTime := time.Now().Add(-7 * 24 * time.Hour) // Last 7 days
endTime := time.Now()

exportPath, err := log.ExportLogs(startTime, endTime)
if err != nil {
    // Handle error
}

fmt.Printf("Logs exported to: %s\n", exportPath)
```

### Changing Log Level

```go
// Change log level at runtime
log.SetLevel(types.LogLevelDebug)

// Now debug messages will be logged
log.Debug("This will be logged", nil)
```

## Log Levels

The logger supports three log levels with the following priority:

1. **DEBUG** (lowest priority): Detailed information for debugging
2. **INFO**: General informational messages
3. **ERROR** (highest priority): Error messages

When you set a log level, all messages at that level and higher will be logged. For example:
- `LogLevelError`: Only ERROR messages are logged
- `LogLevelInfo`: INFO and ERROR messages are logged
- `LogLevelDebug`: All messages (DEBUG, INFO, ERROR) are logged

## Sensitive Data Filtering

The logger automatically filters sensitive information from log entries. The following field names are redacted:

- `password`, `Password`, `PASSWORD`
- `private_key`, `privateKey`, `PrivateKey`
- `ssh_password`, `sshPassword`, `SSHPassword`
- `secret`, `Secret`, `SECRET`
- `token`, `Token`, `TOKEN`

Any field containing these keywords (case-insensitive) will have its value replaced with `[REDACTED]`.

Example:

```go
log.Info("User login", map[string]interface{}{
    "username": "john",
    "password": "secret123",  // Will be redacted
    "ip": "192.168.1.1",
})

// Stored as:
// {
//   "username": "john",
//   "password": "[REDACTED]",
//   "ip": "192.168.1.1"
// }
```

## Storage

### File Storage

Logs are written to a file located at:
- **macOS**: `~/Library/Preferences/MyGUI/app.log`
- **Linux**: `~/.config/MyGUI/app.log`
- **Windows**: `%APPDATA%\MyGUI\app.log`

The log file format is:
```
[2024-02-28 16:06:14] [INFO] [INFO] Database connection established | fields: {"connection_id":"conn-123","database":"mydb","host":"localhost"}
```

### Database Storage

Logs are also stored in a SQLite database at:
- **macOS**: `~/Library/Preferences/MyGUI/config.db`
- **Linux**: `~/.config/MyGUI/config.db`
- **Windows**: `%APPDATA%\MyGUI\config.db`

The `operation_logs` table schema:

```sql
CREATE TABLE operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    level TEXT NOT NULL,
    operation TEXT NOT NULL,
    message TEXT,
    details TEXT,  -- JSON format
    connection_id TEXT
);
```

Indexes are created on `timestamp`, `level`, and `connection_id` for efficient querying.

## Thread Safety

The Logger is thread-safe and can be used concurrently from multiple goroutines. All public methods use a mutex to ensure safe concurrent access.

## Error Handling

If writing to the database fails, the logger will:
1. Continue to write to the log file
2. Log the database error to the log file

This ensures that logs are not lost even if the database is unavailable.

## Testing

The logger includes comprehensive unit tests covering:
- Basic logging functionality (Info, Error, Debug)
- Log level filtering
- Sensitive data filtering
- Log querying with various filters
- Time range filtering
- Log export
- Dynamic log level changes
- Concurrent writes

Run tests with:
```bash
go test -v ./backend/internal/logger/
```

## Requirements Mapping

This component satisfies the following requirements:

- **需求 19.1**: Records all database operations to log file ✓
- **需求 19.2**: Records all errors with timestamp, error type, and stack trace ✓
- **需求 19.3**: Allows users to view application logs ✓
- **需求 19.4**: Allows users to export log files ✓
- **需求 19.5**: Does not log sensitive information (passwords, private keys) ✓
- **需求 19.6**: Provides structured logging with fields ✓
- **非功能性需求-安全性.4**: Does not log sensitive information ✓

## Future Enhancements

Potential improvements for future versions:

1. **Log Rotation**: Implement automatic log file rotation based on size or time
2. **Log Compression**: Compress old log files to save disk space
3. **Remote Logging**: Support sending logs to remote logging services
4. **Structured Logging**: Use a structured logging library like zap or logrus
5. **Log Levels per Module**: Allow different log levels for different modules
6. **Performance Metrics**: Add metrics for log write performance
