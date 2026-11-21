<template>
  <div class="history" v-if="s">
    <div class="stats">
      <div><span class="display num tabnum">{{ s.streak }}</span><span class="muted">day streak</span></div>
      <div><span class="display num tabnum">{{ s.total }}</span><span class="muted">questions logged</span></div>
      <div><span class="display num tabnum">{{ s.testsDone }}</span><span class="muted">{{ s.testsDone === 1 ? "practice test" : "practice tests" }}</span></div>
    </div>
    <p v-if="!rows.length" class="muted">Past days show up here.</p>
    <table v-else>
      <tbody>
        <template v-for="d in rows" :key="d.date">
          <tr>
            <td class="tabnum">{{ d.dateLabel }}</td>
            <td>{{ d.label }}</td>
            <td class="tabnum">{{ d.doneText }}</td>
            <td :class="statusClass(d.status)">{{ statusWord(d.status) }}</td>
          </tr>
          <tr v-for="(c, i) in d.changes ?? []" :key="d.date + i" class="change">
            <td></td>
            <td colspan="3" class="muted"><span class="tabnum">{{ c.time }}</span> {{ c.text }}</td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { snap } from "../../store";

const s = computed(() => snap.value);
const rows = computed(() => (s.value?.strip ?? []).filter((d) => !d.isExam && d.date <= (s.value?.today ?? "")).reverse());

function statusWord(st: string): string {
  return { done: "Done", missed: "Missed", pending: "Today", rest: "Rest", sick: "Sick" }[st] ?? st;
}
function statusClass(st: string): string {
  return st === "done" ? "mint" : st === "missed" ? "coral" : "muted";
}
</script>

<style scoped>
.history { padding: 20px 28px; display: flex; flex-direction: column; gap: 20px; }
.stats { display: flex; gap: 40px; }
.stats div { display: flex; align-items: baseline; gap: 8px; }
.num { font-size: 40px; color: var(--gold); }
table { border-collapse: collapse; width: 100%; max-width: 680px; }
td { padding: 8px 12px 8px 0; border-bottom: 1px solid var(--line); }
tr.change td { border-bottom: 0; padding-top: 0; font-size: 12px; }
</style>
