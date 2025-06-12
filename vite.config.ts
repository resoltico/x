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
    },
    fs: {
      // Allow serving files from one level up to the project root
      allow: ['..']
    }
  },
  build: {
    target: 'esnext',
    chunkSizeWarningLimit: 1000,
    sourcemap: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html')
      },
      output: {
        manualChunks: {
          'vue-vendor': ['vue', 'pinia'],
          'image-processing': ['wasm-vips']
        }
      },
      onwarn(warning, warn) {
        // Skip eval warnings from wasm-vips as they are intentional and safe
        if (warning.code === 'EVAL' && warning.id?.includes('wasm-vips')) {
          return
        }
        // Skip TypeScript issues that are handled by the TypeScript compiler
        if (warning.code === 'PLUGIN_WARNING' && warning.plugin === 'typescript') {
          return
        }
        warn(warning)
      }
    }
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
    ],
    rollupOptions: {
      output: {
        // Ensure workers are built with proper ES module format
        format: 'es',
        entryFileNames: 'assets/[name]-[hash].js'
      }
    }
  },
  assetsInclude: ['**/*.wasm'],
  esbuild: {
    // Enhanced TypeScript compilation options
    target: 'esnext'
  },
  define: {
    // Define global constants for better error handling
    __DEV__: JSON.stringify(process.env.NODE_ENV !== 'production'),
    __VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0')
  }
})