import { createRequestHandler } from "@remix-run/express";
import express from "express";
import { createServer } from "http";

const app = express();
const httpServer = createServer(app);

app.use(express.static("public"));

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