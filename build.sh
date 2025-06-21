#!/bin/bash

# Image Restoration Suite Build Script
# Builds with Mat profiling enabled for memory leak detection

set -e

APP_NAME="Image Restoration Suite"
BINARY_NAME="image-restoration-suite"

echo "Building ${APP_NAME} with Mat profiling enabled..."

# Clean previous builds
echo "Cleaning previous builds..."
rm -f ${BINARY_NAME}*
rm -f *.app

# Build with Mat profiling
echo "Building with -tags matprofile for memory leak detection..."
go build -tags matprofile -ldflags="-s -w" -o ${BINARY_NAME} .

if [ $? -eq 0 ]; then
    echo "✅ Build successful: ${BINARY_NAME}"
    echo "📊 Mat profiling enabled - memory leaks will be tracked"
    echo ""
    echo "To run with memory leak detection:"
    echo "  ./${BINARY_NAME}"
    echo ""
    echo "To build without profiling (production):"
    echo "  go build -ldflags='-s -w' -o ${BINARY_NAME} ."
    echo ""
    echo "Memory profiler endpoints available when running:"
    echo "- Main pprof: http://localhost:6060/debug/pprof/"
    echo "- Mat profile: http://localhost:6060/debug/pprof/gocv.io/x/gocv.Mat"
else
    echo "❌ Build failed"
    exit 1
fi