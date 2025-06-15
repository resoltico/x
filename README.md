# Advanced Image Processing v2.0

A professional-grade image processing application for historical documents with ROI selection, Local Adaptive Algorithms (LAA), and comprehensive quality metrics. Built with cutting-edge Go technologies.

## 🚀 Key Features

### ROI Selection & Processing
- **Rectangle Selection**: Click and drag to select rectangular regions
- **Freehand Selection**: Draw custom polygonal regions with multiple points
- **Mask-Based Processing**: Apply algorithms only to selected areas while preserving the rest
- **Visual Feedback**: Real-time selection overlay with interactive handles

### Advanced Binarization Algorithms
- **Multi-Level Otsu**: 2-level and 3-level thresholding for complex documents
- **Local Adaptive Otsu**: Dynamic window-based thresholding with interpolation
- **True Niblack**: Proper implementation with local mean and standard deviation calculation
- **True Sauvola**: Dynamic range normalization optimized for historical documents
- **Wolf-Jolion**: Enhanced Sauvola variant specifically for degraded documents
- **NICK**: Normalized Image Center of K-means for low-contrast images

### Comprehensive Quality Metrics
- **PSNR**: Peak Signal-to-Noise Ratio for noise assessment
- **SSIM**: Structural Similarity Index for perceptual quality measurement
- **F-Measure**: Precision/recall analysis for binarization quality
- **Real-time Assessment**: Live quality feedback during processing with color indicators
- **Per-Region Metrics**: Quality assessment for individually selected regions

### Professional Workflow
- **Pipeline Processing**: Chain multiple algorithms with live preview
- **Parameter Optimization**: Real-time parameter adjustment with immediate quality feedback
- **Debounced Processing**: Intelligent processing delays to prevent UI lag
- **Thread-Safe Design**: Modern concurrent architecture using Fyne v2.6

## 🛠 System Requirements

- **macOS**: 12.0+ (Apple Silicon or Intel)
- **Go**: 1.24+
- **OpenCV**: 4.11.0 (auto-installed via dependencies)
- **Memory**: 4GB+ RAM recommended
- **Storage**: 100MB+ available space

## 📦 Installation & Setup

### Prerequisites

1. **Install Go 1.24+**:
   ```bash
   # Using Homebrew (recommended)
   brew install go@1.24
   
   # Or download from https://golang.org/dl/
   # Verify installation
   go version  # Should show go1.24.x
   ```

2. **Install OpenCV**:
   ```bash
   # Install via Homebrew
   brew install opencv
   
   # Verify installation
   pkg-config --modversion opencv4  # Should show 4.x.x
   ```

3. **Install Xcode Command Line Tools** (if not already installed):
   ```bash
   xcode-select --install
   ```

### Project Setup

1. **Clone or download the project**:
   ```bash
   git clone <repository-url>
   cd advanced-image-processing
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   go mod tidy
   ```

## 🔨 Building the Application

### Build Types

#### Development Build (Debug)
For debugging, development, and troubleshooting:

```bash
# Build with full debug information
go build ./cmd/app

# Run with debug logging
./app --debug
```

**Debug build characteristics:**
- Binary name: `app` (default from directory name)
- Contains full debug symbols and stack trace information
- Larger file size (~50-100MB)
- Better error messages and debugging capabilities
- Unoptimized code for accurate debugging

#### Production Build (Optimized)
For deployment and end-user distribution:

```bash
# Build optimized release binary
go build -ldflags="-s -w" -o AdvancedImageProcessing ./cmd/app

# Run normally
./AdvancedImageProcessing
```

**Production build characteristics:**
- Binary name: `AdvancedImageProcessing` (custom name)
- Debug symbols and DWARF tables stripped
- Smaller file size (~30-50MB)
- Optimized for performance
- JSON-formatted logging

### Build Flags Explained

#### `-ldflags="-s -w"`
- **`-s`**: Strip symbol table and debug information
- **`-w`**: Strip DWARF debug information
- **Result**: Significantly reduces binary size (30-50% smaller)
- **Trade-off**: Stack traces show memory addresses instead of function names

#### `-o <name>`
- Specifies output binary name
- Without this flag, binary is named after the package directory (`app`)
- Useful for creating user-friendly executable names

### Alternative Build Locations

If you prefer organized build artifacts:

```bash
# Create build directory
mkdir -p build

# Development build
go build -o build/app-debug ./cmd/app

# Production build  
go build -ldflags="-s -w" -o build/AdvancedImageProcessing ./cmd/app

# Run from build directory
./build/AdvancedImageProcessing
```

