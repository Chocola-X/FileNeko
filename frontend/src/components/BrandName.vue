<template>
  <span
    class="brand-wordmark fn-wordmark"
    :class="{ 'brand-wordmark--default': name === 'FileNeko' }"
  >
    <template v-if="name === 'FileNeko'"
      >File<span class="brand-wordmark-tail">Neko</span></template
    >
    <template v-else>{{ name }}</template>
  </span>
</template>

<script setup lang="ts">
import { name } from "@/utils/constants";
</script>
