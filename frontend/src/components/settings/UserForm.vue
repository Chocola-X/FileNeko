<template>
  <div>
    <p v-if="!isDefault && props.user !== null">
      <label for="username">{{ t("settings.username") }}</label>
      <input
        class="input input--block"
        type="text"
        v-model="user.username"
        id="username"
      />
    </p>

    <p v-if="!isDefault">
      <label for="password">{{ t("settings.password") }}</label>
      <input
        class="input input--block"
        type="password"
        :placeholder="passwordPlaceholder"
        v-model="user.password"
        id="password"
      />
    </p>

    <p>
      <label for="scope">{{ t("settings.scope") }}</label>
      <input
        :disabled="automaticHome"
        :placeholder="scopePlaceholder"
        class="input input--block"
        type="text"
        v-model="user.scope"
        id="scope"
      />
    </p>
    <p class="small" v-if="isNew">
      <label>
        <input type="checkbox" v-model="createUserDirData" />
        {{ t("settings.createUserHomeDirectory") }}
      </label>
    </p>

    <p>
      <label for="locale">{{ t("settings.language") }}</label>
      <languages
        class="input input--block"
        id="locale"
        v-model:locale="user.locale"
      ></languages>
    </p>

    <p v-if="!isDefault && user.perm">
      <input
        type="checkbox"
        :disabled="user.perm.admin"
        v-model="user.lockPassword"
      />
      {{ t("settings.lockPassword") }}
    </p>

    <permissions v-model:perm="user.perm" />
    <commands
      v-if="enableExec && user.perm?.admin"
      v-model:commands="user.commands"
    />

    <div v-if="!isDefault">
      <h3>{{ t("settings.rules") }}</h3>
      <p class="small">{{ t("settings.rulesHelp") }}</p>
      <rules v-model:rules="user.rules" />
    </div>
  </div>
</template>

<script setup lang="ts">
import Languages from "./Languages.vue";
import Rules from "./Rules.vue";
import Permissions from "./Permissions.vue";
import Commands from "./Commands.vue";
import { enableExec } from "@/utils/constants";
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const createUserDirData = defineModel<boolean>("createUserDir", {
  default: false,
});
const manualScope = ref("");

const props = defineProps<{
  user: IUserForm;
  isNew: boolean;
  isDefault: boolean;
  homeBasePath?: string;
}>();

const passwordPlaceholder = computed(() =>
  props.isNew ? "" : t("settings.avoidChanges")
);
const automaticHome = computed(() => props.isNew && createUserDirData.value);
const scopePlaceholder = computed(() => {
  if (!automaticHome.value) return "";
  const name = (props.user.username || "")
    .trim()
    .replaceAll("..", "")
    .replace(/[^0-9A-Za-z@_.-]/g, "-")
    .replace(/-+/g, "-");
  const base = (props.homeBasePath || "/users").replace(/\/+$/, "");
  return `${base}/${name || t("settings.username")}`;
});

watch(
  () => props.user,
  () => {
    if (!props.user?.perm?.admin) return;
    props.user.lockPassword = false;
  }
);

watch(automaticHome, (automatic) => {
  if (automatic) {
    manualScope.value = props.user.scope || "";
    props.user.scope = "";
  } else {
    props.user.scope = manualScope.value;
  }
});
</script>
