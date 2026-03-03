package types

import "time"

// QueryResult represents the result of a SQL query
type QueryResult struct {
	Type          string          `json:"type"` // SELECT, INSERT, UPDATE, DELETE, DDL
	Columns       []string        `json:"columns"`
	Rows          [][]interface{} `json:"rows"`
	RowsAffected  int64           `json:"rowsAffected"`
	ExecutionTime time.Duration   `json:"executionTime"`
	Error         *QueryError     `json:"error,omitempty"`
}

// QueryError represents a query execution error
type QueryError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Position int    `json:"position"`
}
