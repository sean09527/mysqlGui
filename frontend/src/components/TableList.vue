<template>
  <div class="table-list">
    <div class="table-list-header">
      <h3>数据表列表</h3>
      <el-button 
        :icon="Refresh" 
        circle 
        size="small" 
        @click="refreshTables"
        :loading="loading"
        title="刷新"
      />
    </div>

    <div class="table-list-filter">
      <el-input
        v-model="filterText"
        placeholder="输入表名进行筛选..."
        clearable
        :prefix-icon="Search"
        @input="handleFilterChange"
      />
      <span v-if="filteredTables.length !== tables.length" class="filter-count">
        显示 {{ filteredTables.length }} / {{ tables.length }} 个表
      </span>
    </div>

    <div class="table-list-content">
      <el-table
        v-if="filteredTables.length > 0"
        v-loading="loading"
        :data="filteredTables"
        stripe
        style="width: 100%"
        :height="tableHeight"
      >
        <el-table-column prop="name" label="表名" min-width="150" />
        <el-table-column prop="engine" label="引擎" width="100" />
        <el-table-column prop="collation" label="排序字符集" width="150" />
        <el-table-column prop="rows" label="数据行数" width="120" align="right">
          <template #default="{ row }">
            {{ formatNumber(row.rows) }}
          </template>
        </el-table-column>
        <el-table-column label="大小" width="100" align="right">
          <template #default="{ row }">
            {{ formatSize(row.dataLength + row.indexLength) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleViewData(row)"
            >
              查看数据
            </el-button>
            <el-button
              type="info"
              size="small"
              @click="handleViewSchema(row)"
            >
              查看结构
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty 
        v-else-if="!loading && !currentDatabase"
        description="请从左侧选择数据库"
      />

      <el-empty 
        v-else-if="!loading && tables.length === 0"
        description="该数据库没有数据表"
      />

      <el-empty 
        v-else-if="!loading && filterText && filteredTables.length === 0"
        description="没有匹配的数据表"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { Refresh, Search } from '@element-plus/icons-vue';
import { useConnectionStore } from '../stores/connection';
import { useDatabaseStore } from '../stores/database';
import { DatabaseAPI } from '../api';
import type { Table } from '../types/api';

// Emits
const emit = defineEmits<{
  viewSchema: [table: string];
  viewData: [table: string];
}>();

// Stores
const connectionStore = useConnectionStore();
const databaseStore = useDatabaseStore();

// State
const loading = ref(false);
const tables = ref<Table[]>([]);
const tableHeight = ref(600);
const filterText = ref('');

// Computed
const currentDatabase = computed(() => databaseStore.currentDatabase);

// 过滤后的表列表
const filteredTables = computed(() => {
  if (!filterText.value) {
    return tables.value;
  }
  
  const searchText = filterText.value.toLowerCase();
  return tables.value.filter(table => 
    table.name.toLowerCase().includes(searchText)
  );
});

// 处理筛选变化
const handleFilterChange = () => {
  // 筛选是响应式的，这里可以添加额外的逻辑
  // 例如：记录筛选历史、统计等
};

// 格式化数字
const formatNumber = (num: number): string => {
  if (num >= 1000000) {
    return `${(num / 1000000).toFixed(2)}M`;
  } else if (num >= 1000) {
    return `${(num / 1000).toFixed(2)}K`;
  }
  return num.toString();
};

// 格式化大小
const formatSize = (bytes: number): string => {
  if (bytes >= 1073741824) {
    return `${(bytes / 1073741824).toFixed(2)} GB`;
  } else if (bytes >= 1048576) {
    return `${(bytes / 1048576).toFixed(2)} MB`;
  } else if (bytes >= 1024) {
    return `${(bytes / 1024).toFixed(2)} KB`;
  }
  return `${bytes} B`;
};

// 加载表列表
const loadTables = async () => {
  if (!connectionStore.currentConnection || !currentDatabase.value) {
    tables.value = [];
    return;
  }

  loading.value = true;
  try {
    const result = await DatabaseAPI.listTables(
      connectionStore.currentConnection.id,
      currentDatabase.value
    );
    tables.value = result || [];
    databaseStore.setTables(result || []);
  } catch (error: any) {
    ElMessage.error(error.message || '加载表列表失败');
    console.error('Failed to load tables:', error);
    tables.value = [];
  } finally {
    loading.value = false;
  }
};

// 刷新表列表
const refreshTables = async () => {
  await loadTables();
  ElMessage.success('表列表已刷新');
};

// 查看表数据
const handleViewData = (table: Table) => {
  databaseStore.setCurrentTable(table.name);
  emit('viewData', table.name);
};

// 查看表结构
const handleViewSchema = (table: Table) => {
  databaseStore.setCurrentTable(table.name);
  emit('viewSchema', table.name);
};

// 监听当前数据库变化
watch(currentDatabase, async (newDb) => {
  if (newDb) {
    filterText.value = ''; // 切换数据库时清空筛选
    await loadTables();
  } else {
    tables.value = [];
    filterText.value = '';
  }
});

// 组件挂载时加载数据
onMounted(async () => {
  if (currentDatabase.value) {
    await loadTables();
  }
  
  // 计算表格高度
  const updateHeight = () => {
    const windowHeight = window.innerHeight;
    tableHeight.value = windowHeight - 200; // 减去头部和其他元素的高度
  };
  
  updateHeight();
  window.addEventListener('resize', updateHeight);
  
  // 清理
  return () => {
    window.removeEventListener('resize', updateHeight);
  };
});
</script>

<style scoped>
.table-list {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 4px;
  overflow: hidden;
}

.table-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.table-list-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.table-list-filter {
  padding: 0 16px 16px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-count {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
}

.table-list-content {
  flex: 1;
  overflow: hidden;
  padding: 0 16px 16px 16px;
}

:deep(.el-table) {
  font-size: 13px;
}

:deep(.el-table th) {
  background-color: #fafafa;
  font-weight: 600;
}

:deep(.el-table td) {
  padding: 8px 0;
}
</style>
