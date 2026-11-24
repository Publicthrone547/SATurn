<template>
  <div class="exam" v-if="s">
    <h1 class="heading">Exam day.</h1>
    <p v-if="s.settings.targetScore" class="score mono">{{ s.settings.targetScore }}</p>
    <p class="muted">You did the work. Go.</p>
    <button class="link close" @click="close">Close</button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { snap } from "../../store";
import { Window } from "@wailsio/runtime";
import { useMock } from "../../api";

defineEmits<{ settings: [] }>();
const s = computed(() => snap.value);
function close() {
  if (!useMock) Window.Hide();
}
</script>

<style scoped>
.exam { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; }
h1 { font-size: 44px; }
.score { font-size: 96px; font-weight: 600; letter-spacing: -0.05em; color: var(--accent); margin: 0; line-height: 1; }
.close { margin-top: 24px; }
</style>
