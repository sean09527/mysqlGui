<template>
  <div class="schema-viewer">
    <el-card v-loading="loading" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="table-title">
            <el-icon><Document /></el-icon>
            {{ schema?.name || '表结构' }}
          </span>
          
        </div>
      </template>

      <el-empty v-if="!schema && !loading" description="请选择一个表查看结构" />

      <div v-else-if="schema" class="schema-content">
        <!-- 表信息 -->
        <div class="info-section">
          <h4>表信息</h4>
          <el-descriptions :column="3" border>
            <el-descriptions-item label="表名">{{ schema.name }}</el-descriptions-item>
            <el-descriptions-item label="引擎">{{ schema.engine }}</el-descriptions-item>
            <el-descriptions-item label="字符集">{{ schema.charset }}</el-descriptions-item>
            <el-descriptions-item label="注释" :span="3">
              {{ schema.comment || '无' }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 列信息 -->
        <div class="info-section">
          <h4>列信息</h4>
          <el-table :data="schema.columns" border stripe>
            <el-table-column prop="name" label="列名" width="180" />
            <el-table-column prop="type" label="数据类型" width="150" />
            <el-table-column label="允许NULL" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.nullable ? 'info' : 'success'" size="small">
                  {{ row.nullable ? '是' : '否' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="默认值" width="150">
              <template #default="{ row }">
                <span v-if="row.defaultValue !== undefined && row.defaultValue !== null">
                  {{ row.defaultValue }}
                </span>
                <span v-else class="text-muted">无</span>
              </template>
            </el-table-column>
            <el-table-column label="自增" width="80" align="center">
              <template #default="{ row }">
                <el-icon v-if="row.autoIncrement" color="#67c23a"><Check /></el-icon>
              </template>
            </el-table-column>
            <el-table-column prop="comment" label="注释" min-width="200" show-overflow-tooltip />
          </el-table>
        </div>

        <!-- 主键信息 -->
        <div v-if="schema.primaryKey && schema.primaryKey.columns.length > 0" class="info-section">
          <h4>主键</h4>
          <el-tag v-for="col in schema.primaryKey.columns" :key="col" type="danger" class="key-tag">
            {{ col }}
          </el-tag>
        </div>

        <!-- 索引信息 -->
        <div v-if="schema.indexes && schema.indexes.length > 0" class="info-section">
          <h4>索引</h4>
          <el-table :data="schema.indexes" border stripe>
            <el-table-column prop="name" label="索引名" width="200" />
            <el-table-column prop="type" label="类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getIndexTypeTag(row.type)" size="small">
                  {{ row.type }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="列" min-width="300">
              <template #default="{ row }">
                <el-tag v-for="col in row.columns" :key="col" class="column-tag">
                  {{ col }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 外键信息 -->
        <div v-if="schema.foreignKeys && schema.foreignKeys.length > 0" class="info-section">
          <h4>外键约束</h4>
          <el-table :data="schema.foreignKeys" border stripe>
            <el-table-column prop="name" label="约束名" width="200" />
            <el-table-column label="列" width="150">
              <template #default="{ row }">
                <el-tag v-for="col in row.columns" :key="col" size="small" class="column-tag">
                  {{ col }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="referencedTable" label="引用表" width="150" />
            <el-table-column label="引用列" width="150">
              <template #default="{ row }">
                <el-tag v-for="col in row.referencedColumns" :key="col" size="small" class="column-tag">
                  {{ col }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="onDelete" label="ON DELETE" width="120" />
            <el-table-column prop="onUpdate" label="ON UPDATE" width="120" />
          </el-table>
        </div>

        <!-- DDL 语句 -->
        <div class="info-section">
          <h4>
            CREATE TABLE DDL
            <el-button 
              :icon="CopyDocument" 
              size="small" 
              text 
              @click="copyDDL"
              style="margin-left: 10px"
            >
              复制
            </el-button>
          </h4>
          <el-input
            v-model="ddl"
            type="textarea"
            :rows="10"
            readonly
            class="ddl-textarea"
          />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { 
  Document, 
  Refresh, 
  Edit, 
  Check, 
  CopyDocument 
} from '@element-plus/icons-vue';
import { useConnectionStore } from '../stores/connection';
import { useDatabaseStore } from '../stores/database';
import { SchemaAPI } from '../api';
import type { TableSchema } from '../types/api';

// Props
interface Props {
  profileId: string;
  database: string;
  table: string;
}

const props = defineProps<Props>();

// Stores (保留用于其他用途)
const connectionStore = useConnectionStore();
const databaseStore = useDatabaseStore();

// State
const loading = ref(false);
const schema = ref<TableSchema | null>(null);
const ddl = ref('');

// Emit
const emit = defineEmits<{
  edit: [schema: TableSchema]
}>();

// 获取索引类型标签颜色
const getIndexTypeTag = (type: string) => {
  switch (type.toUpperCase()) {
    case 'PRIMARY':
      return 'danger';
    case 'UNIQUE':
      return 'warning';
    case 'FULLTEXT':
      return 'success';
    default:
      return 'info';
  }
};

// 加载表结构
const loadSchema = async () => {
  if (!props.profileId || !props.database || !props.table) {
    schema.value = null;
    ddl.value = '';
    return;
  }

  loading.value = true;
  try {
    // 加载表结构
    const tableSchema = await SchemaAPI.getTableSchema(
      props.profileId,
      props.database,
      props.table
    );
    schema.value = tableSchema;

    // 加载 DDL
    const tableDDL = await SchemaAPI.getCreateTableDDL(
      props.profileId,
      props.database,
      props.table
    );
    ddl.value = tableDDL;
  } catch (error: any) {
    ElMessage.error(error.message || '加载表结构失败');
    console.error('Failed to load schema:', error);
    schema.value = null;
    ddl.value = '';
  } finally {
    loading.value = false;
  }
};

// 处理编辑
const handleEdit = () => {
  if (schema.value) {
    emit('edit', schema.value);
  }
};

// 复制 DDL
const copyDDL = async () => {
  try {
    await navigator.clipboard.writeText(ddl.value);
    ElMessage.success('DDL 已复制到剪贴板');
  } catch (error) {
    ElMessage.error('复制失败');
  }
};

// 组件挂载时加载数据
onMounted(() => {
  loadSchema();
});

// 监听 props 变化
watch(() => [props.profileId, props.database, props.table], () => {
  loadSchema();
}, { immediate: false });

// 监听当前表变化
watch(
  () => [databaseStore.currentDatabase, databaseStore.currentTable],
  () => {
    loadSchema();
  },
  { immediate: true }
);
</script>

<style scoped>
.schema-viewer {
  height: 100%;
  overflow-y: auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.schema-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.info-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 500;
  color: #606266;
  display: flex;
  align-items: center;
}

.text-muted {
  color: #909399;
  font-style: italic;
}

.key-tag,
.column-tag {
  margin-right: 8px;
  margin-bottom: 4px;
}

.ddl-textarea {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

:deep(.el-card__body) {
  padding: 20px;
}

:deep(.el-descriptions__label) {
  font-weight: 500;
}
</style>
