import { vitePlugin as remix } from "@remix-run/dev";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [
    remix({
      future: {
        v3_fetcherPersist: true,
        v3_relativeSplatPath: true,
        v3_throwAbortReason: true,
        v3_lazyRouteDiscovery: true,
        v3_singleFetch: true,
      },
      serverBuildFile: "index.js",
    }),
    tsconfigPaths(),
    tailwindcss(),
  ],
  server: {
    port: 3000,
    hmr: {
      port: 3001,
    },
  },
  publicDir: "public",
  resolve: {
    // Allow importing .js files without extension
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json'],
  },
  ssr: {
    // Include the src directory and websocket modules for server-side builds
    noExternal: [/^src/, 'ws', /websocket/],
  },
  build: {
    // Ensure src files and websocket modules are included in the build
    commonjsOptions: {
      include: [/src/, /app/, /node_modules/],
    },
    // Copy public files to build output
    copyPublicDir: true,
    rollupOptions: {
      output: {
        // Preserve exports
        exports: 'named',
        preserveModules: false,
      },
      // Ensure websocket modules are not tree-shaken
      treeshake: {
        moduleSideEffects: (id) => {
          // Keep websocket modules
          if (id.includes('websocket')) return true;
          // Keep entry files
          if (id.includes('entry.server')) return true;
          return 'no-external';
        },
      },
      // Mark certain modules to ensure they're included
      external: [],
      onwarn(warning, warn) {
        // Suppress "Generated an empty chunk" warnings for API routes
        if (warning.code === 'EMPTY_BUNDLE' && 
            (warning.message.includes('api.') || 
             warning.message.includes('api/') ||
             warning.message.includes('_index'))) {
          return;
        }
        // Suppress source map warnings
        if (warning.code === 'SOURCEMAP_ERROR') {
          return;
        }
        warn(warning);
      },
    },
  },
  // Optimize dependencies
  optimizeDeps: {
    include: ['react', 'react-dom', '@remix-run/react', 'ws'],
    exclude: ['sharp'], // Sharp should not be bundled
  },
  // Define globals for server-side code
  define: {
    'global.__wsManager': 'undefined',
  },
});