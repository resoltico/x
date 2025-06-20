# Image Restoration Suite

A modern image processing application built with Go and Fyne, featuring advanced restoration algorithms for historical illustrations and documents. Enhanced with GoCV's MatProfile for robust memory management.

## Features

- **Modern GUI**: Built with Fyne v2.6.1 for cross-platform compatibility
- **Image Processing**: Powered by OpenCV 4.11.0 through GoCV bindings
- **Real-time Preview**: Live preview of transformations
- **Quality Metrics**: PSNR and SSIM calculations for processed images
- **Memory Profiling**: Built-in memory leak detection using GoCV's MatProfile
- **Extensible Architecture**: Modular transformation system for easy algorithm additions

### Currently Supported Transformations

- **2D Otsu Binarization**: Advanced two-dimensional Otsu algorithm optimized for historical illustrations with adjustable parameters:
  - Window Radius (1-20)
  - Epsilon smoothing factor (0.001-0.1)
  - Morphological kernel size (1-15, odd values only)

- **Lanczos4 Scaling**: High-quality image scaling with Lanczos4 interpolation:
  - Scale factor (0.1-5.0)
  - DPI-based scaling
  - Iterative downscaling for large reductions
  - Artifact reduction filters

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

### 2. Verify Dependencies
```bash
make check-deps
```

### 3. Install Go Dependencies
```bash
make deps
```

## Building the Application

### Quick Start
```bash
# Development build with memory profiling
make build-profile

# Production build (optimized, no profiling)
make build
```

### Memory Profiling Build (Recommended for Development)
```bash
# Build with GoCV MatProfile enabled
make build-profile

# Run with memory leak detection
make run-profile
```

The MatProfile build enables comprehensive memory tracking for all GoCV Mat operations. Access the memory profiler at:
- **Main pprof interface**: http://localhost:6060/debug/pprof/
- **Mat-specific profiling**: http://localhost:6060/debug/pprof/gocv.io/x/gocv.Mat

### Cross-Platform Builds

```bash
# Windows
make build-windows

# macOS (Intel)
make build-macos

# macOS (Apple Silicon)
make build-macos-arm64

# macOS (Universal Binary)
make build-macos-universal

# Linux
make build-linux
```

### macOS App Bundle
```bash
# Install fyne packaging tool
go install fyne.io/fyne/v2/cmd/fyne@latest

# Create app bundle
make build-macos-app
```

## Running the Application

### Standard Run
```bash
make run
```

### With Memory Profiling (Recommended for Development)
```bash
make run-profile
```

### Memory Leak Detection
```bash
# Run with enhanced leak detection and monitoring
make check-leaks
```

## Memory Management

This application uses GoCV's built-in MatProfile system for memory management instead of custom tracking. Benefits include:

- **Automatic Tracking**: All Mat allocations and deallocations are tracked
- **Stack Traces**: Know exactly where memory leaks originate  
- **Zero Configuration**: Enabled with `-tags matprofile` build flag

### Monitoring Memory Usage

1. **Build with profiling**:
   ```bash
   make profile
   # or
   ./build.sh
   ```

2. **Run the application**:
   ```bash
   make run-profile
   ```

3. **Check terminal output for MatProfile count**:
   ```
   Initial MatProfile count: 0
   Final MatProfile count: 0  # Should be 0 for no leaks
   ```

### Understanding MatProfile Output

When the application exits, it will show:
```
Final MatProfile count: 0
```

If you see a non-zero count:
```
Final MatProfile count: 5
WARNING: Memory leaks detected! Check MatProfile for details.
```

This indicates memory leaks - 5 Mats were not properly closed.

4. **Check current Mat count**:
   ```bash
   make profile-count
   ```

### Understanding MatProfile Output

Example MatProfile output showing a memory leak:
```
gocv.io/x/gocv.Mat profile: total 1
1 @ 0x40b936c 0x40b93b7 0x40b94e2 0x40b95af 0x402cd87 0x40558e1
#	0x40b936b	gocv.io/x/gocv.newMat+0x4b	/go/src/gocv.io/x/gocv/core.go:153
#	0x40b93b6	gocv.io/x/gocv.NewMat+0x26	/go/src/gocv.io/x/gocv/core.go:159
#	0x40b94e1	main.processImage+0x21	/go/src/image-restoration-suite/pipeline.go:150
```

This shows:
- **total 1**: One unclosed Mat exists
- **Stack trace**: Exact location where the Mat was created
- **File/line**: Precise source location for debugging

### Memory Best Practices

1. **Always defer Close()**: 
   ```go
   mat := gocv.NewMat()
   defer mat.Close()
   ```

2. **Check MatProfile count during development**:
   - Initial count should be 0
   - Final count should return to 0

3. **Build with profiling during development**:
   ```bash
   make profile
   ./build.sh
   ```

## Usage

