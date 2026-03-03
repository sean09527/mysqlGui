package sync

import (
	"database/sql"
	"fmt"

	"mygui/backend/internal/repository"
	"mygui/backend/types"
)

// SchemaDiffer compares schemas between two databases
type SchemaDiffer struct {
	sourceRepo *repository.SchemaRepository
	targetRepo *repository.SchemaRepository
}

// NewSchemaDiffer creates a new SchemaDiffer
func NewSchemaDiffer(sourceDB, targetDB *sql.DB) *SchemaDiffer {
	return &SchemaDiffer{
		sourceRepo: repository.NewSchemaRepository(sourceDB),
		targetRepo: repository.NewSchemaRepository(targetDB),
	}
}

// CompareSchemas compares schemas between source and target databases
func (d *SchemaDiffer) CompareSchemas(sourceDB, targetDB string) (*types.SchemaDiff, error) {
	// Get table lists from both databases
	sourceTables, err := d.sourceRepo.ListTables(sourceDB)
	if err != nil {
		return nil, fmt.Errorf("failed to list source tables: %w", err)
	}

	targetTables, err := d.targetRepo.ListTables(targetDB)
	if err != nil {
		return nil, fmt.Errorf("failed to list target tables: %w", err)
	}

	// Create maps for quick lookup
	sourceTableMap := make(map[string]bool)
	targetTableMap := make(map[string]bool)

	for _, table := range sourceTables {
		sourceTableMap[table.Name] = true
	}

	for _, table := range targetTables {
		targetTableMap[table.Name] = true
	}

	diff := &types.SchemaDiff{
		TablesOnlyInSource: []string{},
		TablesOnlyInTarget: []string{},
		TableDifferences:   []types.TableDiff{},
	}

	// Find tables only in source
	for _, table := range sourceTables {
		if !targetTableMap[table.Name] {
			diff.TablesOnlyInSource = append(diff.TablesOnlyInSource, table.Name)
		}
	}

	// Find tables only in target
	for _, table := range targetTables {
		if !sourceTableMap[table.Name] {
			diff.TablesOnlyInTarget = append(diff.TablesOnlyInTarget, table.Name)
		}
	}

	// Compare tables that exist in both databases
	for _, table := range sourceTables {
		if targetTableMap[table.Name] {
			tableDiff, err := d.compareTable(sourceDB, targetDB, table.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to compare table %s: %w", table.Name, err)
			}

			// Only add to differences if there are actual differences
			if d.hasTableDifferences(tableDiff) {
				diff.TableDifferences = append(diff.TableDifferences, *tableDiff)
			}
		}
	}

	return diff, nil
}

// compareTable compares a single table between source and target
func (d *SchemaDiffer) compareTable(sourceDB, targetDB, tableName string) (*types.TableDiff, error) {
	tableDiff := &types.TableDiff{
		TableName:          tableName,
		ColumnsAdded:       []types.Column{},
		ColumnsRemoved:     []types.Column{},
		ColumnsModified:    []types.ColumnDiff{},
		IndexesAdded:       []types.Index{},
		IndexesRemoved:     []types.Index{},
		ForeignKeysAdded:   []types.ForeignKey{},
		ForeignKeysRemoved: []types.ForeignKey{},
	}

	// Compare columns
	if err := d.compareColumns(sourceDB, targetDB, tableName, tableDiff); err != nil {
		return nil, fmt.Errorf("failed to compare columns: %w", err)
	}

	// Compare indexes
	if err := d.compareIndexes(sourceDB, targetDB, tableName, tableDiff); err != nil {
		return nil, fmt.Errorf("failed to compare indexes: %w", err)
	}

	// Compare foreign keys
	if err := d.compareForeignKeys(sourceDB, targetDB, tableName, tableDiff); err != nil {
		return nil, fmt.Errorf("failed to compare foreign keys: %w", err)
	}

	return tableDiff, nil
}

