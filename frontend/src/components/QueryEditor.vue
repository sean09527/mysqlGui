<template>
  <div class="query-editor">
    <div class="editor-toolbar">
      <el-select
        v-model="selectedDatabase"
        placeholder="选择数据库"
        filterable
        clearable
        style="width: 200px; margin-right: 10px;"
        @change="handleDatabaseChange"
      >
        <el-option
          v-for="db in databases"
          :key="db"
          :label="db"
          :value="db"
        />
      </el-select>
      <el-button 
        type="primary" 
        :icon="VideoPlay" 
        @click="executeQuery"
        :loading="executing"
        :disabled="!sqlText.trim()"
      >
        执行 (Ctrl+Enter)
      </el-button>
      <el-button 
        :icon="Close" 
        @click="cancelQuery"
        :disabled="!executing"
      >
        取消
      </el-button>
      <el-button 
        :icon="FolderAdd" 
        @click="showSaveDialog"
        :disabled="!sqlText.trim()"
      >
        保存查询
      </el-button>
      <el-button 
        :icon="Delete" 
        @click="clearEditor"
      >
        清空
      </el-button>
      <el-divider direction="vertical" />
      <span v-if="lastExecutionTime" class="execution-info">
        执行时间: {{ formatDuration(lastExecutionTime) }}
      </span>
    </div>

    <div class="editor-container">
      <div ref="editorRef" class="codemirror-wrapper"></div>
    </div>

    <div class="result-container">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="查询结果" name="result">
          <div v-if="queryResult">
            <!-- SELECT 查询结果 -->
            <div v-if="queryResult.type === 'SELECT' && queryResult.rows">
              <div class="result-info">
                <el-tag type="success">
                  返回 {{ queryResult.rows.length }} 行
                </el-tag>
                <el-tag type="info" style="margin-left: 10px;">
                  执行时间: {{ formatDuration(queryResult.executionTime) }}
                </el-tag>
              </div>
              <el-table 
                :data="queryResult.rows" 
                border 
                stripe 
                style="width: 100%; margin-top: 10px;"
                max-height="400"
              >
                <el-table-column
                  v-for="(col, index) in queryResult.columns"
                  :key="index"
                  :prop="index.toString()"
                  :label="col"
                  min-width="120"
                >
                  <template #default="scope">
                    {{ formatCellValue(scope.row[index]) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <!-- DML 语句结果 -->
            <div v-else-if="['INSERT', 'UPDATE', 'DELETE'].includes(queryResult.type)">
              <el-result
                icon="success"
                :title="`${queryResult.type} 执行成功`"
                :sub-title="`影响 ${queryResult.rowsAffected} 行，执行时间: ${formatDuration(queryResult.executionTime)}`"
              />
            </div>

            <!-- DDL 语句结果 -->
            <div v-else-if="queryResult.type === 'DDL'">
              <el-result
                icon="success"
                title="DDL 执行成功"
                :sub-title="`执行时间: ${formatDuration(queryResult.executionTime)}`"
              />
            </div>

            <!-- 其他语句结果 -->
            <div v-else>
              <el-result
                icon="success"
                title="执行成功"
                :sub-title="`执行时间: ${formatDuration(queryResult.executionTime)}`"
              />
            </div>
          </div>

          <el-empty v-else description="执行查询后将在此显示结果" />
        </el-tab-pane>

        <el-tab-pane label="错误信息" name="error">
          <div v-if="queryError" class="error-container">
            <el-alert
              type="error"
              :title="`错误代码: ${queryError.code || '未知'}`"
              :description="queryError.message"
              show-icon
              :closable="false"
            />
            <div v-if="queryError.position >= 0" class="error-position">
              错误位置: 第 {{ queryError.position }} 个字符
            </div>
          </div>
          <el-empty v-else description="无错误信息" />
        </el-tab-pane>

        <el-tab-pane label="查询历史" name="history">
          <query-history 
            :profile-id="props.profileId"
            @select-query="handleSelectHistory"
          />
        </el-tab-pane>

        <el-tab-pane name="saved">
          <template #label>
            <span><el-icon><Folder /></el-icon> 已保存</span>
          </template>
          <div class="saved-queries">
            <el-table :data="savedQueries" border stripe>
              <el-table-column prop="name" label="名称" min-width="150" />
              <el-table-column prop="description" label="描述" min-width="200" />
              <el-table-column prop="database" label="数据库" width="120" />
              <el-table-column prop="createdAt" label="创建时间" width="180">
                <template #default="scope">
                  {{ formatDate(scope.row.createdAt) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="scope">
                  <el-button size="small" @click="loadSavedQuery(scope.row)">加载</el-button>
                  <el-button size="small" type="danger" @click="deleteSavedQuery(scope.row.id)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 保存查询对话框 -->
    <el-dialog v-model="saveDialogVisible" title="保存查询" width="500px">
      <el-form :model="saveForm" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="saveForm.name" placeholder="输入查询名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="saveForm.description" type="textarea" :rows="3" placeholder="输入查询描述（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="saveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveQuery" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoPlay, Close, Delete, FolderAdd, Folder } from '@element-plus/icons-vue'
import { EditorView, basicSetup } from 'codemirror'
import { EditorState } from '@codemirror/state'
import { sql as sqlLang, SQLDialect } from '@codemirror/lang-sql'
import { oneDark } from '@codemirror/theme-one-dark'
import { keymap } from '@codemirror/view'
import { ExecuteQuery, CancelQuery, ListDatabases, ListTables, SaveQuery, GetSavedQueries, DeleteSavedQuery } from '../../wailsjs/go/backend/App'
import { query, repository, storage } from '../../wailsjs/go/models'
import QueryHistory from './QueryHistory.vue'

interface Props {
  profileId: string
}

const props = defineProps<Props>()

type QueryResult = query.QueryResult
type QueryError = query.QueryError
type SavedQuery = storage.SavedQuery

const editorRef = ref<HTMLElement>()
let editorView: EditorView | null = null

const sqlText = ref('')
const executing = ref(false)
const queryResult = ref<QueryResult | null>(null)
const queryError = ref<QueryError | null>(null)
const lastExecutionTime = ref<number>(0)
const activeTab = ref('result')
const currentQueryId = ref<string>('')

// 数据库和表相关
const databases = ref<string[]>([])
const selectedDatabase = ref<string>('')
const tables = ref<string[]>([])
const schema = ref<any>(null)

// 保存查询相关
const saveDialogVisible = ref(false)
const saving = ref(false)
const savedQueries = ref<SavedQuery[]>([])
const saveForm = ref({
  name: '',
  description: ''
})

// 初始化 CodeMirror 编辑器
onMounted(async () => {
  // 加载数据库列表
  await loadDatabases()
  // 加载已保存的查询
  await loadSavedQueries()
  
  if (!editorRef.value) return

  try {
    const startState = EditorState.create({
      doc: '',
      extensions: [
        basicSetup,
        sqlLang({
          dialect: SQLDialect.define({}),
          schema: schema.value || {}
        }),
        oneDark,
        keymap.of([
          {
            key: 'Ctrl-Enter',
            run: () => {
              executeQuery()
              return true
            }
          },
          {
            key: 'Cmd-Enter',
            run: () => {
              executeQuery()
              return true
            }
          }
        ]),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            sqlText.value = update.state.doc.toString()
          }
        })
      ]
    })

    editorView = new EditorView({
      state: startState,
      parent: editorRef.value
    })
  } catch (error) {
    console.error('初始化编辑器失败:', error)
    ElMessage.error('初始化 SQL 编辑器失败')
  }
})