## 🎮 Usage Guide

### Basic Workflow

1. **Load Image**: File → Open Image (⌘O)
2. **Select Region** (Optional): 
   - Rectangle tool: Click and drag to create selection
   - Freehand tool: Click multiple points, double-click to close
3. **Choose Algorithm**: Select from Binarization, Morphology, or Filters
4. **Adjust Parameters**: Fine-tune settings in Properties panel
5. **Monitor Quality**: Watch real-time metrics in Metrics panel
6. **Save Result**: File → Save Image (⌘S)

### Supported Image Formats

- **Input**: JPEG, PNG, TIFF, BMP
- **Output**: PNG (recommended for processed images)

### Algorithm Categories

#### Binarization (Local Adaptive)
- **Multi-Level Otsu**: Best for documents with multiple contrast levels
- **Local Adaptive Otsu**: Handles uneven illumination effectively
- **True Niblack**: Ideal for handwritten documents
- **True Sauvola**: Excellent for historical papers with aging
- **Wolf-Jolion**: Enhanced Sauvola for severely degraded documents
- **NICK**: Optimal for very low-contrast images

#### Morphological Operations
- **Erosion**: Removes small noise particles
- **Dilation**: Fills gaps in text characters
- **Opening**: Removes noise while preserving text structure
- **Closing**: Connects broken characters and fills holes

#### Filters
- **Gaussian**: General-purpose noise reduction
- **Median**: Removes salt-and-pepper noise specifically
- **Bilateral**: Edge-preserving smoothing

## 📊 Quality Metrics Guide

### PSNR (Peak Signal-to-Noise Ratio)
- **Range**: 0-∞ dB (higher = better)
- **Excellent**: >40 dB
- **Good**: 30-40 dB  
- **Fair**: 20-30 dB
- **Poor**: <20 dB

### SSIM (Structural Similarity Index)
- **Range**: 0-1 (higher = better)
- **Excellent**: >0.95
- **Good**: 0.8-0.95
- **Fair**: 0.6-0.8
- **Poor**: <0.6

### F-Measure (Binarization Quality)
- **Range**: 0-1 (higher = better)
- **Excellent**: >0.9
- **Good**: 0.8-0.9
- **Fair**: 0.7-0.8
- **Poor**: <0.7

## 📋 Build & Run Summary

| **Aspect** | **Debug Build** | **Production Build** |
|------------|-----------------|---------------------|
| **Build Command** | `go build ./cmd/app` | `go build -ldflags="-s -w" -o AdvancedImageProcessing ./cmd/app` |
| **Binary Name** | `app` | `AdvancedImageProcessing` |
| **File Size** | ~50-100MB | ~30-50MB |
| **Debug Info** | ✅ Full symbols & stack traces | ❌ Stripped for size |
| **Performance** | Unoptimized | ✅ Optimized |
| **Run Command** | `./app --debug` | `./AdvancedImageProcessing` |
| **Logging** | Verbose console output | JSON formatted |
| **Error Details** | ✅ Function names in traces | Memory addresses only |
| **Use Case** | Development & debugging | End-user distribution |
| **Build Flags** | None (default) | `-ldflags="-s -w"` strips debug info<br>`-o` sets custom binary name |

### Quick Commands Reference

```bash
# Development Workflow
go build ./cmd/app && ./app --debug

# Production Workflow  
go build -ldflags="-s -w" -o AdvancedImageProcessing ./cmd/app && ./AdvancedImageProcessing

# With organized build directory
mkdir -p build
go build -o build/app-debug ./cmd/app && ./build/app-debug --debug
go build -ldflags="-s -w" -o build/AdvancedImageProcessing ./cmd/app && ./build/AdvancedImageProcessing
```

## 🔧 Development

### Dependency Management

#### Understanding Go Modules
- **`go.mod`**: Defines module name, Go version, and direct dependencies
- **`go.sum`**: Contains cryptographic checksums for dependency verification
- Both files should be committed to version control

#### Common Dependency Commands

```bash
# Download dependencies (usually automatic)
go mod download

# Add missing dependencies and remove unused ones
go mod tidy

# View all dependencies (direct and indirect)
go list -m all

# Update to latest patch versions (safest - bug fixes only)
go get -u=patch ./...

# Update to latest minor versions (may include new features)
go get -u ./...

# Add a new dependency
go get github.com/example/package

# Remove a dependency (after removing import statements)
go mod tidy
```

