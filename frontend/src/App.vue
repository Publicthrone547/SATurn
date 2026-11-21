<template>
  <Popover v-if="route === 'popover'" />
  <Island v-else-if="route === 'island'" />
  <Main v-else />
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import Popover from "./views/Popover.vue";
import Island from "./views/Island.vue";
import Main from "./views/Main.vue";
import { initStore } from "./store";

// Routing by URL hash: #/popover, #/island, anything else is the main window.
const route = ref(location.hash.replace(/^#\/?/, "") || "main");
window.addEventListener("hashchange", () => (route.value = location.hash.replace(/^#\/?/, "") || "main"));
if (route.value === "island") document.body.style.background = "#000";
onMounted(initStore);
</script>
