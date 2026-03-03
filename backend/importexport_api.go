package backend

import (
	"fmt"

	"mygui/backend/internal/importexport"
	"mygui/backend/internal/repository"
)

// ImportExportAPI 提供数据导入导出相关的 API 方法，供前端调用

// ExportToSQL 导出数据为 SQL INSERT 语句
func (a *App) ExportToSQL(profileID, database, table string, query repository.DataQuery, outputPath string) error {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("导出数据失败: %w", err)
	}
	
	a.logger.Info("Exporting data to SQL", map[string]interface{}{
		"profile_id":  profileID,
		"database":    database,
		"table":       table,
		"output_path": outputPath,
	})
	
	// 创建进度回调函数
	progressCallback := func(current, total int) {
		a.emitEvent("export:progress", map[string]interface{}{
			"profileId":  profileID,
			"database":   database,
			"table":      table,
			"format":     "SQL",
			"current":    current,
			"total":      total,
			"percentage": float64(current) / float64(total) * 100,
		})
	}
	
	err = service.ExportToSQL(database, table, query, outputPath, progressCallback)
	if err != nil {
		a.logger.Error("Failed to export to SQL", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		
		// 发送失败事件
		a.emitEvent("export:failed", map[string]interface{}{
			"profileId": profileID,
			"database":  database,
			"table":     table,
			"format":    "SQL",
			"error":     err.Error(),
		})
		
		return fmt.Errorf("导出数据失败: %w", err)
	}
	
	a.logger.Info("Data exported to SQL successfully", map[string]interface{}{
		"profile_id":  profileID,
		"database":    database,
		"table":       table,
		"output_path": outputPath,
	})
	
	// 发送完成事件
	a.emitEvent("export:completed", map[string]interface{}{
		"profileId":  profileID,
		"database":   database,
		"table":      table,
		"format":     "SQL",
		"outputPath": outputPath,
	})
	
	return nil
}

// ExportToCSV 导出数据为 CSV 格式
func (a *App) ExportToCSV(profileID, database, table string, query repository.DataQuery, outputPath string) error {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("导出数据失败: %w", err)
	}
	
	a.logger.Info("Exporting data to CSV", map[string]interface{}{
		"profile_id":  profileID,
		"database":    database,
		"table":       table,
		"output_path": outputPath,
	})
	
	// 创建进度回调函数
	progressCallback := func(current, total int) {
		a.emitEvent("export:progress", map[string]interface{}{
			"profileId":  profileID,
			"database":   database,
			"table":      table,
			"format":     "CSV",
			"current":    current,
			"total":      total,
			"percentage": float64(current) / float64(total) * 100,
		})
	}
	
	err = service.ExportToCSV(database, table, query, outputPath, progressCallback)
	if err != nil {
		a.logger.Error("Failed to export to CSV", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		
		// 发送失败事件
		a.emitEvent("export:failed", map[string]interface{}{
			"profileId": profileID,
			"database":  database,
			"table":     table,
			"format":    "CSV",
			"error":     err.Error(),
		})
		
		return fmt.Errorf("导出数据失败: %w", err)
	}
	
	a.logger.Info("Data exported to CSV successfully", map[string]interface{}{
		"profile_id":  profileID,
		"database":    database,
		"table":       table,
		"output_path": outputPath,
	})
	
	// 发送完成事件
	a.emitEvent("export:completed", map[string]interface{}{
		"profileId":  profileID,
		"database":   database,
		"table":      table,
		"format":     "CSV",
		"outputPath": outputPath,
	})
	
	return nil
}

