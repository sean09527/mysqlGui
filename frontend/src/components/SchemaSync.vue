<template>
  <div class="schema-sync">
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <span>结构同步配置</span>
        </div>
      </template>

      <el-form :model="syncConfig" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="源数据库">
              <el-select
                v-model="syncConfig.sourceProfileId"
                placeholder="选择源连接"
                style="width: 100%"
                @change="handleSourceProfileChange"
              >
                <el-option
                  v-for="profile in connectionStore.profiles"
                  :key="profile.id"
                  :label="profile.name"
                  :value="profile.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="源库名">
              <el-select
                v-model="syncConfig.sourceDatabase"
                placeholder="选择源数据库"
                style="width: 100%"
                :disabled="!syncConfig.sourceProfileId || loadingSourceDatabases"
                :loading="loadingSourceDatabases"
                @change="handleSourceDatabaseChange"
              >
                <el-option
                  v-for="db in sourceDatabases"
                  :key="db.name"
                  :label="db.name"
                  :value="db.name"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="源表">
              <el-select
                v-model="syncConfig.sourceTables"
                placeholder="选择要同步的表（留空则同步全部）"
                style="width: 100%"
                multiple
                filterable
                :disabled="!syncConfig.sourceDatabase || loadingSourceTables"
                :loading="loadingSourceTables"
                clearable
              >
                <el-option
                  v-for="table in sourceTables"
                  :key="table.name"
                  :label="table.name"
                  :value="table.name"
                />
              </el-select>
            </el-form-item>
          </el-col>

          <el-col :span="12">
            <el-form-item label="目标数据库">
              <el-select
                v-model="syncConfig.targetProfileId"
                placeholder="选择目标连接"
                style="width: 100%"
                @change="handleTargetProfileChange"
              >
                <el-option
                  v-for="profile in connectionStore.profiles"
                  :key="profile.id"
                  :label="profile.name"
                  :value="profile.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="目标库名">
              <el-select
                v-model="syncConfig.targetDatabase"
                placeholder="选择目标数据库"
                style="width: 100%"
                :disabled="!syncConfig.targetProfileId || loadingTargetDatabases"
                :loading="loadingTargetDatabases"
              >
                <el-option
                  v-for="db in targetDatabases"
                  :key="db.name"
                  :label="db.name"
                  :value="db.name"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="">
              <el-alert
                v-if="syncConfig.sourceTables && syncConfig.sourceTables.length > 0"
                :title="`已选择 ${syncConfig.sourceTables.length} 个表进行同步`"
                type="info"
                :closable="false"
                show-icon
              />
              <el-alert
                v-else
                title="未选择表，将同步整个数据库"
                type="warning"
                :closable="false"
                show-icon
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item>
          <el-button
            type="primary"
            :disabled="!canCompare"
            :loading="comparing"
            @click="handleCompare"
          >
            <el-icon><Connection /></el-icon>
            比较结构
          </el-button>
          <el-button
            v-if="diff"
            type="success"
            :disabled="!diff || generating"
            :loading="generating"
            @click="handleGenerateScript"
          >
            <el-icon><Document /></el-icon>
            生成同步脚本
          </el-button>
          <el-button
            v-if="script"
            type="warning"
            :disabled="!script || executing"
            :loading="executing"
            @click="handleExecuteSync"
          >
            <el-icon><VideoPlay /></el-icon>
            执行同步
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 差异查看器 -->
    <DiffViewer
      v-if="diff"
      :diff="diff"
      class="diff-viewer"
    />

    <!-- 脚本预览和编辑 -->
    <el-card v-if="script" class="script-card">
      <template #header>
        <div class="card-header">
          <span>同步脚本 ({{ script.statements.length }} 条语句)</span>
          <div class="header-actions">
            <el-button
              size="small"
              :type="isEditingScript ? 'default' : 'primary'"
              @click="toggleEditMode"
            >
              <el-icon><Edit /></el-icon>
              {{ isEditingScript ? '预览模式' : '编辑模式' }}
            </el-button>
            <el-button
              type="primary"
              size="small"
              @click="handleExportScript"
            >
              <el-icon><Download /></el-icon>
              导出 SQL
            </el-button>
          </div>
        </div>
      </template>

      <div class="script-editor">
        <el-scrollbar v-if="!isEditingScript" height="400px">
          <pre class="sql-script">{{ formattedScript }}</pre>
        </el-scrollbar>
        <el-input
          v-else
          v-model="editableScript"
          type="textarea"
          :rows="20"
          placeholder="编辑 SQL 脚本..."
          class="script-textarea"
        />
      </div>
    </el-card>

    <!-- 执行确认对话框 -->
    <el-dialog
      v-model="showExecuteDialog"
      title="确认执行同步"
      width="600px"
    >
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      >
        <template #title>
          <strong>警告：此操作将修改目标数据库结构，请确认后再执行！</strong>
        </template>
      </el-alert>

      <div class="execute-summary">
        <h4>将要执行的操作：</h4>
        <ul>
          <li v-for="(stmt, index) in script?.statements" :key="index">
            <el-tag :type="getStatementTagType(stmt.type)" size="small">
              {{ stmt.type }}
            </el-tag>
            {{ stmt.description }}
          </li>
        </ul>
      </div>

      <template #footer>
        <el-button @click="showExecuteDialog = false">取消</el-button>
        <el-button
          type="danger"
          :loading="executing"
          @click="confirmExecuteSync"
        >
          确认执行
        </el-button>
      </template>
    </el-dialog>

    <!-- 同步进度对话框 -->
    <el-dialog
      v-model="showProgressDialog"
      title="同步进度"
      width="500px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <el-progress
        :percentage="syncProgress"
        :status="syncStatus"
      />
      <p style="margin-top: 10px; text-align: center">
        {{ syncProgressText }}
      </p>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Connection, Document, VideoPlay, Download, Edit } from '@element-plus/icons-vue';