// compareColumns compares columns between source and target table
func (d *SchemaDiffer) compareColumns(sourceDB, targetDB, tableName string, tableDiff *types.TableDiff) error {
	sourceColumns, err := d.sourceRepo.GetColumns(sourceDB, tableName)
	if err != nil {
		return fmt.Errorf("failed to get source columns: %w", err)
	}

	targetColumns, err := d.targetRepo.GetColumns(targetDB, tableName)
	if err != nil {
		return fmt.Errorf("failed to get target columns: %w", err)
	}

	// Create maps for quick lookup
	sourceColMap := make(map[string]repository.Column)
	targetColMap := make(map[string]repository.Column)

	for _, col := range sourceColumns {
		sourceColMap[col.Name] = col
	}

	for _, col := range targetColumns {
		targetColMap[col.Name] = col
	}

	// Find columns only in source (added)
	for _, col := range sourceColumns {
		if _, exists := targetColMap[col.Name]; !exists {
			tableDiff.ColumnsAdded = append(tableDiff.ColumnsAdded, d.convertColumn(col))
		}
	}

	// Find columns only in target (removed)
	for _, col := range targetColumns {
		if _, exists := sourceColMap[col.Name]; !exists {
			tableDiff.ColumnsRemoved = append(tableDiff.ColumnsRemoved, d.convertColumn(col))
		}
	}

	// Find modified columns
	for _, sourceCol := range sourceColumns {
		if targetCol, exists := targetColMap[sourceCol.Name]; exists {
			if !d.columnsEqual(sourceCol, targetCol) {
				tableDiff.ColumnsModified = append(tableDiff.ColumnsModified, types.ColumnDiff{
					ColumnName: sourceCol.Name,
					OldColumn:  d.convertColumn(targetCol),
					NewColumn:  d.convertColumn(sourceCol),
				})
			}
		}
	}

	return nil
}

// compareIndexes compares indexes between source and target table
func (d *SchemaDiffer) compareIndexes(sourceDB, targetDB, tableName string, tableDiff *types.TableDiff) error {
	sourceIndexes, err := d.sourceRepo.GetIndexes(sourceDB, tableName)
	if err != nil {
		return fmt.Errorf("failed to get source indexes: %w", err)
	}

	targetIndexes, err := d.targetRepo.GetIndexes(targetDB, tableName)
	if err != nil {
		return fmt.Errorf("failed to get target indexes: %w", err)
	}

	// Create maps for quick lookup
	sourceIdxMap := make(map[string]repository.Index)
	targetIdxMap := make(map[string]repository.Index)

	for _, idx := range sourceIndexes {
		sourceIdxMap[idx.Name] = idx
	}

	for _, idx := range targetIndexes {
		targetIdxMap[idx.Name] = idx
	}

	// Find indexes only in source (added)
	for _, idx := range sourceIndexes {
		if _, exists := targetIdxMap[idx.Name]; !exists {
			tableDiff.IndexesAdded = append(tableDiff.IndexesAdded, d.convertIndex(idx))
		}
	}

	// Find indexes only in target (removed)
	for _, idx := range targetIndexes {
		if _, exists := sourceIdxMap[idx.Name]; !exists {
			tableDiff.IndexesRemoved = append(tableDiff.IndexesRemoved, d.convertIndex(idx))
		}
	}

	// Check for modified indexes (same name but different definition)
	for _, sourceIdx := range sourceIndexes {
		if targetIdx, exists := targetIdxMap[sourceIdx.Name]; exists {
			if !d.indexesEqual(sourceIdx, targetIdx) {
				// Treat as remove + add
				tableDiff.IndexesRemoved = append(tableDiff.IndexesRemoved, d.convertIndex(targetIdx))
				tableDiff.IndexesAdded = append(tableDiff.IndexesAdded, d.convertIndex(sourceIdx))
			}
		}
	}

	return nil
}

// compareForeignKeys compares foreign keys between source and target table
func (d *SchemaDiffer) compareForeignKeys(sourceDB, targetDB, tableName string, tableDiff *types.TableDiff) error {
	sourceFKs, err := d.sourceRepo.GetForeignKeys(sourceDB, tableName)
	if err != nil {
		return fmt.Errorf("failed to get source foreign keys: %w", err)
	}

	targetFKs, err := d.targetRepo.GetForeignKeys(targetDB, tableName)
	if err != nil {
		return fmt.Errorf("failed to get target foreign keys: %w", err)
	}

	// Create maps for quick lookup
	sourceFKMap := make(map[string]repository.ForeignKey)
	targetFKMap := make(map[string]repository.ForeignKey)

	for _, fk := range sourceFKs {
		sourceFKMap[fk.Name] = fk
	}

	for _, fk := range targetFKs {
		targetFKMap[fk.Name] = fk
	}

	// Find foreign keys only in source (added)
	for _, fk := range sourceFKs {
		if _, exists := targetFKMap[fk.Name]; !exists {
			tableDiff.ForeignKeysAdded = append(tableDiff.ForeignKeysAdded, d.convertForeignKey(fk))
		}
	}

	// Find foreign keys only in target (removed)
	for _, fk := range targetFKs {
		if _, exists := sourceFKMap[fk.Name]; !exists {
			tableDiff.ForeignKeysRemoved = append(tableDiff.ForeignKeysRemoved, d.convertForeignKey(fk))
		}
	}

	// Check for modified foreign keys (same name but different definition)
	for _, sourceFk := range sourceFKs {
		if targetFk, exists := targetFKMap[sourceFk.Name]; exists {
			if !d.foreignKeysEqual(sourceFk, targetFk) {
				// Treat as remove + add
				tableDiff.ForeignKeysRemoved = append(tableDiff.ForeignKeysRemoved, d.convertForeignKey(targetFk))
				tableDiff.ForeignKeysAdded = append(tableDiff.ForeignKeysAdded, d.convertForeignKey(sourceFk))
			}
		}
	}

	return nil
}

