# Advanced Image Processing Application

A powerful image processing application designed for historical illustrations, engravings, and document scans. Built with Go, Fyne GUI toolkit, and OpenCV (gocv).

## Features

- **Multiple Binarization Algorithms**: Otsu, Niblack, Sauvola
- **Morphological Operations**: Erosion, Dilation, Opening, Closing
- **Noise Reduction**: Gaussian Blur, Median Filter, Bilateral Filter
- **Scaling**: Bilinear, Bicubic, Lanczos interpolation
- **Color Manipulation**: Grayscale conversion, Color overlays
- **Real-time Preview**: Live image processing with parameter adjustment
- **Preset Management**: Save and load transformation sequences
- **Debugging Support**: Comprehensive error reporting and logging
- **Native macOS Interface**: Optimized for Apple Silicon

## System Requirements

- macOS 10.15+ (Catalina or later)
- Apple Silicon (M1/M2/M3) or Intel processor
- At least 4GB RAM
- 500MB free disk space

## Installation and Building

### Prerequisites

1. **Install Go 1.24.4**:
   ```bash
   # Using Homebrew
   brew install go@1.24
   
   # Or download from https://golang.org/dl/
   # Make sure Go 1.24.4 is in your PATH
   go version  # Should show go1.24.4
   ```

2. **Install OpenCV**:
   ```bash
   # Install OpenCV via Homebrew
   brew install opencv
   
   # Verify installation
   pkg-config --modversion opencv4
   ```

3. **Install Xcode Command Line Tools** (if not already installed):
   ```bash
   xcode-select --install
   ```

### Building the Application

1. **Clone or create the project directory**:
   ```bash
   mkdir advanced-image-processing
   cd advanced-image-processing
   ```

2. **Initialize the Go module and install dependencies**:
   ```bash
   # Copy the provided go.mod file to your project directory
   go mod download
   go mod tidy
   ```

3. **Build the application**:
   ```bash
   # Development build (with debug symbols)
   go build -o build/AdvancedImageProcessing ./cmd/app
   
   # Production build (optimized)
   go build -ldflags="-s -w" -o build/AdvancedImageProcessing ./cmd/app
   ```

4. **Test the build**:
   ```bash
   ./build/AdvancedImageProcessing --help
   ./build/AdvancedImageProcessing --debug
   ```

### Creating a macOS App Bundle

1. **Install the Fyne packaging tool**:
   ```bash
   go install fyne.io/fyne/v2/cmd/fyne@latest
   ```

2. **Create the app bundle**:
   ```bash
   # Create app bundle with icon
   fyne package -os darwin -name "Advanced Image Processing" \
     -appVersion "1.0.0" -appBuild 1 -appID "com.strauhmanis.advanced-image-processing" \
     ./cmd/app
   
   # This creates "Advanced Image Processing.app"
   ```

3. **Test the app bundle**:
   ```bash
   open "Advanced Image Processing.app"
   ```

### Code Signing (Personal Use)

Since you mentioned you don't have a developer certificate but want to sign with a personal certificate:

