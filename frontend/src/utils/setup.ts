/**
 * 应用初始化设置
 * 
 * 配置全局错误处理、加载状态等
 */

import type { App } from 'vue';
import { globalErrorHandler, setupGlobalErrorHandlers } from './errorHandler';

/**
 * 设置全局错误处理
 */
export function setupErrorHandling(app: App) {
  // Vue 错误处理
  app.config.errorHandler = globalErrorHandler;

  // 全局未捕获错误处理
  setupGlobalErrorHandlers();

  console.log('[Setup] Global error handling configured');
}

/**
 * 设置全局配置
 */
export function setupApp(app: App) {
  // 设置错误处理
  setupErrorHandling(app);

  // 可以在这里添加其他全局配置
  // 例如：全局组件注册、全局指令等

  console.log('[Setup] Application setup completed');
}
