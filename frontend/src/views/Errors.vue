<template>
  <div class="error-page" :class="{ 'error-page--standalone': showHeader }">
    <header-bar v-if="showHeader" showLogo />
    <status-card
      class="error-card"
      data-tone="pink"
      :icon="info.icon"
      :badge="errorCode === 403 ? 'lock' : 'priority_high'"
      :label="String(errorCode)"
    >
      {{ t(info.message) }}
      <template v-if="showHeader" #actions>
        <router-link class="button" to="/files/">
          <i class="material-icons" aria-hidden="true">home</i>
          {{ t("files.home") }}
        </router-link>
      </template>
    </status-card>
  </div>
</template>

<script setup lang="ts">
import HeaderBar from "@/components/header/HeaderBar.vue";
import StatusCard from "@/components/StatusCard.vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n({});

const errors: {
  [key: number]: {
    icon: string;
    message: string;
  };
} = {
  0: {
    icon: "cloud_off",
    message: "errors.connection",
  },
  403: {
    icon: "gpp_bad",
    message: "errors.forbidden",
  },
  404: {
    icon: "gps_off",
    message: "errors.notFound",
  },
  500: {
    icon: "error_outline",
    message: "errors.internal",
  },
};

const props = withDefaults(
  defineProps<{
    errorCode?: number;
    showHeader?: boolean;
  }>(),
  {
    errorCode: 500,
    showHeader: false,
  }
);

const info = computed(() => {
  return errors[props.errorCode] ? errors[props.errorCode] : errors[500];
});
</script>
