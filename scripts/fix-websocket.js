#!/usr/bin/env node

// Script to fix common WebSocket setup issues

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.join(__dirname, '..');

console.log('🔧 Fixing WebSocket setup...\n');

// Check and create server.websocket.js if missing
const serverWsPath = path.join(rootDir, 'server.websocket.js');
if (!fs.existsSync(serverWsPath)) {
  console.log('📝 Creating server.websocket.js...');
  
  const content = `// Direct websocket manager export for server initialization
// This file provides a direct way to access the websocket manager
// without going through the Remix build system

import { wsManager } from './app/services/websocket.server.js';

// Make it available globally for the server script
if (typeof global !== 'undefined') {
  global.__wsManager = wsManager;
}

export { wsManager };
export default wsManager;
`;
  
  fs.writeFileSync(serverWsPath, content);
  console.log('✅ Created server.websocket.js');
} else {
  console.log('✅ server.websocket.js already exists');
}

// Check if scripts directory exists
const scriptsDir = path.join(rootDir, 'scripts');
if (!fs.existsSync(scriptsDir)) {
  console.log('📁 Creating scripts directory...');
  fs.mkdirSync(scriptsDir);
  console.log('✅ Created scripts directory');
}

// Ensure all TypeScript files have been compiled
console.log('\n🔍 Checking TypeScript compilation...');
const tsFiles = [
  'app/services/websocket.server.ts',
  'app/services/websocket.singleton.ts',
  'app/server-init.ts'
];

let needsCompile = false;
for (const tsFile of tsFiles) {
  const tsPath = path.join(rootDir, tsFile);
  const jsPath = tsPath.replace('.ts', '.js');
  
  if (fs.existsSync(tsPath) && !fs.existsSync(jsPath)) {
    console.log(`⚠️ ${tsFile} needs compilation`);
    needsCompile = true;
  }
}

if (needsCompile) {
  console.log('\n📦 Running TypeScript compilation...');
  console.log('Run "pnpm build" to compile TypeScript files');
} else {
  console.log('✅ All TypeScript files are compiled');
}

console.log('\n✨ WebSocket setup fixes applied!');
console.log('\nNext steps:');
console.log('1. Run "pnpm build" to build the application');
console.log('2. Run "pnpm check" to verify the setup');
console.log('3. Run "pnpm start" to start the server');