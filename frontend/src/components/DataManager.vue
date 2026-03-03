<template>
  <div class="data-manager">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" size="small" @click="showInsertDialog" :icon="Plus">
          插入
        </el-button>
        <el-button
          type="danger"
          size="small"
          @click="handleDelete"
          :disabled="selectedRows.length === 0"
          :icon="Delete"
        >
          删除 ({{ selectedRows.length }})
        </el-button>
        <el-button size="small" @click="refreshData" :icon="Refresh">
          刷新
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-button size="small" @click="toggleFilter" :icon="Filter">
          {{ showFilter ? '隐藏筛选' : '显示筛选' }}
        </el-button>
        <el-button size="small" @click="showExportDialog" :icon="Download">
          导出
        </el-button>
        <el-button size="small" @click="showImportDialog" :icon="Upload">
          导入
        </el-button>
      </div>
    </div>

    <!-- 筛选器 -->
    <div v-if="showFilter" class="filter-container">
      <DataFilter
        :columns="columns"
        @apply="handleFilterApply"
        @clear="handleFilterClear"
        ref="filterRef"
      />
    </div>

    <!-- 数据表格 -->
    <div class="grid-container">
      <DataGrid
        :data="dataRows"
        :columns="columns"
        :column-schemas="columnSchemas"
        :foreign-keys="foreignKeys"
        :profile-id="profileId"
        :database="database"
        :total="totalRows"
        :loading="loading"
        :sortable="true"
        :editable="true"
        :show-pagination="true"
        :page-size="pageSize"
        @selection-change="handleSelectionChange"
        @sort-change="handleSortChange"
        @cell-edit="handleCellEdit"
        @page-change="handlePageChange"
        ref="gridRef"
      />
    </div>

    <!-- 插入对话框 -->
    <DataInsertDialog
      v-model="insertDialogVisible"
      :columns="tableColumns"
      :foreign-keys="foreignKeys"
      :profile-id="profileId"
      :database="database"
      :table="table"
      @success="handleInsertSuccess"
    />

    <!-- 导出对话框 -->
    <ExportDialog
      v-model="exportDialogVisible"
      :profile-id="profileId"
      :database="database"
      :table="table"
      :filters="filters"
      :order-by="orderBy"
      @success="handleExportSuccess"
    />

    <!-- 导入对话框 -->
    <ImportDialog
      v-model="importDialogVisible"
      :profile-id="profileId"
      :database="database"
      :table="table"
      :table-columns="tableColumns"
      @success="handleImportSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Delete, Refresh, Filter, Download, Upload } from '@element-plus/icons-vue';
import DataGrid from './DataGrid.vue';
import DataFilter from './DataFilter.vue';
import DataInsertDialog from './DataInsertDialog.vue';
import ExportDialog from './ExportDialog.vue';
import ImportDialog from './ImportDialog.vue';
import { DataAPI, SchemaAPI } from '../api';
import type { Filter as FilterType, OrderBy, Column, ForeignKey, TableSchema } from '../types/api';

interface Props {
  profileId: string;
  database: string;
  table: string;
}

const props = defineProps<Props>();

const loading = ref(false);
const showFilter = ref(false);
const insertDialogVisible = ref(false);
const exportDialogVisible = ref(false);
const importDialogVisible = ref(false);

// 数据
const dataRows = ref<any[][]>([]);
const columns = ref<string[]>([]);
const totalRows = ref(0);
const pageSize = ref(100);
const currentPage = ref(1);

// 表结构
const tableSchema = ref<TableSchema | null>(null);
const tableColumns = ref<Column[]>([]);
const foreignKeys = ref<ForeignKey[]>([]);

// 列结构映射
const columnSchemas = computed(() => {
  const schemas: Record<string, Column> = {};
  tableColumns.value.forEach(col => {
    schemas[col.name] = col;
  });
  return schemas;
});

// 主键列
const primaryKeyColumns = computed(() => {
  if (!tableSchema.value || !tableSchema.value.indexes) {
    return [];
  }
  // 主键在 indexes 中，type 为 "PRIMARY"
  const primaryIndex = tableSchema.value.indexes.find(idx => idx.type === 'PRIMARY');
  return primaryIndex ? primaryIndex.columns : [];
});

// 查询条件
const filters = ref<FilterType[]>([]);
const orderBy = ref<OrderBy[]>([]);
const selectedRows = ref<any[]>([]);

// 引用
const gridRef = ref<any>(null);
const filterRef = ref<any>(null);

// 加载表结构
const loadTableSchema = async () => {
  try {
    const schema = await SchemaAPI.getTableSchema(props.profileId, props.database, props.table);
    tableSchema.value = schema;
    
    // 安全地处理可能为 null 的数组
    tableColumns.value = schema.columns || [];
    foreignKeys.value = schema.foreignKeys || [];
    
    // 立即从 schema 设置 columns，这样筛选器就能显示列名
    if (tableColumns.value.length > 0) {
      columns.value = tableColumns.value.map(col => col.name);
    }
    
    console.log('Table schema loaded:', {
      columns: columns.value,
      tableColumns: tableColumns.value.length,
      foreignKeys: foreignKeys.value.length
    });
  } catch (error: any) {
    console.error('Failed to load table schema:', error);
    ElMessage.error(`加载表结构失败: ${error.message || error}`);
  }
};

