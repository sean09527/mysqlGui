<template>
  <el-dialog
    v-model="visible"
    title="插入新行"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="120px"
      v-loading="loading"
    >
      <el-form-item
        v-for="column in columns"
        :key="column.name"
        :label="column.name"
        :prop="column.name"
      >
        <template #label>
          <span>{{ column.name }}</span>
          <el-tooltip v-if="column.comment" :content="column.comment" placement="top">
            <el-icon style="margin-left: 4px"><QuestionFilled /></el-icon>
          </el-tooltip>
        </template>

        <!-- 自增列 - 只读 -->
        <el-input
          v-if="column.autoIncrement"
          :model-value="'AUTO_INCREMENT'"
          disabled
          size="small"
        />

        <!-- 外键列 - 关联数据选择器 -->
        <el-select
          v-else-if="isForeignKey(column.name)"
          v-model="formData[column.name]"
          placeholder="选择关联数据"
          size="small"
          filterable
          clearable
          :loading="foreignKeyLoading[column.name]"
          style="width: 100%"
        >
          <el-option
            v-for="option in foreignKeyOptions[column.name]"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </el-select>

        <!-- 枚举类型 - 单选 -->
        <el-radio-group
          v-else-if="isEnumType(column.type)"
          v-model="formData[column.name]"
          size="small"
        >
          <el-radio
            v-for="option in getEnumOptions(column.type)"
            :key="option"
            :label="option"
          >
            {{ option }}
          </el-radio>
        </el-radio-group>

        <!-- 日期时间类型 -->
        <el-date-picker
          v-else-if="isDateTimeType(column.type)"
          v-model="formData[column.name]"
          :type="getDatePickerType(column.type)"
          placeholder="选择日期时间"
          size="small"
          style="width: 100%"
        />

        <!-- 文本类型 -->
        <el-input
          v-else-if="isTextType(column.type)"
          v-model="formData[column.name]"
          type="textarea"
          :rows="3"
          placeholder="输入文本"
          size="small"
        />

        <!-- 数字类型 -->
        <el-input-number
          v-else-if="isNumericType(column.type)"
          v-model="formData[column.name]"
          :controls="true"
          size="small"
          style="width: 100%"
        />

        <!-- 默认输入框 -->
        <el-input
          v-else
          v-model="formData[column.name]"
          placeholder="输入值"
          size="small"
        />

        <div class="field-info">
          <span class="field-type">{{ column.type }}</span>
          <span v-if="!column.nullable" class="field-required">必填</span>
          <span v-if="column.defaultValue" class="field-default">
            默认: {{ column.defaultValue }}
          </span>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose" size="small">取消</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="loading" size="small">
        插入
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { QuestionFilled } from '@element-plus/icons-vue';
import type { Column, ForeignKey } from '../types/api';
import { DataAPI } from '../api';

interface Props {
  modelValue: boolean;
  columns: Column[];
  foreignKeys: ForeignKey[];
  profileId: string;
  database: string;
  table: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  success: [];
}>();

const visible = ref(props.modelValue);
const loading = ref(false);
const formRef = ref<any>(null);
const formData = reactive<Record<string, any>>({});
const foreignKeyOptions = reactive<Record<string, Array<{ label: string; value: any }>>>({});
const foreignKeyLoading = reactive<Record<string, boolean>>({});

// 表单验证规则
const rules = reactive<Record<string, any>>({});

// 监听 modelValue 变化
watch(() => props.modelValue, (val) => {
  visible.value = val;
  if (val) {
    initForm();
  }
});

// 监听 visible 变化
watch(visible, (val) => {
  emit('update:modelValue', val);
});

// 初始化表单
const initForm = () => {
  // 清空表单数据
  Object.keys(formData).forEach(key => delete formData[key]);
  
  // 设置默认值
  props.columns.forEach((column) => {
    if (column.autoIncrement) {
      // 自增列不需要设置
      return;
    }
    
    if (column.defaultValue !== undefined && column.defaultValue !== null) {
      formData[column.name] = column.defaultValue;
    } else if (!column.nullable) {
      // 必填字段设置初始值
      if (isEnumType(column.type)) {
        // 枚举类型默认选择第一个值
        const options = getEnumOptions(column.type);
        formData[column.name] = options.length > 0 ? options[0] : '';
      } else if (isNumericType(column.type)) {
        formData[column.name] = 0;
      } else {
        formData[column.name] = '';
      }
    }
  });

  // 设置验证规则
  props.columns.forEach((column) => {
    if (!column.nullable && !column.autoIncrement) {
      rules[column.name] = [
        { required: true, message: `${column.name} 不能为空`, trigger: 'blur' }
      ];
    }
  });

  // 加载外键选项
  loadForeignKeyOptions();
};

