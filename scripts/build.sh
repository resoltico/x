#!/bin/bash

# Build script for Engraving Processor Pro
# This script handles the build process with proper error handling

set -e  # Exit on any error

echo "🚀 Starting build process for Engraving Processor Pro..."

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    pnpm install
fi

# Type checking
echo "🔍 Running type checks..."
pnpm run typecheck

# Linting
echo "🧹 Running linter..."
pnpm run lint

# Build
echo "🏗️  Building for production..."
pnpm run build

# Check build output
if [ -d "dist" ]; then
    echo "✅ Build completed successfully!"
    echo "📊 Build statistics:"
    echo "   📁 Total files: $(find dist -type f | wc -l)"
    echo "   📦 Total size: $(du -sh dist | cut -f1)"
    echo "   🎯 Main bundle: $(find dist/assets -name "index-*.js" -exec ls -lh {} \; | awk '{print $5}')"
    echo "   🎨 CSS bundle: $(find dist/assets -name "index-*.css" -exec ls -lh {} \; | awk '{print $5}')"
    
    # Check for WASM files
    wasm_files=$(find dist -name "*.wasm" | wc -l)
    if [ $wasm_files -gt 0 ]; then
        echo "   🧮 WASM files: $wasm_files"
    fi
else
    echo "❌ Build failed - no dist directory found"
    exit 1
fi

echo "🎉 Build process completed successfully!"