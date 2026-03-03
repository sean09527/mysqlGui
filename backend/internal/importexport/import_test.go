package importexport

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func setupImportTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}

	// 创建测试数据库
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS test_import_db")
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 创建测试表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test_import_db.test_products (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			price DECIMAL(10, 2),
			stock INT DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 清空测试表
	_, err = db.Exec("TRUNCATE TABLE test_import_db.test_products")
	if err != nil {
		t.Fatalf("清空测试表失败: %v", err)
	}

	return db
}

func cleanupImportTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DROP DATABASE IF EXISTS test_import_db")
	if err != nil {
		t.Errorf("清理测试数据库失败: %v", err)
	}
	db.Close()
}

func TestImportFromSQL(t *testing.T) {
	db := setupImportTestDB(t)
	defer cleanupImportTestDB(t, db)

	// 创建测试 SQL 文件
	sqlContent := `
-- Test SQL import
INSERT INTO test_products (name, price, stock) VALUES ('Product A', 10.50, 100);
INSERT INTO test_products (name, price, stock) VALUES ('Product B', 20.00, 50);
INSERT INTO test_products (name, price, stock) VALUES ('Product C', 15.75, 75);
`
	sqlFilePath := "test_import.sql"
	err := os.WriteFile(sqlFilePath, []byte(sqlContent), 0644)
	if err != nil {
		t.Fatalf("创建测试 SQL 文件失败: %v", err)
	}
	defer os.Remove(sqlFilePath)

	importer := NewImporter(db)

	progressCalled := false
	progressCallback := func(current, total int) {
		progressCalled = true
	}

	result, err := importer.ImportFromSQL("test_import_db", sqlFilePath, progressCallback)
	if err != nil {
		t.Fatalf("导入 SQL 失败: %v", err)
	}

	if !progressCalled {
		t.Error("进度回调未被调用")
	}

	if result.TotalRows != 3 {
		t.Errorf("期望导入 3 行，实际 %d 行", result.TotalRows)
	}

	if result.SuccessRows != 3 {
		t.Errorf("期望成功 3 行，实际 %d 行", result.SuccessRows)
	}

	if result.FailedRows != 0 {
		t.Errorf("期望失败 0 行，实际 %d 行", result.FailedRows)
	}

	// 验证数据是否插入
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_import_db.test_products").Scan(&count)
	if err != nil {
		t.Fatalf("查询数据失败: %v", err)
	}

	if count != 3 {
		t.Errorf("期望表中有 3 行数据，实际 %d 行", count)
	}
}

func TestImportFromCSV(t *testing.T) {
	db := setupImportTestDB(t)
	defer cleanupImportTestDB(t, db)

	// 创建测试 CSV 文件
	csvFilePath := "test_import.csv"
	file, err := os.Create(csvFilePath)
	if err != nil {
		t.Fatalf("创建测试 CSV 文件失败: %v", err)
	}

	writer := csv.NewWriter(file)
	writer.Write([]string{"name", "price", "stock"})
	writer.Write([]string{"Product A", "10.50", "100"})
	writer.Write([]string{"Product B", "20.00", "50"})
	writer.Write([]string{"Product C", "15.75", "75"})
	writer.Flush()
	file.Close()
	defer os.Remove(csvFilePath)

	importer := NewImporter(db)

	mapping := ColumnMapping{
		FileColumns:  []string{"name", "price", "stock"},
		TableColumns: []string{"name", "price", "stock"},
	}

	result, err := importer.ImportFromCSV("test_import_db", "test_products", csvFilePath, mapping, nil)
	if err != nil {
		t.Fatalf("导入 CSV 失败: %v", err)
	}

	if result.TotalRows != 3 {
		t.Errorf("期望导入 3 行，实际 %d 行", result.TotalRows)
	}

	if result.SuccessRows != 3 {
		t.Errorf("期望成功 3 行，实际 %d 行", result.SuccessRows)
	}

	// 验证数据
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_import_db.test_products").Scan(&count)
	if err != nil {
		t.Fatalf("查询数据失败: %v", err)
	}

	if count != 3 {
		t.Errorf("期望表中有 3 行数据，实际 %d 行", count)
	}
}

