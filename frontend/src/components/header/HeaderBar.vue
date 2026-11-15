<template>
  <header>
    <div
      v-if="showLogo"
      class="header-brand-wrap"
      @click="toRoot"
      role="button"
      tabindex="0"
    >
      <img :src="logoURL" alt="File Browser" class="header-logo-img" />
      <div class="header-brand-text">
        <div class="header-brand-main">
          <span class="header-brand-title">{{ name }}</span>
          <span class="skew-badge header-brand-pill">
            <span>{{ t("interface.fileManager") }}</span>
          </span>
        </div>
        <div class="header-brand-sub">
          <span>{{ t("interface.libraryTitle") }}</span>
          <span class="sub-dot">●</span>
          <span class="sub-ver">{{ version }}</span>
        </div>
      </div>
    </div>
    <Action
      v-if="showMenu"
      class="menu-button"
      icon="menu"
      :label="t('buttons.toggleSidebar')"
      @action="layoutStore.showHover('sidebar')"
    />

    <slot />

    <div
      id="dropdown"
      :class="{ active: layoutStore.currentPromptName === 'more' }"
    >
      <slot name="actions" />
    </div>

    <Action
      v-if="ifActionsSlot"
      id="more"
      icon="more_vert"
      :label="t('buttons.more')"
      @action="layoutStore.showHover('more')"
    />

    <div
      class="overlay"
      v-show="layoutStore.currentPromptName == 'more'"
      @click="layoutStore.closeHovers"
    />
  </header>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";

import { logoURL, version, name } from "@/utils/constants";

import Action from "@/components/header/Action.vue";
import { computed, useSlots } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";

defineProps<{
  showLogo?: boolean;
  showMenu?: boolean;
}>();

const layoutStore = useLayoutStore();
const slots = useSlots();
const router = useRouter();

const { t } = useI18n();

const ifActionsSlot = computed(() => (slots.actions ? true : false));

const toRoot = () => {
  router.push({ path: "/files/" });
};
</script>

<style></style>