#### Version Management
Go uses **Semantic Versioning**:
- **MAJOR** (v1 → v2): Breaking changes - **REQUIRES CODE UPDATES**
- **MINOR** (v1.1 → v1.2): New features, backward compatible
- **PATCH** (v1.1.1 → v1.1.2): Bug fixes, backward compatible

**Important**: Go does NOT automatically update major versions. You must explicitly specify them:
```bash
# Manual major version upgrade (requires code changes)
go get github.com/example/package/v2@latest
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🍎 macOS App Bundle Creation

### Using Fyne Package Tool

1. **Install fyne command**:
   ```bash
   go install fyne.io/fyne/v2/cmd/fyne@latest
   ```

2. **Create app bundle**:
   ```bash
   fyne package -os darwin -name "Advanced Image Processing" \
     -appVersion "2.0.0" -appBuild 1 \
     -appID "com.strauhmanis.advanced-image-processing" \
     ./cmd/app
   ```

3. **Run the app bundle**:
   ```bash
   open "Advanced Image Processing.app"
   ```

### Code Signing

#### Option 1: Self-Signed Certificate (Recommended for Personal Use)

1. **Create certificate in Keychain Access**:
   - Open `Keychain Access.app`
   - Go to `Keychain Access` → `Certificate Assistant` → `Create a Certificate`
   - **Name**: `Advanced Image Processing Certificate`
   - **Identity Type**: `Self Signed Root`
   - **Certificate Type**: `Code Signing`
   - **Validity Period**: 365 days (or desired duration)
   - Click `Create`

2. **Sign the application**:
   ```bash
   # Find your certificate name
   security find-identity -v -p codesigning
   
   # Sign the app bundle
   codesign --deep --force --verbose \
     --sign "Advanced Image Processing Certificate" \
     "Advanced Image Processing.app"
   
   # Verify the signature
   codesign --verify --verbose "Advanced Image Processing.app"
   spctl --assess --verbose "Advanced Image Processing.app"
   ```

#### Option 2: Ad-hoc Signing (Simplest, No Certificate Needed)
```bash
# Sign without certificate (ad-hoc signature)
codesign --deep --force --sign - "Advanced Image Processing.app"

# Remove quarantine flag to avoid Gatekeeper warnings
sudo xattr -r -d com.apple.quarantine "Advanced Image Processing.app"
```

#### Option 3: Developer ID Certificate (For Distribution)
If you have an Apple Developer account:
```bash
# Sign with Developer ID
codesign --deep --force --verbose \
  --sign "Developer ID Application: Your Name (TEAM_ID)" \
  "Advanced Image Processing.app"

# Notarize (optional, for distribution)
xcrun notarytool submit "Advanced Image Processing.app" \
  --apple-id your-apple-id@example.com \
  --password your-app-specific-password \
  --team-id YOUR_TEAM_ID
```

#### Handling Gatekeeper Warnings
If macOS shows security warnings:
```bash
# Temporarily disable Gatekeeper (not recommended for production)
sudo spctl --master-disable

# Re-enable after testing
sudo spctl --master-enable

# Or allow specific app in System Preferences:
# System Preferences → Security & Privacy → General → "Allow Anyway"
```

## 🐛 Troubleshooting

### Build Issues

1. **"cannot find package" errors**:
   ```bash
   go mod tidy
   go clean -modcache
   go mod download
   ```

2. **OpenCV not found**:
   ```bash
   # Reinstall OpenCV
   brew reinstall opencv
   
   # Check pkg-config
   pkg-config --libs opencv4
   ```

3. **CGO errors**:
   ```bash
   # Ensure CGO is enabled
   export CGO_ENABLED=1
   go env CGO_ENABLED  # Should show "1"
   ```

### Runtime Issues

1. **"Cannot open app" on macOS**:
   ```bash
   # Remove quarantine
   sudo xattr -r -d com.apple.quarantine ./AdvancedImageProcessing
   ```

2. **Performance issues**:
   - Ensure 4GB+ RAM available
   - Try smaller images for testing
   - Run with `--debug` to identify bottlenecks

3. **Dependency conflicts after updates**:
   ```bash
   # Check for breaking changes
   go list -u -m all
   
   # Revert to previous version if needed
   go get github.com/example/package@v1.2.3
   ```

## 📄 License

MIT License - see source files for details.

## 👨‍💻 Author

**Ervins Strauhmanis**

---

*Built with Go, Fyne v2.6, and OpenCV 4.11*