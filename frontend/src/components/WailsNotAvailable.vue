<template>
  <div class="wails-not-available">
    <el-result
      icon="warning"
      title="Wails 运行时不可用"
      sub-title="应用程序需要通过 Wails 启动才能正常工作"
    >
      <template #extra>
        <div class="instructions">
          <h3>如何启动应用程序：</h3>
          <ol>
            <li>
              <strong>开发模式：</strong>
              <el-code>wails dev</el-code>
            </li>
            <li>
              <strong>生产模式：</strong>
              <el-code>wails build</el-code>
              然后运行生成的可执行文件
            </li>
          </ol>
          
          <el-alert
            title="提示"
            type="info"
            :closable="false"
            show-icon
          >
            <p>如果您看到此消息，说明您可能是通过以下方式启动的：</p>
            <ul>
              <li><code>npm run dev</code> - 这只会启动前端开发服务器</li>
              <li><code>vite</code> - 这只会启动 Vite 开发服务器</li>
            </ul>
            <p>请使用 <code>wails dev</code> 命令来启动完整的应用程序（前端 + 后端）。</p>
          </el-alert>
        </div>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
// This component is shown when Wails runtime is not available
</script>

<style scoped>
.wails-not-available {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

:deep(.el-result) {
  background: white;
  border-radius: 8px;
  padding: 40px;
  max-width: 800px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.instructions {
  text-align: left;
  margin-top: 20px;
}

.instructions h3 {
  margin-bottom: 15px;
  color: #303133;
}

.instructions ol {
  margin: 15px 0;
  padding-left: 20px;
}

.instructions li {
  margin: 10px 0;
  line-height: 1.8;
}

.el-code {
  background: #f5f7fa;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  color: #e83e8c;
  font-weight: 600;
}

:deep(.el-alert) {
  margin-top: 20px;
  text-align: left;
}

:deep(.el-alert ul) {
  margin: 10px 0;
  padding-left: 20px;
}

:deep(.el-alert li) {
  margin: 5px 0;
}

:deep(.el-alert code) {
  background: rgba(64, 158, 255, 0.1);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  color: #409eff;
}
</style>
