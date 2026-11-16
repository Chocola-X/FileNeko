<template>
  <div class="dashboard">
    <header-bar showMenu showLogo />

    <div class="settings-heading">
      <div>
        <span class="settings-eyebrow">FileNeko / SETTINGS</span>
        <h1>{{ t("sidebar.settings") }}<span aria-hidden="true">✦</span></h1>
      </div>
      <span class="settings-pass-tag" aria-hidden="true">{{
        t("interface.settings")
      }}</span>
    </div>

    <div id="nav">
      <div class="wrapper">
        <ul>
          <router-link to="/settings/profile"
            ><li :class="{ active: $route.path === '/settings/profile' }">
              {{ t("settings.profileSettings") }}
            </li></router-link
          >
          <router-link to="/settings/shares" v-if="user?.perm.share"
            ><li :class="{ active: $route.path === '/settings/shares' }">
              {{ t("settings.shareManagement") }}
            </li></router-link
          >
          <router-link to="/settings/global" v-if="user?.perm.admin"
            ><li :class="{ active: $route.path === '/settings/global' }">
              {{ t("settings.globalSettings") }}
            </li></router-link
          >
          <router-link to="/settings/users" v-if="user?.perm.admin"
            ><li
              :class="{
                active:
                  $route.path === '/settings/users' || $route.name === 'User',
              }"
            >
              {{ t("settings.userManagement") }}
            </li></router-link
          >
        </ul>
      </div>
    </div>

    <loading-indicator v-if="loading" />

    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
import LoadingIndicator from "@/components/LoadingIndicator.vue";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const authStore = useAuthStore();
const layoutStore = useLayoutStore();

const user = computed(() => authStore.user);
const loading = computed(() => layoutStore.loading);
</script>
