// An in-memory fake of AppService for working on the interface in a browser
// (npm run dev:mock). It starts from a snapshot the real core produced (a
// math day at 37 of 100) and supports the everyday commands.
// Query flags: ?phase=setup|exam|after, ?task=test, ?red=1, ?session=1, ?done=1,
// ?island=tally.
import base from "./mock-snapshot.json";
import type { Snapshot } from "../bindings/github.com/Publicthrone547/reminder_sat/internal/core/models.js";
import type { IslandPayload } from "../bindings/github.com/Publicthrone547/reminder_sat/models.js";

type Fn = (d: unknown) => void;
const listeners = new Map<string, Set<Fn>>();

export const bus = {
  on(name: string, cb: Fn): () => void {
    if (!listeners.has(name)) listeners.set(name, new Set());
    listeners.get(name)!.add(cb);
    return () => listeners.get(name)?.delete(cb);
  },
  emit(name: string, data: unknown) {
    listeners.get(name)?.forEach((cb) => cb(data));
  },
};

const snap: Snapshot = JSON.parse(JSON.stringify(base));
const log: { task: number; delta: number }[] = [];
const params = typeof location !== "undefined" ? new URLSearchParams(location.search) : new URLSearchParams();

function setup() {
  const phase = params.get("phase");
  if (phase) snap.phase = phase;
  if (params.get("task") === "test") {
    snap.day.tasks = [{ kind: "test", label: "Practice test", target: 1, done: 0, fraction: 0, quickAdd: [0, 0, 0, 0], isTest: true }];
    snap.day.label = "Practice test day";
    snap.day.preset = "test";
  }
  if (params.get("red") === "1") {
    snap.pnr.red = true;
    snap.pnr.fitText = "50 of 100 still fit before 23:00";
    snap.pnr.nowFrac = 0.93;
  }
  if (params.get("session") === "1") {
    snap.session = { endsAt: "", remainingSec: 24 * 60 + 5, totalSec: 50 * 60 };
  }
  if (params.get("done") === "1") {
    const t = snap.day.tasks![0];
    t.done = t.target;
  }
  refresh();
}

function refresh() {
  const tasks = snap.day.tasks ?? [];
  let active = tasks.findIndex((t) => t.done < t.target);
  if (active < 0) active = Math.max(0, tasks.length - 1);
  snap.day.active = active;
  for (const t of tasks) t.fraction = t.target ? t.done / t.target : 0;
  snap.day.done = tasks.length > 0 && tasks.every((t) => t.done >= t.target);
  const t = tasks[active];
  if (t) {
    snap.day.remaining = Math.max(0, t.target - t.done);
    const mins = Math.ceil(snap.day.remaining * (t.kind === "english" ? 1.2 : t.isTest ? 150 : 1.6));
    snap.day.remainingText = snap.day.done
      ? "Done for today"
      : t.isTest
        ? "One full practice test, about 2 h 30 min"
        : `${snap.day.remaining} left, about ${fmt(mins)}`;
    snap.tray.text = snap.day.done ? "✓" : t.isTest ? "test" : `${t.done}/${t.target}`;
    snap.tray.progress = Math.min(1, t.fraction);
    const pnr = 1380 - mins;
    snap.pnr.clock = `${String(Math.floor(pnr / 60)).padStart(2, "0")}:${String(pnr % 60).padStart(2, "0")}`;
    snap.pnr.pnrFrac = Math.min(1, Math.max(0, (pnr - 540) / 840));
  }
  snap.day.canUndo = log.length > 0;
  const today = snap.strip?.find((d) => d.isToday);
  if (today) today.fraction = t ? Math.min(1, t.fraction) : 0;
  bus.emit("state", clone());
}

function fmt(m: number) {
  const h = Math.floor(m / 60);
  const r = m % 60;
  return h === 0 ? `${r} min` : r === 0 ? `${h} h` : `${h} h ${r} min`;
}

function clone(): Snapshot {
  return JSON.parse(JSON.stringify(snap));
}

function done(): Promise<Snapshot> {
  refresh();
  return Promise.resolve(clone());
}

let celebrated = false;
let seq = 0;

export function Snapshot() {
  return Promise.resolve(clone());
}

export function Add(task: number, delta: number, _source: string) {
  const tasks = snap.day.tasks ?? [];
  const i = task < 0 ? snap.day.active : task;
  const t = tasks[i];
  if (!t) return done();
  const before = t.done;
  t.done = Math.max(0, t.done + delta);
  log.push({ task: i, delta: t.done - before });
  const allDone = tasks.every((x) => x.done >= x.target);
  if (allDone && !celebrated) {
    celebrated = true;
    setTimeout(() => bus.emit("celebrate", { kind: "day", id: `mock-${Date.now()}`, title: "Day 1 complete", streak: 1, tomorrow: "Tomorrow: English, 100 questions." }), 0);
  } else if (!allDone) {
    celebrated = false;
    for (const q of [3, 2, 1]) {
      const th = Math.ceil((t.target * q) / 4);
      if (before < th && th <= t.done && !t.isTest) {
        setTimeout(() => bus.emit("celebrate", { kind: "quarter", quarter: q, id: `mock-q-${Date.now()}` }), 0);
        break;
      }
    }
  }
  return done();
}

