<template>
  <div class="settings" v-if="s && f">
    <section>
      <h3>The exam</h3>
      <div class="row">
        <label>Exam date</label>
        <input v-model="f.examDate" class="field" type="date" @change="save" />
      </div>
      <div class="row">
        <label>Target score</label>
        <input
          :value="f.targetScore || ''"
          class="field num"
          type="number"
          min="400"
          max="1600"
          step="10"
          placeholder="none"
          @change="setScore"
        />
      </div>
    </section>

    <section>
      <h3>Reminders</h3>
      <div class="row">
        <label>Morning brief</label>
        <input :value="clock(f.briefAt)" class="field" type="time" @change="setTime('briefAt', $event)" />
      </div>
      <div class="row">
        <label>Remind from</label>
        <input :value="clock(f.remindFrom)" class="field" type="time" @change="setTime('remindFrom', $event)" />
      </div>
      <div class="row">
        <label>Remind until</label>
        <input :value="clock(f.remindUntil)" class="field" type="time" @change="setTime('remindUntil', $event)" />
      </div>
      <div class="row">
        <label>Every</label>
        <select v-model.number="f.intervalMin" class="field" @change="save">
          <option v-for="n in [30, 45, 60, 90, 120]" :key="n" :value="n">{{ n }} minutes</option>
        </select>
      </div>
      <div class="row">
        <label>The day ends at</label>
        <select v-model.number="f.rolloverHour" class="field" @change="save">
          <option v-for="h in [0, 1, 2, 3, 4, 5]" :key="h" :value="h">{{ clock(h * 60) }}</option>
        </select>
      </div>
      <div class="row">
        <label>Send a test reminder</label>
        <span class="inline">
          <select v-model.number="testLevel" class="field">
            <option :value="1">Level 1: banner</option>
            <option :value="2">Level 2: stays, with sound</option>
            <option :value="3">Level 3: urgent</option>
          </select>
          <button class="btn small" @click="api.TestReminder(testLevel)">Send</button>
        </span>
      </div>
    </section>

    <section>
      <h3>Targets and logging</h3>
      <div class="row">
        <label>Math questions a day</label>
        <input v-model.number="f.targets.math" class="field num" type="number" min="1" max="500" @change="save" />
      </div>
      <div class="row">
        <label>English questions a day</label>
        <input v-model.number="f.targets.english" class="field num" type="number" min="1" max="500" @change="save" />
      </div>
      <div v-for="k in ['math', 'english']" :key="k" class="row">
        <label>{{ k === "math" ? "Math" : "English" }} quick add</label>
        <span class="inline">
          <input
            v-for="i in 4"
            :key="i"
            v-model.number="f.quickAdd[k][i - 1]"
            class="field tiny"
            type="number"
            min="1"
            max="200"
            :aria-label="`${k} quick add ${i}`"
            @change="save"
          />
          <button class="link" :title="k === 'math' ? 'A math module has 22 questions' : 'A Reading and Writing module has 27 questions'" @click="module(k)">
            {{ k === "math" ? "Use 22" : "Use 27" }}
          </button>
        </span>
      </div>
      <div class="row">
        <label>Session length</label>
        <select v-model.number="f.sessionMin" class="field" @change="save">
          <option v-for="n in [25, 50, 90]" :key="n" :value="n">{{ n }} minutes</option>
        </select>
      </div>
    </section>

    <section>
      <h3>Menu bar and sound</h3>
      <div class="row">
        <label>Menu bar text</label>
        <select v-model="f.trayText" class="field" @change="save">
          <option value="progress">Progress, 37/100</option>
          <option value="countdown">Countdown, T-13</option>
          <option value="none">Icon only</option>
        </select>
      </div>
      <div class="row">
        <label>Sound</label>
        <Toggle v-model="f.sound" @update:model-value="save" />
      </div>
      <div class="row top">
        <label>Hotkeys</label>
        <div>
          <Toggle v-model="f.hotkeys" @update:model-value="save" />
          <ul class="keys muted">
            <li><kbd>⌃⌥1</kbd> +1 · <kbd>⌃⌥5</kbd> +5 · <kbd>⌃⌥0</kbd> +10</li>
            <li><kbd>⌃⌥Z</kbd> undo · <kbd>⌃⌥S</kbd> open or close the popover</li>
          </ul>
          <p v-for="e in s.system.hotkeyErrors ?? []" :key="e" class="coral small">{{ e }}</p>
        </div>
      </div>
    </section>

    <section>
      <h3>System</h3>
      <div class="row top">
        <label>Launch at login</label>
        <div>
          <Toggle :model-value="s.system.autostart" @update:model-value="(v: boolean) => run(api.SetLaunchAtLogin(v))" />
          <p class="muted small">{{ s.system.autostartNote }}</p>
        </div>
      </div>
      <div class="row top">
        <label>Notifications</label>
        <div>
          <span :class="s.system.notifications ? 'mint' : 'muted'">{{ s.system.notifications ? "Allowed" : "Not allowed yet" }}</span>
          <span class="inline gap">
            <button class="btn small" @click="allow">Allow notifications</button>
            <button class="link" @click="api.OpenNotificationSettings()">Open notification settings</button>
          </span>
          <p class="muted small">Choose the Alerts style so reminders stay on screen.</p>
        </div>
      </div>
      <div class="row top">
        <label>Strict mode</label>
        <div>
          <Toggle v-model="f.strict" @update:model-value="save" />
          <p class="muted small">While today is not finished, SATurn cannot be quit from its menu.</p>
          <div v-if="!s.canQuit" class="bail">
            <template v-if="!bailOpen">
              <button class="link coral" @click="bailOpen = true">Quit anyway</button>
            </template>
            <form v-else class="inline" @submit.prevent="bail">
              <input v-model="phrase" class="field wide" placeholder="Type: I am done for today" aria-label="Bailout phrase" />
              <button class="btn small" type="submit">Quit</button>
              <button class="link" type="button" @click="bailOpen = false">Cancel</button>
            </form>
            <p v-if="bailErr" class="coral small">{{ bailErr }}</p>
          </div>
        </div>
      </div>
    </section>

    <footer class="muted small">
      SATurn {{ s.system.version }}{{ s.system.dry ? " · dry run" : "" }} · data in {{ s.system.dataDir }}
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Toggle from "../../components/Toggle.vue";
import { snap, run } from "../../store";
import { api, errMsg, type Settings } from "../../api";
import { clock, minutes } from "../../format";

