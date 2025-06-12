<template>
  <div id="app" class="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <h1 class="text-2xl font-bold text-gray-900">
              Engraving Processor Pro
            </h1>
            <span class="px-2 py-1 text-xs font-medium bg-primary-100 text-primary-800 rounded-full">
              AI-Powered
            </span>
          </div>
          <div class="flex items-center space-x-4">
            <span class="text-sm text-gray-500">
              v{{ version }}
            </span>
          </div>
        </div>
      </div>
    </header>

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
    <footer class="bg-white border-t border-gray-200 mt-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="text-center text-sm text-gray-500">
          <p>&copy; 2025 Engraving Processor Pro. Built with Vue 3, TypeScript & WebAssembly.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue'
  import ImageInput from '@/components/ImageInput.vue'
  import PreviewRenderer from '@/components/PreviewRenderer.vue'
  import ProcessingControls from '@/components/ProcessingControls.vue'
  import ProgressDisplay from '@/components/ProgressDisplay.vue'

  const version = ref('1.0.0')
</script>

<style scoped>
  /* Component-specific styles if needed */
</style>