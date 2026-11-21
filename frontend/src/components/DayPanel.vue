<template>
  <div v-if="s" class="panel" :class="{ large }" ref="root">
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
        <span v-else class="tabnum">{{ t.done }}</span>
      </button>
    </div>

    <div class="ringwrap" ref="ringBox">
      <Ring
        :fraction="task ? task.fraction : 0"
        :done="task?.done ?? 0"
        :target="task?.target ?? 0"
        :is-test="!!task?.isTest"
        :size="ringSize"
        :number-size="large ? 76 : 56"
        :show-track="pending"
        :now-frac="s.pnr.nowFrac"
        :pnr-frac="s.pnr.pnrFrac"
        :state="ringState"
        :flash="flash"
        :lock="lockKey"
      />
    </div>

    <div class="captions" :class="{ fade: !!cel }">
      <div class="label">{{ off ? s.day.label : s.day.label }}</div>
      <template v-if="off">
        <div class="muted">Reminders are off today.</div>
        <div class="muted">{{ s.day.tomorrow }}</div>
      </template>
      <template v-else-if="s.day.done">
        <div class="mint">Done for today</div>
        <div class="muted">{{ s.day.tomorrow }}</div>
      </template>
      <template v-else>
        <div class="muted">{{ s.day.remainingText }}</div>
        <div v-if="session" class="muted tabnum">Session running · {{ countdown(sessionLeft) }} left</div>
        <div v-else-if="s.promise" class="muted">You said you would start at {{ s.promise.clock }}</div>
        <div v-else-if="s.pnr.red" class="coral">{{ s.pnr.fitText }}</div>
        <div v-else class="muted">Start by {{ s.pnr.clock }} to finish today</div>
      </template>
    </div>

    <div v-if="cel" class="celtext" :key="cel.id">
      <div class="display title">{{ cel.title }}</div>
      <div class="streak">Streak <span class="tabnum pop">{{ streakShown }}</span></div>
      <div class="muted tomorrow">{{ cel.kind === "final" ? cel.totals : cel.tomorrow }}</div>
      <div v-if="cel.kind === 'final'" class="display go">Go get it.</div>
    </div>

    <template v-if="!off">
      <div v-if="task && !task.isTest" class="adds">
        <button v-for="(n, i) in quick" :key="i" class="btn" @click="add(n)">+{{ n }}</button>
      </div>
      <div v-else-if="task" class="testrow">
        <HoldButton v-if="task.done < task.target" class="grow" @confirm="markDone">Hold for test done</HoldButton>
        <button v-else class="btn grow" disabled>Test done</button>
        <input
          v-model="score"
          class="field score"
          inputmode="numeric"
          placeholder="Score"
          aria-label="Practice test score (optional)"
          @change="saveScore"
        />
      </div>

      <button v-if="session" class="btn wide" @click="run(api.EndSession())">
        End session <span class="tabnum muted">{{ countdown(sessionLeft) }}</span>
      </button>
      <button v-else-if="!s.day.done" class="btn primary wide" @click="run(api.StartSession(0))">
        {{ task?.isTest ? `Start the test (${s.settings.testSessionMin} min)` : `Start a ${s.settings.sessionMin} minute session` }}
      </button>
    </template>

    <div class="links">
      <template v-if="off">
        <button class="link" @click="setOff('')">Back to the plan</button>
      </template>
      <template v-else>
        <button class="link" :disabled="!s.day.canUndo" @click="run(api.Undo())">Undo</button>
        <button v-if="task && !task.isTest" class="link" @click="toggleExact">Set exact</button>
        <button class="link" @click="dayOff = dayOff ? '' : 'menu'">Day off ▾</button>
      </template>
    </div>

    <form v-if="exact" class="inline" @submit.prevent="setExact">
      <input ref="exactInput" v-model="exactValue" class="field grow" inputmode="numeric" :placeholder="`Done so far, 0 to ${task?.target}`" />
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

    <Celebration :cx="ringCx" :cy="ringCy" :r="ringSize * 0.43" @play="onCelebrate" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import Ring from "./Ring.vue";
