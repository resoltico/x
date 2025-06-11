import { WebSocketServer, WebSocket } from 'ws';
import { ImageLoader } from '../../src/engine/utils/ImageLoader.js';
import { imageStore } from './imageStore.server';
import { processPreview } from './processing.server';
import type { ProcessingParameters } from '~/types';

interface WebSocketWithAlive extends WebSocket {
  isAlive?: boolean;
  clientId?: string;
}

class WebSocketManager {
  private wss: WebSocketServer | null = null;
  private clients: Map<string, WebSocketWithAlive> = new Map();
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private clientCounter = 0;
  
  // Processing queue to prevent overwhelming the server
  private processingQueue: Map<string, NodeJS.Timeout> = new Map();

  initialize(server: any) {
    console.log('🔌 Initializing WebSocket server...');
    
    try {
      this.wss = new WebSocketServer({ 
        server, 
        path: '/ws',
        perMessageDeflate: {
          zlibDeflateOptions: {
            chunkSize: 1024,
            memLevel: 7,
            level: 3
          },
          zlibInflateOptions: {
            chunkSize: 10 * 1024
          },
          clientNoContextTakeover: true,
          serverNoContextTakeover: true,
          serverMaxWindowBits: 10,
          concurrencyLimit: 10,
          threshold: 1024
        }
      });

      this.wss.on('connection', (ws: WebSocketWithAlive, request) => {
        const clientId = `client-${++this.clientCounter}`;
        ws.clientId = clientId;
        ws.isAlive = true;
        this.clients.set(clientId, ws);
        
        console.log(`✅ New WebSocket connection: ${clientId} from ${request.socket.remoteAddress}`);
        console.log(`📊 Active connections: ${this.clients.size}`);

        // Send welcome message
        this.sendMessage(ws, {
          type: 'connection.established',
          payload: { 
            message: 'Connected to Engraving Processor',
            clientId,
            serverTime: new Date().toISOString()
          }
        });

        // Setup event handlers
        ws.on('pong', () => {
          ws.isAlive = true;
        });

        ws.on('ping', () => {
          ws.pong();
        });

        ws.on('message', async (data: Buffer) => {
          try {
            const message = JSON.parse(data.toString());
            console.log(`📨 Message from ${clientId}:`, message.type);
            await this.handleMessage(ws, message);
          } catch (error) {
            console.error(`❌ WebSocket message error from ${clientId}:`, error);
            this.sendMessage(ws, {
              type: 'error',
              payload: { 
                message: 'Invalid message format',
                error: error instanceof Error ? error.message : 'Unknown error'
              }
            });
          }
        });

        ws.on('close', (code, reason) => {
          console.log(`👋 WebSocket connection closed: ${clientId} (code: ${code}, reason: ${reason?.toString() || 'none'})`);
          this.clients.delete(clientId);
          
          // Clear any pending processing for this client
          if (this.processingQueue.has(clientId)) {
            clearTimeout(this.processingQueue.get(clientId)!);
            this.processingQueue.delete(clientId);
          }
          
          console.log(`📊 Active connections: ${this.clients.size}`);
        });

        ws.on('error', (error) => {
          console.error(`❌ WebSocket client error (${clientId}):`, error);
          this.clients.delete(clientId);
        });
      });

      this.wss.on('error', (error) => {
        console.error('❌ WebSocket server error:', error);
      });

      // Start heartbeat interval to detect broken connections
      this.heartbeatInterval = setInterval(() => {
        const deadClients: string[] = [];
        
        this.clients.forEach((ws, clientId) => {
          if (ws.isAlive === false) {
            console.log(`💔 Terminating inactive connection: ${clientId}`);
            deadClients.push(clientId);
            ws.terminate();
            return;
          }
          
          ws.isAlive = false;
          ws.ping();
        });
        
        // Clean up dead clients
        deadClients.forEach(clientId => this.clients.delete(clientId));
      }, 30000);

      console.log('✅ WebSocket server initialized successfully');
      console.log(`🔗 WebSocket endpoint: ws://localhost:${(server.address() as any)?.port || 3000}/ws`);
      
    } catch (error) {
      console.error('❌ Failed to initialize WebSocket server:', error);
      throw error;
    }
  }

  private sendMessage(ws: WebSocket, message: any) {
    if (ws.readyState === WebSocket.OPEN) {
      try {
        const data = JSON.stringify(message);
        ws.send(data);
        console.log(`📤 Sent ${message.type} to client`);
      } catch (error) {
        console.error('❌ Failed to send WebSocket message:', error);
      }
    } else {
      console.warn(`⚠️ Cannot send message, WebSocket not open (state: ${ws.readyState})`);
    }
  }

