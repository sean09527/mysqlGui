package importexport

import (
	"testing"
)

func TestFormatSQLValue(t *testing.T) {
	exporter := &Exporter{}

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, "NULL"},
		{"string value", "test", "'test'"},
		{"string with quote", "test'quote", "'test''quote'"},
		{"int value", 42, "42"},
		{"float value", 3.14, "3.14"},
		{"bool true", true, "1"},
		{"bool false", false, "0"},
		{"byte array", []byte("data"), "'data'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatSQLValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatSQLValue(%v) = %s, 期望 %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	exporter := &Exporter{}

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, ""},
		{"string value", "test", "test"},
		{"int value", 42, "42"},
		{"byte array", []byte("data"), "data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %s, 期望 %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "table_name", "`table_name`"},
		{"with backticks", "`table_name`", "`table_name`"},
		{"with spaces", "table name", "`table name`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("escapeIdentifier(%s) = %s, 期望 %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateInsertStatement(t *testing.T) {
	exporter := &Exporter{}

	columns := []string{"id", "name", "email"}
	values := []interface{}{1, "Alice", "alice@example.com"}

	result := exporter.generateInsertStatement("test_db", "test_table", columns, values)

	expected := "INSERT INTO `test_db`.`test_table` (`id`, `name`, `email`) VALUES (1, 'Alice', 'alice@example.com');"

	if result != expected {
		t.Errorf("generateInsertStatement() = %s, 期望 %s", result, expected)
	}
}

func TestImportResultInitialization(t *testing.T) {
	result := &ImportResult{
		TotalRows:   10,
		SuccessRows: 8,
		FailedRows:  2,
		Errors: []ImportError{
			{Row: 3, Message: "Error 1"},
			{Row: 7, Message: "Error 2"},
		},
	}

	if result.TotalRows != 10 {
		t.Errorf("TotalRows = %d, 期望 10", result.TotalRows)
	}

	if result.SuccessRows != 8 {
		t.Errorf("SuccessRows = %d, 期望 8", result.SuccessRows)
	}

	if result.FailedRows != 2 {
		t.Errorf("FailedRows = %d, 期望 2", result.FailedRows)
	}

	if len(result.Errors) != 2 {
		t.Errorf("Errors length = %d, 期望 2", len(result.Errors))
	}
}

func TestColumnMapping(t *testing.T) {
	mapping := ColumnMapping{
		FileColumns:  []string{"col1", "col2", "col3"},
		TableColumns: []string{"table_col1", "table_col2", "table_col3"},
	}

	if len(mapping.FileColumns) != len(mapping.TableColumns) {
		t.Error("FileColumns 和 TableColumns 长度应该相等")
	}

	if len(mapping.FileColumns) != 3 {
		t.Errorf("FileColumns length = %d, 期望 3", len(mapping.FileColumns))
	}
}
