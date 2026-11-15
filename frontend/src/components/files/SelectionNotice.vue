<template>
  <div class="active" id="multiple-selection">
    <i class="material-icons selection-mode-icon" aria-hidden="true"
      >checklist</i
    >
    <p role="status">{{ t("files.multipleSelectionEnabled") }}</p>
    <button
      type="button"
      @click="fileStore.multiple = false"
      :title="t('buttons.clear')"
      :aria-label="t('buttons.clear')"
      class="selection-dismiss"
    >
      <i class="material-icons" aria-hidden="true">close</i>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useFileStore } from "@/stores/file";

const { t } = useI18n();
const fileStore = useFileStore();
</script>
