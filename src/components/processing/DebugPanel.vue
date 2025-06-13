<!-- src/components/processing/DebugPanel.vue -->
<template>
  <div class="bg-gray-50 border border-gray-200 rounded-lg p-4">
    <h3 class="text-sm font-medium text-gray-800 mb-2">Debug Information</h3>
    <div class="text-xs text-gray-600 space-y-1 font-mono">
      <div>Worker Status: {{ JSON.stringify(systemStatus) }}</div>
      <div>Has Image: {{ hasImage }}</div>
      <div>Can Process: {{ canProcess }}</div>
      <div>Selected Type: {{ selectedType }}</div>
      <div>Active Tasks: {{ activeTasks.length }}</div>
      <div>Environment: {{ getEnvironmentInfo() }}</div>
      <div>Memory: {{ getMemoryInfo() }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SystemStatus } from '@/modules/processing/SystemStatusManager'
import type { ProcessingTask, ProcessingType } from '@/types'

defineProps<{
  systemStatus: SystemStatus
  hasImage: boolean
  canProcess: boolean
  selectedType: ProcessingType | ''
  activeTasks: ProcessingTask[]
}>()

const getEnvironmentInfo = (): string => {
  if (typeof window === 'undefined') return 'Server-side'
  return `${window.location.protocol}//${window.location.host}`
}

const getMemoryInfo = (): string => {
  const performance = globalThis.performance as any
  if (performance?.memory) {
    const used = Math.round(performance.memory.usedJSHeapSize / 1024 / 1024)
    const total = Math.round(performance.memory.totalJSHeapSize / 1024 / 1024)
    return `${used}/${total} MB`
  }
  return 'N/A'
}
</script>