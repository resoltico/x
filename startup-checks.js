// startup-checks.js
// This script runs basic checks before starting the server

import { fileURLToPath } from "url";
import { dirname, join } from "path";
import fs from "fs";

const __dirname = dirname(fileURLToPath(import.meta.url));

console.log('🔍 Running startup checks...\n');

let hasErrors = false;

// Check Node.js version
function checkNodeVersion() {
  console.log('1️⃣ Checking Node.js version...');
  const version = process.version;
  const major = parseInt(version.split('.')[0].substring(1));
  
  if (major >= 22) {
    console.log(`   ✅ Node.js ${version} (minimum: 22.0.0)`);
  } else {
    console.log(`   ❌ Node.js ${version} is too old (minimum: 22.0.0)`);
    hasErrors = true;
  }
}

// Check required directories
function checkDirectories() {
  console.log('\n2️⃣ Checking required directories...');
  
  const requiredDirs = [
    'build/server',
    'build/client',
    'build/client/assets',
    'public',
    'src/engine'
  ];
  
  for (const dir of requiredDirs) {
    const path = join(__dirname, dir);
    if (fs.existsSync(path)) {
      console.log(`   ✅ ${dir}`);
    } else {
      console.log(`   ❌ ${dir} - Missing! Run 'pnpm build' first`);
      hasErrors = true;
    }
  }
}

// Check critical files
function checkFiles() {
  console.log('\n3️⃣ Checking critical files...');
  
  const criticalFiles = [
    'build/server/index.js',
    'public/favicon.svg',
    'src/engine/pipeline/ProcessingPipeline.js'
  ];
  
  for (const file of criticalFiles) {
    const path = join(__dirname, file);
    if (fs.existsSync(path)) {
      console.log(`   ✅ ${file}`);
    } else {
      console.log(`   ❌ ${file} - Missing!`);
      hasErrors = true;
    }
  }
}

// Check port availability
async function checkPort() {
  console.log('\n4️⃣ Checking port availability...');
  const port = process.env.PORT || 3000;
  
  try {
    const net = await import('net');
    const server = net.createServer();
    
    await new Promise((resolve, reject) => {
      server.once('error', (err) => {
        if (err.code === 'EADDRINUSE') {
          console.log(`   ❌ Port ${port} is already in use`);
          console.log('      Try: lsof -i :3000 (to find process)');
          console.log('      Or set PORT=3001 in your environment');
          hasErrors = true;
          resolve();
        } else {
          reject(err);
        }
      });
      
      server.once('listening', () => {
        console.log(`   ✅ Port ${port} is available`);
        server.close();
        resolve();
      });
      
      server.listen(port);
    });
  } catch (error) {
    console.log(`   ⚠️ Could not check port: ${error.message}`);
  }
}

// Check environment
function checkEnvironment() {
  console.log('\n5️⃣ Checking environment...');
  
  const nodeEnv = process.env.NODE_ENV || 'development';
  console.log(`   ℹ️ NODE_ENV: ${nodeEnv}`);
  
  if (nodeEnv === 'production' && !fs.existsSync(join(__dirname, 'build'))) {
    console.log('   ❌ Production mode but no build found!');
    hasErrors = true;
  } else {
    console.log('   ✅ Environment OK');
  }
}

// Run all checks
async function runChecks() {
  checkNodeVersion();
  checkDirectories();
  checkFiles();
  await checkPort();
  checkEnvironment();
  
  console.log('\n' + '='.repeat(50));
  
  if (hasErrors) {
    console.log('\n❌ Startup checks failed!');
    console.log('   Please fix the issues above before starting the server.');
    process.exit(1);
  } else {
    console.log('\n✅ All startup checks passed!');
    console.log('   Server can be started safely.');
    process.exit(0);
  }
}

runChecks().catch(error => {
  console.error('\n❌ Startup check error:', error);
  process.exit(1);
});