package repository

import (
	"database/sql"
	"fmt"
	"strings"
)

// DataRepository 处理数据访问操作
type DataRepository struct {
	db *sql.DB
}

// NewDataRepository 创建新的 DataRepository 实例
func NewDataRepository(db *sql.DB) *DataRepository {
	return &DataRepository{db: db}
}

// Filter 表示数据筛选条件
type Filter struct {
	Column   string
	Operator string // =, !=, >, <, >=, <=, LIKE, NOT LIKE, IN, NOT IN
	Value    interface{}
}

// OrderBy 表示排序条件
type OrderBy struct {
	Column    string
	Direction string // ASC, DESC
}

// DataQuery 表示数据查询请求
type DataQuery struct {
	Database string
	Table    string
	Columns  []string
	Filters  []Filter
	OrderBy  []OrderBy
	Limit    int
	Offset   int
}

// DataResult 表示查询结果
type DataResult struct {
	Columns []string
	Rows    [][]interface{}
	Total   int64
}

// QueryData 执行参数化查询并返回结果
func (dr *DataRepository) QueryData(query DataQuery) (*DataResult, error) {
	// 构建 SELECT 语句
	selectClause := "*"
	if len(query.Columns) > 0 {
		// 转义列名
		escapedColumns := make([]string, len(query.Columns))
		for i, col := range query.Columns {
			escapedColumns[i] = escapeIdentifier(col)
		}
		selectClause = strings.Join(escapedColumns, ", ")
	}

	// 构建基础查询
	sqlQuery := fmt.Sprintf("SELECT %s FROM %s.%s",
		selectClause,
		escapeIdentifier(query.Database),
		escapeIdentifier(query.Table))

	// 构建 WHERE 子句和参数
	args := []interface{}{}
	if len(query.Filters) > 0 {
		whereClause, whereArgs, err := dr.buildWhereClause(query.Filters)
		if err != nil {
			return nil, fmt.Errorf("构建 WHERE 子句失败: %w", err)
		}
		sqlQuery += " WHERE " + whereClause
		args = append(args, whereArgs...)
	}

	// 构建 ORDER BY 子句
	if len(query.OrderBy) > 0 {
		orderClauses := make([]string, len(query.OrderBy))
		for i, order := range query.OrderBy {
			direction := "ASC"
			if strings.ToUpper(order.Direction) == "DESC" {
				direction = "DESC"
			}
			orderClauses[i] = fmt.Sprintf("%s %s", escapeIdentifier(order.Column), direction)
		}
		sqlQuery += " ORDER BY " + strings.Join(orderClauses, ", ")
	}

	// 添加 LIMIT 和 OFFSET
	if query.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT %d", query.Limit)
	}
	if query.Offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET %d", query.Offset)
	}

	// 执行查询
	rows, err := dr.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %w", err)
	}

	// 读取数据行
	var dataRows [][]interface{}
	for rows.Next() {
		// 创建扫描目标
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描行数据失败: %w", err)
		}

		// 转换字节数组为字符串
		row := make([]interface{}, len(columns))
		for i, val := range values {
			if b, ok := val.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = val
			}
		}
		dataRows = append(dataRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果集失败: %w", err)
	}

	// 获取总行数
	total, err := dr.GetRowCount(query.Database, query.Table, query.Filters)
	if err != nil {
		return nil, fmt.Errorf("获取总行数失败: %w", err)
	}

	return &DataResult{
		Columns: columns,
		Rows:    dataRows,
		Total:   total,
	}, nil
}

// GetRowCount 获取表的行数（支持筛选）
func (dr *DataRepository) GetRowCount(database, table string, filters []Filter) (int64, error) {
	sqlQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s",
		escapeIdentifier(database),
		escapeIdentifier(table))

	args := []interface{}{}
	if len(filters) > 0 {
		whereClause, whereArgs, err := dr.buildWhereClause(filters)
		if err != nil {
			return 0, fmt.Errorf("构建 WHERE 子句失败: %w", err)
		}
		sqlQuery += " WHERE " + whereClause
		args = append(args, whereArgs...)
	}

	var count int64
	err := dr.db.QueryRow(sqlQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询行数失败: %w", err)
	}

	return count, nil
}

// buildWhereClause 构建参数化的 WHERE 子句
func (dr *DataRepository) buildWhereClause(filters []Filter) (string, []interface{}, error) {
	if len(filters) == 0 {
		return "", nil, nil
	}

	var clauses []string
	var args []interface{}

	for _, filter := range filters {
		clause, filterArgs, err := dr.buildFilterClause(filter)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		args = append(args, filterArgs...)
	}

	return strings.Join(clauses, " AND "), args, nil
}

