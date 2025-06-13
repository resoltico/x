<!-- src/components/processing/ProcessingActions.vue -->
<template>
  <div class="flex space-x-3 pt-4 border-t border-slate-200">
    <button
      class="btn btn-secondary flex-1"
      :disabled="!canProcess || isProcessing"
      @click="$emit('preview')"
    >
      <svg v-if="isProcessing" class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      {{ isProcessing ? 'Processing...' : 'Preview' }}
    </button>
    
    <button
      class="btn btn-primary flex-1"
      :disabled="!canProcess || isProcessing"
      @click="$emit('process')"
    >
      Process Full Size
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  canProcess: boolean
  isProcessing: boolean
}>()

defineEmits<{
  preview: []
  process: []
}>()
</script>