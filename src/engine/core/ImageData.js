/**
 * @fileoverview Image data wrapper class for consistent image handling
 * @license MIT
 * @author Ervins Strauhmanis
 */

export class ImageData {
  constructor(data, width, height, channels = 4) {
    this.data = data;
    this.width = width;
    this.height = height;
    this.channels = channels;
  }

  static fromRGBA(rgbaData, width, height) {
    return new ImageData(rgbaData, width, height, 4);
  }

  static fromGrayscale(grayData, width, height) {
    return new ImageData(grayData, width, height, 1);
  }

  static createEmpty(width, height, channels = 4) {
    const size = width * height * channels;
    const data = new Uint8ClampedArray(size);
    return new ImageData(data, width, height, channels);
  }

  getPixel(x, y) {
    if (x < 0 || x >= this.width || y < 0 || y >= this.height) {
      return null;
    }
    const idx = (y * this.width + x) * this.channels;
    if (this.channels === 1) {
      return this.data[idx];
    }
    return Array.from(this.data.slice(idx, idx + this.channels));
  }

  setPixel(x, y, value) {
    if (x < 0 || x >= this.width || y < 0 || y >= this.height) {
      return;
    }
    const idx = (y * this.width + x) * this.channels;
    if (typeof value === 'number' && this.channels === 1) {
      this.data[idx] = value;
    } else if (Array.isArray(value)) {
      for (let i = 0; i < Math.min(value.length, this.channels); i++) {
        this.data[idx + i] = value[i];
      }
    }
  }

  toGrayscale() {
    if (this.channels === 1) return this.clone();
    
    const grayData = new Uint8ClampedArray(this.width * this.height);
    for (let y = 0; y < this.height; y++) {
      for (let x = 0; x < this.width; x++) {
        const idx = (y * this.width + x) * this.channels;
        const gray = Math.round(
          0.299 * this.data[idx] +
          0.587 * this.data[idx + 1] +
          0.114 * this.data[idx + 2]
        );
        grayData[y * this.width + x] = gray;
      }
    }
    return new ImageData(grayData, this.width, this.height, 1);
  }

  toRGBA() {
    if (this.channels === 4) return this.clone();
    
    const rgbaData = new Uint8ClampedArray(this.width * this.height * 4);
    for (let y = 0; y < this.height; y++) {
      for (let x = 0; x < this.width; x++) {
        const srcIdx = y * this.width + x;
        const dstIdx = srcIdx * 4;
        if (this.channels === 1) {
          const gray = this.data[srcIdx];
          rgbaData[dstIdx] = gray;
          rgbaData[dstIdx + 1] = gray;
          rgbaData[dstIdx + 2] = gray;
          rgbaData[dstIdx + 3] = 255;
        } else if (this.channels === 3) {
          const srcIdx3 = srcIdx * 3;
          rgbaData[dstIdx] = this.data[srcIdx3];
          rgbaData[dstIdx + 1] = this.data[srcIdx3 + 1];
          rgbaData[dstIdx + 2] = this.data[srcIdx3 + 2];
          rgbaData[dstIdx + 3] = 255;
        }
      }
    }
    return new ImageData(rgbaData, this.width, this.height, 4);
  }

  clone() {
    return new ImageData(
      new Uint8ClampedArray(this.data),
      this.width,
      this.height,
      this.channels
    );
  }

  getHistogram() {
    const histogram = new Array(256).fill(0);
    if (this.channels === 1) {
      for (let i = 0; i < this.data.length; i++) {
        histogram[this.data[i]]++;
      }
    } else {
      // For multi-channel, compute luminance histogram
      for (let i = 0; i < this.data.length; i += this.channels) {
        const gray = Math.round(
          0.299 * this.data[i] +
          0.587 * this.data[i + 1] +
          0.114 * this.data[i + 2]
        );
        histogram[gray]++;
      }
    }
    return histogram;
  }
}