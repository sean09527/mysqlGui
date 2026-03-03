# Import/Export Service

数据导入导出服务模块，支持 SQL、CSV 和 JSON 格式的数据导入导出。

## 功能特性

### 数据导出

- **SQL 格式导出**: 生成 INSERT 语句，可直接在 MySQL 中执行
- **CSV 格式导出**: 生成标准 CSV 文件，包含列标题
- **JSON 格式导出**: 生成 JSON 数组，每行数据为一个对象

### 数据导入

- **SQL 文件导入**: 执行 SQL 文件中的 INSERT/UPDATE/DELETE 语句
- **CSV 文件导入**: 支持列映射，批量插入数据
- **JSON 文件导入**: 支持列映射，批量插入数据

### 核心特性

1. **进度报告**: 所有导入导出操作支持进度回调
2. **批量处理**: 大数据集分批处理，避免内存溢出
3. **错误处理**: 导入失败时记录错误详情，继续处理剩余数据
4. **数据验证**: 导入前验证文件格式
5. **筛选导出**: 支持导出筛选后的数据

## 使用示例

### 导出数据

```go
import (
    "mygui/backend/internal/importexport"
    "mygui/backend/internal/repository"
)

// 创建服务
service := importexport.NewService(db)

// 定义查询条件
query := repository.DataQuery{
    Database: "mydb",
    Table:    "users",
    Filters: []repository.Filter{
        {Column: "age", Operator: ">", Value: 18},
    },
    Limit: 1000,
}

// 导出为 SQL
err := service.ExportToSQL("mydb", "users", query, "output.sql", func(current, total int) {
    fmt.Printf("导出进度: %d/%d\n", current, total)
})

// 导出为 CSV
err = service.ExportToCSV("mydb", "users", query, "output.csv", nil)

// 导出为 JSON
err = service.ExportToJSON("mydb", "users", query, "output.json", nil)
```

### 导入数据

```go
// 导入 SQL 文件
result, err := service.ImportFromSQL("mydb", "data.sql", func(current, total int) {
    fmt.Printf("导入进度: %d/%d\n", current, total)
})

// 导入 CSV 文件（带列映射）
mapping := importexport.ColumnMapping{
    FileColumns:  []string{"user_name", "user_email", "user_age"},
    TableColumns: []string{"name", "email", "age"},
}
result, err = service.ImportFromCSV("mydb", "users", "data.csv", mapping, nil)

// 导入 JSON 文件
result, err = service.ImportFromJSON("mydb", "users", "data.json", mapping, nil)

// 检查导入结果
fmt.Printf("总行数: %d, 成功: %d, 失败: %d\n", 
    result.TotalRows, result.SuccessRows, result.FailedRows)
for _, err := range result.Errors {
    fmt.Printf("第 %d 行错误: %s\n", err.Row, err.Message)
}
```

### 验证文件格式

```go
// 验证 CSV 格式
err := service.ValidateCSVFormat("data.csv")

// 验证 JSON 格式
err = service.ValidateJSONFormat("data.json")

// 验证 SQL 格式
err = service.ValidateSQLFormat("data.sql")
```

## 数据结构

### ImportResult

```go
type ImportResult struct {
    TotalRows   int           // 总行数
    SuccessRows int           // 成功导入的行数
    FailedRows  int           // 失败的行数
    Errors      []ImportError // 错误详情
}
```

### ImportError

```go
type ImportError struct {
    Row     int    // 错误行号
    Message string // 错误消息
}
```

### ColumnMapping

```go
type ColumnMapping struct {
    FileColumns  []string // 文件中的列名
    TableColumns []string // 表中的列名
}
```

## 性能优化

1. **分批处理**: 导出和导入都采用分批处理，默认每批 1000 行（导出）或 100 行（导入）
2. **流式处理**: CSV 和 SQL 文件采用流式读取，避免一次性加载整个文件
3. **批量插入**: 导入时使用批量插入，提高性能

## 错误处理

- 导入过程中遇到错误时，记录错误详情并继续处理剩余数据
- 所有错误都会记录在 ImportResult.Errors 中
- 支持部分成功的导入操作

## 测试

运行单元测试：

```bash
go test ./backend/internal/importexport/... -v
```

注意：完整的集成测试需要 MySQL 数据库连接。可以通过环境变量 `TEST_MYSQL_DSN` 配置测试数据库连接字符串。

## 需求覆盖

### 需求 17 (数据导出)
- ✅ 17.1: 允许用户选择要导出的表
- ✅ 17.2: 支持导出为 SQL INSERT 语句格式
- ✅ 17.3: 支持导出为 CSV 格式
- ✅ 17.4: 支持导出为 JSON 格式
- ✅ 17.5: 允许用户选择导出全部数据或当前筛选的数据
- ✅ 17.6: 生成导出文件并提示用户保存位置
- ✅ 17.7: 显示导出进度
- ✅ 17.8: 显示成功消息和导出的行数

### 需求 18 (数据导入)
- ✅ 18.1: 允许用户选择要导入的文件
- ✅ 18.2: 支持导入 SQL 文件
- ✅ 18.3: 支持导入 CSV 文件
- ✅ 18.4: 支持导入 JSON 文件
- ✅ 18.5: 导入 CSV 或 JSON 时允许用户映射文件字段到表列
- ✅ 18.6: 在导入前验证数据格式
- ✅ 18.7: 执行数据插入操作
- ✅ 18.8: 显示导入进度
- ✅ 18.9: 显示成功消息、导入的行数和失败的行数
- ✅ 18.10: 导入过程中发生错误时记录错误详情并继续处理剩余数据

## 文件结构

```
backend/internal/importexport/
├── export.go          # 导出功能实现
├── import.go          # 导入功能实现
├── service.go         # 服务封装
├── export_test.go     # 导出功能测试
├── import_test.go     # 导入功能测试
├── service_test.go    # 单元测试
└── README.md          # 本文档
```
