<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row management-page" v-else-if="!layoutStore.loading">
    <div class="column">
      <div class="fn-settings-card card management-card users-manager">
        <div class="card-title">
          <settings-card-title icon="groups">{{
            t("settings.users")
          }}</settings-card-title>
        </div>
        <div class="card-content management-list-content">
          <table class="management-table users-table">
            <thead>
              <tr>
                <th>{{ t("settings.username") }}</th>
                <th>{{ t("settings.admin") }}</th>
                <th>{{ t("settings.scope") }}</th>
                <th :aria-label="t('buttons.update')"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td class="user-name-cell">
                  <span class="user-name"
                    ><i class="material-icons" aria-hidden="true">person</i
                    ><span>{{ user.username }}</span></span
                  >
                </td>
                <td class="user-admin-cell">
                  <span class="management-cell-label">{{
                    t("settings.admin")
                  }}</span>
                  <span
                    class="user-role"
                    :class="{ 'is-admin': user.perm.admin }"
                    role="img"
                    :aria-label="
                      user.perm.admin
                        ? t('settings.administrator')
                        : t('settings.user')
                    "
                  >
                    <i class="material-icons" aria-hidden="true">{{
                      user.perm.admin ? "verified_user" : "remove"
                    }}</i>
                  </span>
                </td>
                <td class="user-scope-cell">
                  <span class="management-cell-label">{{
                    t("settings.scope")
                  }}</span
                  ><code>{{ user.scope }}</code>
                </td>
                <td class="user-actions-cell">
                  <div class="management-actions">
                    <router-link
                      :to="'/settings/users/' + user.id"
                      class="action management-action"
                      data-tone="purple"
                      :aria-label="t('buttons.update')"
                      :title="t('buttons.update')"
                      ><i class="material-icons" aria-hidden="true"
                        >mode_edit</i
                      ></router-link
                    >
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="card-action management-footer">
          <router-link to="/settings/users/new" class="button user-create"
            ><i class="material-icons" aria-hidden="true">person_add</i
            >{{ t("buttons.new") }}</router-link
          >
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SettingsCardTitle from "@/components/settings/SettingsCardTitle.vue";
import { useLayoutStore } from "@/stores/layout";
import { users as api } from "@/api";
import Errors from "@/views/Errors.vue";
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";

const error = ref<StatusError | null>(null);
const users = ref<IUser[]>([]);

const layoutStore = useLayoutStore();
const { t } = useI18n();

onMounted(async () => {
  layoutStore.loading = true;

  try {
    users.value = await api.getAll();
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
});
</script>