  private async handleMessage(ws: WebSocketWithAlive, message: any) {
    const clientId = ws.clientId || 'unknown';
    
    switch (message.type) {
      case 'preview.update':
        // Debounce preview updates per client
        if (this.processingQueue.has(clientId)) {
          clearTimeout(this.processingQueue.get(clientId)!);
        }
        
        const timeout = setTimeout(() => {
          this.handlePreviewUpdate(ws, message.payload);
          this.processingQueue.delete(clientId);
        }, 100);
        
        this.processingQueue.set(clientId, timeout);
        break;
        
      case 'ping':
        this.sendMessage(ws, { type: 'pong', payload: { timestamp: Date.now() } });
        break;
        
      default:
        console.warn(`⚠️ Unknown message type: ${message.type}`);
        this.sendMessage(ws, {
          type: 'error',
          payload: { 
            message: `Unknown message type: ${message.type}`,
            supportedTypes: ['preview.update', 'ping']
          }
        });
    }
  }

  private async handlePreviewUpdate(
    ws: WebSocketWithAlive,
    payload: { imageId: string; parameters: ProcessingParameters }
  ) {
    const clientId = ws.clientId || 'unknown';
    console.log(`🎨 Processing preview update for ${clientId}, image: ${payload?.imageId}`);
    
    if (!payload?.imageId || !payload?.parameters) {
      this.sendMessage(ws, {
        type: 'error',
        payload: { 
          message: 'Invalid preview update payload',
          required: ['imageId', 'parameters']
        }
      });
      return;
    }

    // Send acknowledgment
    this.sendMessage(ws, {
      type: 'preview.processing',
      payload: { 
        imageId: payload.imageId,
        timestamp: Date.now()
      }
    });

    const startTime = Date.now();

    try {
      const storedImage = imageStore.get(payload.imageId);
      if (!storedImage) {
        console.error(`❌ Image not found: ${payload.imageId}`);
        this.sendMessage(ws, {
          type: 'error',
          payload: { 
            message: 'Image not found',
            imageId: payload.imageId,
            suggestion: 'Please re-upload the image'
          }
        });
        return;
      }

      console.log(`📐 Creating preview for image ${payload.imageId} (${storedImage.metadata.width}x${storedImage.metadata.height})`);

      // Create preview-sized version
      const preview = await ImageLoader.createPreview(storedImage.imageData, 512);
      console.log(`📐 Preview size: ${preview.width}x${preview.height}`);

      // Process preview
      console.log(`⚙️ Processing with parameters:`, JSON.stringify(payload.parameters));
      const result = await processPreview(preview, payload.parameters);
      
      const processingTime = Date.now() - startTime;
      console.log(`✅ Preview processed in ${processingTime}ms`);

      // Send result
      this.sendMessage(ws, {
        type: 'preview.result',
        payload: {
          ...result,
          imageId: payload.imageId,
          processingTime
        }
      });
    } catch (error) {
      const processingTime = Date.now() - startTime;
      console.error(`❌ Preview update error after ${processingTime}ms:`, error);
      
      this.sendMessage(ws, {
        type: 'error',
        payload: { 
          message: 'Failed to process preview',
          error: error instanceof Error ? error.message : 'Unknown error',
          imageId: payload.imageId,
          processingTime,
          suggestion: 'Try adjusting the parameters or re-uploading the image'
        }
      });
    }
  }

  broadcast(message: any) {
    const data = JSON.stringify(message);
    let sent = 0;
    let failed = 0;
    
    this.clients.forEach((client, clientId) => {
      if (client.readyState === WebSocket.OPEN) {
        try {
          client.send(data);
          sent++;
        } catch (error) {
          console.error(`❌ Broadcast error for ${clientId}:`, error);
          failed++;
        }
      }
    });
    
    if (sent > 0 || failed > 0) {
      console.log(`📡 Broadcast: sent to ${sent} clients, ${failed} failed`);
    }
  }

  getConnectionCount(): number {
    return this.clients.size;
  }

  getActiveClients(): string[] {
    return Array.from(this.clients.keys());
  }

  close() {
    console.log('🛑 Closing WebSocket server...');
    
    // Clear heartbeat interval
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }

    // Clear processing queue
    this.processingQueue.forEach(timeout => clearTimeout(timeout));
    this.processingQueue.clear();

    if (this.wss) {
      // Close all client connections gracefully
      this.clients.forEach((client, clientId) => {
        if (client.readyState === WebSocket.OPEN) {
          this.sendMessage(client, {
            type: 'server.shutdown',
            payload: { message: 'Server is shutting down' }
          });
          client.close(1001, 'Server shutting down');
        }
        console.log(`👋 Closed connection: ${clientId}`);
      });
      
      this.clients.clear();
      
      this.wss.close((err) => {
        if (err) {
          console.error('❌ Error closing WebSocket server:', err);
        } else {
          console.log('✅ WebSocket server closed');
        }
      });
      
      this.wss = null;
    }
  }
}

// Create singleton instance
const wsManager = new WebSocketManager();

// Ensure the export is not tree-shaken by making it explicit
if (typeof module !== 'undefined' && module.exports) {
  module.exports.wsManager = wsManager;
}

// Export for ES modules
export { wsManager };