import { useConnectionStore } from '../stores/connection';
import { DatabaseAPI, SyncAPI } from '../api';
import DiffViewer from './DiffViewer.vue';
import type { Database, Table, SchemaDiff, SyncScript } from '../types/api';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const connectionStore = useConnectionStore();

// 同步配置
const syncConfig = ref({
  sourceProfileId: '',
  sourceDatabase: '',
  sourceTables: [] as string[],
  targetProfileId: '',
  targetDatabase: '',
});

// 数据库列表
const sourceDatabases = ref<Database[]>([]);
const targetDatabases = ref<Database[]>([]);
const sourceTables = ref<Table[]>([]);
const loadingSourceDatabases = ref(false);
const loadingTargetDatabases = ref(false);
const loadingSourceTables = ref(false);

// 比较结果和脚本
const diff = ref<SchemaDiff | null>(null);
const script = ref<SyncScript | null>(null);

// 加载状态
const comparing = ref(false);
const generating = ref(false);
const executing = ref(false);

// 对话框状态
// 对话框状态
const showExecuteDialog = ref(false);
const showProgressDialog = ref(false);
const syncProgress = ref(0);
const syncStatus = ref<'success' | 'exception' | 'warning' | undefined>(undefined);
const syncProgressText = ref('');

// 脚本编辑状态
const isEditingScript = ref(false);
const editableScript = ref('');

// 计算属性
const canCompare = computed(() => {
  return (
    syncConfig.value.sourceProfileId &&
    syncConfig.value.sourceDatabase &&
    syncConfig.value.targetProfileId &&
    syncConfig.value.targetDatabase
  );
});

const formattedScript = computed(() => {
  if (!script.value) return '';
  return script.value.statements
    .map((stmt, index) => `-- ${index + 1}. ${stmt.description}\n${stmt.sql};`)
    .join('\n\n');
});

// 加载源数据库列表
async function loadSourceDatabases() {
  if (!syncConfig.value.sourceProfileId) return;
  
  loadingSourceDatabases.value = true;
  try {
    sourceDatabases.value = await DatabaseAPI.listDatabases(syncConfig.value.sourceProfileId);
  } catch (error: any) {
    ElMessage.error(`加载源数据库列表失败: ${error.message || error}`);
  } finally {
    loadingSourceDatabases.value = false;
  }
}

