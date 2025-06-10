import { createRequestHandler } from "@remix-run/express";
import express from "express";
import { createServer } from "http";
import { wsManager } from "./build/server/index.js";

const app = express();
const httpServer = createServer(app);

// Initialize WebSocket server
wsManager.initialize(httpServer);

app.use(express.static("public"));

// Remix request handler
app.all(
  "*",
  createRequestHandler({
    build: await import("./build/server/index.js"),
  })
);

const port = process.env.PORT || 3000;

httpServer.listen(port, () => {
  console.log(`Server listening on http://localhost:${port}`);
});

process.on("SIGTERM", () => {
  httpServer.close(() => {
    wsManager.close();
  });
});