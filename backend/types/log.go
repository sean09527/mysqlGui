package types

import "time"

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelError LogLevel = "ERROR"
	LogLevelDebug LogLevel = "DEBUG"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID           int64                  `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Level        LogLevel               `json:"level"`
	Operation    string                 `json:"operation"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	ConnectionID string                 `json:"connectionId,omitempty"`
}

// LogFilter represents filters for querying logs
type LogFilter struct {
	Level        *LogLevel  `json:"level,omitempty"`
	StartTime    *time.Time `json:"startTime,omitempty"`
	EndTime      *time.Time `json:"endTime,omitempty"`
	Keyword      string     `json:"keyword,omitempty"`
	ConnectionID string     `json:"connectionId,omitempty"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}
