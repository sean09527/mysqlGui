package repository

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Database represents a database in MySQL
type Database struct {
	Name string `json:"name"`
}

// Table represents a table in a database
type Table struct {
	Name    string `json:"name"`
	Rows    int64  `json:"rows"`
	Engine  string `json:"engine"`
	Comment string `json:"comment"`
}

// Column represents a column in a table
type Column struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Nullable      bool    `json:"nullable"`
	DefaultValue  *string `json:"defaultValue"`
	AutoIncrement bool    `json:"autoIncrement"`
	Comment       string  `json:"comment"`
}

// Index represents an index on a table
type Index struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"` // PRIMARY, UNIQUE, INDEX, FULLTEXT
	Columns   []string `json:"columns"`
	NonUnique bool     `json:"nonUnique"`
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

// SchemaRepository handles database schema queries
type SchemaRepository struct {
	db            *sql.DB
	databaseCache *cacheEntry
	tableCache    map[string]*cacheEntry
	cacheMutex    sync.RWMutex
	cacheDuration time.Duration
}

// cacheEntry stores cached data with timestamp
type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

// NewSchemaRepository creates a new SchemaRepository
func NewSchemaRepository(db *sql.DB) *SchemaRepository {
	return &SchemaRepository{
		db:            db,
		tableCache:    make(map[string]*cacheEntry),
		cacheDuration: 5 * time.Minute, // Cache for 5 minutes
	}
}

// ClearCache clears all cached data
func (r *SchemaRepository) ClearCache() {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	r.databaseCache = nil
	r.tableCache = make(map[string]*cacheEntry)
}

// ClearTableCache clears cached table list for a specific database
func (r *SchemaRepository) ClearTableCache(database string) {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	delete(r.tableCache, database)
}

// isCacheValid checks if cache entry is still valid
func (r *SchemaRepository) isCacheValid(entry *cacheEntry) bool {
	if entry == nil {
		return false
	}
	return time.Since(entry.timestamp) < r.cacheDuration
}

// ListDatabases returns all databases except system databases
func (r *SchemaRepository) ListDatabases() ([]Database, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if r.isCacheValid(r.databaseCache) {
		databases := r.databaseCache.data.([]Database)
		r.cacheMutex.RUnlock()
		return databases, nil
	}
	r.cacheMutex.RUnlock()

	query := `
		SELECT SCHEMA_NAME 
		FROM INFORMATION_SCHEMA.SCHEMATA
		WHERE SCHEMA_NAME NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
		ORDER BY SCHEMA_NAME
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []Database
	for rows.Next() {
		var db Database
		if err := rows.Scan(&db.Name); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		databases = append(databases, db)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating databases: %w", err)
	}

	// Update cache
	r.cacheMutex.Lock()
	r.databaseCache = &cacheEntry{
		data:      databases,
		timestamp: time.Now(),
	}
	r.cacheMutex.Unlock()

	return databases, nil
}

// ListTables returns all tables in a database
func (r *SchemaRepository) ListTables(database string) ([]Table, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if entry, exists := r.tableCache[database]; exists && r.isCacheValid(entry) {
		tables := entry.data.([]Table)
		r.cacheMutex.RUnlock()
		return tables, nil
	}
	r.cacheMutex.RUnlock()

	query := `
		SELECT TABLE_NAME, IFNULL(TABLE_ROWS, 0), IFNULL(ENGINE, ''), IFNULL(TABLE_COMMENT, '')
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`

	rows, err := r.db.Query(query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.Name, &table.Rows, &table.Engine, &table.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	// Update cache
	r.cacheMutex.Lock()
	r.tableCache[database] = &cacheEntry{
		data:      tables,
		timestamp: time.Now(),
	}
	r.cacheMutex.Unlock()

	return tables, nil
}

// GetColumns returns all columns for a table
func (r *SchemaRepository) GetColumns(database, table string) ([]Column, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			EXTRA,
			IFNULL(COLUMN_COMMENT, '')
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := r.db.Query(query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var isNullable string
		var extra string
		var defaultValue sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &isNullable, &defaultValue, &extra, &col.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col.Nullable = isNullable == "YES"
		col.AutoIncrement = extra == "auto_increment"

		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	return columns, nil
}

// GetIndexes returns all indexes for a table
func (r *SchemaRepository) GetIndexes(database, table string) ([]Index, error) {
	query := `
		SELECT 
			INDEX_NAME,
			COLUMN_NAME,
			NON_UNIQUE,
			INDEX_TYPE
		FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`

	rows, err := r.db.Query(query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	// Group columns by index name
	indexMap := make(map[string]*Index)
	var indexOrder []string

	for rows.Next() {
		var indexName, columnName, indexType string
		var nonUnique int

		if err := rows.Scan(&indexName, &columnName, &nonUnique, &indexType); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}

		if _, exists := indexMap[indexName]; !exists {
			indexType := "INDEX"
			if indexName == "PRIMARY" {
				indexType = "PRIMARY"
			} else if nonUnique == 0 {
				indexType = "UNIQUE"
			}

			indexMap[indexName] = &Index{
				Name:      indexName,
				Type:      indexType,
				Columns:   []string{},
				NonUnique: nonUnique == 1,
			}
			indexOrder = append(indexOrder, indexName)
		}

		indexMap[indexName].Columns = append(indexMap[indexName].Columns, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating indexes: %w", err)
	}

	// Convert map to slice maintaining order
	var indexes []Index
	for _, name := range indexOrder {
		indexes = append(indexes, *indexMap[name])
	}

	return indexes, nil
}

// GetForeignKeys returns all foreign keys for a table
func (r *SchemaRepository) GetForeignKeys(database, table string) ([]ForeignKey, error) {
	query := `
		SELECT 
			kcu.CONSTRAINT_NAME,
			kcu.COLUMN_NAME,
			kcu.REFERENCED_TABLE_NAME,
			kcu.REFERENCED_COLUMN_NAME,
			rc.UPDATE_RULE,
			rc.DELETE_RULE
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
		JOIN INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS rc
			ON kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME
			AND kcu.CONSTRAINT_SCHEMA = rc.CONSTRAINT_SCHEMA
		WHERE kcu.TABLE_SCHEMA = ? 
			AND kcu.TABLE_NAME = ?
			AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY kcu.CONSTRAINT_NAME, kcu.ORDINAL_POSITION
	`

	rows, err := r.db.Query(query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	// Group columns by constraint name
	fkMap := make(map[string]*ForeignKey)
	var fkOrder []string

	for rows.Next() {
		var constraintName, columnName, refTable, refColumn, updateRule, deleteRule string

		if err := rows.Scan(&constraintName, &columnName, &refTable, &refColumn, &updateRule, &deleteRule); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}

		if _, exists := fkMap[constraintName]; !exists {
			fkMap[constraintName] = &ForeignKey{
				Name:              constraintName,
				Columns:           []string{},
				ReferencedTable:   refTable,
				ReferencedColumns: []string{},
				OnUpdate:          updateRule,
				OnDelete:          deleteRule,
			}
			fkOrder = append(fkOrder, constraintName)
		}

		fkMap[constraintName].Columns = append(fkMap[constraintName].Columns, columnName)
		fkMap[constraintName].ReferencedColumns = append(fkMap[constraintName].ReferencedColumns, refColumn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign keys: %w", err)
	}

	// Convert map to slice maintaining order
	var foreignKeys []ForeignKey
	for _, name := range fkOrder {
		foreignKeys = append(foreignKeys, *fkMap[name])
	}

	return foreignKeys, nil
}