1. **Open Image**: Click "OPEN IMAGE" to load an image file
2. **Apply Transformations**: Select transformations from the left panel
3. **Adjust Parameters**: Fine-tune parameters using controls in the Parameters panel
4. **Preview Results**: View real-time preview with memory-efficient processing
5. **Monitor Quality**: Check PSNR and SSIM metrics in the right panel
6. **Monitor Memory**: Watch MatProfile count in terminal output
7. **Save Result**: Click "SAVE IMAGE" to export the processed image
8. **Reset**: Use "Reset" button to clear all transformations

## Project Structure

```
image-restoration-suite/
├── go.mod                  # Go module dependencies
├── Makefile               # Enhanced build system with MatProfile
├── main.go                # Application entry point with pprof server
├── ui.go                  # Main UI implementation
├── pipeline.go            # Simplified pipeline using GoCV memory management
├── transformation.go      # Transformation interface
├── transform_twod_otsu.go # 2D Otsu implementation
├── transform_lanczos4.go  # Lanczos4 scaling implementation
├── debug_gui.go          # GUI debug output (terminal only)
├── debug_pipeline.go     # Pipeline debug output (terminal only)
├── debug_image.go        # Image processing debug output (terminal only)
├── debug_render.go       # Render debug output (terminal only)
├── helpers.go            # Utility functions
├── README.md             # This file
└── README_macOS.md       # macOS-specific build instructions
```

## Architecture

### Memory Management Architecture
- **GoCV MatProfile**: Automatic tracking of all Mat allocations/deallocations (when built with `-tags matprofile`)
- **Zero Custom Tracking**: Removed custom ManagedMat wrapper
- **Simplified Pipeline**: Direct use of gocv.Mat with proper lifecycle management

### Modular Design
- **UI Layer**: Fyne-based graphical interface
- **Pipeline Layer**: Simplified image processing chain management
- **Transformation Layer**: Individual algorithm implementations
- **Debug Layer**: Comprehensive debugging with memory profiling

### Adding New Transformations
1. Implement the `Transformation` interface
2. Add parameter controls in `GetParametersWidget()`
3. Ensure proper Mat lifecycle management (defer Close())
4. Test with MatProfile enabled
5. Register in the transformations list

## Development Workflow

### Setting Up Development Environment
```bash
# Clone and setup
git clone <repository-url>
cd image-restoration-suite
make deps

# Start development with memory profiling
make profile
```

### Testing Memory Management
```bash
# Run tests with profiling
make test

# Monitor memory during development
make run-profile
```

### Debugging Memory Leaks
1. **Enable profiling**: `make profile && make run-profile`
2. **Use the application** (load images, apply transformations)
3. **Check terminal output** for MatProfile count
4. **Look for non-zero final count** indicating leaks
5. **Fix leaks** by adding proper `defer mat.Close()` calls

## Troubleshooting

### Common Issues

#### MatProfile Shows Memory Leaks
```
Final MatProfile count: 5
WARNING: Memory leaks detected!
```
**Solution**: Check your code for Mat allocations without corresponding `defer mat.Close()` calls.

#### OpenCV Not Found
```
Error: opencv not found
```
**Solution**: Install OpenCV and ensure it's in your system PATH:
```bash
make check-deps  # Verify installation
```

#### Go Version Mismatch
```
Error: go version mismatch
```
**Solution**: Upgrade to Go 1.24+ as specified in `go.mod`.

#### Fyne Dependencies Missing
```
Error: fyne dependencies not found
```
**Solution**: 
```bash
make deps
```

### Performance Tips

- **Enable MatProfile during development** to catch memory leaks early
- **Use preview mode** for real-time parameter adjustment (lower memory usage)  
- **Monitor MatProfile count** in terminal output
- **Close Mats promptly** - don't rely on garbage collection
- **Use smaller images** for parameter tuning to reduce memory pressure

## Migration from Custom Memory Management

This version removes the custom `ManagedMat` wrapper and `DebugMemory` system in favor of GoCV's built-in MatProfile. Key changes:

### Removed Files
- `debug_memory.go` - Replaced by GoCV MatProfile
- Custom `ManagedMat` wrapper - Use `gocv.Mat` directly

### New Features
- Automatic memory tracking with `-tags matprofile`
- Terminal-based MatProfile count logging
- Simple leak detection via count comparison

### Code Changes
```go
// Old approach (custom wrapper)
managedMat := NewManagedMat("name", debugMemory)
defer managedMat.Close("name", debugMemory)

// New approach (direct GoCV)
mat := gocv.NewMat()
defer mat.Close()  // MatProfile tracks automatically
```

## Contributing

When contributing to this project:

1. **Always build with profiling**: `make profile`
2. **Test memory management**: `make run-profile`
3. **Verify no leaks**: Ensure MatProfile count returns to 0
4. **Follow Mat lifecycle**: Always pair `gocv.NewMat()` with `defer mat.Close()`

## License, Author

MIT, Ervins Strauhmanis