// buildFilterClause 构建单个筛选条件的子句
func (dr *DataRepository) buildFilterClause(filter Filter) (string, []interface{}, error) {
	column := escapeIdentifier(filter.Column)
	operator := strings.ToUpper(strings.TrimSpace(filter.Operator))

	switch operator {
	case "=", "!=", ">", "<", ">=", "<=":
		return fmt.Sprintf("%s %s ?", column, operator), []interface{}{filter.Value}, nil

	case "LIKE", "NOT LIKE":
		return fmt.Sprintf("%s %s ?", column, operator), []interface{}{filter.Value}, nil

	case "IN", "NOT IN":
		// 处理 IN 操作符
		values, ok := filter.Value.([]interface{})
		if !ok {
			return "", nil, fmt.Errorf("IN/NOT IN 操作符需要数组类型的值")
		}
		if len(values) == 0 {
			return "", nil, fmt.Errorf("IN/NOT IN 操作符的值数组不能为空")
		}

		placeholders := make([]string, len(values))
		for i := range values {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s %s (%s)", column, operator, strings.Join(placeholders, ", ")), values, nil

	default:
		return "", nil, fmt.Errorf("不支持的操作符: %s", operator)
	}
}

// InsertRow 插入新行数据
func (dr *DataRepository) InsertRow(database, table string, data map[string]interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("插入数据不能为空")
	}

	// 构建列名和占位符
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for column, value := range data {
		columns = append(columns, escapeIdentifier(column))
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	// 构建 INSERT 语句
	sqlQuery := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		escapeIdentifier(database),
		escapeIdentifier(table),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// 执行插入
	_, err := dr.db.Exec(sqlQuery, values...)
	if err != nil {
		return fmt.Errorf("执行插入失败: %w", err)
	}

	return nil
}

// UpdateRow 更新行数据（基于主键）
func (dr *DataRepository) UpdateRow(database, table string, pk map[string]interface{}, data map[string]interface{}) error {
	if len(pk) == 0 {
		return fmt.Errorf("主键不能为空")
	}
	if len(data) == 0 {
		return fmt.Errorf("更新数据不能为空")
	}

	// 构建 SET 子句
	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+len(pk))

	for column, value := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", escapeIdentifier(column)))
		values = append(values, value)
	}

	// 构建 WHERE 子句（基于主键）
	whereClauses := make([]string, 0, len(pk))
	for column, value := range pk {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", escapeIdentifier(column)))
		values = append(values, value)
	}

	// 构建 UPDATE 语句
	sqlQuery := fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s",
		escapeIdentifier(database),
		escapeIdentifier(table),
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "))

	// 执行更新
	result, err := dr.db.Exec(sqlQuery, values...)
	if err != nil {
		return fmt.Errorf("执行更新失败: %w", err)
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取受影响行数失败: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("未找到匹配的行进行更新")
	}

	return nil
}

// DeleteRows 删除行数据（基于主键，支持批量删除）
func (dr *DataRepository) DeleteRows(database, table string, pks []map[string]interface{}) (int64, error) {
	if len(pks) == 0 {
		return 0, fmt.Errorf("主键列表不能为空")
	}

	var totalDeleted int64

	// 对每个主键执行删除
	for _, pk := range pks {
		if len(pk) == 0 {
			return totalDeleted, fmt.Errorf("主键不能为空")
		}

		// 构建 WHERE 子句（基于主键）
		whereClauses := make([]string, 0, len(pk))
		values := make([]interface{}, 0, len(pk))

		for column, value := range pk {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", escapeIdentifier(column)))
			values = append(values, value)
		}

		// 构建 DELETE 语句
		sqlQuery := fmt.Sprintf("DELETE FROM %s.%s WHERE %s",
			escapeIdentifier(database),
			escapeIdentifier(table),
			strings.Join(whereClauses, " AND "))

		// 执行删除
		result, err := dr.db.Exec(sqlQuery, values...)
		if err != nil {
			return totalDeleted, fmt.Errorf("执行删除失败: %w", err)
		}

		// 累计删除的行数
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("获取受影响行数失败: %w", err)
		}
		totalDeleted += rowsAffected
	}

	return totalDeleted, nil
}

// escapeIdentifier 转义数据库标识符（表名、列名）
func escapeIdentifier(identifier string) string {
	// 移除可能存在的反引号
	identifier = strings.ReplaceAll(identifier, "`", "")
	// 添加反引号
	return "`" + identifier + "`"
}
