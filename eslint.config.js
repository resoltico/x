// eslint.config.js
import js from '@eslint/js'
import tseslint from '@typescript-eslint/eslint-plugin'
import tsparser from '@typescript-eslint/parser'
import vue from 'eslint-plugin-vue'
import vueParser from 'vue-eslint-parser'

export default [
  // Base JavaScript config
  js.configs.recommended,
  
  // TypeScript files
  {
    files: ['**/*.ts', '**/*.tsx'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module'
      },
      globals: {
        // Browser globals
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        performance: 'readonly',
        console: 'readonly',
        setTimeout: 'readonly',
        setInterval: 'readonly',
        clearTimeout: 'readonly',
        clearInterval: 'readonly',
        URL: 'readonly',
        Blob: 'readonly',
        File: 'readonly',
        FileReader: 'readonly',
        FileList: 'readonly',
        Image: 'readonly',
        HTMLElement: 'readonly',
        HTMLCanvasElement: 'readonly',
        HTMLInputElement: 'readonly',
        HTMLImageElement: 'readonly',
        SVGImageElement: 'readonly',
        HTMLVideoElement: 'readonly',
        CanvasRenderingContext2D: 'readonly',
        OffscreenCanvas: 'readonly',
        OffscreenCanvasRenderingContext2D: 'readonly',
        ImageData: 'readonly',
        ImageBitmap: 'readonly',
        Worker: 'readonly',
        MessageEvent: 'readonly',
        ErrorEvent: 'readonly',
        PromiseRejectionEvent: 'readonly',
        Event: 'readonly',
        DragEvent: 'readonly',
        WheelEvent: 'readonly',
        MouseEvent: 'readonly',
        TouchEvent: 'readonly',
        KeyboardEvent: 'readonly',
        Node: 'readonly',
        EventTarget: 'readonly',
        Transferable: 'readonly',
        ReadableStream: 'readonly',
        MediaSource: 'readonly',
        MessagePort: 'readonly',
        MessageEventSource: 'readonly',
        WindowProxy: 'readonly',
        ServiceWorker: 'readonly',
        createImageBitmap: 'readonly',
        // Browser API types
        ImageBitmapSource: 'readonly',
        ImageBitmapOptions: 'readonly',
        SVGAnimatedLength: 'readonly',
        // Node.js globals for config files
        __dirname: 'readonly',
        process: 'readonly'
      }
    },
    plugins: {
      '@typescript-eslint': tseslint
    },
    rules: {
      ...tseslint.configs.recommended.rules,
      '@typescript-eslint/no-unused-vars': ['error', { 
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_' 
      }],
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-non-null-assertion': 'off',
      '@typescript-eslint/triple-slash-reference': 'off',
      'no-case-declarations': 'off'
    }
  },
  
  // Vue files
  {
    files: ['**/*.vue'],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tsparser,
        ecmaVersion: 'latest',
        sourceType: 'module',
        extraFileExtensions: ['.vue']
      },
      globals: {
        // Browser globals
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        performance: 'readonly',
        console: 'readonly',
        setTimeout: 'readonly',
        setInterval: 'readonly',
        clearTimeout: 'readonly',
        clearInterval: 'readonly',
        URL: 'readonly',
        Blob: 'readonly',
        File: 'readonly',
        FileReader: 'readonly',
        FileList: 'readonly',
        Image: 'readonly',
        HTMLElement: 'readonly',
        HTMLCanvasElement: 'readonly',
        HTMLInputElement: 'readonly',
        CanvasRenderingContext2D: 'readonly',
        OffscreenCanvas: 'readonly',
        OffscreenCanvasRenderingContext2D: 'readonly',
        ImageData: 'readonly',
        Worker: 'readonly',
        MessageEvent: 'readonly',
        ErrorEvent: 'readonly',
        PromiseRejectionEvent: 'readonly',
        Event: 'readonly',
        DragEvent: 'readonly',
        WheelEvent: 'readonly',
        MouseEvent: 'readonly',
        TouchEvent: 'readonly',
        KeyboardEvent: 'readonly',
        Node: 'readonly',
        Transferable: 'readonly'
      }
    },
    plugins: {
      vue,
      '@typescript-eslint': tseslint
    },
    rules: {
      ...vue.configs['vue3-recommended'].rules,
      ...tseslint.configs.recommended.rules,
      'vue/multi-word-component-names': 'off',
      'vue/no-unused-vars': 'error',
      'vue/attributes-order': ['error', {
        'order': [
          'DEFINITION',
          'LIST_RENDERING',
          'CONDITIONALS',
          'RENDER_MODIFIERS',
          'GLOBAL',
          ['UNIQUE', 'SLOT'],
          'TWO_WAY_BINDING',
          'OTHER_DIRECTIVES',
          'OTHER_ATTR',
          'EVENTS',
          'CONTENT'
        ],
        'alphabetical': false
      }],
      '@typescript-eslint/no-unused-vars': ['error', { 
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_' 
      }],
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-non-null-assertion': 'off',
      'no-case-declarations': 'off'
    }
  },

  // Worker files - Enhanced configuration with correct globals
  {
    files: ['src/workers/**/*.ts'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module'
      },
      globals: {
        // Worker globals
        self: 'readonly',
        importScripts: 'readonly',
        WorkerGlobalScope: 'readonly',
        DedicatedWorkerGlobalScope: 'readonly',
        MessageEvent: 'readonly',
        ErrorEvent: 'readonly',
        PromiseRejectionEvent: 'readonly',
        OffscreenCanvas: 'readonly',
        OffscreenCanvasRenderingContext2D: 'readonly',
        ImageData: 'readonly',
        ImageBitmap: 'readonly',
        console: 'readonly',
        setTimeout: 'readonly',
        setInterval: 'readonly',
        clearTimeout: 'readonly',
        clearInterval: 'readonly',
        URL: 'readonly',
        Blob: 'readonly',
        Image: 'readonly',
        Map: 'readonly',
        Set: 'readonly',
        Array: 'readonly',
        Object: 'readonly',
        Promise: 'readonly',
        Error: 'readonly',
        Math: 'readonly',
        Number: 'readonly',
        String: 'readonly',
        Date: 'readonly',
        Uint8ClampedArray: 'readonly',
        ArrayBuffer: 'readonly',
        Event: 'readonly',
        createImageBitmap: 'readonly'
      }
    },
    plugins: {
      '@typescript-eslint': tseslint
    },
    rules: {
      ...tseslint.configs.recommended.rules,
      '@typescript-eslint/no-unused-vars': ['error', { 
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_' 
      }],
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/triple-slash-reference': 'off',
      'no-case-declarations': 'off'
    }
  },
  
  // Test files
  {
    files: ['src/test/**/*', '**/*.test.*', '**/*.spec.*'],
    languageOptions: {
      globals: {
        vi: 'readonly',
        describe: 'readonly',
        it: 'readonly',
        test: 'readonly',
        expect: 'readonly',
        beforeAll: 'readonly',
        afterAll: 'readonly',
        beforeEach: 'readonly',
        afterEach: 'readonly',
        // Browser globals for tests
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        performance: 'readonly',
        Performance: 'readonly',
        console: 'readonly',
        setTimeout: 'readonly',
        setInterval: 'readonly',
        clearTimeout: 'readonly',
        clearInterval: 'readonly',
        URL: 'readonly',
        Blob: 'readonly',
        File: 'readonly',
        FileReader: 'readonly',
        FileList: 'readonly',
        Image: 'readonly',
        HTMLElement: 'readonly',
        HTMLCanvasElement: 'readonly',
        HTMLInputElement: 'readonly',
        CanvasRenderingContext2D: 'readonly',
        OffscreenCanvas: 'readonly',
        OffscreenCanvasRenderingContext2D: 'readonly',
        ImageData: 'readonly',
        ImageBitmap: 'readonly',
        Worker: 'readonly',
        WorkerOptions: 'readonly',
        MessageEvent: 'readonly',
        ErrorEvent: 'readonly',
        PromiseRejectionEvent: 'readonly',
        Event: 'readonly',
        DragEvent: 'readonly',
        WheelEvent: 'readonly',
        MouseEvent: 'readonly',
        TouchEvent: 'readonly',
        KeyboardEvent: 'readonly',
        Node: 'readonly',
        Transferable: 'readonly',
        globalThis: 'writable'
      }
    }
  },

  // Type definition files - Special handling
  {
    files: ['src/types/**/*.d.ts'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module'
      },
      globals: {
        // All browser and worker globals for type definitions
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        performance: 'readonly',
        console: 'readonly',
        setTimeout: 'readonly',
        setInterval: 'readonly',
        clearTimeout: 'readonly',
        clearInterval: 'readonly',
        URL: 'readonly',
        Blob: 'readonly',
        File: 'readonly',
        FileReader: 'readonly',
        FileList: 'readonly',
        Image: 'readonly',
        HTMLElement: 'readonly',
        HTMLCanvasElement: 'readonly',
        HTMLInputElement: 'readonly',
        HTMLImageElement: 'readonly',
        SVGImageElement: 'readonly',
        HTMLVideoElement: 'readonly',
        CanvasRenderingContext2D: 'readonly',
        OffscreenCanvas: 'readonly',
        OffscreenCanvasRenderingContext2D: 'readonly',
        ImageData: 'readonly',
        ImageBitmap: 'readonly',
        ImageBitmapSource: 'readonly',
        ImageBitmapOptions: 'readonly',
        Worker: 'readonly',
        WorkerGlobalScope: 'readonly',
        DedicatedWorkerGlobalScope: 'readonly',
        MessageEvent: 'readonly',
        ErrorEvent: 'readonly',
        PromiseRejectionEvent: 'readonly',
        Event: 'readonly',
        DragEvent: 'readonly',
        WheelEvent: 'readonly',
        MouseEvent: 'readonly',
        TouchEvent: 'readonly',
        KeyboardEvent: 'readonly',
        Node: 'readonly',
        EventTarget: 'readonly',
        Transferable: 'readonly',
        ReadableStream: 'readonly',
        MediaSource: 'readonly',
        MessagePort: 'readonly',
        MessageEventSource: 'readonly',
        WindowProxy: 'readonly',
        ServiceWorker: 'readonly',
        SVGAnimatedLength: 'readonly',
        createImageBitmap: 'readonly',
        globalThis: 'readonly'
      }
    },
    plugins: {
      '@typescript-eslint': tseslint
    },
    rules: {
      ...tseslint.configs.recommended.rules,
      '@typescript-eslint/no-unused-vars': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-non-null-assertion': 'off',
      '@typescript-eslint/triple-slash-reference': 'off',
      'no-undef': 'off', // Disable no-undef for type definition files
      'no-case-declarations': 'off'
    }
  },

  // Config files (Node.js environment)
  {
    files: ['*.config.*', 'eslint.config.js'],
    languageOptions: {
      globals: {
        __dirname: 'readonly',
        process: 'readonly',
        console: 'readonly'
      }
    }
  },
  
  // Global settings
  {
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module'
    },
    rules: {
      'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'no-undef': 'error'
    }
  },
  
  // Ignore patterns
  {
    ignores: [
      'dist/**',
      'node_modules/**',
      '*.d.ts',
      'coverage/**'
    ]
  }
]