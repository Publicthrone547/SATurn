import { describe, expect, it } from "vitest";
import { accelerator, clock, countdown, duration, keyLabel, minutes, tally } from "../format";

describe("tally", () => {
  it("shows one cell per question up to 100", () => {
    expect(tally(100)).toEqual({ cells: 100, per: 1, cols: 25 });
    expect(tally(60)).toEqual({ cells: 60, per: 1, cols: 15 });
    expect(tally(250)).toEqual({ cells: 84, per: 3, cols: 21 });
  });
});

describe("hotkeys", () => {
  const ev = (code: string, m: Partial<Record<"ctrlKey" | "altKey" | "shiftKey" | "metaKey", boolean>> = {}) => ({
    code, ctrlKey: false, altKey: false, shiftKey: false, metaKey: false, ...m,
  });
  it("labels accelerators with the menu glyphs", () => {
    expect(keyLabel("Ctrl+Option+1")).toBe("⌃⌥1");
    expect(keyLabel("Cmd+Shift+k")).toBe("⇧⌘K");
    expect(keyLabel("Option+Space")).toBe("⌥Space");
    expect(keyLabel("")).toBe("");
  });
  it("turns a key press into an accelerator", () => {
    expect(accelerator(ev("Digit1", { ctrlKey: true, altKey: true }))).toBe("Ctrl+Option+1");
    expect(accelerator(ev("KeyK", { metaKey: true, shiftKey: true }))).toBe("Shift+Cmd+K");
    expect(accelerator(ev("F5", { ctrlKey: true }))).toBe("Ctrl+F5");
    expect(accelerator(ev("Digit1", { shiftKey: true }))).toBe("");
    expect(accelerator(ev("ShiftLeft", { ctrlKey: true }))).toBe("");
  });
});

describe("formatters", () => {
  it("formats durations like the core", () => {
    expect(duration(101)).toBe("1 h 41 min");
    expect(duration(60)).toBe("1 h");
    expect(duration(40)).toBe("40 min");
  });
  it("formats and parses clock times", () => {
    expect(clock(1220)).toBe("20:20");
    expect(clock(0)).toBe("00:00");
    expect(minutes("20:20")).toBe(1220);
    expect(minutes("25:00")).toBeNaN();
  });
  it("formats a countdown", () => {
    expect(countdown(24 * 60 + 5)).toBe("24:05");
    expect(countdown(-3)).toBe("0:00");
  });
});