onBeforeUnmount(() => {
  if (editorView) {
    editorView.destroy()
  }
})

// 加载数据库列表
const loadDatabases = async () => {
  if (!props.profileId) return
  
  try {
    const result = await ListDatabases(props.profileId)
    // 提取数据库名称
    databases.value = result?.map((db: repository.Database) => db.name) || []
  } catch (error: any) {
    console.error('加载数据库列表失败:', error)
  }
}

// 加载表列表
const loadTables = async (database: string) => {
  if (!props.profileId || !database) return
  
  try {
    const result = await ListTables(props.profileId, database)
    // 提取表名称
    tables.value = result?.map((t: repository.Table) => t.name) || []
    
    // 构建 schema 对象用于自动补全
    schema.value = {}
    tables.value.forEach(tableName => {
      schema.value[tableName] = []
    })
    
    // 重新配置编辑器以使用新的 schema
    updateEditorSchema()
  } catch (error: any) {
    console.error('加载表列表失败:', error)
    ElMessage.error('加载表列表失败: ' + error.message)
  }
}

// 更新编辑器的 schema 配置
const updateEditorSchema = () => {
  if (!editorView) return
  
  try {
    // 保存当前文档内容
    const currentDoc = editorView.state.doc.toString()
    
    // 创建新的编辑器状态
    const newState = EditorState.create({
      doc: currentDoc,
      extensions: [
        basicSetup,
        sqlLang({
          dialect: SQLDialect.define({}),
          schema: schema.value || {}
        }),
        oneDark,
        keymap.of([
          {
            key: 'Ctrl-Enter',
            run: () => {
              executeQuery()
              return true
            }
          },
          {
            key: 'Cmd-Enter',
            run: () => {
              executeQuery()
              return true
            }
          }
        ]),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            sqlText.value = update.state.doc.toString()
          }
        })
      ]
    })
    
    editorView.setState(newState)
  } catch (error) {
    console.error('更新编辑器 schema 失败:', error)
  }
}