// 加载源表列表
async function loadSourceTables() {
  if (!syncConfig.value.sourceProfileId || !syncConfig.value.sourceDatabase) return;
  
  loadingSourceTables.value = true;
  try {
    sourceTables.value = await DatabaseAPI.listTables(
      syncConfig.value.sourceProfileId,
      syncConfig.value.sourceDatabase
    );
  } catch (error: any) {
    ElMessage.error(`加载源表列表失败: ${error.message || error}`);
  } finally {
    loadingSourceTables.value = false;
  }
}

// 加载目标数据库列表
async function loadTargetDatabases() {
  if (!syncConfig.value.targetProfileId) return;
  
  loadingTargetDatabases.value = true;
  try {
    targetDatabases.value = await DatabaseAPI.listDatabases(syncConfig.value.targetProfileId);
  } catch (error: any) {
    ElMessage.error(`加载目标数据库列表失败: ${error.message || error}`);
  } finally {
    loadingTargetDatabases.value = false;
  }
}

// 处理源连接变化
function handleSourceProfileChange() {
  syncConfig.value.sourceDatabase = '';
  syncConfig.value.sourceTables = [];
  sourceDatabases.value = [];
  sourceTables.value = [];
  diff.value = null;
  script.value = null;
  loadSourceDatabases();
}

// 处理源数据库变化
function handleSourceDatabaseChange() {
  syncConfig.value.sourceTables = [];
  sourceTables.value = [];
  diff.value = null;
  script.value = null;
  loadSourceTables();
}

// 处理目标连接变化
function handleTargetProfileChange() {
  syncConfig.value.targetDatabase = '';
  targetDatabases.value = [];
  diff.value = null;
  script.value = null;
  loadTargetDatabases();
}

// 比较结构
async function handleCompare() {
  comparing.value = true;
  diff.value = null;
  script.value = null;

  try {
    diff.value = await SyncAPI.compareSchemas(
      syncConfig.value.sourceProfileId,
      syncConfig.value.targetProfileId,
      syncConfig.value.sourceDatabase,
      syncConfig.value.targetDatabase,
      syncConfig.value.sourceTables.length > 0 ? syncConfig.value.sourceTables : undefined
    );

    const totalDiffs =
      diff.value.tablesOnlyInSource.length +
      diff.value.tablesOnlyInTarget.length +
      diff.value.tableDifferences.length;

    if (totalDiffs === 0) {
      ElMessage.success('两个数据库结构完全一致，无需同步');
    } else {
      const tableInfo = syncConfig.value.sourceTables.length > 0 
        ? `（已选择 ${syncConfig.value.sourceTables.length} 个表）` 
        : '';
      ElMessage.success(`结构比较完成${tableInfo}，发现 ${totalDiffs} 处差异`);
    }
  } catch (error: any) {
    ElMessage.error(`比较结构失败: ${error.message || error}`);
  } finally {
    comparing.value = false;
  }
}

// 生成同步脚本
async function handleGenerateScript() {
  if (!diff.value) return;

  generating.value = true;
  try {
    script.value = await SyncAPI.generateSyncScript(
      syncConfig.value.sourceProfileId,
      syncConfig.value.targetProfileId,
      syncConfig.value.sourceDatabase,
      diff.value
    );

    // 初始化可编辑脚本
    editableScript.value = script.value.statements
      .map((stmt, index) => `-- ${index + 1}. ${stmt.description}\n${stmt.sql};`)
      .join('\n\n');
    isEditingScript.value = false;

    ElMessage.success(`同步脚本生成完成，共 ${script.value.statements.length} 条语句`);
  } catch (error: any) {
    ElMessage.error(`生成同步脚本失败: ${error.message || error}`);
  } finally {
    generating.value = false;
  }
}

// 切换编辑模式
function toggleEditMode() {
  if (!isEditingScript.value) {
    // 切换到编辑模式，初始化可编辑脚本
    editableScript.value = formattedScript.value;
  } else {
    // 切换到预览模式，解析编辑后的脚本
    try {
      parseEditedScript();
      ElMessage.success('脚本已更新');
    } catch (error: any) {
      ElMessage.warning('脚本格式可能不正确，但已保存更改');
    }
  }
  isEditingScript.value = !isEditingScript.value;
}

