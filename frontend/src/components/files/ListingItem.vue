<template>
  <div
    class="item fb-sticker-item fn-file-card"
    role="button"
    tabindex="0"
    :draggable="isDraggable"
    @dragstart="dragStart"
    @dragover="dragOver"
    @drop="drop"
    @click="itemClick"
    @mousedown="handleMouseDown"
    @mouseup="handleMouseUp"
    @mouseleave="handleMouseLeave"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
    @touchcancel="handleTouchCancel"
    @touchmove="handleTouchMove"
    :data-dir="isDir"
    :data-type="type"
    :data-theme="itemTheme"
    :aria-label="name"
    :aria-selected="isSelected"
    :data-ext="getExtension(name).toLowerCase()"
    @contextmenu="contextMenu"
  >
    <!-- Top-Right Chamfered Folder Tab (Reference Card Header) -->
    <div class="card-folder-tab" v-if="isDir"></div>

    <!-- Bottom-Left Geometric Triangle Corner Slash -->
    <div class="card-corner-slash"></div>

    <!-- Top-Left Selection Check Box Badge -->
    <div class="select-check-badge" @click.stop="toggleSelect">
      <span class="select-check-icon">✓</span>
    </div>

    <!-- Inner Dashed Frame (Exact Student ID / Polaroid Sticker Card Layout) -->
    <div class="card-inner-dashed fn-file-card__surface">
      <div v-if="showThumbnail" class="card-media-banner">
        <img
          :src="thumbnailUrl"
          :alt="name"
          class="card-thumb-img"
          loading="lazy"
          decoding="async"
          @error="failedThumbnail = thumbnailUrl"
        />
        <div class="card-media-overlay">
          <div class="dot-matrix"></div>
          <!-- Cat Postmark SVG inside corner -->
          <svg class="cat-postmark-svg" viewBox="0 0 56 48" fill="none">
            <circle
              cx="22"
              cy="24"
              r="17"
              fill="#FFFFFF"
              stroke="currentColor"
              stroke-width="2"
            />
            <circle
              cx="22"
              cy="24"
              r="14"
              stroke="currentColor"
              stroke-width="1.2"
              stroke-dasharray="2 2"
            />
            <path
              d="M14 20L15.5 14L20 17.5H24L28.5 14L30 20C31 22.5 30.5 27.5 28 29.5C25.5 31.5 18.5 31.5 16 29.5C13.5 27.5 13 22.5 14 20Z"
              fill="#FFFFFF"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linejoin="round"
            />
            <circle cx="19" cy="24" r="1.3" fill="currentColor" />
            <circle cx="25" cy="24" r="1.3" fill="currentColor" />
            <path
              d="M21 26Q22 27.2 23 26"
              stroke="currentColor"
              stroke-width="1.4"
              stroke-linecap="round"
            />
            <path
              d="M36 18Q41 15 46 18T55 18"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
            />
            <path
              d="M36 24Q41 21 46 24T55 24"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
            />
            <path
              d="M36 30Q41 27 46 30T55 30"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
            />
          </svg>
        </div>
        <div class="card-pill-row">
          <span class="skew-badge">
            <span>{{ badgeText }}</span>
          </span>
        </div>
      </div>

      <!-- Non-image top row (Avatar/Icon + Postmark + Skew Tag) -->
      <div v-else class="card-top-row">
        <div class="card-icon-box">
          <i class="material-icons">{{ iconName }}</i>
          <span class="card-icon-badge" v-if="isDir">DIR</span>
          <span class="card-icon-star" v-else>★</span>
        </div>

        <div class="card-meta-col">
          <div class="card-stamp-row">
            <div class="dot-matrix"></div>
            <!-- Cat Postmark SVG -->
            <svg class="cat-postmark-svg" viewBox="0 0 56 48" fill="none">
              <circle
                cx="22"
                cy="24"
                r="17"
                fill="#FFFFFF"
                stroke="currentColor"
                stroke-width="2"
              />
              <circle
                cx="22"
                cy="24"
                r="14"
                stroke="currentColor"
                stroke-width="1.2"
                stroke-dasharray="2 2"
              />
              <path
                d="M14 20L15.5 14L20 17.5H24L28.5 14L30 20C31 22.5 30.5 27.5 28 29.5C25.5 31.5 18.5 31.5 16 29.5C13.5 27.5 13 22.5 14 20Z"
                fill="#FFFFFF"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linejoin="round"
              />
              <circle cx="19" cy="24" r="1.3" fill="currentColor" />
              <circle cx="25" cy="24" r="1.3" fill="currentColor" />
              <path
                d="M21 26Q22 27.2 23 26"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
              />
              <path
                d="M36 18Q41 15 46 18T55 18"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
              />
              <path
                d="M36 24Q41 21 46 24T55 24"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
              />
              <path
                d="M36 30Q41 27 46 30T55 30"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
              />
            </svg>
          </div>

          <div class="card-badge-row">
            <span class="skew-badge">
              <span>{{ badgeText }}</span>
            </span>
          </div>
        </div>
      </div>

      <!-- File/Folder Title and Details -->
      <div class="card-info-row">
        <p class="name" :title="name">{{ name }}</p>
        <div class="card-details-row">
          <p v-if="isDir" class="size" data-order="-1">&mdash;</p>
          <p v-else class="size" :data-order="humanSize()">{{ humanSize() }}</p>
          <p class="modified">
            <time :datetime="modified">{{ humanTime() }}</time>
          </p>
        </div>
      </div>

      <!-- Barcode + Signature Row -->
      <div class="card-barcode-row">
        <div class="css-barcode"></div>
        <span class="font-signature fn-file-card__suffix" lang="en">{{
          signatureText
        }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import { enableThumbs, enableVideoThumbs } from "@/utils/constants";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import { files as api } from "@/api";
import * as upload from "@/utils/upload";
import { computed, inject, ref } from "vue";
import { useRouter } from "vue-router";

const touches = ref<number>(0);

const longPressTimer = ref<number | null>(null);
const longPressTriggered = ref<boolean>(false);
const longPressDelay = ref<number>(500);
const startPosition = ref<{ x: number; y: number } | null>(null);
const moveThreshold = ref<number>(10);

const $showError = inject<IToastError>("$showError")!;
const router = useRouter();

const props = defineProps<{
  name: string;
  isDir: boolean;
  url: string;
  type: string;
  size: number;
  modified: string;
  index: number;
  readOnly?: boolean;
  path?: string;
}>();

const { t } = useI18n();
const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const itemTheme = computed(() => {
  if (props.isDir) return "cobalt";
  if (props.type === "image") return "pink";
  if (props.type === "audio" || props.type === "video") return "cyan";
  return "purple";
});

const badgeText = computed(() => {
  if (props.isDir) return t("interface.folder");
  const type = ["image", "audio", "video", "archive"].includes(props.type)
    ? props.type
    : ["pdf", "text", "textImmutable"].includes(props.type)
      ? "document"
      : "file";
  return t(`interface.${type}`);
});

const signatureText = computed(() =>
  props.isDir
    ? "Folder"
    : getExtension(props.name).replace(/^\./, "").toUpperCase() ||
      t("interface.file")
);

const iconName = computed(() => {
  if (props.isDir) return "folder";
  const iconMap: Record<string, string> = {
    image: "insert_photo",
    audio: "volume_up",
    video: "movie",
    archive: "inventory_2",
    pdf: "picture_as_pdf",
    text: "description",
    blob: "code",
  };
  return iconMap[props.type] || "insert_drive_file";
});

const toggleSelect = (event: MouseEvent) => {
  event.stopPropagation();
  if (isSelected.value) {
    fileStore.removeSelected(props.index);
  } else {
    fileStore.selected.push(props.index);
  }
};

const singleClick = computed(
  () => !props.readOnly && authStore.user?.singleClick
);
const isSelected = computed(
  () => fileStore.selected.indexOf(props.index) !== -1
);
const isDraggable = computed(
  () => !props.readOnly && authStore.user?.perm.rename
);

const canDrop = computed(() => {
  if (!props.isDir || props.readOnly) return false;

  for (const i of fileStore.selected) {
    if (fileStore.req?.items[i].url === props.url) {
      return false;
    }
  }

  return true;
});

const thumbnailUrl = computed(() => {
  const file = {
    path: props.path,
    modified: props.modified,
  };

  const size =
    props.type !== "video" && authStore.user?.viewMode === "mosaic gallery"
      ? "big"
      : "thumb";
  return api.getPreviewURL(file as Resource, size);
});

const isWebImage = computed(() => {
  const ext = getExtension(props.name).toLowerCase();
  return [".png", ".jpg", ".jpeg", ".svg", ".gif", ".webp"].includes(ext);
});

const failedThumbnail = ref("");
const showThumbnail = computed(
  () =>
    !props.isDir &&
    !props.readOnly &&
    enableThumbs &&
    !!authStore.user?.perm.download &&
    (props.type === "image"
      ? isWebImage.value
      : props.type === "video" &&
        enableVideoThumbs &&
        ["mosaic", "mosaic gallery"].includes(
          authStore.user?.viewMode ?? ""
        )) &&
    failedThumbnail.value !== thumbnailUrl.value
);

const humanSize = () => {
  return props.type == "invalid_link" ? "invalid link" : filesize(props.size);
};

const humanTime = () => {
  if (!props.readOnly && authStore.user?.dateFormat) {
    return dayjs(props.modified).format("L LT");
  }
  return dayjs(props.modified).fromNow();
};

const dragStart = () => {
  if (fileStore.selectedCount === 0) {
    fileStore.selected.push(props.index);
    return;
  }

  if (!isSelected.value) {
    fileStore.selected = [];
    fileStore.selected.push(props.index);
  }
};

const dragOver = (event: Event) => {
  if (!canDrop.value) return;

  event.preventDefault();
  let el = event.target as HTMLElement | null;
  if (el !== null) {
    for (let i = 0; i < 5; i++) {
      if (!el?.classList.contains("item")) {
        el = el?.parentElement ?? null;
      }
    }

    if (el !== null) el.style.opacity = "1";
  }
};

const drop = async (event: Event) => {
  if (!canDrop.value) return;
  event.preventDefault();

  if (fileStore.selectedCount === 0) return;

  let el = event.target as HTMLElement | null;
  for (let i = 0; i < 5; i++) {
    if (el !== null && !el.classList.contains("item")) {
      el = el.parentElement;
    }
  }

  const items: any[] = [];

  for (const i of fileStore.selected) {
    if (fileStore.req) {
      items.push({
        from: fileStore.req?.items[i].url,
        to: props.url + encodeURIComponent(fileStore.req?.items[i].name),
        name: fileStore.req?.items[i].name,
        size: fileStore.req?.items[i].size,
        isDir: fileStore.req?.items[i].isDir,
        modified: fileStore.req?.items[i].modified,
        overwrite: false,
        rename: false,
      });
    }
  }

  // Get url from ListingItem instance
  if (el === null) {
    return;
  }
  const path = el.__vue__.url;

  const action = (overwrite?: boolean, rename?: boolean) => {
    const action =
      (event as KeyboardEvent).ctrlKey || (event as KeyboardEvent).metaKey
        ? api.copy
        : api.move;
    action(items, overwrite, rename)
      .then(() => {
        fileStore.reload = true;
      })
      .catch($showError);
  };

  const conflict = await upload.checkConflict(items, path, true);

  if (conflict.length > 0) {
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
      },
      confirm: (event: Event, result: Array<ConflictingResource>) => {
        event.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            items[item.index].rename = true;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            items[item.index].overwrite = true;
          } else {
            items.splice(item.index, 1);
          }
        }
        if (items.length > 0) {
          action();
        }
      },
    });

    return;
  }

  action(false, false);
};

