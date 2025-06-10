/**
 * @fileoverview Image saving utilities
 * @license MIT
 * @author Ervins Strauhmanis
 */

import sharp from 'sharp';
import { ImageLoader } from './ImageLoader.js';

export class ImageSaver {
  static async saveToBuffer(imageData, format = 'png', options = {}) {
    try {
      const buffer = await ImageLoader.toBuffer(imageData);
      
      let sharpInstance = sharp(buffer, {
        raw: {
          width: imageData.width,
          height: imageData.height,
          channels: imageData.channels
        }
      });
      
      // Apply format-specific options
      switch (format) {
        case 'png':
          sharpInstance = sharpInstance.png({
            compressionLevel: options.compressionLevel || 9,
            ...options
          });
          break;
        case 'jpeg':
        case 'jpg':
          sharpInstance = sharpInstance.jpeg({
            quality: options.quality || 90,
            ...options
          });
          break;
        case 'webp':
          sharpInstance = sharpInstance.webp({
            quality: options.quality || 90,
            lossless: options.lossless || false,
            ...options
          });
          break;
        case 'tiff':
          sharpInstance = sharpInstance.tiff({
            compression: options.compression || 'lzw',
            ...options
          });
          break;
        default:
          throw new Error(`Unsupported format: ${format}`);
      }
      
      return await sharpInstance.toBuffer();
    } catch (error) {
      throw new Error(`Failed to save image: ${error.message}`);
    }
  }

  static async saveToFile(imageData, path, format = null, options = {}) {
    try {
      // Detect format from path if not specified
      if (!format) {
        const ext = path.split('.').pop().toLowerCase();
        format = ext;
      }
      
      const buffer = await ImageLoader.toBuffer(imageData);
      
      let sharpInstance = sharp(buffer, {
        raw: {
          width: imageData.width,
          height: imageData.height,
          channels: imageData.channels
        }
      });
      
      // Apply format-specific options
      switch (format) {
        case 'png':
          sharpInstance = sharpInstance.png({
            compressionLevel: options.compressionLevel || 9,
            ...options
          });
          break;
        case 'jpeg':
        case 'jpg':
          sharpInstance = sharpInstance.jpeg({
            quality: options.quality || 90,
            ...options
          });
          break;
        case 'webp':
          sharpInstance = sharpInstance.webp({
            quality: options.quality || 90,
            lossless: options.lossless || false,
            ...options
          });
          break;
        case 'tiff':
          sharpInstance = sharpInstance.tiff({
            compression: options.compression || 'lzw',
            ...options
          });
          break;
        default:
          throw new Error(`Unsupported format: ${format}`);
      }
      
      await sharpInstance.toFile(path);
      
      return {
        path,
        format,
        size: (await sharpInstance.toBuffer()).length
      };
    } catch (error) {
      throw new Error(`Failed to save image to ${path}: ${error.message}`);
    }
  }

  static async toBase64(imageData, format = 'png', options = {}) {
    const buffer = await this.saveToBuffer(imageData, format, options);
    return `data:image/${format};base64,${buffer.toString('base64')}`;
  }
}