# Engraving Processor Pro

Advanced image processing application for historical engravings and documents, built with Remix, React 18+, and Node.js 22+.

## Author

Ervins Strauhmanis

## License

MIT

## Features

- **Advanced Binarization**: Sauvola adaptive thresholding for optimal text extraction
- **Morphological Operations**: Close, open, dilate, and erode operations for cleaning
- **Noise Reduction**: Binary noise removal and median filtering
- **Pixel Art Scaling**: Scale2x algorithm for clean upscaling
- **Real-time Preview**: Instant feedback with WebSocket-powered updates
- **Modular Architecture**: Clean separation of concerns with pluggable algorithms
- **Health Monitoring**: Built-in health checks and diagnostics

## System Requirements

- Node.js 22.0.0 or higher
- macOS (optimized for macOS, but works on other platforms)
- 4GB RAM minimum (8GB recommended for large images)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd engraving-processor-pro
```

2. Install dependencies:
```bash
npm install
```

3. Build the application:
```bash
npm run build
```

## Running the Application

### Development Mode

```bash
npm run dev
```

The application will start on http://localhost:3000

### Production Mode

```bash
npm run build
npm start
```

## Health Checks and Diagnostics

### Startup Checks

Before starting the server, run startup validation:

```bash
node startup-checks.js
```

This will verify:
- Node.js version compatibility
- Required directories exist
- Critical files are present
- Port availability
- Environment configuration

### Server Health Check

After starting the server, verify all systems are operational:

```bash
npm run test:health
```

This checks:
- HTTP server connectivity
- WebSocket functionality
- Static asset serving
- Error handling

### WebSocket Testing

To specifically test WebSocket connectivity:

```bash
npm run test:ws
```

### Live Health Monitoring

Access the health endpoint while the server is running:

```bash
curl http://localhost:3000/health
```

## Troubleshooting

### WebSocket Connection Issues

**Symptom**: "Connection lost. Reconnecting..." message in the UI

**Solutions**:
1. Check server logs for WebSocket initialization errors
2. Verify port 3000 is not blocked by firewall
3. Run `npm run test:ws` to diagnose connection issues
4. Check browser console for detailed error messages
5. Try refreshing the page or clicking "Retry Now"

### Preview Not Generating

**Symptom**: "Generating preview..." spinner runs indefinitely

**Solutions**:
1. Check WebSocket connection status (green "Connected" indicator)
2. Verify the image was uploaded successfully
3. Check server logs for processing errors
4. Try with a smaller image first
5. Reset parameters to defaults and retry

### Port Already in Use

**Symptom**: "Error: listen EADDRINUSE: address already in use :::3000"

**Solutions**:
```bash
# Find process using port 3000
lsof -i :3000

# Kill the process
kill -9 <PID>

