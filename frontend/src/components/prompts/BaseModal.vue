<template>
  <div
    id="modal-background"
    class="theme-backdrop"
    :data-prompt="prompt"
    :data-tone="tone"
    @click="backgroundClick"
  >
    <div ref="modalContainer" role="dialog" aria-modal="true" tabindex="-1">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";

const props = defineProps<{ prompt?: string | null }>();
const tones: Record<string, string> = {
  upload: "pink",
  share: "cyan",
  "share-delete": "pink",
  copy: "purple",
  move: "cyan",
  newFile: "purple",
  delete: "pink",
  deleteUser: "pink",
  replace: "gold",
  "resolve-conflict": "gold",
  discardEditorChanges: "gold",
  "current-password": "purple",
};
const tone = computed(() => tones[props.prompt || ""] || "cobalt");

const emit = defineEmits(["closed"]);

const modalContainer = ref(null);

onMounted(() => {
  const element = document.querySelector("#focus-prompt") as HTMLElement | null;
  if (element) {
    element.focus();
  } else if (modalContainer.value) {
    (modalContainer.value as HTMLElement).focus();
  }
});

const backgroundClick = (event: Event) => {
  const target = event.target as HTMLElement;
  if (target.id == "modal-background") {
    emit("closed");
  }
};

const keydown = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    event.stopImmediatePropagation();
    emit("closed");
  }
};
onMounted(() => window.addEventListener("keydown", keydown));
onUnmounted(() => window.removeEventListener("keydown", keydown));
</script>

<style scoped>
@layer components {
  #modal-background {
    position: fixed;
    inset: 0;
    background-color: #00000096;
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 10000;
    animation: ease-in 150ms opacity-enter;
  }

  @keyframes opacity-enter {
    from {
      opacity: 0;
    }

    to {
      opacity: 1;
    }
  }
}
</style>
