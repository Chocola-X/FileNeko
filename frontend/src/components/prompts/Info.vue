<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.fileInfo") }}</h2>
    </div>

    <div class="card-content">
      <p v-if="selected.length > 1">
        {{ $t("prompts.filesSelected", { count: selected.length }) }}
      </p>

      <p class="break-word" v-if="selected.length < 2">
        <strong>{{ $t("prompts.displayName") }}</strong> {{ name }}
      </p>

      <p v-if="!dir || selected.length > 1 || scanStats">
        <strong>{{ $t("prompts.size") }}:</strong>
        <span id="content_length"></span> {{ humanSize }}
      </p>

      <div v-if="resolution">
        <strong>{{ $t("prompts.resolution") }}:</strong>
        {{ resolution.width }} x {{ resolution.height }}
      </div>

      <p v-if="selected.length < 2" :title="modTime">
        <strong>{{ $t("prompts.lastModified") }}:</strong> {{ humanTime }}
      </p>

      <template v-if="scanStats">
        <p>
          <strong>{{ $t("prompts.numberFiles") }}:</strong>
          {{ scanStats.numFiles }}
        </p>
        <p>
          <strong>{{ $t("prompts.numberDirs") }}:</strong>
          {{ scanStats.numDirs }}
        </p>
        <div class="folder-scan" :aria-busy="scanActive">
          <p class="folder-scan-status" role="status" aria-live="polite">
            <i v-if="scanActive" class="material-icons" aria-hidden="true"
              >hourglass_top</i
            >
            {{ scanStatusText }}
          </p>
          <p class="folder-scan-note">{{ $t("prompts.folderScanNote") }}</p>
          <button
            v-if="scanActive"
            type="button"
            class="button button--flat"
            @click="stopScan"
          >
            {{ $t("buttons.stopScan") }}
          </button>
        </div>
      </template>
      <p v-else-if="hasFolders" class="folder-scan-note">
        {{ $t("prompts.folderScanLinks") }}
      </p>

      <template v-if="!dir">
        <p>
          <strong>MD5: </strong
          ><code
            ><a
              @click="checksum($event, 'md5')"
              @keypress.enter="checksum($event, 'md5')"
              tabindex="2"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA1: </strong
          ><code
            ><a
              @click="checksum($event, 'sha1')"
              @keypress.enter="checksum($event, 'sha1')"
              tabindex="3"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA256: </strong
          ><code
            ><a
              @click="checksum($event, 'sha256')"
              @keypress.enter="checksum($event, 'sha256')"
              tabindex="4"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA512: </strong
          ><code
            ><a
              @click="checksum($event, 'sha512')"
              @keypress.enter="checksum($event, 'sha512')"
              tabindex="5"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
      </template>
    </div>

    <div class="card-action">
      <button
        id="focus-prompt"
        type="submit"
        @click="closeHovers"
        class="button button--flat"
        :aria-label="$t('buttons.ok')"
        :title="$t('buttons.ok')"
      >
        {{ $t("buttons.ok") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import { files as api } from "@/api";

export default {
  name: "info",
  inject: ["$showError"],
  data() {
    return {
      scanController: null,
      scanStats: null,
      scanState: "idle",
    };
  },
  activated() {
    this.startScan();
  },
  deactivated() {
    this.stopScan();
  },
  beforeUnmount() {
    this.stopScan();
  },
  watch: {
    scanKey() {
      if (this.currentPromptName === "info") this.startScan();
    },
  },
  computed: {
    ...mapState(useLayoutStore, ["currentPromptName"]),
    ...mapState(useFileStore, [
      "req",
      "selected",
      "selectedCount",
      "isListing",
    ]),
    infoItems() {
      return this.selectedCount > 0 && this.isListing
        ? this.selected.map((index) => this.req.items[index]).filter(Boolean)
        : [this.req];
    },
    scanKey() {
      return JSON.stringify(
        this.infoItems.map((item) => [item.url, item.isDir, item.isSymlink])
      );
    },
    hasFolders() {
      return this.infoItems.some((item) => item.isDir);
    },
    scanActive() {
      return this.scanState === "scanning" || this.scanState === "queued";
    },
    scanStatusText() {
      const keys = {
        scanning: "folderScanRunning",
        queued: "folderScanQueued",
        complete: this.scanStats?.incomplete
          ? "folderScanIncomplete"
          : "folderScanComplete",
        failed: "folderScanFailed",
        canceled: "folderScanCanceled",
      };
      return this.$t(`prompts.${keys[this.scanState] || "folderScanRunning"}`);
    },
    humanSize: function () {
      if (this.scanStats) return filesize(this.scanStats.size);
      if (this.hasFolders) return "—";
      if (this.selectedCount === 0 || !this.isListing) {
        return filesize(this.req.size);
      }

      let sum = 0;

      for (const selected of this.selected) {
        sum += this.req.items[selected].size;
      }

      return filesize(sum);
    },
    humanTime: function () {
      if (this.selectedCount === 0) {
        return dayjs(this.req.modified).fromNow();
      }

      return dayjs(this.req.items[this.selected[0]].modified).fromNow();
    },
    modTime: function () {
      if (this.selectedCount === 0) {
        return new Date(Date.parse(this.req.modified)).toLocaleString();
      }

      return new Date(
        Date.parse(this.req.items[this.selected[0]].modified)
      ).toLocaleString();
    },
    name: function () {
      return this.selectedCount === 0
        ? this.req.name
        : this.req.items[this.selected[0]].name;
    },
    dir: function () {
      return (
        this.selectedCount > 1 ||
        (this.selectedCount === 0
          ? this.req.isDir
          : this.req.items[this.selected[0]].isDir)
      );
    },
    resolution: function () {
      if (this.selectedCount === 1) {
        const selectedItem = this.req.items[this.selected[0]];
        if (selectedItem && selectedItem.type === "image") {
          return selectedItem.resolution;
        }
      } else if (this.req && this.req.type === "image") {
        return this.req.resolution;
      }
      return null;
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    stopScan() {
      if (!this.scanController) return;
      this.scanController.abort();
      this.scanController = null;
      if (this.scanActive) this.scanState = "canceled";
    },
    async startScan() {
      this.stopScan();
      this.scanStats = null;
      this.scanState = "idle";
      const folders = this.infoItems.filter(
        (item) => item.isDir && !item.isSymlink
      );
      if (folders.length === 0) return;
      const controller = new AbortController();
      this.scanController = controller;
      const files = this.infoItems.filter(
        (item) => !item.isDir && !item.isSymlink
      );
      let totals = {
        size: files.reduce((sum, item) => sum + item.size, 0),
        numFiles: files.length,
        // For a single folder, count its contents; for multiple selections,
        // include the selected folder roots as well as their descendants.
        numDirs: this.infoItems.length > 1 ? folders.length : 0,
        incomplete: false,
      };
      this.scanStats = { ...totals };
      this.scanState = "scanning";
      try {
        for (const folder of folders) {
          const base = { ...totals };
          const result = await api.scanDirectory(
            folder.url,
            controller.signal,
            (stats) => {
              if (
                this.scanController !== controller ||
                controller.signal.aborted
              )
                return;
              this.scanStats = {
                size: base.size + stats.size,
                numFiles: base.numFiles + stats.numFiles,
                numDirs: base.numDirs + stats.numDirs,
                incomplete: base.incomplete || stats.incomplete,
              };
              this.scanState = stats.queued ? "queued" : "scanning";
            }
          );
          if (this.scanController !== controller || controller.signal.aborted)
            return;
          totals = {
            size: base.size + result.size,
            numFiles: base.numFiles + result.numFiles,
            numDirs: base.numDirs + result.numDirs,
            incomplete: base.incomplete || result.incomplete,
          };
        }
        this.scanStats = totals;
        this.scanState = "complete";
      } catch {
        if (this.scanController === controller && !controller.signal.aborted) {
          this.scanState = "failed";
        }
      } finally {
        if (this.scanController === controller) this.scanController = null;
      }
    },
    checksum: async function (event, algo) {
      event.preventDefault();

      let link;

      if (this.selectedCount) {
        link = this.req.items[this.selected[0]].url;
      } else {
        link = this.$route.path;
      }

      try {
        const hash = await api.checksum(link, algo);
        event.target.textContent = hash;
      } catch (e) {
        this.$showError(e);
      }
    },
  },
};
</script>

<style scoped>
.folder-scan {
  margin-top: 16px;
  padding: 12px;
  border: 2px dashed var(--cobalt-light);
  border-radius: 12px;
  background: var(--surface-ice);
}
.folder-scan-status {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0 0 8px;
  color: var(--cobalt-dark);
  font-weight: 700;
}
.folder-scan-note {
  margin: 0;
  color: var(--ink-slate);
  font-size: 0.85em;
  line-height: 1.6;
  overflow-wrap: anywhere;
}
.folder-scan button {
  margin-top: 10px;
}
</style>
