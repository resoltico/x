<!-- src/components/processing/WorkerStatusDisplay.vue -->
<template>
  <div class="mb-4 bg-slate-50 border border-slate-200 rounded-lg p-3">
    <h3 class="text-sm font-medium text-slate-700 mb-2">System Status</h3>
    <div class="grid grid-cols-2 gap-3 text-xs">
      <div class="flex items-center">
        <div 
          class="w-2 h-2 rounded-full mr-2"
          :class="status.initialized ? 'bg-green-500' : 'bg-red-500'"
        ></div>
        <span>{{ status.initialized ? 'Ready' : 'Initializing...' }}</span>
      </div>
      <div>
        <span class="text-slate-500">Workers:</span>
        <span class="ml-1 font-medium">{{ status.availableWorkers }}/{{ status.totalWorkers }}</span>
      </div>
      <div>
        <span class="text-slate-500">Queue:</span>
        <span class="ml-1 font-medium">{{ status.queuedTasks }}</span>
      </div>
      <div>
        <span class="text-slate-500">Environment:</span>
        <span class="ml-1 font-medium text-xs">{{ status.environment }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SystemStatus } from '@/modules/processing/SystemStatusManager'

defineProps<{
  status: SystemStatus
}>()
</script>