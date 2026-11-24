<template>
  <div
    class="tally"
    :class="[size, { done: done >= target }]"
    :style="{ gridTemplateColumns: `repeat(${t.cols}, 1fr)` }"
    role="img"
    :aria-label="`${done} of ${target}`"
  >
    <span
      v-for="i in t.cells"
      :key="i"
      class="cell"
      :class="{ on: i <= lit, fresh: i > from && i <= lit, flash: flashRow === Math.ceil(i / t.cols) - 1 }"
      :style="i > from && i <= lit ? { animationDelay: Math.min(i - from - 1, 24) * 14 + 'ms' } : undefined"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { tally, reducedMotion } from "../format";

// One cell per question (or per few questions above 100), four rows: each
// row is a quarter of the day. New cells light up one after another.
const props = withDefaults(
  defineProps<{ done: number; target: number; size?: "large" | "medium" | "tiny"; flash?: { quarter: number; id: string } | null }>(),
  { size: "medium", flash: null },
);

const t = computed(() => tally(props.target));
const lit = computed(() => Math.min(t.value.cells, Math.floor(Math.max(0, props.done) / t.value.per)));

// The first cell of the latest entry; cells after it play the light-up.
const from = ref(lit.value);
watch(lit, (now, before) => {
  from.value = reducedMotion() || now < before ? now : before;
});

const flashRow = ref(-1);
watch(
  () => props.flash?.id,
  () => {
    if (!props.flash || reducedMotion()) return;
    flashRow.value = props.flash.quarter - 1;
    setTimeout(() => (flashRow.value = -1), 600);
  },
);
</script>

<style scoped>
.tally { display: grid; gap: 3px; }
.cell { height: 12px; border-radius: 2px; background: var(--cell); }
.large .cell { height: 16px; }
.tally.large { gap: 4px; }
.tiny { gap: 1px; }
.tiny .cell { height: 5px; border-radius: 1px; background: #2c2c2c; }
.cell.on { background: var(--accent); }
.cell.fresh { animation: light 420ms var(--burst) both; }
@keyframes light {
  0% { background: var(--cell); transform: scaleY(0.4); }
  45% { background: #f4ffd6; transform: scaleY(1.25); }
  100% { background: var(--accent); transform: none; }
}
.cell.flash { animation: row 560ms var(--calm); }
@keyframes row { 40% { background: #f4ffd6; } }
@media (prefers-reduced-motion: reduce) {
  .cell.fresh, .cell.flash { animation: none; }
}
</style>