import HoldButton from "./HoldButton.vue";
import Celebration from "./Celebration.vue";
import { snap, run, celebration, type CelebrateEvent } from "../store";
import { api, on } from "../api";
import { countdown, reducedMotion } from "../format";

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
const ringSize = computed(() => (props.large ? 260 : 200));
const ringState = computed(() => (off.value ? "off" : s.value?.day.done ? "done" : s.value?.pnr.red ? "red" : "pending"));
const quick = computed(() => task.value?.quickAdd ?? [1, 5, 10, 25]);

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

// Celebration: the quarter flash, the ring lock and the text.
const flash = ref<{ quarter: number; id: string } | null>(null);
const lockKey = ref("");
const cel = ref<CelebrateEvent | null>(null);
const streakShown = ref(0);
let celTimer = 0;
watch(celebration, (e) => {
  if (e?.kind === "quarter" && e.quarter) flash.value = { quarter: e.quarter, id: e.id };
});
function onCelebrate(e: CelebrateEvent) {
  lockKey.value = e.id;
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

// The ring's centre inside the panel, for the celebration overlay.
const root = ref<HTMLElement | null>(null);
const ringBox = ref<HTMLElement | null>(null);
const ringCx = ref(180);
const ringCy = ref(140);
function measure() {
  const a = root.value?.getBoundingClientRect();
  const b = ringBox.value?.getBoundingClientRect();
  if (!a || !b) return;
  ringCx.value = b.left - a.left + b.width / 2;
  ringCy.value = b.top - a.top + b.height / 2;
}
onMounted(() => nextTick(measure));
watch(() => tasks.value.length, () => nextTick(measure));
</script>

<style scoped>
.panel { position: relative; display: flex; flex-direction: column; align-items: stretch; gap: 10px; }
.switch {
  align-self: center; display: inline-flex; background: var(--deep); border-radius: 9px; padding: 2px;
}
.switch button {
  border: 0; background: none; color: var(--dust); font: inherit; padding: 4px 12px; border-radius: 7px; cursor: pointer;
  display: inline-flex; gap: 6px;
}
.switch button.on { background: var(--line); color: var(--ice); }
.ringwrap { align-self: center; }
.captions { text-align: center; display: flex; flex-direction: column; gap: 2px; min-height: 54px; transition: opacity 200ms var(--calm); }
.captions.fade { opacity: 0; }
.label { font-weight: 600; }
.celtext {
  position: absolute; left: 0; right: 0; text-align: center; z-index: 6;
  top: calc(var(--ring) + 6px); pointer-events: none;
}
.panel { --ring: 210px; }
.panel.large { --ring: 270px; }
.celtext .title { font-size: 24px; color: var(--gold); animation: rise 280ms var(--calm) 900ms both; }
.celtext .streak { animation: rise 280ms var(--calm) 1100ms both; }
.celtext .pop { display: inline-block; animation: pop 320ms var(--burst) 1100ms both; }
.celtext .tomorrow { animation: rise 280ms var(--calm) 1400ms both; }
.celtext .go { font-size: 20px; color: var(--gold); animation: rise 280ms var(--calm) 2400ms both; }
@keyframes rise { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: none; } }
@keyframes pop { 0% { transform: scale(1); } 50% { transform: scale(1.2); } 100% { transform: scale(1); } }
.adds { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; }
.testrow { display: flex; gap: 8px; }
.grow { flex: 1; }
.score { width: 84px; height: 36px; }
.wide { width: 100%; }
.links { display: flex; justify-content: space-between; padding: 0 2px; }
.inline { display: flex; align-items: center; gap: 8px; }
.confirm span { color: var(--ice); }
@media (prefers-reduced-motion: reduce) {
  .celtext * { animation: none !important; }
}
</style>
