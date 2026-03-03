<template>
  <div class="cell-editor">
    <!-- 外键列 - 关联数据选择器 -->
    <el-select
      v-if="isForeignKey"
      v-model="editValue"
      placeholder="选择关联数据"
      size="small"
      filterable
      clearable
      :loading="foreignKeyLoading"
      @blur="handleBlur"
      @change="handleChange"
      ref="editorRef"
    >
      <el-option
        v-for="option in foreignKeyOptions"
        :key="option.value"
        :label="option.label"
        :value="option.value"
      />
    </el-select>

    <!-- 枚举类型 - 单选 -->
    <el-radio-group
      v-else-if="isEnumType"
      v-model="editValue"
      size="small"
      @change="handleChange"
      ref="editorRef"
    >
      <el-radio
        v-for="option in enumOptions"
        :key="option"
        :label="option"
      >
        {{ option }}
      </el-radio>
    </el-radio-group>

    <!-- 日期时间类型 -->
    <el-date-picker
      v-else-if="isDateTimeType"
      v-model="editValue"
      :type="datePickerType"
      placeholder="选择日期时间"
      size="small"
      @blur="handleBlur"
      @change="handleChange"
      ref="editorRef"
    />

    <!-- 数字类型 -->
    <el-input-number
      v-else-if="isNumericType"
      v-model="editValue"
      :controls="true"
      size="small"
      @blur="handleBlur"
      @change="handleChange"
      ref="editorRef"
    />

    <!-- 默认输入框 -->
    <el-input
      v-else
      v-model="editValue"
      size="small"
      @blur="handleBlur"
      @keyup.enter="handleChange"
      @keyup.esc="handleCancel"
      ref="editorRef"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import type { Column, ForeignKey } from '../types/api';
import { DataAPI } from '../api';

interface Props {
  value: any;
  column: Column;
  foreignKeys: ForeignKey[];
  profileId: string;
  database: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  change: [value: any];
  cancel: [];
}>();

const editValue = ref(props.value);
const editorRef = ref<any>(null);
const foreignKeyOptions = ref<Array<{ label: string; value: any }>>([]);
const foreignKeyLoading = ref(false);

// 判断是否为外键列
const isForeignKey = computed(() => {
  return props.foreignKeys.some(fk => fk.columns.includes(props.column.name));
});

// 判断是否为枚举类型
const isEnumType = computed(() => {
  return /^enum\(/i.test(props.column.type);
});

// 获取枚举选项
const enumOptions = computed(() => {
  if (!isEnumType.value) return [];
  
  const match = props.column.type.match(/^enum\((.*)\)$/i);
  if (!match) return [];
  
  // 解析枚举值，去掉引号
  return match[1].split(',').map(v => v.trim().replace(/^['"]|['"]$/g, ''));
});

// 判断是否为日期时间类型
const isDateTimeType = computed(() => {
  return /^(date|datetime|timestamp|time|year)$/i.test(props.column.type);
});

// 判断是否为数字类型
const isNumericType = computed(() => {
  return /^(int|integer|tinyint|smallint|mediumint|bigint|float|double|decimal|numeric)(\(.*\))?$/i.test(props.column.type);
});

// 获取日期选择器类型
const datePickerType = computed(() => {
  if (/^datetime$/i.test(props.column.type)) return 'datetime';
  if (/^date$/i.test(props.column.type)) return 'date';
  if (/^time$/i.test(props.column.type)) return 'time';
  if (/^year$/i.test(props.column.type)) return 'year';
  return 'datetime';
});

// 加载外键选项
const loadForeignKeyOptions = async () => {
  if (!isForeignKey.value) return;

  const fk = props.foreignKeys.find(fk => fk.columns.includes(props.column.name));
  if (!fk) return;

  foreignKeyLoading.value = true;

  try {
    const result = await DataAPI.queryData(props.profileId, {
      database: props.database,
      table: fk.referencedTable,
      columns: fk.referencedColumns,
      filters: [],
      orderBy: [],
      limit: 100,
      offset: 0,
    });

    foreignKeyOptions.value = result.rows.map((row) => ({
      label: row.join(' - '),
      value: row[0],
    }));
  } catch (error) {
    console.error('Failed to load foreign key options:', error);
  } finally {
    foreignKeyLoading.value = false;
  }
};

// 验证数据类型
const validateValue = (value: any): boolean => {
  if (value === null || value === undefined || value === '') {
    return props.column.nullable;
  }

  if (isNumericType.value) {
    return !isNaN(Number(value));
  }

  return true;
};

// 处理值变化
const handleChange = () => {
  if (!validateValue(editValue.value)) {
    // 验证失败，恢复原值
    editValue.value = props.value;
    return;
  }

  let finalValue = editValue.value;
  
  // 格式化日期时间为 MySQL 格式
  if (finalValue !== undefined && finalValue !== null && finalValue !== '' && isDateTimeType.value) {
    if (finalValue instanceof Date) {
      // 转换为 MySQL datetime 格式: YYYY-MM-DD HH:MM:SS
      const year = finalValue.getFullYear();
      const month = String(finalValue.getMonth() + 1).padStart(2, '0');
      const day = String(finalValue.getDate()).padStart(2, '0');
      const hours = String(finalValue.getHours()).padStart(2, '0');
      const minutes = String(finalValue.getMinutes()).padStart(2, '0');
      const seconds = String(finalValue.getSeconds()).padStart(2, '0');
      
      if (/^datetime|timestamp$/i.test(props.column.type)) {
        finalValue = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
      } else if (/^date$/i.test(props.column.type)) {
        finalValue = `${year}-${month}-${day}`;
      } else if (/^time$/i.test(props.column.type)) {
        finalValue = `${hours}:${minutes}:${seconds}`;
      } else if (/^year$/i.test(props.column.type)) {
        finalValue = year;
      }
    }
  }

  emit('change', finalValue);
};

// 处理失焦
const handleBlur = () => {
  handleChange();
};

// 处理取消
const handleCancel = () => {
  emit('cancel');
};

// 组件挂载后聚焦
onMounted(() => {
  if (isForeignKey.value) {
    loadForeignKeyOptions();
  }

  // 聚焦编辑器
  setTimeout(() => {
    if (editorRef.value) {
      if (editorRef.value.focus) {
        editorRef.value.focus();
      } else if (editorRef.value.$el) {
        const input = editorRef.value.$el.querySelector('input');
        if (input) input.focus();
      }
    }
  }, 100);
});

// 监听值变化
watch(() => props.value, (val) => {
  editValue.value = val;
});
</script>

<style scoped>
.cell-editor {
  width: 100%;
}
</style>
