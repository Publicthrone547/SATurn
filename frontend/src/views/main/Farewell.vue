<template>
  <div class="bye" v-if="s">
    <h1 class="display">That was the plan.</h1>
    <DayStrip :days="s.strip ?? []" :size="22" class="strip" />
    <p class="totals">{{ s.totals }}</p>
    <p class="muted">SATurn was made for one date. It can leave now.</p>
    <button class="btn primary" @click="api.RetireAndQuit()">Turn off launch at login and quit</button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import DayStrip from "../../components/DayStrip.vue";
import { snap } from "../../store";
import { api } from "../../api";

const s = computed(() => snap.value);
</script>

<style scoped>
.bye { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 14px; text-align: center; }
h1 { font-size: 48px; color: var(--gold); margin: 0; }
.strip { width: 420px; }
.totals { font-size: 16px; margin: 0; }
</style>
