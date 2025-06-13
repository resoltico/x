import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    // Exclude worker files from main test suite as they need special handling
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      '**/cypress/**',
      '**/.{idea,git,cache,output,temp}/**',
      '**/workers/**/*.ts'
    ],
    // Add better error handling and timeout
    testTimeout: 15000,
    hookTimeout: 15000,
    // Enable better stack traces
    include: ['src/**/*.{test,spec}.{js,ts,jsx,tsx}'],
    // Coverage configuration
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test/',
        'src/workers/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/coverage/**'
      ]
    },
    // Suppress expected console output during tests
    onConsoleLog: (log: string) => {
      // Suppress expected logs from tests
      if (log.includes('Mock Worker created for:') ||
          log.includes('Test environment detected') ||
          log.includes('Non-browser environment detected') ||
          log.includes('Falling back to JavaScript-based processing') ||
          log.includes('Processing with Canvas API fallback') ||
          log.includes('OffscreenCanvas not available')) {
        return false
      }
      return true
    },
    // Add pool options for better performance
    pool: 'forks',
    poolOptions: {
      forks: {
        singleFork: true
      }
    },
    // Increase max concurrency for better performance
    maxConcurrency: 1
  },
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
  esbuild: {
    target: 'node14'
  },
  define: {
    __DEV__: true,
    __VERSION__: '"1.0.0-test"'
  }
})