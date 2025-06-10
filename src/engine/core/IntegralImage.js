/**
 * @fileoverview Integral image calculations for efficient local statistics
 * @license MIT
 * @author Ervins Strauhmanis
 */

export class IntegralImage {
  constructor(imageData) {
    this.width = imageData.width;
    this.height = imageData.height;
    this.sum = new Float64Array((this.width + 1) * (this.height + 1));
    this.sqSum = new Float64Array((this.width + 1) * (this.height + 1));
    this.compute(imageData);
  }

  compute(imageData) {
    const grayImage = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    
    // Initialize first row and column to 0
    for (let i = 0; i <= this.width; i++) {
      this.sum[i] = 0;
      this.sqSum[i] = 0;
    }
    for (let i = 0; i <= this.height; i++) {
      this.sum[i * (this.width + 1)] = 0;
      this.sqSum[i * (this.width + 1)] = 0;
    }

    // Compute integral images
    for (let y = 1; y <= this.height; y++) {
      let rowSum = 0;
      let rowSqSum = 0;
      for (let x = 1; x <= this.width; x++) {
        const pixel = grayImage.data[(y - 1) * this.width + (x - 1)];
        rowSum += pixel;
        rowSqSum += pixel * pixel;
        
        const idx = y * (this.width + 1) + x;
        const idxAbove = (y - 1) * (this.width + 1) + x;
        
        this.sum[idx] = rowSum + this.sum[idxAbove];
        this.sqSum[idx] = rowSqSum + this.sqSum[idxAbove];
      }
    }
  }

  getSum(x1, y1, x2, y2) {
    // Clamp coordinates
    x1 = Math.max(0, Math.min(x1, this.width - 1));
    y1 = Math.max(0, Math.min(y1, this.height - 1));
    x2 = Math.max(x1, Math.min(x2, this.width - 1));
    y2 = Math.max(y1, Math.min(y2, this.height - 1));

    // Convert to integral image coordinates (1-indexed)
    const ix1 = x1;
    const iy1 = y1;
    const ix2 = x2 + 1;
    const iy2 = y2 + 1;

    const idx1 = iy2 * (this.width + 1) + ix2;
    const idx2 = iy1 * (this.width + 1) + ix2;
    const idx3 = iy2 * (this.width + 1) + ix1;
    const idx4 = iy1 * (this.width + 1) + ix1;

    return this.sum[idx1] - this.sum[idx2] - this.sum[idx3] + this.sum[idx4];
  }

  getSqSum(x1, y1, x2, y2) {
    // Clamp coordinates
    x1 = Math.max(0, Math.min(x1, this.width - 1));
    y1 = Math.max(0, Math.min(y1, this.height - 1));
    x2 = Math.max(x1, Math.min(x2, this.width - 1));
    y2 = Math.max(y1, Math.min(y2, this.height - 1));

    // Convert to integral image coordinates (1-indexed)
    const ix1 = x1;
    const iy1 = y1;
    const ix2 = x2 + 1;
    const iy2 = y2 + 1;

    const idx1 = iy2 * (this.width + 1) + ix2;
    const idx2 = iy1 * (this.width + 1) + ix2;
    const idx3 = iy2 * (this.width + 1) + ix1;
    const idx4 = iy1 * (this.width + 1) + ix1;

    return this.sqSum[idx1] - this.sqSum[idx2] - this.sqSum[idx3] + this.sqSum[idx4];
  }

  getMean(x1, y1, x2, y2) {
    const area = (x2 - x1 + 1) * (y2 - y1 + 1);
    if (area === 0) return 0;
    return this.getSum(x1, y1, x2, y2) / area;
  }

  getStdDev(x1, y1, x2, y2) {
    const area = (x2 - x1 + 1) * (y2 - y1 + 1);
    if (area === 0) return 0;
    
    const sum = this.getSum(x1, y1, x2, y2);
    const sqSum = this.getSqSum(x1, y1, x2, y2);
    
    const mean = sum / area;
    const variance = (sqSum / area) - (mean * mean);
    
    return Math.sqrt(Math.max(0, variance));
  }
}