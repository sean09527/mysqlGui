package backend

import (
	"fmt"
)

// SyncTableData 同步表数据（逐行处理）
func (a *App) SyncTableData(sourceProfileID, targetProfileID, sourceDB, sourceTable, targetDB, targetTable string) error {
	// 获取源和目标连接
	sourceConn, err := a.connectionManager.GetConnection(sourceProfileID)
	if err != nil {
		a.logger.Error("Failed to get source connection", err, map[string]interface{}{
			"profile_id": sourceProfileID,
		})
		return fmt.Errorf("获取源连接失败: %w", err)
	}

	targetConn, err := a.connectionManager.GetConnection(targetProfileID)
	if err != nil {
		a.logger.Error("Failed to get target connection", err, map[string]interface{}{
			"profile_id": targetProfileID,
		})
		return fmt.Errorf("获取目标连接失败: %w", err)
	}

	a.logger.Info("Starting table data sync", map[string]interface{}{
		"source_profile": sourceProfileID,
		"target_profile": targetProfileID,
		"source_db":      sourceDB,
		"source_table":   sourceTable,
		"target_db":      targetDB,
		"target_table":   targetTable,
	})

	// 1. 清空目标表
	a.logger.Info("Truncating target table", map[string]interface{}{
		"target_db":    targetDB,
		"target_table": targetTable,
	})

	truncateSQL := fmt.Sprintf("TRUNCATE TABLE `%s`.`%s`", targetDB, targetTable)
	_, err = targetConn.Exec(truncateSQL)
	if err != nil {
		a.logger.Error("Failed to truncate target table", err, map[string]interface{}{
			"target_db":    targetDB,
			"target_table": targetTable,
		})
		return fmt.Errorf("清空目标表失败: %w", err)
	}

	// 2. 查询源表所有数据
	a.logger.Info("Querying source table data", map[string]interface{}{
		"source_db":    sourceDB,
		"source_table": sourceTable,
	})

	selectSQL := fmt.Sprintf("SELECT * FROM `%s`.`%s`", sourceDB, sourceTable)
	rows, err := sourceConn.Query(selectSQL)
	if err != nil {
		a.logger.Error("Failed to query source table", err, map[string]interface{}{
			"source_db":    sourceDB,
			"source_table": sourceTable,
		})
		return fmt.Errorf("查询源表数据失败: %w", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		a.logger.Error("Failed to get columns", err, nil)
		return fmt.Errorf("获取列信息失败: %w", err)
	}

	// 3. 准备插入语句
	columnNames := ""
	placeholders := ""
	for i, col := range columns {
		if i > 0 {
			columnNames += ", "
			placeholders += ", "
		}
		columnNames += fmt.Sprintf("`%s`", col)
		placeholders += "?"
	}

	insertSQL := fmt.Sprintf("INSERT INTO `%s`.`%s` (%s) VALUES (%s)",
		targetDB, targetTable, columnNames, placeholders)

	// 准备插入语句
	stmt, err := targetConn.Prepare(insertSQL)
	if err != nil {
		a.logger.Error("Failed to prepare insert statement", err, map[string]interface{}{
			"sql": insertSQL,
		})
		return fmt.Errorf("准备插入语句失败: %w", err)
	}
	defer stmt.Close()

	// 4. 逐行插入数据
	rowCount := 0
	batchSize := 1000
	batchCount := 0

	for rows.Next() {
		// 创建扫描目标
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行
		if err := rows.Scan(valuePtrs...); err != nil {
			a.logger.Error("Failed to scan row", err, map[string]interface{}{
				"row_count": rowCount,
			})
			return fmt.Errorf("扫描行数据失败 (行 %d): %w", rowCount+1, err)
		}

		// 插入到目标表
		_, err = stmt.Exec(values...)
		if err != nil {
			a.logger.Error("Failed to insert row", err, map[string]interface{}{
				"row_count": rowCount,
			})
			return fmt.Errorf("插入数据失败 (行 %d): %w", rowCount+1, err)
		}

		rowCount++

		// 每处理一批数据，发送进度事件
		if rowCount%batchSize == 0 {
			batchCount++
			a.emitEvent("data-sync:progress", map[string]interface{}{
				"sourceProfileId": sourceProfileID,
				"targetProfileId": targetProfileID,
				"sourceDatabase":  sourceDB,
				"sourceTable":     sourceTable,
				"targetDatabase":  targetDB,
				"targetTable":     targetTable,
				"rowsProcessed":   rowCount,
			})
		}
	}

	if err := rows.Err(); err != nil {
		a.logger.Error("Error iterating rows", err, nil)
		return fmt.Errorf("遍历数据失败: %w", err)
	}

	a.logger.Info("Table data sync completed", map[string]interface{}{
		"source_profile": sourceProfileID,
		"target_profile": targetProfileID,
		"rows_synced":    rowCount,
	})

	// 发送完成事件
	a.emitEvent("data-sync:completed", map[string]interface{}{
		"sourceProfileId": sourceProfileID,
		"targetProfileId": targetProfileID,
		"sourceDatabase":  sourceDB,
		"sourceTable":     sourceTable,
		"targetDatabase":  targetDB,
		"targetTable":     targetTable,
		"rowsSynced":      rowCount,
	})

	return nil
}
