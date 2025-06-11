// Direct websocket manager export for server initialization
// This file provides a direct way to access the websocket manager
// without going through the Remix build system

import { wsManager } from './app/services/websocket.server.js';

// Make it available globally for the server script
if (typeof global !== 'undefined') {
  global.__wsManager = wsManager;
}

export { wsManager };
export default wsManager;