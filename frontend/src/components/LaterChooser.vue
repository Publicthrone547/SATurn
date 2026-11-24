<template>
  <div class="later" role="group" aria-label="Start later">
    <span class="q">Start in</span>
    <button v-for="o in options" :key="o.min" class="btn small" :disabled="o.min > max" @click="$emit('choose', o.min)">
      {{ o.label }}
    </button>
    <button class="link" @click="$emit('cancel')">Back</button>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ max?: number }>(), { max: 600 });
defineEmits<{ choose: [minutes: number]; cancel: [] }>();
const options = [
  { min: 15, label: "15 min" },
  { min: 30, label: "30 min" },
  { min: 60, label: "1 hour" },
];
</script>

<style scoped>
.later { display: flex; align-items: center; gap: 6px; }
.q { color: var(--sub); margin-right: 2px; }
</style>
