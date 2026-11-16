<template>
  <section class="folder-empty" role="status">
    <span class="folder-empty-emblem" aria-hidden="true">
      <i class="material-icons">folder_open</i>
      <i class="material-icons folder-empty-sparkle">auto_awesome</i>
    </span>
    <div class="folder-empty-caption">
      <span class="folder-empty-label" aria-hidden="true">EMPTY FOLDER</span>
      <h2>{{ t("files.lonely") }}</h2>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

const { t } = useI18n();
</script>
