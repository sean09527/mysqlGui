<template>
  <div class="database-explorer">
    <div class="explorer-header">
      <h3>数据库浏览器</h3>
      <el-button 
        :icon="Refresh" 
        circle 
        size="small" 
        @click="refreshDatabases"
        :loading="loading"
        title="刷新"
      />
    </div>

    <div class="explorer-content">
      <el-tree
        v-if="treeData.length > 0"
        :data="treeData"
        :props="treeProps"
        node-key="id"
        :expand-on-click-node="false"
        :lazy="true"
        :load="loadNode"
        @node-click="handleNodeClick"
        @node-contextmenu="handleContextMenu"
        :default-expanded-keys="expandedKeys"
      >
        <template #default="{ node, data }">
          <span class="tree-node">
            <el-icon class="node-icon">
              <component :is="getNodeIcon(data.type)" />
            </el-icon>
            <span class="node-label">{{ node.label }}</span>
            <span v-if="data.type === 'table' && data.rows !== undefined" class="node-meta">
              ({{ formatRowCount(data.rows) }} 行)
            </span>
            <span v-if="data.type === 'database' && data.loading" class="node-loading">
              <el-icon class="is-loading"><Loading /></el-icon>
            </span>
          </span>
        </template>
      </el-tree>

      <el-empty 
        v-else-if="!loading && !connectionStore.isConnected"
        description="请先连接到数据库"
      />

      <el-empty 
        v-else-if="!loading"
        description="未找到数据库"
      />
    </div>

    <!-- 右键菜单 -->
    <teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="context-menu"
        :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
        @click="handleMenuClick"
      >
        <template v-if="contextMenuData?.type === 'database'">
          <div class="context-menu-item" @click="handleMenuCommand('create-table')">
            <el-icon><Plus /></el-icon>
            <span>新建表</span>
          </div>
          <div class="context-menu-divider"></div>
          <div class="context-menu-item" @click="handleMenuCommand('refresh-tables')">
            <el-icon><Refresh /></el-icon>
            <span>刷新表列表</span>
          </div>
        </template>
        <template v-else-if="contextMenuData?.type === 'table'">
          <div class="context-menu-item" @click="handleMenuCommand('view-data')">
            <el-icon><Grid /></el-icon>
            <span>查看表数据</span>
          </div>
          <div class="context-menu-item" @click="handleMenuCommand('view-schema')">
            <el-icon><Document /></el-icon>
            <span>查看表结构</span>
          </div>
          <div class="context-menu-item" @click="handleMenuCommand('edit-schema')">
            <el-icon><Edit /></el-icon>
            <span>修改表结构</span>
          </div>
          <div class="context-menu-divider"></div>
          <div class="context-menu-item danger" @click="handleMenuCommand('drop-table')">
            <el-icon><Delete /></el-icon>
            <span>删除表</span>
          </div>
        </template>
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { 
  Refresh, 
  Coin as DatabaseIcon, 
  Document, 
  Grid,
  Delete,
  Folder,
  Edit,
  Loading,
  Plus
} from '@element-plus/icons-vue';
import { useConnectionStore } from '../stores/connection';
import { useDatabaseStore } from '../stores/database';
import { DatabaseAPI, SchemaAPI } from '../api';
import type { Database, Table } from '../types/api';

// Emits
const emit = defineEmits<{
  viewSchema: [];
  viewData: [];
  editSchema: [];
}>();

// Stores
const connectionStore = useConnectionStore();
const databaseStore = useDatabaseStore();

// State
const loading = ref(false);
const treeData = ref<any[]>([]);
const expandedKeys = ref<string[]>([]);
const contextMenuVisible = ref(false);
const contextMenuPosition = ref({ x: 0, y: 0 });
const contextMenuData = ref<any>(null);
const lastClickTime = ref(0);
const lastClickNode = ref<any>(null);

// Cache for table lists - expires after 5 minutes
const tableCache = new Map<string, { data: Table[], timestamp: number }>();
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

// Tree props
const treeProps = {
  children: 'children',
  label: 'label',
  isLeaf: 'isLeaf'
};

// 格式化行数显示
const formatRowCount = (count: number): string => {
  if (count >= 1000000) {
    return `${(count / 1000000).toFixed(1)}M`;
  } else if (count >= 1000) {
    return `${(count / 1000).toFixed(1)}K`;
  }
  return count.toString();
};

