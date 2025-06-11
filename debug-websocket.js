// debug-websocket.js
// Advanced WebSocket debugging utility
// Run this with: node debug-websocket.js

import WebSocket from 'ws';
import readline from 'readline';

const WS_URL = 'ws://localhost:3000/ws';

console.log('🔍 WebSocket Debug Tool\n');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

let ws = null;
let messageCount = 0;

function connect() {
  console.log('🔌 Connecting to', WS_URL);
  ws = new WebSocket(WS_URL);
  
  ws.on('open', () => {
    console.log('✅ Connected successfully!\n');
    showMenu();
  });
  
  ws.on('message', (data) => {
    messageCount++;
    const message = JSON.parse(data.toString());
    console.log(`\n📨 Message #${messageCount}:`);
    console.log(JSON.stringify(message, null, 2));
    
    if (message.type === 'error') {
      console.log('❌ Server reported error');
    }
  });
  
  ws.on('error', (error) => {
    console.error('\n❌ WebSocket error:', error.message);
  });
  
  ws.on('close', (code, reason) => {
    console.log(`\n🔌 Disconnected: code=${code}, reason=${reason || 'none'}`);
    ws = null;
  });
}

function showMenu() {
  console.log('\nAvailable commands:');
  console.log('1. Send ping');
  console.log('2. Send preview request (test)');
  console.log('3. Send invalid message');
  console.log('4. Show connection info');
  console.log('5. Reconnect');
  console.log('6. Exit');
  console.log();
  
  rl.question('Enter command (1-6): ', (answer) => {
    handleCommand(answer);
  });
}

function handleCommand(cmd) {
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    console.log('❌ Not connected. Use command 5 to reconnect.');
    showMenu();
    return;
  }
  
  switch(cmd) {
    case '1':
      console.log('🏓 Sending ping...');
      ws.send(JSON.stringify({ type: 'ping' }));
      setTimeout(showMenu, 1000);
      break;
      
    case '2':
      console.log('🎨 Sending preview request...');
      ws.send(JSON.stringify({
        type: 'preview.update',
        payload: {
          imageId: 'test-' + Date.now(),
          parameters: {
            binarization: {
              method: 'sauvola',
              windowSize: 15,
              k: 0.34,
              r: 128
            },
            morphology: { enabled: false },
            noise: { enabled: false },
            scaling: { method: 'none' }
          }
        }
      }));
      setTimeout(showMenu, 2000);
      break;
      
    case '3':
      console.log('💥 Sending invalid message...');
      ws.send('This is not JSON');
      setTimeout(showMenu, 1000);
      break;
      
    case '4':
      console.log('\n📊 Connection Info:');
      console.log('   State:', ws.readyState === WebSocket.OPEN ? 'OPEN' : 'NOT OPEN');
      console.log('   URL:', ws.url);
      console.log('   Messages received:', messageCount);
      console.log('   Buffer size:', ws.bufferedAmount);
      showMenu();
      break;
      
    case '5':
      console.log('🔄 Reconnecting...');
      if (ws) ws.close();
      setTimeout(connect, 500);
      break;
      
    case '6':
      console.log('👋 Goodbye!');
      if (ws) ws.close();
      process.exit(0);
      break;
      
    default:
      console.log('❓ Unknown command');
      showMenu();
  }
}

// Handle Ctrl+C
process.on('SIGINT', () => {
  console.log('\n\n👋 Shutting down...');
  if (ws) ws.close();
  process.exit(0);
});

// Start
connect();