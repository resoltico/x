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
      // Ignore server-only files in client build
      ignoredRouteFiles: ["**/*.server.{js,ts,jsx,tsx}"],
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
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json'],
  },
  ssr: {
    // Don't bundle these for SSR
    noExternal: [/^src/],
    // External dependencies that should not be bundled
    external: ['sharp'],
  },
  build: {
    // Ensure proper output
    commonjsOptions: {
      include: [/src/, /app/, /node_modules/],
    },
    rollupOptions: {
      output: {
        preserveModules: false,
        format: 'es',
      },
      external: [
        'sharp',
        'bufferutil',
        'utf-8-validate'
      ],
      onwarn(warning, warn) {
        // Suppress certain warnings
        if (warning.code === 'EMPTY_BUNDLE' && 
            (warning.message?.includes('api.') || 
             warning.message?.includes('api/') ||
             warning.message?.includes('_index'))) {
          return;
        }
        if (warning.code === 'SOURCEMAP_ERROR') {
          return;
        }
        warn(warning);
      },
    },
  },
  optimizeDeps: {
    include: ['react', 'react-dom', '@remix-run/react', 'ws'],
    exclude: ['sharp'],
  },
});