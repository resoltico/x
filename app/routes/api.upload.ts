import { type ActionFunctionArgs, json, unstable_parseMultipartFormData, unstable_createMemoryUploadHandler } from "@remix-run/node";
import { v4 as uuidv4 } from 'uuid';
import { ImageLoader } from '../../src/engine/utils/ImageLoader.js';
import { ImageSaver } from '../../src/engine/utils/ImageSaver.js';
import { imageStore } from '~/services/imageStore.server';

export async function action({ request }: ActionFunctionArgs) {
  try {
    const uploadHandler = unstable_createMemoryUploadHandler({
      maxPartSize: 10_000_000, // 10 MB
    });

    const formData = await unstable_parseMultipartFormData(
      request,
      uploadHandler
    );

    const file = formData.get('image') as File;
    if (!file) {
      return json({ error: 'No image file provided' }, { status: 400 });
    }

    // Convert File to Buffer
    const arrayBuffer = await file.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);

    // Load image using ImageLoader
    const { imageData, metadata } = await ImageLoader.loadFromBuffer(buffer);

    // Generate unique ID
    const imageId = uuidv4();

    // Store original image
    imageStore.set(imageId, {
      imageData,
      metadata,
      originalBuffer: buffer,
    });

    // Create preview (max 512px)
    const preview = await ImageLoader.createPreview(imageData, 512);
    const previewBase64 = await ImageSaver.toBase64(preview, 'png');

    return json({
      id: imageId,
      preview: previewBase64,
      metadata: {
        width: metadata.width,
        height: metadata.height,
        channels: metadata.channels,
        format: metadata.format,
      },
    });
  } catch (error) {
    console.error('Upload error:', error);
    return json(
      { error: 'Failed to process image' },
      { status: 500 }
    );
  }
}