<template>
  <div class="plan" v-if="s" :class="{ embedded }">
    <div class="top">
      <div>
        <h2 v-if="!embedded" class="heading">Plan</h2>
        <p class="muted summary">{{ s.planSummary }}. Click a day to change it.</p>
      </div>
      <div class="actions">
        <button v-if="!confirmReset" class="link" @click="confirmReset = true">Reset to the suggestion</button>
        <span v-else class="inline">
          <span class="muted">Replace every day from today?</span>
          <button class="btn small danger" @click="reset">Reset</button>
          <button class="link" @click="confirmReset = false">Cancel</button>
        </span>
      </div>
    </div>

    <div class="days">
      <button
        v-for="d in s.strip ?? []"
        :key="d.date"
        class="day"
        :class="{ today: d.isToday, past: !d.editable && !d.isExam, exam: d.isExam, open: editing === d.date, off: !!d.off, missed: d.status === 'missed' }"
        :disabled="!d.editable || d.isExam"
        @click="edit(d)"
      >
        <span class="wd">{{ d.isToday ? "Today" : d.weekday }}</span>
        <span class="date mono">{{ d.date.slice(8).replace(/^0/, "") }}</span>
        <span class="kind">{{ short(d) }}</span>
        <span class="tgt mono">{{ targetText(d) }}</span>
        <span class="meter"><span class="fill" :style="{ width: (d.isExam ? 0 : Math.min(1, d.fraction) * 100) + '%' }" /></span>
      </button>
    </div>

    <div v-if="day" class="editor">
      <div class="ehead">
        <strong>{{ day.dateLabel }}</strong>
        <span class="muted">{{ day.isToday ? "Today. Progress on a kept subject stays." : "" }}</span>
      </div>
      <div class="presets">
        <button
          v-for="p in presets"
          :key="p.id"
          class="btn small"
          :class="{ primary: preset === p.id }"
          @click="choose(p.id)"
        >
          {{ p.label }}
        </button>
      </div>
      <div v-if="kinds.length" class="targets">
        <label v-for="k in kinds" :key="k">
          <span class="muted">{{ k === "test" ? "Tests" : k === "math" ? "Math questions" : "English questions" }}</span>
          <input v-model.number="targets[k]" class="field mono" type="number" min="1" max="500" @change="save" />
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { snap, run } from "../../store";
import { api, type StripDay } from "../../api";

withDefaults(defineProps<{ embedded?: boolean }>(), { embedded: false });

const s = computed(() => snap.value);
const presets = [
  { id: "math", label: "Math" },
  { id: "english", label: "English" },
  { id: "test", label: "Practice test" },
  { id: "test+math", label: "Test + math" },
  { id: "test+english", label: "Test + English" },
  { id: "rest", label: "Rest" },
  { id: "sick", label: "Sick" },
];
const editing = ref("");
const preset = ref("");
const targets = reactive<Record<string, number>>({});
const confirmReset = ref(false);
const day = computed(() => s.value?.strip?.find((d) => d.date === editing.value) ?? null);
const kinds = computed(() => preset.value.split("+").filter((k) => ["math", "english", "test"].includes(k)));

function short(d: StripDay): string {
  if (d.isExam) return "SAT";
  const map: Record<string, string> = {
    math: "Math", english: "English", test: "Test", "test+math": "Test + math", "test+english": "Test + English", rest: "Rest", sick: "Sick",
  };
  return map[d.preset] ?? d.preset;
}

function targetText(d: StripDay): string {
  if (d.isExam || d.off) return "";
  return (d.tasks ?? [])
    .map((t) => (t.isTest ? (t.done >= t.target ? "done" : "test") : d.editable && !d.isToday ? `${t.target}` : `${t.done}/${t.target}`))
    .join(" + ");
}

function edit(d: StripDay) {
  if (editing.value === d.date) {
    editing.value = "";
    return;
  }
  editing.value = d.date;
  preset.value = d.preset;
  for (const k of Object.keys(targets)) delete targets[k];
  for (const t of d.tasks ?? []) targets[t.kind] = t.target;
}

function choose(p: string) {
  preset.value = p;
  for (const k of kinds.value) {
    if (!targets[k]) targets[k] = s.value?.settings.targets?.[k] ?? (k === "test" ? 1 : 100);
  }
  save();
}

function save() {
  if (!editing.value) return;
  const t: Record<string, number> = {};
  for (const k of kinds.value) if (targets[k] > 0) t[k] = Math.min(500, Math.round(targets[k]));
  run(api.SetDay(editing.value, preset.value, t));
}

async function reset() {
  await run(api.SuggestPlan(true));
  confirmReset.value = false;
}
</script>

<style scoped>
.plan { padding: 22px 28px; display: flex; flex-direction: column; gap: 18px; }
.plan.embedded { padding: 0; }
.top { display: flex; justify-content: space-between; align-items: flex-end; gap: 16px; }
.heading { margin-bottom: 4px; }
.summary { margin: 0; }
.actions { display: flex; align-items: center; gap: 12px; }
.inline { display: inline-flex; align-items: center; gap: 8px; }
/* The days as columns on one line, the exam last. */
.days { display: grid; grid-template-columns: repeat(auto-fill, minmax(60px, 1fr)); gap: 6px; }
.day {
  display: flex; flex-direction: column; align-items: flex-start; gap: 2px; padding: 10px 10px 12px;
  border-radius: 12px; border: 1px solid transparent; background: var(--raised); color: var(--text);
  font: inherit; text-align: left; cursor: pointer; transition: background 120ms var(--calm), border-color 120ms var(--calm);
}
.day:hover:not(:disabled) { background: var(--hover); }
.day:disabled { cursor: default; }
.day.past { background: none; border-color: var(--line); }
.day.off { background: repeating-linear-gradient(-45deg, var(--raised) 0 6px, var(--bg) 6px 12px); }
.day.today { border-color: var(--accent); }
.day.open { border-color: var(--text); }
.day.exam { background: var(--accent); color: var(--ink); }
.day.exam .wd, .day.exam .tgt { color: var(--ink); }
.wd { color: var(--sub); font-size: 11px; }
.day.today .wd { color: var(--accent); }
.date { font-size: 22px; font-weight: 600; letter-spacing: -0.03em; line-height: 1.15; }
.kind { font-weight: 500; margin-top: 4px; }
.tgt { font-size: 11px; color: var(--sub); min-height: 14px; }
.meter { margin-top: 8px; width: 100%; height: 3px; border-radius: 2px; background: var(--cell); overflow: hidden; }
.exam .meter, .off .meter { visibility: hidden; }
.missed .meter { background: var(--alarm); }
.fill { display: block; height: 100%; background: var(--accent); }
.editor { border: 1px solid var(--line); border-radius: 12px; padding: 14px 16px; display: flex; flex-direction: column; gap: 12px; }
.ehead { display: flex; gap: 12px; align-items: baseline; }
.presets { display: flex; flex-wrap: wrap; gap: 6px; }
.targets { display: flex; gap: 20px; }
.targets label { display: flex; align-items: center; gap: 8px; }
.targets .field { width: 76px; }
</style>