1. **Create a self-signed certificate** (if you don't have one):
   ```bash
   # Open Keychain Access
   open /Applications/Utilities/Keychain\ Access.app
   
   # Go to Keychain Access > Certificate Assistant > Create a Certificate
   # Name: "Advanced Image Processing Certificate"
   # Identity Type: Self Signed Root
   # Certificate Type: Code Signing
   # Check "Let me override defaults"
   # Set validity period (e.g., 365 days)
   # Continue through the wizard with default settings
   ```

2. **Sign the application**:
   ```bash
   # Find your certificate name
   security find-identity -v -p codesigning
   
   # Sign the app bundle (replace with your certificate name)
   codesign --deep --force --verbose --sign "Advanced Image Processing Certificate" \
     "Advanced Image Processing.app"
   
   # Verify the signature
   codesign --verify --verbose "Advanced Image Processing.app"
   spctl --assess --verbose "Advanced Image Processing.app"
   ```

3. **Handle Gatekeeper warnings**:
   - The first time you run the app, macOS will show a warning
   - Go to System Preferences > Security & Privacy > General
   - Click "Allow Anyway" next to the blocked app message
   - Alternatively, you can disable Gatekeeper temporarily:
   ```bash
   # Disable Gatekeeper (not recommended for production)
   sudo spctl --master-disable
   
   # Re-enable after testing
   sudo spctl --master-enable
   ```

### Alternative: Ad-hoc Signing

For personal use, you can use ad-hoc signing:

```bash
# Sign with ad-hoc signature (no certificate needed)
codesign --deep --force --sign - "Advanced Image Processing.app"

# Allow the app through Gatekeeper
sudo xattr -r -d com.apple.quarantine "Advanced Image Processing.app"
```

### Running the Application

1. **From command line**:
   ```bash
   # Run with debug mode
   ./build/AdvancedImageProcessing --debug
   
   # Run normally
   ./build/AdvancedImageProcessing
   ```

2. **From app bundle**:
   ```bash
   # Double-click the .app file in Finder, or:
   open "Advanced Image Processing.app"
   
   # Run with debug mode from app bundle
   open "Advanced Image Processing.app" --args --debug
   ```

## Development

### Project Structure

```
advanced-image-processing/
├── cmd/app/main.go              # Application entry point
├── internal/
│   ├── gui/                     # GUI components
│   ├── image_processing/        # Core image processing
│   ├── transforms/              # Transformation algorithms
│   ├── models/                  # Data structures
│   ├── presets/                 # Preset management
│   └── utils/                   # Utilities
├── tests/                       # Test files
├── go.mod                       # Go module definition
└── README.md                    # This file
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/transforms/binarization
```

### Debug Mode

The application supports debug mode for development:

```bash
# Enable debug logging and detailed error reporting
./build/AdvancedImageProcessing --debug
```

Debug mode provides:
- Verbose logging to console
- Detailed error dialogs with stack traces
- Performance monitoring information
- Parameter validation details

## Usage

1. **Load an Image**: File > Open Image... (⌘O)
2. **Add Transformations**: Click on transformation categories in the sidebar
3. **Adjust Parameters**: Select a transformation and modify parameters in the right panel
4. **Preview Results**: View original and processed images in the main area
5. **Save Results**: File > Save Image... (⌘S)
6. **Manage Presets**: Presets menu for saving/loading transformation sequences

### Supported Image Formats

- **Input**: JPEG, PNG, TIFF, BMP
- **Output**: PNG (recommended for processed images)

### Transformation Categories

- **Binarization**: Otsu, Niblack, Sauvola
- **Morphology**: Erosion, Dilation
- **Noise Reduction**: Gaussian Blur
- **Color Manipulation**: Grayscale conversion

## Troubleshooting

### Common Issues

1. **"Cannot open app" error**:
   - The app is not signed or Gatekeeper is blocking it
   - Follow the code signing instructions above
   - Or use: `sudo xattr -r -d com.apple.quarantine "path/to/app"`

2. **OpenCV not found**:
   - Ensure OpenCV is installed: `brew install opencv`
   - Check pkg-config: `pkg-config --libs opencv4`
   - Verify CGO is enabled: `go env CGO_ENABLED` (should be "1")

3. **Build errors**:
   - Ensure Go 1.24.4 is installed and active
   - Run `go mod tidy` to clean up dependencies
   - Check that Xcode command line tools are installed

4. **Performance issues**:
   - Run in debug mode to identify bottlenecks
   - Ensure adequate RAM (4GB+)
   - Try smaller image sizes for testing

### Getting Help

1. **Check Logs**: Run with `--debug` flag for detailed logging
2. **Error Reporting**: Use the built-in error reporting feature
3. **System Info**: Help > About for version information

## License

MIT License - see source files for details.

## Author

Ervins Strauhmanis

---

**Note**: This application is designed for personal use. For distribution through the Mac App Store or to other users, you would need a proper Apple Developer certificate and notarization.