import { type ActionFunctionArgs, json, unstable_parseMultipartFormData, unstable_createMemoryUploadHandler } from "@remix-run/node";
import { v4 as uuidv4 } from 'uuid';
import { ImageLoader } from '../../src/engine/utils/ImageLoader.js';
import { ImageSaver } from '../../src/engine/utils/ImageSaver.js';
import { imageStore } from '~/services/imageStore.server';

export async function action({ request }: ActionFunctionArgs) {
  try {
    console.log('📤 Handling image upload request...');
    
    const uploadHandler = unstable_createMemoryUploadHandler({
      maxPartSize: 10_000_000, // 10 MB
    });

    const formData = await unstable_parseMultipartFormData(
      request,
      uploadHandler
    );

    const file = formData.get('image') as File;
    if (!file) {
      console.error('❌ No image file provided in form data');
      return json({ error: 'No image file provided' }, { status: 400 });
    }

    console.log(`📁 Received file: ${file.name} (${file.type}, ${file.size} bytes)`);

    // Validate file type
    if (!file.type.startsWith('image/')) {
      console.error(`❌ Invalid file type: ${file.type}`);
      return json({ error: 'File must be an image' }, { status: 400 });
    }

    // Convert File to Buffer
    const arrayBuffer = await file.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);
    console.log(`💾 Converted to buffer: ${buffer.length} bytes`);

    // Load image using ImageLoader
    console.log('🖼️ Loading image with Sharp...');
    const { imageData, metadata } = await ImageLoader.loadFromBuffer(buffer);
    console.log(`✅ Image loaded: ${metadata.width}x${metadata.height}, ${metadata.channels} channels`);

    // Generate unique ID
    const imageId = uuidv4();
    console.log(`🆔 Generated image ID: ${imageId}`);

    // Store original image
    imageStore.set(imageId, {
      imageData,
      metadata: {
        ...metadata,
        size: file.size // Use the original file size
      },
      originalBuffer: buffer,
    });
    console.log('💾 Image stored in memory');

    // Create preview (max 512px)
    console.log('🔍 Creating preview...');
    const preview = await ImageLoader.createPreview(imageData, 512);
    const previewBase64 = await ImageSaver.toBase64(preview, 'png');
    console.log(`✅ Preview created: ${preview.width}x${preview.height}`);

    const response = {
      id: imageId,
      preview: previewBase64,
      metadata: {
        width: metadata.width,
        height: metadata.height,
        channels: metadata.channels,
        format: metadata.format,
        size: file.size, // Original file size
      },
    };

    console.log('✅ Upload complete, sending response');
    return json(response);
  } catch (error) {
    console.error('❌ Upload error:', error);
    
    // Provide more specific error messages
    let errorMessage = 'Failed to process image';
    if (error instanceof Error) {
      if (error.message.includes('unsupported image format')) {
        errorMessage = 'Unsupported image format. Please use PNG, JPEG, TIFF, or WebP.';
      } else if (error.message.includes('memory')) {
        errorMessage = 'Image is too large to process. Please use a smaller image.';
      } else {
        errorMessage = `Failed to process image: ${error.message}`;
      }
    }
    
    return json(
      { error: errorMessage },
      { status: 500 }
    );
  }
}