# Or use a different port
PORT=3001 npm start
```

### Build Errors

**Symptom**: Missing directories or files when starting server

**Solutions**:
1. Run a fresh build: `npm run build`
2. Clear build cache: `rm -rf build/`
3. Verify all source files are present
4. Check Node.js version: `node --version` (must be 22+)

## Usage Guide

### 1. Upload an Image

- Click the upload area or drag and drop an image
- Supported formats: PNG, JPEG, TIFF, WebP
- The application will display image metadata and a thumbnail

### 2. Adjust Parameters

The application provides four main parameter categories:

#### Binarization (Always Applied)

- **Method**: Choose between Sauvola, Niblack, or Otsu algorithms
- **Window Size** (5-51): Size of the local window for adaptive thresholding
  - Smaller values: More local detail, may introduce noise
  - Larger values: Smoother results, may lose fine details
- **Threshold K** (0.1-1.0): Controls the threshold bias
  - Lower values: More black pixels (darker result)
  - Higher values: More white pixels (lighter result)
- **Parameter R** (0-255): Dynamic range of standard deviation
  - Lower values: More sensitive to local variations
  - Higher values: Less sensitive, more uniform results

#### Morphology (Optional)

- **Enable**: Toggle morphological operations on/off
- **Operation**:
  - **Close**: Fills gaps in text (dilate then erode)
  - **Open**: Removes small noise (erode then dilate)
  - **Dilate**: Thickens features
  - **Erode**: Thins features
- **Kernel Size** (3-9): Size of the morphological kernel
- **Iterations** (1-3): Number of times to apply the operation

#### Noise Reduction (Optional)

- **Enable**: Toggle noise reduction on/off
- **Method**:
  - **Binary**: Removes isolated pixels based on neighbor count
  - **Median**: Applies median filtering
- **Threshold** (1-8, for binary): Minimum neighbors required
- **Window Size** (3-7, for median): Size of the median filter

#### Scaling (Optional)

- **Method**: None, 2x, 3x, or 4x scaling
- **Algorithm**:
  - **Scale2x/Scale3x**: Pixel art algorithms that preserve sharp edges
  - **Nearest Neighbor**: Simple pixel replication
  - **Bilinear**: Smooth interpolation

### 3. Preview and Process

- The preview updates automatically as you adjust parameters
- Use the toggle button to switch between original and processed views
- The histogram shows the brightness distribution
- Click "Process Full Resolution" to process the entire image
- The processed image will download automatically

## Parameter Tips

### For Historical Engravings

1. Start with default Sauvola settings
2. Adjust window size based on text size:
   - Small text: 11-15
   - Medium text: 15-25
   - Large text: 25-35
3. Fine-tune K parameter:
   - Faded text: Lower K (0.2-0.3)
   - High contrast: Higher K (0.4-0.5)
4. Enable morphological closing if text has gaps
5. Use binary noise reduction for scattered dots

### For Line Art

1. Use larger window sizes (25-35)
2. Lower K values (0.2-0.3) to preserve lines
3. Enable dilation to thicken thin lines
4. Scale2x for clean upscaling

### For Documents with Background

1. Smaller window sizes (11-19) for text
2. Adjust R parameter based on background variation
3. Enable noise reduction for speckled backgrounds
4. Morphological opening to remove background artifacts

## Project Structure

```
engraving-processor-pro/
├── app/                    # Remix application
│   ├── components/         # React components
│   ├── routes/            # Remix routes and API endpoints
│   ├── services/          # Server-side services
│   └── utils/             # Client utilities
├── src/                   # Core processing engine
│   └── engine/
│       ├── algorithms/    # Image processing algorithms
│       ├── core/         # Core data structures
│       ├── pipeline/     # Processing pipeline
│       └── utils/        # Engine utilities
├── public/               # Static assets
├── build/               # Build output (generated)
├── server.js            # Custom server with WebSocket support
├── startup-checks.js    # Startup validation script
├── test-server-health.js # Health check script
└── test-websocket.js    # WebSocket test script
```

## Testing

Run the test suite:

```bash
npm test
```

Run tests with UI:

```bash
npm run test:ui
```

Health checks:

```bash
npm run test:health  # Full health check
npm run test:ws      # WebSocket test only
```

## Development

### Adding New Algorithms

1. Create a new class in the appropriate algorithm folder
2. Extend the base class (e.g., `BaseBinarizer`)
3. Implement required methods
4. Register in `ProcessingPipeline`

Example:
```javascript
export class MyBinarizer extends BaseBinarizer {
  process(imageData) {
    // Implementation
  }
  
  getParameters() {
    // Return current parameters
  }
  
  setParameters(params) {
    // Update parameters
  }
}
```

### Modifying UI

- Components are in `app/components/`
- Use Tailwind CSS for styling
- Follow existing patterns for parameter controls

## Performance Optimization

- The application uses integral images for efficient local statistics
- Preview images are limited to 512px for fast updates
- WebSocket debouncing prevents excessive updates
- Algorithms are optimized for binary images

## Monitoring

The server provides detailed logging for:
- WebSocket connections and messages
- Image processing stages and timing
- Error conditions with context
- Health status and metrics

Check server logs for detailed diagnostics when troubleshooting issues.