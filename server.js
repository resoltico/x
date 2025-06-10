import { createRequestHandler } from "@remix-run/express";
import express from "express";
import { createServer } from "http";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const app = express();
const httpServer = createServer(app);

// Serve static files from public directory
app.use(express.static("public"));

// IMPORTANT: Serve built client assets
app.use("/assets", express.static(path.join(__dirname, "build/client/assets")));

// Import the build
const build = await import("./build/server/index.js");

// Initialize WebSocket server if wsManager is available
if (build.wsManager) {
  build.wsManager.initialize(httpServer);
}

// Remix request handler
app.all(
  "*",
  createRequestHandler({
    build: build,
  })
);

const port = process.env.PORT || 3000;

httpServer.listen(port, () => {
  console.log(`Server listening on http://localhost:${port}`);
});

process.on("SIGTERM", () => {
  httpServer.close(() => {
    if (build.wsManager) {
      build.wsManager.close();
    }
  });
});