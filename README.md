# Image Restoration Suite

A modern image processing application built with Go and Fyne, featuring advanced restoration algorithms for historical illustrations and documents with **enhanced memory management and Mat profiling**.

## Features

- **Modern GUI**: Built with Fyne v2.6.1 for cross-platform compatibility
- **Image Processing**: Powered by OpenCV 4.11.0 through GoCV bindings
- **Real-time Preview**: Live preview of transformations
- **Quality Metrics**: PSNR and SSIM calculations for processed images
- **Memory Leak Detection**: GoCV Mat profiling enabled by default
- **Extensible Architecture**: Modular transformation system

### Currently Supported Transformations

- **2D Otsu Binarization**: Mathematically corrected two-dimensional Otsu algorithm with proper variance calculations
- **Lanczos4 Scaling**: High-quality image scaling with guided filtering

### Supported Image Formats

- **Input**: JPEG, PNG, TIFF
- **Output**: JPEG, PNG, TIFF

## Prerequisites

### System Requirements

- **Go**: Version 1.24+ required
- **OpenCV**: Version 4.11.0+
- **Operating System**: Windows, macOS, or Linux

### OpenCV Installation

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install libopencv-dev
```

#### macOS
```bash
brew install opencv
```

#### Windows
Download and install OpenCV from the official website or use vcpkg:
```cmd
vcpkg install opencv
```

## Installation

### 1. Clone the Repository
```bash
git clone <repository-url>
cd image-restoration-suite
```

### 2. Download Go Dependencies
```bash
go mod download
go mod tidy
```

## Building the Application

### Quick Build (with Memory Profiling)
```bash
chmod +x build.sh
./build.sh
```

### Using Makefile
```bash
# Build with Mat profiling (recommended for development)
make profile

# Build optimized release version
make build

# Run with profiling
make run-profile

# Check for memory leaks
make check-leaks
```

### Manual Build Commands

#### Development Build (with Mat profiling)
```bash
go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite .
```

#### Production Build (optimized)
```bash
go build -ldflags="-s -w" -o image-restoration-suite .
```

### Cross-Platform Builds

#### Windows
```bash
GOOS=windows GOARCH=amd64 go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite.exe .
```

#### macOS
```bash
GOOS=darwin GOARCH=amd64 go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite-mac .
```

#### Linux
```bash
GOOS=linux GOARCH=amd64 go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite-linux .
```

## Running the Application

### With Memory Profiling (Recommended)
```bash
./image-restoration-suite
```

The terminal will show Mat creation/cleanup tracking when profiling is enabled.

### Debug Mode

All builds include comprehensive debugging that outputs to terminal:
- GUI debugging (enabled by default)
- Memory debugging (enabled by default) 
- Pipeline debugging (enabled by default)
- Image processing debugging (disabled by default)
- Render debugging (disabled by default)

## Memory Management

### Mat Profiling
This application uses GoCV's built-in Mat profiling to detect memory leaks:

- All builds include `-tags matprofile` by default
- Mat creation and cleanup is tracked automatically
- Memory leaks are reported in terminal output
- No additional setup required

### Safe Memory Practices
The codebase follows proper GoCV memory management:
- Direct Mat management with proper defer patterns
- Automatic cleanup on panic recovery
- Thread-safe pipeline operations
- Error boundaries around all Mat operations

## Usage

1. **Open Image**: Click "OPEN IMAGE" to load an image file
2. **Apply Transformations**: Select transformations from the left panel
3. **Adjust Parameters**: Fine-tune parameters in the Parameters panel
4. **Preview Results**: View real-time preview with quality metrics
5. **Save Result**: Click "SAVE IMAGE" to export the processed image
6. **Reset**: Use "Reset" button to clear all transformations

## Project Structure

```
image-restoration-suite/
├── go.mod                    # Go module dependencies
├── main.go                   # Application entry point with profiling
├── ui.go                     # Main UI with error boundaries
├── pipeline.go               # Simplified pipeline without SafeMat
├── transformation.go         # Transformation interface
├── transform_twod_otsu.go    # Corrected 2D Otsu implementation  
├── transform_lanczos4.go     # Lanczos4 scaling transformation
├── debug_*.go               # Debug modules (terminal output only)
├── helpers.go               # Utility functions
├── build.sh                 # Build script with profiling
├── Makefile                 # Build automation
└── README.md                # This file
```

## Architecture Changes

### Removed Components
- **SafeMat**: Removed custom wrapper - using direct GoCV Mat management
- **Complex memory tracking**: Replaced with GoCV's native profiling

### Enhanced Components  
- **2D Otsu**: Mathematically corrected implementation
- **SSIM Calculation**: Fixed structural similarity formula
- **Error Boundaries**: Proper panic recovery around all Mat operations
- **Memory Profiling**: Automatic leak detection via GoCV

## Troubleshooting

### Memory Leaks
With Mat profiling enabled, memory leaks are automatically detected:
```bash
# Run and check terminal output for Mat tracking
./image-restoration-suite
```

### Common Issues

#### OpenCV Not Found
```
Error: opencv not found
```
**Solution**: Ensure OpenCV is properly installed and accessible in your system PATH.

#### Go Version Mismatch  
```
Error: go version mismatch
```
**Solution**: Upgrade to Go 1.24+ as specified in `go.mod`.

### Performance Tips

- Mat profiling adds minimal overhead
- Use preview scaling for real-time parameter adjustment  
- Monitor terminal for memory leak warnings
- Close application properly to generate final memory report

## Development

### Adding New Transformations
1. Implement the `Transformation` interface
2. Add parameter controls in `GetParametersWidget()`
3. Ensure proper Mat cleanup with defer patterns
4. Test with Mat profiling enabled

### Memory Debugging
```bash
# Enable all debug modules in main.go:
var debugConfig = DebugConfig{
    GUI:      true,
    Image:    true,
    Memory:   true,
    Pipeline: true, 
    Render:   true,
}
```

## License

MIT License - Ervins Strauhmanis