<template>
  <div
    class="file-listing public-share fn-page--share"
    :style="{
      '--selection-dock-space': dockHeight ? `${dockHeight + 32}px` : '0px',
    }"
  >
    <header-bar showMenu showLogo>
      <title />
      <action
        v-if="fileStore.selectedCount"
        icon="file_download"
        :label="t('buttons.download')"
        @action="download"
        :counter="fileStore.selectedCount"
      />
      <button
        v-if="isSingleFile()"
        class="action copy-clipboard"
        :aria-label="t('buttons.copyDownloadLinkToClipboard')"
        :title="t('buttons.copyDownloadLinkToClipboard')"
        @click="copyToClipboard(linkSelected())"
      >
        <i class="material-icons">content_paste</i>
      </button>
      <action
        v-if="req?.isDir && !error"
        icon="check_circle"
        :label="t('buttons.selectMultiple')"
        @action="toggleMultipleSelection"
      />
    </header-bar>

    <breadcrumbs :base="'/share/' + hash" />
    <div v-if="fileStore.multiple" ref="selectionDock" class="selection-dock">
      <selection-notice />
    </div>

    <loading-indicator v-if="layoutStore.loading" />
    <div v-else-if="error">
      <form
        v-if="error.status === 401"
        id="password"
        class="public-share-card fn-share-card share-password"
        data-tone="purple"
        @submit.prevent="fetchData"
      >
        <div class="public-share-title">
          <i class="material-icons" aria-hidden="true">lock</i>
          <h2>{{ t("login.password") }}</h2>
        </div>
        <div class="public-share-body">
          <p
            v-if="attemptedPasswordLogin"
            class="share-password-error"
            role="alert"
          >
            {{ t("login.wrongCredentials") }}
          </p>
          <label for="share-password">{{ t("login.password") }}</label>
          <input
            id="share-password"
            v-focus
            class="input input--block"
            type="password"
            autocomplete="current-password"
            :placeholder="t('login.password')"
            v-model="password"
            required
          />
          <div class="share-download-actions">
            <button class="button" type="submit">
              {{ t("buttons.submit") }}
            </button>
          </div>
        </div>
      </form>
      <errors v-else :errorCode="error.status" />
    </div>
    <div v-else-if="req !== null" class="public-share-content">
      <div class="public-share-heading">
        <span aria-hidden="true">FileNeko / PUBLIC SHARE</span>
        <h1>{{ t("buttons.share") }}<span aria-hidden="true">✦</span></h1>
      </div>
      <div class="public-share-layout" :class="{ 'is-single': !req.isDir }">
        <section
          class="public-share-card fn-share-card share-summary"
          :data-tone="req.isDir ? 'cobalt' : 'purple'"
        >
          <div class="public-share-title">
            <i class="material-icons" aria-hidden="true">{{ icon }}</i>
            <h2>
              {{
                req.isDir
                  ? t("download.downloadFolder")
                  : t("download.downloadFile")
              }}
            </h2>
          </div>
          <div class="public-share-body">
            <dl class="share-metadata">
              <div>
                <dt>{{ t("prompts.displayName") }}</dt>
                <dd class="share-resource-name">{{ req.name }}</dd>
              </div>
              <template v-if="req.isDir">
                <div>
                  <dt>{{ t("prompts.numberDirs") }}</dt>
                  <dd>{{ req.items.filter((item) => item.isDir).length }}</dd>
                </div>
                <div>
                  <dt>{{ t("prompts.numberFiles") }}</dt>
                  <dd>{{ req.items.filter((item) => !item.isDir).length }}</dd>
                </div>
              </template>
              <template v-else>
                <div>
                  <dt>{{ t("prompts.size") }}</dt>
                  <dd>{{ humanSize }}</dd>
                </div>
                <div>
                  <dt>{{ t("prompts.lastModified") }}</dt>
                  <dd :title="modTime">{{ humanTime }}</dd>
                </div>
              </template>
            </dl>
            <div class="share-download-actions">
              <a
                :href="link"
                class="button"
                data-tone="pink"
                target="_blank"
                rel="noopener noreferrer"
              >
                <i class="material-icons" aria-hidden="true">file_download</i>
                {{ t("buttons.download") }}
              </a>
              <a
                v-if="!req.isDir"
                :href="inlineLink"
                class="button"
                data-tone="cyan"
                target="_blank"
                rel="noopener noreferrer"
              >
                <i class="material-icons" aria-hidden="true">open_in_new</i>
                {{ t("buttons.openFile") }}
              </a>
            </div>
            <details class="share-qr">
              <summary>
                <i class="material-icons" aria-hidden="true">qr_code_2</i>
                {{ t("download.scanToDownload") }}
                <i class="material-icons share-qr-chevron" aria-hidden="true"
                  >expand_more</i
                >
              </summary>
              <div
                class="share-qr-image"
                role="img"
                :aria-label="t('download.scanToDownload')"
              >
                <qrcode-vue
                  :value="link"
                  :size="136"
                  level="M"
                  foreground="#1E2A4A"
                />
              </div>
            </details>
            <section
              v-if="
                previewFile &&
                ['image', 'audio', 'video'].includes(previewFile.type)
              "
              class="share-preview"
              data-tone="cyan"
            >
              <h3>
                <i class="material-icons" aria-hidden="true">visibility</i
                >{{ t("sidebar.preview") }}
              </h3>
              <p v-if="req.isDir" class="share-preview-name">
                {{ previewFile.name }}
              </p>
              <a
                v-if="previewFile.type === 'image'"
                :href="previewURL"
                target="_blank"
                rel="noopener noreferrer"
                class="share-preview-image"
              >
                <img :src="previewURL" :alt="previewFile.name" />
              </a>
              <audio
                v-else-if="previewFile.type === 'audio'"
                :key="previewURL"
                :src="previewURL"
                controls
                preload="metadata"
                :aria-label="previewFile.name"
              />
              <video
                v-else-if="previewFile.type === 'video'"
                :key="previewURL"
                :src="previewURL"
                controls
                preload="metadata"
                :aria-label="previewFile.name"
              />
            </section>
          </div>
        </section>
        <section
          v-if="req.isDir"
          id="shareList"
          class="public-share-card fn-share-card share-files"
          data-tone="cobalt"
        >
          <div class="public-share-title">
            <i class="material-icons" aria-hidden="true">folder_open</i>
            <h2>{{ t("files.files") }}</h2>
            <span class="share-count">{{ req.items.length }}</span>
          </div>
          <div v-if="req.items.length > 0" id="listing" class="list file-icons">
            <div class="public-share-rows">
              <item
                v-for="item in req.items.slice(0, showLimit)"
                :key="base64(item.name)"
                :index="item.index"
                :name="item.name"
                :isDir="item.isDir"
                :url="item.url"
                :modified="item.modified"
                :type="item.type"
                :size="item.size"
                readOnly
              />
            </div>
            <button
              v-if="req.items.length > showLimit"
              class="button share-load-more"
              type="button"
              @click="showLimit += 100"
            >
              {{ t("buttons.more") }} (+{{ req.items.length - showLimit }})
            </button>
          </div>
          <div v-else class="public-share-body"><empty-folder /></div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import LoadingIndicator from "@/components/LoadingIndicator.vue";
