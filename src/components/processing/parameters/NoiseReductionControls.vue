<!-- src/components/processing/parameters/NoiseReductionControls.vue -->
<template>
  <div class="space-y-4">
    <div>
      <label class="label">Method</label>
      <select 
        :model-value="parameters.method" 
        class="input"
        @input="updateMethod(($event.target as HTMLSelectElement).value as any)"
      >
        <option value="median">Median Filter</option>
        <option value="binary-noise-removal">Binary Noise Removal</option>
      </select>
    </div>

    <div v-if="parameters.method === 'median'">
      <label class="label">
        Kernel Size: {{ parameters.kernelSize || 3 }}
      </label>
      <input
        :model-value="parameters.kernelSize || 3"
        type="range"
        min="3"
        max="9"
        step="2"
        class="slider"
        @input="updateKernelSize(parseInt(($event.target as HTMLInputElement).value))"
      />
    </div>

    <div v-if="parameters.method === 'binary-noise-removal'">
      <label class="label">
        Threshold: {{ parameters.threshold || 50 }}
      </label>
      <input
        :model-value="parameters.threshold || 50"
        type="range"
        min="1"
        max="1000"
        step="10"
        class="slider"
        @input="updateThreshold(parseInt(($event.target as HTMLInputElement).value))"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { NoiseReductionParams } from '@/types'

const props = defineProps<{
  parameters: NoiseReductionParams
}>()

const emit = defineEmits<{
  'update:parameters': [parameters: NoiseReductionParams]
}>()

const updateMethod = (method: NoiseReductionParams['method']) => {
  emit('update:parameters', {
    ...props.parameters,
    method
  })
}

const updateKernelSize = (kernelSize: number) => {
  emit('update:parameters', {
    ...props.parameters,
    kernelSize
  })
}

const updateThreshold = (threshold: number) => {
  emit('update:parameters', {
    ...props.parameters,
    threshold
  })
}
</script>