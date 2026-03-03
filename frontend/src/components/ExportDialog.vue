<template>
  <el-dialog
    v-model="visible"
    title="导出数据"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form :model="form" label-width="100px" size="default">
      <!-- 导出格式 -->
      <el-form-item label="导出格式">
        <el-radio-group v-model="form.format">
          <el-radio label="SQL">SQL INSERT 语句</el-radio>
          <el-radio label="CSV">CSV 格式</el-radio>
          <el-radio label="JSON">JSON 格式</el-radio>
        </el-radio-group>
      </el-form-item>

      <!-- 导出范围 -->
      <el-form-item label="导出范围">
        <el-radio-group v-model="form.scope">
          <el-radio label="all">全部数据</el-radio>
          <el-radio label="filtered" :disabled="!hasFilters">当前筛选结果</el-radio>
        </el-radio-group>
        <div v-if="form.scope === 'filtered' && hasFilters" class="scope-info">
          <el-text type="info" size="small">
            将导出当前筛选和排序条件下的数据
          </el-text>
        </div>
      </el-form-item>

      <!-- 文件保存路径 -->
      <el-form-item label="保存路径">
        <el-input
          v-model="form.outputPath"
          placeholder="点击浏览按钮选择保存位置"
          readonly
        >
          <template #append>
            <el-button @click="selectOutputPath">浏览</el-button>
          </template>
        </el-input>
      </el-form-item>

      <!-- 导出进度 -->
      <el-form-item v-if="exporting" label="导出进度">
        <el-progress
          :percentage="progress.percentage"
          :status="progress.status"
        />
        <div class="progress-info">
          <el-text size="small">
            已导出 {{ progress.current }} / {{ progress.total }} 行
          </el-text>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose" :disabled="exporting">取消</el-button>
        <el-button
          type="primary"
          @click="handleExport"
          :loading="exporting"
          :disabled="!form.outputPath"
        >
          {{ exporting ? '导出中...' : '开始导出' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { ExportToSQL, ExportToCSV, ExportToJSON, SaveFileDialog } from '../../wailsjs/go/backend/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { backend, repository } from '../../wailsjs/go/models';
import type { Filter, OrderBy } from '../types/api';

interface Props {
  modelValue: boolean;
  profileId: string;
  database: string;
  table: string;
  filters?: Filter[];
  orderBy?: OrderBy[];
}

const props = withDefaults(defineProps<Props>(), {
  filters: () => [],
  orderBy: () => [],
});

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'success': [];
}>();

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});

const form = ref({
  format: 'SQL' as 'SQL' | 'CSV' | 'JSON',
  scope: 'all' as 'all' | 'filtered',
  outputPath: '',
});

const exporting = ref(false);
const progress = ref({
  current: 0,
  total: 0,
  percentage: 0,
  status: undefined as 'success' | 'exception' | undefined,
});

const hasFilters = computed(() => {
  return props.filters && props.filters.length > 0;
});

// 选择输出路径
const selectOutputPath = async () => {
  try {
    const ext = form.value.format.toLowerCase();
    const filters: backend.FileDialogFilter[] = [
      backend.FileDialogFilter.createFrom({
        displayName: `${form.value.format} 文件`,
        pattern: `*.${ext}`,
      }),
    ];
    
    const defaultFilename = `${props.table}_export.${ext}`;
    
    // 使用后端 API 打开文件保存对话框
    const path = await SaveFileDialog('选择保存位置', defaultFilename, filters);
    
    if (path) {
      form.value.outputPath = path;
    }
  } catch (error: any) {
    console.error('Failed to select output path:', error);
    ElMessage.error(`选择文件失败: ${error.message || error}`);
  }
};

// 构建查询条件
const buildQuery = (): repository.DataQuery => {
  const query: repository.DataQuery = {
    Database: props.database,
    Table: props.table,
    Columns: [],
    Filters: [],
    OrderBy: [],
    Limit: 0,
    Offset: 0,
  };

  // 如果选择导出筛选结果，应用筛选和排序条件
  if (form.value.scope === 'filtered') {
    query.Filters = props.filters.map(f => ({
      Column: f.Column || (f as any).column,
      Operator: f.Operator || (f as any).operator,
      Value: f.Value || (f as any).value,
    }));
    
    query.OrderBy = props.orderBy.map(o => ({
      Column: o.Column || (o as any).column,
      Direction: o.Direction || (o as any).direction,
    }));
  }

  return query;
};

// 处理导出
const handleExport = async () => {
  if (!form.value.outputPath) {
    ElMessage.warning('请选择保存路径');
    return;
  }

  exporting.value = true;
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
        data.table === props.table &&
        data.format === form.value.format
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
        data.table === props.table &&
        data.format === form.value.format
      ) {
        progress.value.status = 'success';
        progress.value.percentage = 100;
        
        ElMessage.success(`导出成功！文件已保存到: ${data.outputPath}`);
        
        setTimeout(() => {
          emit('success');
          handleClose();
        }, 1500);
      }
    };

    const failedHandler = (data: any) => {
      if (
        data.profileId === props.profileId &&
        data.database === props.database &&
        data.table === props.table &&
        data.format === form.value.format
      ) {
        progress.value.status = 'exception';
        ElMessage.error(`导出失败: ${data.error}`);
        exporting.value = false;
      }
    };

    EventsOn('export:progress', progressHandler);
    EventsOn('export:completed', completedHandler);
    EventsOn('export:failed', failedHandler);

    // 构建查询条件
    const query = buildQuery();

    // 调用相应的导出方法
    if (form.value.format === 'SQL') {
      await ExportToSQL(props.profileId, props.database, props.table, query, form.value.outputPath);
    } else if (form.value.format === 'CSV') {
      await ExportToCSV(props.profileId, props.database, props.table, query, form.value.outputPath);
    } else if (form.value.format === 'JSON') {
      await ExportToJSON(props.profileId, props.database, props.table, query, form.value.outputPath);
    }

    // 清理事件监听器
    setTimeout(() => {
      EventsOff('export:progress');
      EventsOff('export:completed');
      EventsOff('export:failed');
    }, 2000);
  } catch (error: any) {
    console.error('Export failed:', error);
    progress.value.status = 'exception';
    ElMessage.error(`导出失败: ${error.message || error}`);
    exporting.value = false;
  }
};

// 关闭对话框
const handleClose = () => {
  if (exporting.value) {
    return;
  }
  
  visible.value = false;
  
  // 重置表单
  setTimeout(() => {
    form.value = {
      format: 'SQL',
      scope: 'all',
      outputPath: '',
    };
    progress.value = {
      current: 0,
      total: 0,
      percentage: 0,
      status: undefined,
    };
    exporting.value = false;
  }, 300);
};

// 监听格式变化，清空输出路径
watch(() => form.value.format, () => {
  form.value.outputPath = '';
});
</script>

<style scoped>
.scope-info {
  margin-top: 8px;
}

.progress-info {
  margin-top: 8px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
