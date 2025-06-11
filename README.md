# Engraving Processor Pro

Advanced image processing application for historical engravings and documents, featuring real-time preview, adaptive binarization, morphological operations, and pixel art scaling algorithms.

## Features

- **Adaptive Binarization**: Sauvola, Niblack, and Otsu algorithms
- **Morphological Operations**: Opening, closing, dilation, and erosion
- **Noise Reduction**: Binary noise removal and median filtering
- **Pixel Art Scaling**: Scale2x/3x/4x algorithms
- **Real-time Preview**: WebSocket-powered instant feedback
- **High Performance**: Efficient processing with progress tracking

## Requirements

- Node.js 22.0.0 or higher
- 4GB RAM minimum (8GB recommended for large images)

## Quick Start

```bash
# Install dependencies
pnpm install

# Build the application
pnpm build

# Start the server (includes all checks)
pnpm start
```

The application will be available at http://localhost:3000

## Available Commands

- `pnpm dev` - Development mode with hot reload
- `pnpm build` - Build for production
- `pnpm start` - Start production server (runs all checks automatically)
- `pnpm test` - Run tests
- `pnpm typecheck` - Check TypeScript types

## Usage Guide

### 1. Upload an Image
Drag and drop or click to browse. Supports PNG, JPEG, TIFF, and WebP formats (max 10MB).

### 2. Adjust Parameters

**Binarization**
- **Sauvola**: Best for documents with uneven lighting
  - Window size: 11-15 for small text, 25-35 for large text
  - K parameter: 0.2-0.3 for faded text, 0.4-0.5 for high contrast
- **Niblack**: Good for high-contrast text
- **Otsu**: Automatic threshold for clear bimodal images

**Morphology**
- **Closing**: Fills gaps in text
- **Opening**: Removes noise particles
- **Dilate**: Thickens features
- **Erode**: Thins features

**Noise Reduction**
- **Binary**: Removes isolated pixels
- **Median**: Smooths while preserving edges

**Scaling**
- **Scale2x/3x/4x**: Pixel art algorithms
- **Nearest/Bilinear**: Traditional scaling

### 3. Preview & Process
- See real-time updates as you adjust settings
- Click "Process Full Resolution" to download the final result

## Architecture

```
engraving-processor-pro/
├── app/                    # Remix application
│   ├── components/         # React UI components
│   ├── routes/            # API and page routes
│   └── services/          # Server services (WebSocket, processing)
├── src/engine/            # Core image processing engine
│   ├── algorithms/        # Processing algorithms
│   ├── core/             # Core data structures
│   ├── pipeline/         # Processing orchestration
│   └── utils/            # Image I/O utilities
├── scripts/              # Server and utility scripts
└── public/              # Static assets
```

## WebSocket Architecture

The application uses WebSockets for real-time preview updates:

1. **Client** uploads image → receives image ID
2. **Client** adjusts parameters → sends via WebSocket
3. **Server** processes preview (512px max) → sends result
4. **Client** displays updated preview instantly

The WebSocket manager is initialized as a singleton on server startup and handles multiple concurrent connections with automatic reconnection.

## Troubleshooting

### Port Already in Use
```bash
lsof -i :3000
kill -9 <PID>
# Or use a different port
PORT=3001 pnpm start
```

### Memory Issues
For large images, increase Node.js memory:
```bash
NODE_OPTIONS="--max-old-space-size=4096" pnpm start
```

### WebSocket Connection Issues
- Check the health endpoint: http://localhost:3000/health
- Ensure port 3000 is not blocked by firewall
- Try refreshing the page

### Build Errors
If you see module resolution errors:
```bash
rm -rf node_modules build
pnpm install
pnpm build
```

## Health Monitoring

The server provides comprehensive health monitoring at `/health`:
- Service status (HTTP, WebSocket)
- Memory usage statistics
- Active connections count
- System diagnostics

## Development

```bash
# Run in development mode
pnpm dev

# Run tests with UI
pnpm test:ui

# Type checking
pnpm typecheck
```

## License

MIT

## Author

Ervins Strauhmanis
