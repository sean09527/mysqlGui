<template>
  <div class="data-grid">
    <!-- 使用虚拟滚动表格处理大数据集 -->
    <div v-if="useVirtualScroll" class="virtual-table-wrapper">
      <el-auto-resizer>
        <template #default="{ height, width }">
          <el-table-v2
            :columns="virtualColumns"
            :data="displayData"
            :width="width"
            :height="height - (showPagination ? 60 : 0)"
            :row-height="42"
            :header-height="42"
            :estimated-row-height="42"
            fixed
            v-loading="loading"
          />
        </template>
      </el-auto-resizer>
    </div>

    <!-- 传统表格用于小数据集 -->
    <el-table
      v-else
      :data="displayData"
      border
      stripe
      highlight-current-row
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
      @cell-dblclick="handleCellDoubleClick"
      style="width: 100%"
      v-loading="loading"
    >
      <!-- 选择列 -->
      <el-table-column type="selection" width="55" />
      
      <!-- 数据列 -->
      <el-table-column
        v-for="column in columns"
        :key="column"
        :prop="column"
        :label="column"
        :sortable="sortable ? 'custom' : false"
        min-width="120"
      >
        <template #default="{ row, $index }">
          <div
            v-if="editingCell.row === $index && editingCell.column === column"
            class="cell-editor-wrapper"
            @click.stop
          >
            <CellEditor
              v-if="columnSchemas && columnSchemas[column]"
              :value="editingValue"
              :column="columnSchemas[column]"
              :foreign-keys="foreignKeys"
              :profile-id="profileId"
              :database="database"
              @change="handleCellEditComplete"
              @cancel="handleCellEditCancel"
            />
            <el-input
              v-else
              v-model="editingValue"
              size="small"
              @blur="handleCellEditComplete"
              @keyup.enter="handleCellEditComplete"
              @keyup.esc="handleCellEditCancel"
              ref="cellInput"
            />
          </div>
          <div 
            v-else 
            class="cell-content"
            @dblclick="handleCellContentDoubleClick($index, column, row[column])"
          >
            {{ formatCellValue(row[column]) }}
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container" v-if="showPagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[50, 100, 200, 500, 1000, 5000]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handlePageSizeChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, h } from 'vue';
import { ElCheckbox } from 'element-plus';
import CellEditor from './CellEditor.vue';
import type { OrderBy, Column, ForeignKey } from '../types/api';

interface Props {
  data?: any[][];
  columns?: string[];
  columnSchemas?: Record<string, Column>;
  foreignKeys?: ForeignKey[];
  profileId?: string;
  database?: string;
  total?: number;
  loading?: boolean;
  sortable?: boolean;
  editable?: boolean;
  showPagination?: boolean;
  pageSize?: number;
  virtualScrollThreshold?: number; // 超过此行数启用虚拟滚动
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  columns: () => [],
  total: 0,
  loading: false,
  sortable: true,
  editable: true,
  showPagination: true,
  pageSize: 100,
  columnSchemas: () => ({}),
  foreignKeys: () => [],
  profileId: '',
  database: '',
  virtualScrollThreshold: 500, // 默认超过500行启用虚拟滚动
});

const emit = defineEmits<{
  selectionChange: [rows: any[]];
  sortChange: [orderBy: OrderBy[]];
  cellEdit: [rowIndex: number, column: string, oldValue: any, newValue: any];
  pageChange: [page: number, pageSize: number];
}>();

// 当前页码
const currentPage = ref(1);
const pageSize = ref(props.pageSize);

// 选中的行
const selectedRows = ref<any[]>([]);
const selectedRowIndices = ref<number[]>([]);

// 编辑状态
const editingCell = ref<{ row: number; column: string }>({ row: -1, column: '' });
const editingValue = ref<any>(null);
const cellInput = ref<any>(null);

// 是否使用虚拟滚动（基于数据量）
const useVirtualScroll = computed(() => {
  return props.data.length > props.virtualScrollThreshold;
});

// 调试：输出 props
console.log('=== DataGrid Props ===');
console.log('editable:', props.editable);
console.log('columns:', props.columns);
console.log('columnSchemas keys:', Object.keys(props.columnSchemas || {}));
console.log('data rows:', props.data?.length);
console.log('useVirtualScroll:', useVirtualScroll.value);

