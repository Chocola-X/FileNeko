<template>
  <button
    @click="action"
    :aria-label="label"
    :title="label"
    :data-tone="tone"
    :data-icon="icon"
    class="action"
  >
    <i class="material-icons">{{ icon }}</i>
    <span>{{ label }}</span>
    <span v-if="counter && counter > 0" class="counter">{{ counter }}</span>
  </button>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { computed } from "vue";

const props = defineProps<{
  icon?: string;
  label?: string;
  counter?: number;
  show?: string;
}>();

const emit = defineEmits<{
  (e: "action"): any;
}>();

const layoutStore = useLayoutStore();

const tones: Record<string, string> = {
  share: "cyan",
  file_download: "cyan",
  forward: "cyan",
  content_copy: "purple",
  code: "purple",
  info: "purple",
  file_upload: "pink",
  delete: "pink",
  help: "gold",
  help_outline: "gold",
};
const tone = computed(() => tones[props.icon || ""] || "cobalt");

const action = () => {
  if (props.show) {
    layoutStore.showHover(props.show);
  }

  emit("action");
};
</script>