// 加载数据
const loadData = async () => {
  loading.value = true;

  try {
    // 查询数据 - 注意：Wails 生成的类型使用大写开头的属性名
    const result = await DataAPI.queryData(props.profileId, {
      Database: props.database,
      Table: props.table,
      Columns: [],
      Filters: filters.value,
      OrderBy: orderBy.value,
      Limit: pageSize.value,
      Offset: (currentPage.value - 1) * pageSize.value,
    } as any);

    // Wails 返回的属性名也是大写开头的
    dataRows.value = result.Rows || [];
    // 只在 columns 为空时才从结果中设置（优先使用 schema 中的列名）
    if (columns.value.length === 0) {
      columns.value = result.Columns || [];
    }
    totalRows.value = result.Total || 0;
    
    console.log('Data loaded:', {
      rows: dataRows.value.length,
      columns: columns.value,
      total: totalRows.value
    });
  } catch (error: any) {
    console.error('Failed to load data:', error);
    ElMessage.error(`加载数据失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

// 刷新数据
const refreshData = () => {
  loadData();
};

// 切换筛选器
const toggleFilter = () => {
  showFilter.value = !showFilter.value;
};

// 显示插入对话框
const showInsertDialog = () => {
  insertDialogVisible.value = true;
};

// 显示导出对话框
const showExportDialog = () => {
  exportDialogVisible.value = true;
};

// 显示导入对话框
const showImportDialog = () => {
  importDialogVisible.value = true;
};

// 处理导出成功
const handleExportSuccess = () => {
  ElMessage.success('导出完成');
};

// 处理导入成功
const handleImportSuccess = () => {
  refreshData();
  ElMessage.success('导入完成');
};

// 处理插入成功
const handleInsertSuccess = () => {
  refreshData();
};

// 处理选择变化
const handleSelectionChange = (rows: any[]) => {
  selectedRows.value = rows;
};

// 处理排序变化
const handleSortChange = (newOrderBy: OrderBy[]) => {
  // 转换为 Wails 需要的大写格式
  orderBy.value = newOrderBy.map(o => ({
    Column: (o as any).column || (o as any).Column,
    Direction: (o as any).direction || (o as any).Direction,
  } as any));
  currentPage.value = 1;
  loadData();
};

// 处理筛选应用
const handleFilterApply = (newFilters: FilterType[]) => {
  // 转换为 Wails 需要的大写格式
  filters.value = newFilters.map(f => ({
    Column: (f as any).column || (f as any).Column,
    Operator: (f as any).operator || (f as any).Operator,
    Value: (f as any).value || (f as any).Value,
  } as any));
  currentPage.value = 1;
  loadData();
};

// 处理筛选清除
const handleFilterClear = () => {
  filters.value = [];
  currentPage.value = 1;
  loadData();
};

// 处理单元格编辑
const handleCellEdit = async (rowIndex: number, column: string, oldValue: any, newValue: any) => {
  loading.value = true;

  try {
    // 获取主键值
    const pk: Record<string, any> = {};
    primaryKeyColumns.value.forEach((pkCol) => {
      const colIndex = columns.value.indexOf(pkCol);
      if (colIndex >= 0) {
        pk[pkCol] = dataRows.value[rowIndex][colIndex];
      }
    });

    if (Object.keys(pk).length === 0) {
      ElMessage.error('无法确定主键，无法更新数据');
      return;
    }

    // 更新数据
    const data: Record<string, any> = { [column]: newValue };
    await DataAPI.updateRow(props.profileId, props.database, props.table, pk, data);

    ElMessage.success('更新成功');
    
    // 更新本地数据
    const colIndex = columns.value.indexOf(column);
    if (colIndex >= 0) {
      dataRows.value[rowIndex][colIndex] = newValue;
    }
  } catch (error: any) {
    console.error('Failed to update row:', error);
    ElMessage.error(`更新失败: ${error.message || error}`);
    
    // 恢复原值
    const colIndex = columns.value.indexOf(column);
    if (colIndex >= 0) {
      dataRows.value[rowIndex][colIndex] = oldValue;
    }
  } finally {
    loading.value = false;
  }
};

// 处理删除
const handleDelete = async () => {
  if (selectedRows.value.length === 0) {
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedRows.value.length} 行数据吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    );
  } catch {
    return;
  }

  loading.value = true;

  try {
    // 构建主键列表
    const pks: Array<Record<string, any>> = [];
    
    for (const row of selectedRows.value) {
      const pk: Record<string, any> = {};
      primaryKeyColumns.value.forEach((pkCol) => {
        if (row[pkCol] !== undefined) {
          pk[pkCol] = row[pkCol];
        }
      });
      
      if (Object.keys(pk).length > 0) {
        pks.push(pk);
      }
    }

    if (pks.length === 0) {
      ElMessage.error('无法确定主键，无法删除数据');
      return;
    }

    await DataAPI.deleteRows(props.profileId, props.database, props.table, pks);

    ElMessage.success(`成功删除 ${pks.length} 行数据`);
    
    // 清除选择
    gridRef.value?.clearSelection();
    selectedRows.value = [];
    
    // 刷新数据
    refreshData();
  } catch (error: any) {
    console.error('Failed to delete rows:', error);
    ElMessage.error(`删除失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

// 处理分页变化
const handlePageChange = (page: number, size: number) => {
  currentPage.value = page;
  pageSize.value = size;
  loadData();
};

// 监听表变化
watch(() => [props.profileId, props.database, props.table], () => {
  currentPage.value = 1;
  filters.value = [];
  orderBy.value = [];
  selectedRows.value = [];
  loadTableSchema();
  loadData();
}, { immediate: false });

// 组件挂载
onMounted(() => {
  loadTableSchema();
  loadData();
});
</script>

<style scoped>
.data-manager {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 16px;
  background-color: #fff;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 8px;
}

.filter-container {
  margin-bottom: 16px;
}

.grid-container {
  flex: 1;
  overflow: auto;
}
</style>
