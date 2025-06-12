// Global type extensions for browser APIs

declare global {
  interface Performance {
    memory?: {
      usedJSHeapSize: number
      totalJSHeapSize: number
      jsHeapSizeLimit: number
    }
  }
  
  interface Navigator {
    hardwareConcurrency?: number
  }
}

export {}