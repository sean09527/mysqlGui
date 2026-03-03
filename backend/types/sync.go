package types

// SchemaDiff represents differences between two database schemas
type SchemaDiff struct {
	TablesOnlyInSource []string    `json:"tablesOnlyInSource"`
	TablesOnlyInTarget []string    `json:"tablesOnlyInTarget"`
	TableDifferences   []TableDiff `json:"tableDifferences"`
}

// TableDiff represents differences in a single table
type TableDiff struct {
	TableName          string       `json:"tableName"`
	ColumnsAdded       []Column     `json:"columnsAdded"`
	ColumnsRemoved     []Column     `json:"columnsRemoved"`
	ColumnsModified    []ColumnDiff `json:"columnsModified"`
	IndexesAdded       []Index      `json:"indexesAdded"`
	IndexesRemoved     []Index      `json:"indexesRemoved"`
	ForeignKeysAdded   []ForeignKey `json:"foreignKeysAdded"`
	ForeignKeysRemoved []ForeignKey `json:"foreignKeysRemoved"`
}

// ColumnDiff represents a difference in a column
type ColumnDiff struct {
	ColumnName string `json:"columnName"`
	OldColumn  Column `json:"oldColumn"`
	NewColumn  Column `json:"newColumn"`
}

// SyncScript represents a synchronization script
type SyncScript struct {
	Statements []SQLStatement `json:"statements"`
}

// SQLStatement represents a single SQL statement
type SQLStatement struct {
	SQL         string `json:"sql"`
	Type        string `json:"type"` // CREATE, ALTER, DROP
	Description string `json:"description"`
}
