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
    // Include the src directory for server-side builds
    noExternal: [/^src/],
  },
  build: {
    // Ensure src files are included in the build
    commonjsOptions: {
      include: [/src/, /app/, /node_modules/],
    },
    // Copy public files to build output
    copyPublicDir: true,
    // Suppress empty chunk warnings for API routes
    rollupOptions: {
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
    include: ['react', 'react-dom', '@remix-run/react'],
  },
});