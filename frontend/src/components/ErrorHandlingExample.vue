<template>
  <div class="error-handling-example">
    <el-card header="错误处理和用户反馈示例">
      <el-space direction="vertical" :size="20" style="width: 100%">
        <!-- 错误处理示例 -->
        <el-card shadow="never">
          <template #header>
            <span>错误处理示例</span>
          </template>
          <el-space wrap>
            <el-button @click="showSuccessMessage">成功消息</el-button>
            <el-button @click="showWarningMessage">警告消息</el-button>
            <el-button @click="showInfoMessage">信息消息</el-button>
            <el-button @click="showErrorMessage">错误消息</el-button>
            <el-button @click="showErrorNotificationExample">错误通知</el-button>
            <el-button @click="simulateConnectionError">模拟连接错误</el-button>
            <el-button @click="simulateSQLError">模拟 SQL 错误</el-button>
          </el-space>
        </el-card>

        <!-- 加载状态示例 -->
        <el-card shadow="never">
          <template #header>
            <span>加载状态示例</span>
          </template>
          <el-space wrap>
            <el-button
              :loading="localLoading"
              @click="simulateLocalLoading"
            >
              局部加载 (3秒)
            </el-button>
            <el-button @click="simulateGlobalLoading">
              全屏加载 (3秒)
            </el-button>
            <el-button @click="simulateProgressLoading">
              进度加载 (5秒)
            </el-button>
            <el-button @click="simulateCancelableLoading">
              可取消加载 (10秒)
            </el-button>
          </el-space>
        </el-card>

        <!-- 确认对话框示例 -->
        <el-card shadow="never">
          <template #header>
            <span>确认对话框示例</span>
          </template>
          <el-space wrap>
            <el-button @click="showBasicConfirm">基本确认</el-button>
            <el-button @click="showDeleteConfirm">删除确认</el-button>
            <el-button @click="showDropTableConfirm">删除表确认</el-button>
            <el-button @click="showSyncConfirm">同步确认</el-button>
            <el-button @click="showAlterTableConfirm">修改表确认</el-button>
          </el-space>
        </el-card>

        <!-- 组合示例 -->
        <el-card shadow="never">
          <template #header>
            <span>完整流程示例</span>
          </template>
          <el-space wrap>
            <el-button type="primary" @click="simulateDeleteOperation">
              模拟删除操作（完整流程）
            </el-button>
            <el-button type="primary" @click="simulateImportOperation">
              模拟导入操作（带进度）
            </el-button>
          </el-space>
        </el-card>

        <!-- 操作日志 -->
        <el-card shadow="never">
          <template #header>
            <span>操作日志</span>
            <el-button
              size="small"
              style="float: right"
              @click="logs = []"
            >
              清空
            </el-button>
          </template>
          <div class="logs">
            <div v-for="(log, index) in logs" :key="index" class="log-item">
              <el-tag :type="log.type" size="small">{{ log.time }}</el-tag>
              <span>{{ log.message }}</span>
            </div>
            <div v-if="logs.length === 0" class="empty-logs">
              暂无日志
            </div>
          </div>
        </el-card>
      </el-space>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import {
  showError,
  showSuccess,
  showWarning,
  showInfo,
  showErrorNotification,
} from '@/utils/errorHandler';
import { useLoading, useLocalLoading } from '@/composables/useLoading';
import { useConfirm } from '@/composables/useConfirm';

// 日志
interface Log {
  time: string;
  message: string;
  type: 'success' | 'warning' | 'info' | 'danger';
}

const logs = ref<Log[]>([]);

function addLog(message: string, type: Log['type'] = 'info') {
  const time = new Date().toLocaleTimeString();
  logs.value.unshift({ time, message, type });
  if (logs.value.length > 20) {
    logs.value.pop();
  }
}

// 局部加载
const { loading: localLoading, withLocalLoading } = useLocalLoading();

// 全局加载
const { withLoading } = useLoading();

// 确认对话框
const {
  confirm,
  confirmDelete,
  confirmDropTable,
  confirmSync,
  confirmAlterTable,
} = useConfirm();

// ===== 错误处理示例 =====

function showSuccessMessage() {
  showSuccess('操作成功！');
  addLog('显示成功消息', 'success');
}

function showWarningMessage() {
  showWarning('请注意数据格式');
  addLog('显示警告消息', 'warning');
}

function showInfoMessage() {
  showInfo('正在处理中...');
  addLog('显示信息消息', 'info');
}

function showErrorMessage() {
  showError('操作失败，请重试');
  addLog('显示错误消息', 'danger');
}

function showErrorNotificationExample() {
  showErrorNotification(
    new Error('这是一个详细的错误信息，包含更多上下文'),
    '操作失败'
  );
  addLog('显示错误通知', 'danger');
}

function simulateConnectionError() {
  const error = new Error('Connection refused: Unable to connect to MySQL server at localhost:3306');
  showError(error);
  addLog('模拟连接错误', 'danger');
}

function simulateSQLError() {
  const error = new Error("SQL syntax error: You have an error in your SQL syntax near 'SELCT' at line 1");
  showError(error);
  addLog('模拟 SQL 错误', 'danger');
}

// ===== 加载状态示例 =====

