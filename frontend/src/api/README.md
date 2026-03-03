# Frontend API 使用说明

## 概述

此目录包含前端与后端通信的 API 包装器。所有 API 调用都通过 Wails 框架的绑定机制与 Go 后端通信。

## Wails 绑定生成

在开发或构建应用时，Wails 会自动生成 TypeScript 绑定文件到 `frontend/wailsjs/go/backend/` 目录。

### 生成绑定的方法

1. **开发模式**：运行 `wails dev`，绑定会自动生成
2. **构建模式**：运行 `wails build`，绑定会在构建前生成

### 绑定文件位置

生成的绑定文件位于：
- `frontend/wailsjs/go/backend/App.js` - JavaScript 绑定
- `frontend/wailsjs/go/backend/App.d.ts` - TypeScript 类型定义

## API 使用

### 导入 API

```typescript
import { ConnectionAPI, DatabaseAPI, SchemaAPI, DataAPI } from '@/api';
```

### 使用示例

#### 连接管理

```typescript
// 获取所有连接配置
const profiles = await ConnectionAPI.listProfiles();

// 测试连接
await ConnectionAPI.testConnection(profile);

// 连接到数据库
await ConnectionAPI.connect(profileId);
```

#### 数据库操作

```typescript
// 获取数据库列表
const databases = await DatabaseAPI.listDatabases(profileId);

// 获取表列表
const tables = await DatabaseAPI.listTables(profileId, database);
```

#### 表结构管理

```typescript
// 获取表结构
const schema = await SchemaAPI.getTableSchema(profileId, database, table);

// 创建表
await SchemaAPI.createTable(profileId, database, tableSchema);
```

#### 数据管理

```typescript
// 查询数据
const result = await DataAPI.queryData(profileId, {
  database: 'mydb',
  table: 'users',
  columns: ['*'],
  filters: [],
  orderBy: [],
  limit: 100,
  offset: 0,
});

// 插入数据
await DataAPI.insertRow(profileId, database, table, {
  name: 'John',
  email: 'john@example.com',
});
```

## 错误处理

所有 API 调用都返回 Promise，应该使用 try-catch 处理错误：

```typescript
try {
  await ConnectionAPI.connect(profileId);
  ElMessage.success('连接成功');
} catch (error) {
  ElMessage.error(`连接失败: ${error.message}`);
}
```

## 事件监听

Wails 支持从后端向前端发送事件。使用 `runtime.EventsOn` 监听事件：

```typescript
import { EventsOn } from '../../wailsjs/runtime/runtime';

// 监听连接状态变化
EventsOn('connection:status:changed', (data) => {
  console.log('Connection status changed:', data);
});
```

## 开发注意事项

1. **绑定未生成时**：如果 Wails 绑定尚未生成，API 调用会抛出错误提示
2. **类型安全**：使用 TypeScript 类型定义确保类型安全
3. **错误处理**：始终处理 API 调用可能抛出的错误
4. **加载状态**：在 API 调用期间显示加载指示器，提升用户体验

## API 列表

### ConnectionAPI
- `createProfile(profile)` - 创建连接配置
- `updateProfile(id, profile)` - 更新连接配置
- `deleteProfile(id)` - 删除连接配置
- `listProfiles()` - 获取所有连接配置
- `testConnection(profile)` - 测试连接
- `connect(profileId)` - 连接到数据库
- `disconnect(profileId)` - 断开连接
- `getConnectionStatus(profileId)` - 获取连接状态

### DatabaseAPI
- `listDatabases(profileId)` - 获取数据库列表
- `listTables(profileId, database)` - 获取表列表
- `getTableRowCount(profileId, database, table)` - 获取表行数

### SchemaAPI
- `getTableSchema(profileId, database, table)` - 获取表结构
- `createTable(profileId, database, schema)` - 创建表
- `alterTable(profileId, database, table, changes)` - 修改表结构
- `dropTable(profileId, database, table)` - 删除表
- `getCreateTableDDL(profileId, database, table)` - 获取 CREATE TABLE DDL

### DataAPI
- `queryData(profileId, query)` - 查询数据
- `insertRow(profileId, database, table, data)` - 插入行
- `updateRow(profileId, database, table, pk, data)` - 更新行
- `deleteRows(profileId, database, table, pks)` - 删除行

### QueryAPI
- `executeQuery(profileId, sql)` - 执行 SQL 查询
- `cancelQuery(profileId, queryId)` - 取消查询

### SyncAPI
- `compareSchemas(sourceProfileId, targetProfileId, sourceDB, targetDB)` - 比较数据库结构
- `generateSyncScript(sourceProfileId, targetProfileId, diff)` - 生成同步脚本
- `executeSyncScript(targetProfileId, script)` - 执行同步脚本

### ImportExportAPI
- `exportData(profileId, database, table, format, query)` - 导出数据
- `importData(profileId, database, table, file, format, mapping)` - 导入数据

### LogAPI
- `getLogs(filter)` - 获取日志
- `exportLogs(startTime, endTime)` - 导出日志
