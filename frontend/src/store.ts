// One reactive snapshot, replaced on every "state" event. The webviews hold
// no state of their own: every button calls the service and renders the
// snapshot that comes back.
import { ref, shallowRef } from "vue";
import { api, on, errMsg, type Snapshot } from "./api";

export const snap = shallowRef<Snapshot | null>(null);
export const error = ref("");

export interface CelebrateEvent {
  kind: string; // quarter | day | test | final
  id: string;
  title?: string;
  body?: string;
  quarter?: number;
  streak?: number;
  tomorrow?: string;
  totals?: string;
  dayNumber?: number;
}

/** The last celebration; components play each id once. */
export const celebration = shallowRef<CelebrateEvent | null>(null);

let started = false;

export async function initStore() {
  if (started) return;
  started = true;
  on<Snapshot>("state", (d) => {
    snap.value = d;
  });
  on<CelebrateEvent>("celebrate", (d) => {
    celebration.value = d;
  });
  try {
    snap.value = await api.Snapshot();
  } catch (e) {
    error.value = errMsg(e);
  }
}

/** Runs a service call and adopts the snapshot it returns. */
export async function run(p: Promise<Snapshot | void | null> | PromiseLike<Snapshot | void | null>): Promise<boolean> {
  try {
    const s = await p;
    if (s && typeof s === "object" && "phase" in s) snap.value = s as Snapshot;
    error.value = "";
    return true;
  } catch (e) {
    error.value = errMsg(e);
    return false;
  }
}