// 处理数据库选择变化
const handleDatabaseChange = async (database: string) => {
  if (database) {
    await loadTables(database)
    ElMessage.success(`已切换到数据库: ${database}`)
  } else {
    tables.value = []
    schema.value = null
    updateEditorSchema()
  }
}

// 执行查询
const executeQuery = async () => {
  const sql = sqlText.value.trim()
  if (!sql) {
    ElMessage.warning('请输入 SQL 语句')
    return
  }

  if (!props.profileId) {
    ElMessage.error('未选择数据库连接')
    return
  }

  executing.value = true
  queryResult.value = null
  queryError.value = null
  activeTab.value = 'result'

  try {
    // 如果选择了数据库且 SQL 中没有 USE 语句，先执行 USE
    if (selectedDatabase.value && !sql.toLowerCase().includes('use ')) {
      try {
        await ExecuteQuery(props.profileId, `USE \`${selectedDatabase.value}\``)
      } catch (error: any) {
        console.warn('USE 语句执行失败，继续执行查询:', error)
        // USE 失败不影响后续查询，可能数据库已经选择或者 SQL 中使用了完全限定名
      }
    }
    
    // 执行实际的查询
    const result = await ExecuteQuery(props.profileId, sql)
    
    // 检查结果是否有效
    if (!result) {
      throw new Error('后端返回空结果')
    }
    
    if (result.error) {
      queryError.value = result.error
      activeTab.value = 'error'
      ElMessage.error('查询执行失败')
    } else {
      queryResult.value = result
      lastExecutionTime.value = result.executionTime || 0
      currentQueryId.value = result.id || ''
      ElMessage.success('查询执行成功')
    }
  } catch (error: any) {
    console.error('查询执行错误:', error)
    const errorMessage = error?.message || error?.toString() || '未知错误'
    queryError.value = {
      code: error?.code || 0,
      message: errorMessage,
      position: -1
    }
    activeTab.value = 'error'
    ElMessage.error('查询执行失败: ' + errorMessage)
  } finally {
    executing.value = false
  }
}

// 取消查询
const cancelQuery = async () => {
  if (!currentQueryId.value) return

  try {
    await CancelQuery(props.profileId, currentQueryId.value)
    ElMessage.success('查询已取消')
    executing.value = false
  } catch (error: any) {
    ElMessage.error('取消查询失败: ' + error.message)
  }
}

