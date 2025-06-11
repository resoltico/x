import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.join(__dirname, '..');

console.log('🔧 Running post-build script...');

// Check if websocket modules are in the build
const serverBuildDir = path.join(rootDir, 'build/server');
const serverIndexPath = path.join(serverBuildDir, 'index.js');

if (fs.existsSync(serverIndexPath)) {
  // Read the server index file
  const serverIndex = fs.readFileSync(serverIndexPath, 'utf-8');
  
  // Check if websocket imports are present
  if (!serverIndex.includes('websocket')) {
    console.warn('⚠️ WebSocket modules may not be included in build');
  } else {
    console.log('✅ WebSocket modules found in build');
  }
  
  // Create a module map file to help with runtime imports
  const assetsDir = path.join(serverBuildDir, 'assets');
  if (fs.existsSync(assetsDir)) {
    const files = fs.readdirSync(assetsDir);
    const wsFiles = files.filter(f => f.includes('websocket'));
    
    if (wsFiles.length > 0) {
      console.log(`✅ Found ${wsFiles.length} websocket module(s) in assets:`, wsFiles);
      
      // Create a module map
      const moduleMap = {
        websocketModules: wsFiles.map(f => `/build/server/assets/${f}`),
        generated: new Date().toISOString()
      };
      
      fs.writeFileSync(
        path.join(serverBuildDir, 'module-map.json'),
        JSON.stringify(moduleMap, null, 2)
      );
      console.log('✅ Created module map');
    }
  }
} else {
  console.error('❌ Server build not found');
  process.exit(1);
}

console.log('✅ Post-build script completed');