<template>
  <div v-if="s" class="panel" :class="{ large }" ref="root">
    <div class="top">
      <div class="label">{{ s.day.label }}</div>
      <div v-if="tasks.length > 1" class="switch" role="tablist" aria-label="Task">
        <button
          v-for="(t, i) in tasks"
          :key="t.kind"
          role="tab"
          :aria-selected="sel === i"
          :class="{ on: sel === i }"
          @click="picked = i"
        >
          {{ t.isTest ? "Test" : t.label }}
          <span v-if="t.isTest">{{ t.done >= t.target ? "✓" : "" }}</span>
          <span v-else class="mono">{{ t.done }}</span>
        </button>
      </div>
    </div>

    <div class="count" ref="countBox">
      <template v-if="task">
        <Odometer :value="task.done" :size="large ? 84 : 64" :class="{ full: task.done >= task.target }" />
        <span class="of mono">/{{ task.target }}{{ task.isTest ? " test" : "" }}</span>
      </template>
    </div>

    <Tally
      v-if="task && !task.isTest"
      :done="task.done"
      :target="task.target"
      :size="large ? 'large' : 'medium'"
      :flash="flash"
    />

    <div class="status" :class="{ fade: !!cel }">
      <template v-if="off">
        <div>{{ s.day.off === "sick" ? "Sick day." : "Rest day." }} Reminders are off.</div>
        <div class="muted">{{ s.day.tomorrow }}</div>
      </template>
      <template v-else-if="s.day.done">
        <div class="ok">Done for today.</div>
        <div class="muted">{{ s.day.tomorrow }}</div>
      </template>
      <template v-else>
        <div>{{ s.day.remainingText }}</div>
        <div v-if="session" class="muted mono">Session running, {{ countdown(sessionLeft) }} left</div>
        <div v-else-if="s.promise" class="muted">You said you would start at {{ s.promise.clock }}</div>
        <div v-else-if="s.pnr.red" class="bad">{{ s.pnr.fitText }}</div>
      </template>
    </div>

    <div v-if="cel" class="celtext" :key="cel.id">
      <div class="celtitle">{{ cel.title }}</div>
      <div class="streak">Streak <span class="mono pop">{{ streakShown }}</span></div>
      <div class="muted tomorrow">{{ cel.kind === "final" ? cel.totals : cel.tomorrow }}</div>
      <div v-if="cel.kind === 'final'" class="go">Go get it.</div>
    </div>

    <DayLine
      v-if="pending && !session"
      :from="clock(s.settings.remindFrom)"
      :until="clock(s.settings.remindUntil)"
      :now-frac="s.pnr.nowFrac"
      :pnr-frac="s.pnr.pnrFrac"
      :pnr-clock="s.pnr.clock"
      :red="s.pnr.red"
    />

    <template v-if="!off">
      <div v-if="task && !task.isTest" class="adds">
        <button
          v-for="(n, i) in quick"
          :key="i"
          class="btn"
          :class="{ primary: i === 0 && !s.day.done }"
          :data-add="n"
          :title="keyFor(n) ? `+${n} from any app: ${keyLabel(keyFor(n))}` : `+${n}`"
          @click="add(n)"
        >
          <span class="mono">+{{ n }}</span>
          <kbd v-if="i === 0 && keyFor(n)">{{ keyLabel(keyFor(n)) }}</kbd>
        </button>
      </div>
      <div v-else-if="task" class="testrow">
        <HoldButton v-if="task.done < task.target" class="grow primary" @confirm="markDone">Hold when the test is done</HoldButton>
        <button v-else class="btn grow" disabled>Test done</button>
        <input
          v-model="score"
          class="field score mono"
          inputmode="numeric"
          placeholder="Score"
          aria-label="Practice test score (optional)"
          @change="saveScore"
        />
      </div>

      <button v-if="session" class="btn ghost wide" @click="run(api.EndSession())">
        End session <span class="mono muted">{{ countdown(sessionLeft) }}</span>
      </button>
      <button v-else-if="!s.day.done" class="btn ghost wide" @click="run(api.StartSession(0))">
        {{ task?.isTest ? `Start the test, ${s.settings.testSessionMin} min` : `Start a ${s.settings.sessionMin} min session` }}
      </button>
    </template>

    <div class="links">
      <template v-if="off">
        <button class="link" @click="setOff('')">Back to the plan</button>
      </template>
      <template v-else>
        <button class="link" :disabled="!s.day.canUndo" @click="run(api.Undo())">Undo</button>
        <button v-if="task && !task.isTest" class="link" @click="toggleExact">Set exact</button>
        <button class="link" @click="dayOff = dayOff ? '' : 'menu'">Day off</button>
      </template>
    </div>

    <form v-if="exact" class="inline" @submit.prevent="setExact">
      <input ref="exactInput" v-model="exactValue" class="field grow mono" inputmode="numeric" :placeholder="`Done so far, 0 to ${task?.target}`" />
      <button class="btn small" type="submit">Set</button>
      <button class="link" type="button" @click="exact = false">Cancel</button>
    </form>

    <div v-if="dayOff === 'menu'" class="inline">
      <button class="btn small" @click="dayOff = 'rest'">Rest</button>
      <button class="btn small" @click="dayOff = 'sick'">Sick</button>
      <button class="link" @click="dayOff = ''">Cancel</button>
    </div>
    <div v-else-if="dayOff" class="inline confirm">
      <span class="grow">{{ dayOff === "rest" ? "Rest" : "Sick" }} today. {{ keepText }}</span>
      <button class="btn small" @click="setOff(dayOff)">Confirm</button>
      <button class="link" @click="dayOff = ''">Cancel</button>
    </div>

    <Celebration :cx="countCx" :cy="countCy" :r="large ? 120 : 92" @play="onCelebrate" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import Tally from "./Tally.vue";