// 解析编辑后的脚本
function parseEditedScript() {
  if (!script.value) return;
  
  // 简单解析：将编辑后的脚本按语句分割
  const statements = editableScript.value
    .split(';')
    .map(s => s.trim())
    .filter(s => s.length > 0 && !s.startsWith('--'));
  
  // 更新脚本对象（保持原有的描述和类型）
  if (statements.length > 0) {
    script.value.statements = statements.map((sql, index) => {
      const originalStmt = script.value!.statements[index];
      return {
        sql: sql,
        type: originalStmt?.type || 'ALTER',
        description: originalStmt?.description || `语句 ${index + 1}`,
      };
    });
  }
}

// 执行同步
function handleExecuteSync() {
  // 如果在编辑模式，先解析脚本
  if (isEditingScript.value) {
    try {
      parseEditedScript();
    } catch (error: any) {
      ElMessage.error('脚本格式错误，请检查后重试');
      return;
    }
  }
  showExecuteDialog.value = true;
}

// 确认执行同步
async function confirmExecuteSync() {
  if (!script.value) return;

  showExecuteDialog.value = false;
  showProgressDialog.value = true;
  executing.value = true;
  syncProgress.value = 0;
  syncStatus.value = undefined;
  syncProgressText.value = '正在执行同步...';

  try {
    await SyncAPI.executeSyncScript(
      syncConfig.value.sourceProfileId,
      syncConfig.value.targetProfileId,
      syncConfig.value.targetDatabase,
      script.value
    );

    syncProgress.value = 100;
    syncStatus.value = 'success';
    syncProgressText.value = '同步完成！';

    setTimeout(() => {
      showProgressDialog.value = false;
      ElMessage.success('数据库结构同步成功');
      // 重置状态
      diff.value = null;
      script.value = null;
    }, 2000);
  } catch (error: any) {
    syncStatus.value = 'exception';
    syncProgressText.value = `同步失败: ${error.message || error}`;
    ElMessage.error(`执行同步失败: ${error.message || error}`);
  } finally {
    executing.value = false;
  }
}

// 导出脚本
function handleExportScript() {
  if (!script.value) return;

  // 使用编辑模式的脚本或格式化的脚本
  const content = isEditingScript.value ? editableScript.value : formattedScript.value;
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `sync_${syncConfig.value.sourceDatabase}_to_${syncConfig.value.targetDatabase}_${Date.now()}.sql`;
  link.click();
  URL.revokeObjectURL(url);

  ElMessage.success('脚本已导出');
}

// 获取语句标签类型
function getStatementTagType(type: string): 'success' | 'warning' | 'danger' {
  switch (type) {
    case 'CREATE':
      return 'success';
    case 'ALTER':
      return 'warning';
    case 'DROP':
      return 'danger';
    default:
      return 'warning';
  }
}

// 监听同步进度事件
onMounted(() => {
  EventsOn('sync:progress', (data: any) => {
    if (
      data.sourceProfileId === syncConfig.value.sourceProfileId &&
      data.targetProfileId === syncConfig.value.targetProfileId
    ) {
      syncProgress.value = Math.round(data.percentage);
      syncProgressText.value = `正在执行: ${data.statement}`;
    }
  });

  EventsOn('sync:completed', (data: any) => {
    if (
      data.sourceProfileId === syncConfig.value.sourceProfileId &&
      data.targetProfileId === syncConfig.value.targetProfileId
    ) {
      syncProgress.value = 100;
      syncStatus.value = 'success';
      syncProgressText.value = '同步完成！';
    }
  });

  EventsOn('sync:failed', (data: any) => {
    if (
      data.sourceProfileId === syncConfig.value.sourceProfileId &&
      data.targetProfileId === syncConfig.value.targetProfileId
    ) {
      syncStatus.value = 'exception';
      syncProgressText.value = `同步失败: ${data.error}`;
    }
  });
});
</script>

<style scoped>
.schema-sync {
  padding: 20px;
}

.config-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.diff-viewer {
  margin-bottom: 20px;
}

.script-card {
  margin-bottom: 20px;
}

.script-editor {
  background-color: #f5f7fa;
  border-radius: 4px;
  padding: 10px;
}

.sql-script {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.script-textarea {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.script-textarea :deep(textarea) {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.execute-summary {
  max-height: 400px;
  overflow-y: auto;
}

.execute-summary h4 {
  margin-top: 0;
  margin-bottom: 10px;
}

.execute-summary ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.execute-summary li {
  padding: 8px;
  margin-bottom: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
