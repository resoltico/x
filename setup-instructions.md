# Setup Instructions

## Create Public Directory Structure

1. Create a `public` directory in your project root if it doesn't exist:
```bash
mkdir -p public
```

2. Create a favicon.ico file. You can either:
   - Use an online favicon generator to convert the SVG I provided to ICO format
   - Or use the SVG directly by creating `public/favicon.ico` as a symlink to `public/favicon.svg`
   - Or use this simple command to create a basic favicon:
   ```bash
   # On macOS/Linux
   echo -e '\x00\x00\x01\x00\x01\x00\x10\x10\x00\x00\x01\x00\x20\x00\x68\x04\x00\x00\x16\x00\x00\x00' > public/favicon.ico
   ```

3. Save the SVG favicon I created as `public/favicon.svg`

## File Structure

Your project should have this structure:
```
engraving-processor-pro/
├── public/
│   ├── favicon.ico
│   └── favicon.svg
├── app/
├── src/
├── build/           (created after running build)
├── server.js
└── package.json
```

## Build and Run

1. Clean previous builds:
```bash
rm -rf build/
```

2. Build the application:
```bash
pnpm build
```

3. Start the server:
```bash
pnpm start
```

## What Was Fixed

1. **Server Static Assets**: The server now properly serves built client assets from `build/client/assets/`
2. **Drag and Drop**: 
   - Added better visual feedback with scaling and shadow effects
   - Fixed drag leave detection to prevent flickering
   - Added clear drop zone boundaries
   - Shows different states: idle, dragging, uploading, success
3. **File Upload**: 
   - Made it clear that files auto-upload on selection
   - Added upload progress indicator
   - Added success feedback
   - Added error handling and display
4. **WebSocket**: Added better error handling and retry logic
5. **Processing**: 
   - Added timeout to prevent infinite polling
   - Better error messages
   - Disabled process button until preview is ready
6. **UI Improvements**:
   - Added file size display
   - Added error alerts
   - Better button states
   - Improved transitions and animations

The drag-and-drop area now has a clear 3px dashed border that changes to blue when dragging, and the entire area is clickable. Files are automatically uploaded when selected or dropped - no separate load button is needed.