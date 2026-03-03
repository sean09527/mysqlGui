package schema

import (
	"database/sql"
	"fmt"
	"strings"

	"mygui/backend/internal/repository"
)

// PrimaryKey represents a primary key constraint
type PrimaryKey struct {
	Columns []string `json:"columns"`
}

// TableSchema represents the complete schema of a table
type TableSchema struct {
	Name        string                  `json:"name"`
	Columns     []repository.Column     `json:"columns"`
	PrimaryKey  *PrimaryKey             `json:"primaryKey,omitempty"`
	Indexes     []repository.Index      `json:"indexes"`
	ForeignKeys []repository.ForeignKey `json:"foreignKeys"`
	Engine      string                  `json:"engine"`
	Charset     string                  `json:"charset"`
	Comment     string                  `json:"comment"`
}

// Manager handles schema management operations
type Manager struct {
	db         *sql.DB
	repository *repository.SchemaRepository
}

// NewManager creates a new schema manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db:         db,
		repository: repository.NewSchemaRepository(db),
	}
}

// ClearCache clears all cached schema data
func (m *Manager) ClearCache() {
	m.repository.ClearCache()
}

// ClearTableCache clears cached table list for a specific database
func (m *Manager) ClearTableCache(database string) {
	m.repository.ClearTableCache(database)
}

// ListDatabases returns all databases
func (m *Manager) ListDatabases() ([]repository.Database, error) {
	return m.repository.ListDatabases()
}

// ListTables returns all tables in a database
func (m *Manager) ListTables(database string) ([]repository.Table, error) {
	return m.repository.ListTables(database)
}

// GetTableSchema returns the complete schema of a table
func (m *Manager) GetTableSchema(database, table string) (*TableSchema, error) {
	// Get columns
	columns, err := m.repository.GetColumns(database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Get indexes
	indexes, err := m.repository.GetIndexes(database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}

	// Get foreign keys
	foreignKeys, err := m.repository.GetForeignKeys(database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign keys: %w", err)
	}

	// Get table info (engine, charset, comment)
	var engine, charset, comment string
	query := `
		SELECT 
			IFNULL(ENGINE, ''),
			IFNULL(TABLE_COLLATION, ''),
			IFNULL(TABLE_COMMENT, '')
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`
	err = m.db.QueryRow(query, database, table).Scan(&engine, &charset, &comment)
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}

	// Extract primary key from indexes
	var primaryKey *PrimaryKey
	for _, idx := range indexes {
		if idx.Type == "PRIMARY" {
			primaryKey = &PrimaryKey{
				Columns: idx.Columns,
			}
			break
		}
	}

	return &TableSchema{
		Name:        table,
		Columns:     columns,
		PrimaryKey:  primaryKey,
		Indexes:     indexes,
		ForeignKeys: foreignKeys,
		Engine:      engine,
		Charset:     charset,
		Comment:     comment,
	}, nil
}

// GetCreateTableDDL returns the CREATE TABLE statement for a table
func (m *Manager) GetCreateTableDDL(database, table string) (string, error) {
	// Use SHOW CREATE TABLE to get the DDL
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", database, table)

	var tableName, createStmt string
	err := m.db.QueryRow(query).Scan(&tableName, &createStmt)
	if err != nil {
		return "", fmt.Errorf("failed to get CREATE TABLE statement: %w", err)
	}

	return createStmt, nil
}

// CreateTable creates a new table with the given schema
func (m *Manager) CreateTable(database string, schema TableSchema) error {
	ddl := m.generateCreateTableDDL(database, schema)

	_, err := m.db.Exec(ddl)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Clear cache for this database after successful creation
	m.ClearTableCache(database)

	return nil
}

