<!-- src/components/processing/parameters/MorphologyControls.vue -->
<template>
  <div class="space-y-4">
    <div>
      <label class="label">Operation</label>
      <select 
        :model-value="parameters.operation" 
        class="input"
        @input="updateOperation(($event.target as HTMLSelectElement).value as any)"
      >
        <option value="opening">Opening (Erosion + Dilation)</option>
        <option value="closing">Closing (Dilation + Erosion)</option>
        <option value="erosion">Erosion</option>
        <option value="dilation">Dilation</option>
      </select>
    </div>

    <div>
      <label class="label">
        Kernel Size: {{ parameters.kernelSize }}
      </label>
      <input
        :model-value="parameters.kernelSize"
        type="range"
        min="3"
        max="15"
        step="2"
        class="slider"
        @input="updateKernelSize(parseInt(($event.target as HTMLInputElement).value))"
      />
    </div>

    <div>
      <label class="label">
        Iterations: {{ parameters.iterations }}
      </label>
      <input
        :model-value="parameters.iterations"
        type="range"
        min="1"
        max="5"
        step="1"
        class="slider"
        @input="updateIterations(parseInt(($event.target as HTMLInputElement).value))"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { MorphologyParams } from '@/types'

const props = defineProps<{
  parameters: MorphologyParams
}>()

const emit = defineEmits<{
  'update:parameters': [parameters: MorphologyParams]
}>()

const updateOperation = (operation: MorphologyParams['operation']) => {
  emit('update:parameters', {
    ...props.parameters,
    operation
  })
}

const updateKernelSize = (kernelSize: number) => {
  emit('update:parameters', {
    ...props.parameters,
    kernelSize
  })
}

const updateIterations = (iterations: number) => {
  emit('update:parameters', {
    ...props.parameters,
    iterations
  })
}
</script>