export function SetExact(task: number, value: number) {
  const t = (snap.day.tasks ?? [])[task < 0 ? snap.day.active : task];
  return t ? Add(task, value - t.done, "main") : done();
}

export function MarkDone(task: number) {
  const t = (snap.day.tasks ?? [])[task < 0 ? snap.day.active : task];
  return t ? Add(task, t.target - t.done, "popover") : done();
}

export function Undo() {
  const e = log.pop();
  const t = e && (snap.day.tasks ?? [])[e.task];
  if (t && e) t.done = Math.max(0, t.done - e.delta);
  return done();
}

export function StartSession(minutes: number) {
  const m = minutes > 0 ? minutes : 50;
  snap.session = { endsAt: new Date(Date.now() + m * 60000).toISOString(), remainingSec: m * 60, totalSec: m * 60 };
  return done();
}

export function EndSession() {
  snap.session = null;
  return done();
}

export function MakePromise(minutes: number) {
  const at = new Date(Date.now() + minutes * 60000);
  snap.promise = { clock: `${String(at.getHours()).padStart(2, "0")}:${String(at.getMinutes()).padStart(2, "0")}`, used: (snap.promise?.used ?? 0) + 1 };
  return done();
}

export function SetOff(_date: string, off: string) {
  snap.day.off = off;
  snap.day.label = off === "rest" ? "Rest day" : off === "sick" ? "Sick day" : "Math day";
  return done();
}

export function SetDay(date: string, preset: string, targets: Record<string, number> | null) {
  const d = snap.strip?.find((x) => x.date === date);
  if (d) {
    d.preset = preset;
    d.off = preset === "rest" || preset === "sick" ? preset : "";
    d.label = { math: "Math day", english: "English day", test: "Practice test day", "test+math": "Practice test and math", "test+english": "Practice test and English", rest: "Rest day", sick: "Sick day" }[preset] ?? preset;
    d.tasks = preset.split("+").filter((k) => ["math", "english", "test"].includes(k)).map((k) => ({
      kind: k, label: k, target: targets?.[k] ?? (k === "test" ? 1 : 100), done: 0, fraction: 0, quickAdd: [1, 5, 10, 25], isTest: k === "test",
    }));
  }
  return done();
}

export function SuggestPlan(_overwrite: boolean) {
  return done();
}

export function UpdateSettings(s: Snapshot["settings"]) {
  snap.settings = s;
  return done();
}

export function CompleteOnboarding(targetScore: number) {
  snap.phase = "study";
  snap.settings.onboarded = true;
  snap.settings.targetScore = targetScore;
  return done();
}

export function SetScore(_date: string, score: number) {
  snap.day.score = score;
  return done();
}

const sample = (): IslandPayload => ({
  seq: ++seq, mode: "remind", level: 2, title: "37 of 100. 63 left, about 1 h 41 min.", body: "13 more closes this quarter.",
  sticky: true, hideAfterMs: 0, canLater: true, laterMax: 349, fraction: 0.37, done: 37, target: 100, isTest: false,
  quickAdd: [1, 5, 10, 25],
  buttons: [{ id: "add:2", label: "+10", primary: false }, { id: "add:3", label: "+25", primary: false }, { id: "studying", label: "I'm studying", primary: true }],
});

export function IslandDone(_seq: number, _reason: string) {
  return Promise.resolve();
}

export function IslandCurrent() {
  if (params.get("island") === "tally") {
    return Promise.resolve({
      seq: ++seq, mode: "tally", level: 0, title: "+1 math", body: "62 left today", sticky: false, hideAfterMs: 0,
      canLater: false, laterMax: 0, fraction: 0.38, done: 38, target: 100, isTest: false, quickAdd: [1, 5, 10, 25], buttons: [],
    });
  }
  return Promise.resolve(sample());
}

export function OpenMain(tab: string) {
  if (tab) bus.emit("navigate", tab);
  return Promise.resolve();
}

export function ShowPopover(_focus: string) {
  return Promise.resolve();
}

export function HidePopover() {
  return Promise.resolve();
}

export function RequestNotifications() {
  snap.system.notifications = true;
  refresh();
  return Promise.resolve(true);
}

export function OpenNotificationSettings() {
  return Promise.resolve();
}

export function SetLaunchAtLogin(on: boolean) {
  snap.system.autostart = on;
  return done();
}

export function TestReminder(_level: number) {
  bus.emit("island", sample());
  return Promise.resolve();
}

export function Bailout(phrase: string) {
  return phrase === "I am done for today" ? Promise.resolve() : Promise.reject(new Error('type "I am done for today" exactly'));
}

export function PauseHotkeys(_paused: boolean) {
  return Promise.resolve();
}

export function RetireAndQuit() {
  return Promise.resolve();
}

setup();
