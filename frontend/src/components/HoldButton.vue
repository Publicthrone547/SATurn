<template>
  <button
    class="hold btn"
    :class="{ holding }"
    @pointerdown="start"
    @pointerup="cancel"
    @pointerleave="cancel"
    @keydown.space.prevent="start"
    @keyup.space.prevent="cancel"
    @keydown.enter.prevent="start"
    @keyup.enter.prevent="cancel"
  >
    <span class="bar" :style="{ transform: `scaleX(${progress})` }" />
    <span class="label"><slot /></span>
  </button>
</template>

<script setup lang="ts">
import { ref, onBeforeUnmount } from "vue";

const props = withDefaults(defineProps<{ ms?: number }>(), { ms: 1200 });
const emit = defineEmits<{ confirm: [] }>();
const progress = ref(0);
const holding = ref(false);
let raf = 0;
let t0 = 0;

function frame(t: number) {
  progress.value = Math.min(1, (t - t0) / props.ms);
  if (progress.value >= 1) {
    holding.value = false;
    progress.value = 0;
    emit("confirm");
    return;
  }
  raf = requestAnimationFrame(frame);
}

function start() {
  if (holding.value) return;
  holding.value = true;
  t0 = performance.now();
  raf = requestAnimationFrame(frame);
}

function cancel() {
  if (!holding.value) return;
  holding.value = false;
  cancelAnimationFrame(raf);
  progress.value = 0;
}

onBeforeUnmount(() => cancelAnimationFrame(raf));
</script>

<style scoped>
.hold { position: relative; overflow: hidden; }
.bar {
  position: absolute; inset: 0; background: currentColor; opacity: 0.18;
  transform-origin: left center; pointer-events: none;
}
.label { position: relative; }
</style>
