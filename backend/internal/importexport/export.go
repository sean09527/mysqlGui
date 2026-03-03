package importexport

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"mygui/backend/internal/repository"
	"os"
	"strings"
)

// Exporter 处理数据导出操作
type Exporter struct {
	db *sql.DB
}

// NewExporter 创建新的 Exporter 实例
func NewExporter(db *sql.DB) *Exporter {
	return &Exporter{db: db}
}

// ExportToSQL 导出数据为 SQL INSERT 语句
func (e *Exporter) ExportToSQL(database, table string, query repository.DataQuery, outputPath string, progressCallback func(current, total int)) error {
	if database == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return fmt.Errorf("表名称不能为空")
	}
	if outputPath == "" {
		return fmt.Errorf("输出路径不能为空")
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer file.Close()

	// 设置查询参数以获取所有数据
	query.Database = database
	query.Table = table
	if query.Limit == 0 {
		query.Limit = 1000 // 分批处理，每批 1000 行
	}

	// 获取数据仓库
	repo := repository.NewDataRepository(e.db)

	// 获取总行数
	totalRows, err := repo.GetRowCount(database, table, query.Filters)
	if err != nil {
		return fmt.Errorf("获取总行数失败: %w", err)
	}

	if totalRows == 0 {
		return fmt.Errorf("没有数据可导出")
	}

	// 写入文件头注释
	_, err = file.WriteString(fmt.Sprintf("-- MySQL dump for table %s.%s\n", database, table))
	if err != nil {
		return fmt.Errorf("写入文件头失败: %w", err)
	}
	_, err = file.WriteString(fmt.Sprintf("-- Total rows: %d\n\n", totalRows))
	if err != nil {
		return fmt.Errorf("写入文件头失败: %w", err)
	}

	// 分批导出数据
	processedRows := 0
	for offset := 0; offset < int(totalRows); offset += query.Limit {
		query.Offset = offset

		// 查询数据
		result, err := repo.QueryData(query)
		if err != nil {
			return fmt.Errorf("查询数据失败: %w", err)
		}

		// 生成 INSERT 语句
		for _, row := range result.Rows {
			insertSQL := e.generateInsertStatement(database, table, result.Columns, row)
			_, err = file.WriteString(insertSQL + "\n")
			if err != nil {
				return fmt.Errorf("写入 INSERT 语句失败: %w", err)
			}

			processedRows++
			if progressCallback != nil {
				progressCallback(processedRows, int(totalRows))
			}
		}
	}

	return nil
}

// ExportToCSV 导出数据为 CSV 格式
func (e *Exporter) ExportToCSV(database, table string, query repository.DataQuery, outputPath string, progressCallback func(current, total int)) error {
	if database == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return fmt.Errorf("表名称不能为空")
	}
	if outputPath == "" {
		return fmt.Errorf("输出路径不能为空")
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer file.Close()

	// 创建 CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 设置查询参数
	query.Database = database
	query.Table = table
	if query.Limit == 0 {
		query.Limit = 1000 // 分批处理
	}

	// 获取数据仓库
	repo := repository.NewDataRepository(e.db)

	// 获取总行数
	totalRows, err := repo.GetRowCount(database, table, query.Filters)
	if err != nil {
		return fmt.Errorf("获取总行数失败: %w", err)
	}

	if totalRows == 0 {
		return fmt.Errorf("没有数据可导出")
	}

	// 写入列标题
	firstBatch := true
	processedRows := 0

	// 分批导出数据
	for offset := 0; offset < int(totalRows); offset += query.Limit {
		query.Offset = offset

		// 查询数据
		result, err := repo.QueryData(query)
		if err != nil {
			return fmt.Errorf("查询数据失败: %w", err)
		}

		// 写入列标题（仅第一批）
		if firstBatch {
			if err := writer.Write(result.Columns); err != nil {
				return fmt.Errorf("写入列标题失败: %w", err)
			}
			firstBatch = false
		}

		// 写入数据行
		for _, row := range result.Rows {
			record := make([]string, len(row))
			for i, val := range row {
				record[i] = e.formatValue(val)
			}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("写入数据行失败: %w", err)
			}

			processedRows++
			if progressCallback != nil {
				progressCallback(processedRows, int(totalRows))
			}
		}
	}

	return nil
}

