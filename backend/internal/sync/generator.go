package sync

import (
	"database/sql"
	"fmt"
	"strings"

	"mygui/backend/internal/repository"
	"mygui/backend/types"
)

// ScriptGenerator generates SQL synchronization scripts
type ScriptGenerator struct {
	sourceRepo *repository.SchemaRepository
}

// NewScriptGenerator creates a new ScriptGenerator
func NewScriptGenerator(sourceDB *sql.DB) *ScriptGenerator {
	return &ScriptGenerator{
		sourceRepo: repository.NewSchemaRepository(sourceDB),
	}
}

// GenerateSyncScript generates a synchronization script from a schema diff
func (g *ScriptGenerator) GenerateSyncScript(sourceDB string, diff *types.SchemaDiff) (*types.SyncScript, error) {
	script := &types.SyncScript{
		Statements: []types.SQLStatement{},
	}

	// Step 1: Drop foreign keys that will be removed or modified
	for _, tableDiff := range diff.TableDifferences {
		for _, fk := range tableDiff.ForeignKeysRemoved {
			stmt := types.SQLStatement{
				SQL:         fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY `%s`;", tableDiff.TableName, fk.Name),
				Type:        "ALTER",
				Description: fmt.Sprintf("Drop foreign key %s from table %s", fk.Name, tableDiff.TableName),
			}
			script.Statements = append(script.Statements, stmt)
		}
	}

	// Step 2: Drop tables that only exist in target
	for _, tableName := range diff.TablesOnlyInTarget {
		stmt := types.SQLStatement{
			SQL:         fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName),
			Type:        "DROP",
			Description: fmt.Sprintf("Drop table %s (only exists in target)", tableName),
		}
		script.Statements = append(script.Statements, stmt)
	}

	// Step 3: Create tables that only exist in source
	for _, tableName := range diff.TablesOnlyInSource {
		createSQL, err := g.generateCreateTableSQL(sourceDB, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate CREATE TABLE for %s: %w", tableName, err)
		}

		stmt := types.SQLStatement{
			SQL:         createSQL,
			Type:        "CREATE",
			Description: fmt.Sprintf("Create table %s (only exists in source)", tableName),
		}
		script.Statements = append(script.Statements, stmt)
	}

	// Step 4: Alter existing tables
	for _, tableDiff := range diff.TableDifferences {
		// Drop indexes that will be removed or modified
		for _, idx := range tableDiff.IndexesRemoved {
			if idx.Name != "PRIMARY" {
				stmt := types.SQLStatement{
					SQL:         fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;", tableDiff.TableName, idx.Name),
					Type:        "ALTER",
					Description: fmt.Sprintf("Drop index %s from table %s", idx.Name, tableDiff.TableName),
				}
				script.Statements = append(script.Statements, stmt)
			}
		}

		// Drop columns
		for _, col := range tableDiff.ColumnsRemoved {
			stmt := types.SQLStatement{
				SQL:         fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`;", tableDiff.TableName, col.Name),
				Type:        "ALTER",
				Description: fmt.Sprintf("Drop column %s from table %s", col.Name, tableDiff.TableName),
			}
			script.Statements = append(script.Statements, stmt)
		}

		// Add columns
		for _, col := range tableDiff.ColumnsAdded {
			colDef := g.generateColumnDefinition(col)
			stmt := types.SQLStatement{
				SQL:         fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s;", tableDiff.TableName, colDef),
				Type:        "ALTER",
				Description: fmt.Sprintf("Add column %s to table %s", col.Name, tableDiff.TableName),
			}
			script.Statements = append(script.Statements, stmt)
		}

		// Modify columns
		for _, colDiff := range tableDiff.ColumnsModified {
			colDef := g.generateColumnDefinition(colDiff.NewColumn)
			stmt := types.SQLStatement{
				SQL:         fmt.Sprintf("ALTER TABLE `%s` MODIFY COLUMN %s;", tableDiff.TableName, colDef),
				Type:        "ALTER",
				Description: fmt.Sprintf("Modify column %s in table %s", colDiff.ColumnName, tableDiff.TableName),
			}
			script.Statements = append(script.Statements, stmt)
		}

		// Add indexes
		for _, idx := range tableDiff.IndexesAdded {
			indexSQL := g.generateAddIndexSQL(tableDiff.TableName, idx)
			stmt := types.SQLStatement{
				SQL:         indexSQL,
				Type:        "ALTER",
				Description: fmt.Sprintf("Add index %s to table %s", idx.Name, tableDiff.TableName),
			}
			script.Statements = append(script.Statements, stmt)
		}
	}

	// Step 5: Add foreign keys (after all table structure changes)
	// Sort foreign keys by dependency order
	sortedFKs := g.sortForeignKeysByDependency(diff.TableDifferences)
	for _, fkInfo := range sortedFKs {
		fkSQL := g.generateAddForeignKeySQL(fkInfo.tableName, fkInfo.fk)
		stmt := types.SQLStatement{
			SQL:         fkSQL,
			Type:        "ALTER",
			Description: fmt.Sprintf("Add foreign key %s to table %s", fkInfo.fk.Name, fkInfo.tableName),
		}
		script.Statements = append(script.Statements, stmt)
	}

	return script, nil
}

// generateCreateTableSQL generates a CREATE TABLE statement
func (g *ScriptGenerator) generateCreateTableSQL(database, tableName string) (string, error) {
	// Get table structure
	columns, err := g.sourceRepo.GetColumns(database, tableName)
	if err != nil {
		return "", fmt.Errorf("failed to get columns: %w", err)
	}

	indexes, err := g.sourceRepo.GetIndexes(database, tableName)
	if err != nil {
		return "", fmt.Errorf("failed to get indexes: %w", err)
	}

	foreignKeys, err := g.sourceRepo.GetForeignKeys(database, tableName)
	if err != nil {
		return "", fmt.Errorf("failed to get foreign keys: %w", err)
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("CREATE TABLE `%s` (", tableName))

	// Add columns
	var columnDefs []string
	for _, col := range columns {
		colDef := g.generateColumnDefinition(g.convertRepoColumn(col))
		columnDefs = append(columnDefs, "  "+colDef)
	}

	// Add primary key
	for _, idx := range indexes {
		if idx.Name == "PRIMARY" {
			pkCols := make([]string, len(idx.Columns))
			for i, col := range idx.Columns {
				pkCols[i] = fmt.Sprintf("`%s`", col)
			}
			columnDefs = append(columnDefs, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(pkCols, ", ")))
			break
		}
	}

	// Add other indexes
	for _, idx := range indexes {
		if idx.Name != "PRIMARY" {
			indexDef := g.generateIndexDefinition(idx)
			columnDefs = append(columnDefs, "  "+indexDef)
		}
	}

	// Add foreign keys
	for _, fk := range foreignKeys {
		fkDef := g.generateForeignKeyDefinition(fk)
		columnDefs = append(columnDefs, "  "+fkDef)
	}

	parts = append(parts, strings.Join(columnDefs, ",\n"))
	parts = append(parts, ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")

	return strings.Join(parts, "\n"), nil
}

// generateColumnDefinition generates a column definition
func (g *ScriptGenerator) generateColumnDefinition(col types.Column) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("`%s`", col.Name))
	parts = append(parts, col.Type)

	if !col.Nullable {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}

	if col.DefaultValue != nil {
		if strings.ToUpper(*col.DefaultValue) == "CURRENT_TIMESTAMP" || strings.ToUpper(*col.DefaultValue) == "NULL" {
			parts = append(parts, fmt.Sprintf("DEFAULT %s", *col.DefaultValue))
		} else {
			parts = append(parts, fmt.Sprintf("DEFAULT '%s'", *col.DefaultValue))
		}
	}

	if col.AutoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}

	if col.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''")))
	}

	return strings.Join(parts, " ")
}

// generateIndexDefinition generates an index definition for CREATE TABLE
func (g *ScriptGenerator) generateIndexDefinition(idx repository.Index) string {
	cols := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		cols[i] = fmt.Sprintf("`%s`", col)
	}

	switch idx.Type {
	case "UNIQUE":
		return fmt.Sprintf("UNIQUE KEY `%s` (%s)", idx.Name, strings.Join(cols, ", "))
	case "FULLTEXT":
		return fmt.Sprintf("FULLTEXT KEY `%s` (%s)", idx.Name, strings.Join(cols, ", "))
	default:
		return fmt.Sprintf("KEY `%s` (%s)", idx.Name, strings.Join(cols, ", "))
	}
}

// generateForeignKeyDefinition generates a foreign key definition for CREATE TABLE
func (g *ScriptGenerator) generateForeignKeyDefinition(fk repository.ForeignKey) string {
	cols := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		cols[i] = fmt.Sprintf("`%s`", col)
	}

	refCols := make([]string, len(fk.ReferencedColumns))
	for i, col := range fk.ReferencedColumns {
		refCols[i] = fmt.Sprintf("`%s`", col)
	}

	return fmt.Sprintf("CONSTRAINT `%s` FOREIGN KEY (%s) REFERENCES `%s` (%s) ON DELETE %s ON UPDATE %s",
		fk.Name,
		strings.Join(cols, ", "),
		fk.ReferencedTable,
		strings.Join(refCols, ", "),
		fk.OnDelete,
		fk.OnUpdate,
	)
}

// generateAddIndexSQL generates an ALTER TABLE ADD INDEX statement
func (g *ScriptGenerator) generateAddIndexSQL(tableName string, idx types.Index) string {
	cols := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		cols[i] = fmt.Sprintf("`%s`", col)
	}

	switch idx.Type {
	case "UNIQUE":
		return fmt.Sprintf("ALTER TABLE `%s` ADD UNIQUE KEY `%s` (%s);", tableName, idx.Name, strings.Join(cols, ", "))
	case "FULLTEXT":
		return fmt.Sprintf("ALTER TABLE `%s` ADD FULLTEXT KEY `%s` (%s);", tableName, idx.Name, strings.Join(cols, ", "))
	default:
		return fmt.Sprintf("ALTER TABLE `%s` ADD KEY `%s` (%s);", tableName, idx.Name, strings.Join(cols, ", "))
	}
}

// generateAddForeignKeySQL generates an ALTER TABLE ADD FOREIGN KEY statement
func (g *ScriptGenerator) generateAddForeignKeySQL(tableName string, fk types.ForeignKey) string {
	cols := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		cols[i] = fmt.Sprintf("`%s`", col)
	}

	refCols := make([]string, len(fk.ReferencedColumns))
	for i, col := range fk.ReferencedColumns {
		refCols[i] = fmt.Sprintf("`%s`", col)
	}

	return fmt.Sprintf("ALTER TABLE `%s` ADD CONSTRAINT `%s` FOREIGN KEY (%s) REFERENCES `%s` (%s) ON DELETE %s ON UPDATE %s;",
		tableName,
		fk.Name,
		strings.Join(cols, ", "),
		fk.ReferencedTable,
		strings.Join(refCols, ", "),
		fk.OnDelete,
		fk.OnUpdate,
	)
}

// foreignKeyInfo holds foreign key with its table name
type foreignKeyInfo struct {
	tableName string
	fk        types.ForeignKey
}

// sortForeignKeysByDependency sorts foreign keys to handle dependencies
func (g *ScriptGenerator) sortForeignKeysByDependency(tableDiffs []types.TableDiff) []foreignKeyInfo {
	var result []foreignKeyInfo
	var selfReferencing []foreignKeyInfo

	// Separate self-referencing foreign keys
	for _, tableDiff := range tableDiffs {
		for _, fk := range tableDiff.ForeignKeysAdded {
			fkInfo := foreignKeyInfo{
				tableName: tableDiff.TableName,
				fk:        fk,
			}

			if fk.ReferencedTable == tableDiff.TableName {
				// Self-referencing, add at the end
				selfReferencing = append(selfReferencing, fkInfo)
			} else {
				result = append(result, fkInfo)
			}
		}
	}

	// Add self-referencing foreign keys at the end
	result = append(result, selfReferencing...)

	return result
}

// convertRepoColumn converts repository.Column to types.Column
func (g *ScriptGenerator) convertRepoColumn(col repository.Column) types.Column {
	return types.Column{
		Name:          col.Name,
		Type:          col.Type,
		Nullable:      col.Nullable,
		DefaultValue:  col.DefaultValue,
		AutoIncrement: col.AutoIncrement,
		Comment:       col.Comment,
	}
}
