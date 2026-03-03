package data

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mygui/backend/internal/repository"
)

func TestNewManager(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.db)
	assert.NotNil(t, manager.repository)
}

func TestQueryData_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow(1, "Alice", "alice@example.com").
		AddRow(2, "Bob", "bob@example.com")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` LIMIT 100").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := repository.DataQuery{
		Database: "testdb",
		Table:    "users",
	}

	result, err := manager.QueryData(query)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, []string{"id", "name", "email"}, result.Columns)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, int64(2), result.Total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithCustomLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` LIMIT 50").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := repository.DataQuery{
		Database: "testdb",
		Table:    "users",
		Limit:    50,
	}

	result, err := manager.QueryData(query)
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_EmptyDatabase(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	query := repository.DataQuery{
		Database: "",
		Table:    "users",
	}

	result, err := manager.QueryData(query)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "数据库名称不能为空")
}

func TestQueryData_EmptyTable(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	query := repository.DataQuery{
		Database: "testdb",
		Table:    "",
	}

	result, err := manager.QueryData(query)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "表名称不能为空")
}

func TestQueryData_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(1, "Alice", 30)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` WHERE `age` > \\? LIMIT 100").
		WithArgs(25).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` > \\?").
		WithArgs(25).
		WillReturnRows(countRows)

	query := repository.DataQuery{
		Database: "testdb",
		Table:    "users",
		Filters: []repository.Filter{
			{Column: "age", Operator: ">", Value: 25},
		},
	}

	result, err := manager.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithOrderBy(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(2, "Bob", 25).
		AddRow(1, "Alice", 30)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` ORDER BY `age` DESC LIMIT 100").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := repository.DataQuery{
		Database: "testdb",
		Table:    "users",
		OrderBy: []repository.OrderBy{
			{Column: "age", Direction: "DESC"},
		},
	}

	result, err := manager.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_WithPagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(11, "User11").
		AddRow(12, "User12")

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` LIMIT 10 OFFSET 10").
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(25)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	query := repository.DataQuery{
		Database: "testdb",
		Table:    "users",
		Limit:    10,
		Offset:   10,
	}

	result, err := manager.QueryData(query)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, int64(25), result.Total)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryData_RepositoryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	mock.ExpectQuery("SELECT \\* FROM `testdb`.`users` LIMIT 100").
		WillReturnError(sql.ErrConnDone)

	query := repository.DataQuery{
		Database: "testdb",
		Table:    "users",
	}

	result, err := manager.QueryData(query)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "查询数据失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRowCount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(100)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnRows(countRows)

	count, err := manager.GetRowCount("testdb", "users", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(100), count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRowCount_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	countRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(25)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users` WHERE `age` >= \\?").
		WithArgs(18).
		WillReturnRows(countRows)

	filters := []repository.Filter{
		{Column: "age", Operator: ">=", Value: 18},
	}

	count, err := manager.GetRowCount("testdb", "users", filters)
	require.NoError(t, err)
	assert.Equal(t, int64(25), count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRowCount_EmptyDatabase(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	count, err := manager.GetRowCount("", "users", nil)
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "数据库名称不能为空")
}

func TestGetRowCount_EmptyTable(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	count, err := manager.GetRowCount("testdb", "", nil)
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "表名称不能为空")
}

func TestGetRowCount_RepositoryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM `testdb`.`users`").
		WillReturnError(sql.ErrConnDone)

	count, err := manager.GetRowCount("testdb", "users", nil)
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.Contains(t, err.Error(), "获取行数失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertRow_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	data := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	}

	mock.ExpectExec("INSERT INTO `testdb`.`users`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = manager.InsertRow("testdb", "users", data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertRow_EmptyDatabase(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	data := map[string]interface{}{
		"name": "Alice",
	}

	err = manager.InsertRow("", "users", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据库名称不能为空")
}

func TestInsertRow_EmptyTable(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	data := map[string]interface{}{
		"name": "Alice",
	}

	err = manager.InsertRow("testdb", "", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "表名称不能为空")
}

func TestInsertRow_EmptyData(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	data := map[string]interface{}{}

	err = manager.InsertRow("testdb", "users", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "插入数据不能为空")
}

func TestInsertRow_RepositoryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	data := map[string]interface{}{
		"name": "Alice",
	}

	mock.ExpectExec("INSERT INTO `testdb`.`users`").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = manager.InsertRow("testdb", "users", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "插入数据失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

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

	err = manager.UpdateRow("testdb", "users", pk, data)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRow_EmptyDatabase(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pk := map[string]interface{}{"id": 1}
	data := map[string]interface{}{"name": "Alice"}

	err = manager.UpdateRow("", "users", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据库名称不能为空")
}

func TestUpdateRow_EmptyTable(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pk := map[string]interface{}{"id": 1}
	data := map[string]interface{}{"name": "Alice"}

	err = manager.UpdateRow("testdb", "", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "表名称不能为空")
}

func TestUpdateRow_EmptyPrimaryKey(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pk := map[string]interface{}{}
	data := map[string]interface{}{"name": "Alice"}

	err = manager.UpdateRow("testdb", "users", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "主键不能为空")
}

func TestUpdateRow_EmptyData(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pk := map[string]interface{}{"id": 1}
	data := map[string]interface{}{}

	err = manager.UpdateRow("testdb", "users", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "更新数据不能为空")
}

func TestUpdateRow_RepositoryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pk := map[string]interface{}{"id": 1}
	data := map[string]interface{}{"name": "Alice"}

	mock.ExpectExec("UPDATE `testdb`.`users` SET").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnError(sql.ErrConnDone)

	err = manager.UpdateRow("testdb", "users", pk, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "更新数据失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pks := []map[string]interface{}{
		{"id": 1},
		{"id": 2},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	deleted, err := manager.DeleteRows("testdb", "users", pks)
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRows_EmptyDatabase(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pks := []map[string]interface{}{
		{"id": 1},
	}

	deleted, err := manager.DeleteRows("", "users", pks)
	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.Contains(t, err.Error(), "数据库名称不能为空")
}

func TestDeleteRows_EmptyTable(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pks := []map[string]interface{}{
		{"id": 1},
	}

	deleted, err := manager.DeleteRows("testdb", "", pks)
	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.Contains(t, err.Error(), "表名称不能为空")
}

func TestDeleteRows_EmptyPrimaryKeyList(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pks := []map[string]interface{}{}

	deleted, err := manager.DeleteRows("testdb", "users", pks)
	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.Contains(t, err.Error(), "主键列表不能为空")
}

func TestDeleteRows_RepositoryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewManager(db)

	pks := []map[string]interface{}{
		{"id": 1},
	}

	mock.ExpectExec("DELETE FROM `testdb`.`users` WHERE `id` = \\?").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	deleted, err := manager.DeleteRows("testdb", "users", pks)
	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	assert.Contains(t, err.Error(), "删除数据失败")

	assert.NoError(t, mock.ExpectationsWereMet())
}
