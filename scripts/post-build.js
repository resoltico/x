import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.join(__dirname, '..');

console.log('🔧 Running post-build script...');

// Check if websocket modules are in the build
const serverBuildDir = path.join(rootDir, 'build/server');
const serverIndexPath = path.join(serverBuildDir, 'index.js');

if (!fs.existsSync(serverIndexPath)) {
  console.error('❌ Server build not found');
  process.exit(1);
}

// Read the server index file
const serverIndex = fs.readFileSync(serverIndexPath, 'utf-8');

// Check if websocket imports are present
if (!serverIndex.includes('websocket')) {
  console.warn('⚠️ WebSocket modules may not be included in build');
  console.log('💡 Tip: Ensure websocket.singleton.server.ts is imported in entry.server.tsx');
} else {
  console.log('✅ WebSocket modules found in build');
}

// Create a simple server-side websocket export if needed
const wsExportPath = path.join(rootDir, 'build/websocket-export.js');
const wsExportContent = `// WebSocket manager export helper
import { getWsManager } from './server/index.js';
export { getWsManager };
export default getWsManager;
`;

fs.writeFileSync(wsExportPath, wsExportContent);
console.log('✅ Created WebSocket export helper');

console.log('✅ Post-build script completed');