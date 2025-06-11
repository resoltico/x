// Server initialization hook - SERVER ONLY
// This file initializes server-side services on startup

// This file is not needed anymore since we're using websocket-init.server.ts
// The initialization is handled by entry.server.tsx importing websocket-init.server.ts

// Initialize websocket manager on server startup
if (typeof window === 'undefined') {
  // We're on the server
  const initializeServices = async () => {
    try {
      // Wait a bit for dynamic imports to complete
      await new Promise(resolve => setTimeout(resolve, 100));
      
      const wsManager = getWsManager();
      
      if (wsManager) {
        console.log('✅ Server services initialized');
      } else {
        console.log('⏳ WebSocket manager initializing...');
      }
    } catch (error) {
      console.error('❌ Failed to initialize server services:', error);
    }
  };
  
  initializeServices();
}