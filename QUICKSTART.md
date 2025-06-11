# Quick Start Guide

## Prerequisites

- Node.js 22.0.0 or higher
- pnpm (or npm/yarn)

## Step-by-Step Setup

### 1. Install Dependencies
```bash
pnpm install
```

### 2. Fix WebSocket Setup (if needed)
```bash
pnpm fix:websocket
```
This creates necessary files and checks the setup.

### 3. Build the Application
```bash
pnpm build
```
This compiles TypeScript files and creates the production build.

### 4. Verify Setup
```bash
pnpm check
```
This runs pre-flight checks to ensure everything is configured correctly.

### 5. Start the Server
```bash
pnpm start
```
The application will be available at http://localhost:3000

## Alternative Start Methods

### Development Mode
```bash
pnpm dev
```
Runs the application in development mode with hot reload.

### Direct Server Start (skip checks)
```bash
pnpm start:direct
```
Starts the server without running pre-flight checks.

## Troubleshooting

### WebSocket Issues

If you see "WebSocket manager not found":

1. Run the fix script:
   ```bash
   pnpm fix:websocket
   ```

2. Rebuild the application:
   ```bash
   pnpm build
   ```

3. Check the setup:
   ```bash
   pnpm check
   ```

### Port Already in Use

If port 3000 is already in use:

```bash
# Find process using port 3000
lsof -i :3000

# Kill the process
kill -9 <PID>

# Or use a different port
PORT=3001 pnpm start
```

### Memory Issues

If you encounter memory errors with large images:

1. Check available memory:
   ```bash
   node -e "console.log(process.memoryUsage())"
   ```

2. Increase Node.js memory limit:
   ```bash
   NODE_OPTIONS="--max-old-space-size=4096" pnpm start
   ```

## Verify Installation

1. **Check health endpoint**:
   ```bash
   curl http://localhost:3000/health
   ```

2. **Look for**:
   - `"status": "healthy"`
   - `"websocket": true`
   - `"wsManager": true`

3. **Test the application**:
   - Open http://localhost:3000
   - Upload an image
   - Adjust parameters
   - Check that preview updates in real-time

## Common Commands

- `pnpm build` - Build the application
- `pnpm start` - Start the server with checks
- `pnpm dev` - Development mode
- `pnpm check` - Verify setup
- `pnpm fix:websocket` - Fix WebSocket issues
- `pnpm test` - Run tests
- `pnpm typecheck` - Check TypeScript types

## Next Steps

1. Read [README.md](README.md) for detailed usage instructions
2. Check [WEBSOCKET-SETUP.md](WEBSOCKET-SETUP.md) for WebSocket architecture details
3. Upload an image and start processing!