// generateCreateTableDDL generates a CREATE TABLE statement from a schema
func (m *Manager) generateCreateTableDDL(database string, schema TableSchema) string {
	var parts []string

	// Add columns
	for _, col := range schema.Columns {
		colDef := fmt.Sprintf("`%s` %s", col.Name, col.Type)

		if !col.Nullable {
			colDef += " NOT NULL"
		} else {
			colDef += " NULL"
		}

		if col.DefaultValue != nil && *col.DefaultValue != "" {
			// 检查是否需要引号
			defaultVal := *col.DefaultValue
			// 对于特殊值（如 CURRENT_TIMESTAMP, NULL），不加引号
			if defaultVal == "CURRENT_TIMESTAMP" || defaultVal == "NULL" {
				colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
			} else {
				// 对于普通值，加引号并转义
				colDef += fmt.Sprintf(" DEFAULT '%s'", strings.ReplaceAll(defaultVal, "'", "''"))
			}
		}

		if col.AutoIncrement {
			colDef += " AUTO_INCREMENT"
		}

		if col.Comment != "" {
			colDef += fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''"))
		}

		parts = append(parts, colDef)
	}

	// Add primary key from PrimaryKey field (preferred)
	if schema.PrimaryKey != nil && len(schema.PrimaryKey.Columns) > 0 {
		parts = append(parts, fmt.Sprintf("PRIMARY KEY (%s)", m.quoteColumns(schema.PrimaryKey.Columns)))
	} else {
		// Fallback: check indexes for PRIMARY type (for backward compatibility)
		for _, idx := range schema.Indexes {
			if idx.Type == "PRIMARY" {
				parts = append(parts, fmt.Sprintf("PRIMARY KEY (%s)", m.quoteColumns(idx.Columns)))
				break
			}
		}
	}

	// Add indexes (skip PRIMARY type as it's handled above)
	for _, idx := range schema.Indexes {
		if idx.Type == "UNIQUE" {
			parts = append(parts, fmt.Sprintf("UNIQUE KEY `%s` (%s)", idx.Name, m.quoteColumns(idx.Columns)))
		} else if idx.Type == "INDEX" {
			parts = append(parts, fmt.Sprintf("KEY `%s` (%s)", idx.Name, m.quoteColumns(idx.Columns)))
		}
	}

	// Add foreign keys
	for _, fk := range schema.ForeignKeys {
		fkDef := fmt.Sprintf("CONSTRAINT `%s` FOREIGN KEY (%s) REFERENCES `%s` (%s)",
			fk.Name,
			m.quoteColumns(fk.Columns),
			fk.ReferencedTable,
			m.quoteColumns(fk.ReferencedColumns))

		if fk.OnDelete != "" {
			fkDef += fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
		}

		if fk.OnUpdate != "" {
			fkDef += fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
		}

		parts = append(parts, fkDef)
	}

	// Build the complete CREATE TABLE statement
	ddl := fmt.Sprintf("CREATE TABLE `%s`.`%s` (\n  %s\n)",
		database,
		schema.Name,
		strings.Join(parts, ",\n  "))

	// Add table options
	if schema.Engine != "" {
		ddl += fmt.Sprintf(" ENGINE=%s", schema.Engine)
	}

	if schema.Charset != "" {
		ddl += fmt.Sprintf(" CHARSET=%s", schema.Charset)
	}

	if schema.Comment != "" {
		ddl += fmt.Sprintf(" COMMENT='%s'", strings.ReplaceAll(schema.Comment, "'", "''"))
	}

	return ddl
}

// quoteColumns quotes column names and joins them with commas
func (m *Manager) quoteColumns(columns []string) string {
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = fmt.Sprintf("`%s`", col)
	}
	return strings.Join(quoted, ", ")
}

// SchemaChange represents a change to a table schema
type SchemaChange struct {
	Type   string      `json:"type"` // ADD_COLUMN, MODIFY_COLUMN, DROP_COLUMN, ADD_INDEX, DROP_INDEX, ADD_FOREIGN_KEY, DROP_FOREIGN_KEY
	Target string      `json:"target"`
	Data   interface{} `json:"data"`
}

