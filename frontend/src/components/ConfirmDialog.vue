<template>
  <el-dialog
    v-model="visible"
    :title="title"
    :width="width"
    :close-on-click-modal="false"
    :close-on-press-escape="!loading"
    :show-close="!loading"
    @close="handleClose"
  >
    <div class="confirm-dialog-content">
      <!-- 图标 -->
      <div class="icon-wrapper" :class="typeClass">
        <el-icon :size="48">
          <WarningFilled v-if="type === 'warning'" />
          <QuestionFilled v-else-if="type === 'info'" />
          <CircleCloseFilled v-else-if="type === 'error'" />
          <InfoFilled v-else />
        </el-icon>
      </div>

      <!-- 消息内容 -->
      <div class="message-wrapper">
        <div class="message">{{ message }}</div>
        <div v-if="detail" class="detail">{{ detail }}</div>
        
        <!-- 额外内容插槽 -->
        <div v-if="$slots.default" class="extra-content">
          <slot></slot>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel" :disabled="loading">
          {{ cancelText }}
        </el-button>
        <el-button
          :type="confirmButtonType"
          @click="handleConfirm"
          :loading="loading"
          :disabled="loading"
        >
          {{ confirmText }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import {
  WarningFilled,
  QuestionFilled,
  CircleCloseFilled,
  InfoFilled,
} from '@element-plus/icons-vue';

/**
 * 对话框类型
 */
export type DialogType = 'warning' | 'info' | 'error' | 'success';

/**
 * Props
 */
interface Props {
  modelValue?: boolean;
  title?: string;
  message: string;
  detail?: string;
  type?: DialogType;
  confirmText?: string;
  cancelText?: string;
  width?: string | number;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  title: '确认操作',
  type: 'warning',
  confirmText: '确认',
  cancelText: '取消',
  width: '500px',
  loading: false,
});

/**
 * Emits
 */
const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  confirm: [];
  cancel: [];
}>();

/**
 * 对话框可见性
 */
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});

/**
 * 类型对应的样式类
 */
const typeClass = computed(() => {
  return `icon-${props.type}`;
});

/**
 * 确认按钮类型
 */
const confirmButtonType = computed(() => {
  switch (props.type) {
    case 'warning':
    case 'error':
      return 'danger';
    case 'info':
      return 'primary';
    case 'success':
      return 'success';
    default:
      return 'primary';
  }
});

/**
 * 处理确认
 */
function handleConfirm() {
  emit('confirm');
}

/**
 * 处理取消
 */
function handleCancel() {
  visible.value = false;
  emit('cancel');
}

/**
 * 处理关闭
 */
function handleClose() {
  if (!props.loading) {
    emit('cancel');
  }
}
</script>

<style scoped lang="scss">
.confirm-dialog-content {
  display: flex;
  gap: 20px;
  padding: 20px 0;

  .icon-wrapper {
    flex-shrink: 0;

    &.icon-warning {
      color: var(--el-color-warning);
    }

    &.icon-error {
      color: var(--el-color-danger);
    }

    &.icon-info {
      color: var(--el-color-primary);
    }

    &.icon-success {
      color: var(--el-color-success);
    }
  }

  .message-wrapper {
    flex: 1;
    min-width: 0;

    .message {
      font-size: 16px;
      font-weight: 500;
      color: var(--el-text-color-primary);
      margin-bottom: 8px;
      word-wrap: break-word;
    }

    .detail {
      font-size: 14px;
      color: var(--el-text-color-regular);
      line-height: 1.6;
      word-wrap: break-word;
    }

    .extra-content {
      margin-top: 16px;
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
