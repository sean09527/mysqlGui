# SQL 查询编辑器组件

## 概述

SQL 查询编辑器提供了一个功能完整的 SQL 查询界面，支持语法高亮、查询执行、结果显示和历史记录管理。

## 组件结构

### QueryEditor.vue
主查询编辑器组件，包含以下功能：

#### 功能特性
1. **数据库选择**
   - 下拉选择器显示所有可用数据库
   - 支持搜索过滤
   - 选择数据库后自动加载表列表用于自动补全

2. **SQL 编辑器**
   - 使用 CodeMirror 6 实现
   - SQL 语法高亮
   - 暗色主题 (One Dark)
   - 快捷键支持 (Ctrl+Enter / Cmd+Enter 执行查询)
   - 表名自动补全（选择数据库后可用）

3. **查询执行**
   - 执行按钮：执行当前 SQL 语句
   - 自动添加 USE 语句（如果选择了数据库）
   - 取消按钮：取消正在执行的查询
   - 清空按钮：清空编辑器内容
   - 执行时间显示

4. **结果显示**
   - **SELECT 查询**：以表格形式显示结果，包含行数和执行时间
   - **DML 语句** (INSERT/UPDATE/DELETE)：显示受影响的行数
   - **DDL 语句**：显示执行成功消息
   - 支持 NULL 值和对象类型的格式化显示

4. **错误处理**
   - 显示错误代码和错误消息
   - 显示错误位置（如果可用）
   - 友好的错误提示

5. **查询历史**
   - 集成 QueryHistory 组件
   - 点击历史记录自动填充到编辑器

### QueryHistory.vue
查询历史记录组件，包含以下功能：

#### 功能特性
1. **历史记录列表**
   - 显示最近 100 条查询记录
   - 显示执行状态（成功/失败）
   - 显示执行时间和受影响行数
   - 显示数据库名称（如果有）

2. **搜索功能**
   - 支持按 SQL 内容搜索
   - 支持按数据库名称搜索
   - 实时过滤

3. **历史管理**
   - 刷新按钮：重新加载历史记录
   - 清空按钮：清空所有历史记录（需确认）

4. **时间显示**
   - 智能时间格式化（刚刚、X分钟前、X小时前、X天前）
   - 超过7天显示完整日期时间

5. **交互**
   - 点击历史记录填充到编辑器
   - 失败的查询用红色边框标识
   - 悬停效果

## 使用方法

### 在视图中使用

```vue
<template>
  <div class="query-view">
    <QueryEditor :profile-id="currentConnection.id" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConnectionStore } from '../stores/connection'
import QueryEditor from '../components/QueryEditor.vue'

const connectionStore = useConnectionStore()
const currentConnection = computed(() => connectionStore.currentConnection)
</script>
```

### Props

#### QueryEditor
- `profileId` (string, required): 数据库连接配置 ID

#### QueryHistory
- `profileId` (string, required): 数据库连接配置 ID

### Events

#### QueryHistory
- `select-query`: 当用户点击历史记录时触发，参数为 SQL 字符串

## 后端 API

组件使用以下后端 API：

### ExecuteQuery
```typescript
ExecuteQuery(profileId: string, sql: string): Promise<QueryResult>
```
执行 SQL 查询（默认 30 秒超时）

### CancelQuery
```typescript
CancelQuery(profileId: string, queryId: string): Promise<void>
```
取消正在执行的查询

### GetQueryHistory
```typescript
GetQueryHistory(profileId: string, limit: number): Promise<QueryHistoryEntry[]>
```
获取查询历史记录

### ClearQueryHistory
```typescript
ClearQueryHistory(profileId: string): Promise<void>
```
清空查询历史记录

## 数据类型

### QueryResult
```typescript
interface QueryResult {
  id: string                    // 查询 ID
  type: string                  // 查询类型: SELECT, INSERT, UPDATE, DELETE, DDL, OTHER
  columns?: string[]            // 列名（SELECT 查询）
  rows?: any[][]               // 数据行（SELECT 查询）
  rowsAffected: number         // 受影响的行数
  executionTime: number        // 执行时间（纳秒）
  error?: QueryError           // 错误信息（如果有）
}
```

### QueryError
```typescript
interface QueryError {
  code: number                 // 错误代码
  message: string              // 错误消息
  position: number             // 错误位置（-1 表示未知）
}
```

### QueryHistoryEntry
```typescript
interface QueryHistoryEntry {
  id: number                   // 历史记录 ID
  timestamp: any               // 时间戳
  connectionId: string         // 连接 ID
  database: string             // 数据库名称
  sql: string                  // SQL 语句
  executionTime: number        // 执行时间（毫秒）
  rowsAffected: number         // 受影响的行数
  success: boolean             // 是否成功
}
```

## 样式定制

组件使用 scoped 样式，可以通过以下方式定制：

```vue
<style>
/* 覆盖编辑器高度 */
.query-editor .editor-container {
  flex: 0 0 400px;
}

/* 覆盖结果表格高度 */
.query-editor .el-table {
  max-height: 500px;
}
</style>
```

## 依赖

- Vue 3
- Element Plus
- CodeMirror 6
  - @codemirror/lang-sql
  - @codemirror/theme-one-dark
- Wails 绑定

## 注意事项

1. **查询超时**：默认超时时间为 30 秒，由后端控制
2. **大结果集**：建议使用 LIMIT 限制结果集大小，避免前端渲染性能问题
3. **查询取消**：取消功能依赖后端的 context 取消机制
4. **历史记录**：历史记录存储在本地 SQLite 数据库中
5. **敏感信息**：历史记录中不会记录密码等敏感信息

## 实现的需求

- ✅ 需求 13.1: 提供 SQL 编辑器供用户输入 SQL 语句，支持 MySQL 函数提示
- ✅ 需求 13.2: 支持 SQL 语法高亮显示
- ✅ 需求 13.3: 对于 SELECT 查询，以表格形式显示查询结果
- ✅ 需求 13.4: 对于 INSERT、UPDATE 或 DELETE 语句，显示受影响的行数
- ✅ 需求 13.5: 如果 SQL 执行失败，显示错误信息和错误位置
- ✅ 需求 13.6: 记录最近执行的 SQL 语句历史
- ✅ 需求 13.7: 在 30 秒内完成 SQL 查询执行，超时则终止查询

## 未来改进

1. **SQL 自动完成**：添加表名、列名、函数的自动完成
2. **SQL 格式化**：添加 SQL 格式化功能
3. **多标签页**：支持多个查询标签页
4. **结果导出**：支持将查询结果导出为 CSV/JSON
5. **查询计划**：显示查询执行计划
6. **语法检查**：在执行前进行基本的语法检查