const s = computed(() => snap.value);
// The generated type marks maps as nullable; the backend always fills them.
type Form = Omit<Settings, "targets" | "quickAdd"> & { targets: Record<string, number>; quickAdd: Record<string, number[]> };
const f = ref<Form | null>(null);
watch(
  () => s.value?.settings,
  (v) => {
    if (v) f.value = JSON.parse(JSON.stringify(v));
  },
  { immediate: true },
);

const testLevel = ref(2);
const bailOpen = ref(false);
const phrase = ref("");
const bailErr = ref("");

function save() {
  if (!f.value) return;
  const copy = JSON.parse(JSON.stringify(f.value)) as Settings;
  if (!copy.targetScore) copy.targetScore = 0;
  run(api.UpdateSettings(copy));
}

function setScore(e: Event) {
  if (!f.value) return;
  const v = Number((e.target as HTMLInputElement).value);
  f.value.targetScore = v >= 400 && v <= 1600 ? Math.round(v) : 0;
  save();
}

function setTime(key: "briefAt" | "remindFrom" | "remindUntil", e: Event) {
  const m = minutes((e.target as HTMLInputElement).value);
  if (!f.value || Number.isNaN(m)) return;
  f.value[key] = m;
  save();
}

function module(k: string) {
  const q = f.value?.quickAdd[k];
  if (!q) return;
  q[3] = k === "math" ? 22 : 27;
  save();
}

async function allow() {
  await api.RequestNotifications();
}

async function bail() {
  bailErr.value = "";
  try {
    await api.Bailout(phrase.value);
  } catch (e) {
    bailErr.value = errMsg(e);
  }
}
</script>

<style scoped>
.settings { padding: 20px 28px 28px; display: flex; flex-direction: column; gap: 22px; max-width: 680px; }
section { display: flex; flex-direction: column; gap: 10px; }
h3 { margin: 0 0 2px; font-size: 13px; font-weight: 600; color: var(--gold); }
.row { display: grid; grid-template-columns: 200px 1fr; align-items: center; gap: 12px; }
.row.top { align-items: start; }
.row label { color: var(--dust); }
.num { width: 90px; }
.tiny { width: 58px; }
.wide { width: 220px; }
.inline { display: inline-flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.gap { margin-left: 12px; }
.small { font-size: 12px; margin: 4px 0 0; }
.keys { list-style: none; padding: 0; margin: 8px 0 0; font-size: 12px; display: flex; flex-direction: column; gap: 2px; }
kbd { font-family: inherit; color: var(--ice); }
.bail { margin-top: 8px; }
footer { margin-top: 8px; }
</style>
