<template>
  <div class="table-editor">
    <el-form :model="form" label-width="100px" class="table-editor-form" v-loading="loading">
      <!-- 表基本信息 -->
      <el-card shadow="never" class="section-card">
        <template #header>
          <span class="section-title">表信息</span>
        </template>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="表名" required>
              <el-input 
                v-model="form.name" 
                placeholder="请输入表名"
                :disabled="!isCreateMode"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="引擎">
              <el-select v-model="form.engine" placeholder="选择引擎">
                <el-option label="InnoDB" value="InnoDB" />
                <el-option label="MyISAM" value="MyISAM" />
                <el-option label="Memory" value="Memory" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="字符集">
              <el-select v-model="form.charset" placeholder="选择字符集">
                <el-option label="utf8mb4" value="utf8mb4" />
                <el-option label="utf8" value="utf8" />
                <el-option label="latin1" value="latin1" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="表注释">
          <el-input 
            v-model="form.comment" 
            placeholder="请输入表注释"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
      </el-card>

      <!-- 列定义 -->
      <el-card shadow="never" class="section-card">
        <template #header>
          <div class="section-header">
            <span class="section-title">列定义</span>
            <el-button type="primary" size="small" :icon="Plus" @click="addColumn">
              添加列
            </el-button>
          </div>
        </template>
        <el-table :data="form.columns" border stripe max-height="400">
          <el-table-column label="列名" width="150">
            <template #default="{ row }">
              <el-input v-model="row.name" placeholder="列名" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="数据类型" width="150">
            <template #default="{ row }">
              <el-select 
                v-model="row.baseType" 
                placeholder="类型" 
                size="small"
                @change="handleTypeChange(row)"
              >
                <el-option-group label="整数类型">
                  <el-option label="TINYINT" value="TINYINT" />
                  <el-option label="SMALLINT" value="SMALLINT" />
                  <el-option label="MEDIUMINT" value="MEDIUMINT" />
                  <el-option label="INT" value="INT" />
                  <el-option label="BIGINT" value="BIGINT" />
                </el-option-group>
                <el-option-group label="小数类型">
                  <el-option label="DECIMAL" value="DECIMAL" />
                  <el-option label="FLOAT" value="FLOAT" />
                  <el-option label="DOUBLE" value="DOUBLE" />
                </el-option-group>
                <el-option-group label="字符串类型">
                  <el-option label="CHAR" value="CHAR" />
                  <el-option label="VARCHAR" value="VARCHAR" />
                  <el-option label="TINYTEXT" value="TINYTEXT" />
                  <el-option label="TEXT" value="TEXT" />
                  <el-option label="MEDIUMTEXT" value="MEDIUMTEXT" />
                  <el-option label="LONGTEXT" value="LONGTEXT" />
                </el-option-group>
                <el-option-group label="二进制类型">
                  <el-option label="BINARY" value="BINARY" />
                  <el-option label="VARBINARY" value="VARBINARY" />
                  <el-option label="TINYBLOB" value="TINYBLOB" />
                  <el-option label="BLOB" value="BLOB" />
                  <el-option label="MEDIUMBLOB" value="MEDIUMBLOB" />
                  <el-option label="LONGBLOB" value="LONGBLOB" />
                </el-option-group>
                <el-option-group label="日期时间">
                  <el-option label="DATE" value="DATE" />
                  <el-option label="TIME" value="TIME" />
                  <el-option label="DATETIME" value="DATETIME" />
                  <el-option label="TIMESTAMP" value="TIMESTAMP" />
                  <el-option label="YEAR" value="YEAR" />
                </el-option-group>
                <el-option-group label="其他类型">
                  <el-option label="ENUM" value="ENUM" />
                  <el-option label="SET" value="SET" />
                  <el-option label="JSON" value="JSON" />
                  <el-option label="BOOLEAN" value="BOOLEAN" />
                  <el-option label="BIT" value="BIT" />
                </el-option-group>
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="长度/值" width="120">
            <template #default="{ row }">
              <el-input 
                v-if="needsLength(row.baseType)"
                v-model="row.length" 
                placeholder="长度" 
                size="small"
                @blur="updateFullType(row)"
              />
              <el-input 
                v-else-if="needsValues(row.baseType)"
                v-model="row.enumValues" 
                placeholder="'a','b','c'" 
                size="small"
                @blur="updateFullType(row)"
              />
              <span v-else class="type-no-length">-</span>
            </template>
          </el-table-column>
          <el-table-column label="小数位" width="90">
            <template #default="{ row }">
              <el-input 
                v-if="needsDecimals(row.baseType)"
                v-model="row.decimals" 
                placeholder="小数" 
                size="small"
                @blur="updateFullType(row)"
              />
              <span v-else class="type-no-length">-</span>
            </template>
          </el-table-column>
          <el-table-column label="UNSIGNED" width="100" align="center">
            <template #default="{ row }">
              <el-checkbox 
                v-if="supportsUnsigned(row.baseType)"
                v-model="row.unsigned"
                @change="updateFullType(row)"
              />
              <span v-else class="type-no-length">-</span>
            </template>
          </el-table-column>
          <el-table-column label="允许NULL" width="100" align="center">
            <template #default="{ row }">
              <el-checkbox v-model="row.nullable" />
            </template>
          </el-table-column>
          <el-table-column label="默认值" width="120">
            <template #default="{ row }">
              <el-input v-model="row.defaultValue" placeholder="默认值" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="自增" width="80" align="center">
            <template #default="{ row }">
              <el-checkbox v-model="row.autoIncrement" />
            </template>
          </el-table-column>
          <el-table-column label="主键" width="80" align="center">
            <template #default="{ row, $index }">
              <el-checkbox 
                :model-value="isPrimaryKey($index)" 
                @change="togglePrimaryKey($index)"
              />
            </template>
          </el-table-column>
          <el-table-column label="注释" min-width="150">
            <template #default="{ row }">
              <el-input v-model="row.comment" placeholder="注释" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right">
            <template #default="{ $index }">
              <el-button 
                type="danger" 
                :icon="Delete" 
                size="small" 
                text
                @click="removeColumn($index)"
              />
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 索引定义 -->
      <el-card shadow="never" class="section-card">
        <template #header>
          <div class="section-header">
            <span class="section-title">索引</span>
            <el-button type="primary" size="small" :icon="Plus" @click="addIndex">
              添加索引
            </el-button>
          </div>
        </template>
        <el-table :data="form.indexes" border stripe>
          <el-table-column label="索引名" width="200">
            <template #default="{ row }">
              <el-input v-model="row.name" placeholder="索引名" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="类型" width="150">
            <template #default="{ row }">
              <el-select v-model="row.type" placeholder="类型" size="small">
                <el-option label="INDEX" value="INDEX" />
                <el-option label="UNIQUE" value="UNIQUE" />
                <el-option label="FULLTEXT" value="FULLTEXT" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="列" min-width="300">
            <template #default="{ row }">
              <el-select 
                v-model="row.columns" 
                multiple 
                placeholder="选择列" 
                size="small"
                style="width: 100%"
              >
                <el-option 
                  v-for="col in form.columns" 
                  :key="col.name" 
                  :label="col.name" 
                  :value="col.name"
                  :disabled="!col.name"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right">
            <template #default="{ $index }">
              <el-button 
                type="danger" 
                :icon="Delete" 
                size="small" 
                text
                @click="removeIndex($index)"
              />
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 外键定义 -->
      <el-card shadow="never" class="section-card">
        <template #header>
          <div class="section-header">
            <span class="section-title">外键约束</span>
            <el-button type="primary" size="small" :icon="Plus" @click="addForeignKey">
              添加外键
            </el-button>
          </div>
        </template>
        <el-table :data="form.foreignKeys" border stripe>
          <el-table-column label="约束名" width="180">
            <template #default="{ row }">
              <el-input v-model="row.name" placeholder="约束名" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="列" width="150">
            <template #default="{ row }">
              <el-select 
                v-model="row.columns" 
                multiple 
                placeholder="选择列" 
                size="small"
              >
                <el-option 
                  v-for="col in form.columns" 
                  :key="col.name" 
                  :label="col.name" 
                  :value="col.name"
                  :disabled="!col.name"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="引用表" width="150">
            <template #default="{ row }">
              <el-input v-model="row.referencedTable" placeholder="表名" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="引用列" width="150">
            <template #default="{ row }">
              <el-input v-model="row.referencedColumns" placeholder="列名(逗号分隔)" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="ON DELETE" width="130">
            <template #default="{ row }">
              <el-select v-model="row.onDelete" size="small">
                <el-option label="RESTRICT" value="RESTRICT" />
                <el-option label="CASCADE" value="CASCADE" />
                <el-option label="SET NULL" value="SET NULL" />
                <el-option label="NO ACTION" value="NO ACTION" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="ON UPDATE" width="130">
            <template #default="{ row }">
              <el-select v-model="row.onUpdate" size="small">
                <el-option label="RESTRICT" value="RESTRICT" />
                <el-option label="CASCADE" value="CASCADE" />
                <el-option label="SET NULL" value="SET NULL" />
                <el-option label="NO ACTION" value="NO ACTION" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right">
            <template #default="{ $index }">
              <el-button 
                type="danger" 
                :icon="Delete" 
                size="small" 
                text
                @click="removeForeignKey($index)"
              />
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-form>
    
    <!-- 操作按钮 -->
    <div class="table-editor-footer">
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">
        {{ isCreateMode ? '创建' : '保存' }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Delete } from '@element-plus/icons-vue';