// ExportToJSON 导出数据为 JSON 格式
func (a *App) ExportToJSON(profileID, database, table string, query repository.DataQuery, outputPath string) error {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("导出数据失败: %w", err)
	}
	
	a.logger.Info("Exporting data to JSON", map[string]interface{}{
		"profile_id":  profileID,
		"database":    database,
		"table":       table,
		"output_path": outputPath,
	})
	
	// 创建进度回调函数
	progressCallback := func(current, total int) {
		a.emitEvent("export:progress", map[string]interface{}{
			"profileId":  profileID,
			"database":   database,
			"table":      table,
			"format":     "JSON",
			"current":    current,
			"total":      total,
			"percentage": float64(current) / float64(total) * 100,
		})
	}
	
	err = service.ExportToJSON(database, table, query, outputPath, progressCallback)
	if err != nil {
		a.logger.Error("Failed to export to JSON", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		
		// 发送失败事件
		a.emitEvent("export:failed", map[string]interface{}{
			"profileId": profileID,
			"database":  database,
			"table":     table,
			"format":    "JSON",
			"error":     err.Error(),
		})
		
		return fmt.Errorf("导出数据失败: %w", err)
	}
	
	a.logger.Info("Data exported to JSON successfully", map[string]interface{}{
		"profile_id":  profileID,
		"database":    database,
		"table":       table,
		"output_path": outputPath,
	})
	
	// 发送完成事件
	a.emitEvent("export:completed", map[string]interface{}{
		"profileId":  profileID,
		"database":   database,
		"table":      table,
		"format":     "JSON",
		"outputPath": outputPath,
	})
	
	return nil
}

// ImportFromSQL 从 SQL 文件导入数据
func (a *App) ImportFromSQL(profileID, database, sqlFilePath string) (*importexport.ImportResult, error) {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
		})
		return nil, fmt.Errorf("导入数据失败: %w", err)
	}
	
	a.logger.Info("Importing data from SQL", map[string]interface{}{
		"profile_id":    profileID,
		"database":      database,
		"sql_file_path": sqlFilePath,
	})
	
	// 创建进度回调函数
	progressCallback := func(current, total int) {
		a.emitEvent("import:progress", map[string]interface{}{
			"profileId":  profileID,
			"database":   database,
			"format":     "SQL",
			"current":    current,
			"total":      total,
			"percentage": float64(current) / float64(total) * 100,
		})
	}
	
	result, err := service.ImportFromSQL(database, sqlFilePath, progressCallback)
	if err != nil {
		a.logger.Error("Failed to import from SQL", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
		})
		
		// 发送失败事件
		a.emitEvent("import:failed", map[string]interface{}{
			"profileId": profileID,
			"database":  database,
			"format":    "SQL",
			"error":     err.Error(),
		})
		
		return result, fmt.Errorf("导入数据失败: %w", err)
	}
	
	a.logger.Info("Data imported from SQL successfully", map[string]interface{}{
		"profile_id":   profileID,
		"database":     database,
		"total_rows":   result.TotalRows,
		"success_rows": result.SuccessRows,
		"failed_rows":  result.FailedRows,
	})
	
	// 发送完成事件
	a.emitEvent("import:completed", map[string]interface{}{
		"profileId":   profileID,
		"database":    database,
		"format":      "SQL",
		"totalRows":   result.TotalRows,
		"successRows": result.SuccessRows,
		"failedRows":  result.FailedRows,
	})
	
	return result, nil
}

// ImportFromCSV 从 CSV 文件导入数据
func (a *App) ImportFromCSV(profileID, database, table, csvFilePath string, mapping importexport.ColumnMapping) (*importexport.ImportResult, error) {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return nil, fmt.Errorf("导入数据失败: %w", err)
	}
	
	a.logger.Info("Importing data from CSV", map[string]interface{}{
		"profile_id":    profileID,
		"database":      database,
		"table":         table,
		"csv_file_path": csvFilePath,
	})
	
	// 创建进度回调函数
	progressCallback := func(current, total int) {
		a.emitEvent("import:progress", map[string]interface{}{
			"profileId":  profileID,
			"database":   database,
			"table":      table,
			"format":     "CSV",
			"current":    current,
			"total":      total,
			"percentage": float64(current) / float64(total) * 100,
		})
	}
	
	result, err := service.ImportFromCSV(database, table, csvFilePath, mapping, progressCallback)
	if err != nil {
		a.logger.Error("Failed to import from CSV", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		
		// 发送失败事件
		a.emitEvent("import:failed", map[string]interface{}{
			"profileId": profileID,
			"database":  database,
			"table":     table,
			"format":    "CSV",
			"error":     err.Error(),
		})
		
		return result, fmt.Errorf("导入数据失败: %w", err)
	}
	
	a.logger.Info("Data imported from CSV successfully", map[string]interface{}{
		"profile_id":   profileID,
		"database":     database,
		"table":        table,
		"total_rows":   result.TotalRows,
		"success_rows": result.SuccessRows,
		"failed_rows":  result.FailedRows,
	})
	
	// 发送完成事件
	a.emitEvent("import:completed", map[string]interface{}{
		"profileId":   profileID,
		"database":    database,
		"table":       table,
		"format":      "CSV",
		"totalRows":   result.TotalRows,
		"successRows": result.SuccessRows,
		"failedRows":  result.FailedRows,
	})
	
	return result, nil
}

