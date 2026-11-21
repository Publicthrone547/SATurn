<template>
  <div class="island" @mouseenter="hover = true" @mouseleave="hover = false">
    <div v-if="p" class="content" :key="p.seq">
      <Ring
        class="mini"
        :fraction="ringFraction"
        :done="p.done"
        :target="p.target"
        :is-test="p.isTest"
        :size="56"
        :number-size="17"
        :lock="lockKey"
      >
        <span class="mid display tabnum">{{ p.isTest ? (p.fraction >= 1 ? "✓" : "T") : p.done }}</span>
      </Ring>
      <div class="text">
        <div class="title">{{ p.title }}</div>
        <div class="body">{{ p.body }}</div>
        <div v-if="laterOpen" class="row">
          <LaterChooser :max="p.laterMax" @choose="later" @cancel="laterOpen = false" />
        </div>
        <div v-else-if="p.buttons?.length" class="row">
          <button
            v-for="b in p.buttons"
            :key="b.id"
            class="btn small"
            :class="{ primary: b.primary }"
            @click="press(b.id)"
          >
            {{ b.label }}
          </button>
          <span class="spacer" />
          <button v-if="p.mode === 'remind' && p.canLater" class="link" @click="laterOpen = true">Later</button>
          <button v-else-if="p.mode === 'remind' || p.mode === 'welcome'" class="link" @click="close('dismiss')">Not now</button>
        </div>
        <button v-if="p.sticky && !p.buttons?.length" class="link close" @click="close('dismiss')">Close</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import Ring from "../components/Ring.vue";
import LaterChooser from "../components/LaterChooser.vue";
import { api, on, type IslandPayload } from "../api";
import { run } from "../store";

const p = ref<IslandPayload | null>(null);
const hover = ref(false);
const laterOpen = ref(false);
const lockKey = ref("");
let timer = 0;
let left = 0;
let startedAt = 0;

const ringFraction = computed(() => (p.value?.mode === "celebrate" ? 1 : p.value?.fraction ?? 0));

function show(payload: IslandPayload) {
  p.value = payload;
  laterOpen.value = false;
  clearTimeout(timer);
  if (payload.mode === "celebrate") lockKey.value = `c${payload.seq}`;
  if (!payload.sticky && payload.hideAfterMs > 0) {
    left = payload.hideAfterMs;
    arm();
  }
}

function arm() {
  clearTimeout(timer);
  startedAt = Date.now();
  timer = window.setTimeout(() => {
    if (hover.value) {
      // Paused while the pointer is over the island; check again shortly.
      left = 1000;
      arm();
      return;
    }
    close("timeout");
  }, left);
}

function close(reason: string) {
  clearTimeout(timer);
  const seq = p.value?.seq;
  if (seq != null) api.IslandDone(seq, reason);
}

async function press(id: string) {
  const q = p.value?.quickAdd ?? [1, 5, 10, 25];
  if (id.startsWith("add:")) {
    await run(api.Add(-1, q[Number(id.slice(4))] ?? 0, "island"));
  } else if (id === "studying") {
    await run(api.StartSession(0));
  } else if (id === "more") {
    await run(api.StartSession(30));
  } else if (id === "done") {
    await run(api.MarkDone(-1));
  } else if (id === "other") {
    close("button");
    api.ShowPopover("exact");
    return;
  }
  close("button");
}

async function later(min: number) {
  if (await run(api.MakePromise(min))) close("button");
}

let off: (() => void) | null = null;
onMounted(async () => {
  off = on<IslandPayload>("island", show);
  const cur = await api.IslandCurrent().catch(() => null);
  if (cur && !p.value) show(cur);
});
onBeforeUnmount(() => {
  off?.();
  clearTimeout(timer);
});
void startedAt;
</script>

<style scoped>
.island { height: 100%; background: #000; color: var(--ice); overflow: hidden; }
.content {
  height: 100%; display: flex; align-items: center; gap: 14px; padding: 6px 18px 8px;
  box-sizing: border-box; animation: enter 280ms var(--calm) both;
}
@keyframes enter { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: none; } }
.mini { flex: none; }
.mid { font-size: 20px; color: var(--gold); }
.text { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 3px; }
.title { font-size: 14px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.body { font-size: 12px; color: var(--dust); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.row { display: flex; align-items: center; gap: 6px; margin-top: 4px; }
.row .btn { background: #1b1f45; }
.row .btn.primary { background: var(--gold); }
.spacer { flex: 1; }
.close { align-self: flex-end; }
@media (prefers-reduced-motion: reduce) { .content { animation: none; } }
</style>
