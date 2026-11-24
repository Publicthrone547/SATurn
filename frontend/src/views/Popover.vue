<template>
  <div class="pop" v-if="s">
    <header>
      <span class="muted">
        <template v-if="s.phase === 'study'">SAT in <span class="mono now">{{ s.daysLeft }}</span> {{ s.daysLeft === 1 ? "day" : "days" }}</template>
        <template v-else>SATurn</template>
      </span>
      <span class="right">
        <span v-if="s.phase === 'study'" class="muted" title="Days in a row with the work done">streak <span class="mono now">{{ s.streak }}</span></span>
        <button class="gear" aria-label="Settings" title="Settings" @click="open('settings')">
          <svg viewBox="0 0 16 16" width="15" height="15" aria-hidden="true">
            <circle cx="8" cy="8" r="2.4" fill="none" stroke="currentColor" stroke-width="1.4" />
            <path
              d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4"
              stroke="currentColor" stroke-width="1.4" stroke-linecap="round"
            />
          </svg>
        </button>
      </span>
    </header>

    <section v-if="s.phase === 'setup'" class="centre">
      <div class="heading">SATurn</div>
      <p class="muted">One count a day until the SAT. Set up the plan first.</p>
      <button class="btn primary" @click="open('')">Finish setup</button>
    </section>

    <section v-else-if="s.phase === 'exam'" class="centre">
      <div class="heading big">Exam day.</div>
      <p v-if="s.settings.targetScore" class="mono score">{{ s.settings.targetScore }}</p>
      <p class="muted">You did the work. Go.</p>
    </section>

    <section v-else-if="s.phase === 'after'" class="centre">
      <div class="heading big">That was the plan.</div>
      <p class="muted">{{ s.totals }}</p>
      <button class="btn primary" @click="open('')">Close SATurn</button>
    </section>

    <template v-else>
      <DayPanel source="popover" @lit="lit = $event" />
      <DayStrip class="strip" :days="s.strip ?? []" :lit="lit" @open="open('plan')" />
    </template>

    <p v-if="error" class="bad err">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import DayPanel from "../components/DayPanel.vue";
import DayStrip from "../components/DayStrip.vue";
import { snap, error } from "../store";
import { api } from "../api";

const s = computed(() => snap.value);
const lit = ref(-1);

function open(tab: string) {
  api.HidePopover();
  api.OpenMain(tab);
}
</script>

<style scoped>
.pop {
  height: 100%; box-sizing: border-box; padding: 12px 16px 12px; display: flex; flex-direction: column; gap: 12px;
  border-radius: 14px; overflow: hidden;
}
header { display: flex; justify-content: space-between; align-items: center; height: 22px; }
.right { display: flex; align-items: center; gap: 8px; }
.now { color: var(--text); }
.gear { background: none; border: 0; color: var(--sub); padding: 3px; border-radius: 6px; cursor: pointer; display: flex; }
.gear:hover { color: var(--text); background: var(--raised); }
.centre { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; text-align: center; }
.big { font-size: 30px; }
.score { font-size: 56px; font-weight: 600; letter-spacing: -0.04em; color: var(--accent); margin: 0; }
.strip { margin-top: auto; }
.err { font-size: 12px; margin: 0; }
</style>
