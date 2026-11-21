<template>
  <div class="main" v-if="s">
    <div class="titlebar drag">
      <nav v-if="s.phase === 'study'" class="tabs" role="tablist">
        <button
          v-for="t in tabs"
          :key="t.id"
          role="tab"
          :aria-selected="tab === t.id"
          :class="{ on: tab === t.id }"
          @click="tab = t.id"
        >
          {{ t.label }}
        </button>
      </nav>
      <span v-else class="muted name">SATurn</span>
    </div>
    <main class="body">
      <Welcome v-if="s.phase === 'setup'" />
      <Exam v-else-if="s.phase === 'exam' && tab !== 'settings'" @settings="tab = 'settings'" />
      <Farewell v-else-if="s.phase === 'after' && tab !== 'settings'" />
      <Today v-else-if="tab === 'today'" @plan="tab = 'plan'" />
      <Plan v-else-if="tab === 'plan'" />
      <History v-else-if="tab === 'history'" />
      <Settings v-else />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import Welcome from "./main/Welcome.vue";
import Today from "./main/Today.vue";
import Plan from "./main/Plan.vue";
import History from "./main/History.vue";
import Settings from "./main/Settings.vue";
import Exam from "./main/Exam.vue";
import Farewell from "./main/Farewell.vue";
import { snap } from "../store";
import { on } from "../api";

const s = computed(() => snap.value);
const tabs = [
  { id: "today", label: "Today" },
  { id: "plan", label: "Plan" },
  { id: "history", label: "History" },
  { id: "settings", label: "Settings" },
];
const tab = ref(new URLSearchParams(location.search).get("tab") ?? "today");
let off: (() => void) | null = null;
onMounted(() => {
  off = on<string>("navigate", (t) => {
    if (tabs.some((x) => x.id === t)) tab.value = t;
  });
});
onBeforeUnmount(() => off?.());
</script>

<style scoped>
.main { height: 100%; display: flex; flex-direction: column; }
.titlebar { height: 38px; flex: none; display: flex; align-items: center; justify-content: center; border-bottom: 1px solid var(--line); }
.tabs { display: flex; gap: 2px; --wails-draggable: no-drag; }
.tabs button {
  border: 0; background: none; color: var(--dust); font: inherit; padding: 5px 14px; border-radius: 8px; cursor: pointer;
}
.tabs button:hover { color: var(--ice); }
.tabs button.on { background: var(--deep); color: var(--ice); }
.name { font-weight: 500; }
.body { flex: 1; overflow: auto; }
</style>
