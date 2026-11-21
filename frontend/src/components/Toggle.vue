<template>
  <label class="toggle">
    <input type="checkbox" role="switch" :checked="modelValue" :disabled="disabled" @change="onChange" />
    <span class="track"><span class="knob" /></span>
    <span v-if="label" class="text">{{ label }}</span>
  </label>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ modelValue: boolean; label?: string; disabled?: boolean }>(), { label: "", disabled: false });
const emit = defineEmits<{ "update:modelValue": [v: boolean] }>();
function onChange(e: Event) {
  emit("update:modelValue", (e.target as HTMLInputElement).checked);
}
</script>

<style scoped>
.toggle { display: inline-flex; align-items: center; gap: 10px; cursor: pointer; }
input { position: absolute; opacity: 0; width: 0; height: 0; }
.track { width: 34px; height: 20px; border-radius: 10px; background: var(--line); position: relative; transition: background 140ms var(--calm); flex: none; }
.knob {
  position: absolute; top: 2px; left: 2px; width: 16px; height: 16px; border-radius: 50%;
  background: var(--ice); transition: transform 180ms var(--calm);
}
input:checked + .track { background: var(--gold); }
input:checked + .track .knob { transform: translateX(14px); background: var(--space); }
input:focus-visible + .track { outline: 2px solid var(--gold); outline-offset: 2px; }
input:disabled + .track { opacity: 0.4; }
</style>