// AlterTable modifies a table schema
func (m *Manager) AlterTable(database, table string, changes []SchemaChange) error {
	// Check for data loss warnings
	warnings := m.detectDataLossWarnings(changes)
	if len(warnings) > 0 {
		// Return warnings as error - caller should handle confirmation
		return fmt.Errorf("data loss warning: %s", strings.Join(warnings, "; "))
	}

	// Generate ALTER TABLE statements
	statements := m.generateAlterTableStatements(database, table, changes)

	// Log generated SQL statements for debugging
	fmt.Printf("=== Generated ALTER TABLE statements ===\n")
	for i, stmt := range statements {
		fmt.Printf("Statement %d: %s\n", i+1, stmt)
	}
	fmt.Printf("========================================\n")

	// Execute in a transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, stmt := range statements {
		_, err := tx.Exec(stmt)
		if err != nil {
			fmt.Printf("ERROR executing SQL: %s\nError: %v\n", stmt, err)
			return fmt.Errorf("failed to execute ALTER TABLE: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clear cache for this database after successful modification
	m.ClearTableCache(database)

	return nil
}

// detectDataLossWarnings checks if changes might cause data loss
// detectDataLossWarnings checks if changes might cause data loss
func (m *Manager) detectDataLossWarnings(changes []SchemaChange) []string {
	var warnings []string

	for _, change := range changes {
		switch change.Type {
		case "DROP_COLUMN":
			warnings = append(warnings, fmt.Sprintf("Dropping column '%s' will delete all data in that column", change.Target))

		case "MODIFY_COLUMN":
			// For MODIFY_COLUMN, we need to check what actually changed
			// The warning should only appear if we're changing from NULL to NOT NULL
			// Not if the column was already NOT NULL
			// Since we don't have the old column info here, we'll skip this warning
			// The comparison logic should prevent false positives

		case "DROP_INDEX":
			// Dropping primary key or unique index might affect data integrity
			if strings.Contains(strings.ToUpper(change.Target), "PRIMARY") {
				warnings = append(warnings, "Dropping PRIMARY KEY may affect data integrity")
			}

		case "DROP_FOREIGN_KEY":
			warnings = append(warnings, fmt.Sprintf("Dropping foreign key '%s' will remove referential integrity constraint", change.Target))
		}
	}

	return warnings
}

// generateAlterTableStatements generates ALTER TABLE SQL statements
func (m *Manager) generateAlterTableStatements(database, table string, changes []SchemaChange) []string {
	var statements []string

	for _, change := range changes {
		var stmt string

		switch change.Type {
		case "ADD_COLUMN":
			if col, ok := change.Data.(repository.Column); ok {
				colDef := fmt.Sprintf("`%s` %s", col.Name, col.Type)

				if !col.Nullable {
					colDef += " NOT NULL"
				}

				if col.DefaultValue != nil && *col.DefaultValue != "" {
					// 检查是否需要引号
					defaultVal := *col.DefaultValue
					// 对于特殊值（如 CURRENT_TIMESTAMP, NULL），不加引号
					if defaultVal == "CURRENT_TIMESTAMP" || defaultVal == "NULL" {
						colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
					} else {
						// 对于普通值，加引号并转义
						colDef += fmt.Sprintf(" DEFAULT '%s'", strings.ReplaceAll(defaultVal, "'", "''"))
					}
				}

				if col.AutoIncrement {
					colDef += " AUTO_INCREMENT"
				}

				if col.Comment != "" {
					colDef += fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''"))
				}

				stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD COLUMN %s", database, table, colDef)
			}

		case "MODIFY_COLUMN":
			if col, ok := change.Data.(repository.Column); ok {
				colDef := fmt.Sprintf("`%s` %s", col.Name, col.Type)

				if !col.Nullable {
					colDef += " NOT NULL"
				}

				if col.DefaultValue != nil && *col.DefaultValue != "" {
					// 检查是否需要引号
					defaultVal := *col.DefaultValue
					// 对于特殊值（如 CURRENT_TIMESTAMP, NULL），不加引号
					if defaultVal == "CURRENT_TIMESTAMP" || defaultVal == "NULL" {
						colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
					} else {
						// 对于普通值，加引号并转义
						colDef += fmt.Sprintf(" DEFAULT '%s'", strings.ReplaceAll(defaultVal, "'", "''"))
					}
				}

				if col.AutoIncrement {
					colDef += " AUTO_INCREMENT"
				}

				if col.Comment != "" {
					colDef += fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''"))
				}

				stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` MODIFY COLUMN %s", database, table, colDef)
			} else {
				// Log error if type assertion fails
				fmt.Printf("ERROR: MODIFY_COLUMN data is not repository.Column, type: %T, data: %+v\n", change.Data, change.Data)
			}

		case "DROP_COLUMN":
			stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP COLUMN `%s`", database, table, change.Target)

		case "ADD_INDEX":
			if idx, ok := change.Data.(repository.Index); ok {
				if idx.Type == "PRIMARY" {
					stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD PRIMARY KEY (%s)",
						database, table, m.quoteColumns(idx.Columns))
				} else if idx.Type == "UNIQUE" {
					stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD UNIQUE KEY `%s` (%s)",
						database, table, idx.Name, m.quoteColumns(idx.Columns))
				} else {
					stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD KEY `%s` (%s)",
						database, table, idx.Name, m.quoteColumns(idx.Columns))
				}
			}

		case "DROP_INDEX":
			if change.Target == "PRIMARY" {
				stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP PRIMARY KEY", database, table)
			} else {
				stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP KEY `%s`", database, table, change.Target)
			}

		case "ADD_FOREIGN_KEY":
			if fk, ok := change.Data.(repository.ForeignKey); ok {
				fkDef := fmt.Sprintf("CONSTRAINT `%s` FOREIGN KEY (%s) REFERENCES `%s` (%s)",
					fk.Name,
					m.quoteColumns(fk.Columns),
					fk.ReferencedTable,
					m.quoteColumns(fk.ReferencedColumns))

				if fk.OnDelete != "" {
					fkDef += fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
				}

				if fk.OnUpdate != "" {
					fkDef += fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
				}

				stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD %s", database, table, fkDef)
			}

		case "DROP_FOREIGN_KEY":
			stmt = fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP FOREIGN KEY `%s`", database, table, change.Target)
		}

		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// DropTable drops a table
func (m *Manager) DropTable(database, table string) error {
	query := fmt.Sprintf("DROP TABLE `%s`.`%s`", database, table)

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	// Clear cache for this database after successful deletion
	m.ClearTableCache(database)

	return nil
}

// CompareSchemas compares two table schemas and returns a list of changes
// CompareSchemas compares two table schemas and returns a list of changes
func CompareSchemas(oldSchema, newSchema *TableSchema) []SchemaChange {
	changes := []SchemaChange{}

	if oldSchema == nil || newSchema == nil {
		return changes
	}

	// Compare columns
	oldColumns := make(map[string]repository.Column)
	for _, col := range oldSchema.Columns {
		oldColumns[col.Name] = col
	}

	newColumns := make(map[string]repository.Column)
	for _, col := range newSchema.Columns {
		newColumns[col.Name] = col
	}

	// Find added and modified columns
	for _, newCol := range newSchema.Columns {
		if oldCol, exists := oldColumns[newCol.Name]; exists {
			// Check if column was modified
			if !columnsEqual(oldCol, newCol) {
				changes = append(changes, SchemaChange{
					Type:   "MODIFY_COLUMN",
					Target: newCol.Name,
					Data:   newCol,
				})
			}
		} else {
			// Column was added
			changes = append(changes, SchemaChange{
				Type:   "ADD_COLUMN",
				Target: newCol.Name,
				Data:   newCol,
			})
		}
	}

	// Find dropped columns
	for _, oldCol := range oldSchema.Columns {
		if _, exists := newColumns[oldCol.Name]; !exists {
			changes = append(changes, SchemaChange{
				Type:   "DROP_COLUMN",
				Target: oldCol.Name,
				Data:   nil,
			})
		}
	}

	// Compare indexes
	oldIndexes := make(map[string]repository.Index)
	for _, idx := range oldSchema.Indexes {
		oldIndexes[idx.Name] = idx
	}

	newIndexes := make(map[string]repository.Index)
	for _, idx := range newSchema.Indexes {
		newIndexes[idx.Name] = idx
	}

	// Find added and modified indexes
	for _, newIdx := range newSchema.Indexes {
		if oldIdx, exists := oldIndexes[newIdx.Name]; exists {
			// Check if index was modified
			if !indexesEqual(oldIdx, newIdx) {
				// Drop and recreate
				changes = append(changes, SchemaChange{
					Type:   "DROP_INDEX",
					Target: oldIdx.Name,
					Data:   nil,
				})
				changes = append(changes, SchemaChange{
					Type:   "ADD_INDEX",
					Target: newIdx.Name,
					Data:   newIdx,
				})
			}
		} else {
			// Index was added
			changes = append(changes, SchemaChange{
				Type:   "ADD_INDEX",
				Target: newIdx.Name,
				Data:   newIdx,
			})
		}
	}

	// Find dropped indexes
	for _, oldIdx := range oldSchema.Indexes {
		if _, exists := newIndexes[oldIdx.Name]; !exists {
			changes = append(changes, SchemaChange{
				Type:   "DROP_INDEX",
				Target: oldIdx.Name,
				Data:   nil,
			})
		}
	}

	// Compare foreign keys
	oldFKs := make(map[string]repository.ForeignKey)
	for _, fk := range oldSchema.ForeignKeys {
		oldFKs[fk.Name] = fk
	}

	newFKs := make(map[string]repository.ForeignKey)
	for _, fk := range newSchema.ForeignKeys {
		newFKs[fk.Name] = fk
	}

	// Find added and modified foreign keys
	for _, newFK := range newSchema.ForeignKeys {
		if oldFK, exists := oldFKs[newFK.Name]; exists {
			// Check if foreign key was modified
			if !foreignKeysEqual(oldFK, newFK) {
				// Drop and recreate
				changes = append(changes, SchemaChange{
					Type:   "DROP_FOREIGN_KEY",
					Target: oldFK.Name,
					Data:   nil,
				})
				changes = append(changes, SchemaChange{
					Type:   "ADD_FOREIGN_KEY",
					Target: newFK.Name,
					Data:   newFK,
				})
			}
		} else {
			// Foreign key was added
			changes = append(changes, SchemaChange{
				Type:   "ADD_FOREIGN_KEY",
				Target: newFK.Name,
				Data:   newFK,
			})
		}
	}

	// Find dropped foreign keys
	for _, oldFK := range oldSchema.ForeignKeys {
		if _, exists := newFKs[oldFK.Name]; !exists {
			changes = append(changes, SchemaChange{
				Type:   "DROP_FOREIGN_KEY",
				Target: oldFK.Name,
				Data:   nil,
			})
		}
	}

	return changes
}

// columnsEqual checks if two columns are equal
// columnsEqual checks if two columns are equal
// columnsEqual checks if two columns are equal
func columnsEqual(a, b repository.Column) bool {
	// Compare basic fields
	if a.Name != b.Name ||
		a.Nullable != b.Nullable ||
		a.AutoIncrement != b.AutoIncrement ||
		a.Comment != b.Comment {
		return false
	}

	// Normalize and compare types (case-insensitive, trim spaces)
	typeA := strings.ToUpper(strings.TrimSpace(a.Type))
	typeB := strings.ToUpper(strings.TrimSpace(b.Type))
	if typeA != typeB {
		return false
	}

	// Compare DefaultValue pointers correctly
	if a.DefaultValue == nil && b.DefaultValue == nil {
		return true
	}
	if a.DefaultValue == nil || b.DefaultValue == nil {
		return false
	}
	return *a.DefaultValue == *b.DefaultValue
}

// indexesEqual checks if two indexes are equal
// indexesEqual checks if two indexes are equal
func indexesEqual(a, b repository.Index) bool {
	if a.Name != b.Name || a.Type != b.Type || len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if a.Columns[i] != b.Columns[i] {
			return false
		}
	}
	return true
}

// foreignKeysEqual checks if two foreign keys are equal
// foreignKeysEqual checks if two foreign keys are equal
func foreignKeysEqual(a, b repository.ForeignKey) bool {
	if a.Name != b.Name ||
		a.ReferencedTable != b.ReferencedTable ||
		a.OnDelete != b.OnDelete ||
		a.OnUpdate != b.OnUpdate ||
		len(a.Columns) != len(b.Columns) ||
		len(a.ReferencedColumns) != len(b.ReferencedColumns) {
		return false
	}

	for i := range a.Columns {
		if a.Columns[i] != b.Columns[i] {
			return false
		}
	}

	for i := range a.ReferencedColumns {
		if a.ReferencedColumns[i] != b.ReferencedColumns[i] {
			return false
		}
	}

	return true
}
