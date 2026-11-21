<template>
  <div class="pop" v-if="s">
    <header>
      <span class="muted tabnum">{{ s.phase === "study" ? `T-${s.daysLeft}` : "SATurn" }}</span>
      <span class="right">
        <span v-if="s.phase === 'study'" class="muted">streak {{ s.streak }}</span>
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
      <div class="display big">SATurn</div>
      <p class="muted">Thirteen rings to the SAT. Set up the plan first.</p>
      <button class="btn primary" @click="open('')">Finish setup</button>
    </section>

    <section v-else-if="s.phase === 'exam'" class="centre">
      <div class="display big">Today.</div>
      <p v-if="s.settings.targetScore" class="display score">{{ s.settings.targetScore }}</p>
      <p class="muted">You did the work. Go.</p>
    </section>

    <section v-else-if="s.phase === 'after'" class="centre">
      <div class="display big">That was the plan.</div>
      <p class="muted">{{ s.totals }}</p>
      <button class="btn primary" @click="open('')">Close SATurn</button>
    </section>

    <template v-else>
      <DayPanel source="popover" @lit="lit = $event" />
      <DayStrip class="strip" :days="s.strip ?? []" :lit="lit" @open="open('plan')" />
    </template>

    <p v-if="error" class="coral err">{{ error }}</p>
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
  height: 100%; box-sizing: border-box; padding: 12px 16px 14px; display: flex; flex-direction: column; gap: 10px;
  border-radius: 14px; overflow: hidden;
}
header { display: flex; justify-content: space-between; align-items: center; height: 22px; }
.right { display: flex; align-items: center; gap: 8px; }
.gear { background: none; border: 0; color: var(--dust); padding: 3px; border-radius: 6px; cursor: pointer; display: flex; }
.gear:hover { color: var(--ice); background: var(--deep); }
.centre { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; text-align: center; }
.big { font-size: 40px; color: var(--gold); }
.score { font-size: 56px; color: var(--gold); margin: 0; }
.strip { margin-top: auto; }
.err { font-size: 12px; margin: 0; }
</style>
