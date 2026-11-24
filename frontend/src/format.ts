// Small pure helpers shared by the views.

/** The cells of a day's tally: up to 100, each worth `per` questions, in
 * four rows so that every row is a quarter of the day. */
export function tally(target: number): { cells: number; per: number; cols: number } {
  const t = Math.max(1, Math.round(target));
  const per = Math.ceil(t / 100);
  const cells = Math.ceil(t / per);
  return { cells, per, cols: Math.ceil(cells / 4) };
}

/** "24:05" style countdown from seconds. */
export function countdown(sec: number): string {
  const s = Math.max(0, Math.round(sec));
  const m = Math.floor(s / 60);
  return `${m}:${String(s % 60).padStart(2, "0")}`;
}

/** Minutes after midnight as "20:20". */
export function clock(min: number): string {
  const m = ((Math.round(min) % 1440) + 1440) % 1440;
  return `${String(Math.floor(m / 60)).padStart(2, "0")}:${String(m % 60).padStart(2, "0")}`;
}

/** "20:20" back to minutes after midnight; NaN when invalid. */
export function minutes(hhmm: string): number {
  const m = /^(\d{1,2}):(\d{2})$/.exec(hhmm.trim());
  if (!m) return NaN;
  const h = Number(m[1]);
  const mm = Number(m[2]);
  if (h > 23 || mm > 59) return NaN;
  return h * 60 + mm;
}

/** Minutes as "40 min", "1 h", "1 h 41 min". */
export function duration(min: number): string {
  const m = Math.max(0, Math.round(min));
  const h = Math.floor(m / 60);
  const r = m % 60;
  if (h === 0) return `${r} min`;
  if (r === 0) return `${h} h`;
  return `${h} h ${r} min`;
}

const modSymbols: [RegExp, string][] = [
  [/^(ctrl|control)$/i, "⌃"],
  [/^(option|alt|optionoralt)$/i, "⌥"],
  [/^shift$/i, "⇧"],
  [/^(cmd|command|cmdorctrl|super)$/i, "⌘"],
];
const keyNames: Record<string, string> = {
  space: "Space", return: "↩", enter: "↩", tab: "⇥", escape: "⎋", backspace: "⌫", delete: "⌦",
  left: "←", right: "→", up: "↑", down: "↓", plus: "+",
};

/** "Ctrl+Option+1" as the macOS menu glyphs, "⌃⌥1". */
export function keyLabel(acc: string): string {
  if (!acc) return "";
  const parts = acc.split("+");
  const key = parts.pop() ?? "";
  const mods = modSymbols.map(([re, sym]) => (parts.some((p) => re.test(p)) ? sym : "")).join("");
  return mods + (keyNames[key.toLowerCase()] ?? key.toUpperCase());
}

const codeKeys: Record<string, string> = {
  Space: "Space", Enter: "Return", Tab: "Tab", ArrowLeft: "Left", ArrowRight: "Right", ArrowUp: "Up", ArrowDown: "Down",
  Minus: "-", Equal: "=", BracketLeft: "[", BracketRight: "]", Semicolon: ";", Quote: "'", Comma: ",", Period: ".",
  Slash: "/", Backslash: "\\", Backquote: "`",
};

/** A key press as a global-hotkey accelerator ("Ctrl+Option+K"), or "" while
 * it is incomplete: a hotkey needs Control, Option or Command. The physical
 * key decides, so Option+1 is "1" and not "¡", in any keyboard layout. */
export function accelerator(e: { code: string; ctrlKey: boolean; altKey: boolean; shiftKey: boolean; metaKey: boolean }): string {
  if (!e.ctrlKey && !e.altKey && !e.metaKey) return "";
  let key = "";
  const c = e.code;
  if (/^Key[A-Z]$/.test(c)) key = c.slice(3);
  else if (/^Digit\d$/.test(c)) key = c.slice(5);
  else if (/^Numpad\d$/.test(c)) key = c.slice(6);
  else if (/^F\d{1,2}$/.test(c)) key = c;
  else key = codeKeys[c] ?? "";
  if (!key) return "";
  const mods = [e.ctrlKey && "Ctrl", e.altKey && "Option", e.shiftKey && "Shift", e.metaKey && "Cmd"].filter(Boolean);
  return [...mods, key].join("+");
}

export function reducedMotion(): boolean {
  try {
    return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  } catch {
    return false;
  }
}
