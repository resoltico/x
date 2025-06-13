<!-- src/components/processing/parameters/ScalingControls.vue -->
<template>
  <div class="space-y-4">
    <div>
      <label class="label">Method</label>
      <select 
        :model-value="parameters.method" 
        class="input"
        @input="updateMethod(($event.target as HTMLSelectElement).value as any)"
      >
        <option value="scale2x">Scale2x (Pixel Art)</option>
        <option value="scale3x">Scale3x (Pixel Art)</option>
        <option value="scale4x">Scale4x (Pixel Art)</option>
        <option value="nearest">Nearest Neighbor</option>
        <option value="bilinear">Bilinear</option>
      </select>
    </div>

    <div v-if="!parameters.method.startsWith('scale')">
      <label class="label">
        Scale Factor: {{ parameters.factor.toFixed(1) }}x
      </label>
      <input
        :model-value="parameters.factor"
        type="range"
        min="0.1"
        max="8"
        step="0.1"
        class="slider"
        @input="updateFactor(parseFloat(($event.target as HTMLInputElement).value))"
      />
    </div>

    <div v-else class="text-sm text-slate-600">
      <p>Pixel art scaling methods use fixed scale factors:</p>
      <ul class="list-disc list-inside mt-1">
        <li>Scale2x: 2x scaling</li>
        <li>Scale3x: 3x scaling</li>
        <li>Scale4x: 4x scaling</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ScalingParams } from '@/types'

const props = defineProps<{
  parameters: ScalingParams
}>()

const emit = defineEmits<{
  'update:parameters': [parameters: ScalingParams]
}>()

const updateMethod = (method: ScalingParams['method']) => {
  let factor = props.parameters.factor
  
  // Auto-set factor for pixel art methods
  if (method === 'scale2x') factor = 2
  else if (method === 'scale3x') factor = 3
  else if (method === 'scale4x') factor = 4
  
  emit('update:parameters', {
    ...props.parameters,
    method,
    factor
  })
}

const updateFactor = (factor: number) => {
  emit('update:parameters', {
    ...props.parameters,
    factor
  })
}
</script>