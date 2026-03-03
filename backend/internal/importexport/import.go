package importexport

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Importer 处理数据导入操作
type Importer struct {
	db *sql.DB
}

// NewImporter 创建新的 Importer 实例
func NewImporter(db *sql.DB) *Importer {
	return &Importer{db: db}
}

// ImportResult 表示导入结果
type ImportResult struct {
	TotalRows   int
	SuccessRows int
	FailedRows  int
	Errors      []ImportError
}

// ImportError 表示导入错误
type ImportError struct {
	Row     int
	Message string
}

// ColumnMapping 表示列映射关系
type ColumnMapping struct {
	FileColumns  []string
	TableColumns []string
}

// ImportFromSQL 从 SQL 文件导入数据
func (i *Importer) ImportFromSQL(database string, sqlFilePath string, progressCallback func(current, total int)) (*ImportResult, error) {
	if database == "" {
		return nil, fmt.Errorf("数据库名称不能为空")
	}
	if sqlFilePath == "" {
		return nil, fmt.Errorf("SQL 文件路径不能为空")
	}

	// 打开 SQL 文件
	file, err := os.Open(sqlFilePath)
	if err != nil {
		return nil, fmt.Errorf("打开 SQL 文件失败: %w", err)
	}
	defer file.Close()

	// 切换到目标数据库
	_, err = i.db.Exec(fmt.Sprintf("USE %s", escapeIdentifier(database)))
	if err != nil {
		return nil, fmt.Errorf("切换数据库失败: %w", err)
	}

	result := &ImportResult{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 增加缓冲区大小以处理大语句

	var currentStatement strings.Builder
	lineNumber := 0

	// 逐行读取并执行 SQL 语句
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}

		// 累积 SQL 语句（直到遇到分号）
		currentStatement.WriteString(line)
		currentStatement.WriteString(" ")

		// 检查是否是完整的语句（以分号结尾）
		if strings.HasSuffix(line, ";") {
			stmt := strings.TrimSpace(currentStatement.String())
			result.TotalRows++

			// 执行 SQL 语句
			_, err := i.db.Exec(stmt)
			if err != nil {
				result.FailedRows++
				result.Errors = append(result.Errors, ImportError{
					Row:     lineNumber,
					Message: fmt.Sprintf("执行 SQL 失败: %v", err),
				})
			} else {
				result.SuccessRows++
			}

			// 重置语句构建器
			currentStatement.Reset()

			// 报告进度
			if progressCallback != nil {
				progressCallback(result.TotalRows, result.TotalRows)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("读取 SQL 文件失败: %w", err)
	}

	return result, nil
}

// ImportFromCSV 从 CSV 文件导入数据
func (i *Importer) ImportFromCSV(database, table string, csvFilePath string, mapping ColumnMapping, progressCallback func(current, total int)) (*ImportResult, error) {
	if database == "" {
		return nil, fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return nil, fmt.Errorf("表名称不能为空")
	}
	if csvFilePath == "" {
		return nil, fmt.Errorf("CSV 文件路径不能为空")
	}

	// 打开 CSV 文件
	file, err := os.Open(csvFilePath)
	if err != nil {
		return nil, fmt.Errorf("打开 CSV 文件失败: %w", err)
	}
	defer file.Close()

	// 创建 CSV reader
	reader := csv.NewReader(file)

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("读取 CSV 表头失败: %w", err)
	}

	// 如果没有提供映射，使用表头作为列名
	if len(mapping.FileColumns) == 0 {
		mapping.FileColumns = headers
		mapping.TableColumns = headers
	}

	// 验证映射
	if len(mapping.FileColumns) != len(mapping.TableColumns) {
		return nil, fmt.Errorf("列映射不匹配: 文件列数 %d, 表列数 %d", len(mapping.FileColumns), len(mapping.TableColumns))
	}

	// 创建列索引映射
	columnIndexMap := make(map[string]int)
	for i, header := range headers {
		columnIndexMap[header] = i
	}

	result := &ImportResult{}
	rowNumber := 1 // 从 1 开始（表头是第 0 行）

	// 批量插入配置
	batchSize := 100
	batch := make([]map[string]interface{}, 0, batchSize)

	// 逐行读取数据
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.FailedRows++
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNumber,
				Message: fmt.Sprintf("读取 CSV 行失败: %v", err),
			})
			rowNumber++
			continue
		}

		result.TotalRows++

		// 构建数据映射
		data := make(map[string]interface{})
		valid := true

		for i, fileCol := range mapping.FileColumns {
			tableCol := mapping.TableColumns[i]
			colIndex, exists := columnIndexMap[fileCol]

			if !exists {
				result.FailedRows++
				result.Errors = append(result.Errors, ImportError{
					Row:     rowNumber,
					Message: fmt.Sprintf("CSV 文件中未找到列: %s", fileCol),
				})
				valid = false
				break
			}

			if colIndex >= len(record) {
				result.FailedRows++
				result.Errors = append(result.Errors, ImportError{
					Row:     rowNumber,
					Message: fmt.Sprintf("CSV 行数据不完整，缺少列: %s", fileCol),
				})
				valid = false
				break
			}

			// 处理空值
			value := record[colIndex]
			if value == "" {
				data[tableCol] = nil
			} else {
				data[tableCol] = value
			}
		}

		if !valid {
			rowNumber++
			continue
		}

		// 添加到批次
		batch = append(batch, data)

		// 当批次满时执行插入
		if len(batch) >= batchSize {
			successCount, errors := i.executeBatchInsert(database, table, batch)
			result.SuccessRows += successCount
			result.FailedRows += len(errors)
			for _, err := range errors {
				result.Errors = append(result.Errors, ImportError{
					Row:     rowNumber - len(batch) + err.Row,
					Message: err.Message,
				})
			}
			batch = batch[:0] // 清空批次

			// 报告进度
			if progressCallback != nil {
				progressCallback(result.TotalRows, result.TotalRows)
			}
		}

		rowNumber++
	}

	// 处理剩余的批次
	if len(batch) > 0 {
		successCount, errors := i.executeBatchInsert(database, table, batch)
		result.SuccessRows += successCount
		result.FailedRows += len(errors)
		for _, err := range errors {
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNumber - len(batch) + err.Row,
				Message: err.Message,
			})
		}
	}

	// 最终进度报告
	if progressCallback != nil {
		progressCallback(result.TotalRows, result.TotalRows)
	}

	return result, nil
}

