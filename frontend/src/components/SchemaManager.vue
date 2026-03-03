<template>
  <div class="schema-manager">
    <div class="toolbar">
      <el-button 
        type="primary" 
        :icon="Plus" 
        @click="handleCreateTable"
        :disabled="!databaseStore.currentDatabase"
      >
        创建表
      </el-button>
    </div>

    <SchemaViewer @edit="handleEditTable" />

    <TableEditor
      v-model="editorVisible"
      :schema="currentSchema"
      :mode="editorMode"
      @save="handleSaveTable"
    />

    <!-- 数据丢失警告对话框 -->
    <el-dialog
      v-model="warningVisible"
      title="警告"
      width="500px"
    >
      <el-alert
        type="warning"
        :closable="false"
        show-icon
      >
        <template #title>
          <div style="font-size: 16px; font-weight: 500;">此操作可能导致数据丢失</div>
        </template>
        <div style="margin-top: 12px;">
          <p>以下修改可能会导致数据丢失：</p>
          <ul style="margin: 8px 0; padding-left: 20px;">
            <li v-for="(warning, index) in dataLossWarnings" :key="index">
              {{ warning }}
            </li>
          </ul>
          <p style="margin-top: 12px; color: #e6a23c;">
            <strong>请确认您已备份数据，并了解此操作的后果。</strong>
          </p>
        </div>
      </el-alert>

      <template #footer>
        <el-button @click="warningVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmSaveWithWarning">
          确认修改
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import { useConnectionStore } from '../stores/connection';
import { useDatabaseStore } from '../stores/database';
import { SchemaAPI } from '../api';
import SchemaViewer from './SchemaViewer.vue';
import TableEditor from './TableEditor.vue';
import type { TableSchema, SchemaChange } from '../types/api';

// Stores
const connectionStore = useConnectionStore();
const databaseStore = useDatabaseStore();

// State
const editorVisible = ref(false);
const editorMode = ref<'create' | 'edit'>('create');
const currentSchema = ref<TableSchema | null>(null);
const pendingSchema = ref<TableSchema | null>(null);
const warningVisible = ref(false);
const dataLossWarnings = ref<string[]>([]);

// 创建表
const handleCreateTable = () => {
  currentSchema.value = null;
  editorMode.value = 'create';
  editorVisible.value = true;
};

// 编辑表
const handleEditTable = (schema: TableSchema) => {
  currentSchema.value = schema;
  editorMode.value = 'edit';
  editorVisible.value = true;
};

// 保存表
const handleSaveTable = async (schema: TableSchema) => {
  if (editorMode.value === 'create') {
    await createTable(schema);
  } else {
    await modifyTable(schema);
  }
};

// 创建表
const createTable = async (schema: TableSchema) => {
  if (!connectionStore.currentConnection || !databaseStore.currentDatabase) {
    ElMessage.error('请先选择数据库');
    return;
  }

  try {
    await SchemaAPI.createTable(
      connectionStore.currentConnection.id,
      databaseStore.currentDatabase,
      schema
    );

    ElMessage.success(`表 "${schema.name}" 创建成功`);
    editorVisible.value = false;

    // 刷新表列表
    // TODO: 触发 DatabaseExplorer 刷新
    
    // 选中新创建的表
    databaseStore.setCurrentTable(schema.name);
  } catch (error: any) {
    ElMessage.error(error.message || '创建表失败');
    console.error('Failed to create table:', error);
  }
};

// 修改表
const modifyTable = async (schema: TableSchema) => {
  if (!connectionStore.currentConnection || !databaseStore.currentDatabase || !currentSchema.value) {
    ElMessage.error('缺少必要信息');
    return;
  }

  // 检测可能导致数据丢失的修改
  const warnings = detectDataLossRisks(currentSchema.value, schema);
  
  if (warnings.length > 0) {
    // 显示警告对话框
    dataLossWarnings.value = warnings;
    pendingSchema.value = schema;
    warningVisible.value = true;
    return;
  }

  // 没有风险，直接执行修改
  await executeTableModification(schema);
};

// 确认带警告的保存
const confirmSaveWithWarning = async () => {
  warningVisible.value = false;
  if (pendingSchema.value) {
    await executeTableModification(pendingSchema.value);
    pendingSchema.value = null;
  }
};

// 执行表修改
const executeTableModification = async (schema: TableSchema) => {
  if (!connectionStore.currentConnection || !databaseStore.currentDatabase || !currentSchema.value) {
    return;
  }

  try {
    // 生成修改操作
    const changes = generateSchemaChanges(currentSchema.value, schema);
    
    if (changes.length === 0) {
      ElMessage.info('没有检测到任何修改');
      editorVisible.value = false;
      return;
    }

    await SchemaAPI.alterTable(
      connectionStore.currentConnection.id,
      databaseStore.currentDatabase,
      currentSchema.value.name,
      changes
    );

    ElMessage.success('表结构修改成功');
    editorVisible.value = false;

    // 刷新表结构显示
    // SchemaViewer 会自动重新加载
  } catch (error: any) {
    ElMessage.error(error.message || '修改表结构失败');
    console.error('Failed to alter table:', error);
  }
};

