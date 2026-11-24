<template>
  <div class="dayline" :class="{ red }" :title="hint">
    <div class="above mono">
      <span v-if="!red" class="deadline" :style="labelStyle">start by {{ pnrClock }}</span>
    </div>
    <div class="track">
      <span class="past" :style="{ width: pct(now) }" />
      <span class="work" :style="{ left: pct(start), width: pct(1 - start) }" />
      <span class="mark" :style="{ left: pct(pnr) }" />
      <span class="now" :style="{ left: pct(now) }" />
    </div>
    <div class="ends mono">
      <span>{{ from }}</span>
      <span>{{ until }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

// The reminder hours as a line: the dot is now, the hatched part is the work
// still to do, placed at the latest time it can start and still end in time.
const props = defineProps<{ from: string; until: string; nowFrac: number; pnrFrac: number; pnrClock: string; red: boolean }>();

const now = computed(() => clamp(props.nowFrac));
const pnr = computed(() => clamp(props.pnrFrac));
const start = computed(() => (props.red ? now.value : pnr.value));
const hint = computed(() =>
  props.red
    ? "The rest of today's work no longer fits before the reminders end"
    : `The hatched part is the work left, placed so it ends at ${props.until}. Start it by ${props.pnrClock}.`,
);
// The deadline label sits over its mark but never runs past the ends.
const labelStyle = computed(() => ({ left: `clamp(50px, ${pnr.value * 100}%, calc(100% - 50px))` }));

function clamp(v: number) {
  return Math.min(1, Math.max(0, v || 0));
}
function pct(v: number) {
  return `${(v * 100).toFixed(2)}%`;
}
</script>

<style scoped>
.dayline { display: flex; flex-direction: column; gap: 5px; }
.above { position: relative; height: 14px; font-size: 10.5px; }
.track { position: relative; height: 8px; border-radius: 2px; background: var(--raised); overflow: visible; }
.past { position: absolute; left: 0; top: 0; bottom: 0; background: var(--line); border-radius: 2px 0 0 2px; }
.work {
  position: absolute; top: 0; bottom: 0; border-radius: 2px;
  background: repeating-linear-gradient(-45deg, var(--sub) 0 2px, transparent 2px 5px);
  opacity: 0.55; transition: left 280ms var(--calm), width 280ms var(--calm);
}
.red .work { background: repeating-linear-gradient(-45deg, var(--alarm) 0 2px, transparent 2px 5px); opacity: 0.9; }
.mark { position: absolute; top: -4px; bottom: -4px; width: 2px; margin-left: -1px; background: var(--text); border-radius: 1px; transition: left 280ms var(--calm); }
.red .mark { display: none; }
.now {
  position: absolute; top: 50%; width: 10px; height: 10px; margin: -5px 0 0 -5px; border-radius: 50%;
  background: var(--accent); box-shadow: 0 0 0 3px var(--bg);
}
.red .now { background: var(--alarm); }
.ends { display: flex; justify-content: space-between; font-size: 10.5px; color: var(--sub); }
.deadline { position: absolute; bottom: 0; transform: translateX(-50%); color: var(--text); white-space: nowrap; }
</style>
