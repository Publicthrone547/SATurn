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
    </section>

    <section>
      <h3>Log from any app</h3>
      <div class="row">
        <label>Global hotkeys</label>
        <Toggle v-model="f.hotkeys" @update:model-value="save" />
      </div>
      <template v-if="f.hotkeys">
        <div v-for="a in hotkeyActions" :key="a.action" class="row">
          <label>{{ a.action === "popover" ? "Open the panel" : a.action === "undo" ? "Undo" : `Add ${a.label}` }}</label>
          <KeyRecorder :model-value="f.keys[a.action] ?? ''" @update:model-value="(v: string) => setKey(a.action, v)" />
        </div>
        <div class="row">
          <span />
          <p class="muted small note">
            Click a key to record a new combination with ⌃, ⌥ or ⌘. Delete clears it.
            Digits and F-keys work in every keyboard layout.
            <button class="link" @click="resetKeys">Restore the defaults</button>
          </p>
        </div>
        <div v-for="e in s.system.hotkeyErrors ?? []" :key="e" class="row">
          <span />
          <p class="bad small">{{ e }}</p>
        </div>
      </template>
      <div class="row top">
        <label>Links</label>
        <div class="links">
          <div v-for="l in links" :key="l.url" class="lrow">
            <code class="mono">{{ l.url }}</code>
            <span class="muted">{{ l.text }}</span>
            <button class="link" @click="copy(l.url)">{{ copied === l.url ? "Copied" : "Copy" }}</button>
          </div>
          <p class="muted small">
            For the Shortcuts app: make a shortcut with the action Open URLs and paste a link. Shortcuts can then run
            it from the menu bar, Spotlight, Siri or a key. Raycast, Alfred and a Stream Deck can open the same links;
            in Terminal, <code class="mono">open -g saturn://add</code>.
          </p>
          <span><button class="btn small" @click="openShortcuts">Open Shortcuts</button></span>
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
          <span :class="s.system.notifications ? 'ok' : 'muted'">{{ s.system.notifications ? "Allowed" : "Not allowed yet" }}</span>
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
              <button class="link bad" @click="bailOpen = true">Quit anyway</button>
            </template>
            <form v-else class="inline" @submit.prevent="bail">
              <input v-model="phrase" class="field wide" placeholder="Type: I am done for today" aria-label="Bailout phrase" />
              <button class="btn small" type="submit">Quit</button>
              <button class="link" type="button" @click="bailOpen = false">Cancel</button>
            </form>
            <p v-if="bailErr" class="bad small">{{ bailErr }}</p>
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
import { Browser, Clipboard } from "@wailsio/runtime";
import Toggle from "../../components/Toggle.vue";
import KeyRecorder from "../../components/KeyRecorder.vue";
import { hotkeyActions } from "../../hotkeys";
import { snap, run } from "../../store";
import { api, errMsg, useMock, type Settings } from "../../api";
import { clock, minutes } from "../../format";

const s = computed(() => snap.value);
// The generated type marks maps as nullable; the backend always fills them.
type Form = Omit<Settings, "targets" | "quickAdd" | "keys"> & {
  targets: Record<string, number>;
  quickAdd: Record<string, number[]>;
  keys: Record<string, string>;
};
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

function setKey(action: string, acc: string) {
  if (!f.value) return;
  f.value.keys = { ...f.value.keys, [action]: acc };
  save();
}

function resetKeys() {
  if (!f.value) return;
  f.value.keys = { add1: "Cmd+Option+1", add5: "Cmd+Option+5", add10: "Cmd+Option+0", undo: "Cmd+Option+Z", popover: "Cmd+Option+S" };
  save();
}

const links = [
  { url: "saturn://add", text: "+1" },
  { url: "saturn://add?n=5", text: "+5, any number up to 200" },
  { url: "saturn://undo", text: "undo the last entry" },
  { url: "saturn://popover", text: "open the panel" },
];
const copied = ref("");
async function copy(text: string) {
  try {
    if (useMock) await navigator.clipboard.writeText(text);
    else await Clipboard.SetText(text);
    copied.value = text;
    setTimeout(() => (copied.value = ""), 1500);
  } catch {
    copied.value = "";
  }
}
function openShortcuts() {
  if (!useMock) Browser.OpenURL("shortcuts://create-shortcut");
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
.settings { padding: 22px 28px 28px; display: flex; flex-direction: column; gap: 26px; max-width: 700px; }
section { display: flex; flex-direction: column; gap: 10px; }
h3 { margin: 0 0 2px; padding-bottom: 8px; font-size: 13px; font-weight: 600; border-bottom: 1px solid var(--line); }
.row { display: grid; grid-template-columns: 200px 1fr; align-items: center; gap: 12px; }
.row.top { align-items: start; }
.row label { color: var(--sub); }
.row > .field { justify-self: start; min-width: 150px; }
.num { width: 90px; }
.tiny { width: 58px; }
.wide { width: 220px; }
.inline { display: inline-flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.gap { margin-left: 12px; }
.small { font-size: 12px; margin: 4px 0 0; }
.note { margin: 0; line-height: 1.5; max-width: 52ch; }
.links { display: flex; flex-direction: column; gap: 6px; }
.lrow { display: grid; grid-template-columns: 150px 1fr auto; align-items: baseline; gap: 10px; }
.links p { line-height: 1.5; max-width: 52ch; }
code { color: var(--text); font-size: 12px; }
.bail { margin-top: 8px; }
footer { margin-top: 8px; }
</style>
