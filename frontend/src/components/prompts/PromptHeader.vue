<template>
  <div class="card-title prompt-header">
    <i class="material-icons prompt-icon" aria-hidden="true">{{ icon }}</i>
    <h2><slot /></h2>
    <button
      type="button"
      class="prompt-close"
      :aria-label="t('buttons.close')"
      :title="t('buttons.close')"
      @click="layoutStore.closeHovers()"
    >
      <i class="material-icons" aria-hidden="true">close</i>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useLayoutStore } from "@/stores/layout";

const layoutStore = useLayoutStore();
const { t } = useI18n();
const icons: Record<string, string> = {
  info: "info",
  help: "help_outline",
  upload: "cloud_upload",
  download: "file_download",
  share: "share",
  "share-delete": "link_off",
  copy: "content_copy",
  move: "drive_file_move",
  rename: "edit",
  newFile: "note_add",
  newDir: "create_new_folder",
  delete: "delete_outline",
  deleteUser: "person_remove",
  replace: "find_replace",
  "resolve-conflict": "difference",
  discardEditorChanges: "edit_off",
  "current-password": "lock",
};
const icon = computed(
  () => icons[layoutStore.currentPromptName || ""] || "folder_open"
);
</script>
