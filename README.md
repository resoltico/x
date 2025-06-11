# Engraving Processor Pro

Advanced image processing application for historical engravings and documents, built with Remix, React 18+, and Node.js 22+.

## Author

Ervins Strauhmanis

## License

MIT

## Features

- **Advanced Binarization**: Adaptive thresholding algorithms for optimal text extraction
- **Morphological Operations**: Image enhancement through mathematical morphology
- **Noise Reduction**: Binary noise removal and median filtering
- **Pixel Art Scaling**: Scale2x/3x/4x algorithms for clean upscaling
- **Real-time Preview**: WebSocket-powered instant feedback
- **Modular Architecture**: Pluggable processing algorithms

## System Requirements

- Node.js 22.0.0 or higher
- 4GB RAM minimum (8GB recommended for large images)

## Quick Start

```bash
# Install dependencies
npm install

# Build the application
npm run build

# Start the server (includes all checks and monitoring)
npm start
```

The application will be available at http://localhost:3000

## Image Processing Algorithms

### Binarization
Converts grayscale images to pure black and white. **Sauvola** adapts to local image variations using statistical analysis within sliding windows, ideal for documents with uneven lighting. **Niblack** uses simpler local mean and standard deviation calculations, working well for high-contrast text. **Otsu** automatically finds the optimal global threshold by maximizing between-class variance, best for images with clear bimodal histograms.

### Morphology
Mathematical operations that process images based on shapes. **Closing** (dilation followed by erosion) fills gaps in text and connects broken characters. **Opening** (erosion followed by dilation) removes small noise particles while preserving larger features. **Dilation** expands white regions to thicken text strokes. **Erosion** shrinks white regions to thin features or separate touching elements.

## Usage

1. **Upload an Image**: Drag and drop or click to browse (supports PNG, JPEG, TIFF, WebP)
2. **Adjust Parameters**: Fine-tune binarization, morphology, noise reduction, and scaling
3. **Preview**: See real-time updates as you adjust settings
4. **Process**: Click "Process Full Resolution" to download the result

## Parameter Guidelines

### For Historical Engravings
- Start with Sauvola binarization
- Window size: 11-15 for small text, 25-35 for large text
- K parameter: 0.2-0.3 for faded text, 0.4-0.5 for high contrast
- Enable morphological closing if text has gaps

### For Documents with Background
- Use smaller window sizes (11-19)
- Enable noise reduction for speckled backgrounds
- Try morphological opening to remove artifacts

## Project Structure

```
engraving-processor-pro/
├── app/                    # Remix application
│   ├── components/         # React components
│   ├── routes/            # API and page routes
│   └── services/          # Server-side services
├── src/engine/            # Core processing engine
│   └── algorithms/        # Image processing algorithms
├── public/               # Static assets
└── script-server.js      # Main server with integrated monitoring
```

## Troubleshooting

### WebSocket Connection Issues
- Check server logs for initialization errors
- Verify port 3000 is not blocked
- Try refreshing the page

### Preview Not Generating
- Ensure WebSocket shows "Connected" status
- Check server logs for processing errors
- Try with a smaller image first

### Port Already in Use
```bash
# Find and kill process using port 3000
lsof -i :3000
kill -9 <PID>

# Or use a different port
PORT=3001 npm start
```

## Health Monitoring

The server provides comprehensive health monitoring at http://localhost:3000/health including:
- Service status (HTTP, WebSocket)
- Memory usage statistics
- Active WebSocket connections
- Detailed error reporting

## Development

```bash
# Development mode with hot reload
npm run dev

# Run tests
npm test

# Type checking
npm run typecheck
```