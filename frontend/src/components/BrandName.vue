<template>
  <span class="brand-wordmark fn-wordmark brand-wordmark--default"
    >{{ head }}<span class="brand-wordmark-tail">{{ tail }}</span></span
  >
</template>

<script setup lang="ts">
import { name } from "@/utils/constants";

// Keep spaces intact and split at the first word boundary when available.
const characters = Array.from(name);
const space = characters.findIndex((character) => /\s/u.test(character));
const midpoint = space >= 0 ? space : Math.ceil(characters.length / 2);
const head = characters.slice(0, midpoint).join("");
const tail = characters.slice(midpoint).join("");
</script>
