package query

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Executor) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	executor := NewExecutor(db)
	return db, mock, executor
}

func TestNewExecutor(t *testing.T) {
	db, _, executor := setupMockDB(t)
	defer db.Close()

	assert.NotNil(t, executor)
	assert.Equal(t, 30*time.Second, executor.defaultTimeout)
	assert.NotNil(t, executor.activeQueries)
}

func TestExecutor_DetermineQueryType(t *testing.T) {
	db, _, executor := setupMockDB(t)
	defer db.Close()

	tests := []struct {
		name     string
		query    string
		expected QueryType
	}{
		{"SELECT query", "SELECT * FROM users", QueryTypeSelect},
		{"SELECT with whitespace", "  select * from users  ", QueryTypeSelect},
		{"SHOW query", "SHOW TABLES", QueryTypeSelect},
		{"DESCRIBE query", "DESCRIBE users", QueryTypeSelect},
		{"DESC query", "DESC users", QueryTypeSelect},
		{"EXPLAIN query", "EXPLAIN SELECT * FROM users", QueryTypeSelect},
		{"INSERT query", "INSERT INTO users VALUES (1, 'test')", QueryTypeInsert},
		{"UPDATE query", "UPDATE users SET name = 'test'", QueryTypeUpdate},
		{"DELETE query", "DELETE FROM users WHERE id = 1", QueryTypeDelete},
		{"CREATE query", "CREATE TABLE test (id INT)", QueryTypeDDL},
		{"ALTER query", "ALTER TABLE users ADD COLUMN age INT", QueryTypeDDL},
		{"DROP query", "DROP TABLE users", QueryTypeDDL},
		{"TRUNCATE query", "TRUNCATE TABLE users", QueryTypeDDL},
		{"RENAME query", "RENAME TABLE old TO new", QueryTypeDDL},
		{"SET query", "SET @var = 1", QueryTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.determineQueryType(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_ExecuteSelect(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow(1, "Alice", "alice@example.com").
		AddRow(2, "Bob", "bob@example.com")

	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	// Execute query
	result, err := executor.Execute("SELECT * FROM users")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeSelect, result.Type)
	assert.Equal(t, []string{"id", "name", "email"}, result.Columns)
	assert.Equal(t, int64(2), result.RowsAffected)
	assert.Len(t, result.Rows, 2)
	assert.Nil(t, result.Error)
	assert.Greater(t, result.ExecutionTime, time.Duration(0))

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteSelectEmpty(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations for empty result
	rows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = 999").WillReturnRows(rows)

	// Execute query
	result, err := executor.Execute("SELECT * FROM users WHERE id = 999")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeSelect, result.Type)
	assert.Equal(t, []string{"id", "name"}, result.Columns)
	assert.Equal(t, int64(0), result.RowsAffected)
	assert.Len(t, result.Rows, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteInsert(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations
	mock.ExpectExec("INSERT INTO users").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute query
	result, err := executor.Execute("INSERT INTO users (name, email) VALUES ('Charlie', 'charlie@example.com')")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeInsert, result.Type)
	assert.Equal(t, int64(1), result.RowsAffected)
	assert.Nil(t, result.Columns)
	assert.Nil(t, result.Rows)
	assert.Nil(t, result.Error)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteUpdate(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations
	mock.ExpectExec("UPDATE users SET").
		WillReturnResult(sqlmock.NewResult(0, 3))

	// Execute query
	result, err := executor.Execute("UPDATE users SET active = 1 WHERE created_at < '2024-01-01'")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeUpdate, result.Type)
	assert.Equal(t, int64(3), result.RowsAffected)
	assert.Nil(t, result.Error)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteDelete(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations
	mock.ExpectExec("DELETE FROM users WHERE").
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Execute query
	result, err := executor.Execute("DELETE FROM users WHERE id IN (1, 2)")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeDelete, result.Type)
	assert.Equal(t, int64(2), result.RowsAffected)
	assert.Nil(t, result.Error)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteDDL(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"CREATE TABLE", "CREATE TABLE test (id INT PRIMARY KEY)"},
		{"ALTER TABLE", "ALTER TABLE users ADD COLUMN age INT"},
		{"DROP TABLE", "DROP TABLE test"},
		{"TRUNCATE TABLE", "TRUNCATE TABLE logs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, executor := setupMockDB(t)
			defer db.Close()

			// Use regex pattern matching for ExpectExec
			mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))

			result, err := executor.Execute(tt.query)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, QueryTypeDDL, result.Type)
			assert.Equal(t, int64(0), result.RowsAffected)
			assert.Nil(t, result.Error)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestExecutor_ExecuteWithTimeout(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with delay
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT (.+) FROM users").
		WillDelayFor(100 * time.Millisecond).
		WillReturnRows(rows)

	// Execute with sufficient timeout
	result, err := executor.ExecuteWithTimeout("SELECT * FROM users", 500*time.Millisecond)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeSelect, result.Type)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteTimeout(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with long delay
	mock.ExpectQuery("SELECT (.+) FROM users").
		WillDelayFor(200 * time.Millisecond).
		WillReturnError(context.DeadlineExceeded)

	// Execute with short timeout
	result, err := executor.ExecuteWithTimeout("SELECT * FROM users", 50*time.Millisecond)

	require.Error(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Error)
	// The error message can vary, just check that an error occurred
	assert.NotEmpty(t, result.Error.Message)
}

func TestExecutor_ExecuteError(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with error
	mock.ExpectQuery("SELECT (.+) FROM nonexistent").
		WillReturnError(sql.ErrNoRows)

	// Execute query
	result, err := executor.Execute("SELECT * FROM nonexistent")

	require.Error(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Message, "no rows")
}

func TestExecutor_Cancel(t *testing.T) {
	db, _, executor := setupMockDB(t)
	defer db.Close()

	// Register a mock query
	ctx, cancel := context.WithCancel(context.Background())
	queryID := "test-query-id"
	executor.registerQuery(queryID, cancel)

	// Cancel the query
	err := executor.Cancel(queryID)
	assert.NoError(t, err)

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context was not cancelled")
	}
}

func TestExecutor_CancelNonExistent(t *testing.T) {
	db, _, executor := setupMockDB(t)
	defer db.Close()

	// Try to cancel non-existent query
	err := executor.Cancel("non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExecutor_ConvertValue(t *testing.T) {
	db, _, executor := setupMockDB(t)
	defer db.Close()

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"nil value", nil, nil},
		{"string value", "test", "test"},
		{"int value", 42, 42},
		{"byte array", []byte("hello"), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock column type
			colType := &sql.ColumnType{}
			result := executor.convertValue(tt.input, colType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_ParseError(t *testing.T) {
	db, _, executor := setupMockDB(t)
	defer db.Close()

	tests := []struct {
		name          string
		err           error
		expectedMsg   string
		expectedCode  int
	}{
		{
			"generic error",
			sql.ErrNoRows,
			"sql: no rows in result set",
			0,
		},
		{
			"connection error",
			sql.ErrConnDone,
			"sql: connection is already closed",
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qErr := executor.parseError(tt.err)
			assert.NotNil(t, qErr)
			assert.Equal(t, tt.expectedMsg, qErr.Message)
			assert.Equal(t, -1, qErr.Position)
		})
	}
}

func TestExecutor_ConcurrentQueries(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations for multiple queries
	for i := 0; i < 5; i++ {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(i)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(rows)
	}

	// Execute queries concurrently
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			result, err := executor.Execute("SELECT * FROM users")
			assert.NoError(t, err)
			assert.NotNil(t, result)
			done <- true
		}()
	}

	// Wait for all queries to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteSelectWithNullValues(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with NULL values
	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow(1, "Alice", nil).
		AddRow(2, nil, "bob@example.com").
		AddRow(3, "Charlie", "charlie@example.com")

	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	// Execute query
	result, err := executor.Execute("SELECT * FROM users")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeSelect, result.Type)
	assert.Len(t, result.Rows, 3)
	
	// Check NULL values are properly handled
	assert.Nil(t, result.Rows[0][2]) // email is NULL
	assert.Nil(t, result.Rows[1][1]) // name is NULL

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteUpdateNoRowsAffected(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations - no rows match the condition
	mock.ExpectExec("UPDATE users SET").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute query
	result, err := executor.Execute("UPDATE users SET active = 1 WHERE id = 999999")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeUpdate, result.Type)
	assert.Equal(t, int64(0), result.RowsAffected)
	assert.Nil(t, result.Error)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteDeleteNoRowsAffected(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations - no rows match the condition
	mock.ExpectExec("DELETE FROM users WHERE").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute query
	result, err := executor.Execute("DELETE FROM users WHERE id = 999999")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeDelete, result.Type)
	assert.Equal(t, int64(0), result.RowsAffected)
	assert.Nil(t, result.Error)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteSyntaxError(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with syntax error
	mock.ExpectQuery("SELECT (.+) FORM users").
		WillReturnError(sql.ErrNoRows)

	// Execute query with syntax error
	result, err := executor.Execute("SELECT * FORM users") // FORM instead of FROM

	// Assertions
	require.Error(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Error)
	assert.NotEmpty(t, result.Error.Message)
}

func TestExecutor_ExecuteDMLError(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with constraint violation
	mock.ExpectExec("INSERT INTO users").
		WillReturnError(sql.ErrConnDone)

	// Execute query
	result, err := executor.Execute("INSERT INTO users (id, name) VALUES (1, 'test')")

	// Assertions
	require.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeInsert, result.Type)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Message, "connection")
}