import { SchemaAPI } from '../api';
import type { TableSchema, Column, Index, ForeignKey } from '../types/api';

// Props
interface Props {
  profileId: string;
  database: string;
  table: string;
  mode: 'create' | 'edit';
}

const props = defineProps<Props>();

// Emit
const emit = defineEmits<{
  success: [];
}>();

// State
const loading = ref(false);
const saving = ref(false);
const isCreateMode = computed(() => props.mode === 'create');

// Form data
const form = ref<TableSchema>({
  name: '',
  columns: [],
  primaryKey: { columns: [] },
  indexes: [],
  foreignKeys: [],
  engine: 'InnoDB',
  charset: 'utf8mb4',
  comment: ''
});

// 加载表结构
const loadSchema = async () => {
  if (isCreateMode.value) {
    initForm();
    return;
  }

  if (!props.profileId || !props.database || !props.table) {
    return;
  }

  loading.value = true;
  try {
    const schema = await SchemaAPI.getTableSchema(
      props.profileId,
      props.database,
      props.table
    );
    form.value = JSON.parse(JSON.stringify(schema));
    
    // 确保数组字段存在
    if (!form.value.columns) form.value.columns = [];
    if (!form.value.indexes) form.value.indexes = [];
    if (!form.value.foreignKeys) form.value.foreignKeys = [];
    if (!form.value.primaryKey) form.value.primaryKey = { columns: [] };
    
    // 解析每列的类型字符串
    form.value.columns.forEach((col: any) => {
      const parsed = parseTypeString(col.type);
      col.baseType = parsed.baseType;
      col.length = parsed.length;
      col.decimals = parsed.decimals;
      col.unsigned = parsed.unsigned;
      col.enumValues = parsed.enumValues;
    });
  } catch (error: any) {
    ElMessage.error(error.message || '加载表结构失败');
    console.error('Failed to load schema:', error);
  } finally {
    loading.value = false;
  }
};