// Helper functions for comparison

func (d *SchemaDiffer) columnsEqual(col1, col2 repository.Column) bool {
	if col1.Name != col2.Name {
		return false
	}
	if col1.Type != col2.Type {
		return false
	}
	if col1.Nullable != col2.Nullable {
		return false
	}
	if col1.AutoIncrement != col2.AutoIncrement {
		return false
	}

	// Compare default values
	if (col1.DefaultValue == nil) != (col2.DefaultValue == nil) {
		return false
	}
	if col1.DefaultValue != nil && col2.DefaultValue != nil {
		if *col1.DefaultValue != *col2.DefaultValue {
			return false
		}
	}

	return true
}

func (d *SchemaDiffer) indexesEqual(idx1, idx2 repository.Index) bool {
	if idx1.Name != idx2.Name {
		return false
	}
	if idx1.Type != idx2.Type {
		return false
	}
	if len(idx1.Columns) != len(idx2.Columns) {
		return false
	}
	for i := range idx1.Columns {
		if idx1.Columns[i] != idx2.Columns[i] {
			return false
		}
	}
	return true
}

func (d *SchemaDiffer) foreignKeysEqual(fk1, fk2 repository.ForeignKey) bool {
	if fk1.Name != fk2.Name {
		return false
	}
	if fk1.ReferencedTable != fk2.ReferencedTable {
		return false
	}
	if fk1.OnDelete != fk2.OnDelete {
		return false
	}
	if fk1.OnUpdate != fk2.OnUpdate {
		return false
	}
	if len(fk1.Columns) != len(fk2.Columns) {
		return false
	}
	for i := range fk1.Columns {
		if fk1.Columns[i] != fk2.Columns[i] {
			return false
		}
	}
	if len(fk1.ReferencedColumns) != len(fk2.ReferencedColumns) {
		return false
	}
	for i := range fk1.ReferencedColumns {
		if fk1.ReferencedColumns[i] != fk2.ReferencedColumns[i] {
			return false
		}
	}
	return true
}

func (d *SchemaDiffer) hasTableDifferences(tableDiff *types.TableDiff) bool {
	return len(tableDiff.ColumnsAdded) > 0 ||
		len(tableDiff.ColumnsRemoved) > 0 ||
		len(tableDiff.ColumnsModified) > 0 ||
		len(tableDiff.IndexesAdded) > 0 ||
		len(tableDiff.IndexesRemoved) > 0 ||
		len(tableDiff.ForeignKeysAdded) > 0 ||
		len(tableDiff.ForeignKeysRemoved) > 0
}

// Conversion functions from repository types to types package

func (d *SchemaDiffer) convertColumn(col repository.Column) types.Column {
	return types.Column{
		Name:          col.Name,
		Type:          col.Type,
		Nullable:      col.Nullable,
		DefaultValue:  col.DefaultValue,
		AutoIncrement: col.AutoIncrement,
		Comment:       col.Comment,
	}
}

func (d *SchemaDiffer) convertIndex(idx repository.Index) types.Index {
	return types.Index{
		Name:    idx.Name,
		Type:    idx.Type,
		Columns: idx.Columns,
	}
}

func (d *SchemaDiffer) convertForeignKey(fk repository.ForeignKey) types.ForeignKey {
	return types.ForeignKey{
		Name:              fk.Name,
		Columns:           fk.Columns,
		ReferencedTable:   fk.ReferencedTable,
		ReferencedColumns: fk.ReferencedColumns,
		OnDelete:          fk.OnDelete,
		OnUpdate:          fk.OnUpdate,
	}
}
