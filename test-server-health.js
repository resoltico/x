// test-server-health.js
// Run this with: node test-server-health.js

import fetch from 'node-fetch';
import WebSocket from 'ws';

const SERVER_URL = 'http://localhost:3000';
const WS_URL = 'ws://localhost:3000/ws';

console.log('🔍 Engraving Processor Pro - Server Health Check\n');

async function checkHttpServer() {
  console.log('1️⃣ Checking HTTP server...');
  try {
    const response = await fetch(`${SERVER_URL}/health`);
    if (response.ok) {
      const data = await response.json();
      console.log('✅ HTTP server is running');
      console.log('   Status:', data.status);
      console.log('   Uptime:', Math.round(data.uptime / 1000), 'seconds');
      console.log('   Services:', JSON.stringify(data.services, null, 2));
      return true;
    } else {
      console.error('❌ HTTP server returned status:', response.status);
      return false;
    }
  } catch (error) {
    console.error('❌ Cannot reach HTTP server:', error.message);
    console.log('\n💡 Make sure the server is running with: pnpm start');
    return false;
  }
}

async function checkWebSocket() {
  console.log('\n2️⃣ Checking WebSocket server...');
  
  return new Promise((resolve) => {
    const ws = new WebSocket(WS_URL);
    let timeout;
    
    timeout = setTimeout(() => {
      console.error('❌ WebSocket connection timeout (5 seconds)');
      ws.close();
      resolve(false);
    }, 5000);
    
    ws.on('open', () => {
      clearTimeout(timeout);
      console.log('✅ WebSocket connected successfully');
      
      // Test ping-pong
      console.log('   Testing ping-pong...');
      ws.send(JSON.stringify({ type: 'ping' }));
    });
    
    ws.on('message', (data) => {
      const message = JSON.parse(data.toString());
      
      if (message.type === 'connection.established') {
        console.log('   ✅ Received welcome message');
        console.log('      Client ID:', message.payload.clientId);
      } else if (message.type === 'pong') {
        console.log('   ✅ Ping-pong successful');
        
        // Test preview functionality
        console.log('   Testing preview request (with invalid image ID)...');
        ws.send(JSON.stringify({
          type: 'preview.update',
          payload: {
            imageId: 'test-invalid-id',
            parameters: {
              binarization: { method: 'sauvola', windowSize: 15, k: 0.34, r: 128 }
            }
          }
        }));
      } else if (message.type === 'error') {
        if (message.payload.message.includes('Image not found')) {
          console.log('   ✅ Error handling working correctly');
          console.log('      Error:', message.payload.message);
          ws.close();
          resolve(true);
        }
      }
    });
    
    ws.on('error', (error) => {
      clearTimeout(timeout);
      console.error('❌ WebSocket error:', error.message);
      resolve(false);
    });
    
    ws.on('close', () => {
      clearTimeout(timeout);
      console.log('   WebSocket connection closed');
    });
  });
}

async function checkStaticAssets() {
  console.log('\n3️⃣ Checking static assets...');
  
  const assets = [
    '/favicon.svg',
    '/build/.vite/manifest.json'
  ];
  
  let allGood = true;
  
  for (const asset of assets) {
    try {
      const response = await fetch(`${SERVER_URL}${asset}`);
      if (response.ok) {
        console.log(`   ✅ ${asset} - OK`);
      } else {
        console.log(`   ❌ ${asset} - Status ${response.status}`);
        allGood = false;
      }
    } catch (error) {
      console.log(`   ❌ ${asset} - Error: ${error.message}`);
      allGood = false;
    }
  }
  
  return allGood;
}

async function runHealthCheck() {
  const httpOk = await checkHttpServer();
  if (!httpOk) {
    console.log('\n❌ Server is not running. Please start it with: pnpm start');
    process.exit(1);
  }
  
  const wsOk = await checkWebSocket();
  const assetsOk = await checkStaticAssets();
  
  console.log('\n📊 Summary:');
  console.log('   HTTP Server:', httpOk ? '✅' : '❌');
  console.log('   WebSocket:', wsOk ? '✅' : '❌');
  console.log('   Static Assets:', assetsOk ? '✅' : '❌');
  
  if (httpOk && wsOk && assetsOk) {
    console.log('\n🎉 All systems operational! The server is ready.');
    process.exit(0);
  } else {
    console.log('\n⚠️ Some issues detected. Please check the errors above.');
    process.exit(1);
  }
}

// Add timeout for the entire check
setTimeout(() => {
  console.error('\n❌ Health check timeout. Server may not be responding.');
  process.exit(1);
}, 10000);

runHealthCheck();