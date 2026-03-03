/**
 * 加载状态管理 Composable
 * 
 * 提供全局和局部加载状态管理，支持操作取消和进度反馈
 */

import { ref, computed } from 'vue';
import { ElLoading, ElNotification } from 'element-plus';
import type { LoadingInstance } from 'element-plus/es/components/loading/src/loading';

/**
 * 操作状态接口
 */
export interface OperationState {
  id: string;
  message: string;
  progress?: number;
  cancelable: boolean;
  cancelFn?: () => void;
  startTime: number;
}

/**
 * 全局加载状态
 */
const operations = ref<Map<string, OperationState>>(new Map());
const loadingInstance = ref<LoadingInstance | null>(null);

/**
 * 加载状态管理 Hook
 */
export function useLoading() {
  /**
   * 是否有正在进行的操作
   */
  const isLoading = computed(() => operations.value.size > 0);

  /**
   * 当前操作列表
   */
  const currentOperations = computed(() => Array.from(operations.value.values()));

  /**
   * 开始一个操作
   */
  function startOperation(
    id: string,
    message: string,
    options?: {
      cancelable?: boolean;
      cancelFn?: () => void;
      showFullscreen?: boolean;
    }
  ) {
    const operation: OperationState = {
      id,
      message,
      cancelable: options?.cancelable ?? false,
      cancelFn: options?.cancelFn,
      startTime: Date.now(),
    };

    operations.value.set(id, operation);

    // 如果需要全屏加载
    if (options?.showFullscreen && !loadingInstance.value) {
      loadingInstance.value = ElLoading.service({
        lock: true,
        text: message,
        background: 'rgba(0, 0, 0, 0.7)',
      });
    }

    return id;
  }

  /**
   * 更新操作进度
   */
  function updateProgress(id: string, progress: number, message?: string) {
    const operation = operations.value.get(id);
    if (operation) {
      operation.progress = progress;
      if (message) {
        operation.message = message;
      }
      operations.value.set(id, operation);

      // 更新全屏加载文本
      if (loadingInstance.value) {
        loadingInstance.value.setText(`${message || operation.message} (${progress}%)`);
      }
    }
  }

  /**
   * 完成操作
   */
  function completeOperation(id: string, successMessage?: string) {
    operations.value.delete(id);

    // 如果没有其他操作，关闭全屏加载
    if (operations.value.size === 0 && loadingInstance.value) {
      loadingInstance.value.close();
      loadingInstance.value = null;
    }

    // 显示成功消息
    if (successMessage) {
      ElNotification({
        type: 'success',
        title: '操作完成',
        message: successMessage,
        duration: 3000,
      });
    }
  }

  /**
   * 取消操作
   */
  function cancelOperation(id: string) {
    const operation = operations.value.get(id);
    if (operation && operation.cancelable && operation.cancelFn) {
      operation.cancelFn();
      operations.value.delete(id);

      // 如果没有其他操作，关闭全屏加载
      if (operations.value.size === 0 && loadingInstance.value) {
        loadingInstance.value.close();
        loadingInstance.value = null;
      }

      ElNotification({
        type: 'info',
        title: '操作已取消',
        message: operation.message,
        duration: 2000,
      });
    }
  }

  /**
   * 取消所有操作
   */
  function cancelAllOperations() {
    operations.value.forEach((operation) => {
      if (operation.cancelable && operation.cancelFn) {
        operation.cancelFn();
      }
    });
    operations.value.clear();

    if (loadingInstance.value) {
      loadingInstance.value.close();
      loadingInstance.value = null;
    }
  }

  /**
   * 包装异步操作，自动管理加载状态
   */
  async function withLoading<T>(
    operation: (updateProgress: (progress: number, message?: string) => void) => Promise<T>,
    options: {
      id?: string;
      message: string;
      successMessage?: string;
      cancelable?: boolean;
      showFullscreen?: boolean;
    }
  ): Promise<T> {
    const operationId = options.id || `operation_${Date.now()}`;
    let cancelled = false;

    const cancelFn = () => {
      cancelled = true;
    };

    startOperation(operationId, options.message, {
      cancelable: options.cancelable,
      cancelFn,
      showFullscreen: options.showFullscreen,
    });

    try {
      const result = await operation((progress: number, message?: string) => {
        if (!cancelled) {
          updateProgress(operationId, progress, message);
        }
      });

      if (!cancelled) {
        completeOperation(operationId, options.successMessage);
      }

      return result;
    } catch (error) {
      operations.value.delete(operationId);
      if (operations.value.size === 0 && loadingInstance.value) {
        loadingInstance.value.close();
        loadingInstance.value = null;
      }
      throw error;
    }
  }

  return {
    isLoading,
    currentOperations,
    startOperation,
    updateProgress,
    completeOperation,
    cancelOperation,
    cancelAllOperations,
    withLoading,
  };
}

/**
 * 局部加载状态 Hook（用于组件内部）
 */
export function useLocalLoading() {
  const loading = ref(false);
  const loadingMessage = ref('');
  const progress = ref(0);

  function startLoading(message: string = '加载中...') {
    loading.value = true;
    loadingMessage.value = message;
    progress.value = 0;
  }

  function updateLoadingProgress(value: number, message?: string) {
    progress.value = value;
    if (message) {
      loadingMessage.value = message;
    }
  }

  function stopLoading() {
    loading.value = false;
    loadingMessage.value = '';
    progress.value = 0;
  }

  async function withLocalLoading<T>(
    operation: () => Promise<T>,
    message: string = '加载中...'
  ): Promise<T> {
    startLoading(message);
    try {
      const result = await operation();
      return result;
    } finally {
      stopLoading();
    }
  }

  return {
    loading,
    loadingMessage,
    progress,
    startLoading,
    updateLoadingProgress,
    stopLoading,
    withLocalLoading,
  };
}
