<template>
  <div class="cel" aria-hidden="true">
    <canvas ref="canvas" class="confetti" />
    <template v-if="playing">
      <div
        class="shock"
        :key="'s' + playing.id"
        :style="{ left: cx - r + 'px', top: cy - r + 'px', width: r * 2 + 'px', height: r * 2 + 'px' }"
      />
      <svg
        class="orbit"
        :key="'o' + playing.id"
        :style="{ left: cx - r * 0.7 + 'px', top: cy - r * 0.7 + 'px', width: r * 1.4 + 'px', height: r * 1.4 + 'px' }"
        viewBox="0 0 100 100"
      >
        <ellipse cx="50" cy="50" rx="46" ry="12" fill="none" stroke="var(--gold)" stroke-width="1.6" transform="rotate(-18 50 50)" />
      </svg>
      <svg
        v-if="playing.kind === 'test'"
        class="rocket"
        :key="'r' + playing.id"
        :style="{ left: cx - 8 + 'px', top: cy - 8 + 'px' }"
        width="16" height="16" viewBox="0 0 16 16"
      >
        <path d="M8 1 L14 15 L8 11.5 L2 15 Z" fill="var(--gold)" />
      </svg>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, shallowRef, watch } from "vue";
import confetti from "canvas-confetti";
import { celebration, type CelebrateEvent } from "../store";
import { reducedMotion } from "../format";

// cx, cy: the ring's centre inside this overlay; r: the ring's radius.
const props = withDefaults(defineProps<{ cx: number; cy: number; r: number; compact?: boolean }>(), { compact: false });
const emit = defineEmits<{ play: [e: CelebrateEvent] }>();

const canvas = ref<HTMLCanvasElement | null>(null);
const playing = shallowRef<CelebrateEvent | null>(null);
const played = new Set<string>();
let fire: ((o: confetti.Options) => Promise<null> | null) | null = null;
let timer = 0;

onMounted(() => {
  const c = canvas.value;
  // happy-dom (tests) has no canvas.
  if (!c || typeof c.getContext !== "function" || !c.getContext("2d")) return;
  try {
    // A blob worker is blocked by the Content-Security-Policy.
    fire = confetti.create(c, { resize: true, useWorker: false });
  } catch {
    fire = null;
  }
});

onBeforeUnmount(() => clearTimeout(timer));

watch(celebration, (e) => {
  if (!e || !e.id || played.has(e.id) || e.kind === "quarter") return;
  played.add(e.id);
  emit("play", e);
  if (reducedMotion()) return;
  playing.value = e;
  clearTimeout(timer);
  timer = window.setTimeout(() => (playing.value = null), e.kind === "test" ? 2600 : 1600);
  const delay = e.kind === "test" ? 1600 : 400;
  if (fire && !props.compact) {
    const burst = (x: number, angle: number) =>
      fire!({
        particleCount: 90, spread: 70, startVelocity: 38, ticks: 160, scalar: 0.9,
        colors: ["#f2c66d", "#d9b28a", "#62e6b4", "#e6ecff"], origin: { x, y: 0.45 }, angle,
      });
    window.setTimeout(() => {
      burst(0.25, 60);
      burst(0.75, 120);
    }, delay);
  }
});
</script>

<style scoped>
.cel { position: absolute; inset: 0; pointer-events: none; overflow: visible; z-index: 5; }
.confetti { position: absolute; inset: 0; width: 100%; height: 100%; }
.shock {
  position: absolute; border-radius: 50%; border: 2px solid var(--gold);
  animation: shock 600ms cubic-bezier(0, 0, 0.2, 1) 350ms both;
}
@keyframes shock {
  from { transform: scale(1); opacity: 0.6; }
  to { transform: scale(2.2); opacity: 0; }
}
.orbit { position: absolute; animation: orbit 600ms var(--burst) 500ms both; }
@keyframes orbit {
  0% { transform: rotate(0deg) scale(0.6); opacity: 0; }
  30% { opacity: 1; }
  100% { transform: rotate(360deg) scale(1); opacity: 0; }
}
.rocket { position: absolute; animation: launch 1600ms cubic-bezier(0.4, 0, 0.6, 1) both; }
@keyframes launch {
  0% { transform: translate(0, 0) rotate(0deg); opacity: 1; }
  40% { transform: translate(20px, -60px) rotate(18deg); opacity: 1; }
  100% { transform: translate(110px, -190px) rotate(32deg); opacity: 0; }
}
</style>