// 获取节点图标
const getNodeIcon = (type: string) => {
  switch (type) {
    case 'database':
      return DatabaseIcon;
    case 'table':
      return Document;
    default:
      return Folder;
  }
};

// 清除缓存
const clearCache = () => {
  tableCache.clear();
};

// 检查缓存是否有效
const isCacheValid = (timestamp: number): boolean => {
  return Date.now() - timestamp < CACHE_DURATION;
};

// 加载数据库列表
const loadDatabases = async () => {
  if (!connectionStore.currentConnection) {
    treeData.value = [];
    return;
  }

  loading.value = true;
  try {
    const databases = await DatabaseAPI.listDatabases(connectionStore.currentConnection.id);
    databaseStore.setDatabases(databases);
    
    // 构建树形数据 - 使用懒加载，不预加载表列表
    treeData.value = databases.map((db: Database) => ({
      id: `db-${db.name}`,
      label: db.name,
      type: 'database',
      database: db.name,
      isLeaf: false,
      loading: false
    }));
  } catch (error: any) {
    ElMessage.error(error.message || '加载数据库列表失败');
    console.error('Failed to load databases:', error);
  } finally {
    loading.value = false;
  }
};

// 懒加载节点 - Element Plus Tree 的 lazy load 回调
const loadNode = async (node: any, resolve: (data: any[]) => void) => {
  // 只处理数据库节点的懒加载
  if (node.level === 0) {
    // 根节点，返回数据库列表
    resolve(treeData.value);
    return;
  }
  
  if (node.data.type === 'database') {
    // 数据库节点，加载表列表
    const dbName = node.data.database;
    
    // 检查缓存
    const cached = tableCache.get(dbName);
    if (cached && isCacheValid(cached.timestamp)) {
      const tableNodes = cached.data.map((table: Table) => ({
        id: `table-${dbName}-${table.name}`,
        label: table.name,
        type: 'table',
        database: dbName,
        table: table.name,
        rows: table.rows,
        engine: table.engine,
        comment: table.comment,
        isLeaf: true
      }));
      resolve(tableNodes);
      return;
    }
    
    // 从服务器加载
    node.data.loading = true;
    try {
      if (!connectionStore.currentConnection) {
        resolve([]);
        return;
      }

      const tables = await DatabaseAPI.listTables(
        connectionStore.currentConnection.id,
        dbName
      );
      
      // 更新缓存
      tableCache.set(dbName, {
        data: tables,
        timestamp: Date.now()
      });
      
      const tableNodes = tables.map((table: Table) => ({
        id: `table-${dbName}-${table.name}`,
        label: table.name,
        type: 'table',
        database: dbName,
        table: table.name,
        rows: table.rows,
        engine: table.engine,
        comment: table.comment,
        isLeaf: true
      }));
      
      // 如果这是当前数据库，更新 store
      if (dbName === databaseStore.currentDatabase) {
        databaseStore.setTables(tables);
      }
      
      resolve(tableNodes);
    } catch (error: any) {
      ElMessage.error(error.message || '加载表列表失败');
      console.error('Failed to load tables:', error);
      resolve([]);
    } finally {
      node.data.loading = false;
    }
  } else {
    // 表节点是叶子节点
    resolve([]);
  }
};

// 处理节点点击
const handleNodeClick = async (data: any, node: any) => {
  const now = Date.now();
  const isDoubleClick = lastClickNode.value === data.id && (now - lastClickTime.value) < 300;
  
  lastClickTime.value = now;
  lastClickNode.value = data.id;
  
  if (data.type === 'database') {
    // 点击数据库节点 - 自动展开
    databaseStore.setCurrentDatabase(data.database);
    databaseStore.setCurrentTable(null);
    
    // 自动展开数据库节点
    if (!node.expanded) {
      node.expanded = true;
    }
  } else if (data.type === 'table') {
    // 单击表节点 - 只选中，不打开
    databaseStore.setCurrentDatabase(data.database);
    databaseStore.setCurrentTable(data.table);
    
    // 双击表节点 - 打开表数据
    if (isDoubleClick) {
      emit('viewData');
    }
  }
};

// 处理右键菜单
const handleContextMenu = (event: MouseEvent, data: any) => {
  event.preventDefault();
  contextMenuData.value = data;
  contextMenuPosition.value = { x: event.clientX, y: event.clientY };
  contextMenuVisible.value = true;
  
  // 点击其他地方关闭菜单
  const closeMenu = () => {
    contextMenuVisible.value = false;
    document.removeEventListener('click', closeMenu);
  };
  setTimeout(() => {
    document.addEventListener('click', closeMenu);
  }, 100);
};

