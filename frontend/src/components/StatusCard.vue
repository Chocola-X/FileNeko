<template>
  <section class="status-card" role="status">
    <span class="status-card-emblem" aria-hidden="true">
      <i class="material-icons">{{ icon }}</i>
      <i class="material-icons status-card-sparkle">{{ badge }}</i>
    </span>
    <div class="status-card-caption">
      <span class="status-card-label" aria-hidden="true">{{ label }}</span>
      <h2><slot /></h2>
    </div>
    <div v-if="$slots.actions" class="status-card-actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ icon: string; label: string; badge?: string }>(), {
  badge: "auto_awesome",
});
</script>
