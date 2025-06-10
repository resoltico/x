import { WebSocketServer, WebSocket } from 'ws';
import { ImageLoader } from '../../src/engine/utils/ImageLoader.js';
import { imageStore } from './imageStore.server';
import { processPreview } from './processing.server';
import type { ProcessingParameters } from '~/types';

interface WebSocketWithAlive extends WebSocket {
  isAlive?: boolean;
}

class WebSocketManager {
  private wss: WebSocketServer | null = null;
  private clients: Set<WebSocketWithAlive> = new Set();
  private heartbeatInterval: NodeJS.Timeout | null = null;

  initialize(server: any) {
    console.log('Initializing WebSocket server...');
    
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
      console.log('New WebSocket connection from:', request.socket.remoteAddress);
      ws.isAlive = true;
      this.clients.add(ws);

      // Send welcome message
      this.sendMessage(ws, {
        type: 'connection.established',
        payload: { message: 'Connected to Engraving Processor' }
      });

      ws.on('pong', () => {
        ws.isAlive = true;
      });

      ws.on('message', async (data: Buffer) => {
        try {
          const message = JSON.parse(data.toString());
          await this.handleMessage(ws, message);
        } catch (error) {
          console.error('WebSocket message error:', error);
          this.sendMessage(ws, {
            type: 'error',
            payload: { message: 'Invalid message format' }
          });
        }
      });

      ws.on('close', (code, reason) => {
        console.log('WebSocket connection closed:', code, reason?.toString());
        this.clients.delete(ws);
      });

      ws.on('error', (error) => {
        console.error('WebSocket client error:', error);
        this.clients.delete(ws);
      });
    });

    this.wss.on('error', (error) => {
      console.error('WebSocket server error:', error);
    });

    // Start heartbeat interval to detect broken connections
    this.heartbeatInterval = setInterval(() => {
      this.clients.forEach((ws) => {
        if (ws.isAlive === false) {
          console.log('Terminating inactive WebSocket connection');
          this.clients.delete(ws);
          return ws.terminate();
        }
        
        ws.isAlive = false;
        ws.ping();
      });
    }, 30000);

    console.log('WebSocket server initialized successfully');
  }

  private sendMessage(ws: WebSocket, message: any) {
    if (ws.readyState === WebSocket.OPEN) {
      try {
        ws.send(JSON.stringify(message));
      } catch (error) {
        console.error('Failed to send WebSocket message:', error);
      }
    }
  }

  private async handleMessage(ws: WebSocket, message: any) {
    switch (message.type) {
      case 'preview.update':
        await this.handlePreviewUpdate(ws, message.payload);
        break;
      case 'ping':
        this.sendMessage(ws, { type: 'pong' });
        break;
      default:
        console.warn('Unknown message type:', message.type);
        this.sendMessage(ws, {
          type: 'error',
          payload: { message: `Unknown message type: ${message.type}` }
        });
    }
  }

  private async handlePreviewUpdate(
    ws: WebSocket,
    payload: { imageId: string; parameters: ProcessingParameters }
  ) {
    if (!payload?.imageId || !payload?.parameters) {
      this.sendMessage(ws, {
        type: 'error',
        payload: { message: 'Invalid preview update payload' }
      });
      return;
    }

    // Send acknowledgment
    this.sendMessage(ws, {
      type: 'preview.processing',
      payload: { imageId: payload.imageId }
    });

    try {
      const storedImage = imageStore.get(payload.imageId);
      if (!storedImage) {
        this.sendMessage(ws, {
          type: 'error',
          payload: { message: 'Image not found', imageId: payload.imageId }
        });
        return;
      }

      // Create preview-sized version
      const preview = await ImageLoader.createPreview(storedImage.imageData, 512);

      // Process preview
      const result = await processPreview(preview, payload.parameters);

      // Send result
      this.sendMessage(ws, {
        type: 'preview.result',
        payload: {
          ...result,
          imageId: payload.imageId
        }
      });
    } catch (error) {
      console.error('Preview update error:', error);
      this.sendMessage(ws, {
        type: 'error',
        payload: { 
          message: 'Failed to process preview',
          error: error instanceof Error ? error.message : 'Unknown error',
          imageId: payload.imageId
        }
      });
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
    
    if (sent > 0) {
      console.log(`Broadcast sent to ${sent} clients`);
    }
  }

  close() {
    console.log('Closing WebSocket server...');
    
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }

    if (this.wss) {
      // Close all client connections gracefully
      this.clients.forEach((client) => {
        if (client.readyState === WebSocket.OPEN) {
          client.close(1001, 'Server shutting down');
        }
      });
      
      this.clients.clear();
      
      this.wss.close((err) => {
        if (err) {
          console.error('Error closing WebSocket server:', err);
        } else {
          console.log('WebSocket server closed');
        }
      });
      
      this.wss = null;
    }
  }
}

// Singleton instance
export const wsManager = new WebSocketManager();