package backend

import (
	"fmt"

	"mygui/backend/types"
)

// SyncAPI 提供数据库结构同步相关的 API 方法，供前端调用

// CompareSchemas 比较两个数据库的结构差异
func (a *App) CompareSchemas(sourceProfileID, targetProfileID, sourceDB, targetDB string) (*types.SchemaDiff, error) {
	return a.CompareSchemasWithTables(sourceProfileID, targetProfileID, sourceDB, targetDB, nil)
}

// CompareSchemasWithTables 比较两个数据库的结构差异（支持指定表列表）
func (a *App) CompareSchemasWithTables(sourceProfileID, targetProfileID, sourceDB, targetDB string, tables []string) (*types.SchemaDiff, error) {
	engine, err := a.getOrCreateSyncEngine(sourceProfileID, targetProfileID)
	if err != nil {
		a.logger.Error("Failed to get sync engine", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
		})
		return nil, fmt.Errorf("比较结构失败: %w", err)
	}

	a.logger.Info("Comparing schemas", map[string]interface{}{
		"source_profile": sourceProfileID,
		"target_profile": targetProfileID,
		"source_db":      sourceDB,
		"target_db":      targetDB,
		"tables":         tables,
	})

	diff, err := engine.CompareSchemas(sourceDB, targetDB)
	if err != nil {
		a.logger.Error("Failed to compare schemas", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
			"source_db":      sourceDB,
			"target_db":      targetDB,
		})
		return nil, fmt.Errorf("比较结构失败: %w", err)
	}

	// 如果指定了表列表，过滤差异结果
	if len(tables) > 0 {
		diff = filterDiffByTables(diff, tables)
	}

	a.logger.Info("Schema comparison completed", map[string]interface{}{
		"source_profile":     sourceProfileID,
		"target_profile":     targetProfileID,
		"tables_only_source": len(diff.TablesOnlyInSource),
		"tables_only_target": len(diff.TablesOnlyInTarget),
		"tables_with_diff":   len(diff.TableDifferences),
	})

	return diff, nil
}

// filterDiffByTables 根据表列表过滤差异结果
func filterDiffByTables(diff *types.SchemaDiff, tables []string) *types.SchemaDiff {
	tableSet := make(map[string]bool)
	for _, table := range tables {
		tableSet[table] = true
	}

	filtered := &types.SchemaDiff{
		TablesOnlyInSource: []string{},
		TablesOnlyInTarget: []string{},
		TableDifferences:   []types.TableDiff{},
	}

	// 过滤仅存在于源的表
	for _, table := range diff.TablesOnlyInSource {
		if tableSet[table] {
			filtered.TablesOnlyInSource = append(filtered.TablesOnlyInSource, table)
		}
	}

	// 过滤仅存在于目标的表
	for _, table := range diff.TablesOnlyInTarget {
		if tableSet[table] {
			filtered.TablesOnlyInTarget = append(filtered.TablesOnlyInTarget, table)
		}
	}

	// 过滤有差异的表
	for _, tableDiff := range diff.TableDifferences {
		if tableSet[tableDiff.TableName] {
			filtered.TableDifferences = append(filtered.TableDifferences, tableDiff)
		}
	}

	return filtered
}

// GenerateSyncScript 生成同步脚本
func (a *App) GenerateSyncScript(sourceProfileID, targetProfileID, sourceDB string, diff *types.SchemaDiff) (*types.SyncScript, error) {
	engine, err := a.getOrCreateSyncEngine(sourceProfileID, targetProfileID)
	if err != nil {
		a.logger.Error("Failed to get sync engine", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
		})
		return nil, fmt.Errorf("生成同步脚本失败: %w", err)
	}

	a.logger.Info("Generating sync script", map[string]interface{}{
		"source_profile": sourceProfileID,
		"target_profile": targetProfileID,
		"source_db":      sourceDB,
	})

	script, err := engine.GenerateSyncScript(sourceDB, diff)
	if err != nil {
		a.logger.Error("Failed to generate sync script", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
		})
		return nil, fmt.Errorf("生成同步脚本失败: %w", err)
	}

	a.logger.Info("Sync script generated", map[string]interface{}{
		"source_profile":  sourceProfileID,
		"target_profile":  targetProfileID,
		"statement_count": len(script.Statements),
	})

	return script, nil
}

