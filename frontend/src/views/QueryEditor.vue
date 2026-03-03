<template>
  <div class="query-editor-view">
    <div v-if="!currentConnection" class="no-connection">
      <el-empty description="请先连接到数据库">
        <el-button type="primary" @click="goToConnections">
          前往连接管理
        </el-button>
      </el-empty>
    </div>
    <div v-else class="editor-container">
      <QueryEditor :profile-id="currentConnection.id" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useConnectionStore } from '../stores/connection'
import QueryEditor from '../components/QueryEditor.vue'

const router = useRouter()
const connectionStore = useConnectionStore()

const currentConnection = computed(() => connectionStore.currentConnection)

const goToConnections = () => {
  router.push('/connections')
}
</script>

<style scoped>
.query-editor-view {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.no-connection {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.editor-container {
  flex: 1;
  overflow: hidden;
}
</style>
