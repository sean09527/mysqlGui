<template>
  <el-card class="diff-viewer">
    <template #header>
      <div class="card-header">
        <span>结构差异</span>
        <el-tag type="info">
          共 {{ totalDifferences }} 处差异
        </el-tag>
      </div>
    </template>

    <el-tabs v-model="activeTab">
      <!-- 仅存在于源数据库的表 -->
      <el-tab-pane
        label="仅源库存在"
        name="source-only"
        :disabled="diff.tablesOnlyInSource.length === 0"
      >
        <template #label>
          <span>
            仅源库存在
            <el-badge
              v-if="diff.tablesOnlyInSource.length > 0"
              :value="diff.tablesOnlyInSource.length"
              type="success"
            />
          </span>
        </template>

        <el-alert
          type="success"
          :closable="false"
          show-icon
          style="margin-bottom: 15px"
        >
          以下表仅存在于源数据库，同步时将在目标数据库中创建
        </el-alert>

        <el-table :data="sourceOnlyTables" stripe border>
          <el-table-column prop="name" label="表名" width="200" />
          <el-table-column label="操作" width="100">
            <template #default>
              <el-tag type="success" size="small">CREATE</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="说明">
            <template #default="{ row }">
              将在目标数据库创建表 {{ row.name }}
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 仅存在于目标数据库的表 -->
      <el-tab-pane
        label="仅目标库存在"
        name="target-only"
        :disabled="diff.tablesOnlyInTarget.length === 0"
      >
        <template #label>
          <span>
            仅目标库存在
            <el-badge
              v-if="diff.tablesOnlyInTarget.length > 0"
              :value="diff.tablesOnlyInTarget.length"
              type="danger"
            />
          </span>
        </template>

        <el-alert
          type="warning"
          :closable="false"
          show-icon
          style="margin-bottom: 15px"
        >
          以下表仅存在于目标数据库，同步时将被删除（请谨慎操作）
        </el-alert>

        <el-table :data="targetOnlyTables" stripe border>
          <el-table-column prop="name" label="表名" width="200" />
          <el-table-column label="操作" width="100">
            <template #default>
              <el-tag type="danger" size="small">DROP</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="说明">
            <template #default="{ row }">
              将从目标数据库删除表 {{ row.name }}
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 结构不同的表 -->
      <el-tab-pane
        label="结构差异"
        name="differences"
        :disabled="diff.tableDifferences.length === 0"
      >
        <template #label>
          <span>
            结构差异
            <el-badge
              v-if="diff.tableDifferences.length > 0"
              :value="diff.tableDifferences.length"
              type="warning"
            />
          </span>
        </template>

        <el-alert
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 15px"
        >
          以下表在两个数据库中都存在，但结构不同，同步时将修改目标数据库的表结构
        </el-alert>

        <el-collapse v-model="expandedTables" accordion>
          <el-collapse-item
            v-for="tableDiff in diff.tableDifferences"
            :key="tableDiff.tableName"
            :name="tableDiff.tableName"
          >
            <template #title>
              <div class="table-diff-title">
                <el-icon><Document /></el-icon>
                <strong>{{ tableDiff.tableName }}</strong>
                <el-tag type="warning" size="small" style="margin-left: 10px">
                  {{ getTotalChanges(tableDiff) }} 处变更
                </el-tag>
              </div>
            </template>

            <!-- 列差异 -->
            <div v-if="hasColumnChanges(tableDiff)" class="diff-section">
              <h4>列变更</h4>

              <!-- 新增的列 -->
              <div v-if="tableDiff.columnsAdded.length > 0" class="change-group">
                <div class="change-header added">
                  <el-icon><Plus /></el-icon>
                  新增列 ({{ tableDiff.columnsAdded.length }})
                </div>
                <el-table :data="tableDiff.columnsAdded" size="small" border>
                  <el-table-column prop="name" label="列名" width="150" />
                  <el-table-column prop="type" label="类型" width="120" />
                  <el-table-column label="允许NULL" width="100" align="center">
                    <template #default="{ row }">
                      <el-tag :type="row.nullable ? 'success' : 'danger'" size="small">
                        {{ row.nullable ? 'YES' : 'NO' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="defaultValue" label="默认值" width="120" />
                  <el-table-column prop="comment" label="注释" min-width="150" />
                </el-table>
              </div>

              <!-- 删除的列 -->
              <div v-if="tableDiff.columnsRemoved.length > 0" class="change-group">
                <div class="change-header removed">
                  <el-icon><Minus /></el-icon>
                  删除列 ({{ tableDiff.columnsRemoved.length }})
                </div>
                <el-table :data="tableDiff.columnsRemoved" size="small" border>
                  <el-table-column prop="name" label="列名" width="150" />
                  <el-table-column prop="type" label="类型" width="120" />
                  <el-table-column label="允许NULL" width="100" align="center">
                    <template #default="{ row }">
                      <el-tag :type="row.nullable ? 'success' : 'danger'" size="small">
                        {{ row.nullable ? 'YES' : 'NO' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="defaultValue" label="默认值" width="120" />
                  <el-table-column prop="comment" label="注释" min-width="150" />
                </el-table>
              </div>

              <!-- 修改的列 -->
              <div v-if="tableDiff.columnsModified.length > 0" class="change-group">
                <div class="change-header modified">
                  <el-icon><Edit /></el-icon>
                  修改列 ({{ tableDiff.columnsModified.length }})
                </div>
                <div
                  v-for="colDiff in tableDiff.columnsModified"
                  :key="colDiff.columnName"
                  class="column-diff"
                >
                  <div class="column-name">
                    <el-icon><ArrowRight /></el-icon>
                    <strong>{{ colDiff.columnName }}</strong>
                  </div>
                  <el-row :gutter="10">
                    <el-col :span="12">
                      <div class="old-value">
                        <div class="label">原值 (目标库)</div>
                        <div class="value">
                          <div><strong>类型:</strong> {{ colDiff.oldColumn.type }}</div>
                          <div><strong>NULL:</strong> {{ colDiff.oldColumn.nullable ? 'YES' : 'NO' }}</div>
                          <div><strong>默认值:</strong> {{ colDiff.oldColumn.defaultValue || '无' }}</div>
                          <div><strong>注释:</strong> {{ colDiff.oldColumn.comment || '无' }}</div>
                        </div>
                      </div>
                    </el-col>
                    <el-col :span="12">
                      <div class="new-value">
                        <div class="label">新值 (源库)</div>
                        <div class="value">
                          <div><strong>类型:</strong> {{ colDiff.newColumn.type }}</div>
                          <div><strong>NULL:</strong> {{ colDiff.newColumn.nullable ? 'YES' : 'NO' }}</div>
                          <div><strong>默认值:</strong> {{ colDiff.newColumn.defaultValue || '无' }}</div>
                          <div><strong>注释:</strong> {{ colDiff.newColumn.comment || '无' }}</div>
                        </div>
                      </div>
                    </el-col>
                  </el-row>
                </div>
              </div>
            </div>

            <!-- 索引差异 -->
            <div v-if="hasIndexChanges(tableDiff)" class="diff-section">
              <h4>索引变更</h4>

              <!-- 新增的索引 -->
              <div v-if="tableDiff.indexesAdded.length > 0" class="change-group">
                <div class="change-header added">
                  <el-icon><Plus /></el-icon>
                  新增索引 ({{ tableDiff.indexesAdded.length }})
                </div>
                <el-table :data="tableDiff.indexesAdded" size="small" border>
                  <el-table-column prop="name" label="索引名" width="200" />
                  <el-table-column prop="type" label="类型" width="120" />
                  <el-table-column label="列">
                    <template #default="{ row }">
                      {{ row.columns.join(', ') }}
                    </template>
                  </el-table-column>
                </el-table>
              </div>

              <!-- 删除的索引 -->
              <div v-if="tableDiff.indexesRemoved.length > 0" class="change-group">
                <div class="change-header removed">
                  <el-icon><Minus /></el-icon>
                  删除索引 ({{ tableDiff.indexesRemoved.length }})
                </div>
                <el-table :data="tableDiff.indexesRemoved" size="small" border>
                  <el-table-column prop="name" label="索引名" width="200" />
                  <el-table-column prop="type" label="类型" width="120" />
                  <el-table-column label="列">
                    <template #default="{ row }">
                      {{ row.columns.join(', ') }}
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>

            <!-- 外键差异 -->
            <div v-if="hasForeignKeyChanges(tableDiff)" class="diff-section">
              <h4>外键变更</h4>

              <!-- 新增的外键 -->
              <div v-if="tableDiff.foreignKeysAdded.length > 0" class="change-group">
                <div class="change-header added">
                  <el-icon><Plus /></el-icon>
                  新增外键 ({{ tableDiff.foreignKeysAdded.length }})
                </div>
                <el-table :data="tableDiff.foreignKeysAdded" size="small" border>
                  <el-table-column prop="name" label="外键名" width="200" />
                  <el-table-column label="列">
                    <template #default="{ row }">
                      {{ row.columns.join(', ') }}
                    </template>
                  </el-table-column>
                  <el-table-column label="引用">
                    <template #default="{ row }">
                      {{ row.referencedTable }} ({{ row.referencedColumns.join(', ') }})
                    </template>
                  </el-table-column>
                </el-table>
              </div>

              <!-- 删除的外键 -->
              <div v-if="tableDiff.foreignKeysRemoved.length > 0" class="change-group">
                <div class="change-header removed">
                  <el-icon><Minus /></el-icon>
                  删除外键 ({{ tableDiff.foreignKeysRemoved.length }})
                </div>
                <el-table :data="tableDiff.foreignKeysRemoved" size="small" border>
                  <el-table-column prop="name" label="外键名" width="200" />
                  <el-table-column label="列">
                    <template #default="{ row }">
                      {{ row.columns.join(', ') }}
                    </template>
                  </el-table-column>
                  <el-table-column label="引用">
                    <template #default="{ row }">
                      {{ row.referencedTable }} ({{ row.referencedColumns.join(', ') }})
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { Document, Plus, Minus, Edit, ArrowRight } from '@element-plus/icons-vue';
import type { SchemaDiff, TableDiff } from '../types/api';

interface Props {
  diff: SchemaDiff;
}

const props = defineProps<Props>();

const activeTab = ref('source-only');
const expandedTables = ref<string[]>([]);

// 计算属性
const totalDifferences = computed(() => {
  return (
    props.diff.tablesOnlyInSource.length +
    props.diff.tablesOnlyInTarget.length +
    props.diff.tableDifferences.length
  );
});

const sourceOnlyTables = computed(() => {
  return props.diff.tablesOnlyInSource.map(name => ({ name }));
});

const targetOnlyTables = computed(() => {
  return props.diff.tablesOnlyInTarget.map(name => ({ name }));
});

// 辅助函数
function getTotalChanges(tableDiff: TableDiff): number {
  return (
    tableDiff.columnsAdded.length +
    tableDiff.columnsRemoved.length +
    tableDiff.columnsModified.length +
    tableDiff.indexesAdded.length +
    tableDiff.indexesRemoved.length +
    tableDiff.foreignKeysAdded.length +
    tableDiff.foreignKeysRemoved.length
  );
}

function hasColumnChanges(tableDiff: TableDiff): boolean {
  return (
    tableDiff.columnsAdded.length > 0 ||
    tableDiff.columnsRemoved.length > 0 ||
    tableDiff.columnsModified.length > 0
  );
}

function hasIndexChanges(tableDiff: TableDiff): boolean {
  return (
    tableDiff.indexesAdded.length > 0 ||
    tableDiff.indexesRemoved.length > 0
  );
}

function hasForeignKeyChanges(tableDiff: TableDiff): boolean {
  return (
    tableDiff.foreignKeysAdded.length > 0 ||
    tableDiff.foreignKeysRemoved.length > 0
  );
}

// 自动展开第一个有差异的表
if (props.diff.tableDifferences.length > 0) {
  expandedTables.value = [props.diff.tableDifferences[0].tableName];
}

// 自动选择有内容的标签页
if (props.diff.tablesOnlyInSource.length > 0) {
  activeTab.value = 'source-only';
} else if (props.diff.tablesOnlyInTarget.length > 0) {
  activeTab.value = 'target-only';
} else if (props.diff.tableDifferences.length > 0) {
  activeTab.value = 'differences';
}
</script>

<style scoped>
.diff-viewer {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-diff-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.diff-section {
  margin-bottom: 20px;
}

.diff-section h4 {
  margin: 15px 0 10px 0;
  color: #303133;
  font-size: 14px;
  font-weight: 600;
}

.change-group {
  margin-bottom: 15px;
}

.change-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 4px;
  font-weight: 600;
  margin-bottom: 10px;
}

.change-header.added {
  background-color: #f0f9ff;
  color: #67c23a;
  border-left: 3px solid #67c23a;
}

.change-header.removed {
  background-color: #fef0f0;
  color: #f56c6c;
  border-left: 3px solid #f56c6c;
}

.change-header.modified {
  background-color: #fdf6ec;
  color: #e6a23c;
  border-left: 3px solid #e6a23c;
}

.column-diff {
  margin-bottom: 15px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.column-name {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 14px;
}

.old-value,
.new-value {
  padding: 10px;
  border-radius: 4px;
}

.old-value {
  background-color: #fef0f0;
  border: 1px solid #fbc4c4;
}

.new-value {
  background-color: #f0f9ff;
  border: 1px solid #b3d8ff;
}

.old-value .label,
.new-value .label {
  font-weight: 600;
  margin-bottom: 8px;
  font-size: 12px;
  text-transform: uppercase;
}

.old-value .label {
  color: #f56c6c;
}

.new-value .label {
  color: #409eff;
}

.old-value .value,
.new-value .value {
  font-size: 13px;
  line-height: 1.8;
}

.old-value .value div,
.new-value .value div {
  margin-bottom: 4px;
}
</style>