// 将二维数组转换为对象数组以供 el-table 使用
const displayData = computed(() => {
  if (!props.data || !Array.isArray(props.data)) {
    return [];
  }
  return props.data.map((row, rowIndex) => {
    const obj: Record<string, any> = { _rowIndex: rowIndex };
    props.columns.forEach((col, index) => {
      obj[col] = row[index];
    });
    return obj;
  });
});

// 虚拟表格列配置
const virtualColumns = computed(() => {
  const cols: any[] = [];
  
  // 选择列
  cols.push({
    key: 'selection',
    dataKey: 'selection',
    title: '',
    width: 55,
    cellRenderer: ({ rowData, rowIndex }: any) => {
      const isSelected = selectedRowIndices.value.indexOf(rowIndex) >= 0;
      return h(ElCheckbox, {
        modelValue: isSelected,
        'onUpdate:modelValue': (checked: boolean) => handleVirtualRowSelection(rowIndex, checked, rowData),
      });
    },
    headerCellRenderer: () => {
      const allSelected = displayData.value.length > 0 && 
        selectedRowIndices.value.length === displayData.value.length;
      return h(ElCheckbox, {
        modelValue: allSelected,
        'onUpdate:modelValue': (checked: boolean) => handleVirtualSelectAll(checked),
      });
    },
  });
  
  // 数据列
  props.columns.forEach((column) => {
    cols.push({
      key: column,
      dataKey: column,
      title: column,
      width: 150,
      cellRenderer: ({ rowData, rowIndex }: any) => {
        const value = rowData[column];
        const formattedValue = formatCellValue(value);
        
        return h('div', {
          class: 'virtual-cell-content',
          title: formattedValue, // 添加 tooltip
        }, formattedValue);
      },
    });
  });
  
  return cols;
});

// 处理虚拟表格行选择
const handleVirtualRowSelection = (rowIndex: number, checked: boolean, rowData: any) => {
  const index = selectedRowIndices.value.indexOf(rowIndex);
  if (checked && index < 0) {
    selectedRowIndices.value.push(rowIndex);
  } else if (!checked && index >= 0) {
    selectedRowIndices.value.splice(index, 1);
  }
  
  // 更新选中的行数据
  const selected = selectedRowIndices.value.map(idx => displayData.value[idx]);
  selectedRows.value = selected;
  emit('selectionChange', selected);
};

// 处理虚拟表格全选
const handleVirtualSelectAll = (checked: boolean) => {
  if (checked) {
    selectedRowIndices.value = displayData.value.map((_, index) => index);
    selectedRows.value = [...displayData.value];
  } else {
    selectedRowIndices.value = [];
    selectedRows.value = [];
  }
  emit('selectionChange', selectedRows.value);
};

// 处理选择变化
const handleSelectionChange = (rows: any[]) => {
  selectedRows.value = rows;
  emit('selectionChange', rows);
};

// 处理排序变化
const handleSortChange = ({ column, prop, order }: any) => {
  if (!prop || !order) {
    emit('sortChange', []);
    return;
  }

  const orderBy: OrderBy = {
    column: prop,
    direction: order === 'ascending' ? 'ASC' : 'DESC',
  };
  emit('sortChange', [orderBy]);
};

// 处理单元格双击（备用方法 - 直接在 cell-content 上监听）
const handleCellContentDoubleClick = (rowIndex: number, columnName: string, value: any) => {
  console.log('=== Cell Content Double-Click ===');
  console.log('Editable:', props.editable);
  console.log('Row index:', rowIndex);
  console.log('Column:', columnName);
  console.log('Value:', value);
  
  if (!props.editable) {
    console.log('❌ Editing disabled');
    return;
  }
  
  editingCell.value = { row: rowIndex, column: columnName };
  editingValue.value = value;
  
  console.log('✅ Started editing:', editingCell.value);
  
  // 聚焦输入框
  nextTick(() => {
    if (cellInput.value) {
      if (Array.isArray(cellInput.value)) {
        cellInput.value[0]?.focus();
      } else {
        cellInput.value.focus();
      }
      console.log('✅ Input focused');
    }
  });
};

