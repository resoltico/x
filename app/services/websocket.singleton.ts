// Global singleton for WebSocket manager
import { wsManager } from './websocket.server';

// Store wsManager in a global variable to prevent tree-shaking
if (typeof global !== 'undefined') {
  (global as any).__wsManager = wsManager;
}

export { wsManager };
export function getWsManager() {
  if (typeof global !== 'undefined' && (global as any).__wsManager) {
    return (global as any).__wsManager;
  }
  return wsManager;
}