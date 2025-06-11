#!/usr/bin/env node

// Quick script to check if WebSocket manager can be loaded

import { fileURLToPath } from 'url';
import path from 'path';
import fs from 'fs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.join(__dirname, '..');

console.log('🔍 Checking WebSocket setup...\n');

const checks = {
  nodeVersion: false,
  serverWebsocket: false,
  appWebsocket: false,
  singleton: false,
  buildExists: false,
};

// Check Node version
const nodeVersion = process.version;
const nodeMajor = parseInt(nodeVersion.split('.')[0].substring(1));
if (nodeMajor >= 22) {
  console.log(`✅ Node.js ${nodeVersion} (OK)`);
  checks.nodeVersion = true;
} else {
  console.log(`❌ Node.js ${nodeVersion} (requires 22.0.0+)`);
}

// Check if server.websocket.js exists
const serverWsPath = path.join(rootDir, 'server.websocket.js');
if (fs.existsSync(serverWsPath)) {
  console.log('✅ server.websocket.js exists');
  checks.serverWebsocket = true;
  
  // Try to import it
  try {
    const { wsManager } = await import(serverWsPath);
    if (wsManager) {
      console.log('   ✅ Can import wsManager from server.websocket.js');
    }
  } catch (e) {
    console.log('   ❌ Failed to import:', e.message);
  }
} else {
  console.log('❌ server.websocket.js not found');
}

// Check app websocket files
const wsServerPath = path.join(rootDir, 'app/services/websocket.server.js');
const wsSingletonPath = path.join(rootDir, 'app/services/websocket.singleton.js');

if (fs.existsSync(wsServerPath)) {
  console.log('✅ app/services/websocket.server.js exists');
  checks.appWebsocket = true;
} else {
  console.log('❌ app/services/websocket.server.js not found');
}

if (fs.existsSync(wsSingletonPath)) {
  console.log('✅ app/services/websocket.singleton.js exists');
  checks.singleton = true;
} else {
  console.log('❌ app/services/websocket.singleton.js not found');
}

// Check if build exists
const buildPath = path.join(rootDir, 'build/server/index.js');
if (fs.existsSync(buildPath)) {
  console.log('✅ Build exists');
  checks.buildExists = true;
} else {
  console.log('❌ Build not found - run "pnpm build" first');
}

// Summary
console.log('\n📊 Summary:');
const passed = Object.values(checks).filter(v => v).length;
const total = Object.values(checks).length;

if (passed === total) {
  console.log(`✅ All checks passed (${passed}/${total})`);
  console.log('\n🚀 Ready to start the server with "pnpm start"');
  process.exit(0);
} else {
  console.log(`⚠️ Some checks failed (${passed}/${total})`);
  console.log('\n📝 Recommendations:');
  
  if (!checks.nodeVersion) {
    console.log('- Upgrade Node.js to version 22 or higher');
  }
  
  if (!checks.serverWebsocket) {
    console.log('- Create server.websocket.js file (see WEBSOCKET-SETUP.md)');
  }
  
  if (!checks.buildExists) {
    console.log('- Run "pnpm build" to create the production build');
  }
  
  process.exit(1);
}