// 处理单元格双击
const handleCellDoubleClick = (row: any, column: any, cell: any, event: any) => {
  console.log('=== Cell Double-Click Event ===');
  console.log('1. Editable prop:', props.editable);
  console.log('2. Row data:', row);
  console.log('3. Column object:', column);
  console.log('4. Column property:', column?.property);
  console.log('5. Available columnSchemas:', Object.keys(props.columnSchemas || {}));
  console.log('6. Display data length:', displayData.value.length);
  
  if (!props.editable) {
    console.log('❌ Editing disabled - editable prop is false');
    return;
  }
  
  const rowIndex = displayData.value.indexOf(row);
  const columnName = column.property;
  
  console.log('7. Calculated row index:', rowIndex);
  console.log('8. Column name:', columnName);
  
  if (rowIndex >= 0 && columnName) {
    editingCell.value = { row: rowIndex, column: columnName };
    editingValue.value = row[columnName];
    
    console.log('✅ Editing cell:', editingCell.value);
    console.log('✅ Initial value:', editingValue.value);
    
    // 聚焦输入框
    nextTick(() => {
      if (cellInput.value) {
        if (Array.isArray(cellInput.value)) {
          cellInput.value[0]?.focus();
        } else {
          cellInput.value.focus();
        }
        console.log('✅ Input focused');
      } else {
        console.log('⚠️ cellInput ref not found');
      }
    });
  } else {
    console.log('❌ Invalid row or column - rowIndex:', rowIndex, 'columnName:', columnName);
  }
};

// 完成单元格编辑
const handleCellEditComplete = (newValue?: any) => {
  const { row, column } = editingCell.value;
  if (row >= 0 && column) {
    const oldValue = displayData.value[row][column];
    const finalValue = newValue !== undefined ? newValue : editingValue.value;
    
    if (oldValue !== finalValue) {
      emit('cellEdit', row, column, oldValue, finalValue);
    }
  }
  
  editingCell.value = { row: -1, column: '' };
  editingValue.value = null;
};

// 取消单元格编辑
const handleCellEditCancel = () => {
  editingCell.value = { row: -1, column: '' };
  editingValue.value = null;
};

// 格式化单元格值
const formatCellValue = (value: any): string => {
  if (value === null || value === undefined) {
    return 'NULL';
  }
  if (typeof value === 'boolean') {
    return value ? 'true' : 'false';
  }
  if (typeof value === 'object') {
    return JSON.stringify(value);
  }
  return String(value);
};

// 处理页码变化
const handlePageChange = (page: number) => {
  emit('pageChange', page, pageSize.value);
};

// 处理每页大小变化
const handlePageSizeChange = (size: number) => {
  currentPage.value = 1;
  emit('pageChange', 1, size);
};

// 获取选中的行
const getSelectedRows = () => {
  return selectedRows.value;
};

// 清除选择
const clearSelection = () => {
  selectedRows.value = [];
};

// 暴露方法给父组件
defineExpose({
  getSelectedRows,
  clearSelection,
});
</script>

<style scoped>
.data-grid {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.virtual-table-wrapper {
  flex: 1;
  min-height: 400px;
  overflow: hidden;
}

.virtual-cell-content {
  padding: 8px 12px;
  min-height: 20px;
  cursor: default;
  user-select: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 26px;
}

.cell-content {
  padding: 4px;
  min-height: 20px;
  cursor: pointer;
  user-select: none;
}

.cell-content:hover {
  background-color: #f5f7fa;
}

.cell-editor-wrapper {
  width: 100%;
}

.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-table__cell) {
  padding: 8px 0;
}

/* 虚拟表格样式优化 */
:deep(.el-table-v2__header-row) {
  background-color: #f5f7fa;
  font-weight: 600;
}

:deep(.el-table-v2__row) {
  border-bottom: 1px solid #ebeef5;
}

:deep(.el-table-v2__row:hover) {
  background-color: #f5f7fa;
}

:deep(.el-table-v2__cell) {
  padding: 0;
  border-right: 1px solid #ebeef5;
}

:deep(.el-table-v2__header-cell) {
  padding: 0 12px;
  border-right: 1px solid #ebeef5;
}
</style>