const itemClick = (event: Event | KeyboardEvent) => {
  // If long press was triggered, prevent normal click behavior
  if (longPressTriggered.value) {
    longPressTriggered.value = false;
    return;
  }

  if (
    singleClick.value &&
    !(event as KeyboardEvent).ctrlKey &&
    !(event as KeyboardEvent).metaKey &&
    !(event as KeyboardEvent).shiftKey &&
    !fileStore.multiple
  )
    open();
  else click(event);
};

const contextMenu = (event: MouseEvent) => {
  event.preventDefault();
  if (
    fileStore.selected.length === 0 ||
    event.ctrlKey ||
    fileStore.selected.indexOf(props.index) === -1
  ) {
    click(event);
  }
};

const click = (event: Event | KeyboardEvent) => {
  if (!singleClick.value && fileStore.selectedCount !== 0)
    event.preventDefault();

  setTimeout(() => {
    touches.value = 0;
  }, 300);

  touches.value++;
  if (touches.value > 1) {
    open();
  }

  if (fileStore.selected.indexOf(props.index) !== -1) {
    if (
      (event as KeyboardEvent).ctrlKey ||
      (event as KeyboardEvent).metaKey ||
      fileStore.multiple
    ) {
      fileStore.removeSelected(props.index);
    } else {
      fileStore.selected = [props.index];
    }
    return;
  }

  if ((event as KeyboardEvent).shiftKey && fileStore.selected.length > 0) {
    let fi = 0;
    let la = 0;

    if (props.index > fileStore.selected[0]) {
      fi = fileStore.selected[0] + 1;
      la = props.index;
    } else {
      fi = props.index;
      la = fileStore.selected[0] - 1;
    }

    for (; fi <= la; fi++) {
      if (fileStore.selected.indexOf(fi) == -1) {
        fileStore.selected.push(fi);
      }
    }

    return;
  }

  if (
    !(event as KeyboardEvent).ctrlKey &&
    !(event as KeyboardEvent).metaKey &&
    !fileStore.multiple
  ) {
    fileStore.selected = [];
  }
  fileStore.selected.push(props.index);
};

