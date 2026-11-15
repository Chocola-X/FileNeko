<template>
  <div class="destination-picker" :aria-busy="navigating">
    <ul class="file-list">
      <li :key="item.url" v-for="item in items">
        <button
          type="button"
          class="destination-folder"
          @click="next"
          :aria-label="item.name"
          :data-url="item.url"
          :disabled="navigating"
        >
          <i
            class="material-icons destination-folder-icon"
            aria-hidden="true"
            >{{ item.name === ".." ? "drive_folder_upload" : "folder" }}</i
          >
          <span>{{ item.name }}</span>
          <i class="material-icons destination-folder-arrow" aria-hidden="true"
            >chevron_right</i
          >
        </button>
      </li>
    </ul>

    <p class="destination-path" aria-live="polite">
      <span>{{ $t("prompts.currentlyNavigating") }}</span>
      <code>{{ nav }}</code>
    </p>
  </div>
</template>

<script>
import { mapState, mapActions } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import url from "@/utils/url";
import { files } from "@/api";
import { StatusError } from "@/api/utils.js";

export default {
  name: "file-list",
  emits: ["update:selected", "update:navigating"],
  props: {
    exclude: {
      type: Array,
      default: () => [],
    },
  },
  data: function () {
    return {
      items: [],
      navigating: false,
      current: window.location.pathname,
      nextAbortController: new AbortController(),
    };
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useFileStore, ["req"]),
    nav() {
      return decodeURIComponent(this.current);
    },
  },
  mounted() {
    this.fillOptions(this.req);
  },
  unmounted() {
    this.abortOngoingNext();
  },
  methods: {
    ...mapActions(useLayoutStore, ["showHover"]),
    abortOngoingNext() {
      this.nextAbortController.abort();
    },
    fillOptions(req) {
      // Sets the current path and resets
      // the current items.
      this.current = req.url;
      this.items = [];

      this.$emit("update:selected", this.current);

      // If the path isn't the root path,
      // show a button to navigate to the previous
      // directory.
      if (req.url !== "/files/") {
        this.items.push({
          name: "..",
          url: url.removeLastDir(req.url) + "/",
        });
      }

      // If this folder is empty, finish here.
      if (req.items === null) return;

      // Otherwise we add every directory to the
      // move options.
      for (const item of req.items) {
        if (!item.isDir) continue;
        if (this.exclude?.includes(item.url)) continue;

        this.items.push({
          name: item.name,
          url: item.url,
        });
      }
    },
    next: async function (event) {
      // A native button emits one click for touch and keyboard activation.
      // Ignore the second click of a double click, even if the first request
      // already replaced the row with a child directory in the same position.
      if (this.navigating || event.detail > 1) return;
      // Retrieves the URL of the directory the user
      // just clicked in and fill the options with its
      // content.
      const uri = event.currentTarget.dataset.url;
      this.abortOngoingNext();
      const controller = new AbortController();
      this.nextAbortController = controller;
      this.navigating = true;
      this.$emit("update:navigating", true);
      try {
        const req = await files.fetch(uri, controller.signal);
        if (!controller.signal.aborted) this.fillOptions(req);
      } catch (e) {
        if (
          controller.signal.aborted ||
          (e instanceof StatusError && e.is_canceled)
        ) {
          return;
        }
        this.$showError(e);
      } finally {
        this.navigating = false;
        this.$emit("update:navigating", false);
      }
    },
    createDir: async function () {
      this.showHover({
        prompt: "newDir",
        action: null,
        confirm: (url) => {
          const paths = url.split("/");
          this.items.push({
            name: paths[paths.length - 2],
            url: url,
          });
        },
        props: {
          redirect: false,
          base: this.current === this.$route.path ? null : this.current,
        },
      });
    },
  },
};
</script>
