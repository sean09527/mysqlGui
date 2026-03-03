<template>
  <el-dialog
    v-model="visible"
    title="导入数据"
    width="700px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form :model="form" label-width="100px" size="default">
      <!-- 文件选择 -->
      <el-form-item label="选择文件">
        <el-input
          v-model="form.filePath"
          placeholder="点击浏览按钮选择文件"
          readonly
        >
          <template #append>
            <el-button @click="selectFile">浏览</el-button>
          </template>
        </el-input>
      </el-form-item>

      <!-- 文件格式 -->
      <el-form-item v-if="form.filePath" label="文件格式">
        <el-tag :type="formatTagType">{{ detectedFormat || '未知' }}</el-tag>
        <el-text v-if="formatValidated" type="success" size="small" style="margin-left: 12px;">
          <el-icon><CircleCheck /></el-icon>
          格式验证通过
        </el-text>
        <el-text v-else-if="formatError" type="danger" size="small" style="margin-left: 12px;">
          <el-icon><CircleClose /></el-icon>
          {{ formatError }}
        </el-text>
      </el-form-item>

      <!-- 列映射配置 (CSV 和 JSON) -->
      <el-form-item
        v-if="showMapping && tableColumns.length > 0"
        label="列映射"
        class="mapping-item"
      >
        <div class="mapping-container">
          <div class="mapping-header">
            <span class="mapping-col">文件列</span>
            <span class="mapping-arrow"></span>
            <span class="mapping-col">表列</span>
          </div>
          <div
            v-for="(fileCol, index) in fileColumns"
            :key="index"
            class="mapping-row"
          >
            <el-input
              v-model="fileColumns[index]"
              size="small"
              placeholder="文件列名"
              class="mapping-input"
            />
            <el-icon class="mapping-arrow-icon"><Right /></el-icon>
            <el-select
              v-model="form.mapping.TableColumns[index]"
              size="small"
              placeholder="选择表列"
              class="mapping-select"
              clearable
            >
              <el-option
                v-for="col in tableColumns"
                :key="col.name"
                :label="`${col.name} (${col.type})`"
                :value="col.name"
              />
            </el-select>
            <el-button
              size="small"
              type="danger"
              :icon="Delete"
              circle
              @click="removeMapping(index)"
            />
          </div>
          <el-button
            size="small"
            type="primary"
            :icon="Plus"
            @click="addMapping"
            style="margin-top: 8px;"
          >
            添加映射
          </el-button>
        </div>
      </el-form-item>

      <!-- 导入进度 -->
      <el-form-item v-if="importing" label="导入进度">
        <el-progress
          :percentage="progress.percentage"
          :status="progress.status"
        />
        <div class="progress-info">
          <el-text size="small">
            已导入 {{ progress.current }} / {{ progress.total }} 行
          </el-text>
        </div>
      </el-form-item>

      <!-- 导入结果 -->
      <el-form-item v-if="importResult" label="导入结果">
        <div class="result-container">
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="总行数">
              {{ importResult.TotalRows }}
            </el-descriptions-item>
            <el-descriptions-item label="成功行数">
              <el-text type="success">{{ importResult.SuccessRows }}</el-text>
            </el-descriptions-item>
            <el-descriptions-item label="失败行数">
              <el-text :type="importResult.FailedRows > 0 ? 'danger' : 'info'">
                {{ importResult.FailedRows }}
              </el-text>
            </el-descriptions-item>
          </el-descriptions>

          <!-- 错误详情 -->
          <div v-if="importResult.Errors && importResult.Errors.length > 0" class="error-details">
            <el-divider content-position="left">错误详情</el-divider>
            <el-scrollbar max-height="200px">
              <div
                v-for="(error, index) in importResult.Errors"
                :key="index"
                class="error-item"
              >
                <el-text type="danger" size="small">
                  第 {{ error.Row }} 行: {{ error.Message }}
                </el-text>
              </div>
            </el-scrollbar>
          </div>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose" :disabled="importing">
          {{ importResult ? '关闭' : '取消' }}
        </el-button>
        <el-button
          v-if="!importResult"
          type="primary"
          @click="handleImport"
          :loading="importing"
          :disabled="!canImport"
        >
          {{ importing ? '导入中...' : '开始导入' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { CircleCheck, CircleClose, Right, Delete, Plus } from '@element-plus/icons-vue';
import {
  ImportFromSQL,
  ImportFromCSV,
  ImportFromJSON,
  ValidateCSVFormat,
  ValidateJSONFormat,
  ValidateSQLFormat,
  OpenFileDialog,
} from '../../wailsjs/go/backend/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { backend, importexport } from '../../wailsjs/go/models';
import type { Column } from '../types/api';

interface Props {
  modelValue: boolean;
  profileId: string;
  database: string;
  table: string;
  tableColumns: Column[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'success': [];
}>();

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});

const form = ref({
  filePath: '',
  mapping: {
    FileColumns: [] as string[],
    TableColumns: [] as string[],
  },
});

const fileColumns = ref<string[]>([]);
const detectedFormat = ref<'SQL' | 'CSV' | 'JSON' | null>(null);
const formatValidated = ref(false);
const formatError = ref('');
const importing = ref(false);
const importResult = ref<importexport.ImportResult | null>(null);

const progress = ref({
  current: 0,
  total: 0,
  percentage: 0,
  status: undefined as 'success' | 'exception' | undefined,
});

const formatTagType = computed(() => {
  if (!detectedFormat.value) return 'info';
  if (formatValidated.value) return 'success';
  if (formatError.value) return 'danger';
  return 'warning';
});

const showMapping = computed(() => {
  return (
    detectedFormat.value === 'CSV' ||
    detectedFormat.value === 'JSON'
  ) && formatValidated.value;
});

const canImport = computed(() => {
  if (!form.value.filePath || !formatValidated.value) {
    return false;
  }
  
  // SQL 文件不需要映射
  if (detectedFormat.value === 'SQL') {
    return true;
  }
  
  // CSV 和 JSON 需要至少一个映射
  return form.value.mapping.TableColumns.some(col => col);
});

// 选择文件
const selectFile = async () => {
  try {
    const filters: backend.FileDialogFilter[] = [
      backend.FileDialogFilter.createFrom({
        displayName: 'SQL 文件',
        pattern: '*.sql',
      }),
      backend.FileDialogFilter.createFrom({
        displayName: 'CSV 文件',
        pattern: '*.csv',
      }),
      backend.FileDialogFilter.createFrom({
        displayName: 'JSON 文件',
        pattern: '*.json',
      }),
      backend.FileDialogFilter.createFrom({
        displayName: '所有文件',
        pattern: '*',
      }),
    ];
    
    // 使用后端 API 打开文件选择对话框
    const path = await OpenFileDialog('选择导入文件', filters);
    
    if (path) {
      form.value.filePath = path;
      detectFormat(path);
    }
  } catch (error: any) {
    console.error('Failed to select file:', error);
    ElMessage.error(`选择文件失败: ${error.message || error}`);
  }
};

// 检测文件格式
const detectFormat = async (filePath: string) => {
  formatValidated.value = false;
  formatError.value = '';
  detectedFormat.value = null;
  
  // 根据文件扩展名检测格式
  const ext = filePath.split('.').pop()?.toLowerCase();
  
  if (ext === 'sql') {
    detectedFormat.value = 'SQL';
    await validateFormat('SQL', filePath);
  } else if (ext === 'csv') {
    detectedFormat.value = 'CSV';
    await validateFormat('CSV', filePath);
  } else if (ext === 'json') {
    detectedFormat.value = 'JSON';
    await validateFormat('JSON', filePath);
  } else {
    formatError.value = '不支持的文件格式';
  }
};

// 验证文件格式
const validateFormat = async (format: 'SQL' | 'CSV' | 'JSON', filePath: string) => {
  try {
    if (format === 'SQL') {
      await ValidateSQLFormat(props.profileId, filePath);
    } else if (format === 'CSV') {
      await ValidateCSVFormat(props.profileId, filePath);
      // CSV 验证通过后，初始化映射
      initializeMappingForCSV();
    } else if (format === 'JSON') {
      await ValidateJSONFormat(props.profileId, filePath);
      // JSON 验证通过后，初始化映射
      initializeMappingForJSON();
    }
    
    formatValidated.value = true;
    formatError.value = '';
  } catch (error: any) {
    formatValidated.value = false;
    formatError.value = error.message || '格式验证失败';
    ElMessage.error(`文件格式验证失败: ${error.message || error}`);
  }
};

// 初始化 CSV 映射
const initializeMappingForCSV = () => {
  // 默认使用表的列名作为文件列名
  fileColumns.value = props.tableColumns.map(col => col.name);
  form.value.mapping.FileColumns = [...fileColumns.value];
  form.value.mapping.TableColumns = [...fileColumns.value];
};

// 初始化 JSON 映射
const initializeMappingForJSON = () => {
  // 默认使用表的列名作为文件列名
  fileColumns.value = props.tableColumns.map(col => col.name);
  form.value.mapping.FileColumns = [...fileColumns.value];
  form.value.mapping.TableColumns = [...fileColumns.value];
};

// 添加映射
const addMapping = () => {
  fileColumns.value.push('');
  form.value.mapping.FileColumns.push('');
  form.value.mapping.TableColumns.push('');
};

// 删除映射
const removeMapping = (index: number) => {
  fileColumns.value.splice(index, 1);
  form.value.mapping.FileColumns.splice(index, 1);
  form.value.mapping.TableColumns.splice(index, 1);
};

// 处理导入
const handleImport = async () => {
  if (!canImport.value) {
    ElMessage.warning('请完成必要的配置');
    return;
  }

  importing.value = true;
  importResult.value = null;
  progress.value = {
    current: 0,
    total: 0,
    percentage: 0,
    status: undefined,
  };

  try {
    // 监听进度事件
    const progressHandler = (data: any) => {
      if (
        data.profileId === props.profileId &&
        data.database === props.database &&
        data.format === detectedFormat.value
      ) {
        progress.value.current = data.current;
        progress.value.total = data.total;
        progress.value.percentage = Math.round(data.percentage);
      }
    };

    const completedHandler = (data: any) => {
      if (
        data.profileId === props.profileId &&
        data.database === props.database &&
        data.format === detectedFormat.value
      ) {
        progress.value.status = 'success';
        progress.value.percentage = 100;
      }
    };

    const failedHandler = (data: any) => {
      if (
        data.profileId === props.profileId &&
        data.database === props.database &&
        data.format === detectedFormat.value
      ) {
        progress.value.status = 'exception';
        ElMessage.error(`导入失败: ${data.error}`);
        importing.value = false;
      }
    };

    EventsOn('import:progress', progressHandler);
    EventsOn('import:completed', completedHandler);
    EventsOn('import:failed', failedHandler);

    // 准备映射数据
    const mapping = importexport.ColumnMapping.createFrom({
      FileColumns: fileColumns.value.filter((_, i) => form.value.mapping.TableColumns[i]),
      TableColumns: form.value.mapping.TableColumns.filter(col => col),
    });

    // 调用相应的导入方法
    let result: importexport.ImportResult;
    
    if (detectedFormat.value === 'SQL') {
      result = await ImportFromSQL(props.profileId, props.database, form.value.filePath);
    } else if (detectedFormat.value === 'CSV') {
      result = await ImportFromCSV(
        props.profileId,
        props.database,
        props.table,
        form.value.filePath,
        mapping
      );
    } else if (detectedFormat.value === 'JSON') {
      result = await ImportFromJSON(
        props.profileId,
        props.database,
        props.table,
        form.value.filePath,
        mapping
      );
    } else {
      throw new Error('不支持的文件格式');
    }

    importResult.value = result;
    
    if (result.FailedRows === 0) {
      ElMessage.success(`导入成功！共导入 ${result.SuccessRows} 行数据`);
      emit('success');
    } else {
      ElMessage.warning(
        `导入完成，成功 ${result.SuccessRows} 行，失败 ${result.FailedRows} 行`
      );
    }

    // 清理事件监听器
    setTimeout(() => {
      EventsOff('import:progress');
      EventsOff('import:completed');
      EventsOff('import:failed');
    }, 2000);
  } catch (error: any) {
    console.error('Import failed:', error);
    progress.value.status = 'exception';
    ElMessage.error(`导入失败: ${error.message || error}`);
  } finally {
    importing.value = false;
  }
};

// 关闭对话框
const handleClose = () => {
  if (importing.value) {
    return;
  }
  
  visible.value = false;
  
  // 重置表单
  setTimeout(() => {
    form.value = {
      filePath: '',
      mapping: {
        FileColumns: [],
        TableColumns: [],
      },
    };
    fileColumns.value = [];
    detectedFormat.value = null;
    formatValidated.value = false;
    formatError.value = '';
    importResult.value = null;
    progress.value = {
      current: 0,
      total: 0,
      percentage: 0,
      status: undefined,
    };
    importing.value = false;
  }, 300);
};

// 监听文件路径变化
watch(() => form.value.filePath, (newPath) => {
  if (!newPath) {
    detectedFormat.value = null;
    formatValidated.value = false;
    formatError.value = '';
    fileColumns.value = [];
    form.value.mapping = {
      FileColumns: [],
      TableColumns: [],
    };
  }
});
</script>

<style scoped>
.mapping-item {
  margin-bottom: 24px;
}

.mapping-container {
  width: 100%;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.mapping-header {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 600;
  color: #606266;
}

.mapping-col {
  flex: 1;
  text-align: center;
}

.mapping-arrow {
  width: 40px;
  text-align: center;
}

.mapping-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.mapping-input,
.mapping-select {
  flex: 1;
}

.mapping-arrow-icon {
  color: #909399;
  font-size: 16px;
}

.progress-info {
  margin-top: 8px;
}

.result-container {
  width: 100%;
}

.error-details {
  margin-top: 16px;
}

.error-item {
  padding: 4px 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
