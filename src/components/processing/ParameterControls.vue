<!-- src/components/processing/ParameterControls.vue -->
<template>
  <div>
    <!-- Binarization Controls -->
    <BinarizationControls
      v-if="processingType === 'binarization'"
      :parameters="parameters.binarization || defaultBinarizationParams"
      @update:parameters="updateBinarizationParams"
    />

    <!-- Morphology Controls -->
    <MorphologyControls
      v-else-if="processingType === 'morphology'"
      :parameters="parameters.morphology || defaultMorphologyParams"
      @update:parameters="updateMorphologyParams"
    />

    <!-- Noise Reduction Controls -->
    <NoiseReductionControls
      v-else-if="processingType === 'noise-reduction'"
      :parameters="parameters.noise || defaultNoiseParams"
      @update:parameters="updateNoiseParams"
    />

    <!-- Scaling Controls -->
    <ScalingControls
      v-else-if="processingType === 'scaling'"
      :parameters="parameters.scaling || defaultScalingParams"
      @update:parameters="updateScalingParams"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { 
  ProcessingType, 
  ProcessingParameters,
  BinarizationParams,
  MorphologyParams,
  NoiseReductionParams,
  ScalingParams
} from '@/types'

// Import child components
import BinarizationControls from './parameters/BinarizationControls.vue'
import MorphologyControls from './parameters/MorphologyControls.vue'
import NoiseReductionControls from './parameters/NoiseReductionControls.vue'
import ScalingControls from './parameters/ScalingControls.vue'

const props = defineProps<{
  processingType: ProcessingType
  parameters: ProcessingParameters
}>()

const emit = defineEmits<{
  'update:parameters': [parameters: ProcessingParameters]
}>()

// Default parameters
const defaultBinarizationParams: BinarizationParams = {
  method: 'otsu',
  windowSize: 15,
  k: 0.2,
  threshold: 128
}

const defaultMorphologyParams: MorphologyParams = {
  operation: 'opening',
  kernelSize: 3,
  iterations: 1
}

const defaultNoiseParams: NoiseReductionParams = {
  method: 'median',
  kernelSize: 3,
  threshold: 50
}

const defaultScalingParams: ScalingParams = {
  method: 'scale2x',
  factor: 2
}

// Update methods
const updateBinarizationParams = (newParams: BinarizationParams) => {
  emit('update:parameters', {
    ...props.parameters,
    binarization: newParams
  })
}

const updateMorphologyParams = (newParams: MorphologyParams) => {
  emit('update:parameters', {
    ...props.parameters,
    morphology: newParams
  })
}

const updateNoiseParams = (newParams: NoiseReductionParams) => {
  emit('update:parameters', {
    ...props.parameters,
    noise: newParams
  })
}

const updateScalingParams = (newParams: ScalingParams) => {
  emit('update:parameters', {
    ...props.parameters,
    scaling: newParams
  })
}
</script>