// 处理菜单点击
const handleMenuClick = (event: Event) => {
  event.stopPropagation();
};

// 处理菜单命令
const handleMenuCommand = async (command: string) => {
  if (!contextMenuData.value) return;

  const data = contextMenuData.value;
  contextMenuVisible.value = false;

  switch (command) {
    case 'create-table':
      // 新建表
      databaseStore.setCurrentDatabase(data.database);
      databaseStore.setCurrentTable(null);
      emit('editSchema'); // 触发编辑模式，但不传表名，表示新建
      break;

    case 'refresh-tables':
      // 清除该数据库的缓存
      tableCache.delete(data.database);
      ElMessage.success('表列表已刷新，请重新展开数据库节点');
      break;

    case 'view-schema':
      databaseStore.setCurrentDatabase(data.database);
      databaseStore.setCurrentTable(data.table);
      emit('viewSchema');
      break;

    case 'edit-schema':
      databaseStore.setCurrentDatabase(data.database);
      databaseStore.setCurrentTable(data.table);
      emit('editSchema');
      break;

    case 'view-data':
      databaseStore.setCurrentDatabase(data.database);
      databaseStore.setCurrentTable(data.table);
      emit('viewData');
      break;

    case 'drop-table':
      await handleDropTable(data);
      break;
  }

  contextMenuData.value = null;
};

// 删除表
const handleDropTable = async (data: any) => {
  if (!connectionStore.currentConnection) return;

  try {
    await ElMessageBox.confirm(
      `确定要删除表 "${data.database}.${data.table}" 吗？此操作不可恢复！`,
      '删除表',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    );

    await SchemaAPI.dropTable(
      connectionStore.currentConnection.id,
      data.database,
      data.table
    );

    ElMessage.success('表已删除');

    // 清除缓存，强制重新加载
    tableCache.delete(data.database);

    // 如果删除的是当前选中的表，清除选中状态
    if (databaseStore.currentTable === data.table) {
      databaseStore.setCurrentTable(null);
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除表失败');
      console.error('Failed to drop table:', error);
    }
  }
};

// 刷新数据库列表
const refreshDatabases = async () => {
  clearCache();
  await loadDatabases();
  ElMessage.success('数据库列表已刷新');
};

// 监听连接状态变化
watch(
  () => connectionStore.currentConnection,
  async (newConnection) => {
    clearCache();
    if (newConnection) {
      await loadDatabases();
    } else {
      treeData.value = [];
      databaseStore.setCurrentDatabase(null);
      databaseStore.setCurrentTable(null);
    }
  }
);

// 组件挂载时加载数据
onMounted(async () => {
  if (connectionStore.currentConnection) {
    await loadDatabases();
  }
});

// 组件卸载时清理缓存
onUnmounted(() => {
  clearCache();
});
</script>

<style scoped>
.database-explorer {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 4px;
  overflow: hidden;
}

.explorer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.explorer-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.explorer-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}

.node-icon {
  font-size: 16px;
  color: #606266;
}

.node-label {
  font-size: 14px;
  color: #303133;
}

.node-meta {
  font-size: 12px;
  color: #909399;
  margin-left: 4px;
}

.node-loading {
  margin-left: 8px;
  color: #409eff;
  font-size: 14px;
}

:deep(.el-tree-node__content) {
  height: 32px;
  padding: 0 8px;
}

:deep(.el-tree-node__content:hover) {
  background-color: #f5f7fa;
}

:deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: #ecf5ff;
  color: #409eff;
}

:deep(.el-tree-node.is-current > .el-tree-node__content .node-icon) {
  color: #409eff;
}

:deep(.el-tree-node.is-current > .el-tree-node__content .node-label) {
  color: #409eff;
  font-weight: 500;
}

/* 右键菜单样式 */
.context-menu {
  position: fixed;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 4px 0;
  min-width: 160px;
  z-index: 9999;
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 14px;
  color: #606266;
  cursor: pointer;
  transition: background-color 0.2s;
}

.context-menu-item:hover {
  background-color: #f5f7fa;
}

.context-menu-item.danger {
  color: #f56c6c;
}

.context-menu-item.danger:hover {
  background-color: #fef0f0;
}

.context-menu-divider {
  height: 1px;
  background-color: #e4e7ed;
  margin: 4px 0;
}
</style>
