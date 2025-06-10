/**
 * @fileoverview Image loading utilities
 * @license MIT
 * @author Ervins Strauhmanis
 */

import sharp from 'sharp';
import { ImageData } from '../core/ImageData.js';

export class ImageLoader {
  static async loadFromBuffer(buffer, options = {}) {
    try {
      const image = sharp(buffer);
      const metadata = await image.metadata();
      
      // Convert to raw pixel data
      const { data, info } = await image
        .raw()
        .toBuffer({ resolveWithObject: true });
      
      // Create ImageData object
      const imageData = new ImageData(
        new Uint8ClampedArray(data),
        info.width,
        info.height,
        info.channels
      );
      
      return {
        imageData,
        metadata: {
          width: info.width,
          height: info.height,
          channels: info.channels,
          format: metadata.format,
          size: buffer.length
        }
      };
    } catch (error) {
      throw new Error(`Failed to load image: ${error.message}`);
    }
  }

  static async loadFromPath(path, options = {}) {
    try {
      const image = sharp(path);
      const metadata = await image.metadata();
      
      // Convert to raw pixel data
      const { data, info } = await image
        .raw()
        .toBuffer({ resolveWithObject: true });
      
      // Create ImageData object
      const imageData = new ImageData(
        new Uint8ClampedArray(data),
        info.width,
        info.height,
        info.channels
      );
      
      return {
        imageData,
        metadata: {
          width: info.width,
          height: info.height,
          channels: info.channels,
          format: metadata.format,
          path
        }
      };
    } catch (error) {
      throw new Error(`Failed to load image from ${path}: ${error.message}`);
    }
  }

  static async createPreview(imageData, maxSize = 512) {
    const ratio = Math.min(maxSize / imageData.width, maxSize / imageData.height);
    
    if (ratio >= 1) {
      return imageData; // No need to resize
    }
    
    const newWidth = Math.round(imageData.width * ratio);
    const newHeight = Math.round(imageData.height * ratio);
    
    // Convert to sharp buffer
    const buffer = await ImageLoader.toBuffer(imageData);
    
    // Resize using sharp
    const resized = await sharp(buffer, {
      raw: {
        width: imageData.width,
        height: imageData.height,
        channels: imageData.channels
      }
    })
      .resize(newWidth, newHeight, {
        kernel: 'lanczos3'
      })
      .raw()
      .toBuffer();
    
    return new ImageData(
      new Uint8ClampedArray(resized),
      newWidth,
      newHeight,
      imageData.channels
    );
  }

  static async toBuffer(imageData) {
    return Buffer.from(imageData.data);
  }
}