import DayLine from "./DayLine.vue";
import Odometer from "./Odometer.vue";
import HoldButton from "./HoldButton.vue";
import Celebration from "./Celebration.vue";
import { snap, run, celebration, type CelebrateEvent } from "../store";
import { api, on } from "../api";
import { clock, countdown, keyLabel, reducedMotion } from "../format";

const props = withDefaults(defineProps<{ large?: boolean; source?: string }>(), { large: false, source: "popover" });
const emit = defineEmits<{ lit: [n: number] }>();

const s = computed(() => snap.value);
const tasks = computed(() => s.value?.day.tasks ?? []);
const picked = ref<number | null>(null);
const sel = computed(() => {
  const n = tasks.value.length;
  if (picked.value !== null && picked.value < n) return picked.value;
  return s.value?.day.active ?? 0;
});
const task = computed(() => tasks.value[sel.value]);
const off = computed(() => !!s.value?.day.off);
const pending = computed(() => s.value?.phase === "study" && !off.value && !s.value?.day.done && tasks.value.length > 0);
const quick = computed(() => task.value?.quickAdd ?? [1, 5, 10, 25]);

// The global hotkey that adds n, when there is one.
function keyFor(n: number): string {
  const st = s.value?.settings;
  if (!st?.hotkeys) return "";
  const action = ({ 1: "add1", 5: "add5", 10: "add10" } as Record<number, string>)[n];
  return action ? (st.keys?.[action] ?? "") : "";
}

watch(
  () => s.value?.day.date,
  () => (picked.value = null),
);

async function add(n: number) {
  await run(api.Add(sel.value, n, props.source));
}
function markDone() {
  run(api.MarkDone(sel.value));
}

// Practice test score (optional).
const score = ref("");
watch(
  () => s.value?.day.score,
  (v) => (score.value = v ? String(v) : ""),
  { immediate: true },
);
function saveScore() {
  const v = score.value.trim() === "" ? 0 : Number(score.value);
  if (Number.isNaN(v)) return;
  run(api.SetScore(s.value!.day.date, v));
}

// Exact value.
const exact = ref(false);
const exactValue = ref("");
const exactInput = ref<HTMLInputElement | null>(null);
async function toggleExact() {
  exact.value = !exact.value;
  exactValue.value = String(task.value?.done ?? "");
  if (exact.value) {
    await nextTick();
    exactInput.value?.focus();
    exactInput.value?.select();
  }
}
async function setExact() {
  const v = Number(exactValue.value);
  if (!Number.isFinite(v) || v < 0) return;
  if (await run(api.SetExact(sel.value, Math.round(v)))) exact.value = false;
}

// Day off.
const dayOff = ref("");
const keepText = computed(() => {
  const left = (s.value?.strip ?? []).filter((d) => !d.isExam && !d.off && d.date > (s.value?.today ?? "")).length;
  return `${left} study ${left === 1 ? "day stays" : "days stay"} in the plan.`;
});
async function setOff(v: string) {
  if (await run(api.SetOff(s.value!.today, v))) dayOff.value = "";
}

// Session countdown, ticking locally from the last snapshot.
const session = computed(() => s.value?.session ?? null);
const sessionLeft = ref(0);
let base = 0;
let baseAt = 0;
watch(
  session,
  (v) => {
    base = v?.remainingSec ?? 0;
    baseAt = Date.now();
    sessionLeft.value = base;
  },
  { immediate: true },
);
const iv = window.setInterval(() => {
  if (session.value) sessionLeft.value = Math.max(0, base - (Date.now() - baseAt) / 1000);
}, 1000);
onBeforeUnmount(() => clearInterval(iv));