// 初始化表单
const initForm = () => {
  form.value = {
    name: props.table || '',
    columns: [],
    primaryKey: { columns: [] },
    indexes: [],
    foreignKeys: [],
    engine: 'InnoDB',
    charset: 'utf8mb4',
    comment: ''
  };
};

// 列操作
const addColumn = () => {
  form.value.columns.push({
    name: '',
    type: 'VARCHAR(255)',
    baseType: 'VARCHAR',
    length: '255',
    decimals: undefined,
    unsigned: false,
    enumValues: undefined,
    nullable: true,
    defaultValue: undefined,
    autoIncrement: false,
    comment: ''
  } as any);
};

// 类型辅助函数
const needsLength = (baseType: string): boolean => {
  const typesNeedingLength = [
    'CHAR', 'VARCHAR', 'BINARY', 'VARBINARY',
    'TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'BIGINT',
    'BIT', 'DECIMAL'
  ];
  return typesNeedingLength.includes(baseType);
};

const needsDecimals = (baseType: string): boolean => {
  return baseType === 'DECIMAL';
};

const needsValues = (baseType: string): boolean => {
  return baseType === 'ENUM' || baseType === 'SET';
};

const supportsUnsigned = (baseType: string): boolean => {
  const unsignedTypes = [
    'TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'BIGINT',
    'DECIMAL', 'FLOAT', 'DOUBLE'
  ];
  return unsignedTypes.includes(baseType);
};

