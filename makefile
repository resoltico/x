# Image Restoration Suite Makefile
.PHONY: build run clean test profile

# Default target
build:
	go build -ldflags="-s -w" -o image-restoration-suite .

# Build with Mat memory profiling enabled
profile:
	go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite .

# Run with profiling
run-profile:
	go run -tags matprofile .

# Regular run
run:
	go run .

# Clean build artifacts
clean:
	rm -f image-restoration-suite
	rm -f image-restoration-suite.exe

# Test with profiling
test:
	go test -tags matprofile ./...

# Cross-platform builds with profiling
build-windows:
	GOOS=windows GOARCH=amd64 go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite.exe .

build-macos:
	GOOS=darwin GOARCH=amd64 go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite-mac .

build-linux:
	GOOS=linux GOARCH=amd64 go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite-linux .

# Universal binary for macOS
build-macos-universal:
	GOOS=darwin GOARCH=arm64 go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite-arm64 .
	GOOS=darwin GOARCH=amd64 go build -tags matprofile -ldflags="-s -w" -o image-restoration-suite-x86_64 .
	lipo -create -output image-restoration-suite image-restoration-suite-arm64 image-restoration-suite-x86_64
	rm image-restoration-suite-arm64 image-restoration-suite-x86_64

# Check for memory leaks with profiling
check-leaks:
	@echo "Running with Mat profiling enabled..."
	@echo "Check terminal output for Mat creation/cleanup tracking"
	go run -tags matprofile .

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Vet code
vet:
	go vet ./...