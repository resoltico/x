import { WebSocketServer, WebSocket } from 'ws';
import { ImageLoader } from '../../src/engine/utils/ImageLoader.js';
import { imageStore } from './imageStore.server';
import { processPreview } from './processing.server';
import type { ProcessingParameters } from '~/types';

class WebSocketManager {
  private wss: WebSocketServer | null = null;
  private clients: Set<WebSocket> = new Set();

  initialize(server: any) {
    this.wss = new WebSocketServer({ server, path: '/ws' });

    this.wss.on('connection', (ws: WebSocket) => {
      console.log('New WebSocket connection');
      this.clients.add(ws);

      ws.on('message', async (data: Buffer) => {
        try {
          const message = JSON.parse(data.toString());
          await this.handleMessage(ws, message);
        } catch (error) {
          console.error('WebSocket message error:', error);
        }
      });

      ws.on('close', () => {
        console.log('WebSocket connection closed');
        this.clients.delete(ws);
      });

      ws.on('error', (error) => {
        console.error('WebSocket error:', error);
        this.clients.delete(ws);
      });
    });
  }

  private async handleMessage(ws: WebSocket, message: any) {
    switch (message.type) {
      case 'preview.update':
        await this.handlePreviewUpdate(ws, message.payload);
        break;
      default:
        console.warn('Unknown message type:', message.type);
    }
  }

  private async handlePreviewUpdate(
    ws: WebSocket,
    payload: { imageId: string; parameters: ProcessingParameters }
  ) {
    try {
      const storedImage = imageStore.get(payload.imageId);
      if (!storedImage) {
        ws.send(JSON.stringify({
          type: 'error',
          payload: { message: 'Image not found' },
        }));
        return;
      }

      // Create preview-sized version
      const preview = await ImageLoader.createPreview(storedImage.imageData, 512);

      // Process preview
      const result = await processPreview(preview, payload.parameters);

      // Send result
      ws.send(JSON.stringify({
        type: 'preview.result',
        payload: result,
      }));
    } catch (error) {
      console.error('Preview update error:', error);
      ws.send(JSON.stringify({
        type: 'error',
        payload: { message: 'Failed to process preview' },
      }));
    }
  }

  broadcast(message: any) {
    const data = JSON.stringify(message);
    this.clients.forEach((client) => {
      if (client.readyState === WebSocket.OPEN) {
        client.send(data);
      }
    });
  }

  close() {
    if (this.wss) {
      this.wss.close();
    }
  }
}

// Singleton instance
export const wsManager = new WebSocketManager();