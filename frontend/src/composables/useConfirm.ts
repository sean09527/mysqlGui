/**
 * 确认对话框 Composable
 * 
 * 提供便捷的确认对话框功能，用于危险操作前的确认
 */

import { ElMessageBox } from 'element-plus';
import type { ElMessageBoxOptions } from 'element-plus';

/**
 * 确认选项
 */
export interface ConfirmOptions {
  title?: string;
  message: string;
  detail?: string;
  type?: 'warning' | 'info' | 'error' | 'success';
  confirmText?: string;
  cancelText?: string;
  dangerouslyUseHTMLString?: boolean;
}

/**
 * 确认对话框 Hook
 */
export function useConfirm() {
  /**
   * 显示确认对话框
   */
  async function confirm(options: ConfirmOptions): Promise<boolean> {
    const {
      title = '确认操作',
      message,
      detail,
      type = 'warning',
      confirmText = '确认',
      cancelText = '取消',
      dangerouslyUseHTMLString = false,
    } = options;

    // 构建消息内容
    let content = message;
    if (detail) {
      content = dangerouslyUseHTMLString
        ? `<p style="font-weight: 500; margin-bottom: 8px;">${message}</p><p style="color: #909399; font-size: 14px;">${detail}</p>`
        : `${message}\n\n${detail}`;
    }

    const messageBoxOptions: ElMessageBoxOptions = {
      title,
      message: content,
      type,
      confirmButtonText: confirmText,
      cancelButtonText: cancelText,
      showCancelButton: true,
      closeOnClickModal: false,
      closeOnPressEscape: true,
      dangerouslyUseHTMLString,
      distinguishCancelAndClose: true,
    };

    try {
      await ElMessageBox.confirm(content, title, messageBoxOptions);
      return true;
    } catch (action) {
      // 用户取消或关闭对话框
      return false;
    }
  }

  /**
   * 删除确认（预设的危险操作确认）
   */
  async function confirmDelete(
    itemName: string,
    itemType: string = '项目'
  ): Promise<boolean> {
    return confirm({
      title: '确认删除',
      message: `确定要删除${itemType} "${itemName}" 吗？`,
      detail: '此操作不可撤销，请谨慎操作。',
      type: 'warning',
      confirmText: '删除',
      cancelText: '取消',
    });
  }

  /**
   * 删除表确认
   */
  async function confirmDropTable(tableName: string): Promise<boolean> {
    return confirm({
      title: '确认删除表',
      message: `确定要删除表 "${tableName}" 吗？`,
      detail: '删除表将永久删除表结构和所有数据，此操作不可撤销！',
      type: 'error',
      confirmText: '删除表',
      cancelText: '取消',
    });
  }

  /**
   * 删除数据确认
   */
  async function confirmDeleteData(count: number = 1): Promise<boolean> {
    return confirm({
      title: '确认删除数据',
      message: `确定要删除选中的 ${count} 条数据吗？`,
      detail: '删除的数据无法恢复，请谨慎操作。',
      type: 'warning',
      confirmText: '删除',
      cancelText: '取消',
    });
  }

  /**
   * 执行同步确认
   */
  async function confirmSync(
    sourceDB: string,
    targetDB: string,
    changeCount: number
  ): Promise<boolean> {
    return confirm({
      title: '确认执行同步',
      message: `确定要将 "${sourceDB}" 的结构同步到 "${targetDB}" 吗？`,
      detail: `此操作将执行 ${changeCount} 个结构变更，可能会影响目标数据库的数据。`,
      type: 'warning',
      confirmText: '执行同步',
      cancelText: '取消',
    });
  }

  /**
   * 修改表结构确认（可能导致数据丢失）
   */
  async function confirmAlterTable(
    tableName: string,
    hasDataLoss: boolean = false
  ): Promise<boolean> {
    return confirm({
      title: '确认修改表结构',
      message: `确定要修改表 "${tableName}" 的结构吗？`,
      detail: hasDataLoss
        ? '此操作可能导致数据丢失或转换错误，请确保已备份重要数据。'
        : '修改表结构可能会影响现有数据，请谨慎操作。',
      type: hasDataLoss ? 'error' : 'warning',
      confirmText: '确认修改',
      cancelText: '取消',
    });
  }

  /**
   * 导入数据确认
   */
  async function confirmImport(
    fileName: string,
    rowCount?: number
  ): Promise<boolean> {
    const detail = rowCount
      ? `将导入约 ${rowCount} 条数据，可能会覆盖或新增数据。`
      : '导入操作可能会覆盖或新增数据。';

    return confirm({
      title: '确认导入数据',
      message: `确定要导入文件 "${fileName}" 吗？`,
      detail,
      type: 'info',
      confirmText: '开始导入',
      cancelText: '取消',
    });
  }

  /**
   * 断开连接确认
   */
  async function confirmDisconnect(connectionName: string): Promise<boolean> {
    return confirm({
      title: '确认断开连接',
      message: `确定要断开与 "${connectionName}" 的连接吗？`,
      detail: '断开连接后，当前所有未保存的操作将丢失。',
      type: 'info',
      confirmText: '断开连接',
      cancelText: '取消',
    });
  }

  /**
   * 取消操作确认
   */
  async function confirmCancel(operationName: string): Promise<boolean> {
    return confirm({
      title: '确认取消',
      message: `确定要取消 "${operationName}" 操作吗？`,
      detail: '取消后，已执行的部分可能无法恢复。',
      type: 'warning',
      confirmText: '确认取消',
      cancelText: '继续操作',
    });
  }

  /**
   * 覆盖文件确认
   */
  async function confirmOverwrite(fileName: string): Promise<boolean> {
    return confirm({
      title: '确认覆盖文件',
      message: `文件 "${fileName}" 已存在，是否覆盖？`,
      detail: '覆盖后，原文件内容将丢失。',
      type: 'warning',
      confirmText: '覆盖',
      cancelText: '取消',
    });
  }

  return {
    confirm,
    confirmDelete,
    confirmDropTable,
    confirmDeleteData,
    confirmSync,
    confirmAlterTable,
    confirmImport,
    confirmDisconnect,
    confirmCancel,
    confirmOverwrite,
  };
}
