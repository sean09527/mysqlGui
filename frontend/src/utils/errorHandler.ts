/**
 * 全局错误处理器
 * 
 * 统一处理应用中的各种错误，提供用户友好的错误消息
 */

import { ElMessage, ElNotification } from 'element-plus';

/**
 * 错误类型枚举
 */
export enum ErrorType {
  CONNECTION = 'connection',
  SQL = 'sql',
  NETWORK = 'network',
  VALIDATION = 'validation',
  PERMISSION = 'permission',
  TIMEOUT = 'timeout',
  UNKNOWN = 'unknown',
}

/**
 * 错误信息接口
 */
export interface ErrorInfo {
  type: ErrorType;
  code?: string | number;
  message: string;
  detail?: string;
  originalError?: any;
}

/**
 * 错误消息映射表
 */
const ERROR_MESSAGES: Record<string, string> = {
  // 连接错误
  'connection_failed': '数据库连接失败',
  'connection_timeout': '连接超时',
  'connection_refused': '连接被拒绝',
  'ssh_connection_failed': 'SSH 连接失败',
  'ssh_auth_failed': 'SSH 认证失败',
  'invalid_credentials': '用户名或密码错误',
  'database_not_found': '数据库不存在',
  
  // SQL 错误
  'sql_syntax_error': 'SQL 语法错误',
  'table_not_found': '表不存在',
  'column_not_found': '列不存在',
  'duplicate_entry': '数据重复',
  'foreign_key_constraint': '外键约束冲突',
  'data_too_long': '数据过长',
  'invalid_data_type': '数据类型不匹配',
  
  // 网络错误
  'network_error': '网络错误',
  'request_timeout': '请求超时',
  'server_error': '服务器错误',
  
  // 验证错误
  'validation_error': '数据验证失败',
  'required_field': '必填字段不能为空',
  'invalid_format': '格式不正确',
  
  // 权限错误
  'permission_denied': '权限不足',
  'access_denied': '访问被拒绝',
  
  // 其他错误
  'operation_cancelled': '操作已取消',
  'file_not_found': '文件不存在',
  'unknown_error': '未知错误',
};

/**
 * 解析错误信息
 */
export function parseError(error: any): ErrorInfo {
  // 如果已经是 ErrorInfo 格式
  if (error && typeof error === 'object' && 'type' in error) {
    return error as ErrorInfo;
  }

  // 解析字符串错误
  if (typeof error === 'string') {
    return classifyError(error);
  }

  // 解析 Error 对象
  if (error instanceof Error) {
    return classifyError(error.message, error);
  }

  // 解析 Wails 错误响应
  if (error && typeof error === 'object') {
    const message = error.message || error.error || error.msg || String(error);
    return classifyError(message, error);
  }

  // 未知错误
  return {
    type: ErrorType.UNKNOWN,
    message: '发生未知错误',
    originalError: error,
  };
}

/**
 * 根据错误消息分类错误类型
 */
function classifyError(message: string, originalError?: any): ErrorInfo {
  const lowerMessage = message.toLowerCase();

  // 连接错误
  if (
    lowerMessage.includes('connection') ||
    lowerMessage.includes('connect') ||
    lowerMessage.includes('连接')
  ) {
    if (lowerMessage.includes('ssh')) {
      return {
        type: ErrorType.CONNECTION,
        message: ERROR_MESSAGES['ssh_connection_failed'],
        detail: message,
        originalError,
      };
    }
    if (lowerMessage.includes('timeout') || lowerMessage.includes('超时')) {
      return {
        type: ErrorType.CONNECTION,
        message: ERROR_MESSAGES['connection_timeout'],
        detail: message,
        originalError,
      };
    }
    if (lowerMessage.includes('refused') || lowerMessage.includes('拒绝')) {
      return {
        type: ErrorType.CONNECTION,
        message: ERROR_MESSAGES['connection_refused'],
        detail: message,
        originalError,
      };
    }
    if (lowerMessage.includes('auth') || lowerMessage.includes('认证') || lowerMessage.includes('password')) {
      return {
        type: ErrorType.CONNECTION,
        message: ERROR_MESSAGES['invalid_credentials'],
        detail: message,
        originalError,
      };
    }
    return {
      type: ErrorType.CONNECTION,
      message: ERROR_MESSAGES['connection_failed'],
      detail: message,
      originalError,
    };
  }

  // SQL 错误
  if (
    lowerMessage.includes('sql') ||
    lowerMessage.includes('syntax') ||
    lowerMessage.includes('query')
  ) {
    if (lowerMessage.includes('syntax')) {
      return {
        type: ErrorType.SQL,
        message: ERROR_MESSAGES['sql_syntax_error'],
        detail: message,
        originalError,
      };
    }
    if (lowerMessage.includes('table') && lowerMessage.includes('not') && lowerMessage.includes('exist')) {
      return {
        type: ErrorType.SQL,
        message: ERROR_MESSAGES['table_not_found'],
        detail: message,
        originalError,
      };
    }
    if (lowerMessage.includes('column') && lowerMessage.includes('not') && lowerMessage.includes('exist')) {
      return {
        type: ErrorType.SQL,
        message: ERROR_MESSAGES['column_not_found'],
        detail: message,
        originalError,
      };
    }
    if (lowerMessage.includes('duplicate')) {
      return {
        type: ErrorType.SQL,
        message: ERROR_MESSAGES['duplicate_entry'],
        detail: message,
        originalError,
      };
    }
    if (lowerMessage.includes('foreign key') || lowerMessage.includes('constraint')) {
      return {
        type: ErrorType.SQL,
        message: ERROR_MESSAGES['foreign_key_constraint'],
        detail: message,
        originalError,
      };
    }
    if (lowerMessage.includes('data too long')) {
      return {
        type: ErrorType.SQL,
        message: ERROR_MESSAGES['data_too_long'],
        detail: message,
        originalError,
      };
    }
    return {
      type: ErrorType.SQL,
      message: '数据库操作失败',
      detail: message,
      originalError,
    };
  }

  // 网络错误
  if (
    lowerMessage.includes('network') ||
    lowerMessage.includes('fetch') ||
    lowerMessage.includes('网络')
  ) {
    if (lowerMessage.includes('timeout') || lowerMessage.includes('超时')) {
      return {
        type: ErrorType.TIMEOUT,
        message: ERROR_MESSAGES['request_timeout'],
        detail: message,
        originalError,
      };
    }
    return {
      type: ErrorType.NETWORK,
      message: ERROR_MESSAGES['network_error'],
      detail: message,
      originalError,
    };
  }

  // 超时错误
  if (lowerMessage.includes('timeout') || lowerMessage.includes('超时')) {
    return {
      type: ErrorType.TIMEOUT,
      message: ERROR_MESSAGES['request_timeout'],
      detail: message,
      originalError,
    };
  }

  // 权限错误
  if (
    lowerMessage.includes('permission') ||
    lowerMessage.includes('access denied') ||
    lowerMessage.includes('权限') ||
    lowerMessage.includes('拒绝访问')
  ) {
    return {
      type: ErrorType.PERMISSION,
      message: ERROR_MESSAGES['permission_denied'],
      detail: message,
      originalError,
    };
  }

  // 验证错误
  if (
    lowerMessage.includes('validation') ||
    lowerMessage.includes('invalid') ||
    lowerMessage.includes('验证')
  ) {
    return {
      type: ErrorType.VALIDATION,
      message: ERROR_MESSAGES['validation_error'],
      detail: message,
      originalError,
    };
  }

  // 默认未知错误
  return {
    type: ErrorType.UNKNOWN,
    message: message || ERROR_MESSAGES['unknown_error'],
    originalError,
  };
}