// ImportFromJSON 从 JSON 文件导入数据
func (i *Importer) ImportFromJSON(database, table string, jsonFilePath string, mapping ColumnMapping, progressCallback func(current, total int)) (*ImportResult, error) {
	if database == "" {
		return nil, fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return nil, fmt.Errorf("表名称不能为空")
	}
	if jsonFilePath == "" {
		return nil, fmt.Errorf("JSON 文件路径不能为空")
	}

	// 打开 JSON 文件
	file, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("打开 JSON 文件失败: %w", err)
	}
	defer file.Close()

	// 解析 JSON 数组
	var jsonData []map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&jsonData); err != nil {
		return nil, fmt.Errorf("解析 JSON 文件失败: %w", err)
	}

	if len(jsonData) == 0 {
		return nil, fmt.Errorf("JSON 文件中没有数据")
	}

	// 如果没有提供映射，使用第一条记录的键作为列名
	if len(mapping.FileColumns) == 0 {
		for key := range jsonData[0] {
			mapping.FileColumns = append(mapping.FileColumns, key)
			mapping.TableColumns = append(mapping.TableColumns, key)
		}
	}

	// 验证映射
	if len(mapping.FileColumns) != len(mapping.TableColumns) {
		return nil, fmt.Errorf("列映射不匹配: 文件列数 %d, 表列数 %d", len(mapping.FileColumns), len(mapping.TableColumns))
	}

	result := &ImportResult{
		TotalRows: len(jsonData),
	}

	// 批量插入配置
	batchSize := 100
	batch := make([]map[string]interface{}, 0, batchSize)

	// 逐条处理数据
	for rowIndex, record := range jsonData {
		// 构建数据映射
		data := make(map[string]interface{})
		valid := true

		for i, fileCol := range mapping.FileColumns {
			tableCol := mapping.TableColumns[i]
			value, exists := record[fileCol]

			if !exists {
				result.FailedRows++
				result.Errors = append(result.Errors, ImportError{
					Row:     rowIndex + 1,
					Message: fmt.Sprintf("JSON 记录中未找到字段: %s", fileCol),
				})
				valid = false
				break
			}

			data[tableCol] = value
		}

		if !valid {
			continue
		}

		// 添加到批次
		batch = append(batch, data)

		// 当批次满时执行插入
		if len(batch) >= batchSize {
			successCount, errors := i.executeBatchInsert(database, table, batch)
			result.SuccessRows += successCount
			result.FailedRows += len(errors)
			for _, err := range errors {
				result.Errors = append(result.Errors, ImportError{
					Row:     rowIndex - len(batch) + err.Row + 1,
					Message: err.Message,
				})
			}
			batch = batch[:0] // 清空批次

			// 报告进度
			if progressCallback != nil {
				progressCallback(rowIndex + 1, len(jsonData))
			}
		}
	}

	// 处理剩余的批次
	if len(batch) > 0 {
		successCount, errors := i.executeBatchInsert(database, table, batch)
		result.SuccessRows += successCount
		result.FailedRows += len(errors)
		for _, err := range errors {
			result.Errors = append(result.Errors, ImportError{
				Row:     len(jsonData) - len(batch) + err.Row + 1,
				Message: err.Message,
			})
		}
	}

	// 最终进度报告
	if progressCallback != nil {
		progressCallback(len(jsonData), len(jsonData))
	}

	return result, nil
}

