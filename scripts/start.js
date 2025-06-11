#!/usr/bin/env node

// Startup script that ensures everything is ready before starting the server

import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import path from 'path';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.join(__dirname, '..');

console.log('🚀 Starting Engraving Processor Pro...\n');

// Run the check script first
console.log('📋 Running pre-flight checks...');
const checkProcess = spawn('node', ['scripts/check-websocket.js'], {
  cwd: rootDir,
  stdio: 'inherit'
});

checkProcess.on('exit', (code) => {
  if (code === 0) {
    console.log('\n🎯 Starting server...\n');
    
    // Start the actual server
    const serverProcess = spawn('node', ['script-server.js'], {
      cwd: rootDir,
      stdio: 'inherit'
    });
    
    // Handle Ctrl+C
    process.on('SIGINT', () => {
      serverProcess.kill('SIGINT');
    });
    
    serverProcess.on('exit', (code) => {
      process.exit(code);
    });
  } else {
    console.error('\n❌ Pre-flight checks failed. Please fix the issues above and try again.');
    process.exit(1);
  }
});