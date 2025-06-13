<!-- src/components/processing/parameters/BinarizationControls.vue -->
<template>
  <div class="space-y-4">
    <div>
      <label class="label">Method</label>
      <select 
        :model-value="parameters.method" 
        class="input"
        @input="updateMethod(($event.target as HTMLSelectElement).value as any)"
      >
        <option value="otsu">Otsu (Global)</option>
        <option value="sauvola">Sauvola (Adaptive)</option>
        <option value="niblack">Niblack (Adaptive)</option>
      </select>
    </div>

    <div v-if="parameters.method === 'sauvola' || parameters.method === 'niblack'">
      <label class="label">
        Window Size: {{ parameters.windowSize }}
      </label>
      <input
        :model-value="parameters.windowSize"
        type="range"
        min="3"
        max="51"
        step="2"
        class="slider"
        @input="updateWindowSize(parseInt(($event.target as HTMLInputElement).value))"
      />
    </div>

    <div v-if="parameters.method === 'sauvola' || parameters.method === 'niblack'">
      <label class="label">
        K Factor: {{ (parameters.k ?? 0.2).toFixed(2) }}
      </label>
      <input
        :model-value="parameters.k ?? 0.2"
        type="range"
        min="-1"
        max="1"
        step="0.01"
        class="slider"
        @input="updateK(parseFloat(($event.target as HTMLInputElement).value))"
      />
    </div>

    <div v-if="parameters.method === 'otsu'">
      <label class="label">
        Threshold: {{ parameters.threshold }}
      </label>
      <input
        :model-value="parameters.threshold"
        type="range"
        min="0"
        max="255"
        step="1"
        class="slider"
        @input="updateThreshold(parseInt(($event.target as HTMLInputElement).value))"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { BinarizationParams } from '@/types'

const props = defineProps<{
  parameters: BinarizationParams
}>()

const emit = defineEmits<{
  'update:parameters': [parameters: BinarizationParams]
}>()

const updateMethod = (method: BinarizationParams['method']) => {
  emit('update:parameters', {
    ...props.parameters,
    method
  })
}

const updateWindowSize = (windowSize: number) => {
  emit('update:parameters', {
    ...props.parameters,
    windowSize
  })
}

const updateK = (k: number) => {
  emit('update:parameters', {
    ...props.parameters,
    k
  })
}

const updateThreshold = (threshold: number) => {
  emit('update:parameters', {
    ...props.parameters,
    threshold
  })
}
</script>