func TestExecutor_ExecuteDDLError(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with DDL error
	mock.ExpectExec("CREATE TABLE").
		WillReturnError(sql.ErrConnDone)

	// Execute query
	result, err := executor.Execute("CREATE TABLE users (id INT)")

	// Assertions
	require.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeDDL, result.Type)
	assert.NotNil(t, result.Error)
}

func TestExecutor_ExecuteWithVeryShortTimeout(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with minimal delay
	mock.ExpectQuery("SELECT (.+)").
		WillDelayFor(10 * time.Millisecond).
		WillReturnError(context.DeadlineExceeded)

	// Execute with very short timeout (1ms)
	result, err := executor.ExecuteWithTimeout("SELECT * FROM users", 1*time.Millisecond)

	// Assertions
	require.Error(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Error)
}

func TestExecutor_ExecuteMultipleDMLStatements(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	tests := []struct {
		name          string
		query         string
		rowsAffected  int64
		expectedType  QueryType
	}{
		{
			"Bulk insert",
			"INSERT INTO users (name) VALUES ('Alice'), ('Bob'), ('Charlie')",
			3,
			QueryTypeInsert,
		},
		{
			"Bulk update",
			"UPDATE users SET active = 1",
			10,
			QueryTypeUpdate,
		},
		{
			"Bulk delete",
			"DELETE FROM users WHERE created_at < '2020-01-01'",
			5,
			QueryTypeDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))

			result, err := executor.Execute(tt.query)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedType, result.Type)
			assert.Equal(t, tt.rowsAffected, result.RowsAffected)
			assert.Nil(t, result.Error)
		})
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecuteSelectWithLargeResultSet(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with large result set
	rows := sqlmock.NewRows([]string{"id", "name"})
	for i := 1; i <= 1000; i++ {
		rows.AddRow(i, "User"+string(rune(i)))
	}

	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	// Execute query
	result, err := executor.Execute("SELECT * FROM users")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, QueryTypeSelect, result.Type)
	assert.Equal(t, int64(1000), result.RowsAffected)
	assert.Len(t, result.Rows, 1000)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_QueryIDGeneration(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT (.+)").WillReturnRows(rows)

	// Execute query
	result, err := executor.Execute("SELECT * FROM users")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	
	// Verify ID is a valid UUID format (36 characters with hyphens)
	assert.Len(t, result.ID, 36)
	assert.Contains(t, result.ID, "-")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutor_ExecutionTimeTracking(t *testing.T) {
	db, mock, executor := setupMockDB(t)
	defer db.Close()

	// Setup mock expectations with delay
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT (.+)").
		WillDelayFor(50 * time.Millisecond).
		WillReturnRows(rows)

	// Execute query
	result, err := executor.Execute("SELECT * FROM users")

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.ExecutionTime, 40*time.Millisecond)
	assert.Less(t, result.ExecutionTime, 200*time.Millisecond)

	assert.NoError(t, mock.ExpectationsWereMet())
}
