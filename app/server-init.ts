// Server initialization hook
// This file is imported by app/root.tsx to ensure websocket manager is available

import { getWsManager } from './services/websocket.singleton';

// Initialize websocket manager on server startup
if (typeof window === 'undefined') {
  // We're on the server
  const wsManager = getWsManager();
  
  // Ensure it's available globally
  if (typeof global !== 'undefined') {
    (global as any).__wsManager = wsManager;
    console.log('✅ WebSocket manager registered globally');
  }
}