// 处理类型变化
const handleTypeChange = (row: any) => {
  // 设置默认长度
  const defaultLengths: Record<string, string> = {
    'TINYINT': '4',
    'SMALLINT': '6',
    'MEDIUMINT': '9',
    'INT': '11',
    'BIGINT': '20',
    'CHAR': '50',
    'VARCHAR': '255',
    'BINARY': '50',
    'VARBINARY': '255',
    'DECIMAL': '10',
    'BIT': '1'
  };
  
  if (needsLength(row.baseType) && !row.length) {
    row.length = defaultLengths[row.baseType] || '';
  }
  
  if (needsDecimals(row.baseType) && !row.decimals) {
    row.decimals = '2';
  }
  
  if (!needsLength(row.baseType)) {
    row.length = undefined;
  }
  
  if (!needsDecimals(row.baseType)) {
    row.decimals = undefined;
  }
  
  if (!needsValues(row.baseType)) {
    row.enumValues = undefined;
  }
  
  if (!supportsUnsigned(row.baseType)) {
    row.unsigned = false;
  }
  
  updateFullType(row);
};

// 更新完整类型字符串
const updateFullType = (row: any) => {
  let fullType = row.baseType;
  
  if (needsValues(row.baseType) && row.enumValues) {
    fullType = `${row.baseType}(${row.enumValues})`;
  } else if (needsDecimals(row.baseType) && row.length && row.decimals) {
    fullType = `${row.baseType}(${row.length},${row.decimals})`;
  } else if (needsLength(row.baseType) && row.length) {
    fullType = `${row.baseType}(${row.length})`;
  }
  
  if (row.unsigned && supportsUnsigned(row.baseType)) {
    fullType += ' UNSIGNED';
  }
  
  row.type = fullType;
};

// 解析现有类型字符串
const parseTypeString = (typeStr: string): any => {
  const result: any = {
    baseType: '',
    length: undefined,
    decimals: undefined,
    unsigned: false,
    enumValues: undefined
  };
  
  // 检查 UNSIGNED
  if (typeStr.toUpperCase().includes('UNSIGNED')) {
    result.unsigned = true;
    typeStr = typeStr.replace(/\s*UNSIGNED\s*/i, '').trim();
  }
  
  // 解析类型和参数
  const match = typeStr.match(/^(\w+)(?:\(([^)]+)\))?$/);
  if (match) {
    result.baseType = match[1].toUpperCase();
    
    if (match[2]) {
      const params = match[2];
      
      // ENUM/SET 类型
      if (result.baseType === 'ENUM' || result.baseType === 'SET') {
        result.enumValues = params;
      }
      // DECIMAL 类型
      else if (result.baseType === 'DECIMAL' && params.includes(',')) {
        const [len, dec] = params.split(',');
        result.length = len.trim();
        result.decimals = dec.trim();
      }
      // 其他带长度的类型
      else {
        result.length = params.trim();
      }
    }
  }
  
  return result;
};

const removeColumn = (index: number) => {
  const columnName = form.value.columns[index].name;
  form.value.columns.splice(index, 1);
  
  // 从主键中移除
  if (form.value.primaryKey && form.value.primaryKey.columns) {
    form.value.primaryKey.columns = form.value.primaryKey.columns.filter(
      col => col !== columnName
    );
  }
  
  // 从索引中移除
  if (form.value.indexes && Array.isArray(form.value.indexes)) {
    form.value.indexes.forEach(idx => {
      if (idx.columns && Array.isArray(idx.columns)) {
        idx.columns = idx.columns.filter(col => col !== columnName);
      }
    });
  }
  
  // 从外键中移除
  if (form.value.foreignKeys && Array.isArray(form.value.foreignKeys)) {
    form.value.foreignKeys.forEach(fk => {
      if (fk.columns && Array.isArray(fk.columns)) {
        fk.columns = fk.columns.filter(col => col !== columnName);
      }
    });
  }
};

// 主键操作
const isPrimaryKey = (index: number): boolean => {
  const columnName = form.value.columns[index]?.name;
  return form.value.primaryKey?.columns.includes(columnName) || false;
};

const togglePrimaryKey = (index: number) => {
  const columnName = form.value.columns[index].name;
  if (!columnName) {
    ElMessage.warning('请先输入列名');
    return;
  }
  
  if (!form.value.primaryKey) {
    form.value.primaryKey = { columns: [] };
  }
  
  const pkIndex = form.value.primaryKey.columns.indexOf(columnName);
  if (pkIndex > -1) {
    form.value.primaryKey.columns.splice(pkIndex, 1);
  } else {
    form.value.primaryKey.columns.push(columnName);
  }
};

