package types

// ImportResult represents the result of a data import operation
type ImportResult struct {
	TotalRows   int           `json:"totalRows"`
	SuccessRows int           `json:"successRows"`
	FailedRows  int           `json:"failedRows"`
	Errors      []ImportError `json:"errors"`
}

// ImportError represents an error during import
type ImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ColumnMapping represents mapping between file columns and table columns
type ColumnMapping struct {
	FileColumns  []string `json:"fileColumns"`
	TableColumns []string `json:"tableColumns"`
}

// ExportOptions represents options for data export
type ExportOptions struct {
	Format   string    `json:"format"` // SQL, CSV, JSON
	Database string    `json:"database"`
	Table    string    `json:"table"`
	Query    DataQuery `json:"query"`
}
