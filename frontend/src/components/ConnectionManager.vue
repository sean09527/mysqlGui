<template>
  <div class="connection-manager">
    <div class="header">
      <h2>连接管理</h2>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        新建连接
      </el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="connectionStore.profiles || []"
      stripe
      style="width: 100%"
      class="connection-table"
    >
      <el-table-column prop="name" label="名称" width="200" />
      <el-table-column label="连接信息" min-width="300">
        <template #default="{ row }">
          <span>{{ row.username }}@{{ row.host }}:{{ row.port }}</span>
          <span v-if="row.database" class="database-name">/ {{ row.database }}</span>
        </template>
      </el-table-column>
      <el-table-column label="SSH 隧道" width="100" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.sshEnabled" type="success" size="small">已启用</el-tag>
          <el-tag v-else type="info" size="small">未启用</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag
            v-if="isCurrentConnection(row.id)"
            type="success"
            size="small"
          >
            已连接
          </el-tag>
          <el-tag v-else type="info" size="small">未连接</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" align="center" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="!isCurrentConnection(row.id)"
            type="primary"
            size="small"
            @click="handleConnect(row)"
          >
            连接
          </el-button>
          <el-button
            v-else
            type="warning"
            size="small"
            @click="handleDisconnect(row)"
          >
            断开
          </el-button>
          <el-button
            type="info"
            size="small"
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          <el-button
            type="danger"
            size="small"
            @click="handleDelete(row)"
            :disabled="isCurrentConnection(row.id)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty
      v-if="!loading && (!connectionStore.profiles || connectionStore.profiles.length === 0)"
      description="暂无连接配置，请创建新连接"
    />

    <!-- 连接配置表单对话框 -->
    <ConnectionForm
      v-model:visible="formVisible"
      :profile="currentProfile"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import { useConnectionStore } from '../stores/connection';
import { ConnectionAPI } from '../api';
import ConnectionForm from './ConnectionForm.vue';
import type { ConnectionProfile } from '../types/api';

const connectionStore = useConnectionStore();
const loading = ref(false);
const formVisible = ref(false);
const currentProfile = ref<ConnectionProfile | null>(null);

const isCurrentConnection = (id: string) => {
  return connectionStore.currentConnection?.id === id;
};

const loadProfiles = async () => {
  loading.value = true;
  try {
    const profiles = await ConnectionAPI.listProfiles();
    // Handle null or undefined response
    connectionStore.setProfiles(profiles || []);
  } catch (error: any) {
    ElMessage.error(`加载连接配置失败: ${error.message || error}`);
    // Set empty array on error
    connectionStore.setProfiles([]);
  } finally {
    loading.value = false;
  }
};

const handleCreate = () => {
  currentProfile.value = null;
  formVisible.value = true;
};

const handleEdit = (profile: ConnectionProfile) => {
  currentProfile.value = profile;
  formVisible.value = true;
};

const handleFormSuccess = () => {
  // 刷新连接列表
  loadProfiles();
};

const handleConnect = async (profile: ConnectionProfile) => {
  loading.value = true;
  try {
    await ConnectionAPI.connect(profile.id);
    connectionStore.setCurrentConnection(profile);
    ElMessage.success(`已连接到 ${profile.name}`);
  } catch (error: any) {
    ElMessage.error(`连接失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

const handleDisconnect = async (profile: ConnectionProfile) => {
  loading.value = true;
  try {
    await ConnectionAPI.disconnect(profile.id);
    connectionStore.setCurrentConnection(null);
    ElMessage.success(`已断开连接 ${profile.name}`);
  } catch (error: any) {
    ElMessage.error(`断开连接失败: ${error.message || error}`);
  } finally {
    loading.value = false;
  }
};

const handleDelete = async (profile: ConnectionProfile) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除连接配置 "${profile.name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    );

    loading.value = true;
    await ConnectionAPI.deleteProfile(profile.id);
    connectionStore.removeProfile(profile.id);
    ElMessage.success('连接配置已删除');
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(`删除失败: ${error.message || error}`);
    }
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadProfiles();
});
</script>

<style scoped>
.connection-manager {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.connection-table {
  background-color: #fff;
  border-radius: 4px;
}

.database-name {
  color: #909399;
  margin-left: 5px;
}
</style>
