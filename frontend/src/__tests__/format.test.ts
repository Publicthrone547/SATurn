import { describe, expect, it } from "vitest";
import { arc, clock, countdown, duration, minutes, quarters } from "../format";

describe("ring quarters", () => {
  it("splits a fraction into four quarters", () => {
    expect(quarters(0)).toEqual([0, 0, 0, 0]);
    expect(quarters(0.37).map((q) => +q.toFixed(2))).toEqual([1, 0.48, 0, 0]);
    expect(quarters(1)).toEqual([1, 1, 1, 1]);
    expect(quarters(1.12)).toEqual([1, 1, 1, 1]);
  });
  it("draws arcs clockwise from the top", () => {
    expect(arc(100, 100, 80, 0, 90)).toBe("M 100.000 20.000 A 80 80 0 0 1 180.000 100.000");
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
