<template>
  <teleport to="body">
    <div v-if="show" ref="contextMenu" class="context-menu">
      <slot />
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { ref, watchPostEffect } from "vue";

const emit = defineEmits(["hide"]);
const props = defineProps<{ show: boolean; pos: { x: number; y: number } }>();
const contextMenu = ref<HTMLElement | null>(null);

const hideContextMenu = () => {
  emit("hide");
};

watchPostEffect((onCleanup) => {
  const menu = contextMenu.value;
  if (!props.show || !menu) return;

  const { x, y } = props.pos;
  const viewport = window.visualViewport;
  const place = () => {
    // Leave room for the theme's shadow as well as the menu's border box.
    const margin = 8;
    const root = document.documentElement;
    const width = Math.min(
      root.clientWidth,
      viewport?.width ?? root.clientWidth
    );
    const height = Math.min(
      root.clientHeight,
      viewport?.height ?? root.clientHeight
    );
    const left = (viewport?.offsetLeft ?? 0) + margin;
    const top = (viewport?.offsetTop ?? 0) + margin;
    const maxWidth = Math.max(0, width - 2 * margin);
    const maxHeight = Math.max(0, height - 2 * margin);
    menu.style.setProperty("--context-menu-available-width", `${maxWidth}px`);
    menu.style.maxHeight = `${maxHeight}px`;

    const rect = menu.getBoundingClientRect();
    menu.style.left = `${Math.max(left, Math.min(x, left + maxWidth - rect.width))}px`;
    menu.style.top = `${Math.max(top, Math.min(y, top + maxHeight - rect.height))}px`;
  };

  place();
  const observer = new ResizeObserver(place);
  observer.observe(menu);
  window.addEventListener("resize", place);
  viewport?.addEventListener("resize", place);
  viewport?.addEventListener("scroll", place);
  document.addEventListener("click", hideContextMenu);
  onCleanup(() => {
    observer.disconnect();
    window.removeEventListener("resize", place);
    viewport?.removeEventListener("resize", place);
    viewport?.removeEventListener("scroll", place);
    document.removeEventListener("click", hideContextMenu);
  });
});
</script>
