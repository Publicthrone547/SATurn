<template>
  <div class="dial" :class="[state, { locking }]" :style="{ width: size + 'px', height: size + 'px' }">
    <svg viewBox="0 0 200 200" :width="size" :height="size" aria-hidden="true">
      <g v-if="showTrack" class="track">
        <circle cx="100" cy="100" r="95" fill="none" stroke="var(--line)" stroke-width="2" />
        <line
          class="pnr"
          x1="100" y1="0.5" x2="100" y2="9.5"
          stroke-width="3" stroke-linecap="round"
          :style="{ transform: `rotate(${pnrFrac * 360}deg)` }"
        />
        <circle class="now" :cx="nowPos[0]" :cy="nowPos[1]" r="3.5" />
      </g>
      <g v-for="(q, i) in qs" :key="i">
        <path :d="paths[i]" fill="none" stroke="var(--line)" :stroke-width="stroke" stroke-linecap="round" />
        <path
          :d="paths[i]"
          class="fill"
          :class="{ flash: flashing === i }"
          fill="none"
          :stroke-width="stroke"
          stroke-linecap="round"
          pathLength="1"
          stroke-dasharray="1"
          :stroke-dashoffset="1 - q"
          :style="{ opacity: q > 0 ? 1 : 0 }"
        />
      </g>
      <g v-if="sparks.length" class="sparks">
        <circle
          v-for="sp in sparks"
          :key="sp.id"
          :cx="sp.x" :cy="sp.y" r="2.2"
          :style="{ '--dx': sp.dx + 'px', '--dy': sp.dy + 'px' }"
        />
      </g>
    </svg>
    <div class="centre">
      <slot>
        <template v-if="isTest">
          <div class="word" :style="{ fontSize: numberSize * 0.6 + 'px' }">Test</div>
          <div class="sub">{{ fraction >= 1 ? "done" : "to do" }}</div>
        </template>
        <template v-else>
          <svg v-if="fraction >= 1" class="check" viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
            <path d="M3 8.5l3.2 3.2L13 5" fill="none" stroke="var(--mint)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <Odometer :value="done" :size="numberSize" />
          <div class="sub">of {{ target }}</div>
        </template>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Odometer from "./Odometer.vue";
import { arc, polar, quarters, reducedMotion } from "../format";

const props = withDefaults(
  defineProps<{
    fraction: number;
    done?: number;
    target?: number;
    size?: number;
    numberSize?: number;
    nowFrac?: number;
    pnrFrac?: number;
    showTrack?: boolean;
    state?: string; // pending | done | off | red
    isTest?: boolean;
    flash?: { quarter: number; id: string } | null;
    lock?: string; // a new value plays the lock pop
  }>(),
  { done: 0, target: 100, size: 200, numberSize: 56, nowFrac: 0, pnrFrac: 1, showTrack: false, state: "pending", isTest: false, flash: null, lock: "" },
);

const stroke = 12;
const qs = computed(() => quarters(props.fraction));
const paths = [0, 1, 2, 3].map((i) => arc(100, 100, 80, i * 90 + 1.5, (i + 1) * 90 - 1.5));
const nowPos = computed(() => polar(100, 100, 95, props.nowFrac * 360));

const flashing = ref(-1);
const sparks = ref<{ id: string; x: number; y: number; dx: number; dy: number }[]>([]);
watch(
  () => props.flash?.id,
  () => {
    const f = props.flash;
    if (!f || reducedMotion()) return;
    const i = f.quarter - 1;
    flashing.value = i;
    setTimeout(() => (flashing.value = -1), 420);
    const end = (i + 1) * 90 - 1.5;
    const [x, y] = polar(100, 100, 80, end);
    sparks.value = Array.from({ length: 12 }, (_, k) => {
      const a = end + (k - 5.5) * 14;
      const [px, py] = polar(0, 0, 14 + (k % 3) * 6, a);
      return { id: `${f.id}-${k}`, x, y, dx: px, dy: py };
    });
    setTimeout(() => (sparks.value = []), 700);
  },
);

const locking = ref(false);
watch(
  () => props.lock,
  (v) => {
    if (!v || reducedMotion()) return;
    locking.value = false;
    requestAnimationFrame(() => {
      locking.value = true;
      setTimeout(() => (locking.value = false), 1000);
    });
  },
);
</script>

<style scoped>
.dial { position: relative; flex: none; }
svg { display: block; overflow: visible; }
.fill { stroke: var(--gold); transition: stroke-dashoffset 280ms var(--calm), opacity 80ms linear; }
.fill.flash { animation: flash 400ms var(--calm); }
@keyframes flash { 50% { stroke: #fff3d0; } }
.pnr {
  stroke: var(--coral);
  transform-box: view-box;
  transform-origin: 100px 100px;
  transition: transform 280ms var(--calm);
}
.now { fill: var(--ice); }
.red .now { fill: var(--coral); }
.centre {
  position: absolute; inset: 0; display: flex; flex-direction: column;
  align-items: center; justify-content: center; pointer-events: none;
}
.sub { color: var(--dust); margin-top: 2px; }
.word { font-family: var(--font-display); color: var(--gold); line-height: 1; }
.check { margin-bottom: 2px; }
.locking svg, .locking .centre { animation: lock 320ms var(--burst) 350ms both; }
@keyframes lock { 0% { transform: scale(1); } 60% { transform: scale(1.08); } 100% { transform: scale(1); } }
.sparks circle { fill: var(--gold); animation: spark 600ms var(--calm) forwards; }
@keyframes spark {
  from { transform: translate(0, 0); opacity: 1; }
  to { transform: translate(var(--dx), var(--dy)); opacity: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .fill, .pnr { transition: none; }
}
</style>
