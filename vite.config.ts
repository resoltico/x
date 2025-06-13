import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import wasm from 'vite-plugin-wasm'
import topLevelAwait from 'vite-plugin-top-level-await'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    wasm(),
    topLevelAwait()
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
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
  preview: {
    port: 4173,
    open: true,
    headers: {
      'Cross-Origin-Embedder-Policy': 'require-corp',
      'Cross-Origin-Opener-Policy': 'same-origin',
    }
  },
  build: {
    target: 'esnext',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          'vue-vendor': ['vue', 'pinia']
        },
        // Ensure worker files are accessible at predictable paths
        assetFileNames: (assetInfo) => {
          if (assetInfo.name?.includes('worker') || assetInfo.name?.includes('Worker')) {
            return 'workers/[name]-[hash][extname]'
          }
          return 'assets/[name]-[hash][extname]'
        },
        chunkFileNames: (chunkInfo) => {
          if (chunkInfo.name?.includes('worker') || chunkInfo.facadeModuleId?.includes('worker')) {
            return 'workers/[name]-[hash].js'
          }
          return 'assets/[name]-[hash].js'
        }
      }
    }
  },
  worker: {
    format: 'es',
    rollupOptions: {
      output: {
        entryFileNames: 'workers/[name]-[hash].js',
        assetFileNames: 'workers/[name]-[hash][extname]'
      }
    }
  },
  optimizeDeps: {
    exclude: ['wasm-vips']
  },
  define: {
    __DEV__: JSON.stringify(process.env.NODE_ENV !== 'production'),
    __VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0')
  },
  // Enhanced logging for debugging
  logLevel: 'info',
  clearScreen: false
})