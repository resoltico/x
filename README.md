# Engraving Processor Pro

Application for processing historical engravings and document scans with advanced image algorithms. Built with Vue 3, TypeScript, and WebAssembly for high-performance local processing.

## 🚀 Features

- **Advanced Image Processing**: Binarization, morphological operations, noise reduction, and pixel art scaling
- **High Performance**: WebAssembly-powered processing with Web Workers for non-blocking operations
- **Interactive UI**: Real-time preview, zoom/pan controls, and parameter adjustment
- **Modern Architecture**: Modular design optimized for AI-driven development and extension
- **Zero-Dependency Processing**: All processing happens locally in the browser

## 🛠️ Tech Stack

- **Frontend**: Vue 3 (Composition API), TypeScript, Tailwind CSS
- **Processing**: wasm-vips, Custom WASM modules
- **Build Tools**: Vite, ESLint, Prettier
- **Testing**: Vitest, Vue Test Utils
- **State Management**: Pinia

## 📦 Installation

```bash
# Clone the repository
git clone <repository-url>
cd engraving-processor-pro

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm test
```

## 🏗️ Project Structure

```
src/
├── components/           # Vue components
│   ├── ImageInput.vue           # File upload and drag-drop
│   ├── PreviewRenderer.vue      # Canvas-based image preview
│   ├── ProcessingControls.vue   # Algorithm parameter controls
│   └── ProgressDisplay.vue      # Task progress and status
├── modules/              # Core processing modules
│   ├── ImageInputModule.ts      # File handling and validation
│   ├── PreviewRendererModule.ts # Canvas rendering and interaction
│   ├── ProcessingModule.ts      # Image processing algorithms
│   └── WorkerOrchestratorModule.ts # Web Worker management
├── stores/               # Pinia state management
│   └── app.ts                   # Main application store
├── types/                # TypeScript type definitions
│   └── index.ts                 # Core types and interfaces
├── utils/                # Utility functions
│   ├── fileValidation.ts        # File validation helpers
│   └── imageHelpers.ts          # Image manipulation utilities
├── workers/              # Web Workers
│   └── imageProcessingWorker.ts # Background image processing
└── test/                 # Test setup and utilities
    └── setup.ts                 # Test environment configuration
```

## 🔧 Processing Algorithms

### Binarization
- **Otsu**: Global thresholding for clear bimodal images
- **Sauvola**: Adaptive thresholding for varying lighting
- **Niblack**: Local adaptive thresholding

### Morphological Operations
- **Opening**: Noise removal
- **Closing**: Gap filling
- **Erosion**: Object shrinking
- **Dilation**: Object expansion

### Noise Reduction
- **Median Filter**: Salt-and-pepper noise removal
- **Binary Noise Removal**: Small component elimination

### Scaling
- **Scale2x/3x/4x**: Pixel art scaling algorithms
- **Nearest Neighbor**: Sharp edge preservation
- **Bilinear**: Smooth scaling

## 🎮 Usage

1. **Upload Image**: Drag and drop or click to browse (PNG, JPEG, TIFF, WebP, max 10MB)
2. **Select Algorithm**: Choose from binarization, morphology, noise reduction, or scaling
3. **Adjust Parameters**: Fine-tune algorithm settings with real-time sliders
4. **Preview**: Generate low-resolution preview for quick feedback
5. **Process**: Run full-resolution processing
6. **Download**: Export processed results

### Keyboard Shortcuts
- `Space`: Toggle between original and processed view
- `F`: Fit image to canvas
- `1`: Zoom to actual size (100%)
- `R`: Reset view

## 🧪 Testing

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Type checking
npm run typecheck

# Linting
npm run lint
npm run lint:fix
```

## 🔌 Plugin System

The application supports a modular plugin architecture for extending processing capabilities:

```typescript
interface Plugin {
  name: string
  version: string
  description: string
  parameters: ControlParameter[]
  process: (data: ArrayBuffer, params: any) => Promise<ArrayBuffer>
}
```

## 📈 Performance

- **Web Workers**: Non-blocking processing using all available CPU cores
- **WebAssembly**: Native-speed image processing algorithms
- **Memory Efficient**: Transferable objects for zero-copy data transfer
- **Streaming**: Chunk-based processing for large images

## 🌐 Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

Requires WebAssembly and Web Workers support.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Vue.js Documentation](https://vuejs.org/)
- [wasm-vips Documentation](https://www.npmjs.com/package/wasm-vips)
- [Tailwind CSS Documentation](https://tailwindcss.com/)
- [Vite Documentation](https://vitejs.dev/)

## 🙏 Acknowledgments

- Built with [wasm-vips](https://github.com/kleisauke/wasm-vips) for high-performance image processing
- UI components styled with [Tailwind CSS](https://tailwindcss.com/)
- Icons from [Heroicons](https://heroicons.com/)

---

Made with ❤️ for historical document preservation and image processing enthusiasts.