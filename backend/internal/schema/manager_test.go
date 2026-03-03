package schema

import (
	"testing"

	"mygui/backend/internal/repository"
)

// TestTableSchemaStruct tests the TableSchema struct
func TestTableSchemaStruct(t *testing.T) {
	schema := TableSchema{
		Name: "users",
		Columns: []repository.Column{
			{
				Name:          "id",
				Type:          "int(11)",
				Nullable:      false,
				AutoIncrement: true,
			},
		},
		Indexes: []repository.Index{
			{
				Name:    "PRIMARY",
				Type:    "PRIMARY",
				Columns: []string{"id"},
			},
		},
		ForeignKeys: []repository.ForeignKey{},
		Engine:      "InnoDB",
		Charset:     "utf8mb4",
		Comment:     "User table",
	}
	
	if schema.Name != "users" {
		t.Errorf("Expected name 'users', got '%s'", schema.Name)
	}
	
	if len(schema.Columns) != 1 {
		t.Errorf("Expected 1 column, got %d", len(schema.Columns))
	}
	
	if schema.Engine != "InnoDB" {
		t.Errorf("Expected engine 'InnoDB', got '%s'", schema.Engine)
	}
}

// TestSchemaChangeStruct tests the SchemaChange struct
func TestSchemaChangeStruct(t *testing.T) {
	change := SchemaChange{
		Type:   "ADD_COLUMN",
		Target: "email",
		Data: repository.Column{
			Name:     "email",
			Type:     "varchar(255)",
			Nullable: false,
		},
	}
	
	if change.Type != "ADD_COLUMN" {
		t.Errorf("Expected type 'ADD_COLUMN', got '%s'", change.Type)
	}
	
	if change.Target != "email" {
		t.Errorf("Expected target 'email', got '%s'", change.Target)
	}
}

// TestNewManager tests creating a new manager
func TestNewManager(t *testing.T) {
	// We can't create a real DB connection in unit tests
	// Just verify the constructor doesn't panic with nil
	manager := NewManager(nil)
	
	if manager == nil {
		t.Error("Expected manager to be created")
	}
}

// TestQuoteColumns tests the quoteColumns helper method
func TestQuoteColumns(t *testing.T) {
	manager := NewManager(nil)
	
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"id"}, "`id`"},
		{[]string{"id", "name"}, "`id`, `name`"},
		{[]string{"user_id", "role_id", "created_at"}, "`user_id`, `role_id`, `created_at`"},
	}
	
	for _, test := range tests {
		result := manager.quoteColumns(test.input)
		if result != test.expected {
			t.Errorf("Expected '%s', got '%s'", test.expected, result)
		}
	}
}

// TestDetectDataLossWarnings tests data loss detection
func TestDetectDataLossWarnings(t *testing.T) {
	manager := NewManager(nil)
	
	tests := []struct {
		name     string
		changes  []SchemaChange
		expected int // number of warnings
	}{
		{
			name: "Drop column warning",
			changes: []SchemaChange{
				{Type: "DROP_COLUMN", Target: "old_field"},
			},
			expected: 1,
		},
		{
			name: "Modify to NOT NULL warning",
			changes: []SchemaChange{
				{
					Type:   "MODIFY_COLUMN",
					Target: "email",
					Data: repository.Column{
						Name:     "email",
						Type:     "varchar(255)",
						Nullable: false,
					},
				},
			},
			expected: 1,
		},
		{
			name: "Drop foreign key warning",
			changes: []SchemaChange{
				{Type: "DROP_FOREIGN_KEY", Target: "fk_user_role"},
			},
			expected: 1,
		},
		{
			name: "Add column no warning",
			changes: []SchemaChange{
				{
					Type:   "ADD_COLUMN",
					Target: "new_field",
					Data: repository.Column{
						Name:     "new_field",
						Type:     "varchar(100)",
						Nullable: true,
					},
				},
			},
			expected: 0,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			warnings := manager.detectDataLossWarnings(test.changes)
			if len(warnings) != test.expected {
				t.Errorf("Expected %d warnings, got %d: %v", test.expected, len(warnings), warnings)
			}
		})
	}
}

// TestGenerateCreateTableDDL tests CREATE TABLE statement generation
func TestGenerateCreateTableDDL(t *testing.T) {
	manager := NewManager(nil)
	
	defaultVal := "0"
	schema := TableSchema{
		Name: "users",
		Columns: []repository.Column{
			{
				Name:          "id",
				Type:          "int(11)",
				Nullable:      false,
				AutoIncrement: true,
			},
			{
				Name:         "name",
				Type:         "varchar(100)",
				Nullable:     false,
				DefaultValue: nil,
			},
			{
				Name:         "status",
				Type:         "int(11)",
				Nullable:     false,
				DefaultValue: &defaultVal,
			},
		},
		Indexes: []repository.Index{
			{
				Name:    "PRIMARY",
				Type:    "PRIMARY",
				Columns: []string{"id"},
			},
		},
		Engine:  "InnoDB",
		Charset: "utf8mb4",
		Comment: "User table",
	}
	
	ddl := manager.generateCreateTableDDL("testdb", schema)
	
	// Verify DDL contains expected elements
	expectedElements := []string{
		"CREATE TABLE `testdb`.`users`",
		"`id` int(11) NOT NULL AUTO_INCREMENT",
		"`name` varchar(100) NOT NULL",
		"`status` int(11) NOT NULL DEFAULT 0",
		"PRIMARY KEY (`id`)",
		"ENGINE=InnoDB",
		"CHARSET=utf8mb4",
		"COMMENT='User table'",
	}
	
	for _, element := range expectedElements {
		if !contains(ddl, element) {
			t.Errorf("Expected DDL to contain '%s', but it doesn't.\nGenerated DDL:\n%s", element, ddl)
		}
	}
}

// TestGenerateAlterTableStatements tests ALTER TABLE statement generation
func TestGenerateAlterTableStatements(t *testing.T) {
	manager := NewManager(nil)
	
	changes := []SchemaChange{
		{
			Type:   "ADD_COLUMN",
			Target: "email",
			Data: repository.Column{
				Name:     "email",
				Type:     "varchar(255)",
				Nullable: false,
			},
		},
		{
			Type:   "DROP_COLUMN",
			Target: "old_field",
		},
		{
			Type:   "ADD_INDEX",
			Target: "idx_email",
			Data: repository.Index{
				Name:    "idx_email",
				Type:    "UNIQUE",
				Columns: []string{"email"},
			},
		},
	}
	
	statements := manager.generateAlterTableStatements("testdb", "users", changes)
	
	if len(statements) != 3 {
		t.Errorf("Expected 3 statements, got %d", len(statements))
	}
	
	// Verify each statement type
	expectedPatterns := []string{
		"ALTER TABLE `testdb`.`users` ADD COLUMN `email` varchar(255) NOT NULL",
		"ALTER TABLE `testdb`.`users` DROP COLUMN `old_field`",
		"ALTER TABLE `testdb`.`users` ADD UNIQUE KEY `idx_email` (`email`)",
	}
	
	for i, expected := range expectedPatterns {
		if i < len(statements) && statements[i] != expected {
			t.Errorf("Statement %d:\nExpected: %s\nGot:      %s", i, expected, statements[i])
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
