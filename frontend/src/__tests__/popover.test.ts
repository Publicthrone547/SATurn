import { beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import base from "../mock-snapshot.json";

const calls: unknown[][] = [];
vi.mock("../api", () => {
  const snap = () => JSON.parse(JSON.stringify(base));
  const api = new Proxy(
    {},
    {
      get: (_t, name: string) =>
        (...args: unknown[]) => {
          calls.push([name, ...args]);
          return Promise.resolve(name === "IslandCurrent" ? null : snap());
        },
    },
  );
  return { api, on: () => () => {}, useMock: true, errMsg: (e: unknown) => String(e) };
});

import Popover from "../views/Popover.vue";
import { initStore } from "../store";

describe("Popover", () => {
  beforeEach(() => {
    calls.length = 0;
  });

  it("renders the mock snapshot and logs with the quick-add buttons", async () => {
    await initStore();
    const w = mount(Popover);
    await flushPromises();
    expect(w.text()).toContain("Math day");
    expect(w.text()).toContain("63 left, about 1 h 41 min");
    expect(w.text()).toContain("T-11");
    const buttons = w.findAll("button").filter((b) => /^\+\d+$/.test(b.text()));
    expect(buttons.map((b) => b.text())).toEqual(["+1", "+5", "+10", "+25"]);
    await buttons[2].trigger("click");
    await flushPromises();
    expect(calls).toContainEqual(["Add", 0, 10, "popover"]);
  });
});
