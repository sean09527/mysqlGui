<script setup lang="ts">
import { ref, onMounted } from 'vue';
import MainLayout from './components/MainLayout.vue';
import WailsNotAvailable from './components/WailsNotAvailable.vue';

const isWailsAvailable = ref(false);
const isChecking = ref(true);

onMounted(() => {
  // Check if Wails runtime is available
  const checkWails = () => {
    if (typeof window !== 'undefined' && 
        (window as any)['go'] !== undefined && 
        (window as any)['go']['backend'] !== undefined &&
        (window as any)['go']['backend']['App'] !== undefined) {
      isWailsAvailable.value = true;
    } else {
      isWailsAvailable.value = false;
    }
    isChecking.value = false;
  };

  // Give Wails a moment to initialize
  setTimeout(checkWails, 100);
});
</script>

<template>
  <div id="app">
    <div v-if="isChecking" class="loading">
      <el-icon class="is-loading" :size="40">
        <Loading />
      </el-icon>
      <p>正在初始化...</p>
    </div>
    <WailsNotAvailable v-else-if="!isWailsAvailable" />
    <MainLayout v-else />
  </div>
</template>

<style>
#app {
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  margin: 0;
  padding: 0;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.loading p {
  margin-top: 20px;
  font-size: 18px;
}
</style>
