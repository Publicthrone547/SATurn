<template>
  <div class="odo" :style="{ fontSize: size + 'px' }" :aria-label="String(value)" role="img">
    <span v-for="(d, i) in digits" :key="digits.length - i" class="col">
      <span class="strip" :style="{ transform: `translateY(${-d}em)` }">
        <span v-for="n in 10" :key="n">{{ n - 1 }}</span>
      </span>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(defineProps<{ value: number; size?: number }>(), { size: 56 });
const digits = computed(() => String(Math.max(0, Math.round(props.value))).split("").map(Number));
</script>

<style scoped>
.odo {
  display: flex; font-family: var(--font-display); color: var(--gold);
  line-height: 1; height: 1em; overflow: hidden; font-variant-numeric: tabular-nums;
}
.col { display: inline-block; width: calc(0.44em + 2px); height: 1em; overflow: hidden; text-align: center; }
.strip { display: flex; flex-direction: column; transition: transform 280ms var(--calm); }
.strip span { height: 1em; }
@media (prefers-reduced-motion: reduce) { .strip { transition: none; } }
</style>