async function simulateLocalLoading() {
  addLog('开始局部加载', 'info');
  await withLocalLoading(async () => {
    await new Promise(resolve => setTimeout(resolve, 3000));
  }, '加载中...');
  addLog('局部加载完成', 'success');
}

async function simulateGlobalLoading() {
  addLog('开始全屏加载', 'info');
  try {
    await withLoading(
      async () => {
        await new Promise(resolve => setTimeout(resolve, 3000));
      },
      {
        message: '正在处理数据',
        successMessage: '处理完成',
        showFullscreen: true,
      }
    );
    addLog('全屏加载完成', 'success');
  } catch (error) {
    addLog('全屏加载失败', 'danger');
  }
}

async function simulateProgressLoading() {
  addLog('开始进度加载', 'info');
  try {
    await withLoading(
      async (updateProgress) => {
        for (let i = 0; i <= 100; i += 10) {
          await new Promise(resolve => setTimeout(resolve, 500));
          updateProgress(i, `处理中... ${i}%`);
        }
      },
      {
        message: '正在处理',
        successMessage: '处理完成',
        showFullscreen: true,
      }
    );
    addLog('进度加载完成', 'success');
  } catch (error) {
    addLog('进度加载失败', 'danger');
  }
}

async function simulateCancelableLoading() {
  addLog('开始可取消加载', 'info');
  try {
    await withLoading(
      async (updateProgress) => {
        for (let i = 0; i <= 100; i += 5) {
          await new Promise(resolve => setTimeout(resolve, 500));
          updateProgress(i, `处理中... ${i}%`);
        }
      },
      {
        message: '正在处理（可取消）',
        successMessage: '处理完成',
        cancelable: true,
        showFullscreen: true,
      }
    );
    addLog('可取消加载完成', 'success');
  } catch (error) {
    addLog('可取消加载被取消或失败', 'warning');
  }
}

// ===== 确认对话框示例 =====

async function showBasicConfirm() {
  const confirmed = await confirm({
    title: '确认操作',
    message: '确定要执行此操作吗？',
    detail: '此操作可能会影响数据。',
    type: 'warning',
  });
  addLog(`基本确认: ${confirmed ? '已确认' : '已取消'}`, confirmed ? 'success' : 'warning');
}

async function showDeleteConfirm() {
  const confirmed = await confirmDelete('测试项目', '项目');
  addLog(`删除确认: ${confirmed ? '已确认' : '已取消'}`, confirmed ? 'success' : 'warning');
}

async function showDropTableConfirm() {
  const confirmed = await confirmDropTable('users');
  addLog(`删除表确认: ${confirmed ? '已确认' : '已取消'}`, confirmed ? 'success' : 'warning');
}

async function showSyncConfirm() {
  const confirmed = await confirmSync('source_db', 'target_db', 15);
  addLog(`同步确认: ${confirmed ? '已确认' : '已取消'}`, confirmed ? 'success' : 'warning');
}

async function showAlterTableConfirm() {
  const confirmed = await confirmAlterTable('users', true);
  addLog(`修改表确认: ${confirmed ? '已确认' : '已取消'}`, confirmed ? 'success' : 'warning');
}

// ===== 完整流程示例 =====

async function simulateDeleteOperation() {
  addLog('开始删除操作流程', 'info');

  // 1. 确认
  const confirmed = await confirmDelete('测试数据', '数据');
  if (!confirmed) {
    addLog('用户取消删除', 'warning');
    return;
  }

  // 2. 执行删除
  try {
    await withLoading(
      async () => {
        await new Promise(resolve => setTimeout(resolve, 2000));
        // 模拟可能的错误
        if (Math.random() > 0.7) {
          throw new Error('删除失败：数据被其他进程锁定');
        }
      },
      {
        message: '正在删除数据',
        showFullscreen: true,
      }
    );
    showSuccess('数据删除成功');
    addLog('删除操作完成', 'success');
  } catch (error) {
    showError(error);
    addLog('删除操作失败', 'danger');
  }
}

async function simulateImportOperation() {
  addLog('开始导入操作流程', 'info');

  try {
    await withLoading(
      async (updateProgress) => {
        const totalRows = 1000;
        const batchSize = 100;
        let processed = 0;

        while (processed < totalRows) {
          await new Promise(resolve => setTimeout(resolve, 500));
          processed += batchSize;
          const progress = Math.min((processed / totalRows) * 100, 100);
          updateProgress(progress, `已导入 ${processed}/${totalRows} 行`);
        }
      },
      {
        message: '正在导入数据',
        successMessage: '数据导入完成',
        cancelable: true,
        showFullscreen: true,
      }
    );
    addLog('导入操作完成', 'success');
  } catch (error) {
    showError(error);
    addLog('导入操作失败', 'danger');
  }
}
</script>

<style scoped lang="scss">
.error-handling-example {
  padding: 20px;

  .logs {
    max-height: 300px;
    overflow-y: auto;

    .log-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 8px;
      border-bottom: 1px solid var(--el-border-color-lighter);

      &:last-child {
        border-bottom: none;
      }

      span {
        flex: 1;
        font-size: 14px;
        color: var(--el-text-color-regular);
      }
    }

    .empty-logs {
      text-align: center;
      padding: 40px;
      color: var(--el-text-color-placeholder);
    }
  }
}
</style>