// 清空编辑器
const clearEditor = () => {
  if (editorView) {
    const currentLength = editorView.state.doc.length
    editorView.dispatch({
      changes: {
        from: 0,
        to: currentLength,
        insert: ''
      }
    })
    sqlText.value = ''
  }
}

// 从历史记录选择查询
const handleSelectHistory = (historySql: string) => {
  if (editorView) {
    const currentLength = editorView.state.doc.length
    editorView.dispatch({
      changes: {
        from: 0,
        to: currentLength,
        insert: historySql
      }
    })
    sqlText.value = historySql
    activeTab.value = 'result'
  }
}

// 格式化时间
const formatDuration = (nanoseconds: number): string => {
  const ms = nanoseconds / 1000000
  if (ms < 1000) {
    return `${ms.toFixed(2)} ms`
  }
  return `${(ms / 1000).toFixed(2)} s`
}

// 格式化单元格值
const formatCellValue = (value: any): string => {
  if (value === null || value === undefined) {
    return 'NULL'
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

// 显示保存对话框
const showSaveDialog = () => {
  saveForm.value = {
    name: '',
    description: ''
  }
  saveDialogVisible.value = true
}

// 保存查询
const handleSaveQuery = async () => {
  if (!saveForm.value.name.trim()) {
    ElMessage.warning('请输入查询名称')
    return
  }

  saving.value = true
  try {
    await SaveQuery(
      props.profileId,
      saveForm.value.name,
      sqlText.value,
      saveForm.value.description,
      selectedDatabase.value
    )
    ElMessage.success('查询已保存')
    saveDialogVisible.value = false
    await loadSavedQueries()
  } catch (error: any) {
    ElMessage.error('保存查询失败: ' + error.message)
  } finally {
    saving.value = false
  }
}

// 加载已保存的查询
const loadSavedQueries = async () => {
  if (!props.profileId) return
  
  try {
    const result = await GetSavedQueries(props.profileId)
    savedQueries.value = result || []
  } catch (error: any) {
    console.error('加载已保存查询失败:', error)
  }
}

// 加载已保存的查询到编辑器
const loadSavedQuery = (query: SavedQuery) => {
  if (editorView) {
    const currentLength = editorView.state.doc.length
    editorView.dispatch({
      changes: {
        from: 0,
        to: currentLength,
        insert: query.sql
      }
    })
    sqlText.value = query.sql
    if (query.database) {
      selectedDatabase.value = query.database
      handleDatabaseChange(query.database)
    }
    activeTab.value = 'result'
    ElMessage.success(`已加载查询: ${query.name}`)
  }
}

// 删除已保存的查询
const deleteSavedQuery = async (id: number) => {
  try {
    await DeleteSavedQuery(id)
    ElMessage.success('查询已删除')
    await loadSavedQueries()
  } catch (error: any) {
    ElMessage.error('删除查询失败: ' + error.message)
  }
}

// 格式化日期
const formatDate = (dateStr: string): string => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}
</script>

<style scoped>
.query-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 20px;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.execution-info {
  color: #606266;
  font-size: 14px;
}

.editor-container {
  flex: 0 0 300px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 20px;
}

.codemirror-wrapper {
  height: 300px;
  overflow: auto;
}

.result-container {
  flex: 1;
  overflow: auto;
}

.result-info {
  margin-bottom: 10px;
}

.error-container {
  padding: 20px;
}

.error-position {
  margin-top: 10px;
  padding: 10px;
  background-color: #fef0f0;
  border-left: 3px solid #f56c6c;
  color: #f56c6c;
}

:deep(.el-tabs__content) {
  padding: 20px;
}

.saved-queries {
  padding: 10px;
}

:deep(.cm-editor) {
  height: 100%;
}

:deep(.cm-scroller) {
  overflow: auto;
}
</style>
