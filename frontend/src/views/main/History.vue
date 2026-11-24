<template>
  <div class="history" v-if="s">
    <div class="stats">
      <div><span class="num mono">{{ s.streak }}</span><span class="muted">day streak</span></div>
      <div><span class="num mono">{{ s.total }}</span><span class="muted">questions logged</span></div>
      <div><span class="num mono">{{ s.testsDone }}</span><span class="muted">{{ s.testsDone === 1 ? "practice test" : "practice tests" }}</span></div>
    </div>
    <p v-if="!rows.length" class="muted">Finished days show up here.</p>
    <table v-else>
      <tbody>
        <template v-for="d in rows" :key="d.date">
          <tr>
            <td class="mono muted">{{ d.dateLabel }}</td>
            <td>{{ d.label }}</td>
            <td class="mono">{{ d.doneText }}</td>
            <td><span class="status" :class="d.status">{{ statusWord(d.status) }}</span></td>
          </tr>
          <tr v-for="(c, i) in d.changes ?? []" :key="d.date + i" class="change">
            <td></td>
            <td colspan="3" class="muted"><span class="mono">{{ c.time }}</span> {{ c.text }}</td>
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
</script>

<style scoped>
.history { padding: 22px 28px; display: flex; flex-direction: column; gap: 24px; }
.stats { display: flex; gap: 12px; }
.stats div {
  flex: 1; max-width: 200px; display: flex; flex-direction: column; gap: 2px;
  padding: 14px 16px; border-radius: 12px; background: var(--raised);
}
.num { font-size: 34px; font-weight: 600; letter-spacing: -0.04em; line-height: 1.1; }
table { border-collapse: collapse; width: 100%; max-width: 680px; }
td { padding: 9px 12px 9px 0; border-bottom: 1px solid var(--line); }
tr.change td { border-bottom: 0; padding-top: 0; font-size: 12px; }
.status { font-size: 12px; padding: 2px 8px; border-radius: 6px; background: var(--raised); color: var(--sub); }
.status.done { background: var(--accent); color: var(--ink); font-weight: 600; }
.status.missed { color: var(--alarm); box-shadow: inset 0 0 0 1px var(--alarm); background: none; }
</style>
