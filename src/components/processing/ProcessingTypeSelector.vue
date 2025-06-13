<!-- src/components/processing/ProcessingTypeSelector.vue -->
<template>
  <div>
    <label class="label">Processing Type</label>
    <select
      :model-value="modelValue"
      class="input"
      :disabled="disabled"
      @input="$emit('update:modelValue', ($event.target as HTMLSelectElement).value as ProcessingType | '')"
    >
      <option value="">Select algorithm...</option>
      <option value="binarization">Binarization</option>
      <option value="morphology">Morphological Operations</option>
      <option value="noise-reduction">Noise Reduction</option>
      <option value="scaling">Pixel Art Scaling</option>
    </select>
  </div>
</template>

<script setup lang="ts">
import type { ProcessingType } from '@/types'

defineProps<{
  modelValue: ProcessingType | ''
  disabled?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: ProcessingType | '']
}>()
</script>