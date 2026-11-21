<template>
  <div class="today" v-if="s">
    <section class="left">
      <div class="head">
        <span class="muted tabnum">Day {{ s.dayNumber }} of {{ s.dayCount }} · T-{{ s.daysLeft }}</span>
        <span class="muted">streak {{ s.streak }}</span>
      </div>
      <DayPanel large source="main" @lit="lit = $event" />
      <DayStrip :days="s.strip ?? []" :size="18" :lit="lit" @open="$emit('plan')" />
    </section>
    <section class="right">
      <h2 class="display">Today's log</h2>
      <p v-if="!s.day.entries?.length" class="muted">Nothing logged yet. Every entry shows up here with its time.</p>
      <ul class="log">
        <li v-for="(e, i) in s.day.entries ?? []" :key="i">
          <span class="tabnum muted">{{ e.time }}</span>
          <span class="tabnum">{{ e.text }}</span>
          <span class="muted src">{{ e.source }}</span>
        </li>
      </ul>
      <template v-if="s.day.changes?.length">
        <h3>Changes today</h3>
        <ul class="log">
          <li v-for="(c, i) in s.day.changes" :key="'c' + i">
            <span class="tabnum muted">{{ c.time }}</span>
            <span>{{ c.text }}</span>
          </li>
        </ul>
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import DayPanel from "../../components/DayPanel.vue";
import DayStrip from "../../components/DayStrip.vue";
import { snap } from "../../store";

defineEmits<{ plan: [] }>();
const s = computed(() => snap.value);
const lit = ref(-1);
</script>

<style scoped>
.today { display: grid; grid-template-columns: 380px 1fr; gap: 32px; padding: 20px 28px; min-height: 100%; box-sizing: border-box; }
.left { display: flex; flex-direction: column; gap: 14px; }
.head { display: flex; justify-content: space-between; }
.right { border-left: 1px solid var(--line); padding-left: 28px; }
h2 { font-size: 26px; margin: 0 0 12px; color: var(--ice); }
h3 { font-size: 13px; font-weight: 600; margin: 20px 0 8px; }
.log { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; }
.log li { display: grid; grid-template-columns: 52px 1fr auto; gap: 10px; padding: 6px 0; border-bottom: 1px solid var(--line); }
.src { font-size: 12px; }
</style>
