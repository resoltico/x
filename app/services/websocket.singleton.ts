// Global singleton for WebSocket manager - SERVER ONLY
// The .server.ts extension ensures this is never bundled for the client

let wsManagerInstance: any;

// Only create instance on server
if (typeof window === 'undefined') {
  // Dynamic import to avoid client bundling
  import('./websocket.server.js').then(({ wsManager }) => {
    wsManagerInstance = wsManager;
    
    // Store in global for server access
    if (typeof global !== 'undefined') {
      (global as any).__wsManager = wsManager;
    }
  }).catch(err => {
    console.error('Failed to load WebSocket manager:', err);
  });
}

export function getWsManager() {
  if (typeof window !== 'undefined') {
    throw new Error('WebSocket manager is server-only');
  }
  
  // Try multiple sources
  if (wsManagerInstance) return wsManagerInstance;
  if (typeof global !== 'undefined' && (global as any).__wsManager) {
    return (global as any).__wsManager;
  }
  
  // This will be undefined initially until the dynamic import completes
  return wsManagerInstance;
}

// Re-export for backward compatibility
export { wsManagerInstance as wsManager };