#!/usr/bin/env node

// Main startup script that runs all checks and starts the server

import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import path from 'path';
import fs from 'fs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.join(__dirname, '..');

// Color codes
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m'
};

const log = {
  info: (msg) => console.log(`${colors.blue}ℹ️${colors.reset} ${msg}`),
  success: (msg) => console.log(`${colors.green}✅${colors.reset} ${msg}`),
  error: (msg) => console.log(`${colors.red}❌${colors.reset} ${msg}`),
  warn: (msg) => console.log(`${colors.yellow}⚠️${colors.reset} ${msg}`),
  header: (msg) => console.log(`\n${colors.bright}${colors.cyan}${msg}${colors.reset}\n`)
};

// Run all pre-flight checks
async function runAllChecks() {
  log.header('🚀 Starting Engraving Processor Pro...');
  
  const checks = {
    nodeVersion: false,
    buildExists: false,
    dependencies: false,
    websocketSetup: false
  };

  // Check Node version
  log.info('Checking Node.js version...');
  const nodeVersion = process.version;
  const nodeMajor = parseInt(nodeVersion.split('.')[0].substring(1));
  if (nodeMajor >= 22) {
    log.success(`Node.js ${nodeVersion} ✓`);
    checks.nodeVersion = true;
  } else {
    log.error(`Node.js ${nodeVersion} - requires 22.0.0 or higher`);
  }

  // Check if build exists
  log.info('Checking build...');
  const buildPath = path.join(rootDir, 'build/server/index.js');
  if (fs.existsSync(buildPath)) {
    log.success('Build found ✓');
    checks.buildExists = true;
  } else {
    log.error('Build not found - run "pnpm build" first');
  }

  // Check dependencies
  log.info('Checking dependencies...');
  const nodeModulesPath = path.join(rootDir, 'node_modules');
  if (fs.existsSync(nodeModulesPath)) {
    log.success('Dependencies installed ✓');
    checks.dependencies = true;
  } else {
    log.error('Dependencies not installed - run "pnpm install" first');
  }

  // Check WebSocket setup
  log.info('Checking WebSocket configuration...');
  const wsServerPath = path.join(rootDir, 'app/services/websocket.server.js');
  const wsSingletonPath = path.join(rootDir, 'app/services/websocket.singleton.server.js');
  
  if (fs.existsSync(wsServerPath) || fs.existsSync(wsSingletonPath)) {
    log.success('WebSocket files present ✓');
    checks.websocketSetup = true;
  } else {
    log.warn('WebSocket files may need compilation');
  }

  // Summary
  const passed = Object.values(checks).filter(v => v).length;
  const total = Object.values(checks).length;
  
  if (passed === total) {
    log.header(`All checks passed (${passed}/${total}) ✨`);
    return true;
  } else {
    log.header(`Some checks failed (${passed}/${total}) ⚠️`);
    
    if (!checks.nodeVersion) {
      log.info('→ Install Node.js 22 or higher');
    }
    if (!checks.dependencies) {
      log.info('→ Run: pnpm install');
    }
    if (!checks.buildExists) {
      log.info('→ Run: pnpm build');
    }
    
    return passed >= 2; // Allow starting with minor issues
  }
}

// Main function
async function main() {
  try {
    const checksOk = await runAllChecks();
    
    if (!checksOk) {
      log.error('Pre-flight checks failed. Please fix the issues above.');
      process.exit(1);
    }

    log.header('Starting server...');
    
    // Start the server
    const serverProcess = spawn('node', ['scripts/server.js'], {
      cwd: rootDir,
      stdio: 'inherit',
      env: { ...process.env }
    });
    
    // Handle signals
    process.on('SIGINT', () => {
      serverProcess.kill('SIGINT');
    });
    
    process.on('SIGTERM', () => {
      serverProcess.kill('SIGTERM');
    });
    
    serverProcess.on('exit', (code) => {
      process.exit(code || 0);
    });
    
  } catch (error) {
    log.error(`Startup failed: ${error.message}`);
    process.exit(1);
  }
}

// Run the startup sequence
main();