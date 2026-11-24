<template>
  <button
    ref="el"
    class="rec"
    :class="{ on: recording, empty: !modelValue }"
    :aria-label="recording ? 'Press the new keys' : `Hotkey ${keyLabel(modelValue) || 'none'}, click to change`"
    @click="start"
    @keydown="onKey"
    @blur="stop"
  >
    <template v-if="recording">{{ partial || "Press keys" }}</template>
    <template v-else-if="modelValue">{{ keyLabel(modelValue) }}</template>
    <template v-else>None</template>
  </button>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref } from "vue";
import { api } from "../api";
import { accelerator, keyLabel } from "../format";

// Records a global hotkey. Esc cancels, Delete clears. The app's hotkeys are
// released while recording, so the current combination can be pressed too.
defineProps<{ modelValue: string }>();
const emit = defineEmits<{ "update:modelValue": [v: string] }>();

const el = ref<HTMLButtonElement | null>(null);
const recording = ref(false);
const partial = ref("");

function start() {
  if (recording.value) return;
  recording.value = true;
  partial.value = "";
  api.PauseHotkeys(true);
}

function stop() {
  if (!recording.value) return;
  recording.value = false;
  api.PauseHotkeys(false);
}

function onKey(e: KeyboardEvent) {
  if (!recording.value) {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      start();
    }
    return;
  }
  e.preventDefault();
  e.stopPropagation();
  const bare = !e.ctrlKey && !e.altKey && !e.metaKey && !e.shiftKey;
  if (bare && e.key === "Escape") return stop();
  if (bare && (e.key === "Backspace" || e.key === "Delete")) {
    emit("update:modelValue", "");
    return stop();
  }
  const acc = accelerator(e);
  if (!acc) {
    // Show the modifiers held so far.
    partial.value = [e.ctrlKey && "⌃", e.altKey && "⌥", e.shiftKey && "⇧", e.metaKey && "⌘"].filter(Boolean).join("");
    return;
  }
  emit("update:modelValue", acc);
  stop();
  el.value?.blur();
}

onBeforeUnmount(stop);
</script>

<style scoped>
.rec {
  justify-self: start; min-width: 92px; height: 28px; padding: 0 10px; border-radius: 8px; border: 1px solid var(--line);
  background: var(--bg); color: var(--text); font-family: var(--font-mono); font-size: 13px; cursor: pointer;
}
.rec:hover { border-color: var(--sub); }
.rec.empty { color: var(--sub); font-family: inherit; font-size: 12px; }
.rec.on { border-color: var(--accent); color: var(--accent); font-family: inherit; font-size: 12px; }
</style>
