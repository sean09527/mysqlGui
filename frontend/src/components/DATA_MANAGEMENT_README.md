# 数据管理组件

本文档描述了数据管理界面的组件结构和使用方法。

## 组件概览

### 1. DataManager.vue
主数据管理组件，集成了所有数据管理功能。

**功能:**
- 数据查看和分页
- 数据插入
- 数据更新（单元格编辑）
- 数据删除（单行和批量）
- 数据筛选和排序
- 自动加载表结构

**Props:**
- `profileId`: 连接配置 ID
- `database`: 数据库名称
- `table`: 表名称

**使用示例:**
```vue
<DataManager
  :profile-id="currentProfileId"
  :database="currentDatabase"
  :table="currentTable"
/>
```

### 2. DataGrid.vue
数据表格组件，使用 Element Plus Table 实现。

**功能:**
- 表格显示数据
- 分页（每页 50/100/200/500 行）
- 列排序
- 单元格双击编辑
- 行选择
- 支持类型化编辑器

**Props:**
- `data`: 二维数组数据
- `columns`: 列名数组
- `columnSchemas`: 列结构映射（可选）
- `foreignKeys`: 外键列表（可选）
- `profileId`: 连接配置 ID（可选）
- `database`: 数据库名称（可选）
- `total`: 总行数
- `loading`: 加载状态
- `sortable`: 是否支持排序
- `editable`: 是否支持编辑
- `showPagination`: 是否显示分页
- `pageSize`: 每页大小

**Events:**
- `selectionChange`: 选择变化
- `sortChange`: 排序变化
- `cellEdit`: 单元格编辑
- `pageChange`: 分页变化

### 3. DataFilter.vue
数据筛选组件，支持多列组合筛选。

**功能:**
- 筛选条件构建器
- 支持多种操作符（等于、不等于、大于、小于、包含、为空等）
- 显示当前筛选条件
- 支持多列组合筛选

**Props:**
- `columns`: 列名数组

**Events:**
- `apply`: 应用筛选
- `clear`: 清除筛选

**支持的操作符:**
- `=`: 等于
- `!=`: 不等于
- `>`: 大于
- `<`: 小于
- `>=`: 大于等于
- `<=`: 小于等于
- `LIKE`: 包含
- `NOT LIKE`: 不包含
- `IN`: 在列表中
- `NOT IN`: 不在列表中
- `IS NULL`: 为空
- `IS NOT NULL`: 不为空

### 4. DataInsertDialog.vue
数据插入对话框组件。

**功能:**
- 根据列类型生成输入控件
- 数据类型验证
- 外键列的关联数据选择器
- 自动填充默认值
- 必填字段验证

**Props:**
- `modelValue`: 对话框显示状态
- `columns`: 列结构数组
- `foreignKeys`: 外键列表
- `profileId`: 连接配置 ID
- `database`: 数据库名称
- `table`: 表名称

**Events:**
- `update:modelValue`: 更新显示状态
- `success`: 插入成功

**支持的列类型:**
- 自增列：显示为只读
- 外键列：下拉选择器
- 布尔类型：开关
- 日期时间类型：日期选择器
- 文本类型：文本域
- 数字类型：数字输入框
- 其他类型：普通输入框

### 5. CellEditor.vue
单元格编辑器组件，支持类型化编辑。

**功能:**
- 根据列类型显示不同的编辑器
- 数据类型验证
- 外键列的关联数据选择器
- 自动聚焦

**Props:**
- `value`: 当前值
- `column`: 列结构
- `foreignKeys`: 外键列表
- `profileId`: 连接配置 ID
- `database`: 数据库名称

**Events:**
- `change`: 值变化
- `cancel`: 取消编辑

## 数据流

1. **加载数据:**
   - DataManager 加载表结构（SchemaAPI.getTableSchema）
   - DataManager 查询数据（DataAPI.queryData）
   - 数据传递给 DataGrid 显示

2. **筛选数据:**
   - 用户在 DataFilter 中设置筛选条件
   - DataFilter 发出 `apply` 事件
   - DataManager 重新查询数据

3. **插入数据:**
   - 用户点击"插入"按钮
   - DataInsertDialog 显示
   - 用户填写表单并提交
   - 调用 DataAPI.insertRow
   - 刷新数据

4. **更新数据:**
   - 用户双击单元格
   - DataGrid 显示 CellEditor
   - 用户编辑并确认
   - DataGrid 发出 `cellEdit` 事件
   - DataManager 调用 DataAPI.updateRow
   - 更新本地数据

5. **删除数据:**
   - 用户选择行并点击"删除"按钮
   - 显示确认对话框
   - DataManager 调用 DataAPI.deleteRows
   - 刷新数据

## API 调用

所有组件通过 `frontend/src/api/index.ts` 中的 API 包装器调用后端方法：

- `DataAPI.queryData()`: 查询数据
- `DataAPI.insertRow()`: 插入行
- `DataAPI.updateRow()`: 更新行
- `DataAPI.deleteRows()`: 删除行
- `SchemaAPI.getTableSchema()`: 获取表结构

## 需求映射

- **需求 8 (表数据查看)**: DataManager + DataGrid
- **需求 9 (数据筛选和排序)**: DataFilter + DataGrid 排序
- **需求 10 (数据插入)**: DataInsertDialog
- **需求 11 (数据更新)**: DataGrid + CellEditor
- **需求 12 (数据删除)**: DataManager 删除功能

## 注意事项

1. **主键要求**: 更新和删除操作需要表有主键
2. **外键支持**: 外键列会自动显示为下拉选择器
3. **类型验证**: 所有输入都会根据列类型进行验证
4. **错误处理**: 所有 API 调用都有错误处理和用户提示
5. **性能优化**: 使用分页和按需加载外键选项
