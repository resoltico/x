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
app.use("/build", express.static(path.join(__dirname, "build/client")));

// Import the build
const build = await import("./build/server/index.js");

// Initialize WebSocket server if wsManager is available
if (build.wsManager) {
  build.wsManager.initialize(httpServer);
  console.log('WebSocket server initialized');
}

// IMPORTANT: Skip Remix handler for WebSocket requests
app.use((req, res, next) => {
  // Check if this is a WebSocket upgrade request
  if (req.headers.upgrade === 'websocket' && req.url === '/ws') {
    // Let the WebSocket server handle this - don't pass to Remix
    return;
  }
  next();
});

// Remix request handler - only for non-WebSocket requests
app.all(
  "*",
  createRequestHandler({
    build: build,
    mode: process.env.NODE_ENV,
  })
);

const port = process.env.PORT || 3000;

httpServer.listen(port, () => {
  console.log(`Server listening on http://localhost:${port}`);
});

// Graceful shutdown
process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);

function shutdown() {
  console.log("Shutting down server...");
  httpServer.close(() => {
    if (build.wsManager) {
      build.wsManager.close();
    }
    process.exit(0);
  });
}