import { pub as api } from "@/api";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import { Base64 } from "js-base64";
import { createURL } from "@/api/utils";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import QrcodeVue from "qrcode.vue";
import Item from "@/components/files/ListingItem.vue";
import EmptyFolder from "@/components/files/EmptyFolder.vue";
import SelectionNotice from "@/components/files/SelectionNotice.vue";
import { useElementSize } from "@vueuse/core";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { computed, inject, onMounted, onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { copy } from "@/utils/clipboard";

const selectionDock = ref<HTMLElement | null>(null);
const { height: dockHeight } = useElementSize(selectionDock);

const error = ref<StatusError | null>(null);
const showLimit = ref<number>(100);
const password = ref<string>("");
const attemptedPasswordLogin = ref<boolean>(false);
const hash = ref<string>("");

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n({});

const route = useRoute();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

watch(route, () => {
  showLimit.value = 100;
  fetchData();
});

const req = computed(() => fileStore.req);

// Define computes

const icon = computed(() => {
  if (req.value === null) return "insert_drive_file";
  if (req.value.isDir) return "folder";
  if (req.value.type === "image") return "insert_photo";
  if (req.value.type === "audio") return "volume_up";
  if (req.value.type === "video") return "movie";
  return "insert_drive_file";
});

const link = computed(() => (req.value ? api.getDownloadURL(req.value) : ""));
const selectedItem = computed(() =>
  fileStore.selectedCount === 1 ? req.value?.items[fileStore.selected[0]] : null
);
const inlineLink = computed(() =>
  req.value ? api.getDownloadURL(req.value, true) : ""
);
const previewFile = computed(() =>
  req.value?.isDir ? selectedItem.value : req.value
);
const previewURL = computed(() => {
  if (!req.value) return "";
  if (!req.value.isDir) return inlineLink.value;
  return selectedItem.value
    ? createURL(`api/public/dl/${hash.value}${selectedItem.value.path}`, {
        inline: "true",
      })
    : "";
});
const humanSize = computed(() => {
  if (req.value) {
    return req.value.isDir
      ? req.value.items.length
      : filesize(req.value.size ?? 0);
  } else {
    return "";
  }
});
const humanTime = computed(() => dayjs(req.value?.modified).fromNow());
const modTime = computed(() =>
  req.value
    ? new Date(Date.parse(req.value.modified)).toLocaleString()
    : new Date().toLocaleString()
);

// Functions
const base64 = (name: any) => Base64.encodeURI(name);
const fetchData = async () => {
  fileStore.reload = false;
  fileStore.selected = [];
  fileStore.multiple = false;
  layoutStore.closeHovers();

  // Set loading to true and reset the error.
  layoutStore.loading = true;
  error.value = null;
  if (password.value !== "") {
    attemptedPasswordLogin.value = true;
  }

  const routePath = route.params.path;
  hash.value = Array.isArray(routePath)
    ? routePath[0]
    : (routePath || "").split("/")[0];

  let url = route.path;
  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;

  try {
    const file = await api.fetch(url, password.value);
    file.hash = hash.value;

    fileStore.updateRequest(file);
    document.title = `${file.name} - ${document.title}`;
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
};

const keyEvent = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    // If we're on a listing, unselect all
    // files and folders.
    if (fileStore.selectedCount > 0) {
      fileStore.selected = [];
    }
  }
};

