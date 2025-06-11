import { useEffect, useRef, useState } from 'react';

export function useWebSocket() {
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const reconnectDelay = 1000; // Start with 1 second
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    
    function connect() {
      if (!mountedRef.current) return;

      try {
        // Clear any existing connection
        if (wsRef.current) {
          const state = wsRef.current.readyState;
          if (state === WebSocket.OPEN || state === WebSocket.CONNECTING) {
            console.log('🔌 Closing existing WebSocket connection...');
            wsRef.current.close();
          }
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        console.log('🔌 Attempting WebSocket connection to:', wsUrl);
        
        const ws = new WebSocket(wsUrl);
        let pingInterval: NodeJS.Timeout | null = null;
        let connectionTimeout: NodeJS.Timeout | null = null;

        // Set connection timeout
        connectionTimeout = setTimeout(() => {
          if (ws.readyState === WebSocket.CONNECTING) {
            console.error('⏱️ WebSocket connection timeout');
            ws.close();
            handleReconnect();
          }
        }, 5000);

        ws.onopen = () => {
          if (!mountedRef.current) return;
          
          // Clear connection timeout
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            connectionTimeout = null;
          }

          console.log('✅ WebSocket connected successfully');
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
              console.log('🏓 Sending ping...');
              ws.send(JSON.stringify({ type: 'ping' }));
            }
          }, 30000);
        };

        ws.onclose = (event) => {
          if (!mountedRef.current) return;
          
          // Clear connection timeout
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            connectionTimeout = null;
          }

          console.log(`🔌 WebSocket disconnected: code=${event.code}, reason=${event.reason || 'none'}`);
          setIsConnected(false);
          wsRef.current = null;

          // Clear ping interval
          if (pingInterval) {
            clearInterval(pingInterval);
            pingInterval = null;
          }

          // Handle different close codes
          if (event.code === 1000 || event.code === 1001) {
            // Normal closure
            console.log('👋 WebSocket closed normally');
            setConnectionError(null);
          } else if (event.code === 1006) {
            // Abnormal closure
            console.error('❌ WebSocket closed abnormally');
            handleReconnect();
          } else {
            // Other closures
            handleReconnect();
          }
        };

        ws.onerror = (error) => {
          console.error('❌ WebSocket error:', error);
          if (!mountedRef.current) return;
          
          // Clear connection timeout
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            connectionTimeout = null;
          }
          
          // Don't update error message if we're already showing a reconnection message
          if (!connectionError?.includes('Reconnecting')) {
            setConnectionError('Connection error. Please check your network.');
          }
        };

        ws.onmessage = (event) => {
          if (!mountedRef.current) return;
          
          try {
            const data = JSON.parse(event.data);
            
            // Handle system messages
            switch (data.type) {
              case 'connection.established':
                console.log('🤝 Server acknowledged connection:', data.payload);
                break;
              case 'pong':
                console.log('🏓 Received pong');
                break;
              case 'error':
                console.error('⚠️ Server error:', data.payload);
                if (data.payload.message !== 'Invalid message format') {
                  setConnectionError(`Server error: ${data.payload.message}`);
                }
                break;
              case 'server.shutdown':
                console.log('🛑 Server is shutting down');
                setConnectionError('Server is shutting down. Please refresh the page.');
                break;
            }
          } catch (error) {
            console.error('❌ Failed to parse WebSocket message:', error);
          }
        };

        wsRef.current = ws;
      } catch (error) {
        console.error('❌ Failed to create WebSocket:', error);
        if (!mountedRef.current) return;
        
        setConnectionError('Failed to establish connection');
        handleReconnect();
      }
    }

    function handleReconnect() {
      if (!mountedRef.current) return;
      
      if (reconnectAttemptsRef.current < maxReconnectAttempts) {
        reconnectAttemptsRef.current++;
        const delay = Math.min(reconnectDelay * Math.pow(1.5, reconnectAttemptsRef.current - 1), 5000);
        
        console.log(`🔄 Reconnecting in ${delay}ms... (attempt ${reconnectAttemptsRef.current}/${maxReconnectAttempts})`);
        setConnectionError(`Connection lost. Reconnecting... (${reconnectAttemptsRef.current}/${maxReconnectAttempts})`);
        
        reconnectTimeoutRef.current = setTimeout(() => {
          if (mountedRef.current) connect();
        }, delay);
      } else {
        console.error('❌ Max reconnection attempts reached');
        setConnectionError('Unable to connect to server. Please check your connection and refresh the page.');
      }
    }

    // Initial connection with a small delay to ensure server is ready
    const initialDelay = setTimeout(() => {
      if (mountedRef.current) {
        console.log('🚀 Initiating WebSocket connection...');
        connect();
      }
    }, 100);

    return () => {
      mountedRef.current = false;
      console.log('🧹 Cleaning up WebSocket connection...');
      
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

  // Provide a manual reconnect function
  const reconnect = () => {
    console.log('🔄 Manual reconnection requested');
    reconnectAttemptsRef.current = 0;
    
    if (wsRef.current) {
      wsRef.current.close();
    }
    
    setConnectionError('Reconnecting...');
    setTimeout(() => {
      if (mountedRef.current) {
        const connect = () => {
          // Re-implement connect logic here or trigger useEffect
          window.location.reload();
        };
        connect();
      }
    }, 100);
  };

  return {
    ws: wsRef.current,
    isConnected,
    connectionError,
    reconnect
  };
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private messageHandlers: Map<string, (data: any) => void> = new Map();
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private url: string;
  private isDestroyed = false;
  private connectPromise: Promise<void> | null = null;

  constructor(url: string) {
    this.url = url;
  }

  async connect(): Promise<void> {
    if (this.isDestroyed) return;
    
    // Return existing connection promise if connecting
    if (this.connectPromise) return this.connectPromise;

    this.connectPromise = new Promise((resolve, reject) => {
      try {
        if (this.ws?.readyState === WebSocket.OPEN) {
          resolve();
          return;
        }
        
        if (this.ws?.readyState === WebSocket.CONNECTING) {
          // Wait for existing connection
          const checkConnection = setInterval(() => {
            if (this.ws?.readyState === WebSocket.OPEN) {
              clearInterval(checkConnection);
              resolve();
            } else if (this.ws?.readyState === WebSocket.CLOSED) {
              clearInterval(checkConnection);
              this.connect().then(resolve).catch(reject);
            }
          }, 100);
          return;
        }

        console.log('🔌 WebSocketClient connecting to:', this.url);
        this.ws = new WebSocket(this.url);

        const connectionTimeout = setTimeout(() => {
          if (this.ws?.readyState === WebSocket.CONNECTING) {
            this.ws.close();
            reject(new Error('Connection timeout'));
          }
        }, 5000);

        this.ws.onopen = () => {
          if (this.isDestroyed) return;
          clearTimeout(connectionTimeout);
          console.log('✅ WebSocketClient connected');
          if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
          }
          resolve();
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
            console.error('❌ Failed to parse WebSocket message:', error);
          }
        };

        this.ws.onerror = (error) => {
          clearTimeout(connectionTimeout);
          console.error('❌ WebSocketClient error:', error);
          reject(error);
        };

        this.ws.onclose = () => {
          if (this.isDestroyed) return;
          console.log('🔌 WebSocketClient disconnected');
          this.ws = null;
          this.connectPromise = null;
          
          // Attempt to reconnect after 3 seconds
          this.reconnectTimeout = setTimeout(() => {
            if (!this.isDestroyed) this.connect().catch(console.error);
          }, 3000);
        };
      } catch (error) {
        console.error('❌ Failed to create WebSocket:', error);
        reject(error);
      }
    });

    return this.connectPromise;
  }

  on(type: string, handler: (data: any) => void) {
    this.messageHandlers.set(type, handler);
  }

  async send(type: string, payload: any) {
    // Ensure connected before sending
    await this.connect();
    
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    } else {
      console.warn('⚠️ WebSocket not connected, message not sent:', type);
      throw new Error('WebSocket not connected');
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