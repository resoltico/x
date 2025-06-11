# Quick Start Guide

## Prerequisites

- Node.js 22.0.0 or higher
- npm or pnpm

## Installation & Running

```bash
# 1. Install dependencies
npm install

# 2. Build the application
npm run build

# 3. Start with automatic checks (recommended)
npm run start:check

# Or start directly
npm start
```

## First Time Setup Checklist

1. **Verify Node.js version**:
   ```bash
   node --version  # Should be v22.0.0 or higher
   ```

2. **Run startup checks**:
   ```bash
   npm run check:startup
   ```

3. **Test the server**:
   ```bash
   npm run test:health
   ```

## Troubleshooting Connection Issues

If you see "Connection lost. Reconnecting..." in the UI:

1. **Check server is running**:
   ```bash
   curl http://localhost:3000/health
   ```

2. **Test WebSocket specifically**:
   ```bash
   npm run test:ws
   ```

3. **Debug WebSocket interactively**:
   ```bash
   npm run debug:ws
   ```

4. **Check for port conflicts**:
   ```bash
   lsof -i :3000
   ```

## Processing Your First Image

1. Open http://localhost:3000 in your browser
2. Drag and drop an image (PNG, JPEG, TIFF, or WebP)
3. Wait for the preview to generate
4. Adjust parameters as needed
5. Click "Process Full Resolution" to download the result

## Recommended Settings

### For Historical Engravings
- Method: Sauvola
- Window Size: 15
- K: 0.34
- R: 128

### For Faded Text
- Method: Sauvola
- Window Size: 11-15
- K: 0.2-0.3
- Enable morphological closing

### For Documents with Noise
- Enable noise reduction (Binary method)
- Threshold: 4
- Consider morphological opening

## Server Commands

- `npm start` - Start the server
- `npm run start:check` - Start with validation checks
- `npm run dev` - Development mode with hot reload
- `npm run build` - Build for production
- `npm test` - Run tests
- `npm run test:health` - Check server health
- `npm run debug:ws` - Interactive WebSocket debugging

## Need Help?

1. Check the server logs in the terminal
2. Look for error messages in the browser console
3. Run health checks: `npm run test:health`
4. Refer to the full README.md for detailed documentation