// 检测数据丢失风险
const detectDataLossRisks = (oldSchema: TableSchema, newSchema: TableSchema): string[] => {
  const warnings: string[] = [];

  // 检测删除的列
  const oldColumnNames = oldSchema.columns.map(c => c.name);
  const newColumnNames = newSchema.columns.map(c => c.name);
  const removedColumns = oldColumnNames.filter(name => !newColumnNames.includes(name));
  
  if (removedColumns.length > 0) {
    warnings.push(`删除列: ${removedColumns.join(', ')}`);
  }

  // 检测列类型修改
  for (const newCol of newSchema.columns) {
    const oldCol = oldSchema.columns.find(c => c.name === newCol.name);
    if (oldCol && oldCol.type !== newCol.type) {
      warnings.push(`修改列 "${newCol.name}" 的类型: ${oldCol.type} → ${newCol.type}`);
    }
  }

  // 检测 NOT NULL 约束添加
  for (const newCol of newSchema.columns) {
    const oldCol = oldSchema.columns.find(c => c.name === newCol.name);
    if (oldCol && oldCol.nullable && !newCol.nullable) {
      warnings.push(`列 "${newCol.name}" 添加 NOT NULL 约束`);
    }
  }

  return warnings;
};

// 生成 SchemaChange 列表
const generateSchemaChanges = (oldSchema: TableSchema, newSchema: TableSchema): SchemaChange[] => {
  const changes: SchemaChange[] = [];

  // 表属性修改
  if (oldSchema.engine !== newSchema.engine) {
    changes.push({
      type: 'MODIFY_TABLE_ENGINE',
      target: 'engine',
      definition: newSchema.engine
    });
  }

  if (oldSchema.charset !== newSchema.charset) {
    changes.push({
      type: 'MODIFY_TABLE_CHARSET',
      target: 'charset',
      definition: newSchema.charset
    });
  }

  if (oldSchema.comment !== newSchema.comment) {
    changes.push({
      type: 'MODIFY_TABLE_COMMENT',
      target: 'comment',
      definition: newSchema.comment
    });
  }

  // 列修改
  const oldColumnNames = oldSchema.columns.map(c => c.name);
  const newColumnNames = newSchema.columns.map(c => c.name);

  // 添加的列
  for (const col of newSchema.columns) {
    if (!oldColumnNames.includes(col.name)) {
      changes.push({
        type: 'ADD_COLUMN',
        target: col.name,
        definition: col
      });
    }
  }

  // 删除的列
  for (const col of oldSchema.columns) {
    if (!newColumnNames.includes(col.name)) {
      changes.push({
        type: 'DROP_COLUMN',
        target: col.name
      });
    }
  }

  // 修改的列
  for (const newCol of newSchema.columns) {
    const oldCol = oldSchema.columns.find(c => c.name === newCol.name);
    if (oldCol && JSON.stringify(oldCol) !== JSON.stringify(newCol)) {
      changes.push({
        type: 'MODIFY_COLUMN',
        target: newCol.name,
        definition: newCol
      });
    }
  }

  // 主键修改
  const oldPK = oldSchema.primaryKey?.columns || [];
  const newPK = newSchema.primaryKey?.columns || [];
  if (JSON.stringify(oldPK) !== JSON.stringify(newPK)) {
    if (oldPK.length > 0) {
      changes.push({
        type: 'DROP_PRIMARY_KEY',
        target: 'PRIMARY'
      });
    }
    if (newPK.length > 0) {
      changes.push({
        type: 'ADD_PRIMARY_KEY',
        target: 'PRIMARY',
        definition: newSchema.primaryKey
      });
    }
  }

  // 索引修改
  const oldIndexNames = oldSchema.indexes.map(i => i.name);
  const newIndexNames = newSchema.indexes.map(i => i.name);

  // 添加的索引
  for (const idx of newSchema.indexes) {
    if (!oldIndexNames.includes(idx.name)) {
      changes.push({
        type: 'ADD_INDEX',
        target: idx.name,
        definition: idx
      });
    }
  }

  // 删除的索引
  for (const idx of oldSchema.indexes) {
    if (!newIndexNames.includes(idx.name)) {
      changes.push({
        type: 'DROP_INDEX',
        target: idx.name
      });
    }
  }

  // 修改的索引（先删除再添加）
  for (const newIdx of newSchema.indexes) {
    const oldIdx = oldSchema.indexes.find(i => i.name === newIdx.name);
    if (oldIdx && JSON.stringify(oldIdx) !== JSON.stringify(newIdx)) {
      changes.push({
        type: 'DROP_INDEX',
        target: newIdx.name
      });
      changes.push({
        type: 'ADD_INDEX',
        target: newIdx.name,
        definition: newIdx
      });
    }
  }

  // 外键修改
  const oldFKNames = oldSchema.foreignKeys.map(fk => fk.name);
  const newFKNames = newSchema.foreignKeys.map(fk => fk.name);

  // 添加的外键
  for (const fk of newSchema.foreignKeys) {
    if (!oldFKNames.includes(fk.name)) {
      changes.push({
        type: 'ADD_FOREIGN_KEY',
        target: fk.name,
        definition: fk
      });
    }
  }

  // 删除的外键
  for (const fk of oldSchema.foreignKeys) {
    if (!newFKNames.includes(fk.name)) {
      changes.push({
        type: 'DROP_FOREIGN_KEY',
        target: fk.name
      });
    }
  }

  // 修改的外键（先删除再添加）
  for (const newFK of newSchema.foreignKeys) {
    const oldFK = oldSchema.foreignKeys.find(fk => fk.name === newFK.name);
    if (oldFK && JSON.stringify(oldFK) !== JSON.stringify(newFK)) {
      changes.push({
        type: 'DROP_FOREIGN_KEY',
        target: newFK.name
      });
      changes.push({
        type: 'ADD_FOREIGN_KEY',
        target: newFK.name,
        definition: newFK
      });
    }
  }

  return changes;
};
</script>

<style scoped>
.schema-manager {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.toolbar {
  padding: 16px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}

:deep(.schema-viewer) {
  flex: 1;
  overflow-y: auto;
}
</style>