// 判断是否为外键列
const isForeignKey = (columnName: string): boolean => {
  return props.foreignKeys.some(fk => fk.columns.includes(columnName));
};

// 判断是否为枚举类型
const isEnumType = (type: string): boolean => {
  return /^enum\(/i.test(type);
};

// 获取枚举选项
const getEnumOptions = (type: string): string[] => {
  const match = type.match(/^enum\((.*)\)$/i);
  if (!match) return [];
  
  // 解析枚举值，去掉引号
  return match[1].split(',').map(v => v.trim().replace(/^['"]|['"]$/g, ''));
};

// 判断是否为日期时间类型
const isDateTimeType = (type: string): boolean => {
  return /^(date|datetime|timestamp|time|year)$/i.test(type);
};

// 判断是否为文本类型
const isTextType = (type: string): boolean => {
  return /^(text|mediumtext|longtext|tinytext)$/i.test(type);
};

// 判断是否为数字类型
const isNumericType = (type: string): boolean => {
  return /^(int|integer|tinyint|smallint|mediumint|bigint|float|double|decimal|numeric)(\(.*\))?$/i.test(type);
};

// 获取日期选择器类型
const getDatePickerType = (type: string): string => {
  if (/^datetime$/i.test(type)) return 'datetime';
  if (/^date$/i.test(type)) return 'date';
  if (/^time$/i.test(type)) return 'time';
  if (/^year$/i.test(type)) return 'year';
  return 'datetime';
};

// 加载外键选项
const loadForeignKeyOptions = async () => {
  for (const fk of props.foreignKeys) {
    for (const column of fk.columns) {
      foreignKeyLoading[column] = true;
      
      try {
        // 查询关联表的数据
        const result = await DataAPI.queryData(props.profileId, {
          database: props.database,
          table: fk.referencedTable,
          columns: fk.referencedColumns,
          filters: [],
          orderBy: [],
          limit: 100,
          offset: 0,
        });

        // 构建选项列表
        foreignKeyOptions[column] = result.rows.map((row) => ({
          label: row.join(' - '),
          value: row[0],
        }));
      } catch (error: any) {
        console.error(`Failed to load foreign key options for ${column}:`, error);
        ElMessage.error(`加载关联数据失败: ${error.message || error}`);
      } finally {
        foreignKeyLoading[column] = false;
      }
    }
  }
};

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
  } catch {
    return;
  }

  loading.value = true;

  try {
    // 过滤掉自增列和空值，并格式化日期时间
    const data: Record<string, any> = {};
    props.columns.forEach((column) => {
      if (column.autoIncrement) return;
      
      let value = formData[column.name];
      
      // 格式化日期时间为 MySQL 格式
      if (value !== undefined && value !== null && value !== '' && isDateTimeType(column.type)) {
        if (value instanceof Date) {
          // 转换为 MySQL datetime 格式: YYYY-MM-DD HH:MM:SS
          const year = value.getFullYear();
          const month = String(value.getMonth() + 1).padStart(2, '0');
          const day = String(value.getDate()).padStart(2, '0');
          const hours = String(value.getHours()).padStart(2, '0');
          const minutes = String(value.getMinutes()).padStart(2, '0');
          const seconds = String(value.getSeconds()).padStart(2, '0');
          
          if (/^datetime|timestamp$/i.test(column.type)) {
            value = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
          } else if (/^date$/i.test(column.type)) {
            value = `${year}-${month}-${day}`;
          } else if (/^time$/i.test(column.type)) {
            value = `${hours}:${minutes}:${seconds}`;
          } else if (/^year$/i.test(column.type)) {
            value = year;
          }
        }
      }
      
      if (value !== undefined && value !== null && value !== '') {
        data[column.name] = value;
      } else if (!column.nullable) {
        // 必填字段但为空，使用默认值或 NULL
        if (column.defaultValue !== undefined) {
          data[column.name] = column.defaultValue;
        }
      }
    });

    await DataAPI.insertRow(props.profileId, props.database, props.table, data);
    
    ElMessage.success('插入成功');
    emit('success');
    handleClose();
  } catch (error: any) {
    console.error('Failed to insert row:', error);
    ElMessage.error(`插入失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

// 关闭对话框
const handleClose = () => {
  visible.value = false;
  formRef.value?.resetFields();
};
</script>

<style scoped>
.field-info {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  display: flex;
  gap: 8px;
}

.field-type {
  color: #909399;
}

.field-required {
  color: #f56c6c;
}

.field-default {
  color: #67c23a;
}
</style>