func TestImportFromJSON(t *testing.T) {
	db := setupImportTestDB(t)
	defer cleanupImportTestDB(t, db)

	// 创建测试 JSON 文件
	jsonData := []map[string]interface{}{
		{"name": "Product A", "price": 10.50, "stock": 100},
		{"name": "Product B", "price": 20.00, "stock": 50},
		{"name": "Product C", "price": 15.75, "stock": 75},
	}

	jsonFilePath := "test_import.json"
	file, err := os.Create(jsonFilePath)
	if err != nil {
		t.Fatalf("创建测试 JSON 文件失败: %v", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(jsonData)
	if err != nil {
		t.Fatalf("写入 JSON 数据失败: %v", err)
	}
	file.Close()
	defer os.Remove(jsonFilePath)

	importer := NewImporter(db)

	mapping := ColumnMapping{
		FileColumns:  []string{"name", "price", "stock"},
		TableColumns: []string{"name", "price", "stock"},
	}

	result, err := importer.ImportFromJSON("test_import_db", "test_products", jsonFilePath, mapping, nil)
	if err != nil {
		t.Fatalf("导入 JSON 失败: %v", err)
	}

	if result.TotalRows != 3 {
		t.Errorf("期望导入 3 行，实际 %d 行", result.TotalRows)
	}

	if result.SuccessRows != 3 {
		t.Errorf("期望成功 3 行，实际 %d 行", result.SuccessRows)
	}

	// 验证数据
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_import_db.test_products").Scan(&count)
	if err != nil {
		t.Fatalf("查询数据失败: %v", err)
	}

	if count != 3 {
		t.Errorf("期望表中有 3 行数据，实际 %d 行", count)
	}
}

func TestImportWithColumnMapping(t *testing.T) {
	db := setupImportTestDB(t)
	defer cleanupImportTestDB(t, db)

	// 创建 CSV 文件，列名与表列名不同
	csvFilePath := "test_import_mapping.csv"
	file, err := os.Create(csvFilePath)
	if err != nil {
		t.Fatalf("创建测试 CSV 文件失败: %v", err)
	}

	writer := csv.NewWriter(file)
	writer.Write([]string{"product_name", "product_price", "product_stock"})
	writer.Write([]string{"Product A", "10.50", "100"})
	writer.Flush()
	file.Close()
	defer os.Remove(csvFilePath)

	importer := NewImporter(db)

	// 映射文件列到表列
	mapping := ColumnMapping{
		FileColumns:  []string{"product_name", "product_price", "product_stock"},
		TableColumns: []string{"name", "price", "stock"},
	}

	result, err := importer.ImportFromCSV("test_import_db", "test_products", csvFilePath, mapping, nil)
	if err != nil {
		t.Fatalf("导入 CSV 失败: %v", err)
	}

	if result.SuccessRows != 1 {
		t.Errorf("期望成功 1 行，实际 %d 行", result.SuccessRows)
	}

	// 验证数据
	var name string
	err = db.QueryRow("SELECT name FROM test_import_db.test_products LIMIT 1").Scan(&name)
	if err != nil {
		t.Fatalf("查询数据失败: %v", err)
	}

	if name != "Product A" {
		t.Errorf("期望产品名为 'Product A'，实际 '%s'", name)
	}
}

func TestImportWithErrors(t *testing.T) {
	db := setupImportTestDB(t)
	defer cleanupImportTestDB(t, db)

	// 创建包含错误数据的 SQL 文件
	sqlContent := `
INSERT INTO test_products (name, price, stock) VALUES ('Product A', 10.50, 100);
INSERT INTO test_products (id, name, price, stock) VALUES (1, 'Product B', 20.00, 50); -- 重复主键
INSERT INTO test_products (name, price, stock) VALUES ('Product C', 15.75, 75);
`
	sqlFilePath := "test_import_errors.sql"
	err := os.WriteFile(sqlFilePath, []byte(sqlContent), 0644)
	if err != nil {
		t.Fatalf("创建测试 SQL 文件失败: %v", err)
	}
	defer os.Remove(sqlFilePath)

	importer := NewImporter(db)

	result, err := importer.ImportFromSQL("test_import_db", sqlFilePath, nil)
	if err != nil {
		t.Fatalf("导入 SQL 失败: %v", err)
	}

	if result.FailedRows == 0 {
		t.Error("期望有失败的行")
	}

	if len(result.Errors) == 0 {
		t.Error("期望有错误记录")
	}

	// 验证成功的数据仍然被插入
	if result.SuccessRows == 0 {
		t.Error("期望有成功的行")
	}
}

func TestValidateCSVFormat(t *testing.T) {
	// 创建有效的 CSV 文件
	validCSV := "test_valid.csv"
	file, _ := os.Create(validCSV)
	writer := csv.NewWriter(file)
	writer.Write([]string{"col1", "col2"})
	writer.Write([]string{"val1", "val2"})
	writer.Flush()
	file.Close()
	defer os.Remove(validCSV)

	db, _ := sql.Open("mysql", "root:password@tcp(localhost:3306)/")
	defer db.Close()

	importer := NewImporter(db)

	err := importer.ValidateCSVFormat(validCSV)
	if err != nil {
		t.Errorf("验证有效 CSV 失败: %v", err)
	}

	// 测试无效文件
	err = importer.ValidateCSVFormat("nonexistent.csv")
	if err == nil {
		t.Error("期望验证不存在的文件时返回错误")
	}
}

func TestValidateJSONFormat(t *testing.T) {
	// 创建有效的 JSON 文件
	validJSON := "test_valid.json"
	jsonData := []map[string]interface{}{
		{"col1": "val1", "col2": "val2"},
	}
	file, _ := os.Create(validJSON)
	json.NewEncoder(file).Encode(jsonData)
	file.Close()
	defer os.Remove(validJSON)

	db, _ := sql.Open("mysql", "root:password@tcp(localhost:3306)/")
	defer db.Close()

	importer := NewImporter(db)

	err := importer.ValidateJSONFormat(validJSON)
	if err != nil {
		t.Errorf("验证有效 JSON 失败: %v", err)
	}

	// 创建无效的 JSON 文件
	invalidJSON := "test_invalid.json"
	os.WriteFile(invalidJSON, []byte("{invalid json}"), 0644)
	defer os.Remove(invalidJSON)

	err = importer.ValidateJSONFormat(invalidJSON)
	if err == nil {
		t.Error("期望验证无效 JSON 时返回错误")
	}
}

func TestValidateSQLFormat(t *testing.T) {
	// 创建有效的 SQL 文件
	validSQL := "test_valid.sql"
	os.WriteFile(validSQL, []byte("INSERT INTO test VALUES (1, 'test');"), 0644)
	defer os.Remove(validSQL)

	db, _ := sql.Open("mysql", "root:password@tcp(localhost:3306)/")
	defer db.Close()

	importer := NewImporter(db)

	err := importer.ValidateSQLFormat(validSQL)
	if err != nil {
		t.Errorf("验证有效 SQL 失败: %v", err)
	}

	// 创建无效的 SQL 文件（只有注释）
	invalidSQL := "test_invalid.sql"
	os.WriteFile(invalidSQL, []byte("-- Only comments\n-- No SQL"), 0644)
	defer os.Remove(invalidSQL)

	err = importer.ValidateSQLFormat(invalidSQL)
	if err == nil {
		t.Error("期望验证无效 SQL 时返回错误")
	}
}

func TestImportInvalidParameters(t *testing.T) {
	db, _ := sql.Open("mysql", "root:password@tcp(localhost:3306)/")
	defer db.Close()

	importer := NewImporter(db)

	tests := []struct {
		name     string
		database string
		table    string
		filePath string
		wantErr  bool
	}{
		{"空数据库名", "", "test_table", "file.csv", true},
		{"空表名", "test_db", "", "file.csv", true},
		{"空文件路径", "test_db", "test_table", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := ColumnMapping{}
			_, err := importer.ImportFromCSV(tt.database, tt.table, tt.filePath, mapping, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("期望错误 = %v, 实际错误 = %v", tt.wantErr, err)
			}
		})
	}
}

func TestImportEmptyJSON(t *testing.T) {
	db, _ := sql.Open("mysql", "root:password@tcp(localhost:3306)/")
	defer db.Close()

	// 创建空 JSON 数组文件
	emptyJSON := "test_empty.json"
	os.WriteFile(emptyJSON, []byte("[]"), 0644)
	defer os.Remove(emptyJSON)

	importer := NewImporter(db)
	mapping := ColumnMapping{}

	_, err := importer.ImportFromJSON("test_db", "test_table", emptyJSON, mapping, nil)
	if err == nil {
		t.Error("期望导入空 JSON 时返回错误")
	}
	if !strings.Contains(err.Error(), "没有数据") {
		t.Errorf("期望错误消息包含'没有数据'，实际: %v", err)
	}
}