// ExportToJSON 导出数据为 JSON 格式
func (e *Exporter) ExportToJSON(database, table string, query repository.DataQuery, outputPath string, progressCallback func(current, total int)) error {
	if database == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return fmt.Errorf("表名称不能为空")
	}
	if outputPath == "" {
		return fmt.Errorf("输出路径不能为空")
	}

	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer file.Close()

	// 设置查询参数
	query.Database = database
	query.Table = table
	if query.Limit == 0 {
		query.Limit = 1000 // 分批处理
	}

	// 获取数据仓库
	repo := repository.NewDataRepository(e.db)

	// 获取总行数
	totalRows, err := repo.GetRowCount(database, table, query.Filters)
	if err != nil {
		return fmt.Errorf("获取总行数失败: %w", err)
	}

	if totalRows == 0 {
		return fmt.Errorf("没有数据可导出")
	}

	// 写入 JSON 数组开始
	_, err = file.WriteString("[\n")
	if err != nil {
		return fmt.Errorf("写入 JSON 开始标记失败: %w", err)
	}

	// 分批导出数据
	processedRows := 0
	firstRow := true

	for offset := 0; offset < int(totalRows); offset += query.Limit {
		query.Offset = offset

		// 查询数据
		result, err := repo.QueryData(query)
		if err != nil {
			return fmt.Errorf("查询数据失败: %w", err)
		}

		// 转换为 JSON 对象
		for _, row := range result.Rows {
			// 构建 JSON 对象
			obj := make(map[string]interface{})
			for i, col := range result.Columns {
				obj[col] = row[i]
			}

			// 序列化为 JSON
			jsonData, err := json.Marshal(obj)
			if err != nil {
				return fmt.Errorf("序列化 JSON 失败: %w", err)
			}

			// 写入 JSON（添加逗号分隔符）
			if !firstRow {
				_, err = file.WriteString(",\n")
				if err != nil {
					return fmt.Errorf("写入 JSON 分隔符失败: %w", err)
				}
			}
			_, err = file.WriteString("  " + string(jsonData))
			if err != nil {
				return fmt.Errorf("写入 JSON 数据失败: %w", err)
			}

			firstRow = false
			processedRows++
			if progressCallback != nil {
				progressCallback(processedRows, int(totalRows))
			}
		}
	}

	// 写入 JSON 数组结束
	_, err = file.WriteString("\n]\n")
	if err != nil {
		return fmt.Errorf("写入 JSON 结束标记失败: %w", err)
	}

	return nil
}

// generateInsertStatement 生成 INSERT 语句
func (e *Exporter) generateInsertStatement(database, table string, columns []string, values []interface{}) string {
	// 转义列名
	escapedColumns := make([]string, len(columns))
	for i, col := range columns {
		escapedColumns[i] = escapeIdentifier(col)
	}

	// 格式化值
	formattedValues := make([]string, len(values))
	for i, val := range values {
		formattedValues[i] = e.formatSQLValue(val)
	}

	return fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s);",
		escapeIdentifier(database),
		escapeIdentifier(table),
		strings.Join(escapedColumns, ", "),
		strings.Join(formattedValues, ", "))
}

// formatSQLValue 格式化 SQL 值
func (e *Exporter) formatSQLValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case string:
		// 转义单引号
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case []byte:
		// 字节数组转为字符串
		escaped := strings.ReplaceAll(string(v), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		// 其他类型转为字符串
		return fmt.Sprintf("'%v'", v)
	}
}

// formatValue 格式化值为字符串（用于 CSV）
func (e *Exporter) formatValue(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// escapeIdentifier 转义数据库标识符
func escapeIdentifier(identifier string) string {
	identifier = strings.ReplaceAll(identifier, "`", "")
	return "`" + identifier + "`"
}
