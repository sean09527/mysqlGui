package types

// Database represents a database
type Database struct {
	Name string `json:"name"`
}

// Table represents a table with metadata
type Table struct {
	Name    string `json:"name"`
	Rows    int64  `json:"rows"`
	Engine  string `json:"engine"`
	Comment string `json:"comment"`
}

// TableSchema represents the complete schema of a table
type TableSchema struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	PrimaryKey  *PrimaryKey  `json:"primaryKey,omitempty"`
	Indexes     []Index      `json:"indexes"`
	ForeignKeys []ForeignKey `json:"foreignKeys"`
	Engine      string       `json:"engine"`
	Charset     string       `json:"charset"`
	Comment     string       `json:"comment"`
}

// Column represents a table column
type Column struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Nullable      bool    `json:"nullable"`
	DefaultValue  *string `json:"defaultValue"`
	AutoIncrement bool    `json:"autoIncrement"`
	Comment       string  `json:"comment"`
}

// PrimaryKey represents a primary key constraint
type PrimaryKey struct {
	Columns []string `json:"columns"`
}

// Index represents a table index
type Index struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"` // PRIMARY, UNIQUE, INDEX, FULLTEXT
	Columns []string `json:"columns"`
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Name              string   `json:"name"`
	Columns           []string `json:"columns"`
	ReferencedTable   string   `json:"referencedTable"`
	ReferencedColumns []string `json:"referencedColumns"`
	OnDelete          string   `json:"onDelete"`
	OnUpdate          string   `json:"onUpdate"`
}

// SchemaChange represents a change to a table schema
type SchemaChange struct {
	Type   string      `json:"type"` // ADD_COLUMN, DROP_COLUMN, MODIFY_COLUMN, ADD_INDEX, DROP_INDEX, etc.
	Target string      `json:"target"`
	Data   interface{} `json:"data"`
}
