import { createRequestHandler } from "@remix-run/express";
import express from "express";
import { createServer } from "http";
import path from "path";
import { fileURLToPath } from "url";
import { WebSocketServer } from "ws";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const app = express();
const httpServer = createServer(app);

// Health check storage
const healthStatus = {
  server: false,
  websocket: false,
  startTime: Date.now(),
  wsConnections: 0,
  lastError: null
};

// Serve static files from public directory
app.use(express.static("public"));

// IMPORTANT: Serve built client assets
app.use("/assets", express.static(path.join(__dirname, "build/client/assets")));
app.use("/build", express.static(path.join(__dirname, "build/client")));

// Health check endpoint
app.get("/health", (req, res) => {
  const uptime = Date.now() - healthStatus.startTime;
  const status = healthStatus.server && healthStatus.websocket ? "healthy" : "unhealthy";
  
  res.json({
    status,
    uptime,
    services: {
      server: healthStatus.server,
      websocket: healthStatus.websocket,
      wsConnections: healthStatus.wsConnections
    },
    lastError: healthStatus.lastError
  });
});

// Import the build and initialize WebSocket
let wsManager = null;

async function initializeServices() {
  console.log("🚀 Starting Engraving Processor Pro...");
  
  try {
    // Import the build
    console.log("📦 Loading application build...");
    const build = await import("./build/server/index.js");
    
    // Initialize WebSocket server
    console.log("🔌 Initializing WebSocket server...");
    if (build.wsManager) {
      wsManager = build.wsManager;
      wsManager.initialize(httpServer);
      healthStatus.websocket = true;
      console.log("✅ WebSocket server initialized successfully");
      
      // Monitor WebSocket connections
      setInterval(() => {
        if (wsManager && wsManager.getConnectionCount) {
          healthStatus.wsConnections = wsManager.getConnectionCount();
        }
      }, 5000);
    } else {
      console.error("❌ WebSocket manager not found in build");
      healthStatus.lastError = "WebSocket manager not found in build";
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

    healthStatus.server = true;
    console.log("✅ Server initialization complete");
    
  } catch (error) {
    console.error("❌ Failed to initialize services:", error);
    healthStatus.lastError = error.message;
    throw error;
  }
}

// Start server
const port = process.env.PORT || 3000;

async function startServer() {
  try {
    await initializeServices();
    
    httpServer.listen(port, () => {
      console.log(`✅ Server listening on http://localhost:${port}`);
      console.log(`📊 Health check available at http://localhost:${port}/health`);
      console.log(`🔌 WebSocket endpoint: ws://localhost:${port}/ws`);
      
      // Run startup diagnostics
      runStartupDiagnostics();
    });
  } catch (error) {
    console.error("❌ Failed to start server:", error);
    process.exit(1);
  }
}

// Startup diagnostics
function runStartupDiagnostics() {
  console.log("\n🔍 Running startup diagnostics...");
  
  // Check WebSocket
  if (healthStatus.websocket) {
    console.log("✅ WebSocket: Ready");
  } else {
    console.log("❌ WebSocket: Not initialized");
  }
  
  // Check memory usage
  const memUsage = process.memoryUsage();
  console.log(`💾 Memory usage: RSS ${Math.round(memUsage.rss / 1024 / 1024)}MB, Heap ${Math.round(memUsage.heapUsed / 1024 / 1024)}MB`);
  
  // Check Node.js version
  console.log(`🟢 Node.js version: ${process.version}`);
  
  console.log("\n✨ Server is ready to process images!\n");
}

// Error monitoring
process.on('uncaughtException', (error) => {
  console.error('❌ Uncaught Exception:', error);
  healthStatus.lastError = error.message;
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('❌ Unhandled Rejection at:', promise, 'reason:', reason);
  healthStatus.lastError = reason?.message || 'Unhandled rejection';
});

// Graceful shutdown
process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);

function shutdown() {
  console.log("\n🛑 Shutting down server...");
  
  healthStatus.server = false;
  healthStatus.websocket = false;
  
  httpServer.close(() => {
    console.log("📴 HTTP server closed");
    if (wsManager && wsManager.close) {
      wsManager.close();
      console.log("📴 WebSocket server closed");
    }
    console.log("👋 Goodbye!");
    process.exit(0);
  });
  
  // Force shutdown after 10 seconds
  setTimeout(() => {
    console.error("❌ Forced shutdown after timeout");
    process.exit(1);
  }, 10000);
}

// Start the server
startServer();