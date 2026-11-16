<template>
  <h2 class="settings-card-title">
    <i class="material-icons settings-card-icon" aria-hidden="true">{{
      icon
    }}</i>
    <span><slot /></span>
  </h2>
</template>

<script setup lang="ts">
defineProps<{ icon: string }>();
</script>
