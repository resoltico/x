import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import wasm from 'vite-plugin-wasm'
import topLevelAwait from 'vite-plugin-top-level-await'
import { resolve } from 'path'

export default defineConfig({
  css: {
    transformer: 'lightningcss',
    lightningcss: {
      // Enable all Lightning CSS features
      drafts: {
        customMedia: true,
      },
      nonStandard: {
        deepSelectorCombinator: true,
      },
    },
  },
  plugins: [
    vue(),
    tailwindcss(),
    wasm(),
    topLevelAwait()
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@/components': resolve(__dirname, 'src/components'),
      '@/modules': resolve(__dirname, 'src/modules'),
      '@/stores': resolve(__dirname, 'src/stores'),
      '@/utils': resolve(__dirname, 'src/utils'),
      '@/types': resolve(__dirname, 'src/types'),
      '@/workers': resolve(__dirname, 'src/workers')
    }
  },
  server: {
    port: 3000,
    open: true,
    headers: {
      'Cross-Origin-Embedder-Policy': 'require-corp',
      'Cross-Origin-Opener-Policy': 'same-origin',
    }
  },
  build: {
    target: 'esnext',
    rollupOptions: {
      output: {
        manualChunks: {
          'vue-vendor': ['vue', 'pinia'],
          'image-processing': ['wasm-vips']
        }
      },
      // Handle eval warnings from wasm-vips
      onwarn(warning, warn) {
        // Skip eval warnings from wasm-vips as they are intentional and safe in this context
        if (warning.code === 'EVAL' && warning.id?.includes('wasm-vips')) {
          return
        }
        warn(warning)
      }
    },
    // Increase chunk size warning limit for WASM files
    chunkSizeWarningLimit: 1000,
    // Enable source maps for debugging
    sourcemap: true
  },
  optimizeDeps: {
    exclude: ['wasm-vips'],
    include: ['vue', 'pinia']
  },
  worker: {
    format: 'es',
    plugins: () => [
      wasm(),
      topLevelAwait()
    ]
  },
  // Ensure proper WASM support
  assetsInclude: ['**/*.wasm']
})