package importexport

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"mygui/backend/internal/repository"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func setupTestDB(t *testing.T) *sql.DB {
	// 使用环境变量或默认值连接测试数据库
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}

	// 创建测试数据库
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS test_export_db")
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 创建测试表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test_export_db.test_users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100),
			age INT,
			active BOOLEAN DEFAULT TRUE
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 清空测试表
	_, err = db.Exec("TRUNCATE TABLE test_export_db.test_users")
	if err != nil {
		t.Fatalf("清空测试表失败: %v", err)
	}

	// 插入测试数据
	testData := []struct {
		name   string
		email  string
		age    int
		active bool
	}{
		{"Alice", "alice@example.com", 25, true},
		{"Bob", "bob@example.com", 30, true},
		{"Charlie", "charlie@example.com", 35, false},
	}

	for _, data := range testData {
		_, err = db.Exec(
			"INSERT INTO test_export_db.test_users (name, email, age, active) VALUES (?, ?, ?, ?)",
			data.name, data.email, data.age, data.active,
		)
		if err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DROP DATABASE IF EXISTS test_export_db")
	if err != nil {
		t.Errorf("清理测试数据库失败: %v", err)
	}
	db.Close()
}

func TestExportToSQL(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	exporter := NewExporter(db)
	outputPath := "test_export.sql"
	defer os.Remove(outputPath)

	query := repository.DataQuery{
		Database: "test_export_db",
		Table:    "test_users",
		Limit:    100,
	}

	progressCalled := false
	progressCallback := func(current, total int) {
		progressCalled = true
		if current > total {
			t.Errorf("进度回调错误: current (%d) > total (%d)", current, total)
		}
	}

	err := exporter.ExportToSQL("test_export_db", "test_users", query, outputPath, progressCallback)
	if err != nil {
		t.Fatalf("导出 SQL 失败: %v", err)
	}

	if !progressCalled {
		t.Error("进度回调未被调用")
	}

	// 验证输出文件
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("读取输出文件失败: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "INSERT INTO") {
		t.Error("输出文件不包含 INSERT 语句")
	}
	if !strings.Contains(contentStr, "Alice") {
		t.Error("输出文件不包含测试数据")
	}
}

func TestExportToCSV(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	exporter := NewExporter(db)
	outputPath := "test_export.csv"
	defer os.Remove(outputPath)

	query := repository.DataQuery{
		Database: "test_export_db",
		Table:    "test_users",
		Limit:    100,
	}

	err := exporter.ExportToCSV("test_export_db", "test_users", query, outputPath, nil)
	if err != nil {
		t.Fatalf("导出 CSV 失败: %v", err)
	}

	// 验证 CSV 文件
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("打开 CSV 文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("读取 CSV 文件失败: %v", err)
	}

	// 验证表头和数据行数（表头 + 3 行数据）
	if len(records) != 4 {
		t.Errorf("期望 4 行（表头 + 3 行数据），实际 %d 行", len(records))
	}

	// 验证表头
	headers := records[0]
	expectedHeaders := []string{"id", "name", "email", "age", "active"}
	for _, expected := range expectedHeaders {
		found := false
		for _, header := range headers {
			if header == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("表头缺少列: %s", expected)
		}
	}
}

func TestExportToJSON(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	exporter := NewExporter(db)
	outputPath := "test_export.json"
	defer os.Remove(outputPath)

	query := repository.DataQuery{
		Database: "test_export_db",
		Table:    "test_users",
		Limit:    100,
	}

	err := exporter.ExportToJSON("test_export_db", "test_users", query, outputPath, nil)
	if err != nil {
		t.Fatalf("导出 JSON 失败: %v", err)
	}

	// 验证 JSON 文件
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("读取 JSON 文件失败: %v", err)
	}

	var jsonData []map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	if err != nil {
		t.Fatalf("解析 JSON 失败: %v", err)
	}

	// 验证数据行数
	if len(jsonData) != 3 {
		t.Errorf("期望 3 行数据，实际 %d 行", len(jsonData))
	}

	// 验证第一条记录
	if len(jsonData) > 0 {
		firstRecord := jsonData[0]
		if firstRecord["name"] != "Alice" {
			t.Errorf("期望第一条记录的 name 为 Alice，实际为 %v", firstRecord["name"])
		}
	}
}

func TestExportWithFilters(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	exporter := NewExporter(db)
	outputPath := "test_export_filtered.json"
	defer os.Remove(outputPath)

	query := repository.DataQuery{
		Database: "test_export_db",
		Table:    "test_users",
		Filters: []repository.Filter{
			{Column: "age", Operator: ">", Value: 25},
		},
		Limit: 100,
	}

	err := exporter.ExportToJSON("test_export_db", "test_users", query, outputPath, nil)
	if err != nil {
		t.Fatalf("导出筛选数据失败: %v", err)
	}

	// 验证筛选结果
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("读取 JSON 文件失败: %v", err)
	}

	var jsonData []map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	if err != nil {
		t.Fatalf("解析 JSON 失败: %v", err)
	}

	// 应该只有 2 条记录（age > 25）
	if len(jsonData) != 2 {
		t.Errorf("期望 2 行筛选数据，实际 %d 行", len(jsonData))
	}
}

func TestExportEmptyTable(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// 清空表
	_, err := db.Exec("TRUNCATE TABLE test_export_db.test_users")
	if err != nil {
		t.Fatalf("清空表失败: %v", err)
	}

	exporter := NewExporter(db)
	outputPath := "test_export_empty.json"
	defer os.Remove(outputPath)

	query := repository.DataQuery{
		Database: "test_export_db",
		Table:    "test_users",
		Limit:    100,
	}

	err = exporter.ExportToJSON("test_export_db", "test_users", query, outputPath, nil)
	if err == nil {
		t.Error("期望导出空表时返回错误")
	}
	if !strings.Contains(err.Error(), "没有数据可导出") {
		t.Errorf("期望错误消息包含'没有数据可导出'，实际: %v", err)
	}
}

func TestExportInvalidParameters(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	exporter := NewExporter(db)

	tests := []struct {
		name       string
		database   string
		table      string
		outputPath string
		wantErr    bool
	}{
		{"空数据库名", "", "test_users", "output.json", true},
		{"空表名", "test_export_db", "", "output.json", true},
		{"空输出路径", "test_export_db", "test_users", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := repository.DataQuery{}
			err := exporter.ExportToJSON(tt.database, tt.table, query, tt.outputPath, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("期望错误 = %v, 实际错误 = %v", tt.wantErr, err)
			}
		})
	}
}
