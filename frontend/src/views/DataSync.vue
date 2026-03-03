<template>
  <div class="data-sync">
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <span>数据同步配置</span>
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
                v-model="syncConfig.sourceTable"
                placeholder="选择要同步的表"
                style="width: 100%"
                filterable
                :disabled="!syncConfig.sourceDatabase || loadingSourceTables"
                :loading="loadingSourceTables"
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
                @change="handleTargetDatabaseChange"
              >
                <el-option
                  v-for="db in targetDatabases"
                  :key="db.name"
                  :label="db.name"
                  :value="db.name"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="目标表">
              <el-select
                v-model="syncConfig.targetTable"
                placeholder="选择目标表"
                style="width: 100%"
                filterable
                :disabled="!syncConfig.targetDatabase || loadingTargetTables"
                :loading="loadingTargetTables"
              >
                <el-option
                  v-for="table in targetTables"
                  :key="table.name"
                  :label="table.name"
                  :value="table.name"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item>
          <el-button
            type="primary"
            :disabled="!canSync"
            :loading="syncing"
            @click="handleSync"
          >
            <el-icon><Connection /></el-icon>
            开始同步数据
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-alert
      type="info"
      :closable="false"
      show-icon
      style="margin-top: 20px"
    >
      <template #title>
        <strong>数据同步说明</strong>
      </template>
      <p>数据同步功能将源表的数据复制到目标表中。</p>
      <ul>
        <li>目标表必须已存在且结构兼容</li>
        <li>同步过程会先清空目标表，然后插入源表数据</li>
        <li>请谨慎操作，建议先备份目标表数据</li>
      </ul>
    </el-alert>

    <!-- 同步进度对话框 -->
    <el-dialog
      v-model="showProgressDialog"
      title="数据同步进度"
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
import { Connection } from '@element-plus/icons-vue';
import { useConnectionStore } from '../stores/connection';
import { DatabaseAPI, DataSyncAPI } from '../api';
import type { Database, Table } from '../types/api';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const connectionStore = useConnectionStore();

// 同步配置
const syncConfig = ref({
  sourceProfileId: '',
  sourceDatabase: '',
  sourceTable: '',
  targetProfileId: '',
  targetDatabase: '',
  targetTable: '',
});

// 数据库和表列表
const sourceDatabases = ref<Database[]>([]);
const targetDatabases = ref<Database[]>([]);
const sourceTables = ref<Table[]>([]);
const targetTables = ref<Table[]>([]);

// 加载状态
const loadingSourceDatabases = ref(false);
const loadingTargetDatabases = ref(false);
const loadingSourceTables = ref(false);
const loadingTargetTables = ref(false);
const syncing = ref(false);

// 对话框状态
const showProgressDialog = ref(false);
const syncProgress = ref(0);
const syncStatus = ref<'success' | 'exception' | 'warning' | undefined>(undefined);
const syncProgressText = ref('');

// 计算属性
const canSync = computed(() => {
  return (
    syncConfig.value.sourceProfileId &&
    syncConfig.value.sourceDatabase &&
    syncConfig.value.sourceTable &&
    syncConfig.value.targetProfileId &&
    syncConfig.value.targetDatabase &&
    syncConfig.value.targetTable
  );
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

// 加载目标表列表
async function loadTargetTables() {
  if (!syncConfig.value.targetProfileId || !syncConfig.value.targetDatabase) return;
  
  loadingTargetTables.value = true;
  try {
    targetTables.value = await DatabaseAPI.listTables(
      syncConfig.value.targetProfileId,
      syncConfig.value.targetDatabase
    );
  } catch (error: any) {
    ElMessage.error(`加载目标表列表失败: ${error.message || error}`);
  } finally {
    loadingTargetTables.value = false;
  }
}

// 处理源连接变化
function handleSourceProfileChange() {
  syncConfig.value.sourceDatabase = '';
  syncConfig.value.sourceTable = '';
  sourceDatabases.value = [];
  sourceTables.value = [];
  loadSourceDatabases();
}

// 处理源数据库变化
function handleSourceDatabaseChange() {
  syncConfig.value.sourceTable = '';
  sourceTables.value = [];
  loadSourceTables();
}

// 处理目标连接变化
function handleTargetProfileChange() {
  syncConfig.value.targetDatabase = '';
  syncConfig.value.targetTable = '';
  targetDatabases.value = [];
  targetTables.value = [];
  loadTargetDatabases();
}

// 处理目标数据库变化
function handleTargetDatabaseChange() {
  syncConfig.value.targetTable = '';
  targetTables.value = [];
  loadTargetTables();
}

// 执行数据同步
async function handleSync() {
  try {
    await ElMessageBox.confirm(
      `确定要将 "${syncConfig.value.sourceDatabase}.${syncConfig.value.sourceTable}" 的数据同步到 "${syncConfig.value.targetDatabase}.${syncConfig.value.targetTable}" 吗？\n\n警告：目标表的现有数据将被清空！`,
      '确认数据同步',
      {
        confirmButtonText: '确认同步',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    );

    showProgressDialog.value = true;
    syncing.value = true;
    syncProgress.value = 0;
    syncStatus.value = undefined;
    syncProgressText.value = '正在同步数据...';

    // 调用后端 API 进行数据同步
    await DataSyncAPI.syncTableData(
      syncConfig.value.sourceProfileId,
      syncConfig.value.targetProfileId,
      syncConfig.value.sourceDatabase,
      syncConfig.value.sourceTable,
      syncConfig.value.targetDatabase,
      syncConfig.value.targetTable
    );

    syncProgress.value = 100;
    syncStatus.value = 'success';
    syncProgressText.value = '数据同步完成！';

    setTimeout(() => {
      showProgressDialog.value = false;
      ElMessage.success('数据同步成功');
    }, 2000);

  } catch (error: any) {
    if (error !== 'cancel') {
      syncStatus.value = 'exception';
      syncProgressText.value = `同步失败: ${error.message || error}`;
      ElMessage.error(`数据同步失败: ${error.message || error}`);
    } else {
      showProgressDialog.value = false;
    }
  } finally {
    syncing.value = false;
  }
}

// 监听数据同步进度事件
onMounted(() => {
  EventsOn('data-sync:progress', (data: any) => {
    if (
      data.sourceProfileId === syncConfig.value.sourceProfileId &&
      data.targetProfileId === syncConfig.value.targetProfileId &&
      data.sourceDatabase === syncConfig.value.sourceDatabase &&
      data.sourceTable === syncConfig.value.sourceTable
    ) {
      syncProgressText.value = `已同步 ${data.rowsProcessed} 行数据...`;
      // 由于不知道总行数，显示一个动态进度
      syncProgress.value = Math.min(90, syncProgress.value + 5);
    }
  });

  EventsOn('data-sync:completed', (data: any) => {
    if (
      data.sourceProfileId === syncConfig.value.sourceProfileId &&
      data.targetProfileId === syncConfig.value.targetProfileId &&
      data.sourceDatabase === syncConfig.value.sourceDatabase &&
      data.sourceTable === syncConfig.value.sourceTable
    ) {
      syncProgress.value = 100;
      syncStatus.value = 'success';
      syncProgressText.value = `同步完成！共同步 ${data.rowsSynced} 行数据`;
    }
  });
});
</script>

<style scoped>
.data-sync {
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
</style>