// executeBatchInsert 执行批量插入
func (i *Importer) executeBatchInsert(database, table string, batch []map[string]interface{}) (int, []ImportError) {
	if len(batch) == 0 {
		return 0, nil
	}

	successCount := 0
	var errors []ImportError

	// 对每条记录执行插入（如果批量插入失败，则逐条插入以记录具体错误）
	for idx, data := range batch {
		err := i.insertRow(database, table, data)
		if err != nil {
			errors = append(errors, ImportError{
				Row:     idx,
				Message: fmt.Sprintf("插入数据失败: %v", err),
			})
		} else {
			successCount++
		}
	}

	return successCount, errors
}

// insertRow 插入单行数据
func (i *Importer) insertRow(database, table string, data map[string]interface{}) error {
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
	_, err := i.db.Exec(sqlQuery, values...)
	if err != nil {
		return err
	}

	return nil
}

// ValidateCSVFormat 验证 CSV 文件格式
func (i *Importer) ValidateCSVFormat(csvFilePath string) error {
	file, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("打开 CSV 文件失败: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("读取 CSV 表头失败: %w", err)
	}

	if len(headers) == 0 {
		return fmt.Errorf("CSV 文件表头为空")
	}

	// 读取第一行数据验证格式
	_, err = reader.Read()
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取 CSV 数据失败: %w", err)
	}

	return nil
}

// ValidateJSONFormat 验证 JSON 文件格式
func (i *Importer) ValidateJSONFormat(jsonFilePath string) error {
	file, err := os.Open(jsonFilePath)
	if err != nil {
		return fmt.Errorf("打开 JSON 文件失败: %w", err)
	}
	defer file.Close()

	// 尝试解析 JSON
	var jsonData []map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&jsonData); err != nil {
		return fmt.Errorf("JSON 格式无效: %w", err)
	}

	if len(jsonData) == 0 {
		return fmt.Errorf("JSON 文件中没有数据")
	}

	return nil
}

// ValidateSQLFormat 验证 SQL 文件格式
func (i *Importer) ValidateSQLFormat(sqlFilePath string) error {
	file, err := os.Open(sqlFilePath)
	if err != nil {
		return fmt.Errorf("打开 SQL 文件失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	hasValidSQL := false

	// 简单的 SQL 语句正则表达式
	sqlPattern := regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|CREATE|ALTER|DROP|SELECT)\s+`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}

		if sqlPattern.MatchString(line) {
			hasValidSQL = true
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 SQL 文件失败: %w", err)
	}

	if !hasValidSQL {
		return fmt.Errorf("SQL 文件中没有有效的 SQL 语句")
	}

	return nil
}