// 索引操作
const addIndex = () => {
  form.value.indexes.push({
    name: '',
    type: 'INDEX',
    columns: []
  });
};

const removeIndex = (index: number) => {
  form.value.indexes.splice(index, 1);
};

// 外键操作
const addForeignKey = () => {
  form.value.foreignKeys.push({
    name: '',
    columns: [],
    referencedTable: '',
    referencedColumns: [],
    onDelete: 'RESTRICT',
    onUpdate: 'RESTRICT'
  });
};

const removeForeignKey = (index: number) => {
  form.value.foreignKeys.splice(index, 1);
};

// 验证表单
const validateForm = (): boolean => {
  if (!form.value.name.trim()) {
    ElMessage.error('请输入表名');
    return false;
  }
  
  if (form.value.columns.length === 0) {
    ElMessage.error('请至少添加一列');
    return false;
  }
  
  // 验证列名不为空
  for (const col of form.value.columns) {
    if (!col.name.trim()) {
      ElMessage.error('列名不能为空');
      return false;
    }
    if (!col.type.trim()) {
      ElMessage.error('列类型不能为空');
      return false;
    }
  }
  
  // 验证索引
  if (form.value.indexes && Array.isArray(form.value.indexes)) {
    for (const idx of form.value.indexes) {
      if (!idx.name.trim()) {
        ElMessage.error('索引名不能为空');
        return false;
      }
      if (idx.columns.length === 0) {
        ElMessage.error(`索引 "${idx.name}" 必须包含至少一列`);
        return false;
      }
    }
  }
  
  // 验证外键
  if (form.value.foreignKeys && Array.isArray(form.value.foreignKeys)) {
    for (const fk of form.value.foreignKeys) {
      if (!fk.name.trim()) {
        ElMessage.error('外键约束名不能为空');
        return false;
      }
      if (fk.columns.length === 0) {
        ElMessage.error(`外键 "${fk.name}" 必须包含至少一列`);
        return false;
      }
      if (!fk.referencedTable.trim()) {
        ElMessage.error(`外键 "${fk.name}" 必须指定引用表`);
        return false;
      }
    }
  }
  
  return true;
};

// 保存
const handleSave = async () => {
  if (!validateForm()) {
    return;
  }

  // 确保所有列的 type 字段都是最新的
  form.value.columns.forEach((col: any) => {
    // 如果 baseType 存在但 type 为空或不完整，重新生成
    if (col.baseType && (!col.type || col.type === col.baseType)) {
      updateFullType(col);
    }
    // 如果 type 存在但 baseType 为空，解析 type
    else if (col.type && !col.baseType) {
      const parsed = parseTypeString(col.type);
      col.baseType = parsed.baseType;
      col.length = parsed.length;
      col.decimals = parsed.decimals;
      col.unsigned = parsed.unsigned;
      col.enumValues = parsed.enumValues;
    }
  });

  saving.value = true;
  try {
    if (isCreateMode.value) {
      await SchemaAPI.createTable(props.profileId, props.database, form.value);
      ElMessage.success('创建表成功');
    } else {
      await SchemaAPI.alterTable(props.profileId, props.database, props.table, form.value);
      ElMessage.success('修改表结构成功');
    }
    emit('success');
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败');
    console.error('Failed to save schema:', error);
  } finally {
    saving.value = false;
  }
};

// 关闭
const handleClose = () => {
  // Dialog is managed by parent component through contentType
};

// 组件挂载时加载数据
onMounted(() => {
  loadSchema();
});

// 监听 props 变化
watch(() => [props.profileId, props.database, props.table, props.mode], () => {
  loadSchema();
}, { immediate: false });
</script>

<style scoped>
.table-editor {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 20px;
  background-color: #fff;
}

.table-editor-form {
  flex: 1;
  overflow-y: auto;
  padding-right: 10px;
}

.table-editor-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 0 0 0;
  border-top: 1px solid #e4e7ed;
  margin-top: 16px;
}

.section-card {
  margin-bottom: 16px;
}

.section-card:last-child {
  margin-bottom: 0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

:deep(.el-card__header) {
  padding: 12px 16px;
  background-color: #f5f7fa;
}

:deep(.el-card__body) {
  padding: 16px;
}

:deep(.el-table) {
  font-size: 13px;
}

:deep(.el-form-item) {
  margin-bottom: 16px;
}

.type-no-length {
  color: #c0c4cc;
  font-size: 12px;
  display: inline-block;
  text-align: center;
  width: 100%;
}
</style>
