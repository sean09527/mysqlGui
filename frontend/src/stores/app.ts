import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface AppSettings {
  language: string;
  theme: string;
  pageSize: number;
  queryTimeout: number;
}

export interface OperationProgress {
  id: string;
  message: string;
  progress: number;
  cancelable: boolean;
}

export const useAppStore = defineStore('app', () => {
  const settings = ref<AppSettings>({
    language: 'zh-CN',
    theme: 'light',
    pageSize: 100,
    queryTimeout: 30,
  });

  const loading = ref(false);
  const loadingMessage = ref('');
  const operations = ref<Map<string, OperationProgress>>(new Map());

  function updateSettings(newSettings: Partial<AppSettings>) {
    settings.value = { ...settings.value, ...newSettings };
  }

  function setLoading(isLoading: boolean, message = '') {
    loading.value = isLoading;
    loadingMessage.value = message;
  }

  function addOperation(id: string, message: string, cancelable: boolean = false) {
    operations.value.set(id, {
      id,
      message,
      progress: 0,
      cancelable,
    });
  }

  function updateOperation(id: string, progress: number, message?: string) {
    const operation = operations.value.get(id);
    if (operation) {
      operation.progress = progress;
      if (message) {
        operation.message = message;
      }
    }
  }

  function removeOperation(id: string) {
    operations.value.delete(id);
  }

  function clearOperations() {
    operations.value.clear();
  }

  return {
    settings,
    loading,
    loadingMessage,
    operations,
    updateSettings,
    setLoading,
    addOperation,
    updateOperation,
    removeOperation,
    clearOperations,
  };
});