const open = () => {
  router.push({ path: props.url });
};

const getExtension = (fileName: string): string => {
  const lastDotIndex = fileName.lastIndexOf(".");
  if (lastDotIndex === -1) {
    return fileName;
  }
  return fileName.substring(lastDotIndex);
};

// Long-press helper functions
const startLongPress = (clientX: number, clientY: number) => {
  startPosition.value = { x: clientX, y: clientY };
  longPressTimer.value = window.setTimeout(() => {
    handleLongPress();
  }, longPressDelay.value);
};

const cancelLongPress = () => {
  if (longPressTimer.value !== null) {
    window.clearTimeout(longPressTimer.value);
    longPressTimer.value = null;
  }
  startPosition.value = null;
};

const handleLongPress = () => {
  if (singleClick.value) {
    longPressTriggered.value = true;
    click(new Event("longpress"));
  }
  cancelLongPress();
};

const checkMovement = (clientX: number, clientY: number): boolean => {
  if (!startPosition.value) return false;

  const deltaX = Math.abs(clientX - startPosition.value.x);
  const deltaY = Math.abs(clientY - startPosition.value.y);

  return deltaX > moveThreshold.value || deltaY > moveThreshold.value;
};

// Event handlers
const handleMouseDown = (event: MouseEvent) => {
  if (event.button === 0) {
    startLongPress(event.clientX, event.clientY);
  }
};

const handleMouseUp = () => {
  cancelLongPress();
};

const handleMouseLeave = () => {
  cancelLongPress();
};

const handleTouchStart = (event: TouchEvent) => {
  if (event.touches.length === 1) {
    const touch = event.touches[0];
    startLongPress(touch.clientX, touch.clientY);
  }
};

const handleTouchEnd = () => {
  cancelLongPress();
};

const handleTouchCancel = () => {
  cancelLongPress();
};

const handleTouchMove = (event: TouchEvent) => {
  if (event.touches.length === 1 && startPosition.value) {
    const touch = event.touches[0];
    if (checkMovement(touch.clientX, touch.clientY)) {
      cancelLongPress();
    }
  }
};
</script>
