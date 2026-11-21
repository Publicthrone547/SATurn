<template>
  <div class="strip" role="list" aria-label="The plan">
    <button
      v-for="(d, i) in days"
      :key="d.date"
      class="day"
      :class="{ lit: lit > i }"
      role="listitem"
      :title="title(d)"
      @click="$emit('open')"
    >
      <MiniRing :fraction="d.fraction" :status="d.status" :off="!!d.off" :exam="d.isExam" :today="d.isToday" :size="size" />
    </button>
  </div>
</template>

<script setup lang="ts">
import MiniRing from "./MiniRing.vue";
import type { StripDay } from "../api";

const props = withDefaults(defineProps<{ days: StripDay[]; size?: number; lit?: number }>(), { size: 16, lit: -1 });
defineEmits<{ open: [] }>();

function title(d: StripDay): string {
  if (d.isExam) return `${d.dateLabel}: the SAT`;
  if (d.off) return `${d.dateLabel}: ${d.off}`;
  const what = (d.tasks ?? []).map((t) => (t.isTest ? (t.done >= t.target ? "test done" : "test") : `${t.kind}, ${t.done} of ${t.target}`)).join(", ");
  return `${d.dateLabel}: ${what}`;
}
void props;
</script>

<style scoped>
.strip { display: flex; align-items: center; justify-content: space-between; gap: 2px; }
.day { background: none; border: 0; padding: 3px; border-radius: 6px; cursor: pointer; }
.day:hover { background: var(--deep); }
.lit { animation: lit 300ms var(--burst); }
@keyframes lit { 50% { transform: scale(1.35); } }
</style>
