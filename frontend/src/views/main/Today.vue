<template>
  <div class="today" v-if="s">
    <section class="left">
      <div class="head muted">
        <span>Day <span class="mono now">{{ s.dayNumber }}</span> of {{ s.dayCount }}, SAT in <span class="mono now">{{ s.daysLeft }}</span> {{ s.daysLeft === 1 ? "day" : "days" }}</span>
        <span>streak <span class="mono now">{{ s.streak }}</span></span>
      </div>
      <DayPanel large source="main" @lit="lit = $event" />
      <DayStrip class="strip" :days="s.strip ?? []" large :lit="lit" @open="$emit('plan')" />
    </section>

    <section class="right">
      <h2 class="heading">Log</h2>
      <p v-if="!s.day.entries?.length" class="muted empty">
        Nothing logged yet today. Press +1 after every question you finish; the entries land here with their time.
      </p>
      <ol v-else class="log">
        <li v-for="(e, i) in s.day.entries ?? []" :key="i">
          <span class="mono muted">{{ e.time }}</span>
          <span class="mono delta" :class="{ minus: e.text.startsWith('-') }">{{ e.text.split(" ")[0] }}</span>
          <span>{{ e.text.split(" ").slice(1).join(" ") }}</span>
          <span class="muted src">{{ sourceName(e.source) }}</span>
        </li>
      </ol>

      <template v-if="s.day.changes?.length">
        <h3>Changes today</h3>
        <ol class="log changes">
          <li v-for="(c, i) in s.day.changes" :key="'c' + i">
            <span class="mono muted">{{ c.time }}</span>
            <span>{{ c.text }}</span>
          </li>
        </ol>
      </template>

      <div class="anywhere">
        <div class="ahead">
          <strong>Log from any app</strong>
          <button class="link" @click="$emit('settings')">Change</button>
        </div>
        <div v-if="s.settings.hotkeys && keys.length" class="keys">
          <span v-for="k in keys" :key="k.action"><kbd>{{ keyLabel(k.acc) }}</kbd> {{ k.label }}</span>
        </div>
        <p v-else class="muted">Hotkeys are off.</p>
        <p class="muted">
          Or the link <code class="mono">saturn://add</code> from Shortcuts, Raycast or a Stream Deck.
        </p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import DayPanel from "../../components/DayPanel.vue";
import DayStrip from "../../components/DayStrip.vue";
import { snap } from "../../store";
import { keyLabel } from "../../format";
import { hotkeyActions } from "../../hotkeys";

defineEmits<{ plan: []; settings: [] }>();
const s = computed(() => snap.value);
const lit = ref(-1);

const keys = computed(() =>
  hotkeyActions
    .map((a) => ({ ...a, acc: s.value?.settings.keys?.[a.action] ?? "" }))
    .filter((k) => k.acc),
);

function sourceName(src: string): string {
  return ({ hotkey: "hotkey", link: "link", popover: "menu bar", main: "window", island: "island", menu: "menu", notification: "reminder" } as Record<string, string>)[src] ?? src;
}
</script>

<style scoped>
.today { display: grid; grid-template-columns: 400px 1fr; gap: 36px; padding: 22px 28px; min-height: 100%; box-sizing: border-box; }
.left { display: flex; flex-direction: column; gap: 16px; }
.head { display: flex; justify-content: space-between; }
.now { color: var(--text); }
.strip { margin-top: 4px; }
.right { display: flex; flex-direction: column; min-width: 0; }
.heading { margin-bottom: 12px; }
.empty { max-width: 40ch; line-height: 1.5; margin: 0; }
h3 { font-size: 13px; font-weight: 600; margin: 22px 0 6px; }
.log { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; }
.log li { display: grid; grid-template-columns: 48px 44px 1fr auto; align-items: baseline; gap: 8px; padding: 7px 0; border-bottom: 1px solid var(--line); }
.log.changes li { grid-template-columns: 48px 1fr; }
.delta { color: var(--accent); font-weight: 600; }
.delta.minus { color: var(--sub); }
.src { font-size: 12px; }
.anywhere {
  margin-top: auto; padding: 14px 16px; border-radius: 12px; background: var(--raised);
  display: flex; flex-direction: column; gap: 8px;
}
.anywhere p { margin: 0; line-height: 1.45; }
.ahead { display: flex; justify-content: space-between; align-items: baseline; }
.keys { display: flex; flex-wrap: wrap; gap: 6px 16px; }
code { color: var(--text); font-size: 12px; }
</style>
