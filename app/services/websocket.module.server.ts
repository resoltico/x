/**
 * Consolidated WebSocket module - SERVER ONLY
 * This file combines all WebSocket functionality into a single server-only module
 */

import { wsManager } from './websocket.server';

// Initialize function for the server
export function initializeWebSocket(httpServer: any): boolean {
  if (!wsManager) {
    console.error('WebSocket manager not available');
    return false;
  }
  
  try {
    wsManager.initialize(httpServer);
    console.log('✅ WebSocket server initialized');
    return true;
  } catch (error) {
    console.error('❌ WebSocket initialization failed:', error);
    return false;
  }
}

// Getter function for the manager
export function getWsManager() {
  return wsManager;
}

// Export everything needed
export { wsManager };

// Make available globally for the server script
if (typeof global !== 'undefined') {
  (global as any).__wsManager = wsManager;
  (global as any).__initializeWebSocket = initializeWebSocket;
  (global as any).__getWsManager = getWsManager;
}

// Default export for convenience
export default {
  wsManager,
  initializeWebSocket,
  getWsManager
};
