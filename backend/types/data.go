package types

// DataQuery represents a query for table data
type DataQuery struct {
	Database string    `json:"database"`
	Table    string    `json:"table"`
	Columns  []string  `json:"columns"`
	Filters  []Filter  `json:"filters"`
	OrderBy  []OrderBy `json:"orderBy"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// Filter represents a filter condition
type Filter struct {
	Column   string      `json:"column"`
	Operator string      `json:"operator"` // =, !=, >, <, >=, <=, LIKE, NOT LIKE, IN, NOT IN
	Value    interface{} `json:"value"`
}

// OrderBy represents a sort order
type OrderBy struct {
	Column    string `json:"column"`
	Direction string `json:"direction"` // ASC, DESC
}

// DataResult represents the result of a data query
type DataResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Total   int64           `json:"total"`
}
