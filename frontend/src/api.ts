// The service the views call: the generated Wails bindings, or an in-memory
// fake when built with VITE_MOCK=1 (npm run dev:mock) for interface work.
import { Events } from "@wailsio/runtime";
import * as real from "../bindings/github.com/Publicthrone547/reminder_sat/appservice.js";
import * as mock from "./mock";

export type * from "../bindings/github.com/Publicthrone547/reminder_sat/internal/core/models.js";
export type { IslandPayload, IslandButton } from "../bindings/github.com/Publicthrone547/reminder_sat/models.js";

export const useMock = import.meta.env.VITE_MOCK === "1";

export const api: typeof real = useMock ? (mock as unknown as typeof real) : real;

/** Subscribe to a Go event; returns the unsubscribe function. */
export function on<T = unknown>(name: string, cb: (data: T) => void): () => void {
  if (useMock) return mock.bus.on(name, cb as (d: unknown) => void);
  return Events.On(name, (ev: { data: unknown }) => cb(ev.data as T));
}

export function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message;
  if (typeof e === "string") return e;
  const m = (e as { message?: string } | null)?.message;
  return m ?? String(e);
}
