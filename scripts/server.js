import { createRequestHandler } from "@remix-run/express";
import express from "express";
import { createServer } from "http";
import path from "path";
import { fileURLToPath } from "url";
import { WebSocketServer } from "ws";
import fs from "fs";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.join(__dirname, '..');

const app = express();
const httpServer = createServer(app);

// Health status tracking
const healthStatus = {
  server: false,
  websocket: false,
  wsManager: false,
  startTime: Date.now(),
  wsConnections: 0,
  lastError: null,
  checks: {
    nodeVersion: false,
    buildExists: false,
    portsAvailable: true,
    staticAssets: false,
    memoryUsage: { rss: 0, heap: 0 }
  }
};

// Color codes for console output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m'
};

// Logging helpers
const log = {
  info: (msg) => console.log(`${colors.blue}ℹ️${colors.reset} ${msg}`),
  success: (msg) => console.log(`${colors.green}✅${colors.reset} ${msg}`),
  error: (msg) => console.log(`${colors.red}❌${colors.reset} ${msg}`),
  warn: (msg) => console.log(`${colors.yellow}⚠️${colors.reset} ${msg}`),
  header: (msg) => console.log(`\n${colors.bright}${colors.cyan}${msg}${colors.reset}`)
};

// Startup checks
async function runStartupChecks() {
  log.header("🔍 Running startup checks...");
  
  // Check Node.js version
  const nodeVersion = process.version;
  const nodeMajor = parseInt(nodeVersion.split('.')[0].substring(1));
  if (nodeMajor >= 22) {
    log.success(`Node.js ${nodeVersion} (minimum: 22.0.0)`);
    healthStatus.checks.nodeVersion = true;
  } else {
    log.error(`Node.js ${nodeVersion} is too old (minimum: 22.0.0)`);
    return false;
  }
  
  // Check build exists
  const buildPath = path.join(rootDir, 'build/server/index.js');
  if (fs.existsSync(buildPath)) {
    log.success("Build found");
    healthStatus.checks.buildExists = true;
  } else {
    log.error("Build not found - run 'pnpm build' first");
    return false;
  }
  
  // Check critical directories
  const requiredDirs = ['build/client/assets', 'public', 'src/engine'];
  for (const dir of requiredDirs) {
    if (!fs.existsSync(path.join(rootDir, dir))) {
      log.error(`Missing directory: ${dir}`);
      return false;
    if (!wsManager) {
      try {
        // Approach 2: Try importing the initialization module from the build
        const buildPath = path.join(rootDir, 'build/server/index.js');
        const buildModule = await import(buildPath);
        
        // Look for the initialization function or wsManager in exports
        if (buildModule.__initializeWebSocket) {
          wsManager = buildModule.__wsManager || global.__wsManager;
          if (wsManager) {
            healthStatus.wsManager = true;
            log.success("WebSocket manager found in build exports");
            
            // Initialize it
            if (buildModule.__initializeWebSocket(httpServer)) {
              healthStatus.websocket = true;
              log.success("WebSocket server initialized from build");
            }
          }
        }
      } catch (e) {
        log.info("Could not find WebSocket exports in build");
      }
    }
  }
  
  healthStatus.checks.staticAssets = true;
  return true;
}

// Serve static files
app.use(express.static(path.join(rootDir, "public")));
app.use("/assets", express.static(path.join(rootDir, "build/client/assets")));
app.use("/build", express.static(path.join(rootDir, "build/client")));

// Enhanced health endpoint
app.get("/health", (req, res) => {
  const uptime = Date.now() - healthStatus.startTime;
  const memUsage = process.memoryUsage();
  
  healthStatus.checks.memoryUsage = {
    rss: Math.round(memUsage.rss / 1024 / 1024),
    heap: Math.round(memUsage.heapUsed / 1024 / 1024)
  };
  
  const isHealthy = healthStatus.server && healthStatus.websocket && healthStatus.wsManager;
  
  res.json({
    status: isHealthy ? "healthy" : "unhealthy",
    uptime: Math.round(uptime / 1000),
    services: {
      server: healthStatus.server,
      websocket: healthStatus.websocket,
      wsManager: healthStatus.wsManager,
      wsConnections: healthStatus.wsConnections
    },
    checks: healthStatus.checks,
    issues: !isHealthy ? {
      wsManager: !healthStatus.wsManager ? "WebSocket manager not initialized" : null,
      websocket: !healthStatus.websocket ? "WebSocket server failed to initialize" : null,
      lastError: healthStatus.lastError
    } : null,
    nodeVersion: process.version,
    timestamp: new Date().toISOString()
  });
});

// Initialize services
let wsManager = null;
let healthCheckInterval = null;

async function initializeServices() {
  log.header("🚀 Starting Engraving Processor Pro...");
  
  try {
    log.info("Loading application build...");
    const build = await import(path.join(rootDir, "build/server/index.js"));
    
    // Try multiple approaches to get wsManager
    
    // Approach 1: Check if initialization function is available globally
    if (global.__initializeWebSocket && global.__wsManager) {
      wsManager = global.__wsManager;
      healthStatus.wsManager = true;
      log.success("WebSocket manager found via global");
      
      // Initialize it
      if (global.__initializeWebSocket(httpServer)) {
        healthStatus.websocket = true;
        log.success("WebSocket server initialized via global function");
      }
    } 
    
    // Approach 2: Try the singleton approach
    if (!wsManager) {
      try {
        const { getWsManager } = await import(path.join(rootDir, "build/server/index.js"));
        if (getWsManager) {
          // Wait for dynamic imports to complete
          await new Promise(resolve => setTimeout(resolve, 200));
          wsManager = getWsManager();
          if (wsManager) {
            healthStatus.wsManager = true;
            log.success("WebSocket manager found via singleton");
          }
        }
      } catch (e) {
        log.info("Singleton approach not available");
      }
    }
    
    // Approach 3: Try direct import from build
    if (!wsManager) {
      try {
        // Look for the websocket module in the build
        const buildFiles = fs.readdirSync(path.join(rootDir, 'build/server'));
        const wsFile = buildFiles.find(f => f.includes('websocket'));
        if (wsFile) {
          const wsModule = await import(path.join(rootDir, 'build/server', wsFile));
          wsManager = wsModule.wsManager || wsModule.default;
          if (wsManager) {
            healthStatus.wsManager = true;
            log.success(`WebSocket manager found in ${wsFile}`);
          }
        }
      } catch (e) {
        log.info("Direct build import failed");
      }
    }
    
    if (wsManager) {
      // Initialize WebSocket
      try {
        wsManager.initialize(httpServer);
        healthStatus.websocket = true;
        log.success("WebSocket server initialized");
        
        // Monitor WebSocket connections
        if (wsManager.getConnectionCount) {
          setInterval(() => {
            healthStatus.wsConnections = wsManager.getConnectionCount();
          }, 5000);
        }
      } catch (wsError) {
        log.error(`WebSocket initialization failed: ${wsError.message}`);
        healthStatus.lastError = `WebSocket init: ${wsError.message}`;
      }
    } else {
      log.warn("WebSocket manager not available - real-time preview will not work");
      log.info("This may happen on first run - try restarting after build completes");
    }

    // Skip Remix handler for WebSocket requests
    app.use((req, res, next) => {
      if (req.headers.upgrade === 'websocket' && req.url === '/ws') {
        return;
      }
      next();
    });

    // Remix request handler
    app.all(
      "*",
      createRequestHandler({
        build: build,
        mode: process.env.NODE_ENV,
      })
    );

    healthStatus.server = true;
    log.success("Server initialization complete");
    
  } catch (error) {
    log.error(`Failed to initialize services: ${error.message}`);
    healthStatus.lastError = error.message;
    throw error;
  }
}

// Continuous health monitoring
function startHealthMonitoring() {
  log.header("📊 Starting health monitoring...");
  
  healthCheckInterval = setInterval(() => {
    const memUsage = process.memoryUsage();
    const wsConnections = wsManager ? wsManager.getConnectionCount() : 0;
    
    // Check for issues
    const issues = [];
    if (!healthStatus.websocket) issues.push("WebSocket not initialized");
    if (!healthStatus.wsManager) issues.push("WebSocket manager missing");
    if (memUsage.heapUsed > 500 * 1024 * 1024) issues.push("High memory usage");
    
    if (issues.length > 0) {
      log.warn(`Health check issues: ${issues.join(", ")}`);
    }
    
    // Update health status
    healthStatus.checks.memoryUsage = {
      rss: Math.round(memUsage.rss / 1024 / 1024),
      heap: Math.round(memUsage.heapUsed / 1024 / 1024)
    };
    healthStatus.wsConnections = wsConnections;
  }, 30000); // Check every 30 seconds
}

// Server startup diagnostics
function runDiagnostics() {
  log.header("🔍 Running diagnostics...");
  
  const memUsage = process.memoryUsage();
  
  // System status
  log.info(`Node.js version: ${process.version}`);
  log.info(`Memory usage: RSS ${Math.round(memUsage.rss / 1024 / 1024)}MB, Heap ${Math.round(memUsage.heapUsed / 1024 / 1024)}MB`);
  log.info(`Environment: ${process.env.NODE_ENV || 'development'}`);
  
  // Service status
  if (healthStatus.server) {
    log.success("HTTP server: Ready");
  } else {
    log.error("HTTP server: Not ready");
  }
  
  if (healthStatus.wsManager) {
    log.success("WebSocket manager: Loaded");
  } else {
    log.error("WebSocket manager: Not found");
  }
  
  if (healthStatus.websocket) {
    log.success("WebSocket server: Ready");
  } else {
    log.error("WebSocket server: Not initialized");
  }
  
  // Overall status
  const isReady = healthStatus.server && healthStatus.websocket && healthStatus.wsManager;
  if (isReady) {
    log.header("✨ Server is ready to process images!");
  } else {
    log.header("⚠️ Server started with issues - some features may not work");
    if (!healthStatus.wsManager || !healthStatus.websocket) {
      log.warn("Without WebSocket support, real-time preview updates will not work");
    }
  }
}

// Start server
const port = process.env.PORT || 3000;

async function startServer() {
  try {
    // Run startup checks
    const checksOk = await runStartupChecks();
    if (!checksOk) {
      log.error("Startup checks failed");
      process.exit(1);
    }
    
    // Initialize services
    await initializeServices();
    
    // Start HTTP server
    httpServer.listen(port, () => {
      log.success(`Server listening on http://localhost:${port}`);
      log.info(`Health endpoint: http://localhost:${port}/health`);
      if (healthStatus.websocket) {
        log.info(`WebSocket endpoint: ws://localhost:${port}/ws`);
      }
      
      // Run diagnostics
      runDiagnostics();
      
      // Start health monitoring
      startHealthMonitoring();
    });
  } catch (error) {
    log.error(`Failed to start server: ${error.message}`);
    process.exit(1);
  }
}

// Error handling
process.on('uncaughtException', (error) => {
  log.error(`Uncaught Exception: ${error.message}`);
  healthStatus.lastError = error.message;
});

process.on('unhandledRejection', (reason, promise) => {
  log.error(`Unhandled Rejection: ${reason}`);
  healthStatus.lastError = reason?.message || 'Unhandled rejection';
});

// Graceful shutdown
process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);

function shutdown() {
  log.header("🛑 Shutting down server...");
  
  healthStatus.server = false;
  healthStatus.websocket = false;
  
  if (healthCheckInterval) {
    clearInterval(healthCheckInterval);
  }
  
  httpServer.close(() => {
    log.info("HTTP server closed");
    if (wsManager && wsManager.close) {
      wsManager.close();
      log.info("WebSocket server closed");
    }
    log.info("👋 Goodbye!");
    process.exit(0);
  });
  
  // Force shutdown after 10 seconds
  setTimeout(() => {
    log.error("Forced shutdown after timeout");
    process.exit(1);
  }, 10000);
}

// Start the server
startServer();
