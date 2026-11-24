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
      <Today v-else-if="tab === 'today'" @plan="tab = 'plan'" @settings="tab = 'settings'" />
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
/* The tabs sit right of the traffic lights, like a native toolbar. */
.titlebar { height: 38px; flex: none; display: flex; align-items: stretch; padding-left: 84px; border-bottom: 1px solid var(--line); }
.tabs { display: flex; gap: 18px; --wails-draggable: no-drag; }
.tabs button {
  position: relative; border: 0; background: none; color: var(--sub); font: inherit; padding: 0 2px; cursor: pointer;
}
.tabs button:hover { color: var(--text); }
.tabs button.on { color: var(--text); font-weight: 600; }
.tabs button.on::after {
  content: ""; position: absolute; left: 0; right: 0; bottom: -1px; height: 2px; background: var(--accent); border-radius: 1px;
}
.name { font-weight: 500; align-self: center; }
.body { flex: 1; overflow: auto; }
</style>
