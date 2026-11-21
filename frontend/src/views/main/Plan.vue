<template>
  <div class="plan" v-if="s" :class="{ embedded }">
    <div class="top">
      <div>
        <h2 v-if="!embedded" class="display">The plan</h2>
        <p class="muted summary">{{ s.planSummary }}</p>
      </div>
      <div class="actions">
        <button class="btn small" @click="run(api.SuggestPlan(false))">Suggest a plan</button>
        <button v-if="!confirmReset" class="link" @click="confirmReset = true">Reset to the suggestion</button>
        <span v-else class="inline">
          <span class="muted">Replace every day from today?</span>
          <button class="btn small" @click="reset">Reset</button>
          <button class="link" @click="confirmReset = false">Cancel</button>
        </span>
      </div>
    </div>

    <div class="cards">
      <button
        v-for="d in s.strip ?? []"
        :key="d.date"
        class="card"
        :class="{ today: d.isToday, past: !d.editable && !d.isExam, exam: d.isExam, open: editing === d.date }"
        :disabled="!d.editable || d.isExam"
        @click="edit(d)"
      >
        <span class="wd">{{ d.weekday }}</span>
        <span class="date tabnum">{{ d.dateLabel.split(" ").slice(1).join(" ") }}</span>
        <MiniRing :fraction="d.fraction" :status="d.status" :off="!!d.off" :exam="d.isExam" :today="d.isToday" :size="22" />
        <span class="kind">{{ short(d) }}</span>
        <span class="muted tgt tabnum">{{ targetText(d) }}</span>
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
          <input v-model.number="targets[k]" class="field tabnum" type="number" min="1" max="500" @change="save" />
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import MiniRing from "../../components/MiniRing.vue";
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
    .map((t) => (t.isTest ? (t.done >= t.target ? "done" : "1 test") : d.editable && !d.isToday ? `${t.target}` : `${t.done}/${t.target}`))
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
.plan { padding: 20px 28px; display: flex; flex-direction: column; gap: 16px; }
.plan.embedded { padding: 0; }
.top { display: flex; justify-content: space-between; align-items: flex-end; gap: 16px; }
h2 { font-size: 26px; margin: 0 0 4px; }
.summary { margin: 0; }
.actions { display: flex; align-items: center; gap: 12px; }
.inline { display: inline-flex; align-items: center; gap: 8px; }
.cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(100px, 1fr)); gap: 8px; }
.card {
  display: flex; flex-direction: column; align-items: flex-start; gap: 4px; padding: 10px 12px;
  border-radius: 12px; border: 1px solid var(--line); background: var(--space); color: var(--ice);
  font: inherit; text-align: left; cursor: pointer; transition: background 140ms var(--calm), border-color 140ms var(--calm);
}
.card:hover:not(:disabled) { background: var(--deep); }
.card:disabled { cursor: default; }
.card.past { opacity: 0.6; }
.card.today { border-color: var(--ice); }
.card.open { border-color: var(--gold); background: var(--deep); }
.card.exam { border-color: var(--gold); }
.wd { color: var(--dust); font-size: 12px; }
.date { font-weight: 600; }
.kind { margin-top: 4px; }
.tgt { font-size: 12px; min-height: 16px; }
.editor { border: 1px solid var(--line); border-radius: 12px; padding: 14px 16px; display: flex; flex-direction: column; gap: 12px; }
.ehead { display: flex; gap: 12px; align-items: baseline; }
.presets { display: flex; flex-wrap: wrap; gap: 6px; }
.targets { display: flex; gap: 20px; }
.targets label { display: flex; align-items: center; gap: 8px; }
.targets .field { width: 76px; }
</style>
