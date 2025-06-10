import { WebSocketServer, WebSocket } from 'ws';
import { ImageLoader } from '../../src/engine/utils/ImageLoader.js';
import { imageStore } from './imageStore.server';
import { processPreview } from './processing.server';
import type { ProcessingParameters } from '~/types';

class WebSocketManager {
  private wss: WebSocketServer | null = null;
  private clients: Set<WebSocket> = new Set();

  initialize(server: any) {
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

    this.wss.on('connection', (ws: WebSocket) => {
      console.log('New WebSocket connection established');
      this.clients.add(ws);

      // Send welcome message
      ws.send(JSON.stringify({
        type: 'connection.established',
        payload: { message: 'Connected to Engraving Processor' }
      }));

      // Setup ping/pong for connection health
      const pingInterval = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.ping();
        }
      }, 30000); // 30 seconds

      ws.on('pong', () => {
        // Connection is alive
      });

      ws.on('message', async (data: Buffer) => {
        try {
          const message = JSON.parse(data.toString());
          await this.handleMessage(ws, message);
        } catch (error) {
          console.error('WebSocket message error:', error);
          ws.send(JSON.stringify({
            type: 'error',
            payload: { message: 'Invalid message format' }
          }));
        }
      });

      ws.on('close', () => {
        console.log('WebSocket connection closed');
        clearInterval(pingInterval);
        this.clients.delete(ws);
      });

      ws.on('error', (error) => {
        console.error('WebSocket error:', error);
        clearInterval(pingInterval);
        this.clients.delete(ws);
      });
    });

    this.wss.on('error', (error) => {
      console.error('WebSocket server error:', error);
    });
  }

  private async handleMessage(ws: WebSocket, message: any) {
    switch (message.type) {
      case 'preview.update':
        await this.handlePreviewUpdate(ws, message.payload);
        break;
      case 'ping':
        ws.send(JSON.stringify({ type: 'pong' }));
        break;
      default:
        console.warn('Unknown message type:', message.type);
        ws.send(JSON.stringify({
          type: 'error',
          payload: { message: `Unknown message type: ${message.type}` }
        }));
    }
  }

  private async handlePreviewUpdate(
    ws: WebSocket,
    payload: { imageId: string; parameters: ProcessingParameters }
  ) {
    // Send acknowledgment
    ws.send(JSON.stringify({
      type: 'preview.processing',
      payload: { imageId: payload.imageId }
    }));

    try {
      const storedImage = imageStore.get(payload.imageId);
      if (!storedImage) {
        ws.send(JSON.stringify({
          type: 'error',
          payload: { message: 'Image not found', imageId: payload.imageId }
        }));
        return;
      }

      // Create preview-sized version
      const preview = await ImageLoader.createPreview(storedImage.imageData, 512);

      // Process preview
      const result = await processPreview(preview, payload.parameters);

      // Send result
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'preview.result',
          payload: {
            ...result,
            imageId: payload.imageId
          }
        }));
      }
    } catch (error) {
      console.error('Preview update error:', error);
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'error',
          payload: { 
            message: 'Failed to process preview',
            error: error.message,
            imageId: payload.imageId
          }
        }));
      }
    }
  }

  broadcast(message: any) {
    const data = JSON.stringify(message);
    let sent = 0;
    
    this.clients.forEach((client) => {
      if (client.readyState === WebSocket.OPEN) {
        try {
          client.send(data);
          sent++;
        } catch (error) {
          console.error('Broadcast error:', error);
          this.clients.delete(client);
        }
      }
    });
    
    console.log(`Broadcast sent to ${sent} clients`);
  }

  close() {
    if (this.wss) {
      // Close all client connections
      this.clients.forEach((client) => {
        if (client.readyState === WebSocket.OPEN) {
          client.close(1000, 'Server shutting down');
        }
      });
      
      this.wss.close(() => {
        console.log('WebSocket server closed');
      });
    }
  }
}

// Singleton instance
export const wsManager = new WebSocketManager();