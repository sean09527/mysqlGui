package backend

import (
	"fmt"

	"mygui/backend/internal/repository"
)

// DataAPI 提供数据管理相关的 API 方法，供前端调用

// QueryData 查询表数据
func (a *App) QueryData(profileID string, query repository.DataQuery) (*repository.DataResult, error) {
	manager, err := a.getOrCreateDataManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get data manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   query.Database,
			"table":      query.Table,
		})
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	
	result, err := manager.QueryData(query)
	if err != nil {
		a.logger.Error("Failed to query data", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   query.Database,
			"table":      query.Table,
		})
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	
	a.logger.Debug("Data queried", map[string]interface{}{
		"profile_id": profileID,
		"database":   query.Database,
		"table":      query.Table,
		"rows":       len(result.Rows),
	})
	
	return result, nil
}

// GetRowCount 获取表的行数（支持筛选）
func (a *App) GetRowCount(profileID, database, table string, filters []repository.Filter) (int64, error) {
	manager, err := a.getOrCreateDataManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get data manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return 0, fmt.Errorf("获取行数失败: %w", err)
	}
	
	count, err := manager.GetRowCount(database, table, filters)
	if err != nil {
		a.logger.Error("Failed to get row count", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return 0, fmt.Errorf("获取行数失败: %w", err)
	}
	
	return count, nil
}

// InsertRow 插入新行数据
func (a *App) InsertRow(profileID, database, table string, data map[string]interface{}) error {
	manager, err := a.getOrCreateDataManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get data manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("插入数据失败: %w", err)
	}
	
	a.logger.Info("Inserting row", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
	})
	
	err = manager.InsertRow(database, table, data)
	if err != nil {
		a.logger.Error("Failed to insert row", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("插入数据失败: %w", err)
	}
	
	a.logger.Info("Row inserted successfully", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
	})
	
	// 发送事件通知前端
	a.emitEvent("data:row:inserted", map[string]interface{}{
		"profileId": profileID,
		"database":  database,
		"table":     table,
	})
	
	return nil
}

// UpdateRow 更新行数据
func (a *App) UpdateRow(profileID, database, table string, pk map[string]interface{}, data map[string]interface{}) error {
	manager, err := a.getOrCreateDataManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get data manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("更新数据失败: %w", err)
	}
	
	a.logger.Info("Updating row", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
	})
	
	err = manager.UpdateRow(database, table, pk, data)
	if err != nil {
		a.logger.Error("Failed to update row", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("更新数据失败: %w", err)
	}
	
	a.logger.Info("Row updated successfully", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
	})
	
	// 发送事件通知前端
	a.emitEvent("data:row:updated", map[string]interface{}{
		"profileId": profileID,
		"database":  database,
		"table":     table,
	})
	
	return nil
}

// DeleteRows 删除行数据（支持批量删除）
func (a *App) DeleteRows(profileID, database, table string, pks []map[string]interface{}) (int64, error) {
	manager, err := a.getOrCreateDataManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get data manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return 0, fmt.Errorf("删除数据失败: %w", err)
	}
	
	a.logger.Info("Deleting rows", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
		"count":      len(pks),
	})
	
	deleted, err := manager.DeleteRows(database, table, pks)
	if err != nil {
		a.logger.Error("Failed to delete rows", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return deleted, fmt.Errorf("删除数据失败: %w", err)
	}
	
	a.logger.Info("Rows deleted successfully", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
		"deleted":    deleted,
	})
	
	// 发送事件通知前端
	a.emitEvent("data:rows:deleted", map[string]interface{}{
		"profileId": profileID,
		"database":  database,
		"table":     table,
		"count":     deleted,
	})
	
	return deleted, nil
}
