<template>
  <status-card class="folder-empty" icon="folder_open" label="EMPTY FOLDER">{{
    t("files.lonely")
  }}</status-card>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import StatusCard from "@/components/StatusCard.vue";

const { t } = useI18n();
</script>
