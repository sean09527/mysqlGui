<template>
  <div class="workspace">
    <!-- 左侧数据库浏览器 -->
    <div class="workspace-sidebar">
      <DatabaseExplorer
        @view-schema="handleViewSchema"
        @view-data="handleViewData"
        @edit-schema="handleEditSchema"
      />
    </div>

    <!-- 右侧内容区域 -->
    <div class="workspace-content">
      <!-- 数据表列表（当选中数据库但未选中表时显示） -->
      <div v-if="contentType === 'table-list'" class="table-list-panel">
        <TableList
          @view-schema="handleViewSchema"
          @view-data="handleViewData"
        />
      </div>

      <!-- 欢迎页面 -->
      <div v-else-if="contentType === 'welcome'" class="welcome-content">
        <el-empty description="请从左侧选择数据库和表">
          <template #image>
            <el-icon :size="100" color="#909399">
              <FolderOpened />
            </el-icon>
          </template>
          <div class="welcome-tips">
            <p><strong>提示：</strong></p>
            <ul>
              <li>单击数据库名自动展开表列表</li>
              <li>双击表名查看表数据</li>
              <li>右键点击表名选择"修改表结构"</li>
              <li>右键点击表名选择"查看表结构"</li>
            </ul>
          </div>
        </el-empty>
      </div>

      <!-- 表结构查看 -->
      <div v-else-if="contentType === 'schema-view'" class="content-panel">
        <div class="content-header">
          <h3>
            <el-icon><Document /></el-icon>
            表结构 - {{ currentDatabase }}.{{ currentTable }}
          </h3>
          <div class="content-actions">
            <el-button size="small" @click="handleEditSchema">
              <el-icon><Edit /></el-icon>
              修改表结构
            </el-button>
            <el-button size="small" @click="handleViewData">
              <el-icon><Grid /></el-icon>
              查看数据
            </el-button>
            <el-button size="small" @click="handleClose">
              <el-icon><Close /></el-icon>
              关闭
            </el-button>
          </div>
        </div>
        <div class="content-body">
          <SchemaViewer
            v-if="currentConnection && currentDatabase && currentTable"
            :profile-id="currentConnection.id"
            :database="currentDatabase"
            :table="currentTable"
          />
        </div>
      </div>

      <!-- 表结构编辑 -->
      <div v-else-if="contentType === 'schema-edit'" class="content-panel">
        <div class="content-header">
          <h3>
            <el-icon><Edit /></el-icon>
            {{ currentTable ? `修改表结构 - ${currentDatabase}.${currentTable}` : `新建表 - ${currentDatabase}` }}
          </h3>
          <div class="content-actions">
            <el-button v-if="currentTable" size="small" @click="handleViewSchema">
              <el-icon><Document /></el-icon>
              查看结构
            </el-button>
            <el-button v-if="currentTable" size="small" @click="handleViewData">
              <el-icon><Grid /></el-icon>
              查看数据
            </el-button>
            <el-button size="small" @click="handleClose">
              <el-icon><Close /></el-icon>
              关闭
            </el-button>
          </div>
        </div>
        <div class="content-body">
          <TableEditor
            v-if="currentConnection && currentDatabase"
            :profile-id="currentConnection.id"
            :database="currentDatabase"
            :table="currentTable || ''"
            :mode="currentTable ? 'edit' : 'create'"
            @success="handleSchemaUpdateSuccess"
          />
        </div>
      </div>

      <!-- 表数据管理 -->
      <div v-else-if="contentType === 'data'" class="content-panel">
        <div class="content-header">
          <h3>
            <el-icon><Grid /></el-icon>
            表数据 - {{ currentDatabase }}.{{ currentTable }}
          </h3>
          <div class="content-actions">
            <el-button size="small" @click="handleViewSchema">
              <el-icon><Document /></el-icon>
              查看结构
            </el-button>
            <el-button size="small" @click="handleEditSchema">
              <el-icon><Edit /></el-icon>
              修改结构
            </el-button>
            <el-button size="small" @click="handleClose">
              <el-icon><Close /></el-icon>
              关闭
            </el-button>
          </div>
        </div>
        <div class="content-body">
          <DataManager
            v-if="currentConnection && currentDatabase && currentTable"
            :profile-id="currentConnection.id"
            :database="currentDatabase"
            :table="currentTable"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useConnectionStore } from '../stores/connection';
import { useDatabaseStore } from '../stores/database';
import DatabaseExplorer from '../components/DatabaseExplorer.vue';
import SchemaViewer from '../components/SchemaViewer.vue';
import TableEditor from '../components/TableEditor.vue';
import DataManager from '../components/DataManager.vue';
import TableList from '../components/TableList.vue';
import {
  FolderOpened,
  Document,
  Edit,
  Grid,
  Close,
} from '@element-plus/icons-vue';

const connectionStore = useConnectionStore();
const databaseStore = useDatabaseStore();

const currentConnection = computed(() => connectionStore.currentConnection);
const currentDatabase = computed(() => databaseStore.currentDatabase);
const currentTable = computed(() => databaseStore.currentTable);

// 内容类型: 'welcome' | 'table-list' | 'schema-view' | 'schema-edit' | 'data'
const contentType = ref<string>('welcome');

// 处理查看表结构
const handleViewSchema = () => {
  if (currentTable.value) {
    contentType.value = 'schema-view';
  }
};

// 处理编辑表结构
const handleEditSchema = () => {
  // 支持新建表（currentTable 为 null）和编辑表（currentTable 有值）
  if (currentDatabase.value) {
    contentType.value = 'schema-edit';
  }
};

// 处理查看表数据
const handleViewData = () => {
  if (currentTable.value) {
    contentType.value = 'data';
  }
};

// 处理关闭
const handleClose = () => {
  // 关闭后显示表列表（如果有选中的数据库）
  if (currentDatabase.value) {
    contentType.value = 'table-list';
  } else {
    contentType.value = 'welcome';
  }
  databaseStore.setCurrentTable(null);
};

// 处理表结构更新成功
const handleSchemaUpdateSuccess = () => {
  // 可以选择切换到查看模式或保持编辑模式
  contentType.value = 'schema-view';
};

// 监听当前数据库变化
watch(currentDatabase, (newDb) => {
  if (newDb && !currentTable.value) {
    // 选中数据库但没有选中表时，显示表列表
    contentType.value = 'table-list';
  } else if (!newDb) {
    contentType.value = 'welcome';
  }
});

// 监听当前表变化
watch(currentTable, (newTable) => {
  if (!newTable && currentDatabase.value) {
    // 取消选中表时，显示表列表
    contentType.value = 'table-list';
  } else if (!newTable && !currentDatabase.value) {
    contentType.value = 'welcome';
  }
});
</script>

<style scoped>
.workspace {
  display: flex;
  height: 100%;
  gap: 0;
  background-color: #f5f7fa;
}

.workspace-sidebar {
  width: 280px;
  min-width: 280px;
  background-color: #fff;
  border-right: 1px solid #e4e7ed;
  overflow: hidden;
}

.workspace-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.welcome-content {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background-color: #fff;
  margin: 16px;
  border-radius: 4px;
}

.table-list-panel {
  height: 100%;
  background-color: #fff;
  margin: 16px;
  border-radius: 4px;
  overflow: hidden;
}

.welcome-tips {
  margin-top: 20px;
  text-align: left;
}

.welcome-tips p {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: #606266;
}

.welcome-tips ul {
  margin: 0;
  padding-left: 20px;
  list-style: disc;
}

.welcome-tips li {
  margin: 8px 0;
  font-size: 13px;
  color: #909399;
}

.content-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #fff;
  margin: 16px;
  border-radius: 4px;
  overflow: hidden;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e4e7ed;
  background-color: #fafafa;
}

.content-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 8px;
}

.content-actions {
  display: flex;
  gap: 8px;
}

.content-body {
  flex: 1;
  overflow: auto;
  padding: 0;
}

/* 移除 DataManager 内部的 padding，因为外层已经有了 */
.content-body :deep(.data-manager) {
  padding: 0;
  height: 100%;
}
</style>
