# Image Restoration Suite

A modern image processing application built with Go and Fyne, featuring advanced restoration algorithms for historical illustrations and documents.

## Features

- **Modern GUI**: Built with Fyne v2.6.1 for cross-platform compatibility
- **Image Processing**: Powered by OpenCV 4.11.0 through GoCV bindings
- **Real-time Preview**: Live preview of transformations with zoom controls
- **Quality Metrics**: PSNR and SSIM calculations for processed images
- **Extensible Architecture**: Modular transformation system for easy algorithm additions

### Currently Supported Transformations

- **2D Otsu Binarization**: Advanced two-dimensional Otsu algorithm optimized for historical illustrations with adjustable parameters:
  - Window Radius (1-20)
  - Epsilon smoothing factor (0.001-0.1)
  - Morphological kernel size (1-15, odd values only)

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
cd advanced-image-processing
```

### 2. Verify Dependencies
Ensure you have the exact versions specified in `go.mod`:
- Go 1.24
- Fyne v2.6.1
- GoCV v0.41.0

### 3. Download Go Dependencies
```bash
go mod download
go mod tidy
```

## Building the Application

### Build for Current Platform
```bash
go build -o image-restoration-suite .
```

### Cross-Platform Builds

#### Windows (from any platform)
```bash
GOOS=windows GOARCH=amd64 go build -o image-restoration-suite.exe .
```

#### macOS (from any platform)
```bash
GOOS=darwin GOARCH=amd64 go build -o image-restoration-suite-mac .
```

#### Linux (from any platform)
```bash
GOOS=linux GOARCH=amd64 go build -o image-restoration-suite-linux .
```

### Optimized Release Build
```bash
go build -ldflags="-s -w" -o image-restoration-suite .
```

## Running the Application

### Standard Run
```bash
go run .
```

### Or run the built executable
```bash
# Windows
./image-restoration-suite.exe

# macOS/Linux
./image-restoration-suite
```

## Debug Mode

The application includes comprehensive debugging capabilities that output to the terminal/console that are enabled by default.

## Usage

1. **Open Image**: Click "OPEN IMAGE" to load an image file
2. **Apply Transformations**: Select "2D Otsu" from the transformations list
3. **Adjust Parameters**: Fine-tune parameters using sliders in the Parameters panel
4. **Preview Results**: View real-time preview with zoom controls
5. **Monitor Quality**: Check PSNR and SSIM metrics in the right panel
6. **Save Result**: Click "SAVE IMAGE" to export the processed image
7. **Reset**: Use "Reset" button to clear all transformations

## Project Structure

```
advanced-image-processing/
├── go.mod                  # Go module dependencies
├── main.go                 # Application entry point
├── ui.go                   # Main UI implementation
├── pipeline.go             # Image processing pipeline
├── transformation.go       # Transformation interface
├── twod_otsu.go           # 2D Otsu implementation
├── debug_gui.go           # GUI debug output (terminal only)
├── debug_pipeline.go      # Pipeline debug output (terminal only)
├── debug_image.go         # Image processing debug output (terminal only)
└── README.md              # This file
```

## Architecture

### Modular Design
- **UI Layer**: Fyne-based graphical interface
- **Pipeline Layer**: Image processing chain management
- **Transformation Layer**: Individual algorithm implementations
- **Debug Layer**: Isolated debugging functionality

### Adding New Transformations
1. Implement the `Transformation` interface
2. Add parameter controls in `GetParametersWidget()`
3. Register in the transformations list

## Troubleshooting

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

#### Fyne Dependencies
```
Error: fyne dependencies not found
```
**Solution**: Run `go mod download` and ensure you have the required system libraries for Fyne.

### Performance Tips

- Enable hardware acceleration for OpenCV if available
- Use smaller images for real-time parameter adjustment
- Monitor memory usage through debug output
- Close the application properly to free OpenCV resources

## License, Author

MIT, Ervins Strauhmanis
