# 错误处理和用户反馈使用指南

本指南介绍如何在应用中使用全局错误处理、加载状态管理和确认对话框功能。

## 目录

1. [错误处理](#错误处理)
2. [加载状态管理](#加载状态管理)
3. [确认对话框](#确认对话框)
4. [最佳实践](#最佳实践)

## 错误处理

### 基本用法

```typescript
import { showError, showSuccess, showWarning, showInfo } from '@/utils/errorHandler';

// 显示错误消息
try {
  await someOperation();
} catch (error) {
  showError(error);
}

// 显示成功消息
showSuccess('操作成功！');

// 显示警告消息
showWarning('请注意数据格式');

// 显示信息消息
showInfo('正在处理中...');
```

### 显示详细错误通知

```typescript
import { showErrorNotification } from '@/utils/errorHandler';

try {
  await complexOperation();
} catch (error) {
  showErrorNotification(error, '操作失败');
}
```

### 包装异步操作

```typescript
import { handleAsyncOperation } from '@/utils/errorHandler';

// 自动处理错误和成功消息
const result = await handleAsyncOperation(
  async () => {
    return await api.saveData(data);
  },
  {
    successMessage: '数据保存成功',
    errorTitle: '保存失败',
  }
);

if (result) {
  // 操作成功
  console.log('Result:', result);
}
```

### 错误类型

系统会自动识别以下错误类型：

- **CONNECTION**: 数据库连接错误、SSH 连接错误
- **SQL**: SQL 语法错误、表不存在、外键约束等
- **NETWORK**: 网络错误、请求超时
- **VALIDATION**: 数据验证失败
- **PERMISSION**: 权限不足
- **TIMEOUT**: 操作超时
- **UNKNOWN**: 未知错误

### 全局错误处理器

在 `main.ts` 中设置全局错误处理：

```typescript
import { globalErrorHandler, setupGlobalErrorHandlers } from '@/utils/errorHandler';

// Vue 错误处理
app.config.errorHandler = globalErrorHandler;

// 全局未捕获错误处理
setupGlobalErrorHandlers();
```

## 加载状态管理

### 全局加载状态

```typescript
import { useLoading } from '@/composables/useLoading';

const { withLoading, startOperation, updateProgress, completeOperation } = useLoading();

// 方式 1: 使用 withLoading 包装器
async function exportData() {
  await withLoading(
    async (updateProgress) => {
      // 执行操作
      for (let i = 0; i <= 100; i += 10) {
        await processChunk(i);
        updateProgress(i, `正在导出... ${i}%`);
      }
    },
    {
      message: '正在导出数据',
      successMessage: '导出完成',
      cancelable: true,
      showFullscreen: true,
    }
  );
}

// 方式 2: 手动管理
async function importData() {
  const operationId = startOperation('import-data', '正在导入数据', {
    cancelable: true,
    showFullscreen: true,
  });

  try {
    for (let i = 0; i <= 100; i += 10) {
      await processChunk(i);
      updateProgress(operationId, i, `正在导入... ${i}%`);
    }
    completeOperation(operationId, '导入完成');
  } catch (error) {
    completeOperation(operationId);
    throw error;
  }
}
```

### 局部加载状态

```typescript
import { useLocalLoading } from '@/composables/useLoading';

const { loading, loadingMessage, withLocalLoading } = useLocalLoading();

// 在模板中使用
<el-button :loading="loading" @click="handleSave">
  {{ loading ? loadingMessage : '保存' }}
</el-button>

// 在方法中使用
async function handleSave() {
  await withLocalLoading(async () => {
    await api.saveData(data);
  }, '正在保存...');
}
```

### 在组件中使用

```vue
<template>
  <div v-loading="loading" :element-loading-text="loadingMessage">
    <!-- 内容 -->
  </div>
</template>

<script setup lang="ts">
import { useLocalLoading } from '@/composables/useLoading';

const { loading, loadingMessage, startLoading, stopLoading } = useLocalLoading();

async function loadData() {
  startLoading('加载数据中...');
  try {
    const data = await api.fetchData();
    // 处理数据
  } finally {
    stopLoading();
  }
}
</script>
```

## 确认对话框

### 使用 useConfirm Composable

```typescript
import { useConfirm } from '@/composables/useConfirm';

const {
  confirm,
  confirmDelete,
  confirmDropTable,
  confirmDeleteData,
  confirmSync,
  confirmAlterTable,
} = useConfirm();

// 通用确认
async function handleOperation() {
  const confirmed = await confirm({
    title: '确认操作',
    message: '确定要执行此操作吗？',
    detail: '此操作可能会影响数据。',
    type: 'warning',
  });

  if (confirmed) {
    // 执行操作
  }
}

// 删除表确认
async function handleDropTable(tableName: string) {
  const confirmed = await confirmDropTable(tableName);
  if (confirmed) {
    await api.dropTable(tableName);
  }
}

// 删除数据确认
async function handleDeleteRows(selectedRows: any[]) {
  const confirmed = await confirmDeleteData(selectedRows.length);
  if (confirmed) {
    await api.deleteRows(selectedRows);
  }
}

// 同步确认
async function handleSync() {
  const confirmed = await confirmSync('source_db', 'target_db', 15);
  if (confirmed) {
    await api.executeSync();
  }
}

// 修改表结构确认
async function handleAlterTable(tableName: string, hasDataLoss: boolean) {
  const confirmed = await confirmAlterTable(tableName, hasDataLoss);
  if (confirmed) {
    await api.alterTable(tableName, changes);
  }
}
```

### 使用 ConfirmDialog 组件

```vue
<template>
  <div>
    <el-button @click="showDialog = true">删除</el-button>

    <ConfirmDialog
      v-model="showDialog"
      title="确认删除"
      message="确定要删除这个项目吗？"
      detail="此操作不可撤销"
      type="warning"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    >
      <!-- 可选：额外内容 -->
      <el-checkbox v-model="deleteRelated">
        同时删除关联数据
      </el-checkbox>
    </ConfirmDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import ConfirmDialog from '@/components/ConfirmDialog.vue';

const showDialog = ref(false);
const deleteRelated = ref(false);

function handleConfirm() {
  console.log('Confirmed, deleteRelated:', deleteRelated.value);
  showDialog.value = false;
  // 执行删除操作
}

function handleCancel() {
  console.log('Cancelled');
}
</script>
```

## 最佳实践

### 1. 组合使用错误处理和加载状态

```typescript
import { useLoading } from '@/composables/useLoading';
import { showError, showSuccess } from '@/utils/errorHandler';

const { withLoading } = useLoading();

async function saveData() {
  try {
    await withLoading(
      async () => {
        await api.saveData(data);
      },
      {
        message: '正在保存数据',
        showFullscreen: true,
      }
    );
    showSuccess('数据保存成功');
  } catch (error) {
    showError(error);
  }
}
```

### 2. 危险操作前确认

```typescript
import { useConfirm } from '@/composables/useConfirm';
import { useLoading } from '@/composables/useLoading';
import { showError, showSuccess } from '@/utils/errorHandler';

const { confirmDropTable } = useConfirm();
const { withLoading } = useLoading();

async function dropTable(tableName: string) {
  // 1. 先确认
  const confirmed = await confirmDropTable(tableName);
  if (!confirmed) return;

  // 2. 显示加载状态并执行
  try {
    await withLoading(
      async () => {
        await api.dropTable(tableName);
      },
      {
        message: '正在删除表',
        showFullscreen: true,
      }
    );
    showSuccess('表删除成功');
  } catch (error) {
    showError(error);
  }
}
```

### 3. 带进度的长时间操作

```typescript
import { useLoading } from '@/composables/useLoading';
import { showError, showSuccess } from '@/utils/errorHandler';

const { withLoading } = useLoading();

async function exportLargeData() {
  try {
    await withLoading(
      async (updateProgress) => {
        const totalRows = await api.getRowCount();
        const batchSize = 1000;
        let processed = 0;

        while (processed < totalRows) {
          await api.exportBatch(processed, batchSize);
          processed += batchSize;
          const progress = Math.min((processed / totalRows) * 100, 100);
          updateProgress(progress, `已导出 ${processed}/${totalRows} 行`);
        }
      },
      {
        message: '正在导出数据',
        successMessage: '数据导出完成',
        cancelable: true,
        showFullscreen: true,
      }
    );
  } catch (error) {
    showError(error);
  }
}
```

### 4. 在 API 调用中统一处理

```typescript
// api/index.ts
import { showError } from '@/utils/errorHandler';

export async function apiCall<T>(fn: () => Promise<T>): Promise<T | null> {
  try {
    return await fn();
  } catch (error) {
    showError(error);
    return null;
  }
}

// 使用
const data = await apiCall(() => api.fetchData());
if (data) {
  // 处理数据
}
```

### 5. 表单验证错误处理

```typescript
import { showWarning } from '@/utils/errorHandler';

function validateForm(formData: any): boolean {
  if (!formData.name) {
    showWarning('请输入名称');
    return false;
  }
  if (!formData.host) {
    showWarning('请输入主机地址');
    return false;
  }
  return true;
}
```

## 注意事项

1. **错误消息应该用户友好**: 避免直接显示技术性错误消息，使用 `errorHandler` 会自动转换为友好消息
2. **长时间操作必须显示进度**: 超过 2 秒的操作应该显示加载指示器
3. **危险操作必须确认**: 删除、修改结构等操作必须先确认
4. **可取消的操作应该提供取消功能**: 导入、导出等长时间操作应该允许用户取消
5. **错误要记录到控制台**: 所有错误都会自动记录到控制台，便于调试

## 示例：完整的数据删除流程

```typescript
import { useConfirm } from '@/composables/useConfirm';
import { useLoading } from '@/composables/useLoading';
import { showError, showSuccess } from '@/utils/errorHandler';

async function deleteSelectedRows(selectedRows: any[]) {
  // 1. 验证选择
  if (selectedRows.length === 0) {
    showWarning('请先选择要删除的数据');
    return;
  }

  // 2. 确认删除
  const { confirmDeleteData } = useConfirm();
  const confirmed = await confirmDeleteData(selectedRows.length);
  if (!confirmed) return;

  // 3. 执行删除
  const { withLoading } = useLoading();
  try {
    await withLoading(
      async (updateProgress) => {
        const total = selectedRows.length;
        for (let i = 0; i < total; i++) {
          await api.deleteRow(selectedRows[i]);
          updateProgress(((i + 1) / total) * 100, `已删除 ${i + 1}/${total} 条`);
        }
      },
      {
        message: '正在删除数据',
        showFullscreen: true,
      }
    );
    showSuccess(`成功删除 ${selectedRows.length} 条数据`);
    
    // 4. 刷新数据
    await refreshData();
  } catch (error) {
    showError(error);
  }
}
```
