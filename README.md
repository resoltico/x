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

### Go Dependency Management

This project uses Go modules for dependency management. Here's how it works:

#### Understanding Go Module Files

- **`go.mod`**: Defines the module name, Go version, and direct dependencies
- **`go.sum`**: Contains cryptographic checksums for dependency verification
- Both files should be committed to version control

#### Common Dependency Commands

```bash
# Download dependencies to local cache (usually silent when successful)
go mod download

# Add missing dependencies and remove unused ones
go mod tidy

# View all dependencies (direct and indirect)
go list -m all

# Check for available updates
go list -u -m all

# Update only patch versions (safest - bug fixes only)
go get -u=patch ./...

# Update to latest compatible minor versions (may include new features)
go get -u ./...

# Manually update to specific major versions (POTENTIALLY BREAKING)
# You must explicitly specify the new major version
go get github.com/example/package/v2@latest
go get github.com/example/package/v3@latest

# Add a new dependency
go get github.com/example/package

# Add a specific version
go get github.com/example/package@v1.2.3

# Remove a dependency (run after removing import statements)
go mod tidy
```

**⚠️ Warning About Dependency Updates:**

- **Patch updates** (`go get -u=patch`): Generally safe - only bug fixes and security patches
- **Minor updates** (`go get -u`): Usually safe - new features but backward compatible within the same major version
- **Major updates**: **POTENTIALLY BREAKING** - Go does NOT automatically update major versions. You must manually specify them (e.g., `/v2`, `/v3`)

**Before updating dependencies:**
1. Commit your current working code
2. Review release notes for major version changes
3. Test thoroughly after updates
4. Consider updating one dependency at a time for easier troubleshooting
5. Use `go mod why <module>` to understand impact before updating critical dependencies

**Finding and handling major version updates:**
```bash
# Check what major versions are available
go list -m -versions github.com/example/package

# See which dependencies have newer major versions available
go list -u -m all | grep -v "indirect"

# Manually upgrade to a new major version (requires code changes)
go get github.com/example/package/v2@latest
```

#### Dependency Analysis

```bash
# See why a dependency is included
go mod why github.com/example/package

# View dependency graph
go mod graph

# Verify dependencies match go.sum
go mod verify

# Show module information
go mod edit -json
```

#### Version Selection

Go uses **Minimal Version Selection (MVS)**:
- Always selects the minimum version that satisfies all requirements
- Prefers semantic versioning (v1.2.3)
- Major version changes (v1 to v2) are treated as different modules
- **Breaking changes** typically occur with major version bumps (v1.x.x → v2.0.0)

**Semantic Versioning in Go:**
- **MAJOR** version (v1 → v2): Incompatible API changes - **BREAKING**
- **MINOR** version (v1.1 → v1.2): New functionality, backward compatible
- **PATCH** version (v1.1.1 → v1.1.2): Bug fixes, backward compatible

**Managing Breaking Changes:**
```bash
# List available versions for a specific module
go list -m -versions github.com/example/package

# Check which dependencies might have major version updates
go list -u -m all

# Manually upgrade to a specific major version (requires import path changes)
go get github.com/example/package/v2@latest

# Pin to a specific version to avoid surprises
go get github.com/example/package@v1.4.2
```

**Important**: Go treats major versions as separate modules. Upgrading from `v1` to `v2` requires:
1. Changing import paths in your code from `github.com/example/package` to `github.com/example/package/v2`
2. Potentially updating your code to match the new API
3. Testing thoroughly as breaking changes are expected

#### Handling Problem Dependencies

```bash
# Replace a dependency with a fork or local version
go mod edit -replace github.com/original/package=github.com/fork/package@v1.0.0

# Replace with local directory
go mod edit -replace github.com/example/package=../local-package

# Remove a replace directive
go mod edit -dropreplace github.com/example/package
```

#### Vendoring (Optional)

```bash
# Copy dependencies to vendor/ directory
go mod vendor

# Build using vendored dependencies
go build -mod=vendor ./cmd/app
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
├── go.sum                       # Dependency checksums
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

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
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

4. **Dependency issues**:
   - Run `go mod verify` to check dependency integrity
   - Use `go clean -modcache` to clear module cache if needed
   - Check `go mod why <module>` to understand why a dependency is needed
   - **After dependency updates**: Test thoroughly as updates may introduce breaking changes
   - If builds break after updates, check release notes and consider reverting to previous versions

5. **Breaking changes from dependency updates**:
   - **Symptoms**: Build errors, changed function signatures, missing methods
   - **Solutions**: 
     - Revert to previous version: `go get github.com/example/package@v1.2.3`
     - Check migration guides in the dependency's documentation
     - Update your code to match new API requirements
     - Use `go mod graph` to see which dependency introduced the breaking change

5. **Performance issues**:
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