// ValidateSyncScript 验证同步脚本
func (a *App) ValidateSyncScript(sourceProfileID, targetProfileID string, script *types.SyncScript) error {
	engine, err := a.getOrCreateSyncEngine(sourceProfileID, targetProfileID)
	if err != nil {
		a.logger.Error("Failed to get sync engine", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
		})
		return fmt.Errorf("验证同步脚本失败: %w", err)
	}

	err = engine.ValidateScript(script)
	if err != nil {
		a.logger.Error("Sync script validation failed", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
		})
		return fmt.Errorf("验证同步脚本失败: %w", err)
	}

	return nil
}

// ExecuteSyncScript 执行同步脚本
func (a *App) ExecuteSyncScript(sourceProfileID, targetProfileID, targetDB string, script *types.SyncScript) error {
	engine, err := a.getOrCreateSyncEngine(sourceProfileID, targetProfileID)
	if err != nil {
		a.logger.Error("Failed to get sync engine", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
		})
		return fmt.Errorf("执行同步脚本失败: %w", err)
	}

	a.logger.Info("Executing sync script", map[string]interface{}{
		"source_profile":  sourceProfileID,
		"target_profile":  targetProfileID,
		"target_db":       targetDB,
		"statement_count": len(script.Statements),
	})

	// 创建进度回调函数，发送进度事件到前端
	progressCallback := func(current, total int, statement string) {
		a.emitEvent("sync:progress", map[string]interface{}{
			"sourceProfileId": sourceProfileID,
			"targetProfileId": targetProfileID,
			"current":         current,
			"total":           total,
			"statement":       statement,
			"percentage":      float64(current) / float64(total) * 100,
		})
	}

	err = engine.ExecuteSyncScript(targetDB, script, progressCallback)
	if err != nil {
		a.logger.Error("Failed to execute sync script", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
			"target_db":      targetDB,
		})

		// 发送失败事件
		a.emitEvent("sync:failed", map[string]interface{}{
			"sourceProfileId": sourceProfileID,
			"targetProfileId": targetProfileID,
			"error":           err.Error(),
		})

		return fmt.Errorf("执行同步脚本失败: %w", err)
	}

	a.logger.Info("Sync script executed successfully", map[string]interface{}{
		"source_profile": sourceProfileID,
		"target_profile": targetProfileID,
		"target_db":      targetDB,
	})

	// 发送完成事件
	a.emitEvent("sync:completed", map[string]interface{}{
		"sourceProfileId": sourceProfileID,
		"targetProfileId": targetProfileID,
	})

	return nil
}

// GetSyncSummary 获取同步操作摘要
func (a *App) GetSyncSummary(sourceProfileID, targetProfileID string, diff *types.SchemaDiff) (string, error) {
	engine, err := a.getOrCreateSyncEngine(sourceProfileID, targetProfileID)
	if err != nil {
		a.logger.Error("Failed to get sync engine", err, map[string]interface{}{
			"source_profile": sourceProfileID,
			"target_profile": targetProfileID,
		})
		return "", fmt.Errorf("获取同步摘要失败: %w", err)
	}

	summary := engine.GetSyncSummary(diff)
	return summary, nil
}

// CompareAndGenerateScript 比较结构并生成同步脚本（便捷方法）
func (a *App) CompareAndGenerateScript(sourceProfileID, targetProfileID, sourceDB, targetDB string) (*types.SchemaDiff, *types.SyncScript, error) {
	// 比较结构
	diff, err := a.CompareSchemas(sourceProfileID, targetProfileID, sourceDB, targetDB)
	if err != nil {
		return nil, nil, err
	}

	// 生成同步脚本
	script, err := a.GenerateSyncScript(sourceProfileID, targetProfileID, sourceDB, diff)
	if err != nil {
		return diff, nil, err
	}

	return diff, script, nil
}
