<template>
  <div class="schema-sync-view">
    <SchemaSync />
  </div>
</template>

<script setup lang="ts">
import SchemaSync from '../components/SchemaSync.vue';
</script>

<style scoped>
.schema-sync-view {
  height: 100%;
  overflow: auto;
}
</style>
