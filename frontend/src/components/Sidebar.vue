<template>
  <div
    v-show="active"
    @click="closeHovers"
    class="overlay theme-backdrop"
  ></div>
  <nav id="sidebar" :class="{ active }">
    <div class="sidebar-content">
      <!-- File navigation -->
      <div class="sidebar-badge-banner">
        <div class="sidebar-badge-pill">
          <span class="badge-tag">{{ $t("interface.navigation") }}</span>
          <span class="badge-title">{{ $t("sidebar.myFiles") }}</span>
        </div>
      </div>

      <template v-if="isLoggedIn">
        <button @click="toAccountSettings" class="action" data-tone="purple">
          <i class="material-icons">person</i>
          <span>{{ user.username }}</span>
        </button>
        <button
          class="action"
          data-tone="cobalt"
          :class="{ active: $route.path.startsWith('/files') }"
          @click="toRoot"
          :aria-label="$t('sidebar.myFiles')"
          :title="$t('sidebar.myFiles')"
        >
          <i class="material-icons">folder</i>
          <span>{{ $t("sidebar.myFiles") }}</span>
        </button>

        <div v-if="user.perm.create">
          <button
            @click="showHover('newDir')"
            class="action"
            data-tone="cyan"
            :aria-label="$t('sidebar.newFolder')"
            :title="$t('sidebar.newFolder')"
          >
            <i class="material-icons">create_new_folder</i>
            <span>{{ $t("sidebar.newFolder") }}</span>
          </button>

          <button
            @click="showHover('newFile')"
            class="action"
            data-tone="purple"
            :aria-label="$t('sidebar.newFile')"
            :title="$t('sidebar.newFile')"
          >
            <i class="material-icons">note_add</i>
            <span>{{ $t("sidebar.newFile") }}</span>
          </button>
        </div>

        <div v-if="user.perm.admin">
          <button
            class="action"
            data-tone="pink"
            :class="{ active: $route.path.startsWith('/settings') }"
            @click="toGlobalSettings"
            :aria-label="$t('sidebar.settings')"
            :title="$t('sidebar.settings')"
          >
            <i class="material-icons">settings_applications</i>
            <span>{{ $t("sidebar.settings") }}</span>
          </button>
        </div>
        <button
          v-if="canLogout"
          @click="logout"
          class="action"
          data-tone="mint"
          id="logout"
          :aria-label="$t('sidebar.logout')"
          :title="$t('sidebar.logout')"
        >
          <i class="material-icons">exit_to_app</i>
          <span>{{ $t("sidebar.logout") }}</span>
        </button>
      </template>
      <template v-else>
        <router-link
          v-if="!hideLoginButton"
          class="action"
          data-tone="cyan"
          to="/login"
          :aria-label="$t('sidebar.login')"
          :title="$t('sidebar.login')"
        >
          <i class="material-icons">exit_to_app</i>
          <span>{{ $t("sidebar.login") }}</span>
        </router-link>

        <router-link
          v-if="signup"
          class="action"
          data-tone="purple"
          to="/login"
          :aria-label="$t('sidebar.signup')"
          :title="$t('sidebar.signup')"
        >
          <i class="material-icons">person_add</i>
          <span>{{ $t("sidebar.signup") }}</span>
        </router-link>
      </template>

      <!-- Bottom Storage Quota Card (Authentic refer2.html student ID style) -->
      <div class="quota-card" v-if="isFiles && !disableUsedPercentage">
        <div class="quota-header">
          <span class="quota-title">
            <i class="material-icons">cloud_queue</i>
            {{ $t("interface.storage") }}
          </span>
          <span class="quota-percent">{{ usage.usedPercentage }}%</span>
        </div>

        <!-- Hard-edged Segmented Progress Bar -->
        <div
          class="quota-track"
          role="progressbar"
          :aria-label="$t('interface.storage')"
          :aria-valuenow="usage.usedPercentage"
          aria-valuemin="0"
          aria-valuemax="100"
        >
          <div
            class="quota-bar"
            :style="{
              width: Math.min(Math.max(usage.usedPercentage, 0), 100) + '%',
            }"
          ></div>
        </div>

        <div class="quota-info">
          <span>{{
            $t("sidebar.diskUsed", { used: usage.used, total: usage.total })
          }}</span>
        </div>
        <div class="quota-barcode-row">
          <div class="css-barcode"></div>
          <span class="quota-id">{{ name }}</span>
        </div>
      </div>

      <p class="credits">
        <span>
          <span v-if="disableExternal">{{ name }}</span>
          <a
            v-else
            rel="noopener noreferrer"
            target="_blank"
            href="https://github.com/Chocola-X/FileNeko"
            >{{ name }}</a
          >
          <span v-if="version"> {{ " " }} {{ version }}</span>
        </span>
        <span>
          <a @click="help">{{ $t("sidebar.help") }}</a>
        </span>
        <project-credits />
      </p>
    </div>
  </nav>
</template>

<script>
import { reactive } from "vue";
import { mapActions, mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import * as auth from "@/utils/auth";
import ProjectCredits from "@/components/ProjectCredits.vue";
import {
  version,
  name,
  signup,
  hideLoginButton,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  logoutPage,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  components: { ProjectCredits },
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    return { usage, usageAbortController: new AbortController() };
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload"]),
    ...mapState(useLayoutStore, ["currentPromptName"]),
    active() {
      return this.currentPromptName === "sidebar";
    },
    signup: () => signup,
    hideLoginButton: () => hideLoginButton,
    version: () => version,
    name: () => name,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    canLogout: () => !noAuth && (loginPage || logoutPage !== "/login"),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers", "showHover"]),
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      const path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";
      let usageStats = USAGE_DEFAULT;
      if (this.disableUsedPercentage) {
        return Object.assign(this.usage, usageStats);
      }
      try {
        this.abortOngoingFetchUsage();
        this.usageAbortController = new AbortController();
        const usage = await api.usage(path, this.usageAbortController.signal);
        usageStats = {
          used: prettyBytes(usage.used, { binary: true }),
          total: prettyBytes(usage.total, { binary: true }),
          usedPercentage:
            usage.total > 0
              ? Math.min(
                  100,
                  Math.max(
                    0,
                    Math.round((usage.used / usage.total) * 1000) / 10
                  )
                )
              : 0,
        };
      } finally {
        return Object.assign(this.usage, usageStats);
      }
    },
    toRoot() {
      this.$router.push({ path: "/files" });
      this.closeHovers();
    },
    toAccountSettings() {
      this.$router.push({ path: "/settings/profile" });
      this.closeHovers();
    },
    toGlobalSettings() {
      this.$router.push({ path: "/settings/global" });
      this.closeHovers();
    },
    help() {
      this.showHover("help");
    },
    logout: auth.logout,
  },
  watch: {
    $route: {
      handler(to) {
        if (to.path.includes("/files")) {
          this.fetchUsage();
        }
      },
      immediate: true,
    },
  },
  unmounted() {
    this.abortOngoingFetchUsage();
  },
};
</script>
