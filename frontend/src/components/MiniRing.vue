<template>
  <svg :width="size" :height="size" viewBox="0 0 16 16" class="mini" :class="{ today }" aria-hidden="true">
    <circle v-if="today" cx="8" cy="8" r="7.6" fill="none" stroke="var(--ice)" stroke-width="0.8" />
    <template v-if="exam">
      <path d="M8 3.2 L12.4 12 L3.6 12 Z" fill="var(--gold)" />
    </template>
    <template v-else-if="off">
      <path :d="crescent" fill="var(--dust)" />
    </template>
    <template v-else>
      <circle cx="8" cy="8" r="5" fill="none" :stroke="status === 'missed' ? 'var(--coral)' : 'var(--line)'" stroke-width="2.5" />
      <circle
        v-if="fraction > 0"
        cx="8" cy="8" r="5" fill="none"
        :stroke="fraction >= 1 ? 'var(--mint)' : 'var(--gold)'"
        stroke-width="2.5"
        pathLength="1"
        stroke-dasharray="1"
        :stroke-dashoffset="1 - Math.min(1, fraction)"
        transform="rotate(-90 8 8)"
      />
    </template>
  </svg>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{ fraction?: number; status?: string; off?: boolean; exam?: boolean; today?: boolean; size?: number }>(),
  { fraction: 0, status: "upcoming", off: false, exam: false, today: false, size: 16 },
);
const crescent = "M10.5 3.2 A5.2 5.2 0 1 0 10.5 12.8 A4 4 0 1 1 10.5 3.2 Z";
</script>

<style scoped>
.mini { display: block; flex: none; }
</style>
