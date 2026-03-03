package repository

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestQueryData_Basic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 模拟查询结果
	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow(1, "Alice", "alice@example.com").
		AddRow(2, "Bob", "bob@example.com")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` LIMIT 10").
		WillReturnRows(rows)

	// 模拟 COUNT 查询
	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Limit:    10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, []string{"id", "name", "email"}, result.Columns)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, int64(2), result.Total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectQuery("SELECT `id`, `name` FROM `testdb`.`users` LIMIT 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Columns:  []string{"id", "name"},
		Limit:    10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Equal(t, []string{"id", "name"}, result.Columns)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(1, "Alice", 30)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `age` > \\? LIMIT 10").
		WithArgs(25).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` > \\?").
		WithArgs(25).
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "age", Operator: ">", Value: 25},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, int64(1), result.Total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithMultipleFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age", "city"}).
		AddRow(1, "Alice", 30, "New York")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `age` > \\? AND `city` = \\? LIMIT 10").
		WithArgs(25, "New York").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` > \\? AND `city` = \\?").
		WithArgs(25, "New York").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "age", Operator: ">", Value: 25},
			{Column: "city", Operator: "=", Value: "New York"},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithLikeFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice").
		AddRow(2, "Alicia")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `name` LIKE \\? LIMIT 10").
		WithArgs("Ali%").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `name` LIKE \\?").
		WithArgs("Ali%").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "name", Operator: "LIKE", Value: "Ali%"},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithInFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "status"}).
		AddRow(1, "Alice", "active").
		AddRow(2, "Bob", "pending")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `status` IN \\(\\?, \\?\\) LIMIT 10").
		WithArgs("active", "pending").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `status` IN \\(\\?, \\?\\)").
		WithArgs("active", "pending").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "status", Operator: "IN", Value: []interface{}{"active", "pending"}},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithOrderBy(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(2, "Bob", 25).
		AddRow(1, "Alice", 30)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` ORDER BY `age` DESC LIMIT 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		OrderBy: []OrderBy{
			{Column: "age", Direction: "DESC"},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithPagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(11, "User11").
		AddRow(12, "User12")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` LIMIT 10 OFFSET 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(25)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Limit:    10,
		Offset:   10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, int64(25), result.Total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRowCount_NoFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(100)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	count, err := repo.GetRowCount("testdb", "users", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(100), count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRowCount_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(25)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` >= \\?").
		WithArgs(18).
		WillReturnRows(countRows)

	filters := []Filter{
		{Column: "age", Operator: ">=", Value: 18},
	}

	count, err := repo.GetRowCount("testdb", "users", filters)
	require.NoError(t, err)
	assert.Equal(t, int64(25), count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuildFilterClause_UnsupportedOperator(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	filter := Filter{
		Column:   "age",
		Operator: "INVALID",
		Value:    25,
	}

	_, _, err = repo.buildFilterClause(filter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的操作符")
}

func TestBuildFilterClause_InWithInvalidValue(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	filter := Filter{
		Column:   "status",
		Operator: "IN",
		Value:    "not-an-array",
	}

	_, _, err = repo.buildFilterClause(filter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "需要数组类型的值")
}

func TestBuildFilterClause_InWithEmptyArray(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	filter := Filter{
		Column:   "status",
		Operator: "IN",
		Value:    []interface{}{},
	}

	_, _, err = repo.buildFilterClause(filter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不能为空")
}

func TestEscapeIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "`users`"},
		{"user_name", "`user_name`"},
		{"`already_escaped`", "`already_escaped`"},
		{"table-name", "`table-name`"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryData_SQLInjectionPrevention(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 尝试 SQL 注入攻击
	maliciousValue := "'; DROP TABLE users; --"

	rows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `name` = \\? LIMIT 10").
		WithArgs(maliciousValue).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `name` = \\?").
		WithArgs(maliciousValue).
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "name", Operator: "=", Value: maliciousValue},
		},
		Limit: 10,
	}

	// 参数化查询应该安全地处理恶意输入
	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` LIMIT 10").
		WillReturnError(sql.ErrConnDone)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Limit:    10,
	}

	result, err := repo.QueryData(query)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "执行查询失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertRow_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	data := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	}

	mock.ExpectExec("INSERT INTO `testdb`.`users`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.InsertRow("testdb", "users", data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertRow_EmptyData(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	data := map[string]interface{}{}

	err = repo.InsertRow("testdb", "users", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "插入数据不能为空")
}

func TestInsertRow_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	data := map[string]interface{}{
		"name": "Alice",
	}

	mock.ExpectExec("INSERT INTO `testdb`.`users`").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = repo.InsertRow("testdb", "users", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "执行插入失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertRow_WithNullValue(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	data := map[string]interface{}{
		"name":  "Alice",
		"email": nil,
	}

	mock.ExpectExec("INSERT INTO `testdb`.`users`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.InsertRow("testdb", "users", data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertRow_SQLInjectionPrevention(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 尝试 SQL 注入攻击
	data := map[string]interface{}{
		"name": "'; DROP TABLE users; --",
	}

	mock.ExpectExec("INSERT INTO `testdb`.`users`").
		WithArgs("'; DROP TABLE users; --").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 参数化查询应该安全地处理恶意输入
	err = repo.InsertRow("testdb", "users", data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pk := map[string]interface{}{
		"id": 1,
	}

	data := map[string]interface{}{
		"name":  "Alice Updated",
		"email": "alice.new@example.com",
	}

	mock.ExpectExec("UPDATE `testdb`.`users` SET").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateRow("testdb", "users", pk, data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_CompositePrimaryKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pk := map[string]interface{}{
		"user_id":  1,
		"group_id": 5,
	}

	data := map[string]interface{}{
		"role": "admin",
	}

	mock.ExpectExec("UPDATE `testdb`.`user_groups` SET").
		WithArgs("admin", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateRow("testdb", "user_groups", pk, data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_EmptyPrimaryKey(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pk := map[string]interface{}{}
	data := map[string]interface{}{
		"name": "Alice",
	}

	err = repo.UpdateRow("testdb", "users", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "主键不能为空")
}

func TestUpdateRow_EmptyData(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pk := map[string]interface{}{
		"id": 1,
	}
	data := map[string]interface{}{}

	err = repo.UpdateRow("testdb", "users", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "更新数据不能为空")
}

func TestUpdateRow_NoRowsAffected(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pk := map[string]interface{}{
		"id": 999,
	}

	data := map[string]interface{}{
		"name": "Alice",
	}

	mock.ExpectExec("UPDATE `testdb`.`users` SET").
		WithArgs(sqlmock.AnyArg(), 999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateRow("testdb", "users", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未找到匹配的行")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pk := map[string]interface{}{
		"id": 1,
	}

	data := map[string]interface{}{
		"name": "Alice",
	}

	mock.ExpectExec("UPDATE `testdb`.`users` SET").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnError(sql.ErrConnDone)

	err = repo.UpdateRow("testdb", "users", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "执行更新失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_SQLInjectionPrevention(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pk := map[string]interface{}{
		"id": 1,
	}

	// 尝试 SQL 注入攻击
	data := map[string]interface{}{
		"name": "'; DROP TABLE users; --",
	}

	mock.ExpectExec("UPDATE `testdb`.`users` SET").
		WithArgs("'; DROP TABLE users; --", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 参数化查询应该安全地处理恶意输入
	err = repo.UpdateRow("testdb", "users", pk, data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_SingleRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pks := []map[string]interface{}{
		{"id": 1},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	deleted, err := repo.DeleteRows("testdb", "users", pks)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_MultipleRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pks := []map[string]interface{}{
		{"id": 1},
		{"id": 2},
		{"id": 3},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))

	deleted, err := repo.DeleteRows("testdb", "users", pks)
	require.NoError(t, err)
	assert.Equal(t, int64(3), deleted)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_CompositePrimaryKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pks := []map[string]interface{}{
		{"user_id": 1, "group_id": 5},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`user_groups` WHERE").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	deleted, err := repo.DeleteRows("testdb", "user_groups", pks)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_EmptyPrimaryKeyList(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pks := []map[string]interface{}{}

	deleted, err := repo.DeleteRows("testdb", "users", pks)
	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.Contains(t, err.Error(), "主键列表不能为空")
}

func TestDeleteRows_EmptyPrimaryKey(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pks := []map[string]interface{}{
		{},
	}

	deleted, err := repo.DeleteRows("testdb", "users", pks)
	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.Contains(t, err.Error(), "主键不能为空")
}

func TestDeleteRows_NoRowsAffected(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pks := []map[string]interface{}{
		{"id": 999},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	deleted, err := repo.DeleteRows("testdb", "users", pks)
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pks := []map[string]interface{}{
		{"id": 1},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	deleted, err := repo.DeleteRows("testdb", "users", pks)
	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.Contains(t, err.Error(), "执行删除失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_PartialSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pks := []map[string]interface{}{
		{"id": 1},
		{"id": 2},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(2).
		WillReturnError(sql.ErrConnDone)

	deleted, err := repo.DeleteRows("testdb", "users", pks)
	assert.Error(t, err)
	assert.Equal(t, int64(1), deleted) // 第一行已删除
	assert.Contains(t, err.Error(), "执行删除失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_SQLInjectionPrevention(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 尝试 SQL 注入攻击
	pks := []map[string]interface{}{
		{"id": "1 OR 1=1"},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs("1 OR 1=1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// 参数化查询应该安全地处理恶意输入
	deleted, err := repo.DeleteRows("testdb", "users", pks)
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// 额外的 SQL 注入防护测试

func TestQueryData_SQLInjectionInTableName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 尝试在表名中注入 SQL
	maliciousTable := "users; DROP TABLE users; --"

	rows := sqlmock.NewRows([]string{"id"})
	// escapeIdentifier 会将表名转义为 `users; DROP TABLE users; --`
	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users; DROP TABLE users; --` LIMIT 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users; DROP TABLE users; --`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    maliciousTable,
		Limit:    10,
	}

	// 标识符转义应该防止 SQL 注入
	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_SQLInjectionInColumnName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 尝试在列名中注入 SQL
	maliciousColumn := "id, (SELECT password FROM admin_users)"

	rows := sqlmock.NewRows([]string{"id, (SELECT password FROM admin_users)"})
	// escapeIdentifier 会将列名转义
	mock.ExpectQuery("SELECT `id, \\(SELECT password FROM admin_users\\)` FROM `testdb`.`users` LIMIT 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Columns:  []string{maliciousColumn},
		Limit:    10,
	}

	// 标识符转义应该防止 SQL 注入
	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertRow_SQLInjectionInColumnName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 尝试在列名中注入 SQL
	data := map[string]interface{}{
		"name); DROP TABLE users; --": "Alice",
	}

	// escapeIdentifier 会转义列名
	mock.ExpectExec("INSERT INTO `testdb`.`users` \\(`name\\); DROP TABLE users; --`\\) VALUES \\(\\?\\)").
		WithArgs("Alice").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 标识符转义应该防止 SQL 注入
	err = repo.InsertRow("testdb", "users", data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_SQLInjectionInPrimaryKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 尝试在主键值中注入 SQL
	pk := map[string]interface{}{
		"id": "1 OR 1=1; DROP TABLE users; --",
	}

	data := map[string]interface{}{
		"name": "Alice",
	}

	mock.ExpectExec("UPDATE `testdb`.`users` SET `name` = \\? WHERE `id` = \\?").
		WithArgs("Alice", "1 OR 1=1; DROP TABLE users; --").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// 参数化查询应该安全地处理恶意输入
	err = repo.UpdateRow("testdb", "users", pk, data)
	assert.Error(t, err) // 应该返回"未找到匹配的行"错误
	assert.Contains(t, err.Error(), "未找到匹配的行")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// 额外的操作符测试

func TestQueryData_NotEqualOperator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "status"}).
		AddRow(1, "Alice", "active").
		AddRow(2, "Bob", "active")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `status` != \\? LIMIT 10").
		WithArgs("inactive").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `status` != \\?").
		WithArgs("inactive").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "status", Operator: "!=", Value: "inactive"},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_LessThanOrEqualOperator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(1, "Alice", 25).
		AddRow(2, "Bob", 30)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `age` <= \\? LIMIT 10").
		WithArgs(30).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` <= \\?").
		WithArgs(30).
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "age", Operator: "<=", Value: 30},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_GreaterThanOrEqualOperator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(2, "Bob", 30).
		AddRow(3, "Charlie", 35)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `age` >= \\? LIMIT 10").
		WithArgs(30).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` >= \\?").
		WithArgs(30).
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "age", Operator: ">=", Value: 30},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_LessThanOperator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(1, "Alice", 25)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `age` < \\? LIMIT 10").
		WithArgs(30).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` < \\?").
		WithArgs(30).
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "age", Operator: "<", Value: 30},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_NotLikeOperator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(3, "Charlie")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `name` NOT LIKE \\? LIMIT 10").
		WithArgs("Ali%").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `name` NOT LIKE \\?").
		WithArgs("Ali%").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "name", Operator: "NOT LIKE", Value: "Ali%"},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_NotInOperator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "status"}).
		AddRow(3, "Charlie", "inactive")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `status` NOT IN \\(\\?, \\?\\) LIMIT 10").
		WithArgs("active", "pending").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `status` NOT IN \\(\\?, \\?\\)").
		WithArgs("active", "pending").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []Filter{
			{Column: "status", Operator: "NOT IN", Value: []interface{}{"active", "pending"}},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// 数据类型验证测试

func TestInsertRow_WithVariousDataTypes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	// 测试各种数据类型
	data := map[string]interface{}{
		"name":       "Alice",           // 字符串
		"age":        30,                // 整数
		"salary":     50000.50,          // 浮点数
		"is_active":  true,              // 布尔值
		"birth_date": "1990-01-01",      // 日期字符串
		"notes":      nil,               // NULL 值
		"score":      int64(95),         // int64
		"rating":     float32(4.5),      // float32
	}

	mock.ExpectExec("INSERT INTO `testdb`.`users`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.InsertRow("testdb", "users", data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_WithVariousDataTypes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	pk := map[string]interface{}{
		"id": 1,
	}

	// 测试各种数据类型
	data := map[string]interface{}{
		"name":      "Alice Updated",
		"age":       31,
		"salary":    55000.75,
		"is_active": false,
		"notes":     "Updated notes",
	}

	mock.ExpectExec("UPDATE `testdb`.`users` SET").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 
			sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateRow("testdb", "users", pk, data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow(1, "Alice", nil).
		AddRow(2, "Bob", "bob@example.com")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` LIMIT 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Limit:    10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)
	assert.Nil(t, result.Rows[0][2]) // email 应该是 nil

	assert.NoError(t, mock.ExpectationsWereMet())
}

// 边界情况测试

func TestQueryData_WithZeroLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users`").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Limit:    0, // 不应用 LIMIT
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithMultipleOrderBy(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age", "city"}).
		AddRow(1, "Alice", 30, "New York").
		AddRow(2, "Bob", 30, "Boston")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` ORDER BY `age` DESC, `name` ASC LIMIT 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		OrderBy: []OrderBy{
			{Column: "age", Direction: "DESC"},
			{Column: "name", Direction: "ASC"},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_OrderByWithInvalidDirection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(1, "Alice", 30)

	// 无效的排序方向应该默认为 ASC
	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` ORDER BY `age` ASC LIMIT 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		OrderBy: []OrderBy{
			{Column: "age", Direction: "INVALID"},
		},
		Limit: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscapeIdentifier_WithBackticks(t *testing.T) {
	// 测试已经包含反引号的标识符
	result := escapeIdentifier("`users`")
	assert.Equal(t, "`users`", result)
}

func TestEscapeIdentifier_WithSpecialCharacters(t *testing.T) {
	// 测试包含特殊字符的标识符
	result := escapeIdentifier("user-table")
	assert.Equal(t, "`user-table`", result)
	
	result = escapeIdentifier("user_table")
	assert.Equal(t, "`user_table`", result)
	
	result = escapeIdentifier("user.table")
	assert.Equal(t, "`user.table`", result)
}

func TestQueryData_ComplexQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDataRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age", "city"}).
		AddRow(1, "Alice", 30, "New York")

	// 复杂查询：多个筛选条件 + 排序 + 分页
	mock.ExpectQuery("SELECT `id`, `name`, `age`, `city` FROM `testdb`.`users` WHERE `age` > \\? AND `city` = \\? ORDER BY `age` DESC LIMIT 20 OFFSET 10").
		WithArgs(25, "New York").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` > \\? AND `city` = \\?").
		WithArgs(25, "New York").
		WillReturnRows(countRows)

	query := DataQuery{
		Database: "testdb",
		Table:    "users",
		Columns:  []string{"id", "name", "age", "city"},
		Filters: []Filter{
			{Column: "age", Operator: ">", Value: 25},
			{Column: "city", Operator: "=", Value: "New York"},
		},
		OrderBy: []OrderBy{
			{Column: "age", Direction: "DESC"},
		},
		Limit:  20,
		Offset: 10,
	}

	result, err := repo.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, int64(1), result.Total)

	assert.NoError(t, mock.ExpectationsWereMet())
}