/**
 * 显示错误消息（使用 Element Plus Message）
 */
export function showError(error: any, title?: string) {
  const errorInfo = parseError(error);
  
  ElMessage({
    type: 'error',
    message: errorInfo.message,
    duration: 5000,
    showClose: true,
  });

  // 如果有详细信息，在控制台输出
  if (errorInfo.detail) {
    console.error('[Error Detail]', errorInfo.detail);
  }
  if (errorInfo.originalError) {
    console.error('[Original Error]', errorInfo.originalError);
  }
}

/**
 * 显示错误通知（使用 Element Plus Notification，用于更详细的错误信息）
 */
export function showErrorNotification(error: any, title: string = '操作失败') {
  const errorInfo = parseError(error);
  
  ElNotification({
    type: 'error',
    title,
    message: errorInfo.detail || errorInfo.message,
    duration: 8000,
    position: 'top-right',
  });

  // 在控制台输出完整错误
  console.error('[Error]', errorInfo);
}

/**
 * 显示成功消息
 */
export function showSuccess(message: string) {
  ElMessage({
    type: 'success',
    message,
    duration: 3000,
    showClose: true,
  });
}

/**
 * 显示警告消息
 */
export function showWarning(message: string) {
  ElMessage({
    type: 'warning',
    message,
    duration: 4000,
    showClose: true,
  });
}

/**
 * 显示信息消息
 */
export function showInfo(message: string) {
  ElMessage({
    type: 'info',
    message,
    duration: 3000,
    showClose: true,
  });
}

/**
 * 包装异步操作，自动处理错误
 */
export async function handleAsyncOperation<T>(
  operation: () => Promise<T>,
  options?: {
    successMessage?: string;
    errorTitle?: string;
    showLoading?: boolean;
    loadingMessage?: string;
  }
): Promise<T | null> {
  try {
    const result = await operation();
    
    if (options?.successMessage) {
      showSuccess(options.successMessage);
    }
    
    return result;
  } catch (error) {
    if (options?.errorTitle) {
      showErrorNotification(error, options.errorTitle);
    } else {
      showError(error);
    }
    return null;
  }
}

/**
 * 全局错误处理器（用于 Vue 错误处理）
 */
export function globalErrorHandler(error: any, instance: any, info: string) {
  console.error('[Vue Error]', error, info);
  showError(error);
}

/**
 * 全局未捕获的 Promise 错误处理
 */
export function setupGlobalErrorHandlers() {
  window.addEventListener('unhandledrejection', (event) => {
    console.error('[Unhandled Promise Rejection]', event.reason);
    showError(event.reason);
    event.preventDefault();
  });

  window.addEventListener('error', (event) => {
    // 忽略 null 错误（通常是无害的）
    if (event.error === null) {
      console.warn('[Global Error] Null error caught (usually harmless):', {
        message: event.message,
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno
      });
      return;
    }
    
    console.error('[Global Error]', event.error);
    showError(event.error);
  });
}
