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
      },
    }),
    tsconfigPaths(),
    tailwindcss(),
  ],
  server: {
    port: 3000,
  },
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
  },
});