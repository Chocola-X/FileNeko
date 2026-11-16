<template>
  <div class="t-container">
    <span>{{ message }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  message: string;
}>();
</script>

<style scoped>
@layer components {
  .t-container {
    width: 100%;
    padding: 5px 0;
    display: flex;
    min-width: 0;
    overflow-wrap: anywhere;
    align-items: center;
  }
}
</style>
