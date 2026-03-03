package repository

import (
	"testing"
)

// TestDatabaseStruct tests the Database struct
func TestDatabaseStruct(t *testing.T) {
	db := Database{
		Name: "testdb",
	}
	
	if db.Name != "testdb" {
		t.Errorf("Expected name 'testdb', got '%s'", db.Name)
	}
}

// TestTableStruct tests the Table struct
func TestTableStruct(t *testing.T) {
	table := Table{
		Name:    "users",
		Rows:    100,
		Engine:  "InnoDB",
		Comment: "User table",
	}
	
	if table.Name != "users" {
		t.Errorf("Expected name 'users', got '%s'", table.Name)
	}
	
	if table.Rows != 100 {
		t.Errorf("Expected 100 rows, got %d", table.Rows)
	}
}

// TestColumnStruct tests the Column struct
func TestColumnStruct(t *testing.T) {
	defaultVal := "0"
	col := Column{
		Name:          "id",
		Type:          "int(11)",
		Nullable:      false,
		DefaultValue:  &defaultVal,
		AutoIncrement: true,
		Comment:       "Primary key",
	}
	
	if col.Name != "id" {
		t.Errorf("Expected name 'id', got '%s'", col.Name)
	}
	
	if col.Nullable {
		t.Error("Expected Nullable to be false")
	}
	
	if !col.AutoIncrement {
		t.Error("Expected AutoIncrement to be true")
	}
	
	if col.DefaultValue == nil || *col.DefaultValue != "0" {
		t.Error("Expected DefaultValue to be '0'")
	}
}

// TestIndexStruct tests the Index struct
func TestIndexStruct(t *testing.T) {
	idx := Index{
		Name:      "idx_email",
		Type:      "UNIQUE",
		Columns:   []string{"email"},
		NonUnique: false,
	}
	
	if idx.Name != "idx_email" {
		t.Errorf("Expected name 'idx_email', got '%s'", idx.Name)
	}
	
	if idx.Type != "UNIQUE" {
		t.Errorf("Expected type 'UNIQUE', got '%s'", idx.Type)
	}
	
	if len(idx.Columns) != 1 || idx.Columns[0] != "email" {
		t.Error("Expected columns to contain 'email'")
	}
}

// TestForeignKeyStruct tests the ForeignKey struct
func TestForeignKeyStruct(t *testing.T) {
	fk := ForeignKey{
		Name:              "fk_user_role",
		Columns:           []string{"role_id"},
		ReferencedTable:   "roles",
		ReferencedColumns: []string{"id"},
		OnDelete:          "CASCADE",
		OnUpdate:          "CASCADE",
	}
	
	if fk.Name != "fk_user_role" {
		t.Errorf("Expected name 'fk_user_role', got '%s'", fk.Name)
	}
	
	if fk.ReferencedTable != "roles" {
		t.Errorf("Expected referenced table 'roles', got '%s'", fk.ReferencedTable)
	}
	
	if fk.OnDelete != "CASCADE" {
		t.Errorf("Expected OnDelete 'CASCADE', got '%s'", fk.OnDelete)
	}
}

// TestNewSchemaRepository tests creating a new repository
func TestNewSchemaRepository(t *testing.T) {
	// We can't create a real DB connection in unit tests
	// Just verify the constructor doesn't panic with nil
	repo := NewSchemaRepository(nil)
	
	if repo == nil {
		t.Error("Expected repository to be created")
	}
}
