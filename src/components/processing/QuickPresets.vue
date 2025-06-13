<!-- src/components/processing/QuickPresets.vue -->
<template>
  <div class="pt-4 border-t border-slate-200">
    <h3 class="text-sm font-medium text-slate-700 mb-3">Quick Presets</h3>
    <div class="grid grid-cols-2 gap-2">
      <button
        class="btn btn-secondary text-sm"
        :disabled="!hasImage || !systemInitialized"
        @click="$emit('apply-preset', 'document')"
      >
        📄 Document
      </button>
      <button
        class="btn btn-secondary text-sm"
        :disabled="!hasImage || !systemInitialized"
        @click="$emit('apply-preset', 'engraving')"
      >
        🖼️ Engraving
      </button>
      <button
        class="btn btn-secondary text-sm"
        :disabled="!hasImage || !systemInitialized"
        @click="$emit('apply-preset', 'pixel-art')"
      >
        🎮 Pixel Art
      </button>
      <button
        class="btn btn-secondary text-sm"
        :disabled="!hasImage || !systemInitialized"
        @click="$emit('apply-preset', 'noise-clean')"
      >
        ✨ Noise Clean
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  hasImage: boolean
  systemInitialized: boolean
}>()

defineEmits<{
  'apply-preset': [presetName: string]
}>()
</script>