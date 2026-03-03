package sync

import (
	"database/sql"
	"fmt"

	"mygui/backend/types"
)

// SyncEngine orchestrates schema synchronization between databases
type SyncEngine struct {
	sourceDB  *sql.DB
	targetDB  *sql.DB
	differ    *SchemaDiffer
	generator *ScriptGenerator
}

// NewSyncEngine creates a new SyncEngine
func NewSyncEngine(sourceDB, targetDB *sql.DB) *SyncEngine {
	return &SyncEngine{
		sourceDB:  sourceDB,
		targetDB:  targetDB,
		differ:    NewSchemaDiffer(sourceDB, targetDB),
		generator: NewScriptGenerator(sourceDB),
	}
}

// CompareSchemas compares schemas between source and target databases
func (e *SyncEngine) CompareSchemas(sourceDB, targetDB string) (*types.SchemaDiff, error) {
	diff, err := e.differ.CompareSchemas(sourceDB, targetDB)
	if err != nil {
		return nil, fmt.Errorf("failed to compare schemas: %w", err)
	}

	return diff, nil
}

// GenerateSyncScript generates a synchronization script from a schema diff
func (e *SyncEngine) GenerateSyncScript(sourceDB string, diff *types.SchemaDiff) (*types.SyncScript, error) {
	script, err := e.generator.GenerateSyncScript(sourceDB, diff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sync script: %w", err)
	}

	return script, nil
}

// ExecuteSyncScript executes a synchronization script on the target database
func (e *SyncEngine) ExecuteSyncScript(targetDB string, script *types.SyncScript, progressCallback func(current, total int, statement string)) error {
	// Start a transaction
	tx, err := e.targetDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Select the target database
	_, err = tx.Exec(fmt.Sprintf("USE `%s`", targetDB))
	if err != nil {
		return fmt.Errorf("failed to select target database %s: %w", targetDB, err)
	}

	// Execute each statement
	total := len(script.Statements)
	for i, stmt := range script.Statements {
		// Report progress
		if progressCallback != nil {
			progressCallback(i+1, total, stmt.Description)
		}

		// Execute the statement
		_, err = tx.Exec(stmt.SQL)
		if err != nil {
			return fmt.Errorf("failed to execute statement %d (%s): %w\nSQL: %s", i+1, stmt.Description, err, stmt.SQL)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ValidateScript validates a synchronization script
func (e *SyncEngine) ValidateScript(script *types.SyncScript) error {
	if script == nil {
		return fmt.Errorf("script is nil")
	}

	if len(script.Statements) == 0 {
		return fmt.Errorf("script has no statements")
	}

	// Basic validation of each statement
	for i, stmt := range script.Statements {
		if stmt.SQL == "" {
			return fmt.Errorf("statement %d has empty SQL", i+1)
		}

		if stmt.Type == "" {
			return fmt.Errorf("statement %d has empty type", i+1)
		}

		// Validate statement type
		validTypes := map[string]bool{
			"CREATE": true,
			"ALTER":  true,
			"DROP":   true,
		}

		if !validTypes[stmt.Type] {
			return fmt.Errorf("statement %d has invalid type: %s", i+1, stmt.Type)
		}
	}

	return nil
}

// CompareAndGenerateScript is a convenience method that combines CompareSchemas and GenerateSyncScript
func (e *SyncEngine) CompareAndGenerateScript(sourceDB, targetDB string) (*types.SchemaDiff, *types.SyncScript, error) {
	// Compare schemas
	diff, err := e.CompareSchemas(sourceDB, targetDB)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compare schemas: %w", err)
	}

	// Generate sync script
	script, err := e.GenerateSyncScript(sourceDB, diff)
	if err != nil {
		return diff, nil, fmt.Errorf("failed to generate sync script: %w", err)
	}

	return diff, script, nil
}

// ExecuteSync is a convenience method that compares, generates, and executes synchronization
func (e *SyncEngine) ExecuteSync(sourceDB, targetDB string, progressCallback func(current, total int, statement string)) (*types.SchemaDiff, error) {
	// Compare and generate script
	diff, script, err := e.CompareAndGenerateScript(sourceDB, targetDB)
	if err != nil {
		return nil, err
	}

	// Validate script
	if err := e.ValidateScript(script); err != nil {
		return diff, fmt.Errorf("script validation failed: %w", err)
	}

	// Execute script
	if err := e.ExecuteSyncScript(targetDB, script, progressCallback); err != nil {
		return diff, fmt.Errorf("failed to execute sync script: %w", err)
	}

	return diff, nil
}

// GetSyncSummary returns a human-readable summary of the sync operations
func (e *SyncEngine) GetSyncSummary(diff *types.SchemaDiff) string {
	summary := "Schema Synchronization Summary:\n\n"

	if len(diff.TablesOnlyInSource) > 0 {
		summary += fmt.Sprintf("Tables to create: %d\n", len(diff.TablesOnlyInSource))
		for _, table := range diff.TablesOnlyInSource {
			summary += fmt.Sprintf("  - %s\n", table)
		}
		summary += "\n"
	}

	if len(diff.TablesOnlyInTarget) > 0 {
		summary += fmt.Sprintf("Tables to drop: %d\n", len(diff.TablesOnlyInTarget))
		for _, table := range diff.TablesOnlyInTarget {
			summary += fmt.Sprintf("  - %s\n", table)
		}
		summary += "\n"
	}

	if len(diff.TableDifferences) > 0 {
		summary += fmt.Sprintf("Tables to alter: %d\n", len(diff.TableDifferences))
		for _, tableDiff := range diff.TableDifferences {
			summary += fmt.Sprintf("  - %s:\n", tableDiff.TableName)

			if len(tableDiff.ColumnsAdded) > 0 {
				summary += fmt.Sprintf("      Columns to add: %d\n", len(tableDiff.ColumnsAdded))
			}
			if len(tableDiff.ColumnsRemoved) > 0 {
				summary += fmt.Sprintf("      Columns to remove: %d\n", len(tableDiff.ColumnsRemoved))
			}
			if len(tableDiff.ColumnsModified) > 0 {
				summary += fmt.Sprintf("      Columns to modify: %d\n", len(tableDiff.ColumnsModified))
			}
			if len(tableDiff.IndexesAdded) > 0 {
				summary += fmt.Sprintf("      Indexes to add: %d\n", len(tableDiff.IndexesAdded))
			}
			if len(tableDiff.IndexesRemoved) > 0 {
				summary += fmt.Sprintf("      Indexes to remove: %d\n", len(tableDiff.IndexesRemoved))
			}
			if len(tableDiff.ForeignKeysAdded) > 0 {
				summary += fmt.Sprintf("      Foreign keys to add: %d\n", len(tableDiff.ForeignKeysAdded))
			}
			if len(tableDiff.ForeignKeysRemoved) > 0 {
				summary += fmt.Sprintf("      Foreign keys to remove: %d\n", len(tableDiff.ForeignKeysRemoved))
			}
		}
		summary += "\n"
	}

	if len(diff.TablesOnlyInSource) == 0 && len(diff.TablesOnlyInTarget) == 0 && len(diff.TableDifferences) == 0 {
		summary += "No differences found. Schemas are identical.\n"
	}

	return summary
}