const toggleMultipleSelection = () => {
  fileStore.toggleMultiple();
};

const isSingleFile = () =>
  fileStore.selectedCount === 1 &&
  !req.value?.items[fileStore.selected[0]].isDir;

const download = () => {
  if (!req.value) return false;

  if (isSingleFile()) {
    api.download(null, hash.value, req.value.items[fileStore.selected[0]].path);
    return true;
  }

  layoutStore.showHover({
    prompt: "download",
    confirm: (format: DownloadFormat) => {
      if (req.value === null) return false;
      layoutStore.closeHovers();

      const files: string[] = [];

      for (const i of fileStore.selected) {
        files.push(req.value.items[i].path);
      }

      api.download(format, hash.value, ...files);
      return true;
    },
  });

  return true;
};

const linkSelected = () => {
  return isSingleFile() && req.value
    ? api.getDownloadURL({
        ...req.value,
        hash: hash.value,
        path: req.value.items[fileStore.selected[0]].path,
      })
    : "";
};

const copyToClipboard = (text: string) => {
  copy({ text }).then(
    () => {
      // clipboard successfully set
      $showSuccess(t("success.linkCopied"));
    },
    () => {
      // clipboard write failed
      copy({ text }, { permission: true }).then(
        () => {
          // clipboard successfully set
          $showSuccess(t("success.linkCopied"));
        },
        (e) => {
          // clipboard write failed
          $showError(e);
        }
      );
    }
  );
};

onMounted(async () => {
  // Created
  window.addEventListener("keydown", keyEvent);
  await fetchData();
});

onBeforeUnmount(() => {
  // Destroyed
  window.removeEventListener("keydown", keyEvent);
});
</script>
