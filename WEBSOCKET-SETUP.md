# WebSocket Setup and Troubleshooting

## Architecture

The application uses WebSockets for real-time preview updates. The WebSocket server is initialized alongside the HTTP server to provide instant feedback when processing parameters change.

## Key Files

1. **app/services/websocket.server.ts** - Main WebSocket server implementation
2. **app/services/websocket.singleton.ts** - Singleton wrapper to prevent tree-shaking
3. **server.websocket.js** - Direct export for server initialization
4. **script-server.js** - Main server script that initializes everything

## How It Works

1. The WebSocket manager is created as a singleton in `websocket.server.ts`
2. The singleton is wrapped in `websocket.singleton.ts` to ensure it's available globally
3. The server script tries multiple approaches to import the WebSocket manager:
   - Global variable (`global.__wsManager`)
   - Direct import from `server.websocket.js`
   - Direct import from app services
   - Singleton getter function

## Troubleshooting

### WebSocket manager not found in build

This happens when the build process tree-shakes the WebSocket exports. Solutions:

1. **Rebuild the application**:
   ```bash
   pnpm build
   ```

2. **Check if server.websocket.js exists**:
   - If not, the file might not have been created
   - Ensure all files are saved

3. **Verify imports in entry files**:
   - `app/root.tsx` should import `./server-init`
   - `app/entry.server.tsx` should import `./services/websocket.singleton`

### WebSocket not initializing

1. **Check server logs** for specific error messages
2. **Verify port availability** - ensure port 3000 is not in use
3. **Check Node.js version** - must be 22.0.0 or higher

### Preview not updating

1. **Check WebSocket connection status** in the UI (top right corner)
2. **Open browser console** and look for WebSocket errors
3. **Check server logs** for processing errors
4. **Verify the /health endpoint**: http://localhost:3000/health

## Testing WebSocket Connection

1. Start the server:
   ```bash
   pnpm start
   ```

2. Check health endpoint:
   ```bash
   curl http://localhost:3000/health
   ```

3. Look for:
   - `"websocket": true`
   - `"wsManager": true`
   - `"wsConnections": <number>`

## Manual WebSocket Test

You can test the WebSocket connection using the browser console:

```javascript
const ws = new WebSocket('ws://localhost:3000/ws');
ws.onopen = () => console.log('Connected');
ws.onmessage = (e) => console.log('Message:', JSON.parse(e.data));
ws.onerror = (e) => console.error('Error:', e);
ws.onclose = () => console.log('Disconnected');

// Send a test ping
ws.send(JSON.stringify({ type: 'ping' }));
```

## Common Issues and Solutions

### Issue: "WebSocket manager not found in build"
**Solution**: The build process is tree-shaking the exports. Ensure:
- `vite.config.ts` includes websocket modules in `ssr.noExternal`
- The post-build script runs successfully
- Try using the direct import approach with `server.websocket.js`

### Issue: "WebSocket connection lost"
**Solution**: 
- Check if the server is still running
- Verify no firewall is blocking WebSocket connections
- Check for memory issues using the health endpoint

### Issue: "Preview generation timeout"
**Solution**:
- Large images may take longer to process
- Check server resources (CPU/memory)
- Try with a smaller image first
- Adjust preview size in `ImageLoader.createPreview()`