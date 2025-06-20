# Image Restoration Suite

A modern image processing application built with Go and Fyne, featuring advanced restoration algorithms for historical illustrations and documents.

## Features

- **Modern GUI**: Built with Fyne v2.6.1 for cross-platform compatibility
- **Image Processing**: Powered by OpenCV 4.11.0 through GoCV bindings
- **Real-time Preview**: Live preview of transformations with memory-efficient processing
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

### Development Build
```bash
# Development build with memory profiling and leak detection
make dev

# Or step by step:
make build-profile
make run-profile
```

### Production Build
```bash
# Production build (optimized, no profiling)
make build
make run
```

### Memory Profiling Build
```bash
# Build with GoCV MatProfile enabled
make build-profile

# Run with comprehensive memory leak detection
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

This application uses proper memory management with comprehensive leak detection:

### Monitoring Memory Usage

1. **Build with profiling** (now default for development):
   ```bash
   make build-profile
   ```

2. **Run the application**:
   ```bash
   make run-profile
   ```

3. **Monitor terminal output**:
   ```
   Initial MatProfile count: 0
   Final MatProfile count: 0  # Should be 0 for no leaks
   ```

4. **Access profiling endpoints**:
   - Main profiler: http://localhost:6060/debug/pprof/
   - Mat profiler: http://localhost:6060/debug/pprof/gocv.io/x/gocv.Mat

### Understanding MatProfile Output

When the application exits, it will show:
```
SUCCESS: No memory leaks detected - all Mats properly closed.
```

If you see leaks:
```
WARNING: Memory leaks detected! 5 Mat(s) not properly closed.
MEMORY LEAK: 5 Mat(s) were created but not cleaned up during session
```

### Memory Best Practices

1. **Always defer Close()**: 
   ```go
   mat := gocv.NewMat()
   defer mat.Close()
   ```

2. **Use cloned Mats for thread safety**:
   ```go
   // Return clones to prevent race conditions
   return processedImage.Clone()
   ```

3. **Proper cleanup in loops**:
   ```go
   for i := 0; i < iterations; i++ {
       temp := gocv.NewMat()
       defer temp.Close() // Added missing cleanup
       // process...
   }
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
├── ui.go                  # Thread-safe UI implementation
├── pipeline.go            # Memory-safe pipeline with proper cleanup
├── transformation.go      # Transformation interface
├── transform_twod_otsu.go # 2D Otsu implementation
├── transform_lanczos4.go  # Lanczos4 implementation
├── debug_*.go            # Debug modules (terminal output only)
├── helpers.go            # Utility functions
├── README.md             # This file
└── README_macOS.md       # macOS-specific build instructions
```

## Architecture

## Development Workflow

### Setting Up Development Environment
```bash
# Complete development setup
make dev
```

### Testing Memory Management
```bash
# Run with comprehensive leak detection
make check-leaks

# Monitor during development
make run-profile
```

### Debugging Memory Issues
1. **Enable profiling**: `make build-profile && make run-profile`
2. **Use the application** (load images, apply transformations)
3. **Check terminal output** for MatProfile count changes
4. **Access profiler**: http://localhost:6060/debug/pprof/gocv.io/x/gocv.Mat
5. **Look for non-zero final count** indicating leaks

## Troubleshooting

### Common Issues

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

## Quick Start

```bash
# 1. Check dependencies
make check-deps

# 2. Set up development environment  
make dev

# 3. Run with memory profiling
make run-profile

# 4. Load an image and apply transformations

# 5. Monitor terminal for memory usage:
#    "SUCCESS: No memory leaks detected" = good
#    "WARNING: Memory leaks detected" = check code
```

## Contributing

When contributing to this project:

1. **Always use profiling build**: `make build-profile`
2. **Test memory management**: `make check-leaks`
3. **Verify no leaks**: Ensure MatProfile count returns to 0
4. **Follow thread safety**: Use proper synchronization
5. **Test UI responsiveness**: Ensure no blocking operations in UI thread

## License, Author

MIT, Ervins Strauhmanis