// The popover's "Other" button from the island focuses the exact field.
let offFocus: (() => void) | null = null;
onMounted(() => {
  offFocus = on<string>("popover:focus", (f) => {
    if (f === "exact" && !exact.value) toggleExact();
  });
});
onBeforeUnmount(() => offFocus?.());

// Celebration: the quarter flash and the text.
const flash = ref<{ quarter: number; id: string } | null>(null);
const cel = ref<CelebrateEvent | null>(null);
const streakShown = ref(0);
let celTimer = 0;
watch(celebration, (e) => {
  if (e?.kind === "quarter" && e.quarter) flash.value = { quarter: e.quarter, id: e.id };
});
function onCelebrate(e: CelebrateEvent) {
  cel.value = e;
  const to = e.streak ?? 0;
  streakShown.value = Math.max(0, to - 1);
  window.setTimeout(() => (streakShown.value = to), reducedMotion() ? 0 : 1100);
  if (e.kind === "final") {
    const n = s.value?.strip?.length ?? 0;
    for (let i = 0; i <= n; i++) window.setTimeout(() => emit("lit", i), 1600 + i * 40);
  }
  clearTimeout(celTimer);
  celTimer = window.setTimeout(() => (cel.value = null), e.kind === "final" ? 9000 : 5500);
}
onBeforeUnmount(() => clearTimeout(celTimer));

// The count's centre inside the panel, for the celebration overlay.
const root = ref<HTMLElement | null>(null);
const countBox = ref<HTMLElement | null>(null);
const countCx = ref(180);
const countCy = ref(80);
function measure() {
  const a = root.value?.getBoundingClientRect();
  const b = countBox.value?.getBoundingClientRect();
  if (!a || !b) return;
  countCx.value = b.left - a.left + b.width / 2;
  countCy.value = b.top - a.top + b.height / 2;
}
onMounted(() => nextTick(measure));
watch(() => tasks.value.length, () => nextTick(measure));
</script>

<style scoped>
.panel { position: relative; display: flex; flex-direction: column; gap: 12px; }
.top { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-height: 24px; }
.label { font-weight: 600; font-size: 14px; }
.switch { display: inline-flex; background: var(--raised); border-radius: 8px; padding: 2px; }
.switch button {
  border: 0; background: none; color: var(--sub); font: inherit; font-size: 12px; padding: 3px 10px; border-radius: 6px; cursor: pointer;
  display: inline-flex; gap: 6px;
}
.switch button.on { background: var(--line); color: var(--text); }

.count { display: flex; align-items: baseline; gap: 6px; min-height: 64px; }
.large .count { min-height: 84px; }
.count :deep(.odo) { color: var(--text); font-weight: 600; letter-spacing: -0.04em; }
.count :deep(.odo.full) { color: var(--accent); }
.of { font-size: 20px; color: var(--sub); }
.large .of { font-size: 26px; }

.status { display: flex; flex-direction: column; gap: 2px; transition: opacity 200ms var(--calm); }
.status.fade { opacity: 0; }

.celtext {
  position: absolute; left: 0; right: 0; z-index: 6; pointer-events: none;
  top: var(--celtop); background: var(--bg); padding-bottom: 6px;
}
.panel { --celtop: 132px; }
.panel.large { --celtop: 168px; }
.celtitle { font-size: 18px; font-weight: 650; color: var(--accent); animation: rise 280ms var(--calm) 900ms both; }
.celtext .streak { animation: rise 280ms var(--calm) 1100ms both; }
.celtext .pop { display: inline-block; animation: pop 320ms var(--burst) 1100ms both; }
.celtext .tomorrow { animation: rise 280ms var(--calm) 1400ms both; }
.celtext .go { font-size: 16px; font-weight: 650; color: var(--accent); animation: rise 280ms var(--calm) 2400ms both; }
@keyframes rise { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: none; } }
@keyframes pop { 0% { transform: scale(1); } 50% { transform: scale(1.25); } 100% { transform: scale(1); } }

.adds { display: grid; grid-template-columns: 1.6fr 1fr 1fr 1fr; gap: 6px; }
.adds .btn { height: 40px; font-size: 15px; }
.adds .btn.primary kbd { background: rgba(18, 19, 20, 0.12); border-color: rgba(18, 19, 20, 0.2); color: var(--ink); }
.testrow { display: flex; gap: 8px; }
.testrow :deep(.hold.primary) { background: var(--accent); color: var(--ink); font-weight: 600; }
.grow { flex: 1; }
.score { width: 84px; height: 34px; }
.wide { width: 100%; }
.links { display: flex; gap: 14px; padding: 0 2px; }
.inline { display: flex; align-items: center; gap: 8px; }
.confirm span { color: var(--text); }
@media (prefers-reduced-motion: reduce) {
  .celtext * { animation: none !important; }
}
</style>
