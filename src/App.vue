<template>
  <div id="app" class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-slate-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <h1 class="text-2xl font-bold text-slate-900">
              Engraving Processor Pro
            </h1>
            <span class="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
              AI-Powered
            </span>
            <!-- System Status -->
            <div class="flex items-center space-x-2 text-xs">
              <div 
                class="w-2 h-2 rounded-full"
                :class="systemStatus.color"
                :title="systemStatus.text"
              ></div>
              <span class="text-slate-500">{{ systemStatus.text }}</span>
            </div>
          </div>
          <div class="flex items-center space-x-4">
            <span class="text-sm text-slate-500">v{{ version }}</span>
          </div>
        </div>
      </div>
    </header>

    <!-- Error Banner -->
    <div v-if="error" class="bg-red-600 text-white px-4 py-2 text-sm">
      <div class="max-w-7xl mx-auto flex items-center justify-between">
        <div class="flex items-center">
          <svg class="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
          <span>{{ error }}</span>
        </div>
        <button class="hover:bg-red-700 px-2 py-1 rounded text-xs" @click="error = null">
          Dismiss
        </button>
      </div>
    </div>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- Left Column: Image Input & Preview -->
        <div class="lg:col-span-2 space-y-6">
          <ImageInput />
          <PreviewRenderer />
        </div>

        <!-- Right Column: Controls & Processing -->
        <div class="space-y-6">
          <ProcessingControls />
          <ProgressDisplay />
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t border-slate-200 mt-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="text-center text-sm text-slate-500">
          <p>&copy; 2025 Ervins Strauhmanis. Built with Vue 3, TypeScript & WebAssembly.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import ImageInput from './components/ImageInput.vue'
import PreviewRenderer from './components/PreviewRenderer.vue'
import ProcessingControls from './components/ProcessingControls.vue'
import ProgressDisplay from './components/ProgressDisplay.vue'

const version = ref('1.0.0')
const error = ref<string | null>(null)
const store = useAppStore()

// System status indicator
const systemStatus = computed(() => {
  if (error.value) return { color: 'bg-red-500', text: 'Error' }
  if (store.isProcessing) return { color: 'bg-blue-500', text: 'Processing' }
  return { color: 'bg-green-500', text: 'Ready' }
})

// Global error handler
window.addEventListener('error', (event) => {
  error.value = `Error: ${event.message}`
})

window.addEventListener('unhandledrejection', (event) => {
  error.value = `Promise rejected: ${event.reason}`
})

console.log('🎛️ App component loaded')
</script>

<style scoped>
/* Minimal component styles */
</style>