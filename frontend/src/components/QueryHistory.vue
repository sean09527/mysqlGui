<template>
  <div class="query-history">
    <div class="history-toolbar">
      <el-button 
        :icon="Refresh" 
        @click="loadHistory"
        :loading="loading"
        size="small"
      >
        刷新
      </el-button>
      <el-button 
        :icon="Delete" 
        @click="clearHistory"
        size="small"
        type="danger"
      >
        清空历史
      </el-button>
      <el-input
        v-model="searchText"
        placeholder="搜索 SQL..."
        :prefix-icon="Search"
        size="small"
        style="width: 300px; margin-left: 10px;"
        clearable
      />
    </div>

    <div class="history-list">
      <el-empty v-if="!loading && filteredHistory.length === 0" description="暂无查询历史" />
      
      <div 
        v-for="entry in filteredHistory" 
        :key="entry.id"
        class="history-item"
        :class="{ 'history-item-error': !entry.success }"
        @click="selectQuery(entry)"
      >
        <div class="history-header">
          <el-tag 
            :type="entry.success ? 'success' : 'danger'" 
            size="small"
          >
            {{ entry.success ? '成功' : '失败' }}
          </el-tag>
          <span class="history-time">{{ formatTime(entry.timestamp) }}</span>
        </div>
        
        <div class="history-sql">
          {{ truncateSQL(entry.sql) }}
        </div>
        
        <div class="history-footer">
          <span v-if="entry.database" class="history-database">
            <el-icon><Document /></el-icon>
            {{ entry.database }}
          </span>
          <span class="history-stats">
            <el-icon><Timer /></el-icon>
            {{ entry.executionTime }} ms
          </span>
          <span v-if="entry.rowsAffected > 0" class="history-stats">
            <el-icon><Document /></el-icon>
            {{ entry.rowsAffected }} 行
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Search, Timer, Document } from '@element-plus/icons-vue'
import { GetQueryHistory, ClearQueryHistory } from '../../wailsjs/go/backend/App'
import { storage } from '../../wailsjs/go/models'

interface Props {
  profileId: string
}

type QueryHistoryEntry = storage.QueryHistoryEntry

const props = defineProps<Props>()
const emit = defineEmits<{
  selectQuery: [sql: string]
}>()

const loading = ref(false)
const history = ref<QueryHistoryEntry[]>([])
const searchText = ref('')

// 过滤后的历史记录
const filteredHistory = computed(() => {
  if (!searchText.value) {
    return history.value
  }
  
  const search = searchText.value.toLowerCase()
  return history.value.filter(entry => 
    entry.sql.toLowerCase().includes(search) ||
    entry.database?.toLowerCase().includes(search)
  )
})

// 加载查询历史
const loadHistory = async () => {
  if (!props.profileId) {
    return
  }

  loading.value = true
  try {
    const result = await GetQueryHistory(props.profileId, 100)
    history.value = result || []
  } catch (error: any) {
    ElMessage.error('加载查询历史失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

// 清空历史
const clearHistory = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有查询历史吗？此操作不可恢复。',
      '确认清空',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await ClearQueryHistory(props.profileId)
    history.value = []
    ElMessage.success('查询历史已清空')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('清空历史失败: ' + error.message)
    }
  }
}

// 选择查询
const selectQuery = (entry: QueryHistoryEntry) => {
  emit('selectQuery', entry.sql)
}

// 格式化时间
const formatTime = (timestamp: string): string => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  // 小于1分钟
  if (diff < 60000) {
    return '刚刚'
  }
  
  // 小于1小时
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)} 分钟前`
  }
  
  // 小于1天
  if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)} 小时前`
  }
  
  // 小于7天
  if (diff < 604800000) {
    return `${Math.floor(diff / 86400000)} 天前`
  }
  
  // 显示完整日期
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 截断 SQL
const truncateSQL = (sql: string): string => {
  const maxLength = 200
  if (sql.length <= maxLength) {
    return sql
  }
  return sql.substring(0, maxLength) + '...'
}

// 组件挂载时加载历史
onMounted(() => {
  loadHistory()
})

// 暴露刷新方法给父组件
defineExpose({
  loadHistory
})
</script>

<style scoped>
.query-history {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.history-toolbar {
  display: flex;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #ebeef5;
}

.history-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.history-item {
  padding: 12px;
  margin-bottom: 10px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
  background-color: #fff;
}

.history-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.history-item-error {
  border-left: 3px solid #f56c6c;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.history-time {
  font-size: 12px;
  color: #909399;
}

.history-sql {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #303133;
  margin-bottom: 8px;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.5;
}

.history-footer {
  display: flex;
  gap: 15px;
  font-size: 12px;
  color: #909399;
}

.history-database,
.history-stats {
  display: flex;
  align-items: center;
  gap: 4px;
}

.history-database .el-icon,
.history-stats .el-icon {
  font-size: 14px;
}
</style>
