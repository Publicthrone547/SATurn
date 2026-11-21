<template>
  <div class="welcome" v-if="s && f">
    <div class="steps muted tabnum">Step {{ step }} of 4</div>

    <section v-if="step === 1" class="step">
      <h1 class="display">When is your SAT?</h1>
      <input v-model="f.examDate" class="field date" type="date" @change="save" />
      <p class="muted">{{ daysText }}</p>
    </section>

    <section v-else-if="step === 2" class="step">
      <h1 class="display">When may SATurn remind you?</h1>
      <div class="grid">
        <label>Morning brief</label>
        <input :value="clock(f.briefAt)" class="field" type="time" @change="setTime('briefAt', $event)" />
        <label>Remind from</label>
        <input :value="clock(f.remindFrom)" class="field" type="time" @change="setTime('remindFrom', $event)" />
        <label>Remind until</label>
        <input :value="clock(f.remindUntil)" class="field" type="time" @change="setTime('remindUntil', $event)" />
      </div>
      <p class="muted">Every hour in between, until the day's work is logged. Never outside these hours.</p>
    </section>

    <section v-else-if="step === 3" class="step wide">
      <h1 class="display">Your plan.</h1>
      <p class="muted">A suggestion counted back from the exam: tests spaced out, rest before the day. Click a day to change it.</p>
      <Plan embedded />
    </section>

    <section v-else class="step">
      <h1 class="display">Last things.</h1>
      <div class="grid">
        <label>Notifications</label>
        <div>
          <span class="inline">
            <button class="btn small" @click="allow">Allow notifications</button>
            <span :class="s.system.notifications ? 'mint' : 'muted'">{{ s.system.notifications ? "Allowed" : notifNote }}</span>
          </span>
          <p class="muted small">
            In System Settings choose the Alerts style for SATurn, so reminders stay on screen.
            <button class="link" @click="api.OpenNotificationSettings()">Open notification settings</button>
          </p>
        </div>
        <label>Launch at login</label>
        <div>
          <Toggle v-model="launch" />
          <p class="muted small">{{ s.system.installed ? "SATurn opens by itself every morning." : s.system.autostartNote }}</p>
        </div>
        <label>Hotkeys</label>
        <p class="muted small keys">⌃⌥1 adds one question from any app. ⌃⌥5 adds five, ⌃⌥0 adds ten, ⌃⌥Z undoes.</p>
        <label>Your target score</label>
        <input v-model.number="score" class="field num" type="number" min="400" max="1600" step="10" placeholder="1500" />
      </div>
      <p class="muted">It appears on exam morning.</p>
    </section>

    <nav>
      <button v-if="step > 1" class="btn" @click="step--">Back</button>
      <span class="grow" />
      <button v-if="step < 4" class="btn primary" @click="step++">Next</button>
      <button v-else class="btn primary" @click="start">Start</button>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Plan from "./Plan.vue";
import Toggle from "../../components/Toggle.vue";
import { snap, run } from "../../store";
import { api, type Settings } from "../../api";
import { clock, minutes } from "../../format";

const s = computed(() => snap.value);
const step = ref(1);
const f = ref<Settings | null>(null);
const launch = ref(true);
const score = ref<number | "">("");
const notifNote = ref("Not allowed yet");

watch(
  () => s.value?.settings,
  (v) => {
    if (v) f.value = JSON.parse(JSON.stringify(v));
  },
  { immediate: true },
);

const daysText = computed(() => {
  if (!f.value || !s.value) return "";
  const a = new Date(s.value.today + "T12:00:00");
  const b = new Date(f.value.examDate + "T12:00:00");
  const n = Math.round((b.getTime() - a.getTime()) / 86400000);
  return n > 0 ? `${n} ${n === 1 ? "day" : "days"} from today` : "That date has passed";
});

function save() {
  if (f.value) run(api.UpdateSettings(JSON.parse(JSON.stringify(f.value))));
}

function setTime(key: "briefAt" | "remindFrom" | "remindUntil", e: Event) {
  const m = minutes((e.target as HTMLInputElement).value);
  if (!f.value || Number.isNaN(m)) return;
  f.value[key] = m;
  save();
}

async function allow() {
  const ok = await api.RequestNotifications();
  if (!ok) notifNote.value = s.value?.system.installed ? "Not allowed. You can allow it in System Settings." : "Available in the copy in Applications";
}

async function start() {
  await run(api.SetLaunchAtLogin(launch.value));
  const n = typeof score.value === "number" ? score.value : 0;
  await run(api.CompleteOnboarding(n >= 400 && n <= 1600 ? n : 0));
}
</script>

<style scoped>
.welcome { min-height: 100%; box-sizing: border-box; padding: 28px 40px; display: flex; flex-direction: column; gap: 20px; }
.steps { font-size: 12px; }
.step { display: flex; flex-direction: column; gap: 14px; max-width: 560px; }
.step.wide { max-width: none; }
h1 { font-size: 40px; color: var(--gold); margin: 0; }
.date { width: 180px; height: 36px; font-size: 15px; }
.grid { display: grid; grid-template-columns: 160px 1fr; gap: 14px 16px; align-items: center; }
.grid label { color: var(--dust); }
.num { width: 100px; }
.small { font-size: 12px; margin: 6px 0 0; }
.keys { margin: 0; }
.inline { display: inline-flex; align-items: center; gap: 10px; }
nav { margin-top: auto; display: flex; gap: 10px; }
.grow { flex: 1; }
</style>