// ImportFromJSON 从 JSON 文件导入数据
func (a *App) ImportFromJSON(profileID, database, table, jsonFilePath string, mapping importexport.ColumnMapping) (*importexport.ImportResult, error) {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return nil, fmt.Errorf("导入数据失败: %w", err)
	}
	
	a.logger.Info("Importing data from JSON", map[string]interface{}{
		"profile_id":     profileID,
		"database":       database,
		"table":          table,
		"json_file_path": jsonFilePath,
	})
	
	// 创建进度回调函数
	progressCallback := func(current, total int) {
		a.emitEvent("import:progress", map[string]interface{}{
			"profileId":  profileID,
			"database":   database,
			"table":      table,
			"format":     "JSON",
			"current":    current,
			"total":      total,
			"percentage": float64(current) / float64(total) * 100,
		})
	}
	
	result, err := service.ImportFromJSON(database, table, jsonFilePath, mapping, progressCallback)
	if err != nil {
		a.logger.Error("Failed to import from JSON", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		
		// 发送失败事件
		a.emitEvent("import:failed", map[string]interface{}{
			"profileId": profileID,
			"database":  database,
			"table":     table,
			"format":    "JSON",
			"error":     err.Error(),
		})
		
		return result, fmt.Errorf("导入数据失败: %w", err)
	}
	
	a.logger.Info("Data imported from JSON successfully", map[string]interface{}{
		"profile_id":   profileID,
		"database":     database,
		"table":        table,
		"total_rows":   result.TotalRows,
		"success_rows": result.SuccessRows,
		"failed_rows":  result.FailedRows,
	})
	
	// 发送完成事件
	a.emitEvent("import:completed", map[string]interface{}{
		"profileId":   profileID,
		"database":    database,
		"table":       table,
		"format":      "JSON",
		"totalRows":   result.TotalRows,
		"successRows": result.SuccessRows,
		"failedRows":  result.FailedRows,
	})
	
	return result, nil
}

// ValidateCSVFormat 验证 CSV 文件格式
func (a *App) ValidateCSVFormat(profileID, csvFilePath string) error {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return fmt.Errorf("验证 CSV 格式失败: %w", err)
	}
	
	err = service.ValidateCSVFormat(csvFilePath)
	if err != nil {
		a.logger.Error("CSV format validation failed", err, map[string]interface{}{
			"profile_id":    profileID,
			"csv_file_path": csvFilePath,
		})
		return fmt.Errorf("验证 CSV 格式失败: %w", err)
	}
	
	return nil
}

// ValidateJSONFormat 验证 JSON 文件格式
func (a *App) ValidateJSONFormat(profileID, jsonFilePath string) error {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return fmt.Errorf("验证 JSON 格式失败: %w", err)
	}
	
	err = service.ValidateJSONFormat(jsonFilePath)
	if err != nil {
		a.logger.Error("JSON format validation failed", err, map[string]interface{}{
			"profile_id":     profileID,
			"json_file_path": jsonFilePath,
		})
		return fmt.Errorf("验证 JSON 格式失败: %w", err)
	}
	
	return nil
}

// ValidateSQLFormat 验证 SQL 文件格式
func (a *App) ValidateSQLFormat(profileID, sqlFilePath string) error {
	service, err := a.getOrCreateImportExportService(profileID)
	if err != nil {
		a.logger.Error("Failed to get import/export service", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return fmt.Errorf("验证 SQL 格式失败: %w", err)
	}
	
	err = service.ValidateSQLFormat(sqlFilePath)
	if err != nil {
		a.logger.Error("SQL format validation failed", err, map[string]interface{}{
			"profile_id":    profileID,
			"sql_file_path": sqlFilePath,
		})
		return fmt.Errorf("验证 SQL 格式失败: %w", err)
	}
	
	return nil
}
