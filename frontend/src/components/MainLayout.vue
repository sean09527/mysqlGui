<template>
  <el-container class="main-layout">
    <el-aside width="200px" class="sidebar">
      <div class="logo">
        <h2>MySQL 管理工具</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        @select="handleMenuSelect"
      >
        <el-menu-item index="/connections">
          <el-icon><Connection /></el-icon>
          <span>连接管理</span>
        </el-menu-item>
        <el-menu-item index="/workspace" :disabled="!isConnected">
          <el-icon><FolderOpened /></el-icon>
          <span>数据库工作台</span>
        </el-menu-item>
        <el-menu-item index="/query" :disabled="!isConnected">
          <el-icon><EditPen /></el-icon>
          <span>SQL 查询</span>
        </el-menu-item>
        <el-menu-item index="/sync" :disabled="!isConnected">
          <el-icon><Refresh /></el-icon>
          <span>结构同步</span>
        </el-menu-item>
        <el-menu-item index="/data-sync" :disabled="!isConnected">
          <el-icon><Connection /></el-icon>
          <span>数据同步</span>
        </el-menu-item>
        <el-menu-item index="/logs">
          <el-icon><Document /></el-icon>
          <span>日志查看</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <span v-if="currentConnection" class="connection-info">
            <el-icon><Link /></el-icon>
            {{ currentConnection.name }} ({{ currentConnection.host }}:{{ currentConnection.port }})
          </span>
          <span v-else class="connection-info">
            <el-icon><WarningFilled /></el-icon>
            未连接
          </span>
        </div>
        <div class="header-right">
          <span v-if="currentDatabase" class="database-info">
            <el-icon><Coin /></el-icon>
            {{ currentDatabase }}
          </span>
          <span v-if="currentTable" class="table-info">
            <el-icon><Document /></el-icon>
            {{ currentTable }}
          </span>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useConnectionStore } from '../stores/connection';
import { useDatabaseStore } from '../stores/database';
import {
  Connection,
  FolderOpened,
  Document,
  Grid,
  EditPen,
  Refresh,
  Upload,
  Link,
  WarningFilled,
  Coin,
} from '@element-plus/icons-vue';

const router = useRouter();
const route = useRoute();
const connectionStore = useConnectionStore();
const databaseStore = useDatabaseStore();

const isConnected = computed(() => connectionStore.isConnected);
const currentConnection = computed(() => connectionStore.currentConnection);
const currentDatabase = computed(() => databaseStore.currentDatabase);
const currentTable = computed(() => databaseStore.currentTable);
const activeMenu = computed(() => route.path);

const handleMenuSelect = (index: string) => {
  router.push(index);
};
</script>

<style scoped>
.main-layout {
  height: 100vh;
  width: 100vw;
}

.sidebar {
  background-color: #304156;
  color: #fff;
  overflow-y: auto;
}

.logo {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid #4a5568;
}

.logo h2 {
  margin: 0;
  font-size: 18px;
  color: #fff;
}

.sidebar-menu {
  border: none;
  background-color: #304156;
}

.sidebar-menu :deep(.el-menu-item) {
  color: #bfcbd9;
}

.sidebar-menu :deep(.el-menu-item:hover) {
  background-color: #263445 !important;
  color: #fff;
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background-color: #409eff !important;
  color: #fff;
}

.sidebar-menu :deep(.el-menu-item.is-disabled) {
  color: #6b7280;
  opacity: 0.5;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left,
.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.connection-info,
.database-info,
.table-info {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 14px;
  color: #606266;
}

.table-info {
  margin-left: 10px;
  padding-left: 10px;
  border-left: 1px solid #e5e7eb;
}

.main-content {
  background-color: #f5f5f5;
  overflow: auto;
}
</style>
