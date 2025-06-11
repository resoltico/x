// test-websocket.js
// Run this with: node test-websocket.js

import WebSocket from 'ws';

const ws = new WebSocket('ws://localhost:3000/ws');

console.log('Attempting to connect to WebSocket server...');

ws.on('open', () => {
  console.log('✅ Connected successfully!');
  console.log('Sending ping...');
  ws.send(JSON.stringify({ type: 'ping' }));
});

ws.on('message', (data) => {
  const message = JSON.parse(data.toString());
  console.log('📥 Received:', message);
  
  if (message.type === 'connection.established') {
    console.log('✅ Server acknowledged connection');
  } else if (message.type === 'pong') {
    console.log('✅ Server responded to ping');
    console.log('\n🎉 WebSocket is working correctly!');
    ws.close();
  }
});

ws.on('error', (error) => {
  console.error('❌ WebSocket error:', error.message);
  console.log('\nTroubleshooting tips:');
  console.log('1. Make sure the server is running (pnpm start)');
  console.log('2. Check that port 3000 is not in use');
  console.log('3. Look for errors in the server logs');
});

ws.on('close', () => {
  console.log('\n👋 Connection closed');
  process.exit(0);
});

// Timeout after 5 seconds
setTimeout(() => {
  console.error('\n❌ Connection timeout - no response from server');
  ws.close();
  process.exit(1);
}, 5000);