<!-- src/components/processing/AlgorithmInfo.vue -->
<template>
  <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
    <h3 class="text-sm font-medium text-blue-800 mb-2">Algorithm Info</h3>
    <p class="text-sm text-blue-700">{{ getAlgorithmDescription() }}</p>
  </div>
</template>

<script setup lang="ts">
import type { ProcessingType } from '@/types'

const props = defineProps<{
  processingType: ProcessingType
}>()

const getAlgorithmDescription = (): string => {
  switch (props.processingType) {
    case 'binarization':
      return 'Converts grayscale images to black and white using various thresholding techniques. Otsu is best for images with clear bimodal histograms, while Sauvola and Niblack work better for documents with varying lighting.'

    case 'morphology':
      return 'Applies morphological operations to binary images. Opening removes noise, closing fills gaps, erosion shrinks objects, and dilation expands them. Useful for cleaning up binary images.'

    case 'noise-reduction':
      return 'Removes noise from images. Median filtering works well for salt-and-pepper noise, while binary noise removal eliminates small unwanted components from binary images.'

    case 'scaling':
      return 'Scales images using different algorithms. Scale2x/3x/4x are specialized for pixel art and maintain sharp edges, while nearest neighbor and bilinear offer traditional scaling approaches.'

    default:
      return 'Select an algorithm to see its description.'
  }
}
</script>