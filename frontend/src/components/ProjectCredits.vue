<template>
  <span v-if="showCredits" class="project-legal">
    <span>© 2026 Chocola-X</span>
    <span class="project-legal-links">
      <a :href="sourceURL" target="_blank" rel="noopener noreferrer">{{
        t("legal.sourceCode")
      }}</a>
      <span aria-hidden="true">·</span>
      <a
        href="https://www.gnu.org/licenses/agpl-3.0.html"
        target="_blank"
        rel="noopener noreferrer"
        >AGPLv3</a
      >
    </span>
  </span>
</template>

<script setup lang="ts">
import { sourceURL, noAuth } from "@/utils/constants";
import { useAuthStore } from "@/stores/auth";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

const { t } = useI18n({});
const authStore = useAuthStore();
const route = useRoute();

const showCredits = computed(
  () =>
    !noAuth &&
    authStore.isLoggedIn &&
    route.matched.some((record) => record.meta.requiresAuth)
);
</script>
