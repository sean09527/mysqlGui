package data

import (
	"database/sql"
	"fmt"
	"mygui/backend/internal/repository"
)

// Manager 处理数据管理的业务逻辑
type Manager struct {
	db         *sql.DB
	repository *repository.DataRepository
}

// NewManager 创建新的 DataManager 实例
func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db:         db,
		repository: repository.NewDataRepository(db),
	}
}

// QueryData 查询表数据
func (m *Manager) QueryData(query repository.DataQuery) (*repository.DataResult, error) {
	if query.Database == "" {
		return nil, fmt.Errorf("数据库名称不能为空")
	}
	if query.Table == "" {
		return nil, fmt.Errorf("表名称不能为空")
	}

	// 设置默认分页参数
	if query.Limit <= 0 {
		query.Limit = 100 // 默认每页 100 行
	}

	result, err := m.repository.QueryData(query)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}

	return result, nil
}

// GetRowCount 获取表的行数（支持筛选）
func (m *Manager) GetRowCount(database, table string, filters []repository.Filter) (int64, error) {
	if database == "" {
		return 0, fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return 0, fmt.Errorf("表名称不能为空")
	}

	count, err := m.repository.GetRowCount(database, table, filters)
	if err != nil {
		return 0, fmt.Errorf("获取行数失败: %w", err)
	}

	return count, nil
}

// InsertRow 插入新行数据
func (m *Manager) InsertRow(database, table string, data map[string]interface{}) error {
	if database == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return fmt.Errorf("表名称不能为空")
	}
	if len(data) == 0 {
		return fmt.Errorf("插入数据不能为空")
	}

	// 执行插入
	err := m.repository.InsertRow(database, table, data)
	if err != nil {
		return fmt.Errorf("插入数据失败: %w", err)
	}

	return nil
}

// UpdateRow 更新行数据（基于主键）
func (m *Manager) UpdateRow(database, table string, pk map[string]interface{}, data map[string]interface{}) error {
	if database == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return fmt.Errorf("表名称不能为空")
	}
	if len(pk) == 0 {
		return fmt.Errorf("主键不能为空")
	}
	if len(data) == 0 {
		return fmt.Errorf("更新数据不能为空")
	}

	// 执行更新
	err := m.repository.UpdateRow(database, table, pk, data)
	if err != nil {
		return fmt.Errorf("更新数据失败: %w", err)
	}

	return nil
}

// DeleteRows 删除行数据（基于主键，支持批量删除）
func (m *Manager) DeleteRows(database, table string, pks []map[string]interface{}) (int64, error) {
	if database == "" {
		return 0, fmt.Errorf("数据库名称不能为空")
	}
	if table == "" {
		return 0, fmt.Errorf("表名称不能为空")
	}
	if len(pks) == 0 {
		return 0, fmt.Errorf("主键列表不能为空")
	}

	// 执行删除
	deleted, err := m.repository.DeleteRows(database, table, pks)
	if err != nil {
		return deleted, fmt.Errorf("删除数据失败: %w", err)
	}

	return deleted, nil
}
