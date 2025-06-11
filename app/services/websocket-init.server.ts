// WebSocket initialization module - SERVER ONLY
// This module handles the WebSocket manager initialization

import { wsManager } from './websocket.server';

// Export the manager
export { wsManager };

// Initialize function for the server to call
export function initializeWebSocket(httpServer: any) {
  if (!wsManager) {
    console.error('WebSocket manager not available');
    return false;
  }
  
  try {
    wsManager.initialize(httpServer);
    console.log('✅ WebSocket initialized via websocket-init.server.ts');
    return true;
  } catch (error) {
    console.error('❌ WebSocket initialization failed:', error);
    return false;
  }
}

// Make manager available globally
if (typeof global !== 'undefined') {
  (global as any).__wsManager = wsManager;
  (global as any).__initializeWebSocket = initializeWebSocket;
}
