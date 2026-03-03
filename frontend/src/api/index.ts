/**
 * API 包装器 - 封装所有前后端通信
 * 
 * 使用 Wails 生成的绑定调用后端方法
 */

import type {
  ConnectionProfile,
  Database,
  Table,
  TableSchema,
  SchemaChange,
  DataQuery,
  DataResult,
  QueryResult,
  SchemaDiff,
  SyncScript,
  ImportResult,
  ColumnMapping,
  LogEntry,
  LogFilter,
} from '../types/api';

// 导入 Wails 生成的绑定
import * as App from '../../wailsjs/go/backend/App';

// 检查 Wails 运行时是否可用
const isWailsAvailable = (): boolean => {
  return typeof window !== 'undefined' && 
         (window as any)['go'] !== undefined && 
         (window as any)['go']['backend'] !== undefined &&
         (window as any)['go']['backend']['App'] !== undefined;
};

// 包装 Wails 调用，提供更好的错误信息
const wailsCall = async <T>(fn: () => Promise<T>, methodName: string): Promise<T> => {
  if (!isWailsAvailable()) {
    throw new Error(
      `Wails 运行时不可用。请确保应用程序通过 Wails 启动。\n` +
      `方法: ${methodName}\n` +
      `提示: 使用 'wails dev' 命令启动应用程序。`
    );
  }
  try {
    return await fn();
  } catch (error: any) {
    // 增强错误信息
    throw new Error(`${methodName} 调用失败: ${error.message || error}`);
  }
};

/**
 * 连接管理 API
 */
export const ConnectionAPI = {
  async createProfile(profile: ConnectionProfile): Promise<void> {
    return wailsCall(() => App.CreateProfile(profile as any), 'CreateProfile');
  },

  async updateProfile(id: string, profile: ConnectionProfile): Promise<void> {
    return wailsCall(() => App.UpdateProfile(id, profile as any), 'UpdateProfile');
  },

  async deleteProfile(id: string): Promise<void> {
    return wailsCall(() => App.DeleteProfile(id), 'DeleteProfile');
  },

  async listProfiles(): Promise<ConnectionProfile[]> {
    return wailsCall(() => App.ListProfiles() as Promise<any>, 'ListProfiles');
  },

  async testConnection(profile: ConnectionProfile): Promise<void> {
    return wailsCall(() => App.TestConnection(profile as any), 'TestConnection');
  },

  async connect(profileId: string): Promise<void> {
    return wailsCall(() => App.Connect(profileId), 'Connect');
  },

  async disconnect(profileId: string): Promise<void> {
    return wailsCall(() => App.Disconnect(profileId), 'Disconnect');
  },

  async getConnectionStatus(profileId: string): Promise<string> {
    return wailsCall(() => App.GetConnectionStatus(profileId), 'GetConnectionStatus');
  },
};

/**
 * 数据库和表管理 API
 */
export const DatabaseAPI = {
  async listDatabases(profileId: string): Promise<Database[]> {
    return wailsCall(() => App.ListDatabases(profileId), 'ListDatabases');
  },

  async listTables(profileId: string, database: string): Promise<Table[]> {
    return wailsCall(() => App.ListTables(profileId, database), 'ListTables');
  },

  async getTableRowCount(profileId: string, database: string, table: string): Promise<number> {
    return wailsCall(() => App.GetRowCount(profileId, database, table, []) as any, 'GetRowCount');
  },
};

/**
 * 表结构管理 API
 */
export const SchemaAPI = {
  async getTableSchema(profileId: string, database: string, table: string): Promise<TableSchema> {
    return wailsCall(() => App.GetTableSchema(profileId, database, table) as Promise<any>, 'GetTableSchema');
  },

  async createTable(profileId: string, database: string, schema: TableSchema): Promise<void> {
    return wailsCall(() => App.CreateTable(profileId, database, schema as any), 'CreateTable');
  },

  async alterTable(profileId: string, database: string, table: string, changes: SchemaChange[]): Promise<void> {
    return wailsCall(() => App.AlterTable(profileId, database, table, changes as any), 'AlterTable');
  },

  async dropTable(profileId: string, database: string, table: string): Promise<void> {
    return wailsCall(() => App.DropTable(profileId, database, table), 'DropTable');
  },

  async getCreateTableDDL(profileId: string, database: string, table: string): Promise<string> {
    return wailsCall(() => App.GetCreateTableDDL(profileId, database, table), 'GetCreateTableDDL');
  },
};

/**
 * 数据管理 API
 */
export const DataAPI = {
  async queryData(profileId: string, query: DataQuery): Promise<DataResult> {
    return wailsCall(() => App.QueryData(profileId, query as any) as Promise<any>, 'QueryData');
  },

  async insertRow(profileId: string, database: string, table: string, data: Record<string, any>): Promise<void> {
    return wailsCall(() => App.InsertRow(profileId, database, table, data), 'InsertRow');
  },

  async updateRow(profileId: string, database: string, table: string, pk: Record<string, any>, data: Record<string, any>): Promise<void> {
    return wailsCall(() => App.UpdateRow(profileId, database, table, pk, data), 'UpdateRow');
  },

  async deleteRows(profileId: string, database: string, table: string, pks: Record<string, any>[]): Promise<void> {
    await wailsCall(() => App.DeleteRows(profileId, database, table, pks), 'DeleteRows');
  },
};

/**
 * SQL 查询 API
 */
export const QueryAPI = {
  async executeQuery(profileId: string, sql: string): Promise<QueryResult> {
    return wailsCall(() => App.ExecuteQuery(profileId, sql) as Promise<any>, 'ExecuteQuery');
  },

  async cancelQuery(profileId: string, queryId: string): Promise<void> {
    return wailsCall(() => App.CancelQuery(profileId, queryId), 'CancelQuery');
  },
};

/**
 * 结构同步 API
 */
export const SyncAPI = {
  async compareSchemas(sourceProfileId: string, targetProfileId: string, sourceDB: string, targetDB: string, tables?: string[]): Promise<SchemaDiff> {
    if (tables && tables.length > 0) {
      // 使用带表过滤的方法
      return wailsCall(() => App.CompareSchemasWithTables(sourceProfileId, targetProfileId, sourceDB, targetDB, tables) as Promise<any>, 'CompareSchemasWithTables');
    } else {
      // 使用原有方法
      return wailsCall(() => App.CompareSchemas(sourceProfileId, targetProfileId, sourceDB, targetDB) as Promise<any>, 'CompareSchemas');
    }
  },

  async generateSyncScript(sourceProfileId: string, targetProfileId: string, sourceDB: string, diff: SchemaDiff): Promise<SyncScript> {
    return wailsCall(() => App.GenerateSyncScript(sourceProfileId, targetProfileId, sourceDB, diff as any) as Promise<any>, 'GenerateSyncScript');
  },

  async executeSyncScript(sourceProfileId: string, targetProfileId: string, targetDatabase: string, script: SyncScript): Promise<void> {
    return wailsCall(() => App.ExecuteSyncScript(sourceProfileId, targetProfileId, targetDatabase, script as any), 'ExecuteSyncScript');
  },
};

/**
 * 导入导出 API
 */
export const ImportExportAPI = {
  async exportData(profileId: string, database: string, table: string, format: string, query: DataQuery): Promise<string> {
    if (format === 'sql') {
      return wailsCall(() => App.ExportToSQL(profileId, database, table, query as any, '') as any, 'ExportToSQL');
    } else if (format === 'csv') {
      return wailsCall(() => App.ExportToCSV(profileId, database, table, query as any, '') as any, 'ExportToCSV');
    } else if (format === 'json') {
      return wailsCall(() => App.ExportToJSON(profileId, database, table, query as any, '') as any, 'ExportToJSON');
    }
    throw new Error(`不支持的导出格式: ${format}`);
  },

  async importData(profileId: string, database: string, table: string, file: string, format: string, mapping: ColumnMapping): Promise<ImportResult> {
    if (format === 'sql') {
      return wailsCall(() => App.ImportFromSQL(profileId, database, file) as Promise<any>, 'ImportFromSQL');
    } else if (format === 'csv') {
      return wailsCall(() => App.ImportFromCSV(profileId, database, table, file, mapping as any) as Promise<any>, 'ImportFromCSV');
    } else if (format === 'json') {
      return wailsCall(() => App.ImportFromJSON(profileId, database, table, file, mapping as any) as Promise<any>, 'ImportFromJSON');
    }
    throw new Error(`不支持的导入格式: ${format}`);
  },
};

/**
 * 日志 API
 */
export const LogAPI = {
  async getLogs(filter: LogFilter): Promise<LogEntry[]> {
    // 日志功能暂未实现
    console.warn('日志功能暂未实现');
    return [];
  },

  async exportLogs(startTime: string, endTime: string): Promise<string> {
    // 日志功能暂未实现
    console.warn('日志功能暂未实现');
    return '';
  },
};

/**
 * 数据同步 API
 */
export const DataSyncAPI = {
  async syncTableData(
    sourceProfileId: string,
    targetProfileId: string,
    sourceDatabase: string,
    sourceTable: string,
    targetDatabase: string,
    targetTable: string
  ): Promise<void> {
    return wailsCall(
      () => App.SyncTableData(sourceProfileId, targetProfileId, sourceDatabase, sourceTable, targetDatabase, targetTable),
      'SyncTableData'
    );
  },
};
