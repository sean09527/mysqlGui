<template>
  <div class="data-filter">
    <div class="filter-header">
      <span class="filter-title">数据筛选</span>
      <el-button type="primary" size="small" @click="addFilter">
        <el-icon><Plus /></el-icon>
        添加条件
      </el-button>
    </div>

    <div class="filter-list" v-if="filters.length > 0">
      <div v-for="(filter, index) in filters" :key="index" class="filter-item">
        <!-- 列选择 -->
        <el-select
          v-model="filter.column"
          placeholder="选择列"
          size="small"
          style="width: 150px"
        >
          <el-option
            v-for="column in columns"
            :key="column"
            :label="column"
            :value="column"
          />
        </el-select>

        <!-- 操作符选择 -->
        <el-select
          v-model="filter.operator"
          placeholder="操作符"
          size="small"
          style="width: 120px"
        >
          <el-option label="等于" value="=" />
          <el-option label="不等于" value="!=" />
          <el-option label="大于" value=">" />
          <el-option label="小于" value="<" />
          <el-option label="大于等于" value=">=" />
          <el-option label="小于等于" value="<=" />
          <el-option label="包含" value="LIKE" />
          <el-option label="不包含" value="NOT LIKE" />
          <el-option label="在列表中" value="IN" />
          <el-option label="不在列表中" value="NOT IN" />
          <el-option label="为空" value="IS NULL" />
          <el-option label="不为空" value="IS NOT NULL" />
        </el-select>

        <!-- 值输入 -->
        <el-input
          v-if="!isNullOperator(filter.operator)"
          v-model="filter.value"
          placeholder="输入值"
          size="small"
          style="flex: 1"
        />

        <!-- 删除按钮 -->
        <el-button
          type="danger"
          size="small"
          @click="removeFilter(index)"
          :icon="Delete"
        />
      </div>
    </div>

    <!-- 当前筛选条件显示 -->
    <div class="current-filters" v-if="appliedFilters.length > 0">
      <div class="current-filters-title">当前筛选条件:</div>
      <el-tag
        v-for="(filter, index) in appliedFilters"
        :key="index"
        closable
        @close="removeAppliedFilter(index)"
        style="margin-right: 8px; margin-bottom: 8px"
      >
        {{ formatFilter(filter) }}
      </el-tag>
    </div>

    <!-- 操作按钮 -->
    <div class="filter-actions">
      <el-button type="primary" size="small" @click="applyFilters">
        应用筛选
      </el-button>
      <el-button size="small" @click="clearFilters">
        清除筛选
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { Plus, Delete } from '@element-plus/icons-vue';
import type { Filter } from '../types/api';

interface Props {
  columns: string[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  apply: [filters: Filter[]];
  clear: [];
}>();

// 当前编辑的筛选条件
const filters = ref<Filter[]>([]);

// 已应用的筛选条件
const appliedFilters = ref<Filter[]>([]);

// 调试：输出 props
console.log('=== DataFilter Props ===');
console.log('columns:', props.columns);
console.log('columns length:', props.columns?.length);

// 添加筛选条件
const addFilter = () => {
  filters.value.push({
    column: props.columns[0] || '',
    operator: '=',
    value: '',
  });
};

// 删除筛选条件
const removeFilter = (index: number) => {
  filters.value.splice(index, 1);
};

// 删除已应用的筛选条件
const removeAppliedFilter = (index: number) => {
  appliedFilters.value.splice(index, 1);
  emit('apply', appliedFilters.value);
};

// 应用筛选
const applyFilters = () => {
  // 验证筛选条件
  const validFilters = filters.value.filter((filter) => {
    if (!filter.column) return false;
    if (!filter.operator) return false;
    if (!isNullOperator(filter.operator) && !filter.value) return false;
    return true;
  });

  appliedFilters.value = validFilters;
  emit('apply', validFilters);
};

// 清除筛选
const clearFilters = () => {
  filters.value = [];
  appliedFilters.value = [];
  emit('clear');
};

// 判断是否为 NULL 操作符
const isNullOperator = (operator: string): boolean => {
  return operator === 'IS NULL' || operator === 'IS NOT NULL';
};

// 格式化筛选条件显示
const formatFilter = (filter: Filter): string => {
  const operatorMap: Record<string, string> = {
    '=': '等于',
    '!=': '不等于',
    '>': '大于',
    '<': '小于',
    '>=': '大于等于',
    '<=': '小于等于',
    'LIKE': '包含',
    'NOT LIKE': '不包含',
    'IN': '在列表中',
    'NOT IN': '不在列表中',
    'IS NULL': '为空',
    'IS NOT NULL': '不为空',
  };

  const operatorText = operatorMap[filter.operator] || filter.operator;
  
  if (isNullOperator(filter.operator)) {
    return `${filter.column} ${operatorText}`;
  }

  let value = filter.value;
  if (filter.operator === 'LIKE' || filter.operator === 'NOT LIKE') {
    value = `%${value}%`;
  }

  return `${filter.column} ${operatorText} ${value}`;
};

// 暴露方法
defineExpose({
  clearFilters,
});
</script>

<style scoped>
.data-filter {
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filter-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.filter-list {
  margin-bottom: 16px;
}

.filter-item {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

.current-filters {
  margin-bottom: 16px;
  padding: 12px;
  background-color: #fff;
  border-radius: 4px;
}

.current-filters-title {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
}

.filter-actions {
  display: flex;
  gap: 8px;
}
</style>
