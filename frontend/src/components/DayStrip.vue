<template>
  <div class="strip" :class="{ large }" role="list" aria-label="The plan">
    <button
      v-for="(d, i) in days"
      :key="d.date"
      class="day"
      :class="[d.isExam ? 'exam' : d.off ? 'off' : d.status, { today: d.isToday, lit: lit > i }]"
      role="listitem"
      :title="title(d)"
      @click="$emit('open')"
    >
      <span class="wd">{{ d.weekday.slice(0, large ? 3 : 1) }}</span>
      <span class="num mono">{{ d.date.slice(8).replace(/^0/, "") }}</span>
      <span v-if="d.isExam" class="bar exam-bar mono">SAT</span>
      <span v-else-if="d.off" class="bar off-bar" />
      <span v-else class="bar"><span class="fill" :style="{ height: Math.min(1, d.fraction) * 100 + '%' }" /></span>
    </button>
  </div>
</template>

<script setup lang="ts">
import type { StripDay } from "../api";

// The days to the exam as columns: a bar fills as the day's work is logged.
withDefaults(defineProps<{ days: StripDay[]; large?: boolean; lit?: number }>(), { large: false, lit: -1 });
defineEmits<{ open: [] }>();

function title(d: StripDay): string {
  if (d.isExam) return `${d.dateLabel}: the SAT`;
  if (d.off) return `${d.dateLabel}: ${d.off}`;
  const what = (d.tasks ?? []).map((t) => (t.isTest ? (t.done >= t.target ? "test done" : "test") : `${t.kind}, ${t.done} of ${t.target}`)).join(", ");
  return `${d.dateLabel}: ${what}`;
}
</script>

<style scoped>
.strip { display: grid; grid-auto-flow: column; grid-auto-columns: 1fr; gap: 2px; }
.day {
  display: flex; flex-direction: column; align-items: center; gap: 3px; padding: 4px 0 5px;
  background: none; border: 0; border-radius: 8px; cursor: pointer; color: var(--sub); font: inherit;
}
.day:hover { background: var(--raised); }
.wd { font-size: 9.5px; text-transform: uppercase; letter-spacing: 0.04em; }
.num { font-size: 11px; color: var(--text); }
.large .wd { font-size: 10.5px; }
.large .num { font-size: 13px; }
.bar {
  position: relative; width: 8px; height: 22px; border-radius: 2px; background: var(--cell); overflow: hidden;
  display: flex; align-items: flex-end;
}
.large .bar { width: 12px; height: 34px; }
.fill { width: 100%; background: var(--accent); transition: height 280ms var(--calm); }
.missed .bar { box-shadow: inset 0 0 0 1.5px var(--alarm); }
.off-bar { background: repeating-linear-gradient(-45deg, var(--line) 0 2px, transparent 2px 4px); }
.exam-bar {
  width: auto; padding: 0 3px; background: var(--accent); color: var(--ink); font-size: 8.5px; font-weight: 700;
  align-items: center; justify-content: center; writing-mode: vertical-rl; letter-spacing: 0.05em;
}
.large .exam-bar { font-size: 10px; }
.day.today .wd, .day.today .num { color: var(--accent); }
.day.today { background: var(--raised); }
.lit { animation: lit 300ms var(--burst); }
@keyframes lit { 50% { transform: translateY(-4px); } }
@media (prefers-reduced-motion: reduce) { .lit { animation: none; } }
</style>
