package backend

import (
	"fmt"

	"mygui/backend/internal/repository"
	"mygui/backend/internal/schema"
)

// SchemaAPI 提供表结构管理相关的 API 方法，供前端调用

// ListDatabases 获取所有数据库列表
func (a *App) ListDatabases(profileID string) ([]repository.Database, error) {
	manager, err := a.getOrCreateSchemaManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get schema manager", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return nil, fmt.Errorf("获取数据库列表失败: %w", err)
	}

	databases, err := manager.ListDatabases()
	if err != nil {
		a.logger.Error("Failed to list databases", err, map[string]interface{}{
			"profile_id": profileID,
		})
		return nil, fmt.Errorf("获取数据库列表失败: %w", err)
	}

	return databases, nil
}

// ListTables 获取指定数据库的所有表列表
func (a *App) ListTables(profileID, database string) ([]repository.Table, error) {
	manager, err := a.getOrCreateSchemaManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get schema manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
		})
		return nil, fmt.Errorf("获取表列表失败: %w", err)
	}

	tables, err := manager.ListTables(database)
	if err != nil {
		a.logger.Error("Failed to list tables", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
		})
		return nil, fmt.Errorf("获取表列表失败: %w", err)
	}

	return tables, nil
}

// GetTableSchema 获取表的完整结构信息
func (a *App) GetTableSchema(profileID, database, table string) (*schema.TableSchema, error) {
	manager, err := a.getOrCreateSchemaManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get schema manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return nil, fmt.Errorf("获取表结构失败: %w", err)
	}

	tableSchema, err := manager.GetTableSchema(database, table)
	if err != nil {
		a.logger.Error("Failed to get table schema", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return nil, fmt.Errorf("获取表结构失败: %w", err)
	}

	return tableSchema, nil
}

// GetCreateTableDDL 获取表的 CREATE TABLE 语句
func (a *App) GetCreateTableDDL(profileID, database, table string) (string, error) {
	manager, err := a.getOrCreateSchemaManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get schema manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return "", fmt.Errorf("获取 DDL 失败: %w", err)
	}

	ddl, err := manager.GetCreateTableDDL(database, table)
	if err != nil {
		a.logger.Error("Failed to get CREATE TABLE DDL", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return "", fmt.Errorf("获取 DDL 失败: %w", err)
	}

	return ddl, nil
}

// CreateTable 创建新表
func (a *App) CreateTable(profileID, database string, tableSchema schema.TableSchema) error {
	manager, err := a.getOrCreateSchemaManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get schema manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      tableSchema.Name,
		})
		return fmt.Errorf("创建表失败: %w", err)
	}

	a.logger.Info("Creating table", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      tableSchema.Name,
	})

	err = manager.CreateTable(database, tableSchema)
	if err != nil {
		a.logger.Error("Failed to create table", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      tableSchema.Name,
		})
		return fmt.Errorf("创建表失败: %w", err)
	}

	a.logger.Info("Table created successfully", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      tableSchema.Name,
	})

	// 发送事件通知前端
	a.emitEvent("schema:table:created", map[string]interface{}{
		"profileId": profileID,
		"database":  database,
		"table":     tableSchema.Name,
	})

	return nil
}

// AlterTable 修改表结构
func (a *App) AlterTable(profileID, database, table string, newSchema schema.TableSchema) error {
	manager, err := a.getOrCreateSchemaManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get schema manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("修改表结构失败: %w", err)
	}

	a.logger.Info("Altering table", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
	})

	// 获取当前表结构
	currentSchema, err := manager.GetTableSchema(database, table)
	if err != nil {
		a.logger.Error("Failed to get current table schema", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("获取当前表结构失败: %w", err)
	}

	// 计算变更
	changes := schema.CompareSchemas(currentSchema, &newSchema)

	if len(changes) == 0 {
		a.logger.Info("No changes detected", map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return nil
	}

	a.logger.Info("Applying schema changes", map[string]interface{}{
		"profile_id":   profileID,
		"database":     database,
		"table":        table,
		"change_count": len(changes),
	})

	err = manager.AlterTable(database, table, changes)
	if err != nil {
		a.logger.Error("Failed to alter table", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("修改表结构失败: %w", err)
	}

	a.logger.Info("Table altered successfully", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
	})

	// 发送事件通知前端
	a.emitEvent("schema:table:altered", map[string]interface{}{
		"profileId": profileID,
		"database":  database,
		"table":     table,
	})

	return nil
}

// DropTable 删除表
func (a *App) DropTable(profileID, database, table string) error {
	manager, err := a.getOrCreateSchemaManager(profileID)
	if err != nil {
		a.logger.Error("Failed to get schema manager", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("删除表失败: %w", err)
	}

	a.logger.Info("Dropping table", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
	})

	err = manager.DropTable(database, table)
	if err != nil {
		a.logger.Error("Failed to drop table", err, map[string]interface{}{
			"profile_id": profileID,
			"database":   database,
			"table":      table,
		})
		return fmt.Errorf("删除表失败: %w", err)
	}

	a.logger.Info("Table dropped successfully", map[string]interface{}{
		"profile_id": profileID,
		"database":   database,
		"table":      table,
	})

	// 发送事件通知前端
	a.emitEvent("schema:table:dropped", map[string]interface{}{
		"profileId": profileID,
		"database":  database,
		"table":     table,
	})

	return nil
}
