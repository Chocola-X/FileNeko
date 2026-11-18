<template>
  <div class="loading-state delayed" role="status" aria-live="polite">
    <div class="neko-loader fn-loader">
      <div
        class="neko-loader__art"
        aria-hidden="true"
        v-html="loadingBook"
      ></div>
      <div class="neko-loader__trail" aria-hidden="true">
        <i></i><i></i><i></i>
      </div>
      <span class="neko-loader__label">{{ t("files.loading") }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import loadingBook from "@/assets/loading-book.svg?raw";

const { t } = useI18n();
</script>
