import { useEffect, useRef, useState } from 'react';

export function useWebSocket() {
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const reconnectDelay = 1000; // Start with 1 second

  useEffect(() => {
    let mounted = true;

    function connect() {
      if (!mounted) return;

      try {
        // Clear any existing connection
        if (wsRef.current?.readyState === WebSocket.OPEN || wsRef.current?.readyState === WebSocket.CONNECTING) {
          wsRef.current.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        console.log('Attempting WebSocket connection to:', wsUrl);
        
        const ws = new WebSocket(wsUrl);
        let pingInterval: NodeJS.Timeout | null = null;

        ws.onopen = () => {
          if (!mounted) return;
          console.log('WebSocket connected successfully');
          setIsConnected(true);
          setConnectionError(null);
          reconnectAttemptsRef.current = 0;
          
          if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
          }

          // Send periodic pings to keep connection alive
          pingInterval = setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
              ws.send(JSON.stringify({ type: 'ping' }));
            }
          }, 30000);
        };

        ws.onclose = (event) => {
          if (!mounted) return;
          console.log('WebSocket disconnected:', event.code, event.reason);
          setIsConnected(false);
          wsRef.current = null;

          // Clear ping interval
          if (pingInterval) {
            clearInterval(pingInterval);
            pingInterval = null;
          }

          // Attempt to reconnect if not a normal closure
          if (event.code !== 1000 && event.code !== 1001 && reconnectAttemptsRef.current < maxReconnectAttempts) {
            reconnectAttemptsRef.current++;
            const delay = Math.min(reconnectDelay * reconnectAttemptsRef.current, 5000);
            
            console.log(`Reconnecting in ${delay}ms... (attempt ${reconnectAttemptsRef.current}/${maxReconnectAttempts})`);
            setConnectionError(`Connection lost. Reconnecting... (${reconnectAttemptsRef.current}/${maxReconnectAttempts})`);
            
            reconnectTimeoutRef.current = setTimeout(() => {
              if (mounted) connect();
            }, delay);
          } else if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
            setConnectionError('Unable to connect to server. Please check your connection and refresh the page.');
          }
        };

        ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          if (!mounted) return;
          
          // Don't update error message if we're already showing a reconnection message
          if (!connectionError?.includes('Reconnecting')) {
            setConnectionError('Connection error. Please check your network.');
          }
        };

        ws.onmessage = (event) => {
          if (!mounted) return;
          
          try {
            const data = JSON.parse(event.data);
            
            // Handle system messages
            if (data.type === 'connection.established') {
              console.log('Server acknowledged connection:', data.payload.message);
            } else if (data.type === 'pong') {
              // Server ponged our ping
            } else if (data.type === 'error') {
              console.error('Server error:', data.payload.message);
              if (data.payload.message !== 'Invalid message format') {
                setConnectionError(`Server error: ${data.payload.message}`);
              }
            }
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        wsRef.current = ws;
      } catch (error) {
        console.error('Failed to create WebSocket:', error);
        if (!mounted) return;
        
        setConnectionError('Failed to establish connection');
        
        // Retry connection
        if (reconnectAttemptsRef.current < maxReconnectAttempts) {
          reconnectAttemptsRef.current++;
          const delay = Math.min(reconnectDelay * reconnectAttemptsRef.current, 5000);
          reconnectTimeoutRef.current = setTimeout(() => {
            if (mounted) connect();
          }, delay);
        }
      }
    }

    // Initial connection with a small delay to ensure server is ready
    const initialDelay = setTimeout(() => {
      if (mounted) connect();
    }, 100);

    return () => {
      mounted = false;
      
      if (initialDelay) {
        clearTimeout(initialDelay);
      }
      
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      
      if (wsRef.current) {
        wsRef.current.close(1000, 'Component unmounted');
      }
    };
  }, []);

  return {
    ws: wsRef.current,
    isConnected,
    connectionError
  };
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private messageHandlers: Map<string, (data: any) => void> = new Map();
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private url: string;
  private isDestroyed = false;

  constructor(url: string) {
    this.url = url;
  }

  connect() {
    if (this.isDestroyed) return;

    try {
      if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
        return; // Already connected or connecting
      }

      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        if (this.isDestroyed) return;
        console.log('WebSocketClient connected');
        if (this.reconnectTimeout) {
          clearTimeout(this.reconnectTimeout);
          this.reconnectTimeout = null;
        }
      };

      this.ws.onmessage = (event) => {
        if (this.isDestroyed) return;
        try {
          const message = JSON.parse(event.data);
          const handler = this.messageHandlers.get(message.type);
          if (handler) {
            handler(message.payload);
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocketClient error:', error);
      };

      this.ws.onclose = () => {
        if (this.isDestroyed) return;
        console.log('WebSocketClient disconnected');
        this.ws = null;
        
        // Attempt to reconnect after 3 seconds
        this.reconnectTimeout = setTimeout(() => {
          if (!this.isDestroyed) this.connect();
        }, 3000);
      };
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
    }
  }

  on(type: string, handler: (data: any) => void) {
    this.messageHandlers.set(type, handler);
  }

  send(type: string, payload: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    } else {
      console.warn('WebSocket not connected, message not sent:', type);
    }
  }

  close() {
    this.isDestroyed = true;
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
    if (this.ws) {
      this.ws.close(1000, 'Client closing